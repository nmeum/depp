## README

No frills static page generator for Git repositories.

### Motivation

Contrary to existing static [page][stagit website] [generator][depp website]
approaches, this software does not strive to be a fully featured git browser
for the web. Instead, the idea is to provide a quick overview for a given
repository, thereby allowing users to decide whether it is interesting enough
to be cloned. As such, this software does intentionally not provide a web
frontend for existing tools like `git-log(1)`, `git-blame(1)`, et cetera. If
more information is needed, the user should simply clone the repository and use
`git(1)` as usual.

Further, this page generator is entirely written in Go using the pure Go Git
library [go-git][go-git github] instead of [libgit2][libgit2 website] to
interact with Git repositories. Thereby, allowing the implementation to be
compiled as a single statically linked binary while also embedding all HTML and
CSS files into the binary through Go's [embed][go embed] package.

### Features

* Easy to deploy, everything is backed into the binaries (no external dependencies).
* Blazingly fast as it only rebuilds files that changed since the last invocation.
* Simple and mobile-friendly web design which can be easily customized.

### Status

I use this for my own [Git server][8pit git]. I am presently not aware of any
bugs and consider it mostly finished software. As I use it myself, I am
committed to maintaining it for the foreseeable future.

### Installation

Installation requires a [Go toolchain][go website]. Assuming a supported Go is
available, the software can be installed either via `go install` or `make`.
Both methods will install two binaries: `depp` for generating HTML files on a
per-repository basis and `depp-index` which can optionally be used to generate
an HTML index page for listing all hosted git repositories. Both commands are
further described in the provided man page, a usage example is provided below.

#### go install

To install to the program using `go install` run the following command:

	$ go install github.com/nmeum/depp/...@latest

Note that this will not install additional documentation files, e.g. man pages.

#### make

Clone the repository manually and ran the following commands:

	$ make
	$ sudo make install

This is the preferred method when packaging this software for a distribution.

### Usage

Assuming you have a web server serving files located at
`/var/www/htdocs/git.example.org`, you want 10 commits on the index
page, and the repository can be cloned via `git://example.org/foo.git`:

	$ depp -c 10 -u git://example.org/foo.git \
		-d /var/www/htdocs/git.example.org/foo \
		<path to git repository to generate pages for>

To automate this process, create a `post-receive` hook for your
repository on your git server, see `githooks(5)` for more information.
Keep in mind that the repository page itself only needs to be regenerated
if the default branch is pushed, since only the default branch is
considered by `depp`. As such, an exemplary `post-receive` hook may look
as follows:

	#!/bin/sh
	
	repo=$(git rev-parse --absolute-git-dir)
	name=${repo##*/}
	
	rebuild=0
	defref=$(git symbolic-ref HEAD)
	while read local_ref local_sha remote_ref remote_sha; do
		if [ "${remote_ref}" = "${defref}" ]; then
			rebuild=1
			break
		fi
	done
	
	# Only rebuild the HTML if the default ref was pushed.
	[ ${rebuild} -eq 1 ] || exit 0
	
	depp -u "git://git.example.org/${name}" \
		-d "/var/www/htdocs/git.example.org/${name}" .

If usage of `deep-index` is also desired, the index page can either be
rebuild as part of the `post-receive` hook or in a separate cronjob.

### README Rendering

Rendering README files written in a chosen markup language (e.g.
markdown) is supported. This is achieved by including an executable file
called `git-render-readme` in the bare Git repository. When executed,
this file receives the README content on standard input and must write
plain HTML to standard output.

As an example, consider the following `git-render-readme` script which
uses the `markdown(1)` program provided by the [discount][discount website]
Markdown implementation:

	#!/bin/sh
	exec markdown -f autolink

### License

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU General Public License as published by the
Free Software Foundation, either version 3 of the License, or (at your
option) any later version.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
Public License for more details.

You should have received a copy of the GNU General Public License along
with this program. If not, see <https://www.gnu.org/licenses/>.

[stagit website]: http://codemadness.nl/git/stagit/log.html
[depp website]: https://depp.brause.cc/depp/
[libgit2 website]: https://libgit2.org/
[8pit git]: https://git.8pit.net/
[go website]: https://golang.org/
[discount website]: http://www.pell.portland.or.us/~orc/Code/discount/
[go embed]: https://pkg.go.dev/embed
[go-git github]: https://github.com/go-git/go-git
