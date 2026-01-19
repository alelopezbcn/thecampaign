package server

import (
	"log"
	"net/http"

	ws "github.com/alelopezbcn/thecampaign/internal/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, you should validate the origin
		return true
	},
}

// Server represents the HTTP server
type Server struct {
	hub *ws.Hub
}

// NewServer creates a new server
func NewServer() *Server {
	hub := ws.NewHub()
	go hub.Run()

	return &Server{
		hub: hub,
	}
}

// handleWebSocket handles WebSocket upgrade requests
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := ws.NewClient(s.hub, conn)
	s.hub.Register(client)
	client.Start()

	log.Printf("New WebSocket connection established")
}

// handleIndex serves the main HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/index.html")
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	// WebSocket endpoint
	http.HandleFunc("/ws", s.handleWebSocket)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	// Main page
	http.HandleFunc("/", s.handleIndex)

	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}
