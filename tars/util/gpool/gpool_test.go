package gpool

import (
	"io/ioutil"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func init() {
	println("using MAXPROC")
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
}

func TestNewPool(t *testing.T) {
	pool := NewPool(1000, 10000)
	defer pool.Release()

	iterations := 1000000
	var counter uint64 = 0

	wg := sync.WaitGroup{}

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		arg := uint64(1)

		job := func() {
			defer wg.Done()
			atomic.AddUint64(&counter, arg)
		}

		pool.JobQueue <- job
	}

	wg.Wait()

	counterFinal := atomic.LoadUint64(&counter)
	if uint64(iterations) != counterFinal {
		t.Errorf("iterations %v is not equal counterFinal %v", iterations, counterFinal)
	}
}

func TestRelease(t *testing.T) {
	grNum := runtime.NumGoroutine()
	pool := NewPool(5, 10)
	defer func() {
		pool.Release()

		if grNum != runtime.NumGoroutine() {
			t.Errorf("grNum %v is not equal to runtime.NumGoroutine %v", grNum, runtime.NumGoroutine())
		}
	}()

	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		job := func() {
			defer wg.Done()
		}

		pool.JobQueue <- job
	}

	wg.Wait()
}

func BenchmarkPool(b *testing.B) {
	pool := NewPool(1, 10)
	defer pool.Release()

	log.SetOutput(ioutil.Discard)

	for n := 0; n < b.N; n++ {
		pool.JobQueue <- func() {
			b.Logf("I am worker of %d\n", n)
		}
	}
}
