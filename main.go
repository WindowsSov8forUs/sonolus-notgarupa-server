package main

import (
	"log"
	"os"

	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/app"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/config"
)

func main() {
	cfg := config.Load(envOrDefault("SONOLUS_CONFIG", "config.ini"))

	router, err := app.BuildRouter(cfg)
	if err != nil {
		log.Fatalf("build router: %v", err)
	}

	log.Printf("sonolus-notgarupa-server listening on %s", cfg.Listen)
	if err := router.Run(cfg.Listen); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
