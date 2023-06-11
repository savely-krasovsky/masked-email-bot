package httpserver

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type Error struct {
	Error string `json:"error"`
}

func (d *delivery) writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if encodingErr := json.NewEncoder(w).Encode(Error{Error: err.Error()}); encodingErr != nil {
		d.logger.Error("Error while encoding error message!", zap.Error(encodingErr))
	}
}
