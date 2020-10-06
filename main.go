package main

import (
	// "github.com/libgit2/git2go/v28"

	"flag"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

type GitRepo struct {
	Title    string
	URL      string
	Branches []string
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
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	repo := GitRepo{
		Title: "Some Repository",
		URL:   "git://git.8pit.net",
		Branches: []string{
			"master",
			"next",
			"feature/foobar",
			"feature/barfoo",
		},
	}

	err := buildPage(*destination, &repo)
	if err != nil {
		log.Fatal(err)
	}
}
