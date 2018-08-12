# wsproxy
wsproxy is a tunnel client and server base on websocket.
It is designed for the situation that you use a network behead a firewall who only let http request go.


wsproxy client connects to wsproxy server with websocket protocol which firewall will treat it as http. wsproxy server can connect to any tcp server whose address is specified by client. 

# Use case

## Connect to a ssh server

To connect to a ssh server listen on example.com:22, run server on a computer A outside the firewall by command:
> ./server -l "0.0.0.0:1234" 

And run client in a firewall protected network with command:
> ./client -s "ws://A:1234" -l ":5678" -r "example.com:22"

Then ssh to the localhost:5678 to connect to the remote ssh server. The data path is
> ssh-client -(tcp)-> 
> wsproxy-client -(websocket)-> 
> wsproxy-server -(tcp)->
> ssh-server
