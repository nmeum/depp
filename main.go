package main

import (
	// "github.com/libgit2/git2go/v28"

	"flag"
	"os"
	"log"
	"path/filepath"
	"html/template"
)

var templateFiles = []string{
	"./tmpl/base.tmpl",
	"./tmpl/style.css",
	"./tmpl/commits.tmpl",
	"./tmpl/tree.tmpl",
	"./tmpl/readme.tmpl",
}

var (
	commits = flag.Int("-c", 5, "amount of recent commits to include")
	destination = flag.String("-d", "./www", "output directory for HTML files")
)

func buildPage(outDir string) error {
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

	err = tmpl.Execute(file, nil)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	err := buildPage(*destination)
	if err != nil {
		log.Fatal(err)
	}
}
