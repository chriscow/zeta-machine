FROM golang:alpine as builder

RUN mkdir /build

ADD . /build/
WORKDIR /build

RUN go build -o webz ./cmd/web/.

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/webz /app/

WORKDIR /app
CMD ["./webz"]