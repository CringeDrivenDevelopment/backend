package youtube

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	FILTER_VIDEO = "EgWKAQIQAWoIEAMQBBAJEAo%3D"
	FILTER_SONGS = "EgWKAQIIAWoOEAMQCRAKEAQQERAQEBU%3D"
)

type Track struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Authors   string `json:"artists"`
	Thumbnail string `json:"thumbnail"`
	Length    int    `json:"length"`
	Explicit  bool   `json:"explicit"`
}

type RawYtMusicSong struct {
	Thumbnail   Thumbnail    `json:"thumbnail"`
	FlexColumns []FlexColumn `json:"flexColumns"`
	Badges      []Badge      `json:"badges,omitempty"`
}

type Thumbnail struct {
	Renderer struct {
		Data struct {
			Items []struct {
				Url string `json:"url"`
			} `json:"thumbnails"`
		} `json:"thumbnail"`
	} `json:"musicThumbnailRenderer"`
}

type FlexColumn struct {
	Renderer struct {
		Data struct {
			Runs []struct {
				Text               string `json:"text"`
				NavigationEndpoint struct {
					WatchEndpoint struct {
						VideoId string `json:"videoId"`
					} `json:"watchEndpoint,omitempty"`
				} `json:"navigationEndpoint,omitempty"`
			} `json:"runs"`
		} `json:"text"`
	} `json:"musicResponsiveListItemFlexColumnRenderer"`
}

type Badge struct {
	Renderer struct {
		Icon struct {
			IconType string `json:"iconType"`
		} `json:"icon"`
	} `json:"musicInlineBadgeRenderer"`
}

func getBestThumbnail(thumbnail Thumbnail) string {
	url := thumbnail.Renderer.Data.Items[0].Url

	fromSize := 60
	toSize := 544

	pattern := `w` + strconv.Itoa(fromSize) + `-h` + strconv.Itoa(fromSize) + `(-l\d+-rj)$`
	re := regexp.MustCompile(pattern)

	replacement := "w" + strconv.Itoa(toSize) + "-h" + strconv.Itoa(toSize) + "$1"

	return re.ReplaceAllString(url, replacement)
}

func checkExplicit(badges []Badge) bool {
	for _, badge := range badges {
		if badge.Renderer.Icon.IconType == "MUSIC_EXPLICIT_BADGE" {
			return true
		}
	}

	return false
}

func getTitleAndID(column FlexColumn) (string, string) {
	return column.Renderer.Data.Runs[0].Text, column.Renderer.Data.Runs[0].NavigationEndpoint.WatchEndpoint.VideoId
}

func parseTime(time string) (int, error) {
	parts := strings.Split(time, ":")
	var seconds int

	switch len(parts) {
	case 2: // MM:SS format
		minutes, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		secs, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
		seconds = minutes*60 + secs

	case 3: // HH:MM:SS format
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %w", err)
		}
		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		secs, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
		seconds = hours*3600 + minutes*60 + secs

	default:
		return 0, fmt.Errorf("unexpected time format: %s", time)
	}

	return seconds, nil
}

func getArtistsAndDuration(column FlexColumn) (string, int, error) {
	data := ""

	for _, text := range column.Renderer.Data.Runs {
		data += text.Text
	}

	splitData := strings.Split(data, " • ")

	duration, err := parseTime(splitData[len(splitData)-1])
	if err != nil {
		return "", 0, err
	}

	return splitData[0], duration, nil
}

func ParseRaw(song *RawYtMusicSong) (Track, error) {
	title, id := getTitleAndID(song.FlexColumns[0])
	artists, duration, err := getArtistsAndDuration(song.FlexColumns[1])
	if err != nil {
		return Track{}, err
	}

	return Track{
		Id:        id,
		Title:     title,
		Authors:   artists,
		Length:    duration,
		Thumbnail: getBestThumbnail(song.Thumbnail),
		Explicit:  checkExplicit(song.Badges),
	}, nil
}
