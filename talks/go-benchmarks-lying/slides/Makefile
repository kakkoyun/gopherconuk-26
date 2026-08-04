ROOT := ../../..
TALK := go-benchmarks-lying

.PHONY: all pdf html watch serve lint/md fix/md format/md check/typos fix/typos install clean help

all: pdf

pdf:
	$(MAKE) -C $(ROOT) build/pdf/$(TALK)

html:
	$(MAKE) -C $(ROOT) build/html/$(TALK)

watch:
	$(MAKE) -C $(ROOT) watch/$(TALK)

serve:
	$(MAKE) -C $(ROOT) serve/$(TALK)

lint/md:
	$(MAKE) -C $(ROOT) lint/md/$(TALK)

fix/md:
	$(MAKE) -C $(ROOT) fix/md/$(TALK)

format/md:
	$(MAKE) -C $(ROOT) format/md/$(TALK)

check/typos:
	$(MAKE) -C $(ROOT) check/typos/$(TALK)

fix/typos:
	$(MAKE) -C $(ROOT) fix/typos/$(TALK)

install:
	$(MAKE) -C $(ROOT) install

clean:
	$(MAKE) -C $(ROOT) clean/$(TALK)

help:
	$(MAKE) -C $(ROOT) help
