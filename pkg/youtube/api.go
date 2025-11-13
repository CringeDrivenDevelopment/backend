package youtube

import (
	"backend/internal/transport/api/dto"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/bytedance/sonic"
	"github.com/klauspost/compress/zstd"
)

type API struct {
	client *http.Client
}

func New() *API {
	return &API{client: &http.Client{}}
}

func (s *API) Search(ctx context.Context, query string) ([]dto.Track, error) {
	body := &SearchRequest{
		Query:  query,
		Params: FILTER_SONGS,
		Context: SearchRequestContext{
			Client: SearchRequestClient{
				Hl:            "en",
				Gl:            "US",
				ClientName:    "WEB_REMIX",
				ClientVersion: "1.20251110.03.00",
				OriginalUrl:   "https://music.youtube.com/",
			},
			User:    SearchRequestUser{LockedSafetyMode: false},
			Request: SearchRequestOptions{UseSsl: true},
		},
	}

	bodyBytes, err := sonic.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://music.youtube.com/youtubei/v1/search", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
	// TODO: set cookie, x-goog-visitor-id
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://music.youtube.com/?cbrd=1")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"142\", \"Not?A_Brand\";v=\"99\"")
	req.Header.Set("Sec-Ch-Ua-Arch", "x86")
	req.Header.Set("Sec-Ch-Ua-Bitness", "64")
	req.Header.Set("Sec-Ch-Ua-Form-Factors", "\"Desktop\"")
	req.Header.Set("Sec-Ch-Ua-Full-Version", "\"142.0.7444.60\"")
	req.Header.Set("Sec-Ch-Ua-Full-Version-List", "\"Chromium\";v=\"142.0.7444.60\", \"Not?A_Brand\";v=\"99.0.0.0\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Model", "\"\"")
	req.Header.Set("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Set("Sec-Ch-Ua-Platform-version", "12.0.0")
	req.Header.Set("Sec-Ch-Ua-Wow64", "?0")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "same-origin")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Youtube-Client-Name", "67")
	req.Header.Set("X-Youtube-Client-Version", "1.20251110.03.00")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Youtube-Bootstrap-Logged-In", "false")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	encoding := resp.Header.Get("Content-Encoding")
	var reader io.Reader

	switch encoding {
	case "gzip":
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func(gzReader *gzip.Reader) {
			err := gzReader.Close()
			if err != nil {
				return
			}
		}(gzReader)
		reader = gzReader
	case "br":
		// Handle brotli compression (you'll need a brotli decoder package)
		brReader := brotli.NewReader(resp.Body)
		reader = brReader
	case "zstd":
		// Handle zstd compression (you'll need a zstd decoder package)
		zstdReader, err := zstd.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer zstdReader.Close()
		reader = zstdReader
		// Add cases for "deflate" if needed
	default:
		// No compression or unsupported encoding
		reader = resp.Body
	}

	respBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	println(string(respBytes))

	var result SearchResponse
	if err := sonic.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}

	var data []struct {
		Data RawYtMusicSong `json:"musicResponsiveListItemRenderer"`
	}

	for _, tab := range result.Contents.TabbedSearchResultsRenderer.Tabs {
		for _, content := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			if content.MusicShelfRenderer.Contents == nil {
				continue
			}

			data = *content.MusicShelfRenderer.Contents
			break
		}
	}

	if data == nil {
		return nil, errors.New("cannot find music shelf")
	}

	tracks := make([]dto.Track, len(data))
	for i, result := range data {
		track, err := parseRaw(&result.Data)
		if err != nil {
			return nil, errors.New("cannot parse music track")
		}

		tracks[i] = track
	}

	return tracks, nil
}
