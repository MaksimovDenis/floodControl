FROM golang:1.21.6-alpine AS builder

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o flood-control-task ./main.go

CMD ["./flood-control-task"]