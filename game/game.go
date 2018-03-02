package game

import (
	//"encoding/base64"
	"encoding/json"
	"github.com/fractalbach/fractalnet/cellular/gameofwar"
	"log"
)

var (
	GAME_WORLD_WIDTH  = 48
	GAME_WORLD_HEIGHT = 48
)

type World struct {

	// Ents is for "Entities".  It's a hash map of {ID number: Entity}
	Ents map[int]*Ent

	// War is the instance of the Game Of War corresponding to this World.
	War *gameofwar.GameInstance

	// private variables include the ID counter (nextid) and map dimensions.
	nextid, w, h int
}

type Ent struct {
	Name     string
	Type     string
	Location Location
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
		Ents:   map[int]*Ent{},
		nextid: 1,
		h:      GAME_WORLD_HEIGHT,
		w:      GAME_WORLD_WIDTH,
		War:    gameofwar.NewGameInstance(GAME_WORLD_WIDTH, GAME_WORLD_HEIGHT),
		//Trees:  CreateRandomInitialTrees(48, 48),
		//LifeGrid: wave.NewLife(GAME_WORLD_WIDTH, GAME_WORLD_HEIGHT),
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

func (w *World) changeEntityLocation(id, x, y int) bool {
	if _, ok := w.Ents[id]; ok {
		w.Ents[id].Location.Set(x, y)
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

func (w *World) stateAllEntities() []byte {
	b, err := json.Marshal(GameState{w.Ents})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return b
}

// ______________________________________________________
// 		Message & Event Handler
// ------------------------------------------------------

// DoGameEvent actually executes the functions to the game world.
// Passing AbstractEvent messages to DoGameEvent will check what kind of
// event it is, and if it has the required parameters, and then attempt to
// do that function.
//
// Sometimes events require a channel for data to be sent back.  For those
// Events, it is assumed that a receiving channel is created prior to
// calling DoGameEvent.  If there is a channel included in the AbstractEvent,
// then it can be utilized by this event handler.
//
func (w *World) DoGameEvent(a *AbstractEvent) interface{} {
	switch a.EventType {

	case "LifeState":
		msg := w.War.LifeStateMessage()
		if a.Response != nil {
			a.Response <- msg
			return true
		}

	case "LifeUpdate":
		w.War.LifeUpdate()
		return true

	case "LifeChange":
		//
		// !NOTE! The naming of this Location is confusing.
		// It is actually a "GridLocation" type in messages.go
		//
		w.War.ChangeAt(a.Location.X, a.Location.Y, a.Value)
		return true

	case "LaBomba":
		return w.War.DropBomb(a.Value, a.Location.X, a.Location.Y)

	case "GameState":
		if a.Response != nil {
			a.Response <- w.stateAllEntities()
			return true
		}

	case "Move":
		return w.changeEntityLocation(a.SourceId, a.Location.X, a.Location.Y)

	case "Login":
		if a.Response != nil {
			id, _ := w.generatePlayer(a.EventBody)
			a.Response <- id
			return true
		}

	case "Logout":
		return w.deleteEntity(a.TargetId)

	case "Create":
	case "Delete":

	}
	return false
}

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
		case event := <-g.eventchan:
			g.w.DoGameEvent(event)
		}
	}
}

// RequestSomething helps send game event messages that requires a response.
// It creates a response channel, sends the message, and awaits response.
// The function returns the value as a byte stream.
func (g *GamePram) RequestSomething(eventType string) []byte {
	r := make(chan interface{})
	event := &AbstractEvent{
		EventType: eventType,
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
	g.eventchan <- event  // Send Event
	a := <-r              // Wait for response
	output, ok := a.(int) // Converts the empty interface into Integer
	if ok {
		return output
	}
	return 0 // If something unexpected happens, return 0.
}

func (g *GamePram) LogoutEvent(playerId int) {
	event := &AbstractEvent{
		EventType: "Logout",
		TargetId:  playerId,
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
