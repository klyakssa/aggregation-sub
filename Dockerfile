FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aggregation-sub ./cmd/aggregator

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN groupadd -r app && useradd -r -g app app

WORKDIR /app

COPY --from=builder /app/aggregation-sub .

RUN chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["./aggregation-sub"]