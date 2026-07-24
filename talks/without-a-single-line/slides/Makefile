.PHONY: all pdf html watch serve lint/md fix/md format/md check/typos fix/typos install clean help

MARP_CONFIG := marp/marp.config.js
SRC         := presentation.md
FLAGS       := --allow-local-files --theme-set marp/css/

# OS detection for optional tool installation
UNAME_S := $(shell uname -s)
TYPOS   := $(shell command -v typos 2>/dev/null)

all: pdf

pdf:
	npx marp --config $(MARP_CONFIG) $(SRC) $(FLAGS) -o presentation.pdf

html:
	npx marp --config $(MARP_CONFIG) $(SRC) $(FLAGS) -o presentation.html

watch:
	npx marp --config $(MARP_CONFIG) $(SRC) $(FLAGS) --watch --preview

serve: html
	python -m http.server 8000

lint/md:
	npx markdownlint-cli2 "**/*.md" "#node_modules"

fix/md:
	npx markdownlint-cli2 --fix "**/*.md" "#node_modules"

format/md:
	npx prettier --write "*.md"

install:
	npm install

install/typos:
ifndef TYPOS
ifeq ($(UNAME_S),Darwin)
	brew install typos-cli
else ifeq ($(UNAME_S),Linux)
	cargo install typos-cli
else
	@echo "Install typos-cli: https://github.com/crate-ci/typos"
	@exit 1
endif
endif

check/typos: install/typos
	typos

fix/typos: install/typos
	typos --write-changes

clean:
	rm -f presentation.pdf presentation.html

help:
	@echo "Presentation:"
	@echo "  make pdf         Build PDF"
	@echo "  make html        Build HTML"
	@echo "  make watch       Live preview"
	@echo "  make serve       HTML + HTTP server at :8000"
	@echo ""
	@echo "Lint & format:"
	@echo "  make lint/md     Lint markdown"
	@echo "  make fix/md      Auto-fix markdown"
	@echo "  make format/md   Format markdown (prettier)"
	@echo "  make check/typos Check for typos"
	@echo "  make fix/typos   Fix typos automatically"
	@echo ""
	@echo "Misc:"
	@echo "  make install     Install npm dependencies"
	@echo "  make clean       Remove output files"
