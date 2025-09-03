package structs

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/lmorg/murex/builtins/pipes/streams"
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

// Output aggregation for ordered results - Phase 2 optimization
type jobResult struct {
	iteration int
	stdout    []byte
	stderr    []byte
	err       error
}

// Result aggregator for maintaining output order
type resultAggregator struct {
	results   map[int]*jobResult
	mutex     sync.Mutex
	nextIndex int
	parent    *lang.Process
}

// Phase 3: Object pools for resource optimization
var (
	// Pool for reusing job result objects
	resultPool = sync.Pool{
		New: func() interface{} {
			return &jobResult{}
		},
	}
)

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

// Worker pool with output aggregation - Phase 2 optimization
func runParallelWorkerPool(execCtx *parallelExecContext, workerCount int, p *lang.Process) error {
	// Create job and result channels with tuned buffering - Phase 3: optimized capacity
	jobCh := make(chan parallelJob, workerCount*2)
	resCh := make(chan *jobResult, workerCount*2)  // Phase 3: 2x parallel for better throughput
	var wg sync.WaitGroup
	var iteration int64 = -1

	// Initialize result aggregator - Phase 2: ordered output
	aggregator := &resultAggregator{
		results:   make(map[int]*jobResult),
		nextIndex: 0,
		parent:    p,
	}

	// Start result aggregator goroutine with wait group
	var aggregatorWg sync.WaitGroup
	aggregatorWg.Add(1)
	go func() {
		defer aggregatorWg.Done()
		aggregator.processResults(resCh)
	}()

	// Start worker goroutines - reuse instead of creating per-iteration
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go parallelWorkerWithAggregation(execCtx, jobCh, resCh, &wg)
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
	close(resCh)

	// Wait for result aggregator to finish processing all results
	aggregatorWg.Wait()
	
	// Flush any remaining results
	aggregator.flush()
	return nil
}

// Worker function with output aggregation - Phase 2 optimization
func parallelWorkerWithAggregation(execCtx *parallelExecContext, jobCh <-chan parallelJob, resultCh chan<- *jobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobCh {
		if execCtx.parentProc.HasCancelled() {
			return
		}

		// Execute job and capture output
		stdout, stderr, err := executeParallelJobWithCapture(execCtx, job)
		
		// Get pooled result object - Phase 3: reduce allocation
		result := resultPool.Get().(*jobResult)
		result.iteration = job.iteration
		result.stdout = stdout
		result.stderr = stderr
		result.err = err
		
		// Send result to aggregator
		resultCh <- result
	}
}

// Execute individual job with output capture and object pooling - Phase 3 optimization
func executeParallelJobWithCapture(execCtx *parallelExecContext, job parallelJob) ([]byte, []byte, error) {
	b, err := convertToByte(job.varValue)
	if err != nil {
		return nil, nil, err
	}

	if len(b) == 0 {
		return nil, nil, nil
	}

	// Create fresh streams for each execution - Phase 3: simplified approach
	// Stream objects maintain complex internal state that can't be easily reset
	stdin := streams.NewStdin()
	stdout := streams.NewStdin() 
	stderr := streams.NewStdin()

	// Lightweight fork with fresh streams - Phase 3: resource optimization
	fork := execCtx.parentProc.Fork(lang.F_FUNCTION | lang.F_BACKGROUND)
	fork.Name.Set("foreach--parallel")
	fork.FileRef = execCtx.parentProc.FileRef
	
	// Use fresh streams
	fork.Stdin = stdin
	fork.Stdout = stdout
	fork.Stderr = stderr

	// Set iteration variable if not using anonymous mode
	if execCtx.varName != "!" {
		err = fork.Variables.Set(execCtx.parentProc, execCtx.varName, job.varValue, job.dataType)
		if err != nil {
			return nil, nil, err
		}
	}

	if !setMetaValues(fork.Process, job.iteration) {
		return nil, nil, nil
	}

	// Set up stdin
	fork.Stdin.SetDataType(job.dataType)
	_, err = fork.Stdin.Writeln(b)
	if err != nil {
		return nil, nil, err
	}

	// Execute the block
	_, err = fork.Execute(execCtx.block)
	
	// Capture outputs from streams
	stdoutBytes, _ := fork.Stdout.ReadAll()
	stderrBytes, _ := fork.Stderr.ReadAll()
	
	return stdoutBytes, stderrBytes, err
}

// Process results in order - Phase 2: ordered output aggregation
func (ra *resultAggregator) processResults(resultCh <-chan *jobResult) {
	for result := range resultCh {
		ra.mutex.Lock()
		ra.results[result.iteration] = result
		ra.flushReady()
		ra.mutex.Unlock()
	}
}

// Flush results that are ready (in order)
func (ra *resultAggregator) flushReady() {
	for {
		result, exists := ra.results[ra.nextIndex]
		if !exists {
			break
		}
		
		// Write outputs in order
		if len(result.stdout) > 0 {
			ra.parent.Stdout.Write(result.stdout)
		}
		if len(result.stderr) > 0 {
			ra.parent.Stderr.Write(result.stderr)
		}
		if result.err != nil {
			ra.parent.Stderr.Writeln([]byte("error: " + result.err.Error()))
		}
		
		// Clean up and return to pool - Phase 3: resource reuse
		delete(ra.results, ra.nextIndex)
		ra.nextIndex++
		
		// Reset and return result object to pool
		result.stdout = nil
		result.stderr = nil
		result.err = nil
		resultPool.Put(result)
	}
}

// Final flush for any remaining results
func (ra *resultAggregator) flush() {
	ra.mutex.Lock()
	defer ra.mutex.Unlock()
	ra.flushReady()
}
