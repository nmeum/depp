package main

import (
	git "github.com/libgit2/git2go"

	"bytes"
	"errors"
	"strconv"
	"strings"
)

var noReadme = errors.New("no Readme file found")

var readmeNames = []string{
	"README",
	"README.txt",
	"README.markdown",
	"README.md",
}

const nonBreakingSpace string = "&nbsp;"

func getReadme(repo *git.Repository, ref *git.Reference) (string, error) {
	commit, err := repo.LookupCommit(ref.Target())
	if err != nil {
		return "", err
	}
	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	for _, name := range readmeNames {
		entry := tree.EntryByName(name)
		if entry == nil {
			continue
		}

		blob, err := repo.LookupBlob(entry.Id)
		if err != nil {
			return "", err
		}

		return string(blob.Contents()), nil
	}

	return "", noReadme
}

func getCommits(commit *git.Commit, n uint) ([]*git.Commit, error) {
	// TODO: Handle/test the case where `availableCommits < n`

	commits := make([]*git.Commit, n)
	for i := uint(0); i < n; i++ {
		commits[i] = commit
		commit = commit.Parent(0)
	}

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
