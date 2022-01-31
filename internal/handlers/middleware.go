package handlers

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//ErrorMiddleware is a middleware that wraps the handlers and returns errors
func ErrorMiddleware(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: handle the case when multierr is returned
		if err := h(NewRenderer(w), r); err != nil {
			res := errBucket{}
			s, ok := status.FromError(err)
			if ok {
				httpStatus := convertStatus(s.Code())
				if httpStatus >= http.StatusInternalServerError {
					res.set("error", "internal server error")
				} else {
					res.set("error", s.Message())
					for _, detail := range s.Details() {
						if fErr, fok := detail.(fieldError); fok {
							res.set(fErr.GetField(), fErr.GetError())
						}
					}
				}
				SetHeaders(w, httpStatus)
			} else {
				SetHeaders(w, http.StatusInternalServerError)
				res.set("error", "unexpected server error occurred")
			}
			WriteJSON(w, http.StatusOK, res)
		}

	}
}

func (s *service) CheckForMaintenanceMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.maintenanceMode {
			// add allowed routes
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Retry-After:", "300")
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
			WriteJSON(w, http.StatusServiceUnavailable, "server is under maintenance")
		}
		next.ServeHTTP(w, r)
	})
}

type fieldError interface {
	GetField() string
	GetError() string
}

type errBucket map[string][]string

func (b errBucket) set(field, msg string) {
	if messages, ok := b[field]; ok {
		b[field] = append(messages, msg)
	}

	b[field] = []string{msg}
}

//convertStatus converts GRPC to HTTP statuses
func convertStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
