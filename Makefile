VERSION = 1.3.0
.PHONY: all clean daemon cli

PKGPATH=pkg
CURDIR=$(shell pwd)
PYDIR=$(shell which python3)
PREFIX     ?= /usr
CONFDIR    ?= /etc
LIBEXEC    ?= libexec
BINDIR      = $(DESTDIR)$(PREFIX)/bin
SYSCONFDIR  = $(DESTDIR)$(CONFDIR)
SYSTEMDDIR  = $(DESTDIR)$(PREFIX)/lib/systemd/system
SRCVERSION  = $(shell git rev-parse --short HEAD 2>/dev/null)
ATUNEVERSION = $(VERSION)$(if $(SRCVERSION),($(SRCVERSION)))
SHELL = /bin/bash

GOFLAGS = -ldflags '-s -w -extldflags "-static" -extldflags "-zrelro" -extldflags "-znow" -extldflags "-ftrapv" -linkmode=external'

all: abs-python daemon cli 

abs-python:
	@if [ $(PYDIR) ] ; then \
		sed -i "s?ExecStart=python3?ExecStart=$(PYDIR)?g" $(CURDIR)/keentuned.service; \
	else \
		echo "no python3 exists."; \
	fi

daemon:
	cd daemon && go build -mod=vendor -v $(GOFLAGS) -o ../$(PKGPATH)/keentuned

cli:
	cd cli && go build -mod=vendor -v $(GOFLAGS) -o ../$(PKGPATH)/keentune

clean:
	rm -rf $(PKGPATH)/*

install: 
	@echo "BEGIN INSTALL keentuned"
	mkdir -p $(BINDIR)
	mkdir -p $(SYSCONFDIR)/keentune/
	mkdir -p $(SYSCONFDIR)/keentune/conf/
	mkdir -p $(SYSTEMDDIR)
	mkdir -p $(DESTDIR)$(PREFIX)/share/bash-completion/completions/
	install -m 0755 $(PKGPATH)/keentune $(BINDIR)
	install -m 0755 $(PKGPATH)/keentuned $(BINDIR)
	cp -rf daemon/examples/* $(SYSCONFDIR)/keentune
	install -m 0644 keentuned.conf $(SYSCONFDIR)/keentune/conf/
	install -m 0644 keentuned.service $(SYSTEMDDIR)
	@echo "END INSTALL keentuned"

startup:
	systemctl daemon-reload
	systemctl restart keentuned
	systemctl restart keentuned

run: all install startup

check: run
	cd ${CURDIR}/test && python3 main.py

authors:
		git shortlog --summary --numbered --email|grep -v openeuler-ci-bot|sed 's/<root@localhost.*//'| awk '{$$1=null;print $$0}'|sed 's/^[ ]*//g' > AUTHORS
