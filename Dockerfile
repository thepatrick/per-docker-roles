FROM golang:1.22-bookworm@sha256:96788441ff71144c93fc67577f2ea99fd4474f8e45c084e9445fe3454387de5b as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
