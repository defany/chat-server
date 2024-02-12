include .env
export

REGISTRY_HOST=docker.io
REGISTRY:=defany
CONTAINER_NAME:=chat-server:v0.0.1

protogen:
	buf generate proto

	go mod tidy

run:
	go run ./app/cmd/main.go

build-push:
	docker buildx build --no-cache --platform linux/amd64 -t $(REGISTRY_HOST)/$(REGISTRY)/$(CONTAINER_NAME) .

	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD) $(REGISTRY_HOST)/$(REGISTRY)

	docker push $(REGISTRY)/$(CONTAINER_NAME)

docker-run:
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)

	docker run -p 50001:50001 $(REGISTRY)/$(CONTAINER_NAME)