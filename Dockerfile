FROM golang:1.23-bookworm@sha256:238546619ba08be3d7896141645f12e81e02a3e37aee88f463c7ab488e78f0fb as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
