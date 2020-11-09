package main

import (
	git "github.com/libgit2/git2go"

	"bytes"
	"strconv"
	"strings"
)

var readmeNames = []string{
	"README",
	"README.txt",
	"README.markdown",
	"README.md",
}

const nonBreakingSpace string = "&nbsp;"

func getCommits(commit *git.Commit, n uint) ([]*git.Commit, error) {
	var i uint

	commits := make([]*git.Commit, n)
	for i = 0; i < n; i++ {
		if commit == nil {
			break
		}

		commits[i] = commit
		commit = commit.Parent(0)
	}

	commits = commits[0:i] // Shrink to appropriate size
	return commits, nil
}

func getRelPath(n int) string {
	var ret string
	for i := 0; i < n; i++ {
		ret += "../"
	}

	if ret == "" {
		return "./"
	} else {
		return ret
	}
}

func getLines(input string) []string {
	// Remove terminating newline (if any)
	if input[len(input)-1] == '\n' {
		input = input[0 : len(input)-1]
	}

	return strings.Split(input, "\n")
}

func getPadding(maxnum int, curnum int) string {
	max := strconv.Itoa(maxnum)
	cur := strconv.Itoa(curnum)

	if len(cur) >= len(max) {
		return ""
	}
	diff := len(max) - len(cur)

	buf := new(bytes.Buffer)
	buf.Grow(diff)

	for i := 0; i < diff; i++ {
		buf.WriteByte(' ')
	}

	return buf.String()
}

func decrement(n int) int {
	return n - 1
}
