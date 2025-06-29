## local test support
Supports local test environments with database, sqs and other required infrastructure support in docker containers.
During local test, the default `make` target is invoked. 

### common dependencies
common dependencies such as sqs are loaded via the `deploy` project. To avoid duplicating specs here, currently `es`
expects `deploy` project checked out at the same level as `es` (for eg. something like  `~/code/es` and `~/code/deploy`)
This practice has mixed acceptance. `es` is just going for convinience and the fact that `deploy` is the place where we
do integration testing for `krypton` components.


### .env file
To ensure some default behavior, a `.env.sample` is provided. Makefile target `.env` will try to fulfill by copying
`.env.sample` to `.env. Make sure you change `.env` to fit local test needs.
