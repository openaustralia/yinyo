.PHONY: image server test build ppa run website apidocs minikube buckets clean skaffold dashboard mocks

all: run

run: install
	yinyo test/scrapers/test-python --output data.sqlite

test:
	go test -short -cover ./...

integration:
	go test -cover ./...

install:
	go install cmd/yinyo/yinyo.go

ppa:
	sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt-get update
	sudo apt-get install golang-go

mocks:
	mockery -all -keeptree

website: apidocs
	cd site; npm install
	# Starts a development web server at http://localhost:1313
	hugo server -s site -D

publish-website:
	cd site; /bin/sh ./publish_to_ghpages.sh

apidocs:
	widdershins --summary openapi/definition.yaml -o openapi/definition.md
	shins openapi/definition.md --layout $(shell pwd)/site/layout.ejs -o site/content/api.html --inline --logo site/static/logo.svg --logo-url / --css site/api-overrides.css

minikube:
	minikube start --memory=3072 --disk-size='30gb'

dashboard:
	minikube dashboard

skaffold:
	skaffold dev --port-forward=true --status-check=false

clean:
	minikube delete
