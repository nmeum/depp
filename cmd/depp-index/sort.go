package main

// byModified sorts Repos by their modified date (latest first).
type byModified []Repo

func (t byModified) Len() int {
	return len(t)
}

func (t byModified) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byModified) Less(i, j int) bool {
	return t[i].Modified.After(t[j].Modified)
}
