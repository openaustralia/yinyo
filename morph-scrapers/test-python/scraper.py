import scraperwiki
import time
import requests
import sys

expected_string = "Hassle-free web scraping."

# Check that proxy gives a working certificate for SSL connections
# If the certificate isn't valid it should throw an exception
html = scraperwiki.scrape("https://morph.io")

if not expected_string in html:
    raise Exception("Not expected result")

# Use requests library to do the same because it gets its CA certs a different way. Oh joy.
r = requests.get('https://morph.io')
if not expected_string in r.text:
    raise Exception("Not expected result")

# Write out to the sqlite database using scraperwiki library
scraperwiki.sqlite.save(unique_keys=['name'], data={"name": "susan", "occupation": "software developer"})

sys.stderr.write("First a little test message to stderr\n")

print "Hello from test-python!"

for i in range(1, 6):
    print "%i..." % i
    time.sleep(1)
