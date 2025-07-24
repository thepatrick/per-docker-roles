FROM golang:1.24-bookworm@sha256:ef8c5c733079ac219c77edab604c425d748c740d8699530ea6aced9de79aea40 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
