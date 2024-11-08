package gitweb

// byType sorts RepoFiles by their object type (directories first).
type byType []RepoFile

func (t byType) Len() int {
	return len(t)
}

func (t byType) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byType) Less(i, j int) bool {
	return t[i].IsDir && !t[j].IsDir
}
