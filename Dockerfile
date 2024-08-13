FROM golang:1.22-bookworm@sha256:96108288c59f09c0deb481579885dcee68e3384bffbf0ce5bf5a68ba40b330f8 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
