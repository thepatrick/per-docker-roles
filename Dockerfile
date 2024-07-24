FROM golang:1.22-bookworm@sha256:af9b40f2b1851be993763b85288f8434af87b5678af04355b1e33ff530b5765f as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
