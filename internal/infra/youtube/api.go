package youtube

import (
	"bytes"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{client: &http.Client{}}
}

func (s *Service) Search(query string, filter string) (*SearchResponse, error) {
	body := &SearchRequest{
		Query: query,
		Context: struct {
			Client struct {
				Hl            string `json:"hl"`
				Gl            string `json:"gl"`
				ClientName    string `json:"clientName"`
				ClientVersion string `json:"clientVersion"`
				OriginalUrl   string `json:"originalUrl"`
			} `json:"client"`
			User struct {
				LockedSafetyMode bool `json:"lockedSafetyMode"`
			} `json:"user"`
			Request struct {
				UseSsl bool `json:"useSsl"`
			} `json:"request"`
		}(struct {
			Client struct {
				Hl            string `json:"hl"`
				Gl            string `json:"gl"`
				ClientName    string `json:"clientName"`
				ClientVersion string `json:"clientVersion"`
				OriginalUrl   string `json:"originalUrl"`
			} `json:"client"`
			User struct {
				LockedSafetyMode bool `json:"lockedSafetyMode"`
			} `json:"user"`
			Request struct {
				UseSsl bool `json:"useSsl"`
			} `json:"req"`
		}{
			Client: struct {
				Hl            string `json:"hl"`
				Gl            string `json:"gl"`
				ClientName    string `json:"clientName"`
				ClientVersion string `json:"clientVersion"`
				OriginalUrl   string `json:"originalUrl"`
			}{Hl: "en", Gl: "US", ClientName: "WEB_REMIX", ClientVersion: "1.20250716.03.00", OriginalUrl: "https://music.youtube.com/"},
			User: struct {
				LockedSafetyMode bool `json:"lockedSafetyMode"`
			}{LockedSafetyMode: false},
			Request: struct {
				UseSsl bool `json:"useSsl"`
			}{UseSsl: true},
		}),
	}

	if filter != "" {
		body.Params = filter
	}

	bodyBytes, err := sonic.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://music.youtube.com/youtubei/v1/search", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Youtube-Client-Name", "67")
	req.Header.Set("X-Youtube-Client-Version", "1.20250716.03.00")
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

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := sonic.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
