all:
	GOOS=linux go build -o ./app .
	docker build -t clay .
	kubectl run --rm -i demo --image=clay --image-pull-policy=Never
