package main

import (
	"flag"

	githubPatches "github.com/0x4f53/github-patches"
)

func main() {
	outputDir := flag.String("outputDir", githubPatches.GithubCacheDir, "the directory to save files to. 'githubCommits/' will be made locally if not specified")
	from := flag.String("from", "", "Starting timestamp in '2006-01-02-15' format")
	to := flag.String("to", "", "Ending timestamp in '2006-01-02-15' format")
	concurrent := flag.Bool("concurrent", false, "Download multiple threads at once")

	flag.Parse()

	githubPatches.GetCommitsInRange(*outputDir, *from, *to, *concurrent)

}
