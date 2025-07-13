FROM ghcr.io/hpinc/krypton/krypton-go-builder as builder

ADD . /go/src/es
WORKDIR /go/src/es

# build the source
RUN make vet build

# use a minimal alpine image for services
FROM ghcr.io/hpinc/krypton/krypton-go-base

# set working directory
WORKDIR /go/bin

ENV GIN_MODE=release
COPY --from=builder /go/src/es/bin/es .
COPY --from=builder /go/src/es/service/config/config.yaml .
COPY --from=builder /go/src/es/service/config/token_config.yaml .
COPY --from=builder /go/src/es/service/db/schema /krypton/schema
COPY --from=builder /go/src/es/service/config/default_policy.json .

USER 1001
EXPOSE 7979/tcp

# run the binary
ENTRYPOINT ["/go/bin/es"]
