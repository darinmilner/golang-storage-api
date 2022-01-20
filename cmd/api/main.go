package main

import (
	"fileuploader/internal/config"
	"fileuploader/internal/filesystem/miniosystem"
	"fileuploader/internal/filesystem/s3aws"
	"fileuploader/internal/handlers"
	"fileuploader/pkg/logger"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	initApp()
	go listenForShutdown()
	ListenAndServe()

}

//initApp starts the app
func initApp() map[string]interface{} {
	config, err := config.NewConfig()
	if err != nil {
		logger.Fatalf("Config can not be created %v", err)
	}
	return createFileSystem(*config, config.RemoteFSName)
}

//createFileSystem creates the file systems for MINIO and AWS
func createFileSystem(config config.Config, providerName string) map[string]interface{} {
	remoteFileSystem := make(map[string]interface{})

	switch providerName {
	case "MINIO":
		if config.Minio.Secret != "" {
			useSLL := false
			if config.Minio.UseSSL {
				useSLL = true
			}

			minio := miniosystem.Minio{
				Endpoint: config.Minio.Endpoint,
				Key:      config.Minio.Key,
				Secret:   config.Minio.Secret,
				UseSSL:   useSLL,
				Region:   config.Minio.Region,
				Bucket:   config.Minio.Bucket,
			}
			remoteFileSystem["MINIO"] = minio
		}
	case "S3":
		s3 := s3aws.S3{
			Key:      config.AWS.Key,
			Secret:   config.AWS.Secret,
			Region:   config.AWS.Region,
			Endpoint: config.AWS.Endpoint,
			Bucket:   config.AWS.Bucket,
		}
		remoteFileSystem["S3"] = s3
	default:
		logger.Errorf("Invalid Remote file system configured")
	}
	return remoteFileSystem
}

//listenForShutdown makes a graceful shutdown
func listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	log.Println("Received signal", s.String())
	shutdown()

	os.Exit(0)
}

//shutdown adds cleanup tasks  TODO:  add clean up tasks
func shutdown() {
	// put any clean up tasks here

	// block until the WaitGroup is empty
	var wg sync.WaitGroup
	wg.Wait()
}

//ListenAndServe starts the server
func ListenAndServe() error {
	handlers := handlers.NewHttpHandler()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      handlers,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	//TODO: Add GRPC
	//go c.listenRPC()
	log.Printf("Listening on port %s", os.Getenv("PORT"))
	return srv.ListenAndServe()
}
