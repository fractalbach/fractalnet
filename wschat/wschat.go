package wschat

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fractalbach/fractalnet/game"
	"github.com/fractalbach/fractalnet/namegen"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024

	// Maximum number of active clients allowed.
	maxActiveClients = 10

	// Number of Messages saved on the server.
	maxSave int = 30
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn // The websocket connection.
	send     chan []byte     // Buffered channel of outbound messages.
	username string          // Username associated with a specific client.
	playerid int
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.broadcast <- []byte(c.username + " has logged out.")
		log.Println("Client Un-Registered: ", c.conn.RemoteAddr())
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Log the message before the additions, so you don't end up
		// with a bunch of duplicate timestamps and addresses in the log.

		log.Println(c.conn.RemoteAddr(), "Player:", c.playerid, string(message))

		// If the Json is not valid, ignore it entirely.

		if !(json.Valid(message)) {
			log.Println("Ignored Invalid Json from ", c.conn.RemoteAddr())
			continue
		}

		// First, the json byte blob is converted into AbstractEvent object(s).
		// Next, the source fields are overwritten to match the player.
		// If any errors are encountered (or the formatting is bad), then the
		// message is rejected and ignored.

		if len(message) < 1 {
			continue
		}

		// The assumption is that a json starting with curly left brace is
		// a single object -> therefore a single event.
		// If the json starts with a square left bracket, then it is an array.

		switch message[0] {
		case '{':
			event, err := game.MakePlayerEvent(message)
			if err != nil {
				log.Println(err)
				continue
			}
			c.eventSwitcher(event)

		case '[':
			eventArr, err := game.MakePlayerEventArray(message)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, event := range *eventArr {
				c.eventSwitcher(&event)
			}

		default:
			continue
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {

	// Check to see if there are too many active clients already.
	if thereAreTooManyActiveClients(hub, maxActiveClients) {
		log.Println("Too many active clients.")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Register a new Client connection into the hub.
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: namegen.GenerateUsername(),
	}

	client.hub.register <- client
	log.Println(
		"Client Registered:", client.conn.RemoteAddr(), client.username)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	// Send a welcoming message, and then request game state messages to be
	// displayed, so that the new player can learn about what is happening.
	client.hub.broadcast <- []byte("Welcome, " + client.username + ".")
	client.hub.broadcast <- client.hub.pram.RequestGameState()
	client.hub.broadcast <- client.hub.pram.RequestTreeState()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//
// Client Event Switch.
//
// eventSwitcher is the first function that an event message gets passed to.
// It can be seen as the "first line of defense" before any messages are
// sent to the game event handler.
//
// Certain events should trigger some specific behavior before entering
// that main loop (specifically chat).  This is where that happens.
//
// NOTE: Try NOT to be redundant with this eventSwitcher.  It's starting to
// look like the DoGameEvent() function...
//
func (c *Client) eventSwitcher(event *game.AbstractEvent) {
	switch event.EventType {
	case "Chat":
		message, err := json.Marshal(game.ChatMessage{
			prettyNow() + " > " + c.username + ": " + event.GetEventBody()})
		if err != nil {
			log.Println(err)
			return
		}
		addMessage(message)
		c.hub.broadcast <- message
		return
	/*
		case "ToggleTree":
			newVal := false
			if event.EventBody == "on" {
				newVal = true
			}
			x, y := event.Location.X, event.Location.Y
			c.hub.broadcast <- c.hub.pram.ToggleTreeEvent(x, y, newVal)
	*/

	default:
		c.hub.pram.CustomPlayerEvent(event)
		//c.hub.broadcast <- c.hub.pram.RequestGameState()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
