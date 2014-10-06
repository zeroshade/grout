// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
)

type GameObject struct {
	*sf.Transformable
	Vel      sf.Vector2f
	prVel    sf.Vector2f
	Accel    sf.Vector2f
	Spr      *SpriteObj
	AniState SpriteState

	InComp InputComponent
	MvComp MovementComponent
	GrComp GraphicsComponent

	onGround bool
}

func NewGameObj(sp *SpriteObj, ic InputComponent, mv MovementComponent, gr GraphicsComponent) *GameObject {
	return &GameObject{sf.NewTransformable(), sf.Vector2f{}, sf.Vector2f{}, sf.Vector2f{}, sp, STAND_RIGHT, ic, mv, gr, false}
}

func (g *GameObject) Draw(target sf.RenderTarget, renderStates sf.RenderStates) {
	g.GrComp.Draw(g, target, renderStates)
}
