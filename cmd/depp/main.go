package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nmeum/depp/css"
	"github.com/nmeum/depp/gitweb"
)

//go:embed tmpl
var templates embed.FS

var (
	commits     = flag.Uint("c", 5, "amount of recent commits to include")
	force       = flag.Bool("f", false, "force rebuilding of all HTML files")
	gitURL      = flag.String("u", "", "clone URL for the Git repository")
	destination = flag.String("d", "./www", "output directory for HTML files")
	verbose     = flag.Bool("v", false, "print the name of each changed file")
)

var tmpl *template.Template

const (
	// Name of file used to record the hash of the generated tree object.
	stateFile = ".tree"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] REPOSITORY\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func walkPages(name string, page *gitweb.RepoPage) error {
	if *verbose {
		fmt.Println(name)
	}

	dest := filepath.Join(*destination, name+".html")
	if page == nil { // file was removed
		err := os.Remove(dest)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		// In case name refers to a (now empty) directory:
		os.Remove(filepath.Join(*destination, name))

		return nil
	} else if isIndexPage(page) {
		dest = filepath.Join(*destination, "index.html")
	}

	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(dest)
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

func buildHTML() (*template.Template, error) {
	var err error

	const name = "base.tmpl"
	tmpl := template.New(name)

	funcMap := template.FuncMap{
		"summarize":    summarize,
		"getRelPath":   getRelPath,
		"increment":    increment,
		"decrement":    decrement,
		"getLines":     getLines,
		"padNumber":    padNumber,
		"relIndex":     relIndex,
		"isIndexPage":  isIndexPage,
		"renderReadme": renderReadme,
	}
	tmpl = tmpl.Funcs(funcMap)

	tmpl, err = tmpl.ParseFS(templates, "tmpl/*.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func generate(repo *gitweb.Repo) error {
	var err error
	tmpl, err = buildHTML()
	if err != nil {
		return err
	}
	err = repo.Walk(walkPages)
	if err != nil {
		return err
	}

	cssPath := filepath.Join(*destination, "style.css")
	_, err = os.Stat(cssPath)
	if *force || errors.Is(err, os.ErrNotExist) {
		err = css.Create(cssPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.Lshortfile)
	if flag.NArg() != 1 {
		usage()
	}

	gitURL, err := url.Parse(*gitURL)
	if err != nil {
		log.Fatal(err)
	}

	path := flag.Arg(0)
	statePath := filepath.Join(*destination, stateFile)

	repo, err := gitweb.NewRepo(path, gitURL, *commits)
	if err != nil {
		log.Fatal(err)
	}
	if !*force {
		err = repo.ReadState(statePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	}

	err = generate(repo)
	if err != nil {
		log.Fatal(err)
	}
	err = repo.WriteState(statePath)
	if err != nil {
		log.Fatal(err)
	}
}
