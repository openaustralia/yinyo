all: run

run: image
	docker run --rm -v `pwd`/cache:/tmp/cache morph-ng /bin/run.sh morph-test-scrapers/test-ruby

# Clean the cache
clean:
	rm -rf cache

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t -v `pwd`/cache:/tmp/cache morph-ng /bin/bash

image:
	docker build -t morph-ng .

lint:
	shellcheck run.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

# This runs the scraper on kubernetes
kubernetes: image
	kubectl apply -f job.yaml
