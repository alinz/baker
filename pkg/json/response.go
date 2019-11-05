package json

import (
	"encoding/json"
	"net/http"
)

func ResponseAsError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
