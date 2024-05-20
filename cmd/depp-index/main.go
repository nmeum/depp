package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/nmeum/depp/css"
	"github.com/nmeum/depp/gitweb"
)

type Repo struct {
	Name     string
	Title    string
	Desc     string
	Modified time.Time
}

type Page struct {
	CurPage  int
	NumPages int

	Title string
	Desc  string
	Repos []Repo
}

//go:embed tmpl
var templates embed.FS

var (
	desc  = flag.String("s", "", "short description of git host")
	title = flag.String("t", "depp-index", "page title")
	dest  = flag.String("d", "./www", "output directory for HTML files")
	strip = flag.Bool("x", false, "strip .git extension from repository name in link")
	items = flag.Int("p", 20, "amount of repos per HTML page, a zero value disables pagination")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] REPOSITORY...\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func repoLink(repo *Repo) string {
	if *strip {
		// Return a post-processed repository name without .git
		return repo.Title
	} else {
		// Return the raw file name, potentially including .git
		return repo.Name
	}
}

func pageName(page int) string {
	if page == 0 {
		return "index.html"
	} else {
		return fmt.Sprintf("%d.html", page)
	}
}

func pageRefs(page Page) []int {
	pages := make([]int, page.NumPages)
	for i := 0; i < page.NumPages; i++ {
		pages[i] = i
	}
	return pages
}

func createHTML(page Page, path string) error {
	const name = "base.tmpl"

	html := template.New(name)
	html.Funcs(template.FuncMap{
		"repoLink": repoLink,
		"pageName": pageName,
		"pageRefs": pageRefs,
	})

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

func getRepos(fps []string) ([]Repo, error) {
	repos := make([]Repo, len(fps))
	for i, fp := range fps {
		r, err := gitweb.NewRepo(fp, nil, 0)
		if err != nil {
			return []Repo{}, err
		}

		commit, err := r.Tip()
		if err != nil {
			return []Repo{}, err
		}
		desc, err := r.Description()
		if err != nil {
			return []Repo{}, err
		}

		sig := commit.Committer()
		repos[i] = Repo{
			Name:     filepath.Base(fp),
			Title:    r.Title,
			Desc:     desc,
			Modified: sig.When,
		}
	}

	sort.Sort(byModified(repos))
	return repos, nil
}

func getPages(repos []Repo) []Page {
	var numPages int
	if *items == 0 {
		numPages = 1
	} else {
		numPages = len(repos) / *items
	}

	pages := make([]Page, numPages)
	for i := 0; i < numPages; i++ {
		var maxrepos int
		if *items == 0 {
			maxrepos = len(repos)
		} else {
			maxrepos = min(*items, len(repos))
		}

		pages[i] = Page{
			CurPage:  i,
			NumPages: numPages,
			Title:    *title,
			Desc:     *desc,
			Repos:    repos[0:maxrepos],
		}

		repos = repos[maxrepos:]
	}

	return pages
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
	pages := getPages(repos)

	err = os.MkdirAll(*dest, 0755)
	if err != nil {
		log.Fatal(err)
	}

	err = css.Create(filepath.Join(*dest, "style.css"))
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		fp := filepath.Join(*dest, pageName(page.CurPage))
		err = createHTML(page, fp)
		if err != nil {
			log.Fatal(err)
		}
	}
}
