from __future__ import print_function

import scraperwiki
import time
import random
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
scraperwiki.sqlite.save(unique_keys=['name'], data={
                        "name": "susan", "occupation": "software developer"})

print("First a little test message to stderr", file=sys.stderr)
print("A second line of error", file=sys.stderr)

print("Hello from test-python!")
print("Stdout gets some extra text")

for i in range(1, 6):
    print("%i..." % i)
    fail = random.choice([0,1])
    if fail == 1:
        print("Simulated oopsie!", file=sys.stderr)
        print("Oppsie: %i" % i, file=sys.stderr)
    else:
        print("Success %i" % i)
        print("This line is just filler %i..." % i)
    time.sleep(1)
