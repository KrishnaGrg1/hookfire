# FROM golang:1.25-alpine AS builder

# WORKDIR /src

# RUN apk add --no-cache ca-certificates

# COPY go.mod go.sum ./
# RUN go mod download

# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/hookfire ./cmd

# FROM alpine:3.21

# WORKDIR /app

# RUN apk add --no-cache ca-certificates tzdata

# COPY --from=builder /out/hookfire /app/hookfire
# COPY migrations /app/migrations

# EXPOSE 8080

# CMD ["/app/hookfire"]

# FROM golang:1.25-alpine

# WORKDIR /app

# RUN apk add --no-cache ca-certificates tzdata

# COPY go.mod go.sum ./
# RUN go mod download

# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hookfire ./cmd

# EXPOSE 8080

# CMD ["./hookfire"]


FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o hookfire ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/hookfire .

EXPOSE 8080

CMD ["./hookfire"]