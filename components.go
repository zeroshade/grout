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

type NullInputComponent struct{}

func (n *NullInputComponent) Update(*GameObject, sf.Event) {}

type PlayerInputEuler struct{}

const (
	WALK_ACCEL = 1000 // pixels per second
	MAX_VEL    = 400
	GRAVITY    = 700
	DAMPING    = 0.92
	MAX_JUMP   = -350
	JUMP_FORCE = -400
)

func (pi *PlayerInputEuler) Update(g *GameObject, e sf.Event) {
	switch ev := e.(type) {
	case sf.EventKeyPressed:
		switch ev.Code {
		case sf.KeyLeft:
			g.Vel.X = -WALK_ACCEL
		case sf.KeyRight:
			g.Vel.X = WALK_ACCEL
		case sf.KeyUp:
			g.Vel.Y = -WALK_ACCEL
		case sf.KeyDown:
			g.Vel.Y = WALK_ACCEL
		}
	case sf.EventKeyReleased:
		switch ev.Code {
		case sf.KeyLeft, sf.KeyRight:
			g.prVel = g.Vel
			g.Vel.X = 0
		case sf.KeyUp, sf.KeyDown:
			g.prVel = g.Vel
			g.Vel.Y = 0
		}
	}
}

type SideScrollInput struct{}

func (s *SideScrollInput) Update(g *GameObject, e sf.Event) {
	switch ev := e.(type) {
	case sf.EventKeyPressed:
		switch ev.Code {
		case sf.KeyRight:
			g.Accel.X = WALK_ACCEL
			g.AniState = WALK_RIGHT
		case sf.KeyLeft:
			g.Accel.X = -WALK_ACCEL
			g.AniState = WALK_LEFT
		case sf.KeyUp:
			if g.onGround {
				g.Vel.Y = JUMP_FORCE
			}
		}
	case sf.EventKeyReleased:
		switch ev.Code {
		case sf.KeyRight:
			g.Accel.X = 0
			g.AniState = STAND_RIGHT
		case sf.KeyLeft:
			g.AniState = STAND_LEFT
			g.Accel.X = 0
		}
	}
}

type MovementComponent interface {
	Update(*GameObject, *Map)
}

type NullMovementComponent struct{}

func (n *NullMovementComponent) Update(*GameObject, *Map) {}

type SideScrollMove struct{}

func (s *SideScrollMove) Update(g *GameObject, m *Map) {
	delta := float32(GetTaskManager().ElpsTime().Seconds())
	if delta > 0.5 {
		delta = 0.5
	}
	g.prVel = g.Vel

	grav := sf.Vector2f{0, GRAVITY * delta}
	accel := g.Accel.TimesScalar(delta)
	// log.Println(accel)
	// log.Println("Accel:", accel)
	// log.Println("Grav:", grav)
	g.Vel = g.Vel.Plus(grav)
	// log.Println("X:", g.Vel.X)
	g.Vel.X = g.Vel.X * DAMPING
	// log.Println("DampX: ", g.Vel.X)
	g.Vel = g.Vel.Plus(accel)

	g.Vel.X = clamp(-MAX_VEL, MAX_VEL, g.Vel.X)
	if math.Abs(float64(g.Vel.X)) < 1 {
		g.Vel.X = 0
	}
	g.Vel.Y = clamp(MAX_JUMP, GRAVITY, g.Vel.Y)
	if math.Abs(float64(g.Vel.Y)) < 1 {
		g.Vel.Y = 0
	}
	// log.Println("Vel: ", g.Vel)
	toMove := g.Vel.TimesScalar(delta)
	// log.Println("ToMove: ", toMove)

	def := sf.DefaultRenderStates()
	tr := g.GetTransform()
	def.Transform.Combine(&tr)
	sprBounds := def.Transform.TransformRect(g.Spr.currAnim.GetBounds())

	var tx int
	var ty int
	tx = int((sprBounds.Left + sprBounds.Width/2) / float32(m.TileWidth))

	if toMove.Y > 0 {
		ty = int((sprBounds.Top + sprBounds.Height - 5) / float32(m.TileHeight))
	} else {
		ty = int((sprBounds.Top) / float32(m.TileHeight))
	}

	onSlope := false
	gids := make([]int, 9)
	for i := 0; i < 9; i++ {
		c := int(i % 3)
		r := int(i / 3)
		t := (tx + (c - 1)) + ((ty + (r - 1)) * int(m.Layers[0].Width))
		if t < 0 {
			continue
		}
		// log.Println(tx, ty, c, r, t)
		gids[i] = t
	}

	gids = append(gids[:4], gids[5:]...)
	gids = append(gids, 0)
	copy(gids[7:], gids[6:])
	gids[6] = gids[2]
	gids = append(gids[:2], gids[3:]...)
	gids[4], gids[6] = gids[6], gids[4]
	gids[0], gids[4] = gids[4], gids[0]

	transPos := sf.TransformIdentity()
	transPos.Translate(toMove.X, toMove.Y)

	curTile := tx + (ty * int(m.Layers[0].Width))
	tileRect := sf.FloatRect{float32(tx * int(m.TileWidth)), float32(ty * int(m.TileHeight)), float32(m.TileWidth), float32(m.TileHeight)}
	if props, ok := m.TData[m.Layers[0].Data.Tiles[curTile].Gid]; ok {
		if props["slope"] == "1" {
			onSlope = true
			// log.Println("TileRect", tileRect)
			bounds := transPos.TransformRect(sprBounds)
			// log.Println("Sprite Boundary", bounds, sprBounds)
			// log.Println("Sloper")
			st := (bounds.Left + (bounds.Width / 2) - tileRect.Left) / tileRect.Width
			fy := (1-st)*(tileRect.Top+tileRect.Height) + st*tileRect.Top
			sprBot := bounds.Top + bounds.Height
			// log.Println(fy, sprBot)
			if sprBot >= fy {
				transPos.Translate(0, -(sprBot - fy))
				g.Vel.Y = 0
			} else if toMove.X < 0 && toMove.Y > 0 {
				xSpeed := float32(math.Abs(float64(toMove.X / 2)))
				// log.Println(xSpeed, toMove.Y)
				if toMove.Y < xSpeed {
					transPos.Translate(0, xSpeed)
				}
			}
		}
	}
	g.onGround = onSlope
	for idx, t := range gids {
		if t >= len(m.Layers[0].Data.Tiles) || t < 0 {
			log.Println("WTF? ", t)
			continue
		}
		ty := uint(t / int(m.Layers[0].Width))
		tx := uint(t % int(m.Layers[0].Width))
		gid := m.Layers[0].Data.Tiles[t].Gid
		if gid > 0 {
			tileRect := sf.FloatRect{float32(tx * m.TileWidth), float32(ty * m.TileHeight), float32(m.TileWidth), float32(m.TileHeight)}
			desiredPos := transPos.TransformRect(sprBounds)
			props, ok := m.TData[gid]
			if isCollide, intersection := desiredPos.Intersects(tileRect); isCollide {
				trslt := sf.Vector2f{0, 0}
				switch idx {
				case 0:
					// tile is below
					if !ok || props["slope"] == "-1" {
						trslt.Y = -intersection.Height
						g.Vel.Y = 0
						g.onGround = true
					}
				case 1:
					// tile is directly above
					trslt.Y = intersection.Height
					g.Vel.Y = 0
				case 2:
					// tile is left
					if !ok || props["slope"] == "1" {
						trslt.X = intersection.Width
					}
				case 3:
					// tile is right
					if (!ok || props["slope"] == "-1") && !onSlope {
						trslt.X = -intersection.Width
					}
				default:
					if ok && (props["slope"] == "1" || props["slope"] == "-1") {
						break
					}
					if intersection.Width > intersection.Height {
						// tile is diagonal, but resolve vertically
						g.Vel.Y = 0
						if idx > 5 {
							trslt.Y = -intersection.Height
						} else {
							trslt.Y = intersection.Height
						}
					} else {
						// tile is diagonal but resolve horizontally
						if idx == 6 || idx == 4 {
							trslt.X = intersection.Width
						} else {
							trslt.X = -intersection.Width
						}
					}
				}
				transPos.Translate(trslt.X, trslt.Y)
			}
		}
	}
	// log.Println(tx, ty)

	g.SetPosition(transPos.TransformPoint(g.GetPosition()))
}

type BaseMovePlayer struct{}

func (s *BaseMovePlayer) Update(g *GameObject, m *Map) {
	delta := float32(GetTaskManager().ElpsTime().Seconds())
	v := sf.Vector2f{delta * (g.Vel.X + delta*g.Accel.X*0.5), delta * (g.Vel.Y + delta*g.Accel.Y*0.5)}
	g.Vel = g.Vel.Plus(g.Accel).TimesScalar(delta)

	if math.Abs(float64(g.Vel.X)) < 0.25 {
		g.Vel.X = 0
	} else if math.Abs(float64(g.Vel.X)) > MAX_VEL {
		if g.Vel.X > 0 {
			g.Vel.X = MAX_VEL
		} else if g.Vel.X < 0 {
			g.Vel.X = -MAX_VEL
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
				log.Println(g.Vel.X, g.Vel.Y, sprBounds, o, rect)
				if g.Vel.X != 0 {
					if rect.Left > sprBounds.Left {
						// collision to the right of sprite
						g.SetPosition(sf.Vector2f{(rect.Left - sprBounds.Width) + sprBounds.Width/2, sprBounds.Top + sprBounds.Height/2})
						// g.Move(sf.Vector2f{-rect.Width, 0})
					} else if rect.Left+rect.Width > sprBounds.Left {
						// collision to the left of sprite
						g.SetPosition(sf.Vector2f{(rect.Left + rect.Width) + sprBounds.Width/2, sprBounds.Top + sprBounds.Height/2})
					}
				} else if g.Vel.Y != 0 {
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

type NullGraphics struct{}

func (n *NullGraphics) Draw(g *GameObject, target sf.RenderTarget, render sf.RenderStates) {
	g.Spr.SetAnim(g.AniState)
	g.Spr.currAnim.Advance()

	t := g.GetTransform()
	render.Transform.Combine(&t)
	target.Draw(g.Spr, render)
}

type SpriteDraw struct{}

func (s *SpriteDraw) Draw(g *GameObject, target sf.RenderTarget, states sf.RenderStates) {
	if g.Vel.X > 0 {
		g.Spr.SetAnim(WALK_RIGHT)
	} else if g.Vel.X < 0 {
		g.Spr.SetAnim(WALK_LEFT)
	} else if g.Vel.Y > 0 {
		g.Spr.SetAnim(WALK_DOWN)
	} else if g.Vel.Y < 0 {
		g.Spr.SetAnim(WALK_UP)
	} else if g.prVel.X > 0 {
		g.Spr.SetAnim(STAND_RIGHT)
	} else if g.prVel.X < 0 {
		g.Spr.SetAnim(STAND_LEFT)
	} else if g.prVel.Y > 0 {
		g.Spr.SetAnim(STAND_DOWN)
	} else if g.prVel.Y < 0 {
		g.Spr.SetAnim(STAND_UP)
	}
	g.Spr.currAnim.Advance()

	t := g.GetTransform()
	states.Transform.Combine(&t)
	target.Draw(g.Spr, states)
}
