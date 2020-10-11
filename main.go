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
	"./tmpl/breadcrumb.tmpl",
}

var (
	commits     = flag.Int("-c", 5, "amount of recent commits to include")
	destination = flag.String("-d", "./www", "output directory for HTML files")
)

var tmpl *template.Template

func walkPages(page *RepoPage) error {
	name := page.CurrentFile.Path
	if page.CurrentFile.Path == "" {
		name = "index"
	} else { // Clear info only used once on the index page
		page.Commits = nil
		page.Readme = ""
	}

	fp := filepath.Join(*destination, name+".html")
	err := os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(fp)
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

func buildTmpl() (*template.Template, error) {
	var err error

	name := filepath.Base(templateFiles[0])
	tmpl := template.New(name)

	funcMap := make(template.FuncMap)
	funcMap["getRelPath"] = getRelPath
	funcMap["decrement"] = decrement
	tmpl = tmpl.Funcs(funcMap)

	tmpl, err = tmpl.ParseFiles(templateFiles[0])
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.ParseFiles(templateFiles[1:]...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
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

	tmpl, err = buildTmpl()
	if err != nil {
		log.Fatal(err)
	}
	err = repo.Walk(walkPages)
	if err != nil {
		log.Fatal(err)
	}
}
