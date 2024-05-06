FROM debian:bookworm as builder

RUN apt-get update && apt-get install git build-essential golang-go -y

WORKDIR /app

COPY . /app/

RUN go build

# TODO: Add a multi-stage build to reduce the size of the final image

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
