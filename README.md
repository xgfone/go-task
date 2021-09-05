# go-task [![Build Status](https://github.com/xgfone/go-task/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-task/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-task)](https://pkg.go.dev/github.com/xgfone/go-task) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-task/master/LICENSE)

It supplies some task runners to run the task. Support `Go1.7+`.


## Installation
```shell
$ go get -u github.com/xgfone/go-task
```

## Example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xgfone/go-task"
)

type tasker struct {
	name string
}

func (t tasker) Name() string { return t.name }
func (t tasker) Run(ctx context.Context, now time.Time) (err error) {
	fmt.Printf("[%s] run task '%s'\n", now.Format(time.RFC3339), t.name)
	return
}

func newTask(name string) task.Task { return tasker{name: name} }

func runTask(ctx context.Context, now time.Time) (err error) {
	t := task.GetTaskFromCtx(ctx)
	fmt.Printf("[%s] run task '%s'\n", now.Format(time.RFC3339), t.Name())
	return
}

func main() {
	config := task.IntervalRunnerConfig{Tick: time.Second, Interval: time.Second * 3}
	runner := task.NewIntervalRunner(&config)

	// Add all the tasks
	runner.AddTask(newTask("task1"))                                        // Use Default Interval: 3s
	runner.AddTask(task.NewTask("task2", runTask))                          // Use Default Interval: 3s
	runner.AddTask(task.NewIntervalFuncTask("task3", time.Second, runTask)) // Use Special Interval: 1s

	// We only run the tasks for 10s.
	time.Sleep(time.Second * 10)
	runner.Stop()

	// Output:
	// [2021-06-03T23:41:14+08:00] run task 'task3'
	// [2021-06-03T23:41:14+08:00] run task 'task1'
	// [2021-06-03T23:41:14+08:00] run task 'task2'
	// [2021-06-03T23:41:16+08:00] run task 'task3'
	// [2021-06-03T23:41:17+08:00] run task 'task2'
	// [2021-06-03T23:41:17+08:00] run task 'task3'
	// [2021-06-03T23:41:17+08:00] run task 'task1'
	// [2021-06-03T23:41:19+08:00] run task 'task3'
	// [2021-06-03T23:41:20+08:00] run task 'task2'
	// [2021-06-03T23:41:20+08:00] run task 'task3'
	// [2021-06-03T23:41:20+08:00] run task 'task1'
	// [2021-06-03T23:41:22+08:00] run task 'task3'
}
```
