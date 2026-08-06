package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

type WorkerResult struct {
	TaskID        string          `json:"task_id"`
	ResultMessage string          `json:"result_message"` // "ack", "success", "failed"
	Error         json.RawMessage `json:"error,omitempty"`
	Output        json.RawMessage `json:"output,omitempty"`
}

func writeResult(res WorkerResult) {
	data, _ := json.Marshal(res)
	fmt.Println(string(data))
}

func main() {
	log.SetOutput(os.Stderr)
	log.Println("Starting AniFlux Kronos Worker...")

	db := database.New()
	defer db.Close()
	client := anilist.New()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var wp WorkerPayload
		if err := json.Unmarshal(line, &wp); err != nil {
			log.Printf("Failed to unmarshal worker payload: %v", err)
			continue
		}

		// 1. Send immediate ACK back over os.Stdout
		writeResult(WorkerResult{
			TaskID:        wp.TaskID,
			ResultMessage: "ack",
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
			taskType = "aniflux_schedule"
		}

		log.Printf("Executing task %s (ID: %s)\n", taskType, wp.TaskID)

		var err error
		switch taskType {
		case "aniflux_schedule":
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
			errBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
			writeResult(WorkerResult{
				TaskID:        wp.TaskID,
				ResultMessage: "failed",
				Error:         errBytes,
			})
		} else {
			log.Printf("Task %s completed successfully", wp.TaskID)
			writeResult(WorkerResult{
				TaskID:        wp.TaskID,
				ResultMessage: "success",
			})
		}
	}
}
