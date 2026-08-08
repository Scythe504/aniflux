package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
)

// Service represents a service that interacts with a database.
type Service interface {
	UpsertAnime(ctx context.Context, a AnimeRecord) error
	BulkUpsertAnime(ctx context.Context, records []AnimeRecord) error
	GetCurrentAiring(ctx context.Context, page, perPage int) ([]AnimeRecord, error)
	GetWeeklySchedule() ([]AiringRecord, error)
	UpsertAiringSchedule(a AiringRecord) error

	// Health returns a map of health status information.
	Health() map[string]string

	// Close terminates the database connection.
	Close() error
}

type service struct {
	pool *pgxpool.Pool
}

var (
	dbInstance *service
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	url := os.Getenv("BLUEPRINT_DB_URL")
	if url == "" {
		url = os.Getenv("DB_URL")
	}
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		url = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"
	}


	// Initialize pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Initializing database connection pool...")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatalf("failed to create pgxpool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Database connection pool established successfully.")

	dbInstance = &service{
		pool: pool,
	}

	return dbInstance
}

// Migrate runs goose database migrations for PostgreSQL.
func Migrate(url string) error {
	sqlDb, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migration: %w", err)
	}
	defer sqlDb.Close()

	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log.New(os.Stderr, "", log.LstdFlags))

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(sqlDb, "migrations"); err != nil {
		return err
	}

	return nil
}

// Health checks the health of the database connection by pinging the database.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database pool
	err := s.pool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db health check failed: %v", err) // Log the error but don't terminate
		return stats
	}

	// Database is up, add pool statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	poolStat := s.pool.Stat()
	stats["open_connections"] = strconv.Itoa(int(poolStat.TotalConns()))
	stats["in_use"] = strconv.Itoa(int(poolStat.AcquiredConns()))
	stats["idle"] = strconv.Itoa(int(poolStat.IdleConns()))
	stats["max_connections"] = strconv.Itoa(int(poolStat.MaxConns()))

	return stats
}

// Close closes the database connection.
func (s *service) Close() error {
	log.Println("Disconnected from database pool.")
	s.pool.Close()
	return nil
}
