package githubPatches

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var GithubCacheDir = ".githubCommits/"

func GetISO8601Timestamps(from, to string) []string {

	var timestamps []string

	if isValidTimestamp(from) && !isValidTimestamp(to) {
		fmt.Println(timeOverflowError)
		return timestamps
	}

	if (from == "" && to != "") || (from != "" && to == "") {
		return timestamps
	}

	if (!isValidTimestamp(from) || !isValidTimestamp(to)) && (from != "" && to != "") {
		fmt.Println(timestampFormatError)
		return timestamps
	}

	if from == "" && to == "" {
		now := time.Now().UTC()
		previousHour := now.Add(-1 * time.Hour)
		timestamps = append(timestamps, previousHour.Format("2006-01-02-15"))
		return timestamps
	}

	fromTime, err := time.Parse("2006-01-02-15", from)
	if err != nil {
		fmt.Println("Invalid 'from' timestamp format. Use dd-mm-yyyy-H.")
		return timestamps
	}

	toTime, err := time.Parse("2006-01-02-15", to)
	if err != nil {
		fmt.Println("Invalid 'to' timestamp format. Use dd-mm-yyyy-H.")
		return timestamps
	}

	var step time.Duration
	if fromTime.After(toTime) {
		step = -1 * time.Hour
	} else {
		step = 1 * time.Hour
	}

	for t := fromTime; ; t = t.Add(step) {
		timestamps = append(timestamps, t.Format("2006-01-02-15"))
		if (step > 0 && !t.Before(toTime)) || (step < 0 && !t.After(toTime)) {
			break
		}
	}

	// remove zeroes from the front to work with gharchive.org
	var finalTimestamps []string

	for _, timestamp := range timestamps {
		parts := strings.Split(timestamp, "-")
		if len(parts) == 4 && strings.HasPrefix(parts[3], "0") {
			parts[3] = strings.TrimPrefix(parts[3], "0")
		}
		timestamp = strings.Join(parts, "-")
		finalTimestamps = append(finalTimestamps, timestamp)
	}

	return finalTimestamps
}

func PrintGharchiveChunkUrls(from, to string) []string {

	var urls []string

	timestamps := GetISO8601Timestamps(from, to)

	for _, timestamp := range timestamps {
		url := fmt.Sprintf("https://data.gharchive.org/%s.json.gz", timestamp)
		urls = append(urls, url)
	}

	return urls

}

func downloadAndExtract(url string) error {

	makeDir(GithubCacheDir)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	fileName := filepath.Base(url)
	outFile, err := os.Create(GithubCacheDir + fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err = outFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	gzFileName := GithubCacheDir + fileName

	gzFile, err := os.Open(gzFileName)

	if err != nil {
		return fmt.Errorf("failed to open gz file: %w", err)
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		os.Remove(gzFileName)
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	jsonFileName := gzFileName[:len(gzFileName)-3]
	jsonFile, err := os.Create(jsonFileName)
	if err != nil {
		return fmt.Errorf("failed to create json file: %w", err)
	}
	defer jsonFile.Close()

	if _, err = io.Copy(jsonFile, gzReader); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	fmt.Printf("Saved and extracted to %s\n", jsonFileName)
	os.Remove(gzFileName)

	return nil

}

func makeDir(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to create directory: %w", err)
		}
	}
	return nil
}

func isValidTimestamp(input string) bool {
	const layout = "2006-01-02-15"
	_, err := time.Parse(layout, input)
	return err == nil
}

var timestampFormatError = "Error: Please specify both to and from timestamps in the format '2006-01-02-15'."
var timeOverflowError = "The timestamp cannot be greater than the current UTC time!"

// Function to read and parse the JSON file
//
// Inputs:
//
// outputDirectory (string): the directory to save files to. "githubCommits/" will be made locally if not specified
//
// From (string): from timestamp as string in the format "2006-01-02-15"
//
// To (string): from timestamp as string in the format "2006-01-02-15"
//
// Concurrent (bool): download all the patch files at the same time. Note: this ramps up network, disk and CPU usage. Limit this
// to short timeframes.
//
// Note: if no timestamp is specified, the previous hour's JSON file will be downloaded.
//
// Output: A JSON file downloaded to the directory as specified
func GetCommitsInRange(outputDirectory, from, to string, concurrent bool) {

	if outputDirectory != "" {
		outputDirectory = outputDirectory + "/"
		makeDir(outputDirectory)
		GithubCacheDir = outputDirectory
	}

	urls := PrintGharchiveChunkUrls(from, to)

	if concurrent {
		fmt.Println("Downloading and extracting concurrently...")
		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				fmt.Printf("Downloading: %s\n", url)
				defer wg.Done()
				if err := downloadAndExtract(url); err != nil {
					fmt.Printf("Error processing %s: %v\n", url, err)
				}
			}(url)
		}
		wg.Wait()
	} else {
		fmt.Println("Downloading and extracting non-concurrently...")
		for _, url := range urls {
			fmt.Printf("Downloading: %s\n", url)
			if err := downloadAndExtract(url); err != nil {
				fmt.Printf("Error processing %s: %v\n", url, err)
			}
		}
	}
}

// Parser function

type Event struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Actor     Actor   `json:"actor"`
	Repo      Repo    `json:"repo"`
	Payload   Payload `json:"payload"`
	Public    bool    `json:"public"`
	CreatedAt string  `json:"created_at"`
	Org       *Org    `json:"org,omitempty"` // Org is optional
	PatchUrl  string  `json:"patchURL,omitempty"`
}

type Actor struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	DisplayLogin string `json:"display_login"`
	GravatarID   string `json:"gravatar_id"`
	URL          string `json:"url"`
	AvatarURL    string `json:"avatar_url"`
}

type Repo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Payload struct {
	RepositoryID int          `json:"repository_id,omitempty"`
	PushID       int64        `json:"push_id,omitempty"`
	Size         int          `json:"size,omitempty"`
	DistinctSize int          `json:"distinct_size,omitempty"`
	Ref          string       `json:"ref,omitempty"`
	Head         string       `json:"head,omitempty"`
	Before       string       `json:"before,omitempty"`
	Commits      []Commit     `json:"commits,omitempty"`
	Action       string       `json:"action,omitempty"`
	Number       int          `json:"number,omitempty"`
	PullRequest  *PullRequest `json:"pull_request,omitempty"`
}

type Commit struct {
	SHA      string `json:"sha"`
	Author   Author `json:"author"`
	Message  string `json:"message"`
	Distinct bool   `json:"distinct"`
	URL      string `json:"url"`
}

type Author struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type PullRequest struct {
	URL string `json:"url"`
	// Add additional fields as needed
}

type Org struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	GravatarID string `json:"gravatar_id"`
	URL        string `json:"url"`
	AvatarURL  string `json:"avatar_url"`
}

func MakePatchURL(apiURL string) string {
	webPrefix := strings.Replace(apiURL, "https://api.github.com/repos/", "https://github.com/", 1)
	webCommitURL := strings.Replace(webPrefix, "/commits/", "/commit/", 1) + ".patch"
	return webCommitURL
}

// Function to read and parse the JSON file
//
// Input:
//
// filename (string): A path of the downloaded JSON from Gharchive
//
// Output ([]Event): A parsed slice containing all events
func ParseJSONFile(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var events []Event
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}
		commits := event.Payload.Commits

		if len(commits) > 0 {
			for _, commit := range commits {
				event.PatchUrl = MakePatchURL(commit.URL)
			}
		}

		events = append(events, event)

	}

	return events, nil
}
