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
