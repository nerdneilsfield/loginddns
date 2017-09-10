package main

import (
	"fmt"
	"net"
	"net/http"
)

func getIP(w http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", req.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", req.RemoteAddr)
		return
	}
	fmt.Fprintf(w, "%s", ip)
}

func main() {
	http.HandleFunc("/", getIP)
	http.ListenAndServe("0.0.0.0:1080", nil)
}
