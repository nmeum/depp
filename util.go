package main

import (
	git "github.com/libgit2/git2go"
)

func getCommits(commit *git.Commit, n uint) ([]*git.Commit, error) {
	// TODO: Handle/test the case where `availableCommits < n`

	commits := make([]*git.Commit, n)
	for i := uint(0); i < n; i++ {
		commits[i] = commit
		commit = commit.Parent(0)
	}

	return commits, nil
}
