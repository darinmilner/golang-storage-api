package handlers

import (
	"fileuploader/internal/config"
	"fileuploader/internal/filesystem"
	"fileuploader/internal/filesystem/miniosystem"
	"fileuploader/internal/filesystem/s3aws"
	"fileuploader/internal/services/uploads"
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
	config        config.Config
	fs            filesystem.FS
	uploadService uploads.Upload
}

//NewHandler sets up the handler struct
func NewHandler(config config.Config, fs filesystem.FS, uploadService uploads.Upload) *handler {
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
func (s *service) PostUploadToFS(rend Renderer, r *http.Request) error {
	fileName, err := getFileToUpload(r, "formFile")
	if err != nil {
		return rend.JSON(err.Error(), http.StatusInternalServerError)

	}

	// get file to post to remote storage
	uploadType := r.Form.Get("upload-type")

	switch uploadType {
	case "MINIO":
		fs := s.minio
		err = fs.Put(fileName, "")
		if err != nil {
			return rend.JSON(err.Error(), http.StatusInternalServerError)
		}
	}

	// send json
	return rend.JSON("Upload successful", http.StatusCreated)
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

//DeleteFromFS deletes a file from the remote storage
func (s *service) DeleteFromFS(rend Renderer, r *http.Request) error {
	var fs filesystem.FS
	fsType := r.URL.Query().Get("fs_type")
	item := r.URL.Query().Get("file")
	switch fsType {
	case "MINIO":
		f := s.minio
		fs = &f
	case "AWS":
		f := s.s3
		fs = &f
	}

	deleted := fs.Delete([]string{item})
	if deleted {
		//send json
		return rend.JSON(fmt.Sprintf("%s was deleted", item), http.StatusOK)
	}

	return rend.JSON("Something went wrong", http.StatusBadRequest)
}

func (s *service) PostUpload(rend Renderer, r *http.Request) error {
	err := s.uploadService.UploadFile(r, "", "formFile", &s.s3)
	if err != nil {
		msg := fmt.Sprintf("upload err: %v", err)
		rend.JSON(msg, http.StatusInternalServerError)
	}

	return rend.JSON("image successfully uploaded. ", http.StatusCreated)
}

func (s *service) MaintenanceMode(rend Renderer, r *http.Request) error {
	var payload struct {
		IsMaintenanceMode bool   `json:"isMaintenanceMode"`
		AdminKey          string `json:"adminKey"`
	}
	err := ReadJSON(r, &payload)
	if err != nil {
		logger.Errorf("error unmarshalling json %v", err)
		return err
	}

	if payload.AdminKey != os.Getenv("ADMIN_KEY") {
		logger.Error("unauthorized")
		return rend.JSON("You are not authorized ", http.StatusForbidden)
	}
	maintenance := payload.IsMaintenanceMode
	if maintenance {
		s.maintenanceMode = true
	} else {
		s.maintenanceMode = false
	}

	rpcClient(maintenance)

	return rend.JSON("maintenance mode has been changed ", http.StatusOK)
}
