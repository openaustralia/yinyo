all:
	GOOS=linux go build -o ./app .
	docker build -t clay .
	kubectl replace -f ../kubernetes/clay.yaml --force
	# Now can access the clay server at http://localhost:8080
