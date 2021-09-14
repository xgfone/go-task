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
	"log"
	"math/rand"
	"sync"
	"time"
)

// IntervalTask is the interval task to be run.
type IntervalTask interface {
	Interval() time.Duration
	Task
}

type intervalTask struct {
	interval time.Duration
	Task
}

func (t intervalTask) Interval() time.Duration { return t.interval }

// NewIntervalTask returns a new interval task.
func NewIntervalTask(interval time.Duration, task Task) IntervalTask {
	return intervalTask{
		interval: interval,
		Task:     task,
	}
}

// NewIntervalFuncTask is equal to NewIntervalTask(interval, NewTask(name, run)).
func NewIntervalFuncTask(name string, interval time.Duration, run TaskFunc) IntervalTask {
	return NewIntervalTask(interval, NewTask(name, run))
}

////////////////////////////////////////////////////////////////////////////

type taskWrapper struct {
	Running bool
	Last    time.Time
	Lock    sync.RWMutex
	Task
}

func (tw *taskWrapper) GetState() (running bool, last time.Time) {
	tw.Lock.RLock()
	running, last = tw.Running, tw.Last
	tw.Lock.RUnlock()
	return
}

func (tw *taskWrapper) SetState(running bool, now time.Time) {
	tw.Lock.Lock()
	tw.Running = running
	if !now.IsZero() {
		tw.Last = now
	}
	tw.Lock.Unlock()
}

// IntervalRunnerConfig is used to configure the interval runner.
type IntervalRunnerConfig struct {
	// The default interval that the task is run between twice.
	//
	// Default: 1m
	Interval time.Duration

	// Jitter is used to produce a random duration, [0, Jitter],
	// to be added to the interval duration of the task to avoid
	// converging on periodic behavior. If 0, disalbe it.
	//
	// Default: 0
	Jitter time.Duration

	// The clock tick interval.
	//
	// Default: 5s
	Tick time.Duration

	// Report whether the tasks can be run or not.
	// If nil, it is equal to return true.
	//
	// Default: nil
	CanRun func() bool

	// ErrorLog is used to log the error returned by the task, if set.
	//
	// Default: log.Printf
	ErrorLog func(fmt string, args ...interface{})
}

func (c *IntervalRunnerConfig) init() {
	if c.Interval == 0 {
		c.Interval = time.Minute
	}
	if c.Tick == 0 {
		c.Tick = time.Second * 5
	}
	if c.ErrorLog == nil {
		c.ErrorLog = log.Printf
	}
}

// intervalRunner is a task runner to run the tasks at regular intervals.
type intervalRunner struct {
	jitter   int64
	random   *rand.Rand
	interval time.Duration
	logerror func(string, ...interface{})
	canRun   func() bool
	tasks    map[string]*taskWrapper
	lock     sync.RWMutex

	once    sync.Once
	cancel  func()
	context context.Context
}

// NewIntervalRunner returns a new runner to run the tasks at regular intervals.
func NewIntervalRunner(config *IntervalRunnerConfig) Runner {
	ctx, cancel := context.WithCancel(context.Background())

	var c IntervalRunnerConfig
	if config != nil {
		c = *config
	}

	c.init()
	src := rand.NewSource(time.Now().UnixMicro())
	r := &intervalRunner{
		jitter:   int64(c.Jitter),
		random:   rand.New(src),
		interval: c.Interval,
		logerror: c.ErrorLog,
		canRun:   c.CanRun,
		tasks:    make(map[string]*taskWrapper, 16),

		cancel:  cancel,
		context: ctx,
	}
	go r.loop(c.Tick)

	return r
}

// AddTask adds teh task to the runner, and panic if the task has been added.
func (r *intervalRunner) AddTask(task Task) {
	name := task.Name()
	r.lock.Lock()
	_, ok := r.tasks[name]
	if !ok {
		r.tasks[name] = &taskWrapper{Task: task}
	}
	r.lock.Unlock()

	if ok {
		panic(fmt.Errorf("the task '%s' has been added", name))
	}
}

// DelTask deletes the task from the runner.
func (r *intervalRunner) DelTask(task Task) {
	name := task.Name()
	r.lock.Lock()
	delete(r.tasks, name)
	r.lock.Unlock()
}

// Stop stops the runner.
func (r *intervalRunner) Stop() { r.once.Do(r.cancel) }

func (r *intervalRunner) loop(tick time.Duration) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-r.context.Done():
			return
		case now := <-ticker.C:
			if r.canRun == nil || r.canRun() {
				r.lock.RLock()
				for _, task := range r.tasks {
					running, last := task.GetState()
					if running {
						continue
					}

					interval := r.interval
					if itask, ok := task.Task.(IntervalTask); ok {
						if iv := itask.Interval(); iv > 0 {
							interval = iv
						}
					}

					if r.jitter > 0 {
						interval += time.Duration(r.random.Int63n(r.jitter))
					}

					if now.Sub(last) >= interval {
						task.SetState(true, time.Time{})
						go r.runTask(task, now)
					}
				}
				r.lock.RUnlock()
			}
		}
	}
}

func (r *intervalRunner) runTask(task *taskWrapper, now time.Time) {
	defer task.SetState(false, now)
	if err := task.Run(setTaskToCtx(r.context, task.Task), now); err != nil {
		r.logerror("task '%s' error: %s", task.Name(), err.Error())
	}
}
