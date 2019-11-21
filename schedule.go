package cron

import (
	"time"
	"errors"
)

const (
	Minute = 60
	Hour = Minute * 60
	Day = Hour * 24
	Week = Day * 7
)

type Schedule interface {
	Next (time.Time) time.Time
}

type ConstDelaySchedule struct {
	Delay time.Duration
}

func (s ConstDelaySchedule) Next(t time.Time) time.Time {
	if t.IsZero() {
		t = time.Now()
	}
	return t.Add(s.Delay -  time.Duration(t.Nanosecond()))
}

type DailySchedule struct {
	ConstDelaySchedule
	Hour int
	Min int
}

func (s DailySchedule) Next(t time.Time) time.Time {
	if t.IsZero() {
		hour := s.Hour
		min := s.Min
		wday :=  -1
		t := time.Now()
		return t.Add(Round(GetDiff(hour, min, wday), Day) -  time.Duration(t.Nanosecond()))
	}
	return t.Add(s.Delay -  time.Duration(t.Nanosecond()))
}

type WeeklySchedule struct {
	DailySchedule
	Wday int
}

func (s WeeklySchedule) Next(t time.Time) time.Time {
	if t.IsZero() {
		hour := s.Hour
		min := s.Min
		wday := s.Wday
		t := time.Now()
		return t.Add(time.Duration(Round(GetDiff(hour, min, wday), Week)) -  time.Duration(t.Nanosecond()))
	}
	return t.Add(s.Delay -  time.Duration(t.Nanosecond()))
}

func GetDiff(hour int, min int, wday int) int{
	t := time.Now()
	var weekSeconds = func (hour int, min int, sec int, wday int) int{
		return sec + min * Minute + hour * Hour + wday * Day
	}
	delta1 := weekSeconds(t.Hour(), t.Minute(), t.Second(), int(t.Weekday()))
	if wday == -1 {
		wday = int(t.Weekday())
	}
	delta2 := weekSeconds(hour, min, 0, wday)
	return delta2 - delta1
}

func Round(x int, mod int) time.Duration{
	if x > 0 {
		return time.Duration(x) * time.Second
	} else {
		return time.Duration(x + mod) * time.Second
	}
}

func NewSchedule(arg ...int) (Schedule, error){
	tol := len(arg)

	var cdf = func (delay int) (ConstDelaySchedule, error) {
		var d = time.Duration(delay) * time.Second
		return ConstDelaySchedule {d}, nil
	}

	var df = func(delay int) (DailySchedule, error){
		f, err := cdf(delay)
		if err != nil {
			return DailySchedule{}, err
		}

		hour := arg[0]
		min := arg[1]
		if hour < 0 || hour > 23 {
			return DailySchedule{}, errors.New("invalid hour")
		} else if  min < 0  || min > 59 {
			return DailySchedule{}, errors.New("invalid min")
		}
		return DailySchedule{
			ConstDelaySchedule : f,
			Hour : hour,
			Min : min,
		}, nil
	}

	var wf = func(delay int) (WeeklySchedule, error) {
		f, err := df(delay)
		if err != nil {
			return WeeklySchedule{}, err
		}

		wday := arg[2]
		if wday < 0 || wday > 6 {
			return WeeklySchedule{}, errors.New("invalid wday")
		}
		return WeeklySchedule{
			DailySchedule : f,
			Wday : wday,
		}, nil
	}

	if tol == 1 {
		return cdf(arg[0])
	} else if tol == 2 {
		return df(Day)
	} else if tol == 3 {
		return wf(Week)
	} else {
		return nil, errors.New("invalid arg")
	}
}
