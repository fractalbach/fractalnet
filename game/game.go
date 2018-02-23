package game

import (
	"encoding/base64"
	"encoding/json"
	"log"
)

type World struct {
	Ents   map[int]*Ent
	nextid int
	h      int //height (number of tiles in y direction)
	w      int //width  (number of tiles in x direction)
	Trees  *BoolGrid
}

type Ent struct {
	Name     string
	Type     string
	Location []float64
}

type GameState struct {
	State map[int]*Ent
}

type TreeState struct {
	Trees string
}

type TreeLoc struct {
	TreeLocs string
}

// ______________________________________________________
// 		Creating a new World
// ------------------------------------------------------

func MakeNewWorld() *World {
	return &World{
		Ents:   map[int]*Ent{},
		nextid: 1,
		h:      48,
		w:      48,
		Trees:  CreateRandomInitialTrees(48, 48),
	}
}

// ______________________________________________________
// 		Manipulating the World
// ------------------------------------------------------

func (w *World) getTrees() *BoolGrid {
	return w.Trees
}

func (w *World) getMapHeight() int {
	return w.h
}

func (w *World) getMapWidth() int {
	return w.w
}

func (w *World) GetNextId() int {
	output := w.nextid
	w.nextid++
	return output
}

func (w *World) GeneratePlayer(username string) (int, bool) {
	id := w.GetNextId()
	ok := w.AddEntity(&Ent{Name: username, Type: "player"}, id)
	if ok {
		return id, true
	}
	return 0, false
}

func AppendEnt(w *World, e *Ent) bool {
	return w.AddEntity(e, w.GetNextId())
}

func (w *World) ChangeLocationEntity(id int, l []float64) {
	if _, ok := w.Ents[id]; ok {
		w.Ents[id].Location = l
	}
}

func (w *World) AddEntity(e *Ent, id int) bool {
	if _, ok := w.Ents[id]; ok {
		return false
	}
	w.Ents[id] = e
	return true
}

func (w *World) DeleteEntity(id int) bool {
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

func (w *World) DoGameEvent(a *AbstractEvent) {
	switch a.EventType {
	case "Move":
		w.ChangeLocationEntity(a.SourceId, a.GetPosition())

	case "ToggleTree":
		x := int(a.GetPosition()[0])
		y := int(a.GetPosition()[1])
		w.Trees.FlipBool(x, y)

	case "Login":
	case "Logout":
	case "Create":
	case "Delete":
	}
}

func (w *World) DoAdminEvent(a *AbstractEvent) {
	switch a.EventType {
	case "UpdateTrees":
		w.Trees.NextGeneration()
	}
}

// ______________________________________________________
// 		Game State
// ------------------------------------------------------

func (w *World) StateAllEntities() []byte {
	b, err := json.Marshal(GameState{w.Ents})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return b
}

func (w *World) StateAllTrees() []byte {
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

func (w *World) StateAllTreesLocations() []byte {
	intList, err := ConvertBoolGridToLocationList(w.Trees.grid)
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	b64 := base64.StdEncoding.EncodeToString(intList)
	msg, err := json.Marshal(TreeLoc{b64})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return msg
}

// ______________________________________________________
//  The Game PRAM
// ------------------------------------------------------
// 	NOTE: 	For the Future.  Will help to separate
// 			game events from the connection handling,
//			and things like the chat room.
// ------------------------------------------------------
/*
type GamePram struct {
	msgchan   chan Message
	adminchan chan AdminMessage
}

func NewGamePram() *GamePram {
	storedpram := &GamePram{
		msgchan:   make(chan Message),
		adminchan: make(chan AdminMessage),
	}
	go storedpram.run()
	return storedpram
}

func (gp *GamePram) run() {
	for {
		select {
		case incoming := <-gp.msgchan:
			EventHandler(incoming)

		case incoming := <-gp.adminchan:
			AdminEventHandler(incoming)

		}
	}
}
*/
