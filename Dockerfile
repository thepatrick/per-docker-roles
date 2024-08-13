FROM golang:1.22-bookworm@sha256:605adc6fa6a40acd8abf8f3610d9d7037f17f6501c0133dbf16c3abc02f525c4 as builder

WORKDIR /app

COPY . /app/

RUN go version \
  && go build

# The runtime image, used to just run the code provided its virtual environment
FROM scratch

COPY --from=builder /app/per-docker-roles /per-docker-roles

CMD ["/per-docker-roles"]
