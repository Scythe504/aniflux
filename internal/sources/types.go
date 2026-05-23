package sources

type TorznabItem struct {
	Title     string `json:"title"`
	MagnetURI string `json:"magnet_uri"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	Peers     int    `json:"peers"`
	Size      int64  `json:"size"`
	InfoHash  string `json:"infohash"`
}

