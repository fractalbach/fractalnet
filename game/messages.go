// messages.go converts external and internal game messages into useful forms.
//
//
package game

import (
	"encoding/json"
)

type ChatMessage struct {
	Chat string
}

// Message is the result of a JSON Marshalling.
type Message struct {

	//Type is the "type of message", such as "move" or "login".
	Type string

	//Body includes all other details; Body can take many forms.
	Body interface{}
}

// AbstractEvent is a structured form of a message, ready to be used by game.
// It is created from either Player or Admin Messages.
//
type AbstractEvent struct {

	// EventId is a unique number used to compare different events going
	// through the system.  Currently NOT IN USE.
	EventId int

	// EventType distinguishes different kinds of events.
	// CURRENTLY NOT IN USE.
	EventType string

	// EventBody is for an arbitrary String, mostly used in chat messages.
	EventBody string

	// SourceId is usually the PlayerId of the client who sent initial request.
	SourceId int

	// SourceType is tagged on by the system, usually "player" or "admin".
	SourceType string

	// TargetId is usually a PlayerId used for "attack that person".
	// CURRENTLY NOT IN USE.
	TargetId int

	// TargetType could be something like "player" or "tree".
	// CURRENTLY NOT IN USE.
	TargetType string

	// Position is an array of values that imply a map location.
	// Can relate to either source or target, depending on the event type.
	Position []float64

	// GridLocation designates a specific x and y integer Location.
	Location Location

	// Response is a channel that used to return values back to the caller.
	Response chan interface{}
}

// _____________________________________________
//  Get - Accessors
// -------------------------------------------

func (m *Message) GetMessageType() string {
	return m.Type
}

func (m *Message) GetBody() interface{} {
	return m.Body
}

func (a *AbstractEvent) GetPosition() []float64 {
	return a.Position
}

func (a *AbstractEvent) GetEventBody() string {
	return a.EventBody
}

// _____________________________________________
//  Set - Mutators
// -------------------------------------------

func (a *AbstractEvent) SetEventBody(s string) {
	a.EventBody = s
}

func (a *AbstractEvent) SetPosition(newPos []float64) {
	a.Position = newPos
}

func (a *AbstractEvent) SetSourceId(id int) {
	a.SourceId = id
}

// _____________________________________________
//  Has - I haz that
// -------------------------------------------

func (a *AbstractEvent) HasResponseChannel() bool {
	if a.Response != nil {
		return true
	}
	return false
}

// ______________________________________________________________
//  Creating Events
// ==============================================================

// MakePlayerEvent converts a player's json byte stream into a useful event.
// After it fills in the event json as best as possible, it overwrites the
// SourceType and SourceId fields to those of a player.
//
// Example:
// https://play.golang.org/p/0ekubkpy_Ou
//
func MakePlayerEvent(jsonblob []byte) (*AbstractEvent, error) {
	m := new(AbstractEvent)
	err := json.Unmarshal([]byte(jsonblob), m)
	m.SourceType = "Player"
	return m, err
}

// MakePlayerEventArray returns a pointer to an array of player events.
// Similar to MakePlayerEvent(), the SourceType is overwritten to "Player",
// and the SourceId is overwritten to the given integer: "playerId".
//
// Example:
// https://play.golang.org/p/KqZwMOPwcbf
//
func MakePlayerEventArray(jsonblob []byte) (*[]AbstractEvent, error) {
	marr := new([]AbstractEvent)
	err := json.Unmarshal(jsonblob, marr)
	if err != nil {
		return marr, err
	}
	for i := len(*marr) - 1; i >= 0; i-- {
		(*marr)[i].SourceType = "Player"
	}
	return marr, err
}

/*  ~~~~ BEGIN DISABLED ~~~~

// _____________________________________________
//  Other & older methods
// -------------------------------------------

// PlayerJsonToEvent converts a byte stream into an useful event object.
//
// When using this function, be sure to check for errors, to
// ensure that the message was formatted correctly.
//
// The PlayerJsonToEvent is intended to be used by other programs in game.
// The byte array is aggressively checked for correct formatting.
// Any issues with formating will "bubble up" to this function, and return
// "false" for the boolean return value.
//

func PlayerJsonToEvent(jsonBlob []byte, playerId int) (*AbstractEvent, bool) {
	m := ParsePlayerMessage(jsonBlob)
	if e, ok := m.ConvertToEvent(); ok {
		e.SetSourceId(playerId)
		return e, true
	}
	log.Println(playerId, "sent a message that can't become an event.")
	return &AbstractEvent{}, false
}

func ParsePlayerMessage(jsonBlob []byte) *Message {
	m := &Message{}
	if !json.Valid(jsonBlob) {
		log.Println("Cannot Parse Player Message; Invalid Json")
		return m
	}
	err := json.Unmarshal(jsonBlob, m)
	if err != nil {
		log.Println("error UnMarshalling Message:", err)
	}
	return m
}

func (m *Message) ConvertToEvent() (*AbstractEvent, bool) {
	switch m.GetMessageType() {

	case "Move", "move":
		if a, ok := m.intoMoveEvent(); ok {
			return a, true
		}

	case "Chat", "chat":
		if a, ok := m.intoChatEvent(); ok {
			return a, true
		}

	case "ToggleTree", "toggleTree":
		if a, ok := m.intoToggleTreeEvent(); ok {
			return a, true
		}
	}

	return &AbstractEvent{}, false
}

// intoMoveEvent converts a message (with type and body) into an actual event.
//
// The body is expected to be an array of numbers, with no "null" or "nil".
// If there any any errors found when converting the message into the event,
// then the function returns an empty AbstractEvent, and a FALSE bool.
//
func (m *Message) intoMoveEvent() (*AbstractEvent, bool) {
	a := &AbstractEvent{EventType: "Move"}
	if newPos, ok := parseLocationArray(m.GetBody()); ok {
		a.SetPosition(newPos)
		return a, true
	}
	return a, false
}

func (m *Message) intoChatEvent() (*AbstractEvent, bool) {
	a := &AbstractEvent{EventType: "Chat"}
	if s, ok := m.GetBody().(string); ok {
		a.SetEventBody(s)
		return a, true
	}
	return a, false
}

// intoCreateTreeEvent accepts a message from a player, and expects json format:
//
//      {
//          Type: "ToggleTree",
//          Body: [1,2,3]
//      }
//

func (m *Message) intoToggleTreeEvent() (*AbstractEvent, bool) {
	a := &AbstractEvent{EventType: "ToggleTree"}
	if newPos, ok := parseLocationArray(m.GetBody()); ok {
		if len(newPos) >= 2 {
			a.SetPosition(newPos)
			return a, true
		}
	}
	return a, false
}

func (m *Message) intoCreateEvent() (*AbstractEvent, bool) {
	a := &AbstractEvent{EventType: "Create"}
	return a, false
}

// parseLocationArray converts the empty interface into a useful float array.
//
func parseLocationArray(i interface{}) ([]float64, bool) {
	if vals, ok := i.([]interface{}); ok {
		if pos, ok := convertLocationRequestToFloat(vals); ok {
			return pos, true
		}
	}
	log.Println("cannot parse location array")
	return []float64{}, false
}

func convertLocationRequestToFloat(vals []interface{}) ([]float64, bool) {
	var pos []float64
	for _, v := range vals {
		if f, ok := v.(float64); ok {
			pos = append(pos, f)
		} else {
			log.Println("badly formatted location; cannot convert to floats")
			return []float64{}, false
		}
	}
	return pos, true
}

~~~~ END DISABLED ~~~~ */
