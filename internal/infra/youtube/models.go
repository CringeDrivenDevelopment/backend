package youtube

const (
	FILTER_VIDEO = "EgWKAQIQAWoIEAMQBBAJEAo%3D"
	FILTER_SONGS = "EgWKAQIIAWoOEAMQCRAKEAQQERAQEBU%3D"
)

type SearchResponse struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
								MusicShelfRenderer struct {
									Contents *[]struct {
										Data RawYtMusicSong `json:"musicResponsiveListItemRenderer"`
									} `json:"contents"`
								} `json:"musicShelfRenderer"`
							} `json:"contents"`
						} `json:"sectionListRenderer"`
					} `json:"content"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"tabbedSearchResultsRenderer"`
	} `json:"contents"`
}

type SearchRequest struct {
	Context struct {
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
	} `json:"context"`
	Query  string `json:"query"`
	Params string `json:"params"`
}

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
