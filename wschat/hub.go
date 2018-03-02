package wschat

import (
	"github.com/fractalbach/fractalnet/game"
	"github.com/fractalbach/fractalnet/namegen"
	"log"
	"time"
)

var numberOfActiveClients int

// hub maintains the game world and the set of active clients
type Hub struct {

	// Game Parallel Random Access Machine
	pram *game.GamePram

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		pram:       game.NewGamePram(),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {

	// Set a Timer to Update the Tree Generations
	// treeUpdateTicker := time.NewTicker(1 * time.Second)
	// go h.treeUpdateTimer(treeUpdateTicker)
	lifeUpdateTicker := time.NewTicker(1 * time.Second)
	go h.lifeUpdateTimer(lifeUpdateTicker)

	// Enter Hub Loop; waiting for messages to arrive from clients.
	for {
		select {

		case client := <-h.register:
			h.clients[client] = true
			numberOfActiveClients++
			h.clientAutoLogin(client)
			sendAllSavedMessages(client)
			log.Println("There are now", numberOfActiveClients, "online.")

		case client := <-h.unregister:
			h.clientAutoLogout(client)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				numberOfActiveClients--
			}
			log.Println("There are now", numberOfActiveClients, "online.")

		// Messages sent to the hub's broadcast channel,
		// are sent to all other active clients.  If a message is unable
		// to receive a broadcast message, that connection is dropped.
		case message := <-h.broadcast:
			if len(message) == 0 {
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		} // End of Select
	} // End of For Loop
} // End of Hub Definition

// thereAreTooManyActiveClients counts the list of registered clients, and
// returns TRUE if there are more than the "max".
//
func thereAreTooManyActiveClients(hub *Hub, max int) bool {
	return len(hub.clients) > max
}

// prettyNow returns a string with a human-readable time stamp.
// Useful for adding to messages.  for the day, use: "_2 Jan, "
//
//      return time.Now().Format("3:04:05 PM")
//
func prettyNow() string {
	return time.Now().Format("3:04 PM")
}

// lifeUpdateTimer defines the actions of the timer, but not the rate of it.
// It sends a game event to trigger an update to the next generation,
// and a game event request for game state.
// The Game state is broadcast to all active clients.
func (h *Hub) lifeUpdateTimer(lifeUpdateTicker *time.Ticker) {
	for t := range lifeUpdateTicker.C {
		if numberOfActiveClients <= 0 {
			continue
		}
		log.Println("Life Update:", t)
		h.pram.UpdateLifeEvent()
		h.broadcast <- h.pram.RequestSomething("LifeState")
	}
}

// clientAutoLogin is temporary and essentially makes a guest account.
//
// A username is randomly generated using the "namegen" package, and a login
// event is created and sent to the game, which should take of the assignment
// of a object ID number.
func (h *Hub) clientAutoLogin(c *Client) {
	name := namegen.GenerateUsername()
	playerId := h.pram.LoginEvent(name)
	if playerId == 0 {
		log.Println("Player Entity could not be created! Login failed!")
		return
	}
	c.playerid = playerId
	c.username = name
	log.Println("New Login: (ID):", playerId, "(Username):", name)
}

// clientAutoLogout forces the logout of the player associated with this
// connection.  As a result the playerID.
func (h *Hub) clientAutoLogout(c *Client) {
	if c.playerid == 0 {
		log.Println("ClientAutoLogout not needed; Client has no playerid.")
		return
	}
	log.Println("Attempting to Logout:", c)
	h.pram.LogoutEvent(c.playerid)
}

var savedChatMessages [][]byte
var maxMessages int = 40

func saveNewMessage() {
	if len(savedChatMessages) >= 40 {
		savedChatMessages = savedChatMessages[1:40]
	}
}

func addMessage(m []byte) {
	if len(savedChatMessages) >= maxMessages {
		savedChatMessages = savedChatMessages[1:maxMessages]
	}
	savedChatMessages = append(savedChatMessages, m)
}

func sendAllSavedMessages(c *Client) {
	for _, v := range savedChatMessages {
		c.send <- v
	}
}
