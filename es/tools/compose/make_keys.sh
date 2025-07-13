#!/bin/sh
# this script is meant to be run from a container to
# keep a steady openssl version to reduce key version conflicts
# generates a 2048 bit key by default
PREFIX=${1:-$1}

openssl genrsa -out /tmp/"$PREFIX"privateKey.pem
openssl rsa -in /tmp/"$PREFIX"privateKey.pem -pubout -out /tmp/"$PREFIX"publicKey.pem
