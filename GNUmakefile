NAME = depp

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man
DOCDIR ?= $(PREFIX)/share/doc/$(NAME)

# GOFLAGS is automatically picked up by go itself.
# See `go help environment` for more information.
GOFLAGS += -trimpath

$(NAME):
	go build

install:
	install -Dm755 $(NAME) "$(DESTDIR)$(BINDIR)/$(NAME)"
	install -Dm644 README.md "$(DESTDIR)$(DOCDIR)/README.md"

.PHONY: $(NAME)
