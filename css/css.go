package css

import (
	"embed"
	"html/template"
	"os"
)

//go:embed tmpl
var templates embed.FS

func Create(path string) error {
	const name = "base.tmpl"
	stylesheet := template.New(name)

	t, err := stylesheet.ParseFS(templates, "tmpl/*.tmpl")
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = t.Execute(file, nil)
	if err != nil {
		return err
	}

	return nil
}
