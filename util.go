package main

import (
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

func isIndexPage(page *RepoPage) bool {
	return page.CurrentFile.Path == ""
}

func decrement(n int) int {
	return n - 1
}
