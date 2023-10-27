package develop

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// The port on which we are hosting the reload server has to be hardcoded on the client-side too.
	reloadAddress = "localhost:12450"
)

//go:embed livereload_client.js
var livereloadClientJS string

var (
	newline            = []byte{'\n'}
	PageReloadClientJS = "<script type=\"text/javascript\">" + fmt.Sprintf(livereloadClientJS, reloadAddress) + "</script>"
)

// ReloadServer implements a websocket communication to implement a similar behavior like the "live-reload" package
type ReloadServer struct {
	hub *hub
}

// NewReloadServer returns a new instance
func NewReloadServer() *ReloadServer {
	return &ReloadServer{
		hub: newHub(),
	}
}

// Start initiates a new server instance which handles reload commands
func (rs *ReloadServer) Start() {
	go rs.hub.run()
	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		// serveWs handles websocket requests from the peer.
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		client := &client{hub: rs.hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client
		go client.writePump()
		client.readPump()
	})
	go func() {
		err := http.ListenAndServe(reloadAddress, nil)
		if err != nil {
			log.Println("Failed to start up the Reload server: ", err)
			return
		}
	}()
	fmt.Printf("🔄 Reload server is listening at: '%s'\n", reloadAddress)

	// once the client is registered "fire" the reload command
	go func() {
		<-rs.hub.registered
		message := bytes.TrimSpace([]byte("reload"))
		rs.hub.broadcast <- message
	}()
}

// Reload sends a command to connected clients to reload the current page
func (rs *ReloadServer) Reload() {
	message := bytes.TrimSpace([]byte("reload"))
	rs.hub.broadcast <- message
}

// --------------------------------------------------------------------------
//   client type
// --------------------------------------------------------------------------

// client is an middleman between the websocket connection and the hub.
type client struct {
	hub *hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("An error happened when reading from the Websocket client: %v", err)
			}
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *client) write(mt int, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// --------------------------------------------------------------------------
//   hub type
// --------------------------------------------------------------------------

// hub maintains the set of active clients and broadcasts messages to the clients.
type hub struct {
	// Registered clients.
	clients map[*client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client

	// registered notifies others if a new client has registered
	registered chan *client
}

func newHub() *hub {
	return &hub{
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[*client]bool),
		registered: make(chan *client),
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.registered <- client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
