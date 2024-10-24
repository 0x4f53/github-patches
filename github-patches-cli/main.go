package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x4f53/github-patches"
)

func main() {
	outputDir := flag.String("outputDir", githubPatches.GithubCacheDir, "the directory to save files to. 'githubCommits/' will be made locally if not specified")
	from := flag.String("from", "", "Starting timestamp in '2006-01-02-15' format")
	to := flag.String("to", "", "Ending timestamp in '2006-01-02-15' format")
	concurrent := flag.Bool("concurrent", false, "Download multiple threads at once")

	verbose := flag.Bool("verbose", false, "Print output to console")

	flag.Parse()

	githubPatches.GetCommitsInRange(*outputDir, *from, *to, *concurrent)

	if *verbose {
		var chunks []string
		filepath.Walk(githubPatches.GithubCacheDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				chunks = append(chunks, path)
			}
			return nil
		})

		for _, chunk := range chunks {
			data, _ := githubPatches.ParsePushEvents(chunk)
			for _, pushEvent := range data {
				data, _ := json.Marshal(pushEvent)
				fmt.Println(string(data))
			}
		}
	}

}
