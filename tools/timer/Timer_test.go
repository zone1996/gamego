package timer

import (
	"fmt"
	"testing"
	"time"
)

var fmtp = fmt.Println

func TestTimer(t *testing.T) {
	timer := NewTimer(nil)
	timer.Start()

	f1 := func() {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), "执行定时器任务-test-1")
	}

	f2 := func() {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), "执行定时器任务-test-2")
	}

	f3 := func() {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), "执行定时器任务-RunOnce")
	}

	fmt.Println("Start:", time.Now().Format("2006-01-02 15:04:05.000"))
	firstTime := time.Now().Add(2 * time.Second) // 2s后执行
	timer.RunAtFixedRate("test-1", f1, firstTime, 5*time.Second)
	timer.RunEveryNMinutes("test-2", f2, 1)
	timer.RunOnceAt("RunOnce", f3, firstTime)

	time.Sleep(2 * time.Minute)
}
