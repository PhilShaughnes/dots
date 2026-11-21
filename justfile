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

# release the next patch / minor / major version after the latest v* tag
release-patch: (_bump "patch")
release-minor: (_bump "minor")
release-major: (_bump "major")

_bump PART:
	#!/usr/bin/env bash
	set -euo pipefail
	git fetch -q --tags origin
	last=$(git tag -l "v*" --sort=-v:refname | head -1)
	IFS=. read -r ma mi pa <<< "${last#v}"
	ma=${ma:-0}; mi=${mi:-0}; pa=${pa:-0}
	case "{{PART}}" in
		major) ma=$((ma+1)); mi=0; pa=0 ;;
		minor) mi=$((mi+1)); pa=0 ;;
		patch) pa=$((pa+1)) ;;
	esac
	next="$ma.$mi.$pa"
	echo "${last:-none} -> v$next"
	{{just_executable()}} release "$next"
