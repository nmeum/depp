package main

import (
	"flag"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

var templateFiles = []string{
	"./tmpl/base.tmpl",
	"./tmpl/style.css",
	"./tmpl/commits.tmpl",
	"./tmpl/tree.tmpl",
	"./tmpl/readme.tmpl",
	"./tmpl/blob.tmpl",
}

var (
	commits     = flag.Int("-c", 5, "amount of recent commits to include")
	destination = flag.String("-d", "./www", "output directory for HTML files")
)

var tmpl *template.Template

func walkPages(page *RepoPage) error {
	var name string
	if page.CurrentFile == "" {
		name = "index"
	} else {
		name = page.CurrentFile
	}
	name += ".html"

	fp := filepath.Join(*destination, name)
	err := os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		log.Fatal(err)
		return err
	}

	file, err := os.Create(fp)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	// If this is not the index, remove some information
	// TODO: Make sure this information is not calculated in the first place
	if page.CurrentFile != "" {
		page.Commits = nil
		page.Readme = ""
	}

	err = tmpl.Execute(file, page)
	if err != nil {
		log.Fatal(err)
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
	repo, err := NewRepo(path)
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err = template.ParseFiles(templateFiles...)
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Walk(walkPages)
	if err != nil {
		log.Fatal(err)
	}
}
