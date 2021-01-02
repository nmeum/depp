package main

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/nmeum/depp/gitweb"
)

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

func padNumber(maxnum int, curnum int) template.HTML {
	digitsReq := func(n int) int {
		r := 1
		for n/10 > 0 {
			n /= 10
			r++
		}
		return r
	}

	max := digitsReq(maxnum)
	cur := digitsReq(curnum)

	diff := max - cur
	if diff == 0 {
		return ""
	}

	buf := new(bytes.Buffer)
	buf.Grow(diff)

	for i := 0; i < diff; i++ {
		buf.WriteString(nonBreakingSpace)
	}

	return template.HTML(buf.String())
}

func relIndex(file *gitweb.RepoFile) string {
	elems := file.PathElements()
	return getRelPath(len(elems) - 1)
}

func isIndexPage(page *gitweb.RepoPage) bool {
	return page.CurrentFile.Path == ""
}

func increment(n int) int {
	return n + 1
}

func decrement(n int) int {
	return n - 1
}
