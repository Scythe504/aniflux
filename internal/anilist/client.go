package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: os.Getenv("ANILIST_URL"),
	}
}

const mediaFields = `id
		episodes
		bannerImage
		description
		season
		popularity
		duration
		coverImage {
			color
			large
			extraLarge
			medium
		}
		countryOfOrigin
		averageScore
		format
		genres
		title {
			english
			romaji
			native
			userPreferred
		}
		status
		seasonYear
		nextAiringEpisode {
			airingAt
			episode
			id
			mediaId
			timeUntilAiring
		}`

func (c *Client) do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) FetchAnilistMedia(ctx context.Context, id int) (*Media, error) {
	query := fmt.Sprintf(`query Media($id: Int) {
		Media(id: $id, type: ANIME) {
			%s
		}
	}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]int{
			"id": id,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Media Media `json:"Media"`
		} `json:"data"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data.Media, nil
}

func (c *Client) FetchAnilistTrending(ctx context.Context, page int, perPage int) ([]Media, error) {
	query := fmt.Sprintf(`query($page: Int, $perPage: Int) {
		Page(page: $page, perPage: $perPage) {
			media(type: ANIME, sort: [TRENDING_DESC]) {
				%s
			}
		}
	}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]int{
			"page":    page,
			"perPage": perPage,
		},
	}
	body, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Page struct {
				Media []Media `json:"media"`
			} `json:"page"`
		} `json:"data"`
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Data.Page.Media, nil
}

func (c *Client) FetchMediaBySeason(ctx context.Context, page int, perPage int, season *SEASON, year *int) ([]Media, error) {
	query := fmt.Sprintf(`query($page: Int, $perPage: Int, $season: MediaSeason, $seasonYear: Int) {
  Page(page: $page, perPage: $perPage) {
    	media(type: ANIME, season: $season, seasonYear: $seasonYear) {
      	%s
    	}
  	}
	}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":    page,
			"perPage": perPage,
		},
	}

	if varsMap, ok := payload["variables"].(map[string]any); ok {
		if season != nil {
			varsMap["season"] = *season
		}
		if year != nil {
			varsMap["seasonYear"] = *year
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Page struct {
				Media []Media `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Page.Media, nil
}

func (c *Client) FetchMediaByGenre(ctx context.Context, page int, perPage int, genre []string) ([]Media, error) {
	query := fmt.Sprintf(`query ($page: Int, $perPage: Int, $genreIn: [String]) {
  Page(page: $page, perPage: $perPage) {
    media(type: ANIME, genre_in: $genreIn) {
      %s
    }
  }
}`, mediaFields)
	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":    page,
			"perPage": perPage,
			"genreIn": genre,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var response struct {
		Data struct {
			Page struct {
				Media []Media `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Data.Page.Media, nil
}

// func (c *Client) FetchAiringMedia(ctx context.Context, page int, perPage int) ([]Media, error) {
// 	query := fmt.Sprintf(``)
// }

func (c *Client) Search(ctx context.Context, page, perPage int, searchQuery string) ([]Media, error) {
	query := fmt.Sprintf(`query Query($page: Int, $perPage: Int, $search: String) {
		Page(page: $page, perPage: $perPage) {
			media(search: $search, type: ANIME) {
			%s
			}
		}
	}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":    page,
			"perPage": perPage,
			"search":  searchQuery,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Page struct {
				Media []Media `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Page.Media, nil
}

func (c *Client) FetchRecommendations(ctx context.Context, page, perPage, anilistId int) ([]Media, error) {
	query := fmt.Sprintf(`query Media($page: Int, $perPage: Int, $mediaId: Int) {
  Media(id: $mediaId, type: ANIME) {
    recommendations(page: $page, perPage: $perPage) {
      nodes {
        mediaRecommendation {
          %s
        }
      }
    }
  }
}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":    page,
			"perPage": perPage,
			"mediaId": anilistId,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Media struct {
				Recommendations struct {
					Nodes []struct {
						MediaRecommendation Media `json:"mediaRecommendation"`
					} `json:"nodes"`
				} `json:"recommendations"`
			} `json:"Media"`
		} `json:"data"`
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	mediaRec := &response.Data.Media.Recommendations.Nodes
	media := make([]Media, len(*mediaRec))
	for idx, m := range *mediaRec {
		media[idx] = m.MediaRecommendation
	}

	return media, nil
}

func (c *Client) FetchAiringSchedule(ctx context.Context, page int, perPage int, airingAtLesser int, airingAtGreater int) ([]AiringSchedule, error) {
	query := fmt.Sprintf(`query AiringSchedule($page: Int, $perPage: Int, $airingAtLesser: Int, $airingAtGreater: Int) {
		Page(page: $page, perPage: $perPage) {
			airingSchedules(sort: [TIME_DESC], airingAt_lesser: $airingAtLesser, airingAt_greater: $airingAtGreater) {
				id
				timeUntilAiring
				airingAt
				episode
				media {
					%s
				}
			}
		}
	}`, mediaFields)

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":            page,
			"perPage":         perPage,
			"airingAtLesser":  airingAtLesser,
			"airingAtGreater": airingAtGreater,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Page struct {
				AiringSchedules []AiringSchedule `json:"airingSchedules"`
			} `json:"Page"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Page.AiringSchedules, nil
}

