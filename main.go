package main

import (
	git "github.com/libgit2/git2go"

	"flag"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

type GitRepo struct {
	Title     string
	URL       string
	CurBranch string
	Branches  []string

	// Optional fields
	Commits []*git.Commit
	Tree    []os.FileInfo
	Readme  string
}

var templateFiles = []string{
	"./tmpl/base.tmpl",
	"./tmpl/style.css",
	"./tmpl/commits.tmpl",
	"./tmpl/tree.tmpl",
	"./tmpl/readme.tmpl",
}

var (
	commits     = flag.Int("-c", 5, "amount of recent commits to include")
	destination = flag.String("-d", "./www", "output directory for HTML files")
)

var repo *git.Repository

func buildPage(outDir string, repo *GitRepo) error {
	tmpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return err
	}

	indexPath := filepath.Join(outDir, "index.html")
	file, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, repo)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var err error
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	if flag.NArg() != 1 {
		os.Exit(1)
	}

	path := flag.Arg(0)
	repo, err = git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}

	var files []os.FileInfo
	for _, fp := range []string{"tmpl", "main.go", "README.md"} {
		stat, err := os.Stat(fp)
		if err != nil {
			log.Fatal(err)
		}

		files = append(files, stat)
	}

	readme := `
# Example Readme

This is an an example Readme file.
`
	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		log.Fatal(err)
	}

	repo := GitRepo{
		Title:     "Some Repository",
		URL:       "git://git.8pit.net",
		Branches:  []string{"foo", "bar", "baz"},
		CurBranch: "bar",
		Commits:   []*git.Commit{commit},
		Tree:      files,
		Readme:    readme,
	}

	err = buildPage(*destination, &repo)
	if err != nil {
		log.Fatal(err)
	}
}
