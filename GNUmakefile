PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man
DOCDIR ?= $(PREFIX)/share/doc/depp

all: depp depp-index
depp:
	go build -trimpath -o $@ ./cmd/depp
depp-index:
	go build -trimpath -o $@ ./cmd/depp-index

install: depp depp-index depp.1 README.md
	install -Dm755 depp "$(DESTDIR)$(BINDIR)/depp"
	install -Dm755 depp "$(DESTDIR)$(BINDIR)/depp-index"
	install -Dm644 depp.1 "$(DESTDIR)$(MANDIR)/man1/depp.1"
	install -Dm644 README.md "$(DESTDIR)$(DOCDIR)/README.md"

.PHONY: all depp depp-index install
