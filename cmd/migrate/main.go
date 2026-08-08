package main

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/scythe504/aniflux/internal/database"
)

func main() {
	url := os.Getenv("BLUEPRINT_DB_URL")
	if url == "" {
		url = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"
	}

	log.Println("Running database migrations...")
	if err := database.Migrate(url); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations completed successfully!")
}
