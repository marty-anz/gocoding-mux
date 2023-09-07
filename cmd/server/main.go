package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type authUserID struct{}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID := "a-special-user-id"

		ctx := context.WithValue(req.Context(), authUserID{}, userID)

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(authUserID{}).(string)
	if !ok {
		return ""
	}

	return userID
}

func profiling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		log.Printf("%s %s %s", req.Method, req.RequestURI, start)
		log.Printf("userID: %s", UserIDFromContext(req.Context()))

		next.ServeHTTP(w, req)

		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()
	healthHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			// req.Method == http.MethodGet
			// req.Context()
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")

			userID := UserIDFromContext(req.Context()))
			if _, err := w.Write([]byte(fmt.Sprintf(`{"status": "ok", "userId": "%s"}`, userID))); err != nil {
				panic(err)
			}
		})

	// type controller struct {}
	//
	// func (c *controller) serveHTTP(w http.ResponseWriter, req *http.Request) {
	mux.Handle("/health", auth(profiling(healthHandler)))

	server := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              ":" + port(),
		Handler:           mux,
	}

	_ = server.ListenAndServe()
}

func port() string {
	if os.Getenv("PORT") != "" {
		return os.Getenv("PORT")
	} else {
		return "8080"
	}
}
