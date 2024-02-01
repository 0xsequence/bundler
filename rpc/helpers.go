package rpc

import (
	"encoding/json"
	"net/http"
)

func (s *RPC) renderJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	buf, err := json.Marshal(v)
	if err != nil {
		s.GetLogger(r.Context()).Error("json.Marshal: failed to serialize response body", "err", err)
		// TODO: similar.. errorHandler(w, proto.ErrorInternal("failed to serialize response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf)
}
