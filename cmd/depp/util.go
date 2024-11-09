package main

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/nmeum/depp/gitweb"
)

func summarize(msg string) string {
	newline := strings.IndexByte(msg, '\n')
	if newline != -1 {
		msg = msg[0:newline]
	}
	return msg
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

func getLines(data string) []string {
	if len(data) == 0 {
		return []string{""} // empty file
	}

	// Remove terminating newline (if any)
	if data[len(data)-1] == '\n' {
		data = data[0 : len(data)-1]
	}

	return strings.Split(data, "\n")
}

func padNumber(maxnum int, curnum int) template.HTML {
	const nonBreakingSpace string = "&nbsp;"
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
