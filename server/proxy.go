package server

import (
	"context"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func ProxyMain(ctx context.Context, addr string) {
	proxyHandler := goproxy.NewProxyHttpServer()
	proxyServer := &http.Server{
		Addr:    addr,
		Handler: proxyHandler,
	}
	defer proxyServer.Close()
	go log.Fatal(proxyServer.ListenAndServe())
	select {
	case <-ctx.Done():
		return
	}
}
