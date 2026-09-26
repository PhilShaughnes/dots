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
