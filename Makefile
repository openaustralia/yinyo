.PHONY: image server

all: run

run:
	./client.sh test/scrapers/test-python data.sqlite

# If you want an interactive shell in the container
shell:
	docker run --rm -i -t openaustralia/clay-scraper:v1 /bin/bash

lint:
	shellcheck build/package/clay-scraper/*.sh client.sh

minio_access_key = $(shell grep access_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_clay_access_key = $(shell grep store_access_key configs/secrets-clay-server.env | cut -d "=" -f 2)
minio_clay_secret_key = $(shell grep store_secret_key configs/secrets-clay-server.env | cut -d "=" -f 2)

buckets:
	mc config host add minio http://localhost:9000 $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_clay_access_key) $(minio_clay_secret_key)
	mc admin policy add minio clay configs/minio-clay-policy.json
	mc admin policy set minio clay user=$(minio_clay_access_key)
	mc mb -p minio/clay
