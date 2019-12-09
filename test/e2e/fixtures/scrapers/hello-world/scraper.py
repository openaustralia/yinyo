import os

file = open("output.txt", "w")
file.write("Hello World!")
file.close()

print(os.environ['HELLO'])
