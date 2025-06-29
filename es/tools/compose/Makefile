# we assume all test machines have either docker-compose or
# the compose plugin in docker command.
DOCKER_COMPOSE=docker-compose
HAS_DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef HAS_DOCKER_COMPOSE
  DOCKER_COMPOSE=docker compose
  DOCKER_COMPOSE_QUIET_PULL=--quiet-pull
endif
KEYS_GEN_IMAGE:=ghcr.io/hpinc/krypton/krypton-test-runner

include .env

start: local deps

local: jwtprivateKey.pem
	$(DOCKER_COMPOSE) -p$(PROJECT) up $(DOCKER_COMPOSE_QUIET_PULL) -d

clean: stop
	-rm *.pem

test: local
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-test.yml up $(DOCKER_COMPOSE_QUIET_PULL) \
		--exit-code-from $(TEST)

deps: dstsprivateKey.pem
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-deps.yml up -d

jwtprivateKey.pem:
	touch jwtprivateKey.pem jwtpublicKey.pem # avoid docker file create permissions
	docker run --rm -v$(CURDIR):/tmp $(KEYS_GEN_IMAGE) /tmp/make_keys.sh jwt

dstsprivateKey.pem:
	touch dstsprivateKey.pem dstspublicKey.pem # avoid docker file create permissions
	docker run --rm -v$(CURDIR):/tmp $(KEYS_GEN_IMAGE) /tmp/make_keys.sh dsts

stop:
	-$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose.yml \
		-f docker-compose-deps.yml \
		-f docker-compose-test.yml down

show_data:
	docker exec $(DB).$(DOMAIN) psql -U $(DB_USER) -d $(DB_NAME) -c 'select id, tenant_id, device_id, user_id, status from enroll order by created_at desc'

clean_data:
	docker exec $(DB).$(DOMAIN) psql -U $(DB_USER) -d $(DB_NAME) -c 'delete from enroll'

.PHONY: start stop keys dsts local test
