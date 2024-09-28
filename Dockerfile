FROM golang:1.23-bookworm@sha256:eac972dedeafc7b375b606672c0a453e4697a7eac308a205c5e3907b1eed2ab6 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
