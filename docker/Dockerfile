FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git
RUN mkdir /opt/dist

WORKDIR /opt/build

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Must be at repository root
COPY . .

ARG GOOS=linux
ARG GOARCH=amd64

WORKDIR /opt/build/cmd/jobber
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -o /opt/dist/jobber .

FROM scratch AS export
COPY --from=builder /opt/dist/jobber /
