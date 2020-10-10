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

func buildPage(outDir string, page *RepoPage) error {
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

	err = tmpl.Execute(file, page)
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
	repo, err := NewRepo(path)
	if err != nil {
		log.Fatal(err)
	}

	head, err := repo.git.Head()
	if err != nil {
		log.Fatal(err)
	}
	page, err := repo.GetPage(head, "")
	if err != nil {
		log.Fatal(err)
	}

	err = buildPage(*destination, page)
	if err != nil {
		log.Fatal(err)
	}
}
