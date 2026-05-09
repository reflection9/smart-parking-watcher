FROM golang:1.26-alpine AS builder

ARG SERVICE_NAME

WORKDIR /src

COPY shared/observability/ /src/shared/observability/

COPY app/${SERVICE_NAME}/go.mod app/${SERVICE_NAME}/go.sum /src/app/${SERVICE_NAME}/

WORKDIR /src/app/${SERVICE_NAME}
RUN go mod download

COPY app/${SERVICE_NAME}/ /src/app/${SERVICE_NAME}/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/service /app/service

CMD ["/app/service"]
