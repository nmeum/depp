package main

import (
	"errors"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nmeum/depp/gitweb"
)

const renderScript = "git-render-readme"

func runWithInput(cmd *exec.Cmd, input string) (string, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func renderReadme(repo *gitweb.RepoPage) (template.HTML, error) {
	readme, err := repo.Readme()
	if err != nil {
		return "", err
	}

	fp := filepath.Join(repo.Path, renderScript)
	renderer, err := exec.LookPath(fp)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return template.HTML(`<pre class="raw">` + readme + "</pre>"), nil
		}

		return "", err
	}

	cmd := exec.Command(renderer)
	out, err := runWithInput(cmd, readme)
	if err != nil {
		return "", err
	}

	return template.HTML(out), nil
}
