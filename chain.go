package cron

import (
	"runtime"
	"fmt"
)

var DefaultChain = NewChain()

type Chain struct {
	wrappers []JobWrapper
}

func (c Chain) Then(j Job) Job{
	tol := len(c.wrappers)
	for i, _ := range c.wrappers {
		j = c.wrappers[tol - i - 1](j)
	}
	return j
}

func NewChain(w ... JobWrapper) Chain {
	return Chain{
		wrappers : w,
	}
}

type JobWrapper func(Job)Job

func Recover(logger Logger) JobWrapper{
	return func (j Job) Job{
		return FuncJob(func() {
			defer func () {
				if r := recover(); r != nil {
					buf := make([]byte, 64 << 10)
					buf = buf[:runtime.Stack(buf, false)]
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					logger.Error(err, "panic", "stack", "...\n" + string(buf))
				}
			}()
			j.Run()
		})
	}
}
