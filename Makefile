all: run

run: image
	# TODO: Mount cache directory as volume as well
	docker run --rm -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/run.sh

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/bash

image:
	docker build -t openaustralia/herokuish .
