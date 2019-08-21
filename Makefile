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

buckets:
	mc config host add minio $(shell minikube service --url minio-service -n clay-system | cut -f 2 -d \ ) admin changeme
	mc mb minio/clay
	mc mb minio/morph
