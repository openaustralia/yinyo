from __future__ import print_function

import sys

# This tests that stdout and stderr are sent in the correct order

print("Line 1 (to stderr)", file=sys.stderr)
print("Line 2 (to stderr)", file=sys.stderr)

print("Line 3 (to stdout)")
print("Line 4 (to stdout)")
