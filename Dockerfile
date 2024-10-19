FROM golang:1.23-bookworm@sha256:2341ddffd3eddb72e0aebab476222fbc24d4a507c4d490a51892ec861bdb71fc as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
