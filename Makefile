SHA=$(shell git rev-parse --short HEAD)

build:
	@docker build -t goodeggs/ecs-logs:latest --squash .
	@docker history goodeggs/ecs-logs:latest

release: build
	@docker tag goodeggs/ecs-logs:latest goodeggs/ecs-logs:$(SHA)
	@docker push goodeggs/ecs-logs:$(SHA)
	@docker push goodeggs/ecs-logs:latest
