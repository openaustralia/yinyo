all: run

run: image
	docker run --rm -v `pwd`/test-scrapers/ruby:/tmp/app openaustralia/herokuish /bin/herokuish test

image:
	docker build -t openaustralia/herokuish .
