package handlers

import (
	"fileuploader/internal/filesystem/miniosystem"
	"fileuploader/internal/filesystem/s3aws"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v4"
	"github.com/go-chi/cors"
)

type service struct {
	minio miniosystem.Minio
	s3    s3aws.S3
}

//TODO: populate minio and s3 structs
//NewHttpHandler sets up the routes and handler service
func NewHttpHandler() http.Handler {
	s := &service{
		minio: miniosystem.Minio{},
		s3:    s3aws.S3{},
	}
	return s.routes()
}

func (s *service) routes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	// if c.Debug {
	// 	mux.Use(middleware.Logger)
	// }
	mux.Use(middleware.Recoverer)

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Vary", "Authorization", "Content-Type", "X-CSRF-Token", "XMLHttpRequest", "Access-Control-Allow-Origin", "Origin"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Get("/list-fs", s.ListFS)
	mux.Post("/files/upload", s.PostUploadToFS)
	mux.Get("/delete-from-fs", s.DeleteFromFS)

	return mux
}
