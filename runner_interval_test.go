// Copyright 2021 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"context"
	"fmt"
	"time"
)

func newTask(name string) Task {
	return NewTask(name, func(ctx context.Context, now time.Time) error {
		fmt.Printf("%s: run task '%s'\n", now.UTC().Format(time.RFC3339), name)
		return nil
	})
}

func ExampleNewIntervalRunner() {
	runner := NewIntervalRunner(&IntervalRunnerConfig{
		Interval: time.Second * 3,
		Tick:     time.Second,
	})

	runner.AddTask(newTask("task1"))
	runner.AddTask(newTask("task2"))
	runner.AddTask(NewIntervalTask(time.Second*5, newTask("task3")))
	runner.AddTask(NewIntervalTask(time.Second*10, newTask("task4")))

	time.Sleep(time.Second * 25)
	runner.Stop()

	// It will output like this:
	// 2021-06-03T14:56:36Z: run task 'task3'
	// 2021-06-03T14:56:36Z: run task 'task4'
	// 2021-06-03T14:56:36Z: run task 'task1'
	// 2021-06-03T14:56:36Z: run task 'task2'
	// 2021-06-03T14:56:39Z: run task 'task2'
	// 2021-06-03T14:56:39Z: run task 'task1'
	// 2021-06-03T14:56:42Z: run task 'task2'
	// 2021-06-03T14:56:42Z: run task 'task3'
	// 2021-06-03T14:56:42Z: run task 'task1'
	// 2021-06-03T14:56:46Z: run task 'task1'
	// 2021-06-03T14:56:46Z: run task 'task4'
	// 2021-06-03T14:56:46Z: run task 'task2'
	// 2021-06-03T14:56:48Z: run task 'task3'
	// 2021-06-03T14:56:49Z: run task 'task1'
	// 2021-06-03T14:56:49Z: run task 'task2'
	// 2021-06-03T14:56:53Z: run task 'task2'
	// 2021-06-03T14:56:53Z: run task 'task3'
	// 2021-06-03T14:56:53Z: run task 'task1'
	// 2021-06-03T14:56:56Z: run task 'task4'
	// 2021-06-03T14:56:56Z: run task 'task1'
	// 2021-06-03T14:56:56Z: run task 'task2'
	// 2021-06-03T14:56:58Z: run task 'task3'

}
