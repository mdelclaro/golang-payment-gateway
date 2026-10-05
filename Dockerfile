FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-gateway ./cmd/payment-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /bank-simulator ./cmd/bank-simulator

FROM alpine:3.22 AS runtime-base

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S app -G app

USER app

FROM runtime-base AS bank-simulator
COPY --from=build /bank-simulator /usr/local/bin/bank-simulator
EXPOSE 8080
ENTRYPOINT ["bank-simulator"]

FROM runtime-base AS gateway
COPY --from=build /payment-gateway /usr/local/bin/payment-gateway

EXPOSE 8090

ENTRYPOINT ["payment-gateway"]
