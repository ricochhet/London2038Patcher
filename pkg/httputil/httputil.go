package httputil

import (
	"net/http"
)

type ContentTypeValue string

const (
	ContentTypeZip  ContentTypeValue = "application/zip"
	ContentTypeJSON ContentTypeValue = "application/json; charset=utf-8"
	ContentTypeHTML ContentTypeValue = "text/html; charset=utf-8"
)

func ContentType(w http.ResponseWriter, ct ContentTypeValue) {
	w.Header().Set("Content-Type", string(ct))
}
