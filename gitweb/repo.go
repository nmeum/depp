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
	curTree  *object.Tree
	prevTree *object.Tree // may be nil

	git        *git.Repository
	maxCommits uint

	Path  string
	Title string
	URL   string
}

type WalkFunc func(string, *RepoPage) error

const (
	// File name of the git description file.
	descFn = "description"
)

func NewRepo(fp string, cloneURL *url.URL, commits uint) (*Repo, error) {
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}

	r := &Repo{Path: absFp, Title: repoTitle(absFp), maxCommits: commits}
	if cloneURL != nil {
		r.URL = cloneURL.String()
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

	// TODO: Make head a public member of the Repository struct.
	head, err := r.Tip()
	if err != nil {
		return nil, err
	}
	r.curTree, err = head.Tree()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repo) ReadState(fp string) error {
	stateFile, err := os.Open(fp)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	h, err := readHashFile(stateFile)
	if err != nil {
		return err
	}

	r.prevTree, err = r.git.TreeObject(h)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) WriteState(fp string) error {
	stateFile, err := os.Create(fp)
	if err != nil {
		return err
	}

	_, err = stateFile.WriteString(r.curTree.Hash.String())
	if err != nil {
		return err
	}

	return stateFile.Close()
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

func (r *Repo) indexPage() *RepoPage {
	return &RepoPage{
		Repo:        r,
		tree:        r.curTree,
		CurrentFile: RepoFile{mode: filemode.Dir, Path: ""},
	}
}

func (r *Repo) walkTree(fn WalkFunc) error {
	err := fn(".", r.indexPage())
	if err != nil {
		return err
	}

	walker := object.NewTreeWalker(r.curTree, true, nil)
	defer walker.Close()
	for {
		fp, entry, err := walker.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		page, err := r.page(entry.Hash, entry.Mode, fp)
		if err != nil {
			return err
		}
		err = fn(fp, page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) walkDiff(fn WalkFunc) error {
	changes, err := object.DiffTree(r.prevTree, r.curTree)
	if err != nil {
		return err
	}
	patch, err := changes.Patch()
	if err != nil {
		return err
	}

	rebuildDirs := make(map[string]bool)
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if to == nil { // file was removed
			rebuildDirs[filepath.Dir(from.Path())] = true
			err = fn(from.Path(), nil)
			if err != nil {
				return err
			}
			continue
		} else if from == nil { // created a new file
			rebuildDirs[filepath.Dir(to.Path())] = true
		}

		fp := to.Path()
		if isReadme(fp) {
			rebuildDirs[filepath.Dir(fp)] = true
		}

		page, err := r.page(to.Hash(), to.Mode(), fp)
		if err != nil {
			return err
		}
		err = fn(fp, page)
		if err != nil {
			return err
		}
	}

	for dir, _ := range rebuildDirs {
		var page *RepoPage
		if dir == "." {
			page = r.indexPage()
		} else {
			entry, err := r.curTree.FindEntry(dir)
			if err == object.ErrEntryNotFound {
				// If we can't find the directory anymore, then the file
				// contained in it was removed and was the only file in it.
				err = fn(dir, nil)
				if err != nil {
					return err
				}
				continue
			} else if err != nil {
				return err
			}

			page, err = r.page(entry.Hash, entry.Mode, dir)
			if err != nil {
				return err
			}
		}

		err = fn(dir, page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) Walk(fn WalkFunc) error {
	if r.prevTree == nil {
		return r.walkTree(fn)
	} else {
		return r.walkDiff(fn)
	}
}

func (r *Repo) page(hash plumbing.Hash, mode filemode.FileMode, fp string) (*RepoPage, error) {
	page := &RepoPage{
		Repo:        r,
		tree:        nil,
		CurrentFile: RepoFile{mode, filepath.ToSlash(fp)},
	}

	var err error
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
