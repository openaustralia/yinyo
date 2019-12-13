.PHONY: image server test build ppa run website apidocs

all: build

build:
	go build ./...

run: install
	yinyo client test/scrapers/test-python --output data.sqlite

test: install
	go test ./... -short

integration: install
	go test ./...

install:
	go install ./...

ppa:
	sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt-get update
	sudo apt-get install golang-go

mocks:
	mockery -all -inpkg

website: apidocs
	# Starts a development web server at http://localhost:1313
	hugo server -s site -D

apidocs:
	widdershins --summary openapi/definition.yaml -o openapi/definition.md
	shins openapi/definition.md -o site/content/api.html --inline

minio_access_key = $(shell grep access_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_yinyo_access_key = $(shell grep store_access_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)
minio_yinyo_secret_key = $(shell grep store_secret_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)

buckets:
	echo "Waiting for Minio to start up..."
	kubectl wait --for condition=ready pod -l app=minio --timeout=60s --namespace yinyo-system
	echo "Minio is running..."
	mc config host add minio http://localhost:9000 $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_yinyo_access_key) $(minio_yinyo_secret_key)
	mc admin policy add minio yinyo configs/minio-yinyo-policy.json
	mc admin policy set minio yinyo user=$(minio_yinyo_access_key)
	mc mb -p minio/yinyo
