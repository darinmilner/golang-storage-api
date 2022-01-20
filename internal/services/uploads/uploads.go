package uploads

import (
	"errors"
	"fileuploader/internal/config"
	"fileuploader/internal/filesystem"
	"fileuploader/pkg/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
)

type service struct {
	Config config.Config
}

func NewUploadService(config config.Config) *service {
	return &service{
		Config: config,
	}
}

func (s *service) getFileToUpload(r *http.Request, fieldName string) (string, error) {
	if err := r.ParseMultipartForm(s.Config.MaxUploadSize); err != nil {
		log.Println(err)
		return "", err
	}

	file, header, err := r.FormFile(fieldName) //name on form value
	if err != nil {
		return "", err
	}

	defer file.Close()

	// check mime type
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	// back to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		logger.Errorf("Upload Service Error: %v", err)
		return "", err
	}

	if !inSlice(s.Config.AllowedMimeTypes, mimeType.String()) {
		return "", errors.New("invalid file type uploaded")
	}

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

func inSlice(slice []string, val string) bool {
	for _, i := range slice {
		if i == val {
			return true
		}
	}

	return false
}

func (s *service) UploadFile(r *http.Request, destination string, field string, fs filesystem.FS) error {
	fileName, err := s.getFileToUpload(r, field)
	if err != nil {
		logger.Errorf("Upload Service: Error getting file %v", err)
		return err
	}

	if fs != nil {
		err = fs.Put(fileName, destination)
		if err != nil {
			logger.Errorf("Upload Service: Error putting file on remote server %v", err)
			return err
		}
	} else {
		// Save file on local server
		err = os.Rename(fileName, fmt.Sprintf("%s/%s", destination, path.Base(fileName)))
		if err != nil {
			logger.Errorf("Upload Service: Error saving file on local server %v", err)
			return err
		}
	}

	return nil
}
