package main

import (
	git "github.com/libgit2/git2go"

	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// RepoFile represents information for a single file/blob.
type RepoFile struct {
	Path  string // Slash separated path
	Type  git.ObjectType
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

func (f *RepoFile) IsDir() bool {
	return f.Type == git.ObjectTree
}

func (f *RepoFile) PathElements() []string {
	return strings.SplitN(f.Path, "/", -1)
}

// Repo represents information required per repository.
type Repo struct {
	git        *git.Repository
	path       string
	numCommits uint

	Title  string
	URL    string
	Readme string
}

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	tree   *git.Tree
	commit *git.Commit

	CurrentFile RepoFile
	Commits     []*git.Commit
}

func NewRepo(fp string, gitServer *url.URL, commits uint) (*Repo, error) {
	var err error
	r := &Repo{path: fp}

	r.git, err = git.OpenRepository(fp)
	if err != nil {
		return nil, err
	}

	r.Title = filepath.Base(fp)
	r.URL = gitServer.String()

	head, err := r.git.Head()
	if err != nil {
		return nil, err
	}
	r.Readme, err = getReadme(r.git, head)
	if err == noReadme || len(r.Readme) == 0 {
		r.Readme = ""
	} else if err != nil {
		return nil, err
	}

	r.numCommits = commits
	return r, nil
}

func (r *Repo) Walk(fn func(*RepoPage) error) error {
	head, err := r.git.Head()
	if err != nil {
		return err
	}

	indexPage, err := r.GetPage(head, "")
	if err != nil {
		return err
	}

	var ret error
	indexPage.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root == "" {
			err = fn(indexPage)
			if err != nil {
				ret = err
				return -1
			}
		}

		// TODO: Explizit handling for git submodules needed
		if e.Type == git.ObjectCommit {
			return 1 // Skip git submodules
		}

		fp := filepath.Join(root, e.Name)
		page, err := r.GetPage(head, fp)
		if err != nil {
			ret = err
			return -1
		}

		ret = fn(page)
		if err != nil {
			ret = err
			return -1
		}

		return 0
	})

	return ret
}

func (r *Repo) GetPage(ref *git.Reference, fp string) (*RepoPage, error) {
	var err error
	page := &RepoPage{Repo: *r}

	oid := ref.Target()
	page.commit, err = r.git.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	page.tree, err = page.commit.Tree()
	if err != nil {
		return nil, err
	}

	// TODO: Find out how to retrieve the TreeEntry for /
	page.CurrentFile = RepoFile{Path: filepath.ToSlash(fp)}
	if page.CurrentFile.Path != "" {
		entry, err := page.tree.EntryByPath(fp)
		if err != nil {
			return nil, err
		}
		page.CurrentFile.Type = entry.Type

		if page.CurrentFile.IsDir() {
			page.tree, err = r.git.LookupTree(entry.Id)
			if err != nil {
				panic(err)
				return nil, err
			}
		}
	} else {
		page.CurrentFile.Type = git.ObjectTree
	}

	page.Commits, err = getCommits(page.commit, page.numCommits)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (r *RepoPage) Files() ([]RepoFile, error) {
	var ret error
	var entries []RepoFile
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != "" {
			return 1 /* Skip passed entry */
		}

		basepath := filepath.Base(r.CurrentFile.Path)
		relpath := filepath.Join(basepath, e.Name)

		file := RepoFile{
			Path:  filepath.ToSlash(relpath),
			Type:  e.Type,
		}

		entries = append(entries, file)
		return 0
	})
	if ret != nil {
		return nil, ret
	}

	return entries, nil
}

func (r *RepoPage) GetBlob(fp string) (string, error) {
	entry, err := r.tree.EntryByPath(fp)
	if err != nil {
		return "", err
	}

	oid := entry.Id
	blob, err := r.git.LookupBlob(oid)
	if err != nil {
		return "", err
	}

	return string(blob.Contents()), nil
}
