package githubPatches

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type File struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	Size     int    `json:"size"`
}

type Files struct {
	JandedobbeleerPyOMPJSON File `json:"jandedobbeleer_py.omp.json"`
}

type Owner struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GistData struct {
	URL         string  `json:"url"`
	ForksURL    string  `json:"forks_url"`
	CommitsURL  string  `json:"commits_url"`
	ID          string  `json:"id"`
	NodeID      string  `json:"node_id"`
	GitPullURL  string  `json:"git_pull_url"`
	GitPushURL  string  `json:"git_push_url"`
	HTMLURL     string  `json:"html_url"`
	Files       Files   `json:"files"`
	Public      bool    `json:"public"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Description string  `json:"description"`
	Comments    int     `json:"comments"`
	User        *string `json:"user"`
	Owner       Owner   `json:"owner"`
	Truncated   bool    `json:"truncated"`
	RawURL      string  `json:"raw_url"`
}

func GetLast100Gists() string {

	url := "https://api.github.com/gists/public?per_page=100"

	response, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Error: received status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	return string(body)

}

func ParseGistData(data string) ([]GistData, error) {

	var gists []GistData
	err := json.Unmarshal([]byte(data), &gists)
	if err != nil {
		return nil, err
	}

	for i := range gists {
		gists[i].RawURL = gists[i].HTMLURL + "/raw"
	}

	return gists, nil
}
