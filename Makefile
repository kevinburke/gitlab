DEP := $(GOPATH)/bin/dep
MEGACHECK := $(GOPATH)/bin/megacheck

test: vet
	@# this target should always be listed first so "make" runs the tests.
	go list ./... | grep -v vendor | xargs go test -short

race-test: vet
	go list ./... | grep -v vendor | xargs go test -race

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
