package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	git "github.com/libgit2/git2go/v30"
	"github.com/nmeum/depp/css"
)

type Repo struct {
	Name     string
	Desc     string
	Modified time.Time
}

type Page struct {
	Title  string
	Desc   string
	Repos  []Repo
}

//go:embed tmpl
var templates embed.FS

var (
	desc   = flag.String("s", "", "short description of git host")
	title  = flag.String("t", "depp-index", "page title")
	dest   = flag.String("d", "./www", "output directory for HTML files")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] REPOSITORY...\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func createHTML(page Page, path string) error {
	const name = "base.tmpl"
	html := template.New(name)

	tmpl, err := html.ParseFS(templates, "tmpl/*.tmpl")
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		return err
	}

	return nil
}

func getDescription(fp string) (string, error) {
	// XXX: Code duplication with ./gitweb/repo.go
	descFp := filepath.Join(fp, "description")

	desc, err := os.ReadFile(descFp)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return string(desc), nil
}

func getRepos(fps []string) ([]Repo, error) {
	repos := make([]Repo, len(fps))
	for i, fp := range fps {
		r, err := git.OpenRepository(fp)
		if err != nil {
			return []Repo{}, err
		}

		head, err := r.Head()
		if err != nil {
			return []Repo{}, err
		}

		oid := head.Target()
		commit, err := r.LookupCommit(oid)
		if err != nil {
			return []Repo{}, err
		}
		desc, err := getDescription(fp)
		if err != nil {
			return []Repo{}, err
		}

		sig := commit.Committer()
		repos[i] = Repo{
			Name:     filepath.Base(fp),
			Desc:     desc,
			Modified: sig.When,
		}
	}

	sort.Sort(byModified(repos))
	return repos, nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.Lshortfile)
	if flag.NArg() == 0 {
		usage()
	}

	repos, err := getRepos(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	page := Page{
		Title:  *title,
		Desc:   *desc,
		Repos:  repos,
	}

	err = css.Create(filepath.Join(*dest, "style.css"))
	if err != nil {
		log.Fatal(err)
	}
	err = createHTML(page, filepath.Join(*dest, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
}
