FROM golang:1.15-alpine

WORKDIR /go/src/kube-database-creator

COPY . .

RUN go install -v ./...

ENTRYPOINT [ "/go/bin/kube-database-creator" ]
