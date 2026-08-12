# shellcheck disable=SC2046,SC2086
SHELL := /bin/bash

TALKS := go-benchmarks-lying without-a-single-line
THEME_DIR := themes/gophercon-datadog
THEME_SET := $(THEME_DIR)
THEME_CSS := $(THEME_DIR)/gophercon-datadog.css
THEME_MINIMAL_CSS := $(THEME_DIR)/gophercon-datadog-minimal.css
SLIDE_MARKDOWN := $(foreach talk,$(TALKS),talks/$(talk)/slides/presentation.md)
THEME_MARKDOWN := $(THEME_DIR)/README.md $(THEME_DIR)/deck.md
SLIDE_PDFS := talks/go-benchmarks-lying/slides/presentation.pdf talks/without-a-single-line/slides/presentation.pdf
GO_FILES := $(shell find talks tools -name '*.go' -type f -print | sort)
MARP := npx marp

.DEFAULT_GOAL := help

.PHONY: all build/html build/pdf build/html/% build/pdf/% watch/% serve/% clean clean/% \
	check check/fast check/slides check/css check/theme check/footer check/go \
	lint/md lint/md/% fix/md/% format/md format/md/% check/typos check/typos/% fix/typos/% \
	pre-commit sync/theme hooks install help

all: check

build/html: build/html/go-benchmarks-lying build/html/without-a-single-line

build/html/%: sync/theme
	$(MARP) --config talks/$*/slides/marp/marp.config.js talks/$*/slides/presentation.md --allow-local-files --theme-set "$(THEME_SET)" -o talks/$*/slides/presentation.html

build/pdf: build/pdf/go-benchmarks-lying build/pdf/without-a-single-line

build/pdf/%: sync/theme
	$(MARP) --config talks/$*/slides/marp/marp.config.js talks/$*/slides/presentation.md --allow-local-files --theme-set "$(THEME_SET)" --pdf -o talks/$*/slides/presentation.pdf

watch/%: sync/theme
	$(MARP) --config talks/$*/slides/marp/marp.config.js talks/$*/slides/presentation.md --allow-local-files --theme-set "$(THEME_SET)" --watch --preview

serve/%: build/html/%
	python3 -m http.server --directory talks/$*/slides 8000

clean: clean/go-benchmarks-lying clean/without-a-single-line

clean/%:
	rm -rf talks/$*/slides/presentation.html talks/$*/slides/presentation.pdf talks/$*/slides/assets talks/$*/slides/fonts

check: check/slides check/go

check/fast: check/theme lint/md check/typos

check/slides: check/css lint/md check/typos build/html check/footer

check/css: check/theme

check/footer: build/pdf
	python3 tools/check_slide_footer.py $(SLIDE_PDFS)

check/theme:
	@set -o errexit -o nounset -o pipefail; \
	output="$(THEME_DIR)/.theme-check.pdf"; \
	trap 'rm -f "$$output"' EXIT; \
	$(MARP) "$(THEME_DIR)/deck.md" --allow-local-files --theme-set "$(THEME_SET)" --pdf -o "$$output"; \
	test -s "$$output"; \
	$(MARP) "$(THEME_DIR)/deck.md" --allow-local-files --theme-set "$(THEME_SET)" --theme gophercon-datadog-minimal --pdf -o "$$output"; \
	test -s "$$output"

check/go:
	command -v goimports >/dev/null || { printf '%s\n' 'Install goimports: go install golang.org/x/tools/cmd/goimports@v0.45.0' >&2; exit 1; }
	if goimports -l $(GO_FILES) | grep -q .; then \
		printf 'Run goimports on:\n' >&2; \
		goimports -l $(GO_FILES) >&2; \
		exit 1; \
	fi
	if gofmt -l $(GO_FILES) | grep -q .; then \
		printf 'Run gofmt on:\n' >&2; \
		gofmt -l $(GO_FILES) >&2; \
		exit 1; \
	fi
	printf '%s\n' '==> talks/go-benchmarks-lying/demo'
	cd talks/go-benchmarks-lying/demo && go vet ./... && go test -race -count=1 ./...
	printf '%s\n' '==> tools/cli/benchenv'
	cd tools/cli/benchenv && go vet ./... && go test -race -count=1 ./...
	printf '%s\n' '==> tools/cli/benchgate'
	cd tools/cli/benchgate && go vet ./... && go test -race -count=1 ./...
	printf '%s\n' '==> tools/cli/honestbench'
	cd tools/cli/honestbench && go vet ./... && go test -race -count=1 ./...
	printf '%s\n' '==> tools/cli/kubectl-obi'
	cd tools/cli/kubectl-obi && go vet ./... && go test -race -count=1 ./...

lint/md:
	$(MARP) --version >/dev/null
	npx markdownlint-cli2 --config .markdownlint-cli2.yaml $(SLIDE_MARKDOWN) $(THEME_MARKDOWN)

lint/md/%:
	npx markdownlint-cli2 --config .markdownlint-cli2.yaml talks/$*/slides/presentation.md

fix/md/%:
	npx markdownlint-cli2 --config .markdownlint-cli2.yaml --fix talks/$*/slides/presentation.md

format/md: format/md/go-benchmarks-lying format/md/without-a-single-line

format/md/%:
	npx prettier --write talks/$*/slides/presentation.md

check/typos:
	typos $(SLIDE_MARKDOWN) $(THEME_MARKDOWN)

check/typos/%:
	typos talks/$*/slides/presentation.md

fix/typos/%:
	typos --write-changes talks/$*/slides/presentation.md

pre-commit: check/fast

sync/theme:
	@rsync -a --delete "$(THEME_DIR)/fonts/" talks/go-benchmarks-lying/slides/fonts/
	@rsync -a --delete "$(THEME_DIR)/assets/" talks/go-benchmarks-lying/slides/assets/
	@rsync -a --delete "$(THEME_DIR)/fonts/" talks/without-a-single-line/slides/fonts/
	@rsync -a --delete "$(THEME_DIR)/assets/" talks/without-a-single-line/slides/assets/

hooks:
	@git config core.hooksPath "$$(git rev-parse --show-toplevel)/.githooks"
	@printf 'Configured core.hooksPath for this worktree.\n'

install:
	npm install

help:
	@printf '%s\n' 'Build:'
	@printf '%s\n' '  make build/html          Build both HTML decks'
	@printf '%s\n' '  make build/pdf           Build both PDF decks'
	@printf '%s\n' '  make watch/<talk>        Open a live preview for one talk'
	@printf '%s\n' ''
	@printf '%s\n' 'Verify:'
	@printf '%s\n' '  make check/fast          Theme, Markdown, and typo checks'
	@printf '%s\n' '  make check               Full slides, footer, CSS/theme, and Go checks'
	@printf '%s\n' '  make sync/theme          Stage output-local theme resources'
	@printf '%s\n' ''
	@printf '%s\n' 'Hooks:'
	@printf '%s\n' '  make hooks               Enable the repository pre-commit hook'
