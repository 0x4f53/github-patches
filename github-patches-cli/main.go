package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	githubPatches "github.com/0x4f53/github-patches"
)

func main() {
	outputDir := flag.String("outputDir", githubPatches.GithubCacheDir, "the directory to save files to. 'githubCommits/' will be made locally if not specified")
	from := flag.String("from", "", "Starting timestamp in '2006-01-02-15' format")
	to := flag.String("to", "", "Ending timestamp in '2006-01-02-15' format")
	concurrent := flag.Bool("concurrent", false, "Download multiple threads at once")
	gists := flag.Bool("gists", false, "Get the last 100 gists on GitHub")

	verbose := flag.Bool("verbose", false, "Print output to console")

	flag.Parse()

	if *gists {
		last100Gists := githubPatches.GetLast100Gists()
		gistData, _ := githubPatches.ParseGistData(last100Gists)

		fmt.Println(len(gistData))

		//for _, gist := range gistData {
		//	data, _ := json.Marshal(gist)
		//	fmt.Println(string(data))
		//}

		return

	}

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
			data, _ := githubPatches.ParseGitHubCommits(chunk)
			for _, pushEvent := range data {
				data, _ := json.Marshal(pushEvent)
				fmt.Println(string(data))
			}
		}

	}

}
