package httpserver

import (
	"errors"
	"net/http"
)

func (d *delivery) handleOAuth2Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorMsg := r.URL.Query().Get("error")

	if errorMsg != "" {
		d.writeError(w, http.StatusBadRequest, errors.New(errorMsg))
		return
	}

	if code == "" || state == "" {
		d.writeError(w, http.StatusBadRequest, errors.New("code or/and state are empty"))
		return
	}

	if err := d.service.HandleRedirect(r.Context(), code, state); err != nil {
		d.writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
