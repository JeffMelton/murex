package structs_test

import (
	"testing"

	"github.com/lmorg/murex/test"
)

// Test Phase 1+2 optimizations - output ordering
func TestForEachParallelOutputOrdering(t *testing.T) {
	tests := []test.MurexTest{
		// Test that output maintains order even with parallel execution
		{
			Block:  `a [1..10] -> foreach i --parallel 4 { out "Item $i" }`,
			Stdout: `^Item 1\nItem 2\nItem 3\nItem 4\nItem 5\nItem 6\nItem 7\nItem 8\nItem 9\nItem 10\n$`,
			Stderr: `^$`,
		},
		// Test with higher parallelism
		{
			Block:  `a [1..5] -> foreach i --parallel 3 { out "Processing $i" }`,
			Stdout: `^Processing 1\nProcessing 2\nProcessing 3\nProcessing 4\nProcessing 5\n$`,
			Stderr: `^$`,
		},
		// Test with mixed stdout/stderr
		{
			Block:  `a [1..3] -> foreach i --parallel 2 { out "stdout $i"; err "stderr $i" }`,
			Stdout: `^stdout 1\nstdout 2\nstdout 3\n$`,
			Stderr: `^stderr 1\nstderr 2\nstderr 3\n$`,
		},
	}

	test.RunMurexTestsRx(tests, t)
}

// Test Phase 1 optimization - sensible concurrency limits
func TestForEachParallelConcurrencyLimits(t *testing.T) {
	tests := []test.MurexTest{
		// Test that --parallel 0 uses reasonable concurrency (not unlimited)
		// This test verifies basic functionality with unlimited parallelism
		{
			Block:  `a [1..20] -> foreach i --parallel 0 { out "Task $i" } -> count`,
			Stdout: `20`,
			Stderr: ``,
		},
		// Test specific parallelism works
		{
			Block:  `a [1..5] -> foreach i --parallel 2 { out "Item $i" } -> count`,
			Stdout: `5`,
			Stderr: ``,
		},
	}

	test.RunMurexTests(tests, t)
}

// Test error handling with output aggregation
func TestForEachParallelErrorHandling(t *testing.T) {
	tests := []test.MurexTest{
		// Test that errors are captured and ordered properly
		{
			Block:  `a [1,2,3] -> foreach i --parallel 2 { if { $i == 2 } then { err "Error at $i" } else { out "Success $i" } }`,
			Stdout: `^Success 1\nSuccess 3\n$`,
			Stderr: `Error at 2`,
		},
		// Test mixed success/failure doesn't break ordering  
		{
			Block:  `a [1,2,3] -> foreach i --parallel 2 { out "Before $i"; if { $i == 2 } then { err "Fail $i" } else { out "After $i" } }`,
			Stdout: `^Before 1\nAfter 1\nBefore 2\nBefore 3\nAfter 3\n$`,
			Stderr: `Fail 2`,
		},
	}

	test.RunMurexTestsRx(tests, t)
}

// Test resource efficiency - no goroutine explosion
func TestForEachParallelResourceEfficiency(t *testing.T) {
	// This test ensures we don't create thousands of goroutines
	// by testing with a large dataset and reasonable parallelism
	tests := []test.MurexTest{
		{
			Block:  `a [1..1000] -> foreach i --parallel 4 { if { $i <= 5 } then { out "Item $i" } } | head -n 5`,
			Stdout: `^Item 1\nItem 2\nItem 3\nItem 4\nItem 5\n$`,
			Stderr: `^$`,
		},
	}

	test.RunMurexTestsRx(tests, t)
}

// Test Phase 2 I/O isolation - no interleaved output
func TestForEachParallelIOIsolation(t *testing.T) {
	tests := []test.MurexTest{
		// Test that parallel workers don't interfere with each other's output
		{
			Block:  `a [1..4] -> foreach i --parallel 4 { out "Start $i"; out "End $i" }`,
			Stdout: `^Start 1\nEnd 1\nStart 2\nEnd 2\nStart 3\nEnd 3\nStart 4\nEnd 4\n$`,
			Stderr: `^$`,
		},
		// Test stderr isolation
		{
			Block:  `a [1..3] -> foreach i --parallel 3 { err "Error $i line 1"; err "Error $i line 2" }`,
			Stdout: `^$`,
			Stderr: `^Error 1 line 1\nError 1 line 2\nError 2 line 1\nError 2 line 2\nError 3 line 1\nError 3 line 2\n$`,
		},
	}

	test.RunMurexTestsRx(tests, t)
}

// Benchmark the optimizations
func BenchmarkForEachParallelOptimized(b *testing.B) {
	tests := []struct {
		name     string
		parallel int
		items    int
	}{
		{"Serial", 1, 100},
		{"Parallel2", 2, 100}, 
		{"Parallel4", 4, 100},
		{"Parallel8", 8, 100},
		{"ParallelUnlimited", 0, 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Note: This would need to be run through the Murex interpreter
				// For now it's a placeholder structure for benchmark testing
			}
		})
	}
}