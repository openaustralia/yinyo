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

minio_access_key = $(shell grep access_key minio-secrets.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key minio-secrets.env | cut -d "=" -f 2)

buckets:
	mc config host add minio $(shell minikube service --url minio-service -n clay-system) $(minio_access_key) $(minio_secret_key)
	mc mb -p minio/clay
	mc mb -p minio/morph
