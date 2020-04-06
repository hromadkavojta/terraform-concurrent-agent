FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN mkdir /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/agent ./cmd/agent

FROM alpine:latest

COPY --from=builder /build .
COPY terraform /usr/bin

RUN ls /usr/bin
ENTRYPOINT ["./agent"]