package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dlshle/dockman/internal/probe"
)

func respond(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Write(body)
}

func handleProbeRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respond(w, http.StatusInternalServerError, []byte("error reading request body"))
		return
	}
	var probeRequest probe.ProbeRequest
	err = json.Unmarshal(body, &probeRequest)
	if err != nil {
		respond(w, http.StatusBadRequest, []byte("invalid request format, failed to unmarshal: "+err.Error()))
		return
	}
	switch probeRequest.Config.Type {
	case probe.ProbeTypeHTTP:
		err = probeHTTP(probeRequest.Config.HTTP)
		if err != nil {
			respond(w, http.StatusInternalServerError, []byte("http probe failed: "+err.Error()))
			return
		}
		respond(w, http.StatusOK, []byte("probe succeeded"))
	default:
		respond(w, http.StatusBadRequest, []byte("unsupported probe type: "+probeRequest.Config.Type))
	}
}
