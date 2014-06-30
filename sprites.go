package grout

import (
	sf "bitbucket.org/krepa098/gosfml2"
	"encoding/xml"
	"errors"
	"log"
	"os"
)

func LoadAnimation(filename string) (*AnimMap, error) {
	c := GetTaskManager().GetSettings()
	sprpath := c.Paths.Res + "/" + c.Paths.Spr + "/"

	file, err := os.Open(sprpath + filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	animInfo := &DFEAnimations{}
	if err = xml.NewDecoder(file).Decode(animInfo); err != nil {
		log.Fatal(err)
		return nil, err
	}

	file.Close()
	if file, err = os.Open(sprpath + animInfo.SheetFileName); err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err = xml.NewDecoder(file).Decode(&animInfo.Sheet); err != nil {
		log.Fatal(err)
		return nil, err
	}

	animInfo.Sheet.Texture, err = sf.NewTextureFromFile(sprpath+animInfo.Sheet.Img, nil)
	for _, v := range animInfo.Sheet.Defs.Defs {
		if v.Spr, err = sf.NewSprite(animInfo.Sheet.Texture); err != nil {
			return nil, err
		}
		v.Spr.SetTextureRect(sf.IntRect{v.X, v.Y, v.W, v.H})
	}

	am := make(AnimMap)
	for _, a := range animInfo.Anims {
		for _, c := range a.Cells {
			var cell AniCell
			cell.Spr = animInfo.Sheet.Defs.Defs[c.Spr.ImgName].Spr
			cell.RenderState = sf.DefaultRenderStates()
			cell.RenderState.Transform.Translate(c.Spr.XOff, c.Spr.YOff)
			cell.Delay = c.Delay
			anim := am[a.Name]
			anim.Cells = append(anim.Cells, cell)
			am[a.Name] = anim
		}
	}

	return &am, nil
}

type AnimMap map[string]Animation

type Animation struct {
	CurrIndex int
	Position  sf.Vector2f
	Cells     []AniCell
}

type AniCell struct {
	Spr         *sf.Sprite
	RenderState sf.RenderStates
	Delay       int
}

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
