package anizip

import (
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
		baseURL: os.Getenv("ANIZIP_URL"),
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

func (c *Client) FetchAnizipData(ctx context.Context, mappingName MappingName, mappingId int) (*Resp, error) {
	apiUrl := fmt.Sprintf("%s/mappings?%s=%d", c.baseURL, mappingName, mappingId)
	resp, err := c.do(ctx, http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response Resp

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
