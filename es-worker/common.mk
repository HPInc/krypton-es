GHCR=ghcr.io/hpinc/krypton
REPO=images
all:
tag:
	docker tag $(DOCKER_IMAGE_NAME) $(GHCR)/$(REPO)/$(DOCKER_IMAGE_NAME)
publish: tag
	docker push $(GHCR)/$(REPO)/$(DOCKER_IMAGE_NAME)
