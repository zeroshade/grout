// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"encoding/xml"
	"errors"
	"log"
	"os"
)

type DFEAnimations struct {
	XMLName       xml.Name       `xml:"animations"`
	SheetFileName string         `xml:"spriteSheet,attr"`
	Sheet         DFESpriteSheet `xml:"-"`
	Anims         []*DFEAnim     `xml:"anim"`
}

type DFEAnim struct {
	XMLName xml.Name   `xml:"anim"`
	Name    string     `xml:"name,attr"`
	Cells   []*DFECell `xml:"cell"`
}

type DFECell struct {
	XMLName xml.Name  `xml:"cell"`
	Delay   int       `xml:"delay,attr"`
	Spr     DFESprite `xml:"spr"`
}

type DFESprite struct {
	XMLName xml.Name `xml:"spr"`
	ImgName string   `xml:"name,attr"`
	XOff    float32  `xml:"x,attr"`
	YOff    float32  `xml:"y,attr"`
	Z       int      `xml:"z,attr"`
}

type DFESpriteSheet struct {
	XMLName xml.Name `xml:"img"`
	Img     string   `xml:"name,attr"`
	W       int      `xml:"w,attr"`
	H       int      `xml:"h,attr"`
	Defs    DFEDefs  `xml:"definitions"`
	Texture *sf.Texture
}

type DFEDefs struct {
	Defs map[string]*DFESpr
}

func (df *DFEDefs) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	df.Defs = make(map[string]*DFESpr)
	var dir string
	for {
		token, err := d.Token()
		if token == nil || err != nil {
			return errors.New("WTF")
		}
		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "dir" {
				dir = se.Attr[0].Value
			} else if se.Name.Local == "spr" {
				spr := new(DFESpr)
				d.DecodeElement(spr, &se)
				df.Defs[dir+spr.Name] = spr
			}
		case xml.EndElement:
			if se.Name.Local == "definitions" {
				return nil
			}
		}
	}
}

type DFESpr struct {
	XMLName string `xml:"spr"`
	Name    string `xml:"name,attr"`
	X       int    `xml:"x,attr"`
	Y       int    `xml:"y,attr"`
	W       int    `xml:"w,attr"`
	H       int    `xml:"h,attr"`
	Spr     *sf.Sprite
}

func NewSpriteObj() *SpriteObj {
	return &SpriteObj{Transformable: sf.NewTransformable()}
}

type SpriteObj struct {
	*sf.Transformable
	Animations AnimMap
	currAnim   *Animation
	XVel, YVel float32
}

func (s *SpriteObj) SetAnim(name string) {
	s.currAnim = s.Animations[name]
}

func (s *SpriteObj) Update() {
	if s.XVel != 0 || s.YVel != 0 {
		delta := float32(GetTaskManager().ElpsTime().Seconds())
		s.Move(sf.Vector2f{s.XVel * delta, s.YVel * delta})
		s.currAnim.Advance()
	} else {
		s.currAnim.Reset()
	}
}

func (s *SpriteObj) Draw(target sf.RenderTarget, renderStates sf.RenderStates) {
	t := s.GetTransform()
	renderStates.Transform.Combine(&t)
	target.Draw(s.currAnim, renderStates)
}

type AnimMap map[string]*Animation

type Animation struct {
	currIndex int
	cells     []AniCell
	fc        int
}

func (s *SpriteObj) LoadAnimations(filename string) error {
	c := GetTaskManager().GetSettings()
	sprpath := c.Paths.Res + "/" + c.Paths.Spr + "/"

	file, err := os.Open(sprpath + filename)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	animInfo := &DFEAnimations{}
	if err = xml.NewDecoder(file).Decode(animInfo); err != nil {
		log.Fatal(err)
		return err
	}

	file.Close()
	if file, err = os.Open(sprpath + animInfo.SheetFileName); err != nil {
		log.Fatal(err)
		return err
	}

	if err = xml.NewDecoder(file).Decode(&animInfo.Sheet); err != nil {
		log.Fatal(err)
		return err
	}

	animInfo.Sheet.Texture, err = sf.NewTextureFromFile(sprpath+animInfo.Sheet.Img, nil)
	for _, v := range animInfo.Sheet.Defs.Defs {
		if v.Spr, err = sf.NewSprite(animInfo.Sheet.Texture); err != nil {
			return err
		}
		v.Spr.SetTextureRect(sf.IntRect{v.X, v.Y, v.W, v.H})
		v.Spr.SetOrigin(sf.Vector2f{float32(v.W / 2), float32(v.H / 2)})
		// v.Spr.SetScale(sf.Vector2f{2, 2})
	}

	if s.Animations == nil {
		s.Animations = make(AnimMap)
	}
	for _, a := range animInfo.Anims {
		for _, c := range a.Cells {
			anim := s.Animations[a.Name]
			if anim == nil {
				anim = &Animation{}
			}
			var cell AniCell
			cell.Spr = animInfo.Sheet.Defs.Defs[c.Spr.ImgName].Spr
			cell.RenderState = sf.DefaultRenderStates()
			cell.RenderState.Transform.Translate(c.Spr.XOff, c.Spr.YOff)
			cell.Delay = c.Delay
			anim.cells = append(anim.cells, cell)
			s.Animations[a.Name] = anim
		}
	}

	return nil
}

func (a *Animation) Reset() {
	a.currIndex = len(a.cells) - 1
	a.fc = 0
}

func (a *Animation) Advance() {
	if a.fc > a.cells[a.currIndex].Delay {
		if a.currIndex == len(a.cells)-1 {
			a.currIndex = 0
		} else {
			a.currIndex++
		}
		a.fc = 0
	}
	a.fc++
}

func (a *Animation) Draw(target sf.RenderTarget, renderStates sf.RenderStates) {
	// t := a.GetTransform()
	// t.Combine(&a.cells[a.currIndex].RenderState.Transform)	
	renderStates.Transform.Combine(&a.cells[a.currIndex].RenderState.Transform)
	gb := renderStates.Transform.TransformRect(a.cells[a.currIndex].Spr.GetGlobalBounds())

	// log.Println(renderStates.Transform)
	target.Draw(a.cells[a.currIndex].Spr, renderStates)

	if GetTaskManager().GetSettings().Debug.ShowSprBound {
		rs, _ := sf.NewRectangleShape()
		rs.SetSize(sf.Vector2f{gb.Width, gb.Height})
		rs.SetPosition(sf.Vector2f{gb.Left, gb.Top})
		rs.SetOutlineThickness(1)
		rs.SetOutlineColor(sf.ColorBlack())
		rs.SetFillColor(sf.ColorTransparent())

		target.Draw(rs, sf.DefaultRenderStates())
	}
}

type AniCell struct {
	Spr         *sf.Sprite
	RenderState sf.RenderStates
	Delay       int
}
