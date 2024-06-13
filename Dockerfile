FROM golang:1.22-bookworm@sha256:3a751e5facec722492e78570374237c3c77f389c1d6b9f8372d1029f75078b88 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
