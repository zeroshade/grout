// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"container/list"
	"log"
	"time"
)

type Task interface {
	Start() bool
	OnSuspend()
	Update()
	Draw(*sf.RenderWindow)
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
func (b *BasicTask) CanKill() bool         { return b.canKill }
func (b *BasicTask) SetCanKill(k bool)     { b.canKill = k }
func (b *BasicTask) GetPriority() int      { return b.priority }
func (b *BasicTask) OnSuspend()            {}
func (b *BasicTask) OnResume()             {}
func (b *BasicTask) Draw(*sf.RenderWindow) {}

type SimpleTask struct {
	BasicTask
	updateFunc func()
	drawFunc   func(*sf.RenderWindow)
}

func (d *SimpleTask) Start() bool { return true }
func (d *SimpleTask) Stop()       {}
func (d *SimpleTask) Update() {
	if d.updateFunc != nil {
		d.updateFunc()
	}
}
func (d *SimpleTask) Draw(w *sf.RenderWindow) {
	if d.drawFunc != nil {
		d.drawFunc(w)
	}
}

func NewSimpleTask(p int, updateFunc func(), drawFunc func(*sf.RenderWindow)) *SimpleTask {
	return &SimpleTask{NewBasicTask(p), updateFunc, drawFunc}
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

type inputTask struct {
	BasicTask
	evQue *list.List
}

func (i *inputTask) Start() bool {
	i.evQue = new(list.List)
	return i.evQue != nil
}

func (i *inputTask) Stop() {
	i.evQue = nil
}

func (i *inputTask) Update() {
	// win := GetTaskManager().getWindow()

	// i.evQue.Init()
	// for event := win.PollEvent(); event != nil; event = win.PollEvent() {
	// 	switch ev := event.(type) {
	// 	case sf.EventClosed:
	// 		GetTaskManager().KillAllTasks()
	// 	case sf.EventResized:
	// 		v := GetTaskManager().getWindow().GetView()
	// 		v.Reset(sf.FloatRect{0, 0, float32(ev.Width), float32(ev.Height)})
	// 		GetTaskManager().getWindow().SetView(v)
	// 	default:
	// 		i.evQue.PushBack(event)
	// 	}
	// }

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
