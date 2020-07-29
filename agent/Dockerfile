FROM golang:latest AS builder

WORKDIR /app

#Get dependencies list and download them
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN mkdir /build

COPY . .

#compile the go files
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/agent ./cmd/agent

#second image to create light-weighted image
FROM alpine:latest

COPY --from=builder /build .

#download git dependencies
RUN apk update && \
	apk add openssh && \
	apk add git

#copy rsa files and terraform binary
COPY terraform /usr/bin
COPY id_rsa /root/.ssh/
COPY id_rsa.pub /root/.ssh/
COPY known_hosts /root/.ssh/
RUN chmod 600 /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa.pub
RUN chmod 600 /root/.ssh/known_hosts

COPY vojtah-sandbox-82d417bb102c.json .
RUN export SSH_PRIVATE_KEY=/root/.ssh/id_rsa
#default credentials for google cloud platform added
ENV GOOGLE_APPLICATION_CREDENTIALS="vojtah-sandbox-82d417bb102c.json"

ENTRYPOINT ["./agent"]