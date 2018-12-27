FROM golang:1.10-alpine3.8


RUN apk add --no-cache upx

WORKDIR /go/src/wsproxy
COPY . .

RUN go build -ldflags "-w -s" -v -o /wsproxy wsproxy && \
    upx /wsproxy


FROM alpine:3.8
RUN apk add --no-cache ca-certificates

WORKDIR /wsproxy
EXPOSE 1443 5004
ENTRYPOINT ["/wsproxy/run.sh"]
CMD ["/wsproxy/wsproxy"]

COPY --from=0 /wsproxy /go/src/wsproxy/run.sh ./
