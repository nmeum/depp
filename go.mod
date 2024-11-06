module github.com/nmeum/depp

go 1.21

require github.com/libgit2/git2go/v34 v34.0.0-20221005011223-4b14d29c207e

require (
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c // indirect
	golang.org/x/sys v0.0.0-20201204225414-ed752295db88 // indirect
)

replace github.com/libgit2/git2go/v34 v34.0.0-20221005011223-4b14d29c207e => github.com/nmeum/git2go/v34 v34.0.1-0.20241106182505-99577a2a0e18
