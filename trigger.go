// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

func triggerUpdater(pri int) *listTask {
	return &listTask{BasicTask: NewBasicTask(pri), f: func(l ListItem) {
		it := l.(Trigger)
		it.Tick()
	}}
}

type Trigger interface {
	Kill()
	Tick()
	IsAlive() bool
}

type BaseTrigger struct {
	t         func() bool
	h         func()
	bFireOnce bool
	alive     bool
}

func (b *BaseTrigger) Kill()         { b.alive = false }
func (b *BaseTrigger) IsAlive() bool { return b.alive }
func (b *BaseTrigger) Tick() {
	if b.t() {
		b.h()
		if b.bFireOnce {
			b.Kill()
		}
	}
}
