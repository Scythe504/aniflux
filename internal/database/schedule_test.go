package database

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func Test30HourAnimeScheduleIntegration(t *testing.T) {
	url := os.Getenv("BLUEPRINT_DB_URL")
	if url == "" {
		url = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"
	}
	if err := Migrate(url); err != nil {
		t.Fatalf("failed to migrate database for test: %v", err)
	}

	db := New()
	defer db.Close()

	ctx := context.Background()

	// Generate a unique anime ID to avoid collisions
	randomID := fmt.Sprintf("test_anime_%d", rand.Intn(1000000))

	// 1. Create a dummy AnimeRecord first to satisfy foreign key constraints
	dummyAnime := AnimeRecord{
		ID:            randomID,
		Type:          "anime",
		Title:         "Test 30H Schedule Anime",
		OriginalTitle: "Test 30H Schedule Anime",
		Status:        "RELEASING",
		UpdatedAt:     time.Now().UnixMilli(),
		CreatedAt:     time.Now().UnixMilli(),
	}

	err := db.UpsertAnime(ctx, dummyAnime)
	if err != nil {
		t.Fatalf("failed to insert dummy anime: %v", err)
	}

	// Make sure we clean up at the end
	defer func() {
		sImpl := db.(*service)
		_, _ = sImpl.pool.Exec(ctx, "DELETE FROM airing_schedule WHERE anime_id = $1", randomID)
		_, _ = sImpl.pool.Exec(ctx, "DELETE FROM anime WHERE id = $1", randomID)
	}()

	// 2. Tuesday 1:30 AM JST is:
	// Year: 2026, Month: August, Day: 11, Hour: 1, Minute: 30
	jst := time.FixedZone("JST", 9*60*60)
	airingTime := time.Date(2026, 8, 11, 1, 30, 0, 0, jst)

	airingRec := AiringRecord{
		AnimeID:         randomID,
		Episode:         1,
		AiringAt:        airingTime.Unix(),
		TimeUntilAiring: 3600,
	}

	// 3. Upsert the schedule
	err = db.UpsertAiringSchedule(airingRec)
	if err != nil {
		t.Fatalf("failed to upsert airing schedule: %v", err)
	}

	// 4. Query it back to verify the 30-hour clock fields are saved correctly
	sImpl := db.(*service)
	var scheduleDay, timing string
	err = sImpl.pool.QueryRow(ctx, "SELECT schedule_day, timing FROM airing_schedule WHERE anime_id = $1 AND episode = $2", randomID, 1).Scan(&scheduleDay, &timing)
	if err != nil {
		t.Fatalf("failed to query saved schedule: %v", err)
	}

	if scheduleDay != "Monday" {
		t.Errorf("expected schedule_day to be 'Monday', got '%s'", scheduleDay)
	}

	if timing != "25:30" {
		t.Errorf("expected timing to be '25:30', got '%s'", timing)
	}
}
