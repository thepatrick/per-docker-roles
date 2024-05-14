FROM golang:1.22-bookworm@sha256:a53599b1a71631df417f5c45aeb9b34cc05122d0c83c635d17811191659b0936 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
