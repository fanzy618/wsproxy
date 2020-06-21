build:
	go build -ldflags "-w -s" -v -o wsproxy github.com/fanzy618/wsproxy

buildarm:
	env GOOS=linux GOARCH=arm go build -ldflags "-w -s" -v -o wsproxy.arm github.com/fanzy618/wsproxy
