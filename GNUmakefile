NAME = depp

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
DOCDIR ?= $(PREFIX)/share/doc/$(NAME)

$(NAME):
	go build -trimpath -o $@

install: $(NAME) README.md
	install -Dm755 $(NAME) "$(DESTDIR)$(BINDIR)/$(NAME)"
	install -Dm644 README.md "$(DESTDIR)$(DOCDIR)/README.md"

.PHONY: $(NAME) install
