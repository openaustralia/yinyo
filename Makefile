all: run

run: image
	docker run --rm -v `pwd`/cache:/tmp/cache -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/run.sh

# Clean the cache
clean:
	rm -rf cache

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/bash

image:
	docker build -t openaustralia/herokuish .
