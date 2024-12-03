package gitweb

import (
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/hash"
)

var readmeRegex = regexp.MustCompile(`README|(README\.[a-zA-Z0-9]+)`)

func isReadme(fp string) bool {
	name := filepath.Base(fp)
	return readmeRegex.MatchString(name)
}

func readHashFile(r io.Reader) (plumbing.Hash, error) {
	var hashData = make([]byte, hash.HexSize)
	_, err := r.Read(hashData)
	if err != nil {
		return plumbing.Hash{}, err
	}

	// TODO: Consider building our own hex decoder?
	return plumbing.NewHash(string(hashData)), nil
}

func repoTitle(path string) string {
	title := filepath.Base(path)
	ext := strings.LastIndex(title, ".git")
	if ext > 0 {
		title = title[0:ext]
	}

	return title
}
