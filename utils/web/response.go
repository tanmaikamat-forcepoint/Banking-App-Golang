package web

import (
	"encoding/json"
	"net/http"
)

type WebResponse struct {
	StatusCode int
	Message    string
	Data       interface{}
}

func SendResponse(w http.ResponseWriter, response WebResponse) {
	w.WriteHeader(response.StatusCode)
	json.NewEncoder(w).Encode(response)

}
