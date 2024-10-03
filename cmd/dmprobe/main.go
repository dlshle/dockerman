package main

import (
	"flag"
	"net/http"
)

// a simple container application for container health checks
func main() {
	port := "18080"
	flag.StringVar(&port, "port", port, "port to listen on")
	// this is the dockman proxy that executes probe commands(http, shell, tcp, grpc(maybe))
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("only POST method is supported"))
			return
		}
		handleProbeRequest(w, r)
	})
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
