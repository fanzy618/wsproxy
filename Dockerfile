FROM golang:1.10-alpine3.8


RUN apk add --no-cache upx 

WORKDIR /go/src/wsproxy
COPY . .

RUN go build -ldflags "-w -s" -v -o /server wsproxy/server && \
    go build -ldflags "-w -s" -v -o /client wsproxy/client && \
    upx /server && upx /client


FROM alpine:3.8
RUN apk add --no-cache ca-certificates

WORKDIR /wsproxy
EXPOSE 1443 5004
ENTRYPOINT ["/wsproxy/run.sh"]
CMD ["/wsproxy/server"]

COPY --from=0 /server /client /go/src/wsproxy/run.sh ./
