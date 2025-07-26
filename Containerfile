FROM golang:1.24@sha256:ef5b4be1f94b36c90385abd9b6b4f201723ae28e71acacb76d00687333c17282 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM ubuntu:24.04@sha256:a08e551cb33850e4740772b38217fc1796a66da2506d312abe51acda354ff061

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates=20240203 \
  && rm -rf /var/lib/apt/lists/*
RUN groupadd --gid 2000 appuser \
  && useradd --uid 3000 --gid appuser --shell /usr/sbin/nologin --system appuser

COPY --from=builder /app/server /usr/local/bin/server

RUN chown appuser:appuser /usr/local/bin/server \
  && chmod 750 /usr/local/bin/server

USER appuser

WORKDIR /home/appuser

EXPOSE 8080

ENTRYPOINT ["server"]
CMD []