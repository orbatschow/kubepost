FROM golang:1.15 AS build

COPY . /go/src/kubepost
WORKDIR /go/src/kubepost
RUN go mod vendor && go build -o /go/bin/kubepost

FROM debian:stretch-slim

COPY --from=build /go/bin/kubepost /usr/bin/kubepost
