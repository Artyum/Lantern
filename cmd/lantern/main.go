package main

import (
	"log"
	"net/http"
	"os"

	"lantern/internal/config"
	"lantern/internal/icons"
	"lantern/internal/server"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	configPath := env("CONFIG_PATH", "/config/config.json")
	cacheDir := env("CACHE_DIR", "/data/cache")
	listen := env("LISTEN", ":8080")

	store, err := config.NewStore(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	resolver, err := icons.NewResolver(cacheDir, nil)
	if err != nil {
		log.Printf("icon cache unavailable (%v); serving fallbacks only", err)
		resolver, err = icons.NewResolver(os.TempDir(), nil)
		if err != nil {
			log.Fatalf("icons: %v", err)
		}
	}

	cfg, _, err := store.Get()
	if err == nil {
		resolver.Warmup(cfg)
	}

	log.Printf("Lantern listening on %s", listen)
	if err := http.ListenAndServe(listen, server.New(store, resolver)); err != nil {
		log.Fatal(err)
	}
}
