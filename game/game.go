package game

import (
	//"encoding/base64"
	"encoding/json"
	"github.com/fractalbach/fractalnet/cellular/wave"
	"log"
)

var (
	GAME_WORLD_WIDTH  = 48
	GAME_WORLD_HEIGHT = 48
)

type World struct {
	Ents   map[int]*Ent
	nextid int
	h      int //height (number of tiles in y direction)
	w      int //width  (number of tiles in x direction)
	// Trees  *BoolGrid
	LifeGrid *wave.Life
}

type Ent struct {
	Name     string
	Type     string
	Location []float64
}

type Location struct {
	X int
	Y int
}

type GameState struct {
	State map[int]*Ent
}

type TreeState struct {
	Trees string
}

// ______________________________________________________
// 		Creating a new World
// ------------------------------------------------------

func MakeNewWorld() *World {
	return &World{
		Ents:     map[int]*Ent{},
		nextid:   1,
		h:        GAME_WORLD_HEIGHT,
		w:        GAME_WORLD_WIDTH,
		LifeGrid: wave.NewLife(GAME_WORLD_WIDTH, GAME_WORLD_HEIGHT),
		//Trees:  CreateRandomInitialTrees(48, 48),
	}
}

// ______________________________________________________
// 		Manipulating the World
// ------------------------------------------------------

func (w *World) MapHeight() int {
	return w.h
}

func (w *World) MapWidth() int {
	return w.w
}

func (w *World) makeNextId() int {
	output := w.nextid
	w.nextid++
	return output
}

func (w *World) generatePlayer(username string) (int, bool) {
	id := w.makeNextId()
	ok := w.addEntity(&Ent{Name: username, Type: "player"}, id)
	if ok {
		return id, true
	}
	return 0, false
}

func (w *World) changeEntityLocation(id int, l []float64) bool {
	if _, ok := w.Ents[id]; ok {
		w.Ents[id].Location = l
		return true
	}
	return false
}

func (w *World) addEntity(e *Ent, id int) bool {
	if _, ok := w.Ents[id]; ok {
		return false
	}
	w.Ents[id] = e
	return true
}

func (w *World) deleteEntity(id int) bool {
	if id == 0 {
		return false
	}
	if _, ok := w.Ents[id]; ok {
		delete(w.Ents, id)
		return true
	}
	return false
}

// ______________________________________________________
// 		Message & Event Handler
// ------------------------------------------------------

func (w *World) DoGameEvent(a *AbstractEvent) interface{} {
	switch a.EventType {

	case "LifeState":
		return w.LifeGrid.LifeStateMessage()

	case "LifeUpdate":
		w.LifeGrid.Step()
		return true

	case "LifeChange":
		w.LifeGrid.AlterAt(a.Location.X, a.Location.Y, a.Value)
		return true

	case "GameState":
		return w.stateAllEntities()

	case "Move":
		return w.changeEntityLocation(a.SourceId, a.Position)

	case "Login":
		id, _ := w.generatePlayer(a.EventBody)
		return id

	case "Logout":
		return w.deleteEntity(a.TargetId)

		/*
			case "ToggleTree":
				newVal := false
				if a.EventBody == "on" {
					newVal = true
				}
				x, y := a.Location.X, a.Location.Y
				w.Trees.Set(x, y, newVal)
				return true

			case "UpdateTrees":
				w.Trees.NextGeneration()
				return true

			case "TreeState":
				return w.stateAllTrees()
		*/

	case "Create":
	case "Delete":

	}
	return false
}

/*
func (w *World) DoAdminEvent(a *AbstractEvent) interface{} {
	switch a.EventType {
	case "Something Special":
	case "Create World in 7 days":
	case "Begin SunToRedGiant Expansion Protocol":
	}
	return false
}
*/
// ______________________________________________________
// 		Game State
// ------------------------------------------------------

func (w *World) stateAllEntities() []byte {
	b, err := json.Marshal(GameState{w.Ents})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return b
}

/*
func (w *World) stateAllTrees() []byte {
	t, _, err := CompressBoolGrid(w.Trees.grid)
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	b64 := base64.StdEncoding.EncodeToString(t)
	msg, err := json.Marshal(TreeState{b64})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return msg
}
*/
// ______________________________________________________
//  The Game PRAM
// ------------------------------------------------------
// 	NOTE: 	For the Future.  Will help to separate
// 			game events from the connection handling,
//			and things like the chat room.
// ------------------------------------------------------

type GamePram struct {
	w         *World
	eventchan chan *AbstractEvent
}

func NewGamePram() *GamePram {
	storedpram := &GamePram{
		w:         MakeNewWorld(),
		eventchan: make(chan *AbstractEvent),
	}
	log.Println("The World is now running in the Game PRAM...")
	go storedpram.run()
	return storedpram
}

func (g *GamePram) run() {
	for {
		select {
		case e := <-g.eventchan:
			r := g.w.DoGameEvent(e)
			if e.HasResponseChannel() {
				e.Response <- r
			}

			/*		case e := <-g.adminchan:
					r := g.w.DoAdminEvent(e)
					if e.HasResponseChannel() {
						e.Response <- r
					}*/
		}
	}
}

func (g *GamePram) RequestTreeState() []byte {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: "TreeState",
		Response:  r,
	}
	g.eventchan <- event
	a := <-r
	output, ok := a.([]byte)
	if ok {
		return output
	}
	return []byte{}
}

func (g *GamePram) RequestLifeState() []byte {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: "LifeState",
		Response:  r,
	}
	g.eventchan <- event
	a := <-r
	if output, ok := a.([]byte); ok {
		return output
	}
	return []byte{}
}

func (g *GamePram) ToggleTreeEvent(x, y int, newVal bool) []byte {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: "ToggleTree",
		Location:  Location{x, y},
		Response:  r,
	}
	if newVal {
		event.EventBody = "on"
	}
	g.eventchan <- event
	a := <-r
	output, ok := a.([]byte)
	if ok {
		return output
	}
	return []byte{}
}

func (g *GamePram) RequestGameState() []byte {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: "GameState",
		Response:  r,
	}
	g.eventchan <- event
	a := <-r
	output, ok := a.([]byte)
	if ok {
		return output
	}
	return []byte{}
}

// LoginEvent returns playerId; If playerId returns 0, Login failed!
func (g *GamePram) LoginEvent(username string) int {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: "Login",
		EventBody: username,
		Response:  r,
	}
	g.eventchan <- event
	a := <-r
	output, ok := a.(int)
	if ok {
		return output
	}
	return 0
}

func (g *GamePram) LogoutEvent(playerId int) {
	event := &AbstractEvent{
		EventType: "Logout",
		TargetId:  playerId,
	}
	g.eventchan <- event
}

func (g *GamePram) UpdateTreesEvent() {
	event := &AbstractEvent{
		EventType: "UpdateTrees",
	}
	g.eventchan <- event
}

func (g *GamePram) UpdateLifeEvent() {
	event := &AbstractEvent{
		EventType: "LifeUpdate",
	}
	g.eventchan <- event
}

func (g *GamePram) CustomPlayerEvent(event *AbstractEvent) {
	g.eventchan <- event
}
