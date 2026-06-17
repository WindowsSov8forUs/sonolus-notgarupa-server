package main

import (
	"flag"
	"log"
	"os"

	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/app"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/config"
)

func main() {
	cfg := config.Load(configPath(os.Args[1:]))

	router, err := app.BuildRouter(cfg)
	if err != nil {
		log.Fatalf("build router: %v", err)
	}

	log.Printf("sonolus-notgarupa-server listening on %s", cfg.Server.Listen)
	if err := router.Run(cfg.Server.Listen); err != nil {
		log.Fatal(err)
	}
}

func configPath(args []string) string {
	flags := flag.NewFlagSet("sonolus-notgarupa-server", flag.ExitOnError)
	defaultPath := envOrDefault("SONOLUS_CONFIG", "config.ini")
	path := flags.String("config", defaultPath, "path to config file")
	flags.StringVar(path, "c", defaultPath, "path to config file")
	flags.Parse(args)
	return *path
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
