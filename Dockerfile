FROM golang:1.19-alpine3.18 AS buildstg
ENV GOPATH /wd
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
COPY . /go
WORKDIR /go
RUN go build -o /service cmd/service/main.go && \
    go build -o /worker cmd/worker/*.go

FROM alpine:3.18.6
COPY --from=buildstg /service /bin/service
COPY --from=buildstg /worker /bin/worker
COPY --from=buildstg /go/defaults.yml /defaults.yml
CMD ["service"] 
# default is service 
