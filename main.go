package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nmeum/depp/gitweb"
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
	commits     = flag.Uint("c", 5, "amount of recent commits to include")
	gitRawURL   = flag.String("g", "git://localhost", "base URL of Git clone server")
	destination = flag.String("d", "./www", "output directory for HTML files")
)

var tmpl *template.Template

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] REPOSITORY\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func walkPages(page *gitweb.RepoPage) error {
	name := page.CurrentFile.Path
	if isIndexPage(page) {
		name = "index"
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
	funcMap["getLines"] = getLines
	funcMap["getPadding"] = getPadding
	funcMap["relIndex"] = relIndex
	funcMap["isIndexPage"] = isIndexPage
	funcMap["renderReadme"] = renderReadme
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

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	gitURL, err := url.Parse(*gitRawURL)
	if err != nil {
		log.Fatal(err)
	}

	path := flag.Arg(0)
	repo, err := gitweb.NewRepo(path, gitURL, *commits)
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
