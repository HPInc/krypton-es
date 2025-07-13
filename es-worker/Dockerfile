FROM es-worker-cached-builder as builder

ADD . /go/src/es-worker
WORKDIR /go/src/es-worker

# build the source
RUN make vet build

# use a minimal alpine image for services
FROM ghcr.io/hpinc/krypton/krypton-go-base

# set working directory
WORKDIR /go/bin

COPY --from=builder /go/src/es-worker/bin/es-worker .
COPY --from=builder /go/src/es-worker/service/config/config.yaml .

USER 1001

# Expose REST port for health and metrics.
EXPOSE 8181

# run the binary
CMD ["/go/bin/es-worker"]
