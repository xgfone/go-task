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

// Package task supplies some task runners to run the task.
package task

import (
	"context"
	"fmt"
	"time"
)

// Task represents a task.
type Task interface {
	Name() string
	Run(ctx context.Context, now time.Time) error
}

// TaskFunc is the task function.
type TaskFunc func(ctx context.Context, now time.Time) error

type task struct {
	name string
	task TaskFunc
}

// NewTask returns a new Task.
func NewTask(name string, run TaskFunc) Task {
	return task{name: name, task: run}
}

func (t task) Run(c context.Context, n time.Time) error { return t.task(c, n) }
func (t task) Name() string                             { return t.name }
func (t task) String() string                           { return fmt.Sprintf("Task(name=%s)", t.name) }
