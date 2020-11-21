# depp

No frills static side generator for Git repositories.

## Motivation

Dynamic git repository viewers like [cgit][cgit website] or
[gitweb][gitweb website] inherit the generael disadvantages of dynamic
web applications. For this reason, static page generators for git (e.g.
[git-arr][git-arr website] or [stagit][stagit website]) emerged
recently. However, these page generators are usual not compatible with
large repository as they generate lots of HTML files (e.g. one for each
commit).

Contrary to existing static page generator approaches, this software
does not strive to be a fully featured git browser for the web. Instead,
the idea is to provide a quick overview for a given repository, thereby
allowing a users to decide whether or not it is interesting enough to be
cloned. As such, this software does intentionally not provide a web
frontend for existing tools like `git-log(1)`, `git-blame(1)`, et
cetera. If more information is needed the user should simply clone the
repository and use `git(1)` as usual.

## Status

Proof of concept, buggy and incomplete. Currently requires an unreleased
version of the go compiler in order to embed template files into the binary.

## Dependencies

This software has the following dependencies:

* [libgit2][libgit2 website]
* [go][go website] >= 1.16.0 (`embed` package needed)
* C compiler, pkg-config, … for linking against libgit2

## Installation

This program can be installed using `go get`, if go itself is configured
properly. If not, simply build using `go build -trimpath`. The two
methods are described further below.

### go get

To install to the program using `go get` run the following command:

	$ go get github.com/nmeum/depp

### go build

To install the program without configuring a `$GOPATH`:

	$ git clone --recursive git://github.com/nmeum/depp
	$ go build -trimpath

Afterwards, copy `./depp` to a directory in your `$PATH`.

## Usage Example

Assuming you have a web server serving files located at
`/var/www/htdocs/git.example.org`, you want 10 commits on the index
page, and `git-daemon(1)` is running on the same domain:

	$ ./depp -c 10 -g git://git.example.org \
		-d /var/www/htdocs/git.example.org \
		<path to git repository to generate pages for>

To automate this process create a `post-receive` hook in your git
repository, see `githooks(5)` for more information on this topic.

## README Rendering

Rendering README files written in a chosen markup language (e.g.
markdown) is supported. This is achieved by including an executable file
called `git-render-readme` in the bare Git repository. When executed,
this file receives the README content on standard input and must write
plain HTML to standard output.

For example, consider the following `git-render-readme` script which
uses the `markdown(1)` program provided by the [discount][discount website]
markdown implementation:

	#!/bin/sh
	exec markdown

## Caveats

Existing HTML files are not tracked, thus generated HTML for files
removed from the repository `HEAD` are not automatically removed.

## License

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

[cgit website]: https://git.zx2c4.com/cgit/
[gitweb website]: https://git-scm.com/docs/gitweb
[git-arr website]: https://blitiri.com.ar/p/git-arr/
[stagit website]: http://codemadness.nl/git/stagit/log.html
[libgit2 website]: https://libgit2.org/
[go website]: https://golang.org/
[discount website]: http://www.pell.portland.or.us/~orc/Code/discount/
[git2go repo]: https://github.com/libgit2/git2go
[git2go build]: https://github.com/libgit2/git2go#installing
