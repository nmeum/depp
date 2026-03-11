package gitweb

import (
	"html/template"

	"github.com/go-git/go-git/v5"
)

const (
	// Name of the depp-specific Git configuration section.
	confSec = "depp"
)

type Config struct {
	HeaderExtra template.HTML
}

func loadConfig(repo *git.Repository) (Config, error) {
	c, err := repo.Config()
	if err != nil {
		return Config{}, err
	}

	raw := c.Raw
	if !raw.HasSection(confSec) {
		return Config{}, nil
	}

	sec := raw.Section(confSec)
	cnf := Config{
		HeaderExtra: template.HTML(sec.Option("extra-head-content")),
	}

	return cnf, nil
}
