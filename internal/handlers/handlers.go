package handlers

import (
	"fileuploader/internal/config"
	"fileuploader/internal/filesystem"
	"fileuploader/internal/filesystem/miniosystem"
	"fileuploader/internal/filesystem/s3aws"
	"fileuploader/pkg/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

//handler is the handler struct
type handler struct {
	config config.Config
	fs     filesystem.FS
}

//NewHandler sets up the handler struct
func NewHandler(config config.Config, fs filesystem.FS) *handler {
	return &handler{
		config: config,
		fs:     fs,
	}
}

//ListFS shows files in an external cloud storage
func (s *service) ListFS(w http.ResponseWriter, r *http.Request) {
	var fs filesystem.FS
	var list []filesystem.Listing
	fsType := ""
	if r.URL.Query().Get("fs-type") != "" {
		fsType = r.URL.Query().Get("fs-type")
	}

	curPath := "/"
	if r.URL.Query().Get("curPath") != "" {
		curPath = r.URL.Query().Get("curPath")
		curPath, _ = url.QueryUnescape(curPath)
	}

	if fsType != "" {
		switch fsType {
		case "MINIO":
			var f miniosystem.Minio
			fs = &f
			fsType = "MINIO"
		case "S3":
			var f s3aws.S3
			fs = &f
			fsType = "S3"
		}

		l, err := fs.List(curPath)
		if err != nil {
			log.Println(err)
			return
		}

		list = l
	}

	//TODO: Add JSON

	logger.Info(list)
}

//PostUploadToFS uploads a file to the FS
func (s *service) PostUploadToFS(w http.ResponseWriter, r *http.Request) {
	fileName, err := getFileToUpload(r, "formFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get file to post to remote storage
	uploadType := r.Form.Get("upload-type")

	switch uploadType {
	case "MINIO":
		fs := s.minio
		err = fs.Put(fileName, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//TODO: send json
}

//getFileToUpload uploads a file to the remote storage system
func getFileToUpload(r *http.Request, fieldName string) (string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		return "", err
	}

	file, header, err := r.FormFile(fieldName) //name on form value
	if err != nil {
		return "", err
	}

	defer file.Close()

	// create a file
	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

//DeleteFromFS deletes a file from the remote storage  TODO: add AWS
func (s *service) DeleteFromFS(w http.ResponseWriter, r *http.Request) {
	var fs filesystem.FS
	fsType := r.URL.Query().Get("fs_type")
	item := r.URL.Query().Get("file")
	switch fsType {
	case "MINIO":
		f := s.minio
		fs = &f
	}

	deleted := fs.Delete([]string{item})
	if deleted {
		//send json
		WriteJSON(w, http.StatusOK, fmt.Sprintf("%s was deleted", item))
	}

	WriteJSON(w, http.StatusBadRequest, "Something went wrong")
}
