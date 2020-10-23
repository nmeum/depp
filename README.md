# depp

No frills static side generator for Git repositories.

## Motivation

Dynamic git repository viewers like [cgit][cgit website] or
[gitweb][gitweb website] inherit the disadvantage of dynamic web
applications. For this reason, static page generators for git (e.g.
[git-arr][git-arr website] or [stagit][stagit website]) emerged
recently. However, these page generators are usual not compatible with
large repository as they generate lots of HTML files (e.g. one for each
commit).

Contrary to existing approaches, this software does not strive to be a
fully featured git browser for the web. Instead, the purpose of this
software is to provide a quick overview for a given repository. As such,
this software does intentionally not provide a web frontend for existing
tools like `git-log(1)`, `git-blame(1)`, et cetera. If more information
is needed the person viewing the generated web page should simply clone
the repository and use `git(1)` as usual.

## Status

Proof of concept, buggy and incomplete.

## Usage

Currently, this software is not intended to be installed system-wide.
Instead, use it directly from the repository. For normal operation,
[libgit2][libgit2 website] and [go][go website] is required. Afterwards
installing libgit2, clone this repository using:

	$ git clone --recursive https://github.com/nmeum/depp

Change into the newly cloned repository and build the software using:

	$ go build

Afterwards, HTML for a given git repository can be generated using the
`./depp` binary. For example, assuming you have a web server serving
files located at `/var/www/htdocs/git.example.org`, you want 10 commits
on the index page, and `git-daemon(1)` is running on the same domain:

	$ ./depp -c 10  -g git://git.example.org \
		-d /var/www/htdocs/git.example.org \
		<path to git repository to generate pages for>

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
[go webseite]: https://golang.org/
