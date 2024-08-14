FROM golang:1.23-bookworm@sha256:537d736a44beb60e4eb013e2bc056a7e9d6fe6eb63e3363530c5edd83d64a93b as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
