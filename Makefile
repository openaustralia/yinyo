all:
	docker run --rm -v `pwd`/test-scrapers/ruby:/tmp/app gliderlabs/herokuish /bin/herokuish test
