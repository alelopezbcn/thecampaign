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
	"knight": {
		Description: `𝐇𝐞𝐚𝐯𝐢𝐥𝐲 𝐚𝐫𝐦𝐨𝐫𝐞𝐝 𝐰𝐚𝐫𝐫𝐢𝐨𝐫.
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Can be moved to field in any phase.
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Can attack with Swords. Deals 2x damage to Archers.
𝐒𝐩𝐞𝐜𝐢𝐚𝐥 𝐏𝐨𝐰𝐞𝐫: Protects an ally with Shield.
𝐃𝐚𝐦𝐚𝐠𝐞 𝐓𝐚𝐤𝐞𝐧: Takes 2x damage from Mages.`,
		Image: "knight.webp",
	},
	"archer": {
		Description: `𝐀 𝐬𝐰𝐢𝐟𝐭 𝐫𝐚𝐧𝐠𝐞𝐝 𝐟𝐢𝐠𝐡𝐭𝐞𝐫.
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Can be moved to field in any phase.
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Can attack with Arrows. Deals 2x damage to Mages.
𝐒𝐩𝐞𝐜𝐢𝐚𝐥 𝐏𝐨𝐰𝐞𝐫: Instant Kill.
𝐃𝐚𝐦𝐚𝐠𝐞 𝐓𝐚𝐤𝐞𝐧: Takes 2x damage from Knights.`,
		Image: "archer.webp",
	},
	"mage": {
		Description: `𝐀 𝐦𝐲𝐬𝐭𝐢𝐜𝐚𝐥 𝐬𝐩𝐞𝐥𝐥𝐜𝐚𝐬𝐭𝐞𝐫.
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Can be moved to field in any phase.
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Can attack with Poison. Deals 2x damage to Knights.
𝐒𝐩𝐞𝐜𝐢𝐚𝐥 𝐏𝐨𝐰𝐞𝐫: Heal an ally or self.
𝐃𝐚𝐦𝐚𝐠𝐞 𝐓𝐚𝐤𝐞𝐧: Takes 2x damage from Arrows.`,
		Image: "mage.webp",
	},
	"dragon": {
		Description: `𝐀 𝐦𝐢𝐠𝐡𝐭𝐲 𝐛𝐞𝐚𝐬𝐭.
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Can be moved to field in any phase.
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Can attack with any weapon.
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: Cannot use Special Powers.
𝐃𝐚𝐦𝐚𝐠𝐞 𝐓𝐚𝐤𝐞𝐧: Takes equal damage from all weapons. Instant kill takes 10 DMG.`,
		Image: "dragon.webp",
	},
	"mercenary": {
		Description: `𝐍𝐞𝐮𝐭𝐫𝐚𝐥 𝐖𝐚𝐫𝐫𝐢𝐨𝐫 𝐟𝐨𝐫 𝐇𝐢𝐫𝐞
𝐑𝐞𝐜𝐫𝐮𝐢𝐭𝐦𝐞𝐧𝐭 𝐂𝐨𝐬𝐭: 8+ 🪙.
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Can be recruited directly to your field.
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Can attack using any weapon.
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: Cannot use Special Powers.`,
		Image: "mercenary.webp",
	},
	"sword": {
		Description: `𝐒𝐰𝐨𝐫𝐝 (𝐖𝐞𝐚𝐩𝐨𝐧)
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Deals 2x damage to Archers.
𝐔𝐬𝐞𝐝 𝐁𝐲: Knights, Dragons, Mercenaries.
𝐔𝐭𝐢𝐥𝐢𝐭𝐲: A value-1 Sword can construct your Castle.
𝐓𝐫𝐚𝐝𝐞𝐚𝐛𝐥𝐞: Yes.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "sword.webp",
	},
	"arrow": {
		Description: `𝐀𝐫𝐫𝐨𝐰 (𝐖𝐞𝐚𝐩𝐨𝐧)
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Deals 2x damage to Mages.
𝐔𝐬𝐞𝐝 𝐁𝐲: Archers, Dragons, Mercenaries.
𝐔𝐭𝐢𝐥𝐢𝐭𝐲: A value-1 Arrow can construct your Castle.
𝐓𝐫𝐚𝐝𝐞𝐚𝐛𝐥𝐞: Yes.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "arrow.webp",
	},
	"poison": {
		Description: `𝐏𝐨𝐢𝐬𝐨𝐧 (𝐖𝐞𝐚𝐩𝐨𝐧)
𝐂𝐨𝐦𝐛𝐚𝐭 𝐀𝐛𝐢𝐥𝐢𝐭𝐲: Deals 2x damage to Knights.
𝐔𝐬𝐞𝐝 𝐁𝐲: Mages, Dragons, Mercenaries.
𝐔𝐭𝐢𝐥𝐢𝐭𝐲: A value-1 Poison can construct your Castle.
𝐓𝐫𝐚𝐝𝐞𝐚𝐛𝐥𝐞: Yes.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "poison.webp",
	},
	"resource": {
		Description: `𝐆𝐨𝐥𝐝 𝐑𝐞𝐬𝐨𝐮𝐫𝐜𝐞
𝐄𝐟𝐟𝐞𝐜𝐭: Spend 2 Coins to buy 1 Card.
𝐌𝐞𝐫𝐜𝐞𝐧𝐚𝐫𝐲: Spend 8+ Coins to recruit directly to field.
𝐔𝐭𝐢𝐥𝐢𝐭𝐲: A value-1 Gold can construct your Castle.
𝐏𝐡𝐚𝐬𝐞: Buy 💰 / Build 🏰`,
		Image: "gold.webp",
	},
	"specialpower": {
		Description: `𝐒𝐩𝐞𝐜𝐢𝐚𝐥 𝐏𝐨𝐰𝐞𝐫
𝐊𝐧𝐢𝐠𝐡𝐭: Shields an ally warrior.
𝐀𝐫𝐜𝐡𝐞𝐫: Instantly kills an enemy.
𝐌𝐚𝐠𝐞: Fully heals an ally or self.
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: Cannot be used by Dragons / Mercenaries.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "specialpower.webp",
	},
	"harpoon": {
		Description: `𝐇𝐚𝐫𝐩𝐨𝐨𝐧
𝐄𝐟𝐟𝐞𝐜𝐭: A powerful weapon that kills Dragons in one hit.
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: Can ONLY be used against Dragons.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "harpoon.webp",
	},
	"bloodrain": {
		Description: `𝐁𝐥𝐨𝐨𝐝 𝐑𝐚𝐢𝐧
𝐄𝐟𝐟𝐞𝐜𝐭: A devastating area attack.
𝐃𝐚𝐦𝐚𝐠𝐞: Deals 4 damage to ALL enemy warriors.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "bloodrain.webp",
	},
	"spy": {
		Description: `𝐒𝐩𝐲
𝐄𝐟𝐟𝐞𝐜𝐭: Peek at an opponent's full hand OR the top 5 cards of the deck.
𝐏𝐡𝐚𝐬𝐞: Spy / Steal / Sabotage 🎭`,
		Image: "spy.webp",
	},
	"thief": {
		Description: `𝐓𝐡𝐢𝐞𝐟
𝐄𝐟𝐟𝐞𝐜𝐭: Steal a random card from an opponent's hand.
𝐏𝐡𝐚𝐬𝐞: Spy / Steal / Sabotage 🎭`,
		Image: "thief.webp",
	},
	"catapult": {
		Description: `𝐂𝐚𝐭𝐚𝐩𝐮𝐥𝐭
𝐄𝐟𝐟𝐞𝐜𝐭: Destroy 1 Gold resource from an enemy Castle.
𝐈𝐦𝐩𝐚𝐜𝐭: Reduces their total Castle value.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "catapult.webp",
	},
	"fortress": {
		Description: `𝐅𝐨𝐫𝐭𝐫𝐞𝐬𝐬
𝐄𝐟𝐟𝐞𝐜𝐭: Fortify a Castle to block the next Catapult attack.
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: The wall is destroyed after one hit.
𝐏𝐡𝐚𝐬𝐞: Build 🏰`,
		Image: "fortress.webp",
	},
	"resurrection": {
		Description: `𝐑𝐞𝐬𝐮𝐫𝐫𝐞𝐜𝐭𝐢𝐨𝐧
𝐄𝐟𝐟𝐞𝐜𝐭: Return a random warrior from the Cemetery to the field.
𝐓𝐚𝐫𝐠𝐞𝐭: Can be used on your field or an ally's.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "resurrection.webp",
	},
	"sabotage": {
		Description: `𝐒𝐚𝐛𝐨𝐭𝐚𝐠𝐞
𝐄𝐟𝐟𝐞𝐜𝐭: Destroy a random card from an opponent's hand.
𝐍𝐨𝐭𝐞: The card is discarded, not stolen.
𝐏𝐡𝐚𝐬𝐞: Spy / Steal / Sabotage 🎭`,
		Image: "sabotage.webp",
	},
	"treason": {
		Description: `𝐓𝐫𝐞𝐚𝐬𝐨𝐧
𝐄𝐟𝐟𝐞𝐜𝐭: Convince a weakened enemy warrior to betray their team.
𝐂𝐨𝐧𝐝𝐢𝐭𝐢𝐨𝐧: Target must have 5 HP or less.
𝐏𝐡𝐚𝐬𝐞: Attack ⚔️`,
		Image: "treason.webp",
	},
	"ambush": {
		Description: `𝐀𝐦𝐛𝐮𝐬𝐡
𝐃𝐞𝐩𝐥𝐨𝐲𝐦𝐞𝐧𝐭: Placed face-down. Triggers on regular weapon attacks.
𝐍𝐨𝐭𝐞: Finishes attack phase when used.
𝐏𝐨𝐬𝐬𝐢𝐛𝐥𝐞 𝐄𝐟𝐟𝐞𝐜𝐭𝐬:
• Reflect Damage (23%)
• Cancel Attack (23%)
• Steal Weapon (23%)
• Drain Life (23%)
• Instant Kill (8%)
𝐋𝐢𝐦𝐢𝐭𝐚𝐭𝐢𝐨𝐧: Only one Ambush per field.
𝐏𝐡𝐚𝐬𝐞: Attack 💰`,
		Image: "ambush.webp",
	},
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
