FROM golang:1.24-bookworm@sha256:08268bff0df66aff6d4f7fcf1b625fcf4f86fb7e6dbb5fdb8bb94f0920025ceb as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
