package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/indexer"
)

type WorkerPayload struct {
	TaskID  string          `json:"task_id"`
	Slug    string          `json:"slug"`
	Payload json.RawMessage `json:"payload"`
}

type TaskParams struct {
	TaskType   string `json:"task_type"`
	AnilistID  int    `json:"anilist_id"`
	Season     string `json:"season"`
	SeasonYear int    `json:"season_year"`
}

type WorkerResultMessage string

const (
	WorkerResultSuccessMessage WorkerResultMessage = "success"
	WorkerResultFailedMessage  WorkerResultMessage = "failed"
	WorkerResultACKMessage     WorkerResultMessage = "ack"
)

type WorkerResult struct {
	TaskID        string              `json:"task_id"`
	ResultMessage WorkerResultMessage `json:"result_message"`
	Error         json.RawMessage     `json:"error,omitempty"`
	Timestamp     time.Time           `json:"timestamp,omitempty"`
}

func writeResult(res WorkerResult) {
	_ = json.NewEncoder(os.Stdout).Encode(res)
}

func writeError(taskID string, errMsg string) {
	errBytes, _ := json.Marshal(errMsg)
	writeResult(WorkerResult{
		TaskID:        taskID,
		ResultMessage: WorkerResultFailedMessage,
		Error:         errBytes,
		Timestamp:     time.Now(),
	})
}

func main() {
	log.SetOutput(os.Stderr)
	log.Println("Starting AniFlux Kronos Worker...")

	db := database.New()
	defer db.Close()
	client := anilist.New()

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Println("No input received on stdin, exiting.")
		return
	}
	line := scanner.Bytes()

	var wp WorkerPayload
	if err := json.Unmarshal(line, &wp); err != nil {
		log.Printf("Failed to unmarshal worker payload: %v", err)
		return
	}

	// 1. Send immediate ACK back over os.Stdout
	writeResult(WorkerResult{
		TaskID:        wp.TaskID,
		ResultMessage: WorkerResultACKMessage,
		Timestamp:     time.Now(),
	})

	// 2. Process task
	ctx := context.Background()
	var params TaskParams
	if len(wp.Payload) > 0 {
		_ = json.Unmarshal(wp.Payload, &params)
	}

	taskType := params.TaskType
	if taskType == "" {
		taskType = wp.Slug
	}
	if taskType == "" || taskType == "aniflux" {
		taskType = "aniflux-daily-schedule"
	}

	log.Printf("Executing task %s (ID: %s)\n", taskType, wp.TaskID)

	var err error
	switch taskType {
	case "aniflux_schedule", "aniflux-daily-schedule", "aniflux-release-check":
		err = indexer.UpdateWeeklySchedule(ctx, db, client)
	case "aniflux_season":
		if params.Season == "" || params.SeasonYear == 0 {
			err = fmt.Errorf("season and season_year required for aniflux_season")
		} else {
			err = indexer.UpdateSeasonalMedia(ctx, db, client, params.Season, params.SeasonYear)
		}
	case "aniflux_media_upsert":
		if params.AnilistID == 0 {
			err = fmt.Errorf("anilist_id required for aniflux_media_upsert")
		} else {
			err = indexer.UpdateMediaEntry(ctx, db, client, params.AnilistID)
		}
	default:
		err = fmt.Errorf("unknown task type: %s", taskType)
	}

	if err != nil {
		log.Printf("Task %s failed: %v", wp.TaskID, err)
		writeError(wp.TaskID, err.Error())
	} else {
		log.Printf("Task %s completed successfully", wp.TaskID)
		writeResult(WorkerResult{
			TaskID:        wp.TaskID,
			ResultMessage: WorkerResultSuccessMessage,
			Timestamp:     time.Now(),
		})
	}
}
