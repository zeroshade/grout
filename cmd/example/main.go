package main

import (
	sf "bitbucket.org/krepa098/gosfml2"
	eng "github.com/zeroshade/grout"
	"log"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

type MainMenu struct {
	rect  *sf.RectangleShape
	stuff *sf.Text
	font  *sf.Font
}

func (m *MainMenu) Init(w *sf.RenderWindow) {
	m.rect, _ = sf.NewRectangleShape()
	m.rect.SetSize(sf.Vector2f{80, 80})
	m.rect.SetOutlineThickness(3)
	m.rect.SetOutlineColor(sf.ColorBlack())
	m.rect.SetFillColor(sf.Color{0, 153, 204, 255})
	m.rect.SetOrigin(sf.Vector2f{40, 40})
	m.rect.SetPosition(sf.Vector2f{100, 30})

	m.font, _ = sf.NewFontFromFile("resources/DroidSans.ttf")

	m.stuff, _ = sf.NewText(m.font)
	m.stuff.SetPosition(sf.Vector2f{110, 40})
	m.stuff.SetString("Hello World")
	m.stuff.SetColor(sf.ColorBlack())

}

func (m *MainMenu) OnPause() {}
func (m *MainMenu) OnResume(w *sf.RenderWindow) {
	w.SetView(w.GetDefaultView())
}

func (m *MainMenu) Draw(w *sf.RenderWindow) {
	w.Clear(sf.ColorRed())
	w.Draw(m.rect, sf.DefaultRenderStates())
	w.Draw(m.stuff, sf.DefaultRenderStates())
}

func (m *MainMenu) Update() (eng.GameState, bool) {

	evQ := eng.GetTaskManager().GetEventQueue()
	for e := evQ.Front(); e != nil; e = e.Next() {
		switch ev := e.Value.(type) {
		case sf.EventKeyPressed:
			switch ev.Code {
			case sf.KeyQ:
				ms, err := NewMapScroll("resources/tutcsv.tmx")
				if err != nil {
					log.Fatal(err)
				}
				return ms, false
			case sf.KeyEscape:
				return nil, true
			}
		}
	}

	return nil, false
}

type MapScroll struct {
	v     *sf.View
	m     *eng.Map
	crono *eng.GameObject
	circ  *sf.CircleShape
}

func (m *MapScroll) OnPause()                    {}
func (m *MapScroll) OnResume(w *sf.RenderWindow) { w.SetView(m.v) }

func (m *MapScroll) Init(w *sf.RenderWindow) {
	m.v = w.GetView()
	// m.v.SetCenter(sf.Vector2f{float32(m.m.Width*m.m.TileWidth) / 2, float32(m.m.Height*m.m.TileHeight) / 2})
	// m.v.Move(sf.Vector2f{30, 40})
	w.SetView(m.v)

	crono := eng.NewSpriteObj()
	if err := crono.LoadAnimations("crono.anim"); err != nil {
		log.Fatal(err)
	}
	crono.SetAnim(eng.STAND_RIGHT)
	m.crono = eng.NewGameObj(crono, &eng.PlayerInputEuler{}, &eng.MovePlayerOnMap{}, &eng.SpriteDraw{})
	m.crono.SetPosition(sf.Vector2f{40, 90})
	m.crono.XVel = 0

	m.circ, _ = sf.NewCircleShape()
	m.circ.SetRadius(2)
	m.circ.SetFillColor(sf.ColorBlack())
	m.circ.SetOrigin(sf.Vector2f{1, 1})
	m.circ.SetPosition(sf.Vector2f{40, 38})
}

func (m *MapScroll) Update() (eng.GameState, bool) {
	// m.v.Move(sf.Vector2f{newX, newY})

	evQ := eng.GetTaskManager().GetEventQueue()
	for e := evQ.Front(); e != nil; e = e.Next() {
		switch ev := e.Value.(type) {
		case sf.EventKeyPressed:
			switch ev.Code {
			case sf.KeyEscape:
				return nil, true
			default:
				m.crono.InComp.Update(m.crono, e.Value.(sf.Event))
			}
		default:
			m.crono.InComp.Update(m.crono, e.Value.(sf.Event))
		}
	}

	m.crono.MvComp.Update(m.crono, m.m)

	return nil, false
}

func (m *MapScroll) Draw(w *sf.RenderWindow) {
	w.Clear(sf.ColorBlack())
	w.SetView(m.v)

	m.m.NoTop()
	w.Draw(m.m, sf.DefaultRenderStates())
	w.Draw(m.crono, sf.DefaultRenderStates())
	m.m.OnlyTop()
	w.Draw(m.m, sf.DefaultRenderStates())
	w.Draw(m.circ, sf.DefaultRenderStates())
}

func NewMapScroll(f string) (ms *MapScroll, err error) {
	ms = new(MapScroll)
	ms.m, err = eng.LoadMapInfo(f)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = ms.m.LoadImageData()
	if err != nil {
		log.Fatal(err)
	}

	return
}

func main() {

	tm := eng.GetTaskManager()
	mm := MainMenu{}
	eng.InitialGameState(&mm)

	tm.Execute()

}
