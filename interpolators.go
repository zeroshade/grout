// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	"time"
)

func interpolatorUpdater(pri int) *listTask {
	return &listTask{BasicTask: NewBasicTask(pri), f: func(l ListItem) {
		it := l.(Interpolator)
		if !it.IsFrozen() {
			it.Update(time.Since(timerUpdate.t).Seconds() * 1000.0)
		}
	}}
}

// Interpolators are designed so that they can update a value based on
// the elapsed time between calls. If they are frozen they remain but will
// not be updated, they could also be killed and then removed.
//
// IsFrozen() should handle whether or not the Interpolator is frozen
// Update(dT float64) will be called with dT being the duration since the last
//                    call to update
// Kill() being called should result in IsAlive() returning false otherwise IsAlive()
//        should return true
type Interpolator interface {
	IsFrozen() bool
	Update(dT float64)
	Kill()
	IsAlive() bool
	GetValue() float64
}

type timeBasedInterpolator struct {
	frozen           bool
	elpTime, totTime float64
	alive            bool
	val              float64
}

func (t *timeBasedInterpolator) Kill()             { t.alive = false }
func (t *timeBasedInterpolator) IsAlive() bool     { return t.alive }
func (t *timeBasedInterpolator) Freeze()           { t.frozen = true }
func (t *timeBasedInterpolator) Thaw()             { t.frozen = false }
func (t *timeBasedInterpolator) GetValue() float64 { return t.val }

type LinearTimeInterpolator struct {
	timeBasedInterpolator
	startVal, endVal float64
}

// LinearTimeInterpolator
//
// It linearly goes from the start value to the end value over timespan milliseconds
func NewLinearTimeInterpolator(timespan, start, end float64) *LinearTimeInterpolator {
	return &LinearTimeInterpolator{timeBasedInterpolator{false, 0, timespan, true, 0}, start, end}
}

func clamp(min, max, val float64) float64 {
	if val < min {
		return min
	} else if val > max {
		return max
	} else {
		return val
	}
}

func (l *LinearTimeInterpolator) Update(dT float64) {
	l.elpTime += dT

	b := clamp(0, 1, l.elpTime/l.totTime)
	l.val = l.startVal*(1-b) + l.endVal*b

	if l.elpTime > l.totTime {
		l.Kill()
	}
}

type QuadraticTimeInterpolator struct {
	timeBasedInterpolator
	startVal, endVal, midVal float64
}

func NewQuadraticTimeInterpolator(timespan, start, end, mid float64) *QuadraticTimeInterpolator {
	return &QuadraticTimeInterpolator{timeBasedInterpolator{false, 0, timespan, true, 0}, start, end, mid}
}

func (q *QuadraticTimeInterpolator) Update(dT float64) {
	q.elpTime += dT

	b := clamp(0, 1, q.elpTime/q.totTime)
	a := 1 - b
	q.val = q.startVal*a*a + q.midVal*2*a*b + q.endVal*b*b

	if q.elpTime > q.totTime {
		q.Kill()
	}
}

type CubicTimeInterpolator struct {
	timeBasedInterpolator
	startVal, endVal, midVal1, midVal2 float64
}

func NewCubicTimeInterpolator(timespan, start, end, mid1, mid2 float64) *CubicTimeInterpolator {
	return &CubicTimeInterpolator{timeBasedInterpolator{false, 0, timespan, true, 0}, start, end, mid1, mid2}
}

func (c *CubicTimeInterpolator) Update(dT float64) {
	c.elpTime += dT

	b := clamp(0, 1, c.elpTime/c.totTime)
	a := 1 - b
	c.val = c.startVal*a*a*a + c.midVal1*3*a*a*b + c.midVal2*3*a*b*b + c.endVal*b*b*b

	if c.elpTime > c.totTime {
		c.Kill()
	}
}
