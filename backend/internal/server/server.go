package server

import (
	"encoding/json"
	"log"
	"net/http"

	ws "github.com/alelopezbcn/thecampaign/internal/websocket"
	"github.com/gorilla/websocket"
)

// Version is set at build time via -ldflags
var Version = "dev"

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

// handleVersion returns the current server version as JSON
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": Version})
}

// CardConfigEntry holds display metadata for a single card type.
type CardConfigEntry struct {
	Description string `json:"description"`
	Image       string `json:"image"`
}

// cardConfig is the authoritative source of card descriptions and images.
// Key is the lowercase card sub_type or type used by the frontend.
// Adding a new card = one entry here.
var cardConfig = map[string]CardConfigEntry{
	"knight":       {Description: "A heavily armored warrior. Can attack with Swords. Special Power: Shield. Takes double damage from Poison.", Image: "knight.webp"},
	"archer":       {Description: "A swift ranged fighter. Can attack with Arrows. Special Power: Instant Kill. Takes double damage from Swords.", Image: "archer.webp"},
	"mage":         {Description: "A mystical spellcaster. Can attack with Poison. Special Power: Heal. Takes double damage from Arrows.", Image: "mage.webp"},
	"dragon":       {Description: "A mighty beast. Can attack with any weapon. Takes equal damage from all weapons. Instant kill takes 10 DMG. Cannot use Special Powers.", Image: "dragon.webp"},
	"mercenary":    {Description: "A neutral warrior for hire. Costs 6+ gold to recruit directly to your field. Can attack with any weapon. 15 HP. Cannot use Special Powers.", Image: "mercenary.webp"},
	"sword":        {Description: "Deals double damage to Archers. Used by Knights and Dragons. A value-1 Sword can also construct your castle. Can be traded.", Image: "sword.webp"},
	"arrow":        {Description: "Deals double damage to Mages. Used by Archers and Dragons. A value-1 Arrow can also construct your castle. Can be traded.", Image: "arrow.webp"},
	"poison":       {Description: "Deals double damage to Knights. Used by Mages and Dragons. A value-1 Poison can also construct your castle. Can be traded.", Image: "poison.webp"},
	"resource":     {Description: "Spend gold to buy cards (2 coins = 1 card). A value-1 Gold can also construct your castle.", Image: "gold.webp"},
	"specialpower": {Description: "Knight: shields an ally warrior. Archer: instantly kills an enemy. Mage: fully heals an ally. Cannot be used by Dragons.", Image: "specialpower.webp"},
	"harpoon":      {Description: "A powerful weapon that kills Dragons from one hit. Can only be used on Dragons.", Image: "harpoon.webp"},
	"bloodrain":    {Description: "A devastating attack that affects all enemy warriors. Deals 4 damage to all enemy warriors.", Image: "bloodrain.webp"},
	"spy":          {Description: "Peek at an opponent's full hand or the top 5 cards of the deck.", Image: "spy.webp"},
	"thief":        {Description: "Steal a random card from an opponent's hand.", Image: "thief.webp"},
	"catapult":     {Description: "Destroy one gold resource from a constructed enemy castle, reducing their castle value.", Image: "catapult.webp"},
	"fortress":     {Description: "Fortify your castle (or an ally's) to block the next catapult attack. The wall is destroyed when hit.", Image: "fortress.webp"},
	"resurrection": {Description: "Resurrect a random fallen warrior from the cemetery and place it on your field (or an ally's).", Image: "resurrection.webp"},
	"sabotage":     {Description: "Destroy a random card from an opponent's hand. The card is discarded, not stolen.", Image: "sabotage.webp"},
}

// handleCardConfig serves card display metadata (descriptions and images).
func (s *Server) handleCardConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cardConfig)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	// WebSocket endpoint
	http.HandleFunc("/ws", s.handleWebSocket)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	// Version endpoint
	http.HandleFunc("/api/version", s.handleVersion)

	// Card config endpoint
	http.HandleFunc("/api/card-config", s.handleCardConfig)

	// Main page
	http.HandleFunc("/", s.handleIndex)

	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}
