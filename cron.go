package cron

import (
	"time"
	"sync"
	"context"
	"strconv"
	"strings"
)

var maxID = 100000000

// cron the total timer main structure
// control all timer obj
type Cron struct {
	entries EntryHeap
	stop chan struct{}
	add chan *Entry
	remove chan int
	snapshot chan chan []*Entry
	running bool
	cronMu sync.Mutex
	wg sync.WaitGroup
	nextID int
	chain Chain
	cache map[int]*Entry
	logger Logger
}

func New(opts ...Option) *Cron{
	c := &Cron{
		entries : make(EntryHeap, 0),
		stop : make(chan struct{}),
		add : make(chan *Entry),
		remove : make(chan int),
		snapshot : make (chan chan []*Entry),
		running : false,
		chain : DefaultChain,
		logger : DefaultLogger,
		cache : make(map[int]*Entry),
	}

	for _, opt := range opts {
		opt(c)
	}
	c.logger.Info("new")
	return c
}

// Start set cron running true and loop begin
func (c *Cron) Start() {
	c.cronMu.Lock()
	defer c.cronMu.Unlock()
	if c.running {
		return
	}
	c.running = true
	go c.loop()
	c.logger.Info("start")
}

func (c *Cron) loop() {
	for {
		now := time.Now()
		var timer *time.Timer
		if len(c.entries) == 0 {
			timer = time.NewTimer(1000 * time.Hour)
		} else {
			e := c.entries.Top()
			delta := e.Next.Sub(now)
			timer = time.NewTimer(delta)
		}

		for {
			select {
			case now = <-timer.C:
				c.logger.Info("timeout", "now", now)
				for {
					e := c.entries.Top()
					if e == nil || e.Next.After(now) {
						break
					}
					e = c.entries.Pop()
					c.processJob(e)
				}
			case <-c.stop:
				timer.Stop()
				c.logger.Info("stop")
				return
			case e := <-c.add:
				timer.Stop()
				c.cache[e.ID] = e
				c.entries.Push(e)
				c.logger.Info("add", "new entry", e)
			case id := <-c.remove:
				if e, ok := c.cache[id]; ok {
					e.IsCancel = true
					c.logger.Info("remove", "entryID", id)
				}
				continue
			case che := <-c.snapshot:
				entries := c.entrySnapshot()
				che <- entries
				c.logger.Info("snapshot", "total", len(entries))
				continue
			}

			break
		}
	}
}

// processJob 
func (c *Cron) processJob(e *Entry){
	if e.IsCancel {
		delete(c.cache, e.ID)
		return
	}

	if e.IsFreq {
		e.GenNextTime()
		c.entries.Push(e)
	} else {
		delete(c.cache, e.ID)
	}

	c.logger.Info("run", "entryID", e.ID, "next", e.Next)
	c.startJob(e.WrappedJob)
}

// startJob goroutine run job
func (c *Cron) startJob(j Job){
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		j.Run()
	}()
}

// entrySnapshot get copy of entries 
func (c *Cron) entrySnapshot() []*Entry{
	var entries []*Entry
	for _, e := range c.entries {
		entries = append(entries, e)
	}

	return entries
}

func (c *Cron) addJob(j Job, isFreq bool, schedule Schedule) int {
	c.cronMu.Lock()
	defer c.cronMu.Unlock()
	if !c.running {
		return -1
	}
	c.nextID++
	e := &Entry{
		ID : c.nextID,
		IsCancel : false,
		IsFreq : isFreq,
		Schedule : schedule,
		Job : j,
		WrappedJob : c.chain.Then(j),
	}

	e.Next = schedule.Next(e.Next)
	c.add <- e
	return c.nextID
}

// CallOut delay seconds call f() 
// if running = false,  do nothing
func (c *Cron) CallOut(delay int, f func()) int {
	sch, _ := NewSchedule(delay)
	return c.addJob(FuncJob(f), false, sch)
}

func (c *Cron)CallFre(freqency int, f func()) int{
	sch, _ := NewSchedule(freqency)
	return c.addJob(FuncJob(f), true, sch)
}

func (c *Cron) Daily(hour int, min int, f func()) int {
	sch, err := NewSchedule(hour, min)
	if err != nil {
		return -1
	}
	return c.addJob(FuncJob(f), true, sch)
}

func (c *Cron) Weekly(wday int, hour int, min int, f func()) int {
	sch, err := NewSchedule(hour, min, wday)
	if err != nil {
		return -1
	}
	return c.addJob(FuncJob(f), true, sch)
}

// RemoveCallOut remove id job
func (c *Cron) RemoveCallOut(id int){
	c.cronMu.Lock()
	defer c.cronMu.Unlock()
	if !c.running {
		return
	}
	c.remove <- id
}

// Entries get cron all job copy
func (c *Cron) Entries() []*Entry{
	c.cronMu.Lock()
	defer c.cronMu.Unlock()
	if !c.running {
		return nil
	}

	var che = make(chan []*Entry)
	c.snapshot <- che
	return <-che
}

// Stop stop cron 
func (c *Cron) Stop() context.Context{
	c.cronMu.Lock()
	defer c.cronMu.Unlock()
	if c.running {
		c.running = false
		c.stop <- struct{}{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func(){
		c.wg.Wait()
		cancel()
	}()
	return ctx
}

// Job represent timer job
type Job interface {
	Run()
}

// FuncJob func as a job
type FuncJob func()

func (f FuncJob) Run() {
	f()
}

// EntryHeap store all timer sorted by heap
type EntryHeap []*Entry

// Top get min entry not remove from list
func (eh EntryHeap) Top() *Entry {
	if len(eh) <= 0 {
		return nil
	}
	return eh[0]
}

// Pop get min entry and remove from list
func (eh *EntryHeap) Pop() *Entry {
	total := len(*eh)
	if total <= 0 {
		return nil
	}
	entry := (*eh)[0]
	(*eh)[0], (*eh)[total - 1] = (*eh)[total - 1], nil
	*eh = (*eh)[:total - 1]

	eh.DownAdjustHeap(0)
	return entry
}

// Insert entry into list
func (eh *EntryHeap) Push(entry *Entry) {
	*eh = append(*eh, entry)

	total := len(*eh)
	eh.UpAdjustHeap(total - 1)
}

// DownAdjustHeap Adjust heap down to child
func (eh EntryHeap) DownAdjustHeap(index int) {
	total := len(eh)
	if index * 2 + 1 >= total {
		return
	}
	var minChild =  2 * index + 1
	if minChild + 1 < total &&  eh[minChild + 1].Less(eh[minChild]) {
		minChild++
	}

	if eh[minChild].Less(eh[index]) {
		eh[index], eh[minChild] = eh[minChild], eh[index]
		eh.DownAdjustHeap(minChild)
	}
}

// UpAdjustHeap Adjust heap up to father
func (eh EntryHeap) UpAdjustHeap(index int) {
	if index <= 0 {
		return
	}

	father := (index - 1) / 2
	if eh[index].Less(eh[father]) {
		eh[index], eh[father] = eh[father], eh[index]
		eh.UpAdjustHeap(father)
	}
}

// Entry conrtain job info
type Entry struct {
	ID int
	IsCancel bool
	IsFreq bool
	Next time.Time
	Schedule Schedule
	WrappedJob Job
	Job Job
}

func (e *Entry) Less (o *Entry) bool{
	return e.Next.Before(o.Next)
}

// GenNextTime gen next call time 
// just freq call effect
func (e *Entry) GenNextTime() {
	e.Next = e.Schedule.Next(time.Now())
}

func (e *Entry) String() string {
	var buf strings.Builder
	buf.WriteString("ID:")
	buf.WriteString(strconv.Itoa(e.ID))
	buf.WriteString(" unix:")
	buf.WriteString(strconv.FormatInt(e.Next.Unix(), 10))
	return buf.String()
}
