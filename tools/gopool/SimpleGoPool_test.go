package gopool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSimpleGoPool(t *testing.T) {

	var wg sync.WaitGroup

	command := func() {
		time.Sleep(100 * time.Millisecond) // sleep 0.1s
		wg.Done()
	}

	executor := NewSimpleExecutor(100, 10000) // step1
	fmt.Println("Size:", executor.Size())

	begin := time.Now().UnixNano()
	for i := 0; i < 5000; i++ {
		wg.Add(1)
		executor.Execute(command) // step2
	}
	wg.Wait()
	end := time.Now().UnixNano()

	fmt.Println("Total ms:", (end-begin)/int64(time.Millisecond))
}
