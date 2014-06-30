// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"log"
	"time"
)

type Task interface {
	Start() bool
	OnSuspend()
	Update()
	OnResume()
	Stop()

	CanKill() bool
	SetCanKill(k bool)
	GetPriority() int
}

type BasicTask struct {
	canKill  bool
	priority int
}

func NewBasicTask(p int) BasicTask {
	return BasicTask{false, p}
}
func (b *BasicTask) CanKill() bool     { return b.canKill }
func (b *BasicTask) SetCanKill(k bool) { b.canKill = k }
func (b *BasicTask) GetPriority() int  { return b.priority }
func (b *BasicTask) OnSuspend()        {}
func (b *BasicTask) OnResume()         {}

type SimpleTask struct {
	BasicTask
	updateFunc func()
}

func (d *SimpleTask) Start() bool { return true }
func (d *SimpleTask) Stop()       {}
func (d *SimpleTask) Update()     { d.updateFunc() }

func NewSimpleTask(p int, updateFunc func()) *SimpleTask {
	return &SimpleTask{NewBasicTask(p), updateFunc}
}

type fpsTask struct {
	BasicTask
	c int
	t time.Time
	p bool
}

func (f *fpsTask) Start() bool {
	f.t = time.Now()
	f.c = 0
	return true
}
func (f *fpsTask) Stop() {}
func (f *fpsTask) Update() {
	f.c++

	if f.p && (f.c%100 == 0) {
		d := time.Since(f.t)
		log.Printf("FPS: %f\n", float64(f.c)/d.Seconds())
	}
}

type timerTask struct {
	BasicTask
	t time.Time
}

func (t *timerTask) Stop()   {}
func (t *timerTask) Update() { t.t = time.Now() }
func (t *timerTask) Start() bool {
	t.t = time.Now()
	return true
}

type videoTask struct {
	BasicTask
	win    *sf.RenderWindow
	ticker *time.Ticker
	w      uint
	h      uint
}

func (v *videoTask) Start() bool {
	v.win = sf.NewRenderWindow(sf.VideoMode{v.w, v.h, 32}, "Testing", sf.StyleDefault, sf.DefaultContextSettings())
	v.ticker = time.NewTicker(time.Second / GetTaskManager().GetSettings().Video.FPS)
	return v.win != nil && v.ticker != nil
}

func (v *videoTask) Update() {
	select {
	case <-v.ticker.C:
		for event := v.win.PollEvent(); event != nil; event = v.win.PollEvent() {
			switch ev := event.(type) {
			case sf.EventKeyReleased:
				switch ev.Code {
				case sf.KeyEscape:
					// v.win.Close()
					// v.SetCanKill(true)
				}
			case sf.EventClosed:
				v.win.Close()
				v.SetCanKill(true)
			}
		}
	}
	v.win.Display()
}

func (v *videoTask) Stop() {
	if v.win.IsOpen() {
		v.win.Close()
	}
	GetTaskManager().KillAllTasks()
}

type ListItem interface {
	IsAlive() bool
}

type listTask struct {
	BasicTask
	list []ListItem
	f    func(ListItem)
}

func (l *listTask) Stop() { l.list = nil }
func (l *listTask) Start() bool {
	l.list = make([]ListItem, 0)
	return true
}
func (l *listTask) Update() {
	for j := 0; j < len(l.list); j++ {
		it := l.list[j]
		if it.IsAlive() {
			l.f(it)
		} else {
			l.list = append(l.list[:j], l.list[j+1:]...)
			j--
		}
	}
}
