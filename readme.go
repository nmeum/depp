package main

import (
	"io"
	"os"
	"os/exec"
	"html/template"
	"path/filepath"

	"github.com/nmeum/depp/gitweb"
)

const renderScript = "git-readme-render"

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
	fp := filepath.Join(repo.Path, renderScript)
	renderer, err := exec.LookPath(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	cmd := exec.Command(renderer)
	readme, err := repo.Readme()
	if err != nil {
		return "", err
	}

	out, err := runWithInput(cmd, readme)
	if err != nil {
		return "", err
	}

	return template.HTML(out), nil
}
