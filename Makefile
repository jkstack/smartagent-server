.PHONY: all

OUTDIR=release

VERSION=2.0.2
TIMESTAMP=`date +%s`

BRANCH=`git rev-parse --abbrev-ref HEAD`
HASH=`git log -n1 --pretty=format:%h`
REVERSION=`git log --oneline|wc -l|tr -d ' '`
BUILD_TIME=`date +'%Y-%m-%d %H:%M:%S'`
LDFLAGS="-X 'main.gitBranch=$(BRANCH)' \
-X 'main.gitHash=$(HASH)' \
-X 'main.gitReversion=$(REVERSION)' \
-X 'main.buildTime=$(BUILD_TIME)' \
-X 'main.version=$(VERSION)'"

all:
	go mod vendor
	rm -fr $(OUTDIR)/$(VERSION)
	mkdir -p $(OUTDIR)/$(VERSION)/opt/smartagent-server/bin \
		$(OUTDIR)/$(VERSION)/opt/smartagent-server/conf
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod vendor -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent-server/bin/smartagent-server code/*.go
	cp conf/server.conf $(OUTDIR)/$(VERSION)/opt/smartagent-server/conf/server.conf
	echo $(VERSION) > $(OUTDIR)/$(VERSION)/opt/smartagent-server/.version
	cd $(OUTDIR)/$(VERSION) && fakeroot tar -czvf smartagent-server_$(VERSION).tar.gz \
		--warning=no-file-changed opt
	rm -fr $(OUTDIR)/$(VERSION)/opt $(OUTDIR)/$(VERSION)/etc
	cp CHANGELOG.md $(OUTDIR)/CHANGELOG.md
version:
	@echo $(VERSION)
distclean:
	rm -fr $(OUTDIR)
