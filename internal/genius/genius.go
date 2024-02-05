package genius

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type SongsResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Songs    []Song `json:"songs"`
		NextPage *int   `json:"next_page"`
	} `json:"response"`
}

type Song struct {
	ArtistNames              string   `json:"artist_names"`
	FullTitle                string   `json:"full_title"`
	HeaderImageThumbnailURL  string   `json:"header_image_thumbnail_url"`
	HeaderImageURL           string   `json:"header_image_url"`
	SongID                   int      `json:"id"`   // "id" is from genius.com
	ID                       string   `json:"uuid"` // UUID is a reserved word in DynamoDB
	Path                     string   `json:"path"`
	ReleaseDateForDisplay    string   `json:"release_date_for_display"`
	SongArtImageThumbnailURL string   `json:"song_art_image_thumbnail_url"`
	SongArtImageURL          string   `json:"song_art_image_url"`
	Title                    string   `json:"title"`
	URL                      string   `json:"url"`
	Lyrics                   []string `json:"lyrics"` // only present after scraping, does not come from genius.com
}

func RequestPage(artistId int, artistName string, pageNumber int) ([]Song, *int) {
	path := fmt.Sprintf("/artists/%s/songs", strconv.Itoa(artistId))
	url := fmt.Sprintf("https://api.genius.com%s", path)
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	if pageNumber > 0 {
		query.Add("page", strconv.Itoa(pageNumber))
	}
	maxPageSize := "50"
	query.Add("per_page", maxPageSize)

	req.URL.RawQuery = query.Encode()

	client := &http.Client{}

	// slog.Info(fmt.Sprintf("Requesting %s", req.URL.String()))
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed request.", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", err)
	}
	var data SongsResponse
	json.Unmarshal(body, &data)

	songs := []Song{}

	// Only include songs where Young Thug is the only artist
	for _, song := range data.Response.Songs {
		artists := song.ArtistNames
		if artists == artistName {
			songs = append(songs, song)
		}
	}

	return songs, data.Response.NextPage
}
