.PHONY: image server

morph_scraper_name = morph-test-scrapers/test-python

all: run

run:
	./morph.sh $(morph_scraper_name)

# If you want an interactive shell in the container
shell:
	docker run --rm -i -t openaustralia/clay-scraper:v1 /bin/bash

lint:
	shellcheck images/clay-scraper/run.sh images/clay-scraper/clay.sh morph.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

minio_access_key = $(shell grep access_key secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key secrets-minio.env | cut -d "=" -f 2)
minio_clay_access_key = clay
minio_clay_secret_key = changeme123
minio_morph_access_key = morph
minio_morph_secret_key = changeme123

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
