package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

//payload is the json payload struct
var payload struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

//WriteJSON writes arbitrary data as JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Write(out)
	SetHeaders(w, status)
	return nil
}

//SetHeaders sets JSON headers
func SetHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

//ReadJSON reads arbitrary data as JSON
func ReadJSON(r *http.Request, data interface{}) error {
	//maxBytes := 1048576

	//r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single json value")
	}

	return nil
}

//BadRequest sends a bad request error ss JSON
func BadRequest(w http.ResponseWriter, r *http.Request, err error) error {
	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	SetHeaders(w, http.StatusBadRequest)
	w.Write(out)
	return nil
}

func InvalidCredentials(w http.ResponseWriter, msg string) error {
	payload.Error = true
	payload.Message = msg

	err := WriteJSON(w, http.StatusForbidden, payload)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

//ServerError sends a bad request error ss JSON
func ServerError(w http.ResponseWriter, err error) error {
	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	SetHeaders(w, http.StatusInternalServerError)
	w.Write(out)
	return nil
}
