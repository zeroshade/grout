// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"container/list"
	"sort"
	"time"
)

type TaskManager interface {
	Execute()
	AddTask(t Task) bool
	SuspendTask(t Task)
	RemoveTask(t Task)
	ResumeTask(t Task)
	KillAllTasks()
	ElpsTime() time.Duration

	GetSettings() *Config
	getWindow() *sf.RenderWindow
	GetEventQueue() *list.List
}

type taskList []Task

func (slice taskList) Len() int {
	return len(slice)
}

func (slice taskList) Less(i, j int) bool {
	return slice[i].GetPriority() < slice[j].GetPriority()
}

func (slice taskList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice taskList) Find(t Task) int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i].GetPriority() >= t.GetPriority() })
	if i < len(slice) && slice[i].GetPriority() == t.GetPriority() {
		for k := i; k < len(slice); k++ {
			if slice[k] == t {
				return k
			}
		}
	}
	return -1
}

func (slice *taskList) Remove(i int) {
	s := *slice
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	s = s[:len(s)-1]
	*slice = s
}

type taskMgr struct {
	taskList       taskList
	pausedTaskList taskList
	conf           Config
	prev           time.Time
	win            *sf.RenderWindow
	w              uint
	h              uint
	ticker         *time.Ticker
}

func newTaskMgr() *taskMgr {
	t := &taskMgr{taskList: make([]Task, 0), pausedTaskList: make([]Task, 0)}
	if err := loadSettings(&t.conf); err != nil {
		panic("Failed to load settings")
	}

	t.w = t.conf.Video.W
	t.h = t.conf.Video.H
	t.win = sf.NewRenderWindow(sf.VideoMode{t.w, t.h, 32}, "Testing", sf.StyleDefault, sf.DefaultContextSettings())
	t.ticker = time.NewTicker(time.Second / t.conf.Video.FPS)
	return t
}

var (
	_This         TaskManager    = newTaskMgr()
	vidUpdate     *SimpleTask    = &SimpleTask{NewBasicTask(1000), nil, func(w *sf.RenderWindow) { w.Display() }}
	stateUpdate   *gameStateTask = &gameStateTask{BasicTask: NewBasicTask(500)}
	fpsUpdate     *fpsTask       = &fpsTask{BasicTask: NewBasicTask(1), p: _This.GetSettings().Debug.PrintFPS}
	timerUpdate   *timerTask     = &timerTask{BasicTask: NewBasicTask(2)}
	interUpdate   *listTask      = interpolatorUpdater(3)
	triggerUpdate *listTask      = triggerUpdater(4)
	inputUpdate   *inputTask     = &inputTask{NewBasicTask(5), nil}
)

func RegisterTrigger(t Trigger) {
	triggerUpdate.list = append(triggerUpdate.list, t)
}

func RegisterInterpolator(i Interpolator) {
	interUpdate.list = append(interUpdate.list, i)
}

func init() {
	_This.AddTask(vidUpdate)
	_This.AddTask(stateUpdate)
	_This.AddTask(fpsUpdate)
	_This.AddTask(timerUpdate)
	_This.AddTask(interUpdate)
	_This.AddTask(inputUpdate)
}

func InitialGameState(g GameState) {
	stateUpdate.stk[0] = g
	stateUpdate.stk[0].(GameState).Init(_This.getWindow())
}

func GetTaskManager() TaskManager {
	return _This
}

func (tm *taskMgr) getWindow() *sf.RenderWindow {
	return tm.win
}

func (tm *taskMgr) GetEventQueue() *list.List {
	return inputUpdate.evQue
}

func (tm *taskMgr) GetSettings() *Config {
	return &tm.conf
}

func (tm *taskMgr) AddTask(t Task) bool {
	if !t.Start() {
		return false
	}

	tm.taskList = append(tm.taskList, t)
	sort.Sort(tm.taskList)

	return true
}

func (tm *taskMgr) SuspendTask(t Task) {
	if i := tm.taskList.Find(t); i != -1 {
		t.OnSuspend()
		tm.taskList.Remove(i)
		tm.pausedTaskList = append(tm.pausedTaskList, t)
		sort.Sort(tm.pausedTaskList)
	}
}

func (tm *taskMgr) ResumeTask(t Task) {
	if i := tm.pausedTaskList.Find(t); i != -1 {
		t.OnResume()
		tm.pausedTaskList.Remove(i)
		tm.taskList = append(tm.taskList, t)
		sort.Sort(tm.taskList)
	}
}

func (tm *taskMgr) RemoveTask(t Task) {
	if i := tm.taskList.Find(t); i != -1 {
		t.SetCanKill(true)
	}
}

func (tm *taskMgr) KillAllTasks() {
	for _, t := range tm.taskList {
		t.SetCanKill(true)
	}
}

func (tm *taskMgr) ElpsTime() time.Duration { return time.Since(tm.prev) }

func (tm *taskMgr) Execute() {
	defer tm.win.Close()
	for len(tm.taskList) > 0 {
		select {
		case <-tm.ticker.C:
			tm.prev = timerUpdate.t
			for _, t := range tm.taskList {
				if !t.CanKill() {
					t.Update()
				}
			}
			for _, t := range tm.taskList {
				if !t.CanKill() {
					t.Draw(tm.win)
				}
			}
			for i := 0; i < len(tm.taskList); i++ {
				if tm.taskList[i].CanKill() {
					tm.taskList[i].Stop()
					tm.taskList.Remove(i)
					i--
				}
			}
		}
	}
}
