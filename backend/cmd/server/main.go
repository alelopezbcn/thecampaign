package main

import (
	"log"
	"os"

	"github.com/alelopezbcn/thecampaign/internal/server"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewServer()
	addr := ":" + port

	log.Printf("Starting The Campaign card game server...")
	log.Printf("Visit http://localhost:%s to play", port)

	if err := srv.Start(addr); err != nil {
		log.Fatal("Server error:", err)
	}
}
