.PHONY: image server

all: run

run:
	./morph.sh -d morph-scrapers/test-python

# If you want an interactive shell in the container
shell:
	docker run --rm -i -t openaustralia/clay-scraper:v1 /bin/bash

lint:
	shellcheck images/clay-scraper/run.sh images/clay-scraper/clay.sh morph.sh

minio_access_key = $(shell grep access_key secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key secrets-minio.env | cut -d "=" -f 2)
minio_clay_access_key = $(shell grep store_access_key secrets-clay-server.env | cut -d "=" -f 2)
minio_clay_secret_key = $(shell grep store_secret_key secrets-clay-server.env | cut -d "=" -f 2)
minio_morph_access_key = $(shell grep store_access_key secrets-morph.env | cut -d "=" -f 2)
minio_morph_secret_key = $(shell grep store_secret_key secrets-morph.env | cut -d "=" -f 2)

buckets:
	mc config host add minio $(shell minikube service --url minio-service -n clay-system) $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_clay_access_key) $(minio_clay_secret_key)
	mc admin user add minio $(minio_morph_access_key) $(minio_morph_secret_key)
	mc admin policy add minio clay minio-clay-policy.json
	mc admin policy add minio morph minio-morph-policy.json
	mc admin policy set minio clay user=$(minio_clay_access_key)
	mc admin policy set minio morph user=$(minio_morph_access_key)
	mc mb -p minio/clay
	mc mb -p minio/morph
