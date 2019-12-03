.PHONY: image server test build ppa run website apidocs

all: run

run: install
	yinyo client test/scrapers/test-python --output data.sqlite

test: install
	go test ./...

install:
	go install ./...

ppa:
	sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt-get update
	sudo apt-get install golang-go

mocks:
	mockery -all -inpkg

shins: widdershins
	docker run --rm -v `pwd`:/app tchaypo/shins-docker /app/openapi/definition.md -o /app/site/content/api.html --inline

widdershins:
	docker run --rm -v `pwd`:/app quay.io/verygoodsecurity/widdershins-docker --summary /app/openapi/definition.yaml -o /app/openapi/definition.md

apidocs: widdershins shins

website: apidocs
	docker run --rm --name "yinyo-docs" -P -v $$(pwd):/src -p 1313:1313 klakegg/hugo server -s /src/site -D


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
