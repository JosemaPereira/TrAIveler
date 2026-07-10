package middleware

import (
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 10 * 1024 * 1024 // 10 MB

// BodySize returns middleware that rejects requests whose body exceeds a
// 10 MB limit. Clients that declare an oversized Content-Length are rejected
// eagerly, before any body is read; requests with unknown or understated
// length (e.g. chunked transfer-encoding) are enforced lazily via
// http.MaxBytesReader as the body is consumed by the next handler.
func BodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxBodyBytes {
			writeTooLarge(w, r)
			return
		}

		responded := false
		r.Body = &limitedBodyReader{
			inner:     http.MaxBytesReader(w, r.Body, maxBodyBytes),
			w:         w,
			r:         r,
			responded: &responded,
		}
		next.ServeHTTP(w, r)
	})
}

func writeTooLarge(w http.ResponseWriter, r *http.Request) {
	writeErrorEnvelope(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body exceeds the 10 MB limit")
}

// limitedBodyReader wraps an http.MaxBytesReader so that, the first time a
// read fails because the limit was exceeded, the 413 error envelope is
// written automatically instead of leaving that translation to every caller.
// This assumes next reads the body before writing its own response headers;
// if next has already started responding, the 413 write is a best-effort
// no-op per net/http's superfluous-WriteHeader handling.
type limitedBodyReader struct {
	inner     io.ReadCloser
	w         http.ResponseWriter
	r         *http.Request
	responded *bool
}

func (l *limitedBodyReader) Read(p []byte) (int, error) {
	n, err := l.inner.Read(p)
	if err != nil && !*l.responded {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			*l.responded = true
			writeTooLarge(l.w, l.r)
		}
	}
	return n, err
}

func (l *limitedBodyReader) Close() error {
	return l.inner.Close()
}
