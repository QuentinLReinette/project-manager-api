# base image
FROM golang:1.26-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# dev image with Air for hot-reloading
FROM base AS development
RUN go install github.com/air-verse/air@latest
COPY . . 
CMD ["air"]

# build image for production
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /project-manager-api src/main.go

# final image for production
FROM alpine:3.23 AS production
RUN adduser -D appuser
USER appuser
WORKDIR /
COPY --from=builder /project-manager-api /project-manager-api
EXPOSE 8080
CMD ["/project-manager-api"]
