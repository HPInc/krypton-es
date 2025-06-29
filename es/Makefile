DOCKER_IMAGE_NAME=krypton-es
BIN=bin/es
ifndef SRC
  SRC=./service
endif

build:
	go build -o $(BIN) service/main.go

test: ci_test

docker: protos
	docker build -t $(DOCKER_IMAGE_NAME) .

# ci_test has a minimal dependency path in a self contained test env
ci_test: protos
	make -C $(COMPOSE_DIR) test


clean_tests: stop_deps clean_test_cache

clean_test_cache:
	go clean -testcache

test_docker: docker-image
	docker run -it -p7979:7979 --rm $(DOCKER_IMAGE_NAME)


.PHONY: run build docker-image test_docker clean check_goimports stop stop_deps test ci_test docker

include common.mk
