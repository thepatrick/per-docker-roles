FROM golang:1.22-bookworm@sha256:69114624152f6cf230e19a47630184f0586581860ed7127224d252085aa908f9 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
