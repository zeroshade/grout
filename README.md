Grout
=====

### 2D Game Engine using Go and SFML
---

**Requires**

- SFML2 Go library via 'go get bitbucket.org/krepa098/gosfml2'
- Tiled Map Editor (http://www.mapeditor.org) for producing xml tilemap definitions
- darkFunction Editor (http://www.darkfunction.com) for producing sprite sheets and animations

*Tiled and the darkFunction Editor aren't necessarily required, they just happened to be what I used when created the assets and thus used their formats for reading stuff in. Anything using the same formats will work.*

---

It is a task-based game engine that I'm now adding entity/component architecture for handling game objects.

---

A basic main using the engine would be:

```go
func main() {
  tm := grout.GetTaskManager()
  initalState := InitialGameStateObject{}
  grout.InitialGameState(&initialState)
  tm.Execute()
}
```

Where the InitialGameStateObject has the following functions defined on it

- OnPause() *when the state is pushed down on the stack*
- OnResume(w \*sf.RenderWindow) *when the state is put back on top of the stack, given the render window for any setup*
- Init(w \*sf.RenderWindow) *when the engine first puts the state on the stack, this is called*
- Update() (grout.GameState, bool) *called once per frame for updating state, return a new gamestate to push onto the stack or nil, return bool to say whether this current state should be popped off*
- Draw(w \*sf.RenderWindow) *called once per frame to draw whatever is needed by the state*

---
