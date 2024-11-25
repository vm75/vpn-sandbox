default: build

build:
	docker buildx build --platform linux/amd64 --format docker -t vm75/vpn-sandbox .

run:
	docker run --rm -p 8080:80 -p 1080:1080 -p 3128:3128 vm75/vpn-sandbox

test-ci:
	act -s DOCKER_USERNAME -s DOCKER_PASSWORD -s GITHUB_TOKEN
