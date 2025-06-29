## es-worker
Device enrollment worker for Krypton

### Dependencies
This project is compiled and run on linux virtual machines. Most of the testing is done in an Ubuntu vm.
```
NAME="Ubuntu"
VERSION="18.04.6 LTS (Bionic Beaver)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 18.04.6 LTS"
VERSION_ID="18.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=bionic
UBUNTU_CODENAME=bionic
```
Additionally,
- go version go1.19.1 linux/amd64
- make (whichever make comes standard with the base os. this one happens to be GNU Make 4.1)
- Docker version 20.10.17, build 100c701
- docker-compose version 1.17.1, build unknown
- default shell (bash should work but is not required)
### Project Principles
This project and all other projects under Krypton are trying to take the following principles forward
- Allow anybody to run the project with minimal local dependencies (see above)
- Allow anybody to package this project for local integration tests (see the deploy project)
- Have no additional package restrictions to deploy to cloud (in other words, build enough flexibility via config to deploy the locally tested containers to cloud)

### How to run locally
Use `make` to run without building. Default `make` target is designed to be in sync with the project principles of letting anybody run locally with minimal dependencies.

### How to package for distribution or integration tests
Use `make docker-image` to build docker image. Check out the deploy project on minimal integration samples.
