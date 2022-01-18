package handlers

import (
	"net/http"
)

type Handler func(rend Renderer, req *http.Request) error

type Renderer interface {
	JSON(interface{}, int) error
	Empty(int) error
}

type renderer struct {
	w http.ResponseWriter
}

func (r *renderer) Empty(i int) error {
	r.w.WriteHeader(i)
	return nil
}

//NewRenderer is the function that creates the renderer interface
func NewRenderer(w http.ResponseWriter) Renderer {
	return &renderer{w: w}
}

//JSON is the function that sends json
func (r *renderer) JSON(i interface{}, status int) error {
	SetHeaders(r.w, status)
	err := WriteJSON(r.w, status, i)
	if err != nil {
		ServerError(r.w, err)
		return err
	}

	return nil
}
