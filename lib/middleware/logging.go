package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Creem un wrapper pel ResponseWriter per capturar l'status i la mida
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK, // valor per defecte
		}

		// Cridem al següent handler
		next.ServeHTTP(wrapped, r)

		// Logging al format típic d'Apache:
		// IP - [Data] "Mètode Path Protocol" Status Mida "Referer" "User-Agent" Temps
		log.Printf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %v",
			r.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.Path,
			r.Proto,
			wrapped.status,
			wrapped.size,
			r.Referer(),
			r.UserAgent(),
			time.Since(start),
		)
	})
}
