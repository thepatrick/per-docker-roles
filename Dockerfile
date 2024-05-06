FROM golang:1.22-bookworm@sha256:d0902bacefdde1cf45528c098d14e55d78c107def8a22d148eabd71582d7a99f as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# TODO: Add a multi-stage build to reduce the size of the final image

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
