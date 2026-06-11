package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	green  = "\033[32m"
	dim    = "\033[2m"
)

type step struct {
	title   string
	content string
}

var steps = []step{
	{
		title: "Install GoReleaser",
		content: `GoReleaser builds and publishes your Go binaries.

Install it with Homebrew:
  ` + dim + `$ brew install goreleaser` + reset + `

Or via Go:
  ` + dim + `$ go install github.com/goreleaser/goreleaser/v2@latest` + reset + `

Verify:
  ` + dim + `$ goreleaser --version` + reset,
	},
	{
		title: "Initialize the GoReleaser config",
		content: `Run this in your repo root to generate a starter .goreleaser.yaml:

  ` + dim + `$ goreleaser init` + reset + `

This creates a .goreleaser.yaml with sensible defaults.
You'll customize it in the next steps.`,
	},
	{
		title: "Configure your .goreleaser.yaml",
		content: `A minimal config for a Go CLI with Homebrew support:

` + dim + `version: 2

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: your-tool
    repository:
      owner: your-github-username
      name: homebrew-tap     # must be named homebrew-<something>
    homepage: https://github.com/your-github-username/your-repo
    description: "Your tool description"` + reset + `

Key fields:
  • goos / goarch  — platforms to build for
  • brews          — tells GoReleaser to update your Homebrew tap`,
	},
	{
		title: "Create a Homebrew tap repo",
		content: `Homebrew taps are just GitHub repos named homebrew-<something>.

1. Create a new repo on GitHub, e.g.:
   ` + dim + `github.com/your-username/homebrew-tap` + reset + `

2. It can start empty — GoReleaser will push the formula file
   on each release.

3. Make sure the repo is public (Homebrew requires public taps
   for the default install flow).`,
	},
	{
		title: "Set up a GitHub token",
		content: `GoReleaser needs a token to publish releases and push to your tap.

1. Go to: github.com → Settings → Developer settings
             → Personal access tokens → Fine-grained tokens

2. Grant permissions:
   • Contents: Read & Write  (for your main repo)
   • Contents: Read & Write  (for your homebrew-tap repo)

3. Export the token in your shell or CI:
   ` + dim + `$ export GITHUB_TOKEN=ghp_your_token_here` + reset + `

For GitHub Actions, store it as a repository secret named
GITHUB_TOKEN (Actions provides one automatically with enough
permissions for releases, but you'll need a PAT for writing
to a separate tap repo).`,
	},
	{
		title: "Set up GitHub Actions for automated releases",
		content: `Create .github/workflows/release.yml:

` + dim + `name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          TAP_GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}` + reset + `

The fetch-depth: 0 is required — GoReleaser reads your full
git history to generate changelogs.`,
	},
	{
		title: "Cut your first release",
		content: `Tag a commit and push the tag — Actions does the rest.

  ` + dim + `$ git tag v0.1.0
$ git push origin v0.1.0` + reset + `

GoReleaser will:
  1. Build binaries for each platform
  2. Create a GitHub release with the binaries attached
  3. Generate a changelog from your commits
  4. Push a formula file to your homebrew-tap repo

Once it completes, anyone can install your tool with:
  ` + dim + `$ brew tap your-username/tap
$ brew install your-tool` + reset,
	},
	{
		title: "Test your release locally first",
		content: `Before pushing a tag, do a dry run:

  ` + dim + `$ goreleaser release --snapshot --clean` + reset + `

  • --snapshot  builds without publishing (no tag required)
  • --clean     removes the dist/ folder first

Inspect the output in dist/ to confirm binaries look right.
When you're satisfied, push a real tag to trigger the full release.`,
	},
}

func pause(scanner *bufio.Scanner) {
	fmt.Printf("\n%s[press enter to continue]%s ", dim, reset)
	scanner.Scan()
}

func header(current, total int, title string) {
	fmt.Printf("\n%s%s── Step %d of %d: %s %s%s\n\n",
		bold, cyan, current, total, title, reset, reset)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	total := len(steps)

	fmt.Printf("\n%s%sGoReleaser + Homebrew Setup Walkthrough%s\n", bold, green, reset)
	fmt.Printf("%s%d steps · press enter to advance · ctrl+c to quit%s\n", dim, total, reset)

	for i, s := range steps {
		pause(scanner)
		header(i+1, total, s.title)
		fmt.Println(s.content)
	}

	fmt.Printf("\n%s%s✓ All done!%s\n", bold, green, reset)
	fmt.Println("Push your first tag and watch GoReleaser do its thing.")

	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) == "--summary" {
		fmt.Printf("\n%sSummary of steps:%s\n", bold, reset)
		for i, s := range steps {
			fmt.Printf("  %d. %s\n", i+1, s.title)
		}
	}

	fmt.Println()
}
