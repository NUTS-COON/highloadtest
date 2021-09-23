FROM alpine:latest as container

FROM golang:1.17 as dep
WORKDIR /tmp/build
COPY go.mod go.sum ./
RUN go mod download

FROM dep as builder
ARG service
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app main.go

FROM container
ARG service
WORKDIR /app
COPY --from=builder /tmp/build/${service}/app .
COPY config.yaml .
CMD ["./app"]