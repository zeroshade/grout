// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
)

type gameStateTask struct {
	BasicTask
	stk gameStateStack
}

func (gs *gameStateTask) Start() bool {
	gs.stk = make(gameStateStack, 1)
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
	state, pop := gs.stk.Top().Update(vidUpdate.win)
	if pop {
		gs.stk.Pop()
		if state == nil && len(gs.stk) != 0 {
			gs.stk.Top().OnResume(vidUpdate.win)
		}
	}
	if state != nil {
		if len(gs.stk) != 0 {
			gs.stk.Top().OnPause()
		}
		gs.stk.Push(state)
		gs.stk.Top().Init(vidUpdate.win)
	}
	if len(gs.stk) == 0 {
		gs.SetCanKill(true)
	}
}

// Interface to define individual game state behavior  
type GameState interface {
	Init(*sf.RenderWindow)
	Update(*sf.RenderWindow) (GameState, bool)
	OnPause()
	OnResume(*sf.RenderWindow)
}

type gameStateStack []GameState

func (s *gameStateStack) Push(gs GameState) {
	*s = append(*s, gs)
}

func (s *gameStateStack) Pop() GameState {
	ret := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return ret
}

func (s *gameStateStack) Top() GameState {
	return (*s)[len(*s)-1]
}
