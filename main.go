package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/webdav"
	errors "golang.org/x/xerrors"
)

func main() {
	dir := flag.String("dir", "./", "webdav root dir")
	addr := flag.String("listen", "", "webdav listen addr")
	user := flag.String("user", "", "HTTP Basic Auth user(Optional)")
	password := flag.String("password", "", "HTTP Basic Auth password(Optional)")

	flag.Parse()

	if *addr == "" {
		_, _ = fmt.Fprintln(os.Stderr, "-listen must be specified")
		os.Exit(1)
	}

	var enableBasicAuth bool
	if *user != "" && *password != "" {
		enableBasicAuth = true
	}

	if _, err := os.Stat(*dir); err != nil {
		err = errors.Errorf("check dir %s failed: %w", err)
		log.Fatalf("%+v", err)
	}

	var handler http.Handler = &webdav.Handler{
		LockSystem: webdav.NewMemLS(),
		FileSystem: webdav.Dir(*dir),
	}

	mux := http.NewServeMux()

	if enableBasicAuth {
		handler = basicAuth(*user, *password, handler)
	}

	mux.Handle("/", handler)

	server := http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	idleWait := make(chan struct{})

	go func() {
		<-signalCh

		timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		if err := server.Shutdown(timeout); err != nil {
			err = errors.Errorf("shutdown webdav failed: %+w", err)
			log.Printf("%+v", err)
		}

		close(idleWait)
	}()

	err := server.ListenAndServe()
	switch {
	default:
		err = errors.Errorf("listen and serve failed: %w", err)
		log.Fatalf("%+v", err)

	case errors.Is(err, http.ErrServerClosed):
	}

	<-idleWait
}

func basicAuth(user, password string, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != password {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		handler.ServeHTTP(w, r)
	}
}
