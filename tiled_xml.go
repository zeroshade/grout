// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	FLIPPED_HORIZONTALLY_FLAG uint = 0x80000000
	FLIPPED_VERTICALLY_FLAG   uint = 0x40000000
	FLIPPED_DIAGONALLY_FLAG   uint = 0x20000000
)

func LoadMapInfo(file string) (*Map, error) {
	xmlFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()
	m := new(Map)
	err = xml.NewDecoder(xmlFile).Decode(m)
	if err != nil {
		return nil, err
	}

	for _, og := range m.Objects {
		if og.Name != "Collision" {
			continue
		}
		m.Collidables = make([]sf.FloatRect, len(og.Objs))
		for _, r := range og.Objs {
			m.Collidables = append(m.Collidables, sf.FloatRect{r.X, r.Y, r.W, r.H})
		}
	}

	return m, nil
}

type Map struct {
	XMLName     xml.Name    `xml:"map"`
	Ver         string      `xml:"version,attr"`
	Ori         string      `xml:"orientation,attr"`
	Width       uint        `xml:"width,attr"`
	Height      uint        `xml:"height,attr"`
	TileWidth   uint        `xml:"tilewidth,attr"`
	TileHeight  uint        `xml:"tileheight,attr"`
	TSets       []*TileSet  `xml:"tileset"`
	Layers      []*Layer    `xml:"layer"`
	Objects     []*ObjGroup `xml:"objectgroup"`
	Collidables []sf.FloatRect
	TSprites    []*sf.Sprite
	drawTop     bool
	TData       map[uint]map[string]string
}

type ObjGroup struct {
	XMLName xml.Name  `xml:"objectgroup"`
	Name    string    `xml:"name,attr"`
	Width   uint      `xml:"width,attr"`
	Height  uint      `xml:"height,attr"`
	Objs    []*Object `xml:"object"`
}

type Object struct {
	XMLName xml.Name `xml:"object"`
	X       float32  `xml:"x,attr"`
	Y       float32  `xml:"y,attr"`
	W       float32  `xml:"width,attr"`
	H       float32  `xml:"height,attr"`
}

func (m *Map) OnlyTop() {
	m.drawTop = true
}

func (m *Map) NoTop() {
	m.drawTop = false
}

func (m *Map) LoadImageData() (err error) {
	m.drawTop = false
	for _, ts := range m.TSets {
		numWide := uint(math.Floor(float64(ts.Image.Width) / float64(ts.TileWidth)))
		numHigh := uint(math.Floor(float64(ts.Image.Height) / float64(ts.TileHeight)))
		ts.LGid = numWide*numHigh + ts.FGid - 1
		ts.Texture, err = sf.NewTextureFromFile(GetTaskManager().GetSettings().Paths.Res+"/"+ts.Image.Src, nil)
		if err != nil {
			return
		}
		ts.Texture.SetSmooth(true)
		if len(m.TSprites) <= int(ts.LGid) {
			t := make([]*sf.Sprite, ts.LGid+1)
			copy(t, m.TSprites[0:])
			m.TSprites = t
		}
		for x := uint(0); x < numWide; x++ {
			for y := uint(0); y < numHigh; y++ {
				srcY := int(y * ts.TileHeight)
				srcX := int(x * ts.TileWidth)
				gid := ts.FGid + x + (y * numWide)

				if m.TSprites[gid], err = sf.NewSprite(ts.Texture); err != nil {
					return err
				}
				m.TSprites[gid].SetTextureRect(sf.IntRect{srcX, srcY, int(ts.TileWidth), int(ts.TileHeight)})
			}
		}
		if len(ts.TileInfo) > 0 {
			log.Println("TIle Info")
			if m.TData == nil {
				m.TData = make(map[uint]map[string]string)
			}
			for _, ti := range ts.TileInfo {
				if _, ok := m.TData[ti.Gid+1]; !ok {
					m.TData[ti.Gid+1] = make(map[string]string)
				}
				for _, v := range ti.Props {
					log.Println("Tile:", ti.Gid+1, "Name:", v.Name, "Val:", v.Value)
					m.TData[ti.Gid+1][v.Name] = v.Value
				}
			}
		}
	}

	return
}

func (m *Map) Draw(target sf.RenderTarget, renderStates sf.RenderStates) {
	for _, layer := range m.Layers {
		if m.drawTop && layer.Name != "Top" {
			continue
		} else if !m.drawTop && layer.Name == "Top" {
			continue
		}
		v := target.GetView()
		sz := v.GetSize()
		ce := v.GetCenter()
		startX := uint(ce.X-sz.X/2) / m.TileWidth
		startY := uint(ce.Y-sz.Y/2) / m.TileHeight
		endX := uint(ce.X+sz.X/2)/m.TileWidth + 1
		endY := uint(ce.Y+sz.Y/2)/m.TileHeight + 1

		startX = uint(math.Max(math.Min(float64(layer.Width), float64(startX)), 0))
		startY = uint(math.Max(math.Min(float64(layer.Height), float64(startY)), 0))
		endX = uint(math.Max(math.Min(float64(layer.Width), float64(endX)), 0))
		endY = uint(math.Max(math.Min(float64(layer.Height), float64(endY)), 0))

		for y := startY; y < endY; y++ {
			for x := startX; x < endX; x++ {
				coord := x + (y * layer.Width)
				tile := layer.Data.Tiles[coord]
				if tile.Gid > uint(0) {
					s := m.TSprites[tile.Gid]
					r := s.GetTextureRect()
					s.SetPosition(sf.Vector2f{X: float32(x * m.TileWidth), Y: float32(y * m.TileHeight)})
					if tile.FlipHoriz && tile.FlipVert {
						// flip diagonally
						s.SetTextureRect(sf.IntRect{r.Left + r.Width, r.Top + r.Height, -r.Width, -r.Height})
					} else if tile.FlipDiag && tile.FlipHoriz {
						s.SetOrigin(sf.Vector2f{X: 0, Y: float32(r.Height)})
						s.SetRotation(90)
					} else if tile.FlipDiag && tile.FlipVert {
						s.SetOrigin(sf.Vector2f{X: float32(r.Width), Y: 0})
						s.SetRotation(270)
					} else if tile.FlipHoriz {
						// flip horizontally
						s.SetTextureRect(sf.IntRect{r.Left + r.Width, r.Top, -r.Width, r.Height})
					} else if tile.FlipVert {
						// flip vertically
						s.SetTextureRect(sf.IntRect{r.Left, r.Top + r.Height, r.Width, -r.Height})
					}
					target.Draw(s, renderStates)
					s.SetTextureRect(r)
					s.SetRotation(0)
					s.SetOrigin(sf.Vector2f{X: 0, Y: 0})
				}
			}
		}
	}
	if GetTaskManager().GetSettings().Debug.ShowSprBound {
		for _, o := range m.Collidables {
			rs, _ := sf.NewRectangleShape()
			rs.SetSize(sf.Vector2f{o.Width, o.Height})
			rs.SetPosition(sf.Vector2f{o.Left, o.Top})
			rs.SetOutlineThickness(1)
			rs.SetOutlineColor(sf.ColorBlack())
			rs.SetFillColor(sf.ColorTransparent())

			target.Draw(rs, sf.DefaultRenderStates())
		}
	}
}

type TileSet struct {
	XMLName    xml.Name   `xml:"tileset"`
	FGid       uint       `xml:"firstgid,attr"`
	Name       string     `xml:"name,attr"`
	TileWidth  uint       `xml:"tilewidth,attr"`
	TileHeight uint       `xml:"tileheight,attr"`
	Image      *ImgInfo   `xml:"image"`
	TileInfo   []TileInfo `xml:"tile"`
	LGid       uint
	Texture    *sf.Texture
}

type TileInfo struct {
	XMLName xml.Name   `xml:"tile"`
	Gid     uint       `xml:"id,attr"`
	Props   []Property `xml:"properties>property"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type ImgInfo struct {
	XMLName xml.Name `xml:"image"`
	Src     string   `xml:"source,attr"`
	Width   uint     `xml:"width,attr"`
	Height  uint     `xml:"height,attr"`
}

type Layer struct {
	XMLName xml.Name `xml:"layer"`
	Name    string   `xml:"name,attr"`
	Width   uint     `xml:"width,attr"`
	Height  uint     `xml:"height,attr"`
	Data    Data     `xml:"data"`
}

type Data struct {
	XMLName xml.Name `xml:"data"`
	Tiles   []*Tile  `xml:"tile"`
}

func (d *Data) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if len(start.Attr) > 0 && start.Attr[0].Name.Local == "encoding" {
		if start.Attr[0].Value == "csv" {
			var csvData string
			if err := dec.DecodeElement(&csvData, &start); err != nil {
				return err
			}
			csvData = strings.Replace(csvData, "\n", "", -1)
			rdr := csv.NewReader(strings.NewReader(csvData))
			rec, err := rdr.Read()
			if err != nil && err != io.EOF {
				return err
			}
			for _, s := range rec {
				g, _ := strconv.Atoi(s)
				d.Tiles = append(d.Tiles, NewTile(uint(g)))
			}
		}
	} else {
		for {
			token, _ := dec.Token()
			if token == nil {
				return errors.New("WTF")
			}
			switch se := token.(type) {
			case xml.StartElement:
				if se.Name.Local == "tile" {
					t := new(Tile)
					dec.DecodeElement(t, &se)
					d.Tiles = append(d.Tiles, t)
				}
			case xml.EndElement:
				return nil
			}
		}
	}
	return nil
}

func NewTile(gid uint) *Tile {
	flipH := (gid & FLIPPED_HORIZONTALLY_FLAG) != 0
	flipV := (gid & FLIPPED_VERTICALLY_FLAG) != 0
	flipD := (gid & FLIPPED_DIAGONALLY_FLAG) != 0

	gid &= ^(FLIPPED_HORIZONTALLY_FLAG | FLIPPED_VERTICALLY_FLAG | FLIPPED_DIAGONALLY_FLAG)

	return &Tile{Gid: gid, FlipHoriz: flipH, FlipVert: flipV, FlipDiag: flipD}
}

type Tile struct {
	XMLName   xml.Name `xml:"tile"`
	Gid       uint     `xml:"gid,attr"`
	FlipHoriz bool
	FlipVert  bool
	FlipDiag  bool
}

func (t *Tile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(t, &start); err != nil {
		return err
	}
	t = NewTile(t.Gid)
	return nil
}
