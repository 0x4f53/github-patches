# github-patches

does what is says on the tin

this tool helps you scrape commit metadata pushed to github (and grab their patch files) in a given time range using the github api. It returns the patch URL, repo type and more!

# Usage

```bash
cd github-patches-cli/

go run main.go -h

Usage of ./github-patches:
  -from string
        Starting timestamp in '2006-01-02-15' format
  -outputDir string
        the directory to save files to. 'githubCommits/' will be made locally if not specified (default ".githubCommits/")
  -to string
        Ending timestamp in '2006-01-02-15' format
```

# Examples

**Note:** use the `--concurrent` flag to download files concurrently (at the same time!)

### Patches from the last hour

```bash
go run main.go

cat .githubCommits/2024-10-14-15.json

{"id":"42807516937","type":"IssuesEvent","actor":{"id":41898282,"login":"github-actions[bot]","display_login":"github-actions","gravatar_id":"","url":"https://api.github.com/users/github-actions[bot]","avatar_url":"https://avatars.githubusercontent.com/u/41898282?"},"repo":{"id":859486274,"name":"leyu-sports/leyuio","url":"https://api.github.com/repos/leyu-sports/leyuio"},"payload":{"action":"opened","issue":{"url":"https://api.github...
{"id":"42807516940","type":"IssuesEvent","actor":{"id":41898282,"login":"github-actions[bot]","display_login":"github-actions","gravatar_id":"","url":"https://api.github.com/users/github-actions[bot]","avatar_url":"https://avatars.githubusercontent.com/u/41898282?"},"repo":{"id":869028658,"name":"long8guoji/long8ty","url":"https://api.github.com/repos/long8guoji/long8ty"},"payload":{"action":"opened","issue":{"url":"https://api.github.com/repos/long8guoji/long8ty/issues/34119","repository_url":"https://api.github.com/repos/long8guoji/long8ty",...
...
```

### Patches from October 14, 2024 at 3 PM UTC to October 15, 2024 at 2 AM 

```bash
go run main.go --from=2024-10-14-15 --to=2024-10-15-2

cat .githubCommits/2024-10-14-15.json

{"id":"42807516937","type":"IssuesEvent","actor":{"id":41898282,"login":"github-actions[bot]",...
...
```

### Patches for a single timestamp

```bash
go run main.go --from=2024-10-14-15 --to=2024-10-14-15

cat .githubCommits/2024-10-14-15.json

{"id":"42807516937","type":"IssuesEvent","actor":{"id":41898282,"login":"github-actions[bot]",...
...
```

### Patches from October 15, 2024 at 2 AM to October 14, 2024 at 3 PM UTC (in reverse)

```bash
go run main.go --to=2024-10-14-15 --from=2024-10-15-2

cat .githubCommits/2024-10-15-2.json

{"id":"42807516937","type":"IssuesEvent","actor":{"id":41898282,"login":"github-actions[bot]",...
...
```