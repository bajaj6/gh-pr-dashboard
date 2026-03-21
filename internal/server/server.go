package server

import (
	"context"
	"fmt"
	htmlpkg "html"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cli/go-gh/v2/pkg/browser"
)

// Logger is a minimal interface for progress logging.
type Logger interface {
	Logf(format string, a ...interface{})
}

// Serve starts an HTTP server that regenerates the dashboard on each request.
func Serve(port int, noOpen bool, log Logger, generateFn func() (string, error)) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.NotFound(w, r)
			return
		}

		html, err := generateFn()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<html><body><h1>Error generating dashboard</h1><pre>%s</pre></body></html>", htmlpkg.EscapeString(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write([]byte(html))
	})

	url := fmt.Sprintf("http://localhost:%d", port)
	log.Logf("Serving dashboard at %s", url)
	log.Logf("Press Ctrl+C to stop")

	if !noOpen {
		go func() {
			time.Sleep(500 * time.Millisecond)
			b := browser.New("", os.Stdout, os.Stderr)
			_ = b.Browse(url)
		}()
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown on Ctrl+C
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Logf("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	return server.Serve(ln)
}
