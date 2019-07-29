all: run

run: image
	# TODO: Mount cache directory as volume as well
	docker run --rm -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/run.sh

image:
	docker build -t openaustralia/herokuish .
