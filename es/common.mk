GHCR=ghcr.io/hpinc
REPO=krypton
DSTS_PROTOS_DIR=protos/dsts
DSTS_PROTO_IMAGE=ghcr.io/hpinc/krypton/krypton-dstsprotos:latest
DSTS_PROTO_LOCAL=dsts_proto_local
COMPOSE_DIR=tools/compose

tag:
	docker tag $(DOCKER_IMAGE_NAME) $(GHCR)/$(REPO)/$(DOCKER_IMAGE_NAME)

publish: tag
	docker push $(GHCR)/$(REPO)/$(DOCKER_IMAGE_NAME)

stop: stop_deps

clean: stop_deps clean_protos clean_tests
	go clean

clean_protos: clean_$(DSTS_PROTOS_DIR)

clean_$(DSTS_PROTOS_DIR):
	-rm -rf $(DSTS_PROTOS_DIR)
	-docker rmi $(DSTS_PROTO_IMAGE)

run: deps
	-ES_DB_SCHEMA_MIGRATION_SCRIPTS=$(CURDIR)/service/db/schema \
	AWS_ACCESS_KEY_ID=test \
	AWS_SECRET_ACCESS_KEY=test \
	AWS_REGION=us-east-1 \
	ES_NOTIFICATION_ENDPOINT=http://localhost:9324 \
	ES_DB_ENROLL_EXPIRY_MINUTES=1 \
	ES_CONFIG_FILE=./service/config/config.yaml \
	ES_TOKEN_CONFIG_FILE=./service/config/token_config_local.yaml \
	ES_DEFAULT_POLICY_FILE=./service/config/default_policy.json \
	go run $(SRC)/...

deps: protos
	make -C $(COMPOSE_DIR)

protos: $(DSTS_PROTOS_DIR)
$(DSTS_PROTOS_DIR):
	docker pull $(DSTS_PROTO_IMAGE)
	mkdir -p $(DSTS_PROTOS_DIR) && \
	docker create --name $(DSTS_PROTO_LOCAL) $(DSTS_PROTO_IMAGE) "" && \
	docker cp $(DSTS_PROTO_LOCAL):/protos $(DSTS_PROTOS_DIR) && \
	docker rm $(DSTS_PROTO_LOCAL) && \
	mv $(DSTS_PROTOS_DIR)/protos/* $(DSTS_PROTOS_DIR) && rm -rf $(DSTS_PROTOS_DIR)/protos

schema_migrate: deps_test
	ES_DB_SCHEMA_MIGRATION_SCRIPTS=$(CURDIR)/service/db/schema \
	ES_DB_SCHEMA_MIGRATION_ENABLED=true \
	ES_MODE_SCHEMA_MIGRATION=true go run $(SRC)/...

stop_deps:
	make -C $(COMPOSE_DIR) stop

check_goimports:
	@which goimports >/dev/null 2>&1 || (echo "goimports not found";exit 1)

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

show_data:
	make -C $(COMPOSE_DIR) show_data

clean_data:
	make -C $(COMPOSE_DIR) clean_data

.PHONY: fmt imports vet tidy gosec show_data clean_data
