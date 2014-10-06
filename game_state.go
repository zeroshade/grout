// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
)

type gameStateTask struct {
	BasicTask
	stk stack
}

func (gs *gameStateTask) Start() bool {
	gs.stk = make(stack, 1)
	return true
}

func (gs *gameStateTask) Stop() {
	for len(gs.stk) != 0 {
		gs.stk.Pop()
	}
	GetTaskManager().KillAllTasks()
}

func (gs *gameStateTask) Update() {
	if len(gs.stk) == 0 {
		return
	}
	state, pop := gs.stk.Top().(GameState).Update()
	if pop {
		gs.stk.Pop()
		if state == nil && len(gs.stk) != 0 {
			gs.stk.Top().(GameState).OnResume(GetTaskManager().getWindow())
		}
	}
	if state != nil {
		if len(gs.stk) != 0 {
			gs.stk.Top().(GameState).OnPause()
		}
		gs.stk.Push(state)
		gs.stk.Top().(GameState).Init(GetTaskManager().getWindow())
	}
	if len(gs.stk) == 0 {
		gs.SetCanKill(true)
	}
}

func (gs *gameStateTask) Draw(w *sf.RenderWindow) {
	if len(gs.stk) == 0 {
		return
	}
	gs.stk.Top().(GameState).Draw(w)
}

// Interface to define individual game state behavior
type GameState interface {
	Init(*sf.RenderWindow)
	Update() (GameState, bool)
	Draw(*sf.RenderWindow)
	OnPause()
	OnResume(*sf.RenderWindow)
}

type stack []interface{}

func (s *stack) Push(i interface{}) {
	*s = append(*s, i)
}

func (s *stack) Pop() interface{} {
	ret := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return ret
}

func (s *stack) Top() interface{} {
	return (*s)[len(*s)-1]
}
