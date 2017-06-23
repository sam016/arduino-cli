/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */
package common

import (
	"math"
	"sync"

	lcf "github.com/Robpol86/logrus-custom-formatter"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

func init() {
	log = logrus.WithFields(logrus.Fields{})
	logrus.SetFormatter(lcf.NewFormatter("%[message]s", nil))
}

// A TaskWrapper wraps a task to be executed to allow
// Useful messages to be print. It is used to pretty
// print operations.
//
// All Message arrays use VERBOSITY as index.
type TaskWrapper struct {
	beforeMessage []string
	task          Task
	afterMessage  []string
	errorMessage  []string
	Result        interface{}
}

// verbosity represents the verbosity level of the message.
//
// Examples:
//
// verbosity 0 Message ""
// verbosity 1 Message "Hi"
// verbosity 2 Message "Hello folk, how are you?"
// type verbosity int

// Task represents a function which can be safely wrapped into a TaskWrapper.
//
// It may provide a result but always provides an error.
type Task func() TaskResult

//TaskResult represents a result from a task, or an error.
type TaskResult struct {
	Result interface{}
	Error  error
}

//TaskSequence represents a sequence of tasks.
type TaskSequence func() []TaskResult

// Execute executes a task while printing messages to describe what is happening.
func (tw TaskWrapper) Execute(verb int) TaskResult {
	var maxUsableVerb int
	var msg string
	maxUsableVerb = minVerb(verb, tw.beforeMessage)
	msg = tw.beforeMessage[maxUsableVerb]
	if msg != "" {
		log.Infof("%s ... ", msg)
	}

	ret := tw.task()

	if ret.Error != nil {
		maxUsableVerb = minVerb(verb, tw.errorMessage)
		msg = tw.errorMessage[maxUsableVerb]
		log.Warn("ERROR\n")
		if msg != "" {
			log.Warnf("%s\n", msg)
		}
	} else {
		maxUsableVerb = minVerb(verb, tw.afterMessage)
		msg = tw.afterMessage[maxUsableVerb]
		log.Info("OK\n")
		if msg != "" {
			log.Infof("%s\n", msg)
		}
	}
	return ret
}

// Task returns task of this wrapper.
func (tw TaskWrapper) Task() Task {
	return tw.task
}

// minVerb tells which is the max level of verbosity for the specified verbosity level (set by another
// function call) and the provided array of strings.
//
// Refer to TaskRunner struct for the usage of the array.
func minVerb(verb1 int, sentences []string) int {
	return int(math.Min(float64(verb1), float64(len(sentences)-1)))
}

// CreateTaskSequence returns a task to execute parameter tasks in sequence.
//
// if abortOnFailure = true then the sequence is aborted with the error,
// otherwise there is just an error logged.
func CreateTaskSequence(taskWrappers []TaskWrapper, ignoreOnFailure []bool, verbosity int) TaskSequence {
	results := make([]TaskResult, 0, 10)

	return TaskSequence(func() []TaskResult {
		for i, taskWrapper := range taskWrappers {
			result := taskWrapper.Execute(verbosity)
			results = append(results, result)
			if result.Error != nil {
				if ignoreOnFailure[i] {
					break
				}
				log.Warnf("Warning from task %d: %s", i, result.Error)
			}
		}
		return results
	})
}

// ExecuteSequence creates and executes an array of tasks in strict sequence.
func ExecuteSequence(taskWrappers []TaskWrapper, ignoreOnFailure []bool, verbosity int) []TaskResult {
	return CreateTaskSequence(taskWrappers, ignoreOnFailure, verbosity)()
}

// ExecuteParallel executes a set of TaskWrappers in parallel, handling concurrency for results.
func ExecuteParallel(taskWrappers []TaskWrapper, verbosity int) []TaskResult {
	results := make(chan TaskResult)
	wg := sync.WaitGroup{}

	wg.Add(len(taskWrappers))

	for _, task := range taskWrappers {
		go func(task TaskWrapper) {
			defer wg.Done()
			results <- task.Execute(verbosity)
		}(task)
	}
	wg.Wait()
	array := make([]TaskResult, len(results))
	for i := range array {
		array[i] = <-results
	}
	return array
}