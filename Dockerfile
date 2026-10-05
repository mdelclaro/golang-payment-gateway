FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-gateway ./cmd/payment-gateway

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S app -G app

COPY --from=build /payment-gateway /usr/local/bin/payment-gateway

USER app

EXPOSE 8090

ENTRYPOINT ["payment-gateway"]
