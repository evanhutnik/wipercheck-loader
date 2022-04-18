DOCKER_REPO=evanhutnik/wipercheck-loader

build:
	docker build --tag ${DOCKER_REPO} .

push:
	docker push ${DOCKER_REPO}:latest
