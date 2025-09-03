# Foreach --Parallel Performance Optimizations

## Overview

This document details the comprehensive performance optimizations implemented for the `foreach --parallel` feature in Murex. The optimizations address critical bottlenecks that were causing poor performance and system instability, particularly on different hardware architectures (Linux x86_64 vs MacBook Pro M1).

## Problem Analysis

### Original Issues
- **Unlimited Concurrency**: Used `MAX_INT` for parallel limit, creating thousands of goroutines
- **Goroutine-per-iteration**: Created new goroutine for every input item
- **Process Forking Overhead**: Heavy process creation without resource reuse
- **Output Race Conditions**: Parallel workers writing directly to parent streams
- **Memory Allocation**: Constant allocation/deallocation of result objects
- **Poor Channel Design**: Fixed buffer sizes regardless of parallelism level

### Performance Impact
- System thrashing under high concurrency
- Inconsistent performance across different hardware
- Memory pressure from excessive goroutine creation
- Race conditions causing output interleaving
- Poor scalability with large datasets

## Optimization Phases

### Phase 1: Worker Pool Architecture ✅

**Goal**: Replace goroutine-per-iteration with fixed worker pools and sensible concurrency limits.

**Implementation**:
```go
// Sensible concurrency limits
if parallel < 1 {
    parallel = runtime.NumCPU() * 8 // Much more reasonable than MAX_INT
}

// Worker pool pattern
for i := 0; i < workerCount; i++ {
    wg.Add(1)
    go parallelWorkerWithAggregation(execCtx, jobCh, resCh, &wg)
}
```

**Benefits**:
- Eliminated goroutine explosion (MAX_INT → runtime.NumCPU() * 8)
- Persistent workers reduce creation/teardown overhead
- Predictable resource usage across different systems
- Better CPU utilization through controlled parallelism

### Phase 2: Output Aggregation and I/O Isolation ✅

**Goal**: Prevent race conditions and maintain output ordering despite parallel execution.

**Implementation**:
```go
type resultAggregator struct {
    results   map[int]*jobResult
    mutex     sync.Mutex
    nextIndex int
    parent    *lang.Process
}

// Per-worker I/O streams with ordered output
func (ra *resultAggregator) flushReady() {
    for {
        result, exists := ra.results[ra.nextIndex]
        if !exists { break }
        
        // Write outputs in order
        if len(result.stdout) > 0 {
            ra.parent.Stdout.Write(result.stdout)
        }
        // ... process stderr and errors
        ra.nextIndex++
    }
}
```

**Benefits**:
- Maintains correct output sequence despite parallel execution
- Eliminates I/O race conditions between workers
- Proper error handling and ordering
- Clean separation of worker output and parent process streams

### Phase 3: Object Pooling and Resource Optimization ✅

**Goal**: Reduce memory allocation overhead and garbage collection pressure.

**Implementation**:
```go
var (
    // Pool for reusing job result objects
    resultPool = sync.Pool{
        New: func() interface{} {
            return &jobResult{}
        },
    }
)

// Optimized channel buffering
jobCh := make(chan parallelJob, workerCount*2)
resCh := make(chan *jobResult, workerCount*2)  // 2x parallel for better throughput
```

**Benefits**:
- Reduced GC pressure through object reuse
- Optimized channel buffer sizes based on worker count
- Better memory utilization patterns
- Consistent performance under sustained load

## Architecture Details

### Worker Pool Flow

```
Input Stream → Job Queue → Worker Pool → Result Aggregator → Ordered Output
     │              │            │              │                   │
     └─ ReadArray   └─ Jobs      └─ Workers     └─ Results         └─ Parent Streams
        WithType       Channel      (Fixed)        Ordered           (Stdout/Stderr)
```

### Key Components

1. **parallelExecContext**: Shared context containing block code and variables
2. **parallelJob**: Work unit with iteration number and data
3. **jobResult**: Result container with output capture and error handling
4. **resultAggregator**: Ordered output processor with synchronization
5. **Object Pools**: Resource reuse for frequent allocations

### Concurrency Model

- **Job Distribution**: Single producer (input reader) → Multiple consumers (workers)
- **Result Collection**: Multiple producers (workers) → Single consumer (aggregator)
- **Output Ordering**: Map-based result storage with sequential index processing
- **Synchronization**: WaitGroups for worker lifecycle + Mutex for result access

## Performance Testing

### Test Coverage

The optimization includes comprehensive tests for:

```go
// Output ordering despite parallel execution
TestForEachParallelOutputOrdering

// Sensible concurrency limits (not unlimited)
TestForEachParallelConcurrencyLimits  

// Proper error handling and ordering
TestForEachParallelErrorHandling

// No goroutine explosion with large datasets
TestForEachParallelResourceEfficiency

// I/O isolation between workers
TestForEachParallelIOIsolation
```

### Benchmarking Results

Performance tests show:
- Consistent execution times across different parallelism levels
- No system thrashing with large datasets (1000+ items)
- Proper output ordering maintained under high concurrency
- Stable memory usage patterns

### Validation Commands

```bash
# Build and test
go build -o bin/murex
go test ./builtins/core/structs/ -run TestForEachParallel

# Performance testing
./test_parallel_performance.sh

# Example usage
./bin/murex -c 'a [1..1000] -> foreach i --parallel 8 { out "Processing $i" }'
```

## Implementation Files

### Core Implementation
- `builtins/core/structs/foreach_parallel.go` - Main optimization implementation
- `builtins/core/structs/foreach_parallel_optimized_test.go` - Comprehensive test suite

### Key Functions
- `cmdForEachParallel()` - Entry point with concurrency limits
- `runParallelWorkerPool()` - Worker pool orchestration
- `parallelWorkerWithAggregation()` - Individual worker implementation  
- `executeParallelJobWithCapture()` - Job execution with I/O capture
- `resultAggregator.processResults()` - Ordered output processing

## Future Optimization Opportunities (Phase 4)

While the core performance issues have been resolved, additional optimizations could include:

1. **Parse-once Optimization**: Cache parsed code blocks for repeated execution
2. **Context Pooling**: Reuse execution contexts to reduce allocation overhead
3. **Advanced Buffer Tuning**: Dynamic buffer sizing based on workload patterns
4. **Memory Pre-allocation**: Pre-allocate result maps and slices
5. **CPU Affinity**: Advanced worker thread management for NUMA systems

## Migration Notes

The optimizations maintain full backward compatibility:
- All existing `foreach --parallel` syntax works unchanged
- Output format and ordering remain identical
- Error handling behavior is preserved
- Flag processing and variable scoping unchanged

## Monitoring and Debugging

For performance monitoring:
```bash
# Monitor goroutine count
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Memory allocation profiling  
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profiling during execution
go tool pprof http://localhost:6060/debug/pprof/profile
```

## Conclusion

These optimizations transform `foreach --parallel` from a resource-intensive operation prone to system thrashing into a well-behaved, predictable parallel processing tool. The improvements provide:

- **Stability**: Consistent performance across different hardware architectures
- **Scalability**: Handles large datasets without system resource exhaustion  
- **Correctness**: Maintains output ordering and proper error handling
- **Efficiency**: Reduced memory allocation and CPU overhead

The implementation serves as a model for high-performance parallel processing in shell environments while maintaining the simplicity and predictability expected by users.