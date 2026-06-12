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
  - formats:           # note: "formats" (list), not "format" (string)
      - tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

brews:                 # use "brews", not "homebrew" (free GoReleaser v2)
  - name: your-tool
    repository:
      owner: your-github-username
      name: homebrew-tools      # must be named homebrew-<something>
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"
    homepage: https://github.com/your-github-username/your-repo
    description: "Your tool description"
    license: MIT` + reset + `

Key fields:
  • formats            — list of archive formats (not "format")
  • brews              — use this, not "homebrew" (that's GoReleaser Pro)
  • token              — PAT for pushing the formula to your tap repo`,
	},
	{
		title: "Create a Homebrew tap repo",
		content: `Homebrew taps are just GitHub repos named homebrew-<something>.

1. Create a new public repo on GitHub, e.g.:
   ` + dim + `github.com/your-username/homebrew-tools` + reset + `

2. IMPORTANT: Create this repo BEFORE pushing your first release tag.
   If GoReleaser can't find it, the release will partially publish
   (binaries uploaded, formula push fails), leaving a stale release
   that blocks future attempts.

3. The repo can start empty — GoReleaser pushes the formula on release.

4. The name after "homebrew-" becomes the tap name:
   homebrew-tools  →  brew tap your-username/tools`,
	},
	{
		title: "Set up a GitHub token for your tap",
		content: `Actions' built-in GITHUB_TOKEN can publish releases to your main repo,
but it cannot write to a separate repo (your tap). You need a PAT.

1. Go to: GitHub → Settings → Developer settings
             → Personal access tokens → Fine-grained tokens
   (Use fine-grained, not classic — you can scope it to specific repos.)

2. Under "Repository access", select BOTH repos:
   • your-tool repo       (Contents: Read & Write)
   • homebrew-tools repo  (Contents: Read & Write)

   ` + yellow + `Both repos must be selected — omitting the tap repo causes a 403.` + reset + `

3. Add the token as a secret in your MAIN repo only:
   Repo → Settings → Secrets and variables → Actions
   Name it: TAP_GITHUB_TOKEN

   You do NOT need to add it to the tap repo.`,
	},
	{
		title: "Set up GitHub Actions for automated releases",
		content: `Create .github/workflows/release.yml:

` + dim + `name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write    # required — don't rely on the repo-level setting

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0    # required for changelog generation

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

` + yellow + `permissions: contents: write must be declared explicitly in the workflow.
Setting it in the repo UI alone is not reliable.` + reset,
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
	{
		title: "Cut your first release",
		content: `Tag a commit and push the tag — Actions does the rest.

  ` + dim + `$ git tag v0.1.0
$ git push origin v0.1.0` + reset + `

GoReleaser will:
  1. Build binaries for each platform
  2. Create a GitHub release with the binaries attached
  3. Generate a changelog from your commits
  4. Push a formula file to your homebrew-tools repo

Once it completes, anyone can install your tool with:
  ` + dim + `$ brew tap your-username/tools
$ brew install your-tool` + reset,
	},
	{
		title: "Recovering from a failed release",
		content: `If a release run fails after binaries are already uploaded, the next
run will error with "already_exists" on the assets. Fix it by deleting
the partial release before retriggering:

  ` + dim + `$ gh release delete v0.1.0 --repo your-username/your-tool --yes
$ git push origin --delete v0.1.0
$ git tag -d v0.1.0
$ git tag v0.1.0
$ git push origin v0.1.0` + reset + `

This clears the stale release and lets GoReleaser start clean.

Install the GitHub CLI if you don't have it:
  ` + dim + `$ brew install gh && gh auth login` + reset,
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
	fmt.Printf("%sIf anything goes wrong, step %d covers recovery.%s\n", dim, len(steps), reset)

	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) == "--summary" {
		fmt.Printf("\n%sSummary of steps:%s\n", bold, reset)
		for i, s := range steps {
			fmt.Printf("  %d. %s\n", i+1, s.title)
		}
	}

	fmt.Println()
}
