// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
)

type GameObject struct {
	*sf.Transformable
	XVel, YVel     float32
	prXVel, prYVel float32
	XAccel, YAccel float32
	Spr            *SpriteObj

	InComp InputComponent
	MvComp MovementComponent
	GrComp GraphicsComponent
}

func NewGameObj(sp *SpriteObj, ic InputComponent, mv MovementComponent, gr GraphicsComponent) *GameObject {
	return &GameObject{sf.NewTransformable(), 0, 0, 0, 0, 0, 0, sp, ic, mv, gr}
}

func (g *GameObject) Draw(target sf.RenderTarget, renderStates sf.RenderStates) {
	g.GrComp.Draw(g, target, renderStates)
}
