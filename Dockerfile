FROM golang:1.22-bookworm@sha256:1e08436b27d2f6a4e3c06a3bf135161a4e358190e5d36cf54bedd850f15037c7 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
