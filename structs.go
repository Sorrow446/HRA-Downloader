package main

import "time"

type MyTransport struct{}

type WriteCounter struct {
	Total      uint64
	TotalStr   string
	Downloaded uint64
	Percentage int
}

type Config struct {
	Email            string
	Password         string
	Urls             []string
	OutPath          string
	TrackTemplate    string
	DownloadBooklets bool
	MaxCoverSize     bool
	KeepCover        bool
	Language         string
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	OutPath string   `arg:"-o"`
}

type Auth struct {
	ResponseStatus  string `json:"response_status"`
	UserID          string `json:"user_id"`
	Country         string `json:"country"`
	Lastname        string `json:"lastname"`
	Firstname       string `json:"firstname"`
	SessionID       string `json:"session_id"`
	HasSubscription bool   `json:"has_subscription"`
	Status          string `json:"status"`
	Filter          string `json:"filter"`
	Hasfilter       string `json:"hasfilter"`
}

type TrackMeta struct {
	PlaylistAdd string `json:"playlistAdd"`
	IsFavorite  string `json:"isFavorite"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	ArtistID    string `json:"artistId"`
	Label       string `json:"label"`
	LabelID     string `json:"labelId"`
	Licensor    string `json:"licensor"`
	LicensorID  string `json:"licensorId"`
	TrackNumber int    `json:"trackNumber"`
	Copyright   string `json:"copyright"`
	Genre       string `json:"genre"`
	Playtime    int    `json:"playtime"`
	UPC         string `json:"upc"`
	ISRC        string `json:"isrc"`
	URL         string `json:"url"`
	Format      string `json:"format"`
}

type Covers struct {
	Title      string `json:"title"`
	Type       string `json:"type"`
	DocumentID string `json:"document_id"`
	Master     struct {
		FileURL string `json:"file_url"`
	} `json:"master"`
	Preview struct {
		FileURL string `json:"file_url"`
	} `json:"preview"`
	Thumbnail struct {
		FileURL string `json:"file_url"`
	} `json:"thumbnail"`
}

type AlbumMeta struct {
	Status         string `json:"status"`
	ResponseStatus string `json:"response_status"`
	Test           string `json:"test"`
	Data           struct {
		Results struct {
			ShopURL          string      `json:"shop_url"`
			AvailableFrom    time.Time   `json:"availableFrom"`
			AvailableUntil   time.Time   `json:"availableUntil"`
			Booklet          string      `json:"booklet"`
			PublishingStatus string      `json:"publishingStatus"`
			ID               string      `json:"id"`
			Title            string      `json:"title"`
			Artist           string      `json:"artist"`
			ArtistID         string      `json:"artistId"`
			Biography        string      `json:"biography"`
			Label            string      `json:"label"`
			LabelID          string      `json:"labelId"`
			Licensor         string      `json:"licensor"`
			LicensorID       string      `json:"licensorId"`
			DdexDate         string      `json:"ddexDate"`
			Copyright        string      `json:"copyright"`
			TrackCount       int         `json:"trackCount"`
			IsLeakBlock      bool        `json:"isLeakBlock"`
			CopyrightYear    int         `json:"copyrightYear"`
			ImportDate       string      `json:"importDate"`
			Genre            string      `json:"genre"`
			Playtime         int         `json:"playtime"`
			ProductionYear   int         `json:"productionYear"`
			ReleaseDate      time.Time   `json:"releaseDate"`
			Subgenre         string      `json:"subgenre"`
			UPC              string      `json:"upc"`
			ShortDescription string      `json:"shortDescription"`
			Caption          string      `json:"caption"`
			Tags             string      `json:"tags"`
			IsFavorite       string      `json:"isFavorite"`
			Tracks           []TrackMeta `json:"tracks"`
			Cover            Covers      `json:"cover"`
		} `json:"results"`
	} `json:"data"`
}
