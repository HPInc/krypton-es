DOCKER_IMAGE_NAME=krypton-es-worker
CACHE_IMAGE_NAME=es-worker-cached-builder
# See README.md on why we have protos coming in as docker containers
CA_IMAGE=ghcr.io/hpinc/krypton/krypton-ca:latest
CA_PROTO=ghcr.io/hpinc/krypton/krypton-caprotos:latest
CA_PROTO_LOCAL=ca_proto_local
CA_PROTOS_DIR=protos/ca
DSTS_IMAGE=ghcr.io/hpinc/krypton/krypton-dsts:latest
DSTS_PROTO=ghcr.io/hpinc/krypton/krypton-dstsprotos:latest
DSTS_PROTO_LOCAL=dsts_proto_local
DSTS_PROTOS_DIR=protos/dsts
BIN=bin/es-worker
SRC=./service
include common.mk

run: deps
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn \
	AWS_ACCESS_KEY_ID=test \
	AWS_SECRET_ACCESS_KEY=test \
	AWS_REGION=us-east-1 \
	ESW_NOTIFICATION_ENDPOINT=http://localhost:9324 \
	ESW_CONFIG_FILE=./service/config/config.yaml \
	go run $(SRC)/...

deps: protos
	make -C tools/compose

protos: $(CA_PROTOS_DIR) $(DSTS_PROTOS_DIR)
$(CA_PROTOS_DIR):
	docker pull $(CA_PROTO)
	mkdir -p $(CA_PROTOS_DIR) && \
	docker create --name $(CA_PROTO_LOCAL) $(CA_PROTO) "" && \
	docker cp $(CA_PROTO_LOCAL):/protos $(CA_PROTOS_DIR) && \
	docker rm $(CA_PROTO_LOCAL) && \
	mv $(CA_PROTOS_DIR)/protos/* $(CA_PROTOS_DIR) && rm -rf $(CA_PROTOS_DIR)/protos

$(DSTS_PROTOS_DIR):
	docker pull $(DSTS_PROTO)
	mkdir -p $(DSTS_PROTOS_DIR) && \
	docker create --name $(DSTS_PROTO_LOCAL) $(DSTS_PROTO) "" && \
	docker cp $(DSTS_PROTO_LOCAL):/protos $(DSTS_PROTOS_DIR) && \
	docker rm $(DSTS_PROTO_LOCAL) && \
	mv $(DSTS_PROTOS_DIR)/protos/* $(DSTS_PROTOS_DIR) && rm -rf $(DSTS_PROTOS_DIR)/protos

fmt:
	go fmt $(SRC)/...

imports: check_goimports
	goimports -w .

vet:
	go vet $(SRC)/...

tidy:
	go mod tidy

gosec:
	gosec $(SRC)/...

lint: protos
	./tools/run_linter.sh

build:
	go build -o $(BIN) service/main.go

# build cache if no local cache exist
docker_cache:
	@if [ $$(docker images -q $(CACHE_IMAGE_NAME) | wc -l) -eq 0 ]; then docker build -t $(CACHE_IMAGE_NAME) -f docker/Dockerfile.cache .; fi

docker: $(CA_PROTOS_DIR) $(DSTS_PROTOS_DIR) docker_cache
	docker build -t $(DOCKER_IMAGE_NAME) .

test:
	go test $(SRC)/...

test_docker: docker-image
	docker run -it --rm $(DOCKER_IMAGE_NAME)

clean: clean_deps clean_build_cache
	go clean

clean_deps: clean_$(CA_PROTOS_DIR) clean_$(DSTS_PROTOS_DIR)
	make -C tools/compose clean

clean_$(CA_PROTOS_DIR):
	-rm -rf $(CA_PROTOS_DIR)
	-docker rmi $(CA_PROTO)
	-docker rmi $(CA_IMAGE)

clean_$(DSTS_PROTOS_DIR):
	-rm -rf $(DSTS_PROTOS_DIR)
	-docker rmi $(DSTS_PROTO)
	-docker rmi $(DSTS_IMAGE)

clean_build_cache:
	-docker rmi $(CACHE_IMAGE_NAME)

check_goimports:
	@which goimports >/dev/null 2>&1 || (echo "goimports not found";exit 1)

.PHONY: run fmt tidy build docker-image test_docker clean check_goimports lint protos clean_build_cache docker docker_cache gosec
