# https://cheatography.com/linux-china/cheat-sheets/justfile/

set dotenv-load := true

default:
	@just --list

# build the binary
build:
	go build -o dots .

# run tests
test:
	go test ./...

# smoke test: build and run against a temp dir (safe, no real files touched)
smoke: build
	bash scripts/smoke.sh ./dots

# watch and run a go file
watch PATH:
	ls {{PATH}}/* | entr -c go run {{PATH}}/*.go

# watch and run a go file
wtest PATH:
	ls {{PATH}}/* | entr -c go test {{PATH}}/*.go

# count without tests
count:
	tokei . --exclude '*_test.go'

# local release build into dist/ (nothing is published)
snapshot:
	mise exec goreleaser@latest -- goreleaser release --snapshot --clean

# tag pushed main as vVERSION and push the tag; the release workflow publishes it
release VERSION:
	#!/usr/bin/env bash
	set -euo pipefail
	v="{{VERSION}}"; tag="v${v#v}"   # accepts 0.1.0 or v0.1.0
	git fetch -q origin
	main=$(git rev-parse main)
	[ "$main" = "$(git rev-parse origin/main)" ] || { echo "main and origin/main differ - push main first"; exit 1; }
	git tag -a "$tag" -m "$tag" "$main"
	git push origin "$tag"
	echo "tagged $tag - watch: gh run list --workflow release.yml"
