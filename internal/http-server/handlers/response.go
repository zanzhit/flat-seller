package handlers

import (
	"net/http"

	"github.com/go-chi/render"
	resp "github.com/zanzhit/flat-seller/internal/lib/api/response"
)

func Error(w http.ResponseWriter, r *http.Request, statusCode int, err resp.Response) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, err)
}
