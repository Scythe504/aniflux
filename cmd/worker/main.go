package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/indexer"
)

type TaskType string

const (
	TaskScheduleIndex TaskType = "aniflux_schedule"
	TaskSeasonIndex   TaskType = "aniflux_season"
	TaskMediaUpsert   TaskType = "aniflux_media_upsert"
)

type Payload struct {
	TaskId     string   `json:"task_id"`
	TaskType   TaskType `json:"task_type"`
	AnilistId  int      `json:"anilist_id"`
	Season     string   `json:"season"`
	SeasonYear int      `json:"season_year"`
}

type WorkerResult struct {
	Success bool   `json:"success"`
	Task    string `json:"task_type"`
	Message string `json:"message"`
}

func main() {
	log.Println("Starting AniFlux Kronos Worker...")

	// Read JSON payload from Stdin until EOF
	payloadBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
		os.Exit(1)
	}

	var p Payload
	if len(payloadBytes) > 0 {
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			log.Printf("Warning: payload unmarshal error (%v), falling back to default schedule task\n", err)
		}
	}

	// Default to schedule indexing if task_type is unassigned
	if p.TaskType == "" {
		p.TaskType = TaskScheduleIndex
	}

	log.Printf("Executing task %s (ID: %s)\n", p.TaskType, p.TaskId)

	// Initialize connections and clients
	db := database.New()
	defer db.Close()

	client := anilist.New()
	ctx := context.Background()

	// Dispatch by TaskType
	switch p.TaskType {
	case TaskScheduleIndex:
		err = indexer.UpdateWeeklySchedule(ctx, db, client)
	case TaskSeasonIndex:
		if p.Season == "" || p.SeasonYear == 0 {
			err = fmt.Errorf("season and season_year are required for %s", TaskSeasonIndex)
		} else {
			err = indexer.UpdateSeasonalMedia(ctx, db, client, p.Season, p.SeasonYear)
		}
	case TaskMediaUpsert:
		if p.AnilistId == 0 {
			err = fmt.Errorf("anilist_id is required for %s", TaskMediaUpsert)
		} else {
			err = indexer.UpdateMediaEntry(ctx, db, client, p.AnilistId)
		}
	default:
		err = fmt.Errorf("unknown task type: %s", p.TaskType)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Task failed with error: %v\n", err)
		os.Exit(1)
	}

	result := WorkerResult{
		Success: true,
		Task:    string(p.TaskType),
		Message: "AniFlux indexing completed successfully",
	}

	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	log.Println("Task finished successfully.")
	os.Exit(0)
}
