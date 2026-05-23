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
			Timeout: 10 * time.Second,
		},
		baseURL: os.Getenv("ANILIST_URL"),
	}
}

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

func (c *Client) FetchAnilistMedia(ctx context.Context, id int, page *int, perPage *int) (*Media, error) {
	query := `query Media($id: Int) {
		Media(id: $id, type: ANIME) {
			id
			episodes
			bannerImage
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
			}
		}
	}`

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
	query :=
		`query($page: Int, $perPage: Int) {
		Page(page: $page, perPage: $perPage) {
			media(type: ANIME, sort: [TRENDING_DESC]) {
				id
				episodes
				bannerImage
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
			}
		}
	}`

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

func (c *Client) FetchMediaBySeason(ctx context.Context, page int, perPage int, season SEASON, year int) ([]Media, error) {
	query := `query($page: Int, $perPage: Int, $season: MediaSeason, $seasonYear: Int) {
  Page(page: $page, perPage: $perPage) {
    media(type: ANIME, season: $season, seasonYear: $seasonYear) {
      id
			episodes
			bannerImage
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
    }
  }
}`

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"page":       page,
			"perPage":    perPage,
			"season":     season,
			"seasonYear": year,
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
	resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Page.Media, nil
}

func (c *Client) FetchMediaByGenre(ctx context.Context, page int, perPage int, genre ...Genre) ([]Media, error) {
	query := `query ($page: Int, $perPage: Int, $genreIn: [String]) {
  Page(page: $page, perPage: $perPage) {
    media(type: ANIME, genre_in: $genreIn) {
      id
			episodes
			bannerImage
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
    }
  }
}`
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
