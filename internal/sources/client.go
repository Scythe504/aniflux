package sources

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: os.Getenv("SOURCE_URL"),
		apiKey:  os.Getenv("SOURCE_API_KEY"),
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

func (item *TorznabItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw struct {
		Title string `xml:"title"`
		Attrs []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"attr"`
	}

	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	item.Title = raw.Title

	seen := map[string]bool{}
	for _, attr := range raw.Attrs {
		if seen[attr.Name] {
			continue
		}

		seen[attr.Name] = true
		switch attr.Name {
		case "magneturl":
			item.MagnetURI = attr.Value
		case "seeders":
			item.Seeders, _ = strconv.Atoi(attr.Value)
		case "peers":
			item.Peers, _ = strconv.Atoi(attr.Value)
		case "size":
			item.Size, _ = strconv.ParseInt(attr.Value, 10, 64)
		case "infohash":
			item.InfoHash = attr.Value
		}
	}

	return nil
}

func (c *Client) FetchSources(ctx context.Context, anidbEid int) ([]TorznabItem, error) {
	apiUrl := fmt.Sprintf("%s/api?t=search&cat=5070&extended=1&eid=%d&apikey=%s", c.baseURL, anidbEid, c.apiKey)

	resp, err := c.do(ctx, http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []TorznabItem `xml:"item"`
		} `xml:"channel"`
	}
	if err = xml.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Channel.Items, nil
}
