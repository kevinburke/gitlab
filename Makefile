DEP := $(GOPATH)/bin/dep
DIFFER := $(GOPATH)/bin/differ
BUMP_VERSION := $(GOPATH)/bin/bump_version
MEGACHECK := $(GOPATH)/bin/megacheck
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

test: vet
	@# this target should always be listed first so "make" runs the tests.
	go list ./... | grep -v vendor | xargs go test -short

race-test: vet
	go list ./... | grep -v vendor | xargs go test -race

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

$(DIFFER):
	go get github.com/kevinburke/differ

$(MEGACHECK):
	go get honnef.co/go/tools/cmd/megacheck

vet: $(MEGACHECK)
	@# We can't vet the vendor directory, it fails.
	go list ./... | grep -v vendor | xargs go vet
	go list ./... | grep -v vendor | xargs $(MEGACHECK) --ignore='github.com/kevinburke/gitlab/*/*.go:S1002'

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

deps: | $(DEP)
	$(DEP) ensure
	$(DEP) prune

release: race-test | $(BUMP_VERSION) $(DIFFER) $(RELEASE)
ifndef version
	@echo "Please provide a version"
	exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Please set GITHUB_TOKEN in the environment"
	exit 1
endif
	$(DIFFER) $(MAKE) authors
	$(BUMP_VERSION) minor main.go
	git push origin --tags
	mkdir -p releases/$(version)
	GOOS=linux GOARCH=amd64 go build -o releases/$(version)/gitlab-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o releases/$(version)/gitlab-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o releases/$(version)/gitlab-windows-amd64 .
	# These commands are not idempotent, so ignore failures if an upload repeats
	$(RELEASE) release --user kevinburke --repo gitlab --tag $(version) || true
	$(RELEASE) upload --user kevinburke --repo gitlab --tag $(version) --name gitlab-linux-amd64 --file releases/$(version)/gitlab-linux-amd64 || true
	$(RELEASE) upload --user kevinburke --repo gitlab --tag $(version) --name gitlab-darwin-amd64 --file releases/$(version)/gitlab-darwin-amd64 || true
	$(RELEASE) upload --user kevinburke --repo gitlab --tag $(version) --name gitlab-windows-amd64 --file releases/$(version)/gitlab-windows-amd64 || true

$(WRITE_MAILMAP):
	go get github.com/kevinburke/write_mailmap

AUTHORS.txt: | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt
	write_mailmap > AUTHORS.txt
