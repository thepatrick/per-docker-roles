FROM golang:1.24-bookworm@sha256:3ce988c30fa67dc966ca716ee0ce7ad08d7330573e808cb68ada7e419bdf23de as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
