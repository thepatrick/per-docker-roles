FROM golang:1.22-bookworm@sha256:cabb8f5368ec3f6efde889d67dded25b8cafff861d83e493e53c83e4c8c1a023 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
