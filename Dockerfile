FROM golang:1.23-bookworm@sha256:441f59f8a2104b99320e1f5aaf59a81baabbc36c81f4e792d5715ef09dd29355 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
