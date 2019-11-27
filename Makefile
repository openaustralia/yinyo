.PHONY: image server test build ppa run

all: run

run: install
	clay client test/scrapers/test-python --output data.sqlite

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

minio_access_key = $(shell grep access_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_clay_access_key = $(shell grep store_access_key configs/secrets-clay-server.env | cut -d "=" -f 2)
minio_clay_secret_key = $(shell grep store_secret_key configs/secrets-clay-server.env | cut -d "=" -f 2)

buckets:
	echo "Waiting for Minio to start up..."
	kubectl wait --for condition=ready pod -l app=minio --timeout=60s --namespace clay-system
	echo "Minio is running..."
	mc config host add minio http://localhost:9000 $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_clay_access_key) $(minio_clay_secret_key)
	mc admin policy add minio clay configs/minio-clay-policy.json
	mc admin policy set minio clay user=$(minio_clay_access_key)
	mc mb -p minio/clay
