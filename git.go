package main

import (
	git "github.com/libgit2/git2go"

	"errors"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// RepoFile represents information for a single file/blob.
type RepoFile struct {
	Path string // Slash separated path
	Type git.ObjectType
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

func (f *RepoFile) FilePath() string {
	return filepath.FromSlash(f.Path)
}

func (f *RepoFile) IsDir() bool {
	return f.Type == git.ObjectTree
}

func (f *RepoFile) IsSubmodule() bool {
	return f.Type == git.ObjectCommit
}

func (f *RepoFile) PathElements() []string {
	return strings.SplitN(f.Path, "/", -1)
}

// Repo represents information required per repository.
type Repo struct {
	git        *git.Repository
	path       string
	numCommits uint

	Title string
	URL   string
}

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	tree   *git.Tree
	commit *git.Commit

	CurrentFile RepoFile
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
			Path: filepath.ToSlash(relpath),
			Type: e.Type,
		}

		entries = append(entries, file)
		return 0
	})
	if ret != nil {
		return nil, ret
	}

	sort.Sort(byType(entries))
	return entries, nil
}

func (r *RepoPage) GetReadme() (string, error) {
	head, err := r.git.Head()
	if err != nil {
		return "", err
	}

	commit, err := r.git.LookupCommit(head.Target())
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

		blob, err := r.git.LookupBlob(entry.Id)
		if err != nil {
			return "", err
		}

		return string(blob.Contents()), nil
	}

	return "", nil
}

func (r *RepoPage) GetCommits() ([]*git.Commit, error) {
	var i uint

	commit := r.commit
	commits := make([]*git.Commit, r.numCommits)

	for i = 0; i < r.numCommits; i++ {
		if commit == nil {
			break
		}

		commits[i] = commit
		commit = commit.Parent(0)
	}

	commits = commits[0:i] // Shrink to appropriate size
	return commits, nil

}

func (r *RepoPage) GetBlob(file *RepoFile) (string, error) {
	if file.Type != git.ObjectBlob {
		return "", errors.New("given RepoFile is not a blob")
	}
	fp := file.FilePath()

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

func (r *RepoPage) GetSubmodule(file *RepoFile) (*git.Submodule, error) {
	if !file.IsSubmodule() {
		return nil, errors.New("given RepoFile is not a submodule")
	}
	fp := file.FilePath()

	submodule, err := r.git.Submodules.Lookup(fp)
	if err != nil {
		return nil, err
	}

	return submodule, nil
}
