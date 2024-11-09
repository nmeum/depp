package gitweb

import (
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/go-git/go-billy/v5/osfs"
)

type Repo struct {
	git        *git.Repository
	maxCommits uint

	Path  string
	Title string
	URL   string
}

// File name of the git description file.
const descFn = "description"

func NewRepo(fp string, cloneURL *url.URL, commits uint) (*Repo, error) {
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}

	r := &Repo{Path: absFp, maxCommits: commits}
	r.Title = filepath.Base(absFp)
	if cloneURL != nil {
		r.URL = cloneURL.String()
	}

	ext := strings.LastIndex(r.Title, ".git")
	if ext > 0 {
		r.Title = r.Title[0:ext]
	}

	fs := osfs.New(absFp)
	if _, err := fs.Stat(git.GitDirName); err == nil {
		// If this is not a bare repository, we change into
		// the .git directory so that we can treat it as such.
		fs, err = fs.Chroot(git.GitDirName)
		if err != nil {
			return nil, err
		}
	}

	s := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())

	r.git, err = git.Open(s, fs)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repo) Tip() (*object.Commit, error) {
	head, err := r.git.Head()
	if err != nil {
		return nil, err
	}

	hash := head.Hash()
	commit, err := r.git.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (r *Repo) Walk(fn func(*RepoPage) error) error {
	head, err := r.Tip()
	if err != nil {
		return err
	}

	tree, err := head.Tree()
	if err != nil {
		return err
	}

	// . is not included by tree.Walk()
	indexPage := &RepoPage{
		Repo:        r,
		tree:        tree,
		CurrentFile: RepoFile{filemode.Dir, ""},
	}
	err = fn(indexPage)
	if err != nil {
		return err
	}

	var seen map[plumbing.Hash]bool
	walker := object.NewTreeWalker(tree, true, seen)
	defer walker.Close()
	for {
		fp, entry, err := walker.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		repoFile := RepoFile{
			mode: entry.Mode,
			Path: filepath.ToSlash(fp),
		}
		page, err := r.page(entry.Hash, repoFile)
		if err != nil {
			return err
		}

		err = fn(page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) page(hash plumbing.Hash, rf RepoFile) (*RepoPage, error) {
	var err error
	page := &RepoPage{
		Repo:        r,
		tree:        nil,
		CurrentFile: rf,
	}

	if page.CurrentFile.IsDir() {
		page.tree, err = r.git.TreeObject(hash)
		if err != nil {
			return nil, err
		}
	}

	return page, nil
}

func (r *Repo) Description() (string, error) {
	fp := filepath.Join(r.Path, descFn)

	desc, err := os.ReadFile(fp)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	descText := string(desc)
	return strings.TrimSpace(descText), nil
}
