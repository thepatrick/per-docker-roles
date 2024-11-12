FROM golang:1.23-bookworm@sha256:1f001ad8c8d90281cd9d6e0ae4a40363039c148c503bcd483ff38c946b3d4f6d as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
