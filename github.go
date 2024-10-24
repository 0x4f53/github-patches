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

	"github.com/0x4f53/textsubs"
)

var GithubCacheDir = ".githubCommits/"

func GetISO8601Timestamps(from, to string) []string {

	var timestamps []string

	if isValidTimestamp(from) && !isValidTimestamp(to) {
		fmt.Println(timeOverflowError)
		os.Exit(-1)
	}

	if (from == "" && to != "") || (from != "" && to == "") {
		return timestamps
	}

	if (!isValidTimestamp(from) || !isValidTimestamp(to)) && (from != "" && to != "") {
		fmt.Println(timestampFormatError)
		os.Exit(-1)
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
		os.Exit(-1)
	}

	toTime, err := time.Parse("2006-01-02-15", to)
	if err != nil {
		fmt.Println("Invalid 'to' timestamp format. Use dd-mm-yyyy-H.")
		os.Exit(-1)
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

	fileName := filepath.Base(url)
	gzFileName := GithubCacheDir + fileName
	jsonFileName := gzFileName[:len(gzFileName)-3]

	if _, err := os.Stat(jsonFileName); err == nil {
		fmt.Printf("JSON file %s already exists. Continuing...\n", jsonFileName)
		return nil
	}

	cachedFiles, _ := listCachedFiles()

	gzFileExists := false
	for _, item := range cachedFiles {
		if item == gzFileName {
			gzFileExists = true
			break
		}
	}

	if gzFileExists {
		fmt.Printf("Extracting %s to %s\n", gzFileName, jsonFileName)

		gzFile, err := os.Open(gzFileName)
		if err != nil {
			return fmt.Errorf("failed to open gz file: %w", err)
		}
		defer gzFile.Close()

		gzReader, err := gzip.NewReader(gzFile)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()

		jsonFile, err := os.Create(jsonFileName)
		if err != nil {
			return fmt.Errorf("failed to create json file: %w", err)
		}
		defer jsonFile.Close()

		if _, err = io.Copy(jsonFile, gzReader); err != nil {
			return fmt.Errorf("failed to write json file: %w", err)
		}

		fmt.Printf("Extracted to %s\n", jsonFileName)

		os.Remove(gzFileName)
	} else {
		fmt.Printf("Downloading %s\n", url)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()

		outFile, err := os.Create(gzFileName)
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

		fmt.Printf("Extracting %s to %s\n", gzFileName, jsonFileName)

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

		jsonFile, err := os.Create(jsonFileName)
		if err != nil {
			return fmt.Errorf("failed to create json file: %w", err)
		}
		defer jsonFile.Close()

		if _, err = io.Copy(jsonFile, gzReader); err != nil {
			return fmt.Errorf("failed to write json file: %w", err)
		}

		fmt.Printf("Downloaded, saved and extracted to %s\n", jsonFileName)

		os.Remove(gzFileName)
	}

	return nil
}

func makeDir(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	return nil
}

func listCachedFiles() ([]string, error) {
	files, err := os.ReadDir(GithubCacheDir)
	if err != nil {
		return nil, err
	}

	var fileNames []string

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
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
		fmt.Println("Downloading and extracting multiple files concurrently...")
		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
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
			if err := downloadAndExtract(url); err != nil {
				fmt.Printf("Error processing %s: %v\n", url, err)
			}
		}
	}
}

func MakePatchURL(apiURL string) string {
	webPrefix := strings.Replace(apiURL, "https://api.github.com/repos/", "https://github.com/", 1)
	webCommitURL := strings.Replace(webPrefix, "/commits/", "/commit/", 1) + ".patch"
	return webCommitURL
}

type Actor struct {
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	ID         int    `json:"id"`
	Login      string `json:"login"`
	URL        string `json:"url"`
}

type Author struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Commit struct {
	Author   Author `json:"author"`
	Distinct bool   `json:"distinct"`
	Message  string `json:"message"`
	Sha      string `json:"sha"`
	URL      string `json:"url"`
	PatchURL string `json:"patch_url"`
}

type Payload struct {
	Before       string   `json:"before"`
	Commits      []Commit `json:"commits"`
	DistinctSize int      `json:"distinct_size"`
	Head         string   `json:"head"`
	PushID       int      `json:"push_id"`
	Ref          string   `json:"ref"`
	Size         int      `json:"size"`
	Action       string   `json:"action"`
	Gist         Gist     `json:"gist"`
}

type Repo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Event struct {
	Domains   []string  `json:"domains"`
	Actor     Actor     `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Payload   Payload   `json:"payload"`
	Public    bool      `json:"public"`
	Repo      Repo      `json:"repo"`
	Type      string    `json:"type"`
}

type User struct {
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	ID         int    `json:"id"`
	Login      string `json:"login"`
	URL        string `json:"url"`
}

type Gist struct {
	Comments    int                    `json:"comments"`
	CreatedAt   time.Time              `json:"created_at"`
	Description string                 `json:"description"`
	Files       map[string]interface{} `json:"files"` // Map for files, since it's an empty object here
	GitPullURL  string                 `json:"git_pull_url"`
	GitPushURL  string                 `json:"git_push_url"`
	HtmlURL     string                 `json:"html_url"`
	ID          string                 `json:"id"`
	Public      bool                   `json:"public"`
	UpdatedAt   time.Time              `json:"updated_at"`
	URL         string                 `json:"url"`
	User        User                   `json:"user"`
}

// Function to read and parse the JSON file
//
// Input:
//
// filenames ([]string): A list of filenames containing gharchive files. You can use filepath.Walk() here
//
// Output ([]Event): A parsed slice containing all push events
func ParseGitHubCommits(filename string) ([]Event, error) {

	var events []Event

	file, err := os.Open(filename)
	if err != nil {
		return events, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()

		var event Event
		err := json.Unmarshal(line, &event)
		if err != nil {
			return nil, err
		}

		commits := event.Payload.Commits

		for index := range commits {
			commits[index].PatchURL = MakePatchURL(commits[index].URL)
		}

		capturedDomains, _ := textsubs.DomainsOnly(string(line), false)
		capturedDomains = removeBlacklistedDomains(capturedDomains)
		event.Domains = append(event.Domains, capturedDomains...)

		events = append(events, event)

	}

	return events, nil
}
func removeBlacklistedDomains(domains []string) []string {

	var blacklist = []string{
		"github.dev",
		"github.com",
		"githubusercontent.com",
		"gravatar.com",
		"akamai.net",
	}

	blacklistMap := make(map[string]bool)
	for _, blacklistedDomain := range blacklist {
		blacklistMap[blacklistedDomain] = true
	}

	var validDomains []string

	for _, domain := range domains {
		if !blacklistMap[domain] {
			validDomains = append(validDomains, domain)
		}
	}

	return validDomains
}
