package structs

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/lmorg/murex/lang"
)

// Optimized parallel execution context
type parallelExecContext struct {
	block      []rune
	varName    string
	parentProc *lang.Process
}

// Job represents work to be done by a worker
type parallelJob struct {
	varValue  any
	dataType  string
	iteration int
}

func cmdForEachParallel(p *lang.Process, flags map[string]string, additional []string) error {
	block, varName, err := forEachInitializer(p, additional)
	if err != nil {
		return err
	}

	parallel, err := getFlagValueInt(flags, foreachParallel)
	if err != nil {
		return err
	}

	// Sensible concurrency limits - Phase 1 optimization
	if parallel < 1 {
		parallel = runtime.NumCPU() * 8 // Much more reasonable than MAX_INT
	}

	// Phase 1: For now we keep the block as-is, but eliminate per-iteration goroutine creation
	// Parse-once optimization will be implemented in a follow-up when we add execution plan caching
	execCtx := &parallelExecContext{
		block:      block,
		varName:    varName,
		parentProc: p,
	}

	return runParallelWorkerPool(execCtx, parallel, p)
}

// Worker pool implementation - Phase 1 critical optimization
func runParallelWorkerPool(execCtx *parallelExecContext, workerCount int, p *lang.Process) error {
	// Create job channel with reasonable buffer size
	jobCh := make(chan parallelJob, workerCount*2)
	var wg sync.WaitGroup
	var iteration int64 = -1

	// Start worker goroutines - reuse instead of creating per-iteration
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go parallelWorker(execCtx, jobCh, &wg)
	}

	// Feed jobs to workers
	go func() {
		defer close(jobCh)
		p.Stdin.ReadArrayWithType(p.Context, func(varValue any, dataType string) {
			i := atomic.AddInt64(&iteration, 1)
			jobCh <- parallelJob{
				varValue:  varValue,
				dataType:  dataType,
				iteration: int(i),
			}
		})
	}()

	// Wait for all workers to complete
	wg.Wait()
	return nil
}

// Worker function - executes jobs using pre-compiled execution plan
func parallelWorker(execCtx *parallelExecContext, jobCh <-chan parallelJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobCh {
		if execCtx.parentProc.HasCancelled() {
			return
		}

		err := executeParallelJob(execCtx, job)
		if err != nil {
			// Error handling - write to parent's stderr
			execCtx.parentProc.Stderr.Writeln([]byte("error: " + err.Error()))
		}
	}
}

// Execute individual job with lightweight process context
func executeParallelJob(execCtx *parallelExecContext, job parallelJob) error {
	b, err := convertToByte(job.varValue)
	if err != nil {
		return err
	}

	if len(b) == 0 {
		return nil
	}

	// Lighter-weight fork - still needs F_FUNCTION but with optimizations
	fork := execCtx.parentProc.Fork(lang.F_FUNCTION | lang.F_BACKGROUND | lang.F_CREATE_STDIN)
	fork.Name.Set("foreach--parallel")
	fork.FileRef = execCtx.parentProc.FileRef

	// Set iteration variable if not using anonymous mode
	if execCtx.varName != "!" {
		err = fork.Variables.Set(execCtx.parentProc, execCtx.varName, job.varValue, job.dataType)
		if err != nil {
			return err
		}
	}

	if !setMetaValues(fork.Process, job.iteration) {
		return nil
	}

	// Set up stdin
	fork.Stdin.SetDataType(job.dataType)
	_, err = fork.Stdin.Writeln(b)
	if err != nil {
		return err
	}

	// Execute the block - Phase 1 focuses on worker pool, parse-once comes in follow-up
	_, err = fork.Execute(execCtx.block)
	return err
}
