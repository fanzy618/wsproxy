FROM golang:1.13-alpine


RUN apk add --no-cache upx

WORKDIR /go/src
RUN mkdir -p /go/src/github.com/fanzy618/wsproxy
COPY . /go/src/github.com/fanzy618/wsproxy/

RUN go build -ldflags "-w -s" -v -o /wsproxy github.com/fanzy618/wsproxy && \
    upx /wsproxy


FROM alpine:3.9
RUN apk add --no-cache ca-certificates

WORKDIR /wsproxy
EXPOSE 1443 5004
ENTRYPOINT ["/bin/sh"]
CMD ["/wsproxy/wsproxy"]

COPY --from=0 /wsproxy ./
