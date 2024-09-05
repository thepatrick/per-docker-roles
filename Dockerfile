FROM golang:1.23-bookworm@sha256:32096e84705b30bb39cc9c65ef2896efacc4268203b7876049847763cefc934d as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
