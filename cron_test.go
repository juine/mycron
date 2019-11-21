package cron

import (
	"fmt"
	"time"
	"math/rand"
	"errors"

	"testing"
)


func TestEntryHeap(t *testing.T){
	var eh EntryHeap
	e := Entry{
		IsCancel : false,
		IsFreq : false,
	}

	total := 0
	now := time.Now()
	rand.Seed(time.Now().Unix())
	for i := 0; i < total; i ++ {
		var e1 = e
		e1.ID = i
		e1.Next = now.Add( time.Second * time.Duration(rand.Intn(1000)))
		eh.Push(&e1)
	}
	fmt.Println(eh)
	fmt.Println("---------")
	for i := 0; i < total; i ++ {
		fmt.Println(eh.Pop())
	}
}

func TestLogger(t *testing.T){
	printlog := func (logger Logger) {
		logger.Info("game start", "userid", "1234", "age", 18)
		err := errors.New("new errro")
		logger.Error(err, "game start", "userid", "1234", "age", 18)
		logger.Warning("game start", "userid", "2234", "age", 218)
	}

	var logger = DefaultLogger
	printlog(logger)
	logger = NewLogger(map[string]string{})
	printlog(logger)
}

type stu struct {
	id int
	name  string
}

func (s stu) fun() {
	fmt.Println(s.id, s.name)
}

func TestCron(t *testing.T){
	logger := NewLogger(map[string]string{})
	c := New(WithLogger(logger),
			WithChain(Recover(logger)))
	c.Start()
	defer c.Stop()
	f := stu {
		id : 1000,
		name : "juine",
	}
	f1 := f
	f1.id = 1001
	c.CallOut(10, func(){ panic("panic error")})
	timeidx := c.CallFre(2, func(){ f.fun()})
	c.Daily(21, 27, func(){ f1.fun()})

	t1 := time.NewTimer(70 * time.Second)
	t2 := time.NewTimer(7 * time.Second)
	for {
		select {
		case <- t1.C:
			;
		case <- t2.C:
			c.RemoveCallOut(timeidx)
			continue
		}
		break
	}
}

func TestSchedule(t *testing.T) {
	var now time.Time
	s, _ := NewSchedule(23, 31)
	fmt.Println(s.Next(now))

	s, _ = NewSchedule(5, 31, 3)
	fmt.Println(s.Next(now))
}
