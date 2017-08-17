.PHONY: test test_race test_integrations get_deps build all install

all: get_deps install test

TMROOT = $${TMROOT:-$$HOME/.tendermint}
define NEWLINE

endef
NOVENDOR = go list github.com/tendermint/mintnet/... | grep -v /vendor/

install: 
	go install github.com/tendermint/mintnet

test: 
	go test `${NOVENDOR}`
	
test_race: 
	go test -race `${NOVENDOR}`

test_integrations: get_vendor_deps install test_race
	bash ./test/test.sh

get_deps:
	go get -d `${NOVENDOR}`
	go list -f '{{join .TestImports "\n"}}' github.com/tendermint/mintnet/... | \
		grep -v /vendor/ | sort | uniq | \
		xargs go get

update_deps:
	go get -d -u github.com/tendermint/mintnet

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
