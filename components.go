// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"log"
	"math"
)

type InputComponent interface {
	Update(*GameObject, sf.Event)
}

type PlayerInputEuler struct{}

const (
	WALK_ACCEL = 120 // pixels per second
	MAX_VEL    = 120
)

func (pi *PlayerInputEuler) Update(g *GameObject, e sf.Event) {
	switch ev := e.(type) {
	case sf.EventKeyPressed:
		switch ev.Code {
		case sf.KeyLeft:
			g.XVel = -WALK_ACCEL
		case sf.KeyRight:
			g.XVel = WALK_ACCEL
		case sf.KeyUp:
			g.YVel = -WALK_ACCEL
		case sf.KeyDown:
			g.YVel = WALK_ACCEL
		}
	case sf.EventKeyReleased:
		switch ev.Code {
		case sf.KeyLeft, sf.KeyRight:
			g.prXVel = g.XVel
			g.prYVel = g.YVel
			g.XVel = 0
		case sf.KeyUp, sf.KeyDown:
			g.prYVel = g.YVel
			g.prXVel = g.XVel
			g.YVel = 0
		}
	}
}

type MovementComponent interface {
	Update(*GameObject, *Map)
}

type BaseMovePlayer struct{}

func (s *BaseMovePlayer) Update(g *GameObject, m *Map) {
	delta := float32(GetTaskManager().ElpsTime().Seconds())
	v := sf.Vector2f{delta * (g.XVel + delta*g.XAccel*0.5), delta * (g.YVel + delta*g.YAccel*0.5)}
	g.XVel += delta * g.XAccel
	g.YVel += delta * g.YAccel

	if math.Abs(float64(g.XVel)) < 0.25 {
		g.XVel = 0
	} else if math.Abs(float64(g.XVel)) > MAX_VEL {
		if g.XVel > 0 {
			g.XVel = MAX_VEL
		} else if g.XVel < 0 {
			g.XVel = -MAX_VEL
		}
	}

	g.Move(v)
}

type MovePlayerOnMap struct {
	BaseMovePlayer
}

func (s *MovePlayerOnMap) Update(g *GameObject, m *Map) {
	s.BaseMovePlayer.Update(g, m)

	if m != nil {
		def := sf.DefaultRenderStates()
		tr := g.GetTransform()
		def.Transform.Combine(&tr)
		sprBounds := def.Transform.TransformRect(g.Spr.currAnim.GetBounds())
		for _, o := range m.Collidables {
			// log.Println(sprBounds, o)
			if t, rect := sprBounds.Intersects(o); t {
				log.Println(g.XVel, g.YVel, sprBounds, o, rect)
				if g.XVel != 0 {
					if rect.Left > sprBounds.Left {
						// collision to the right of sprite
						g.SetPosition(sf.Vector2f{(rect.Left - sprBounds.Width) + sprBounds.Width/2, sprBounds.Top + sprBounds.Height/2})
						// g.Move(sf.Vector2f{-rect.Width, 0})
					} else if rect.Left+rect.Width > sprBounds.Left {
						// collision to the left of sprite
						g.SetPosition(sf.Vector2f{(rect.Left + rect.Width) + sprBounds.Width/2, sprBounds.Top + sprBounds.Height/2})
					}
				} else if g.YVel != 0 {
					if rect.Top > sprBounds.Top {
						// collision below sprite
						g.SetPosition(sf.Vector2f{})
						g.Move(sf.Vector2f{0, -rect.Height})
					} else if rect.Top+rect.Height > sprBounds.Top {
						// collision above
						g.Move(sf.Vector2f{0, rect.Height})
					}
				}
			}
		}
	}
}

type GraphicsComponent interface {
	Draw(*GameObject, sf.RenderTarget, sf.RenderStates)
}

type SpriteDraw struct{}

func (s *SpriteDraw) Draw(g *GameObject, target sf.RenderTarget, states sf.RenderStates) {
	if g.XVel > 0 {
		g.Spr.SetAnim(WALK_RIGHT)
	} else if g.XVel < 0 {
		g.Spr.SetAnim(WALK_LEFT)
	} else if g.YVel > 0 {
		g.Spr.SetAnim(WALK_DOWN)
	} else if g.YVel < 0 {
		g.Spr.SetAnim(WALK_UP)
	} else if g.prXVel > 0 {
		g.Spr.SetAnim(STAND_RIGHT)
	} else if g.prXVel < 0 {
		g.Spr.SetAnim(STAND_LEFT)
	} else if g.prYVel > 0 {
		g.Spr.SetAnim(STAND_DOWN)
	} else if g.prYVel < 0 {
		g.Spr.SetAnim(STAND_UP)
	}
	g.Spr.currAnim.Advance()

	t := g.GetTransform()
	states.Transform.Combine(&t)
	target.Draw(g.Spr, states)
}
