FROM golang:1.22-bookworm@sha256:d08a72b4dcc74f4b77f058fc025081bf86271de4bed01c76168dddbd0403b625 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
