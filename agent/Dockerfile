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

RUN apk update && \
	apk add openssh && \
	apk add git

COPY terraform /usr/bin
COPY id_rsa /root/.ssh/
COPY id_rsa.pub /root/.ssh/
COPY known_hosts /root/.ssh/
RUN chmod 600 /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa.pub
RUN chmod 600 /root/.ssh/known_hosts

COPY vojtah-sandbox-82d417bb102c.json .
RUN export SSH_PRIVATE_KEY=/root/.ssh/id_rsa
RUN export GOOGLE_APPLICATION_CREDENTIALS="vojtah-sandbox-82d417bb102c.json"

ENTRYPOINT ["./agent"]