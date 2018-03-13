/*
topdog is a service used to demonstrate Istio features.

It is designed to run in a 3-tier mode, with a UI, a middle tier, and a backend tier.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/facebookgo/flagenv"
)

var (
	appName = "topdog"

	dogs = []string{
		"amit",
		"cameron",
		"dan",
		"HD",
		"mike",
		"prashanth",
		"reuben",
	}

	port       = flag.Int("service_port", 5000, "Service port")
	staticPath = flag.String("static", filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "ancientlore", "topdog", "static"), "Location of static files")
	backendURL = flag.String("backend", "http://localhost:5000", "Location of backend API")
	midtierURL = flag.String("midtier", "http://localhost:5000", "Location of midtier API")
	version    = flag.Int("version", 1, "Version (1, 2, or 3)")
)

func main() {
	// parse flags & env vars
	flag.Parse()
	flagenv.Parse()

	// initialize logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// check static folder
	fi, err := os.Stat(*staticPath)
	if err != nil {
		log.Fatal(err)
	} else if !fi.IsDir() {
		log.Fatal(*staticPath, " is not a directory")
	}

	// initialize routes - all tiers
	http.Handle("/health", healthCheck)

	// initialize routes - backend tier
	http.Handle("/backend", gziphandler.GzipHandler(http.HandlerFunc(backEnd)))

	// initialize routes - mid tier
	http.Handle("/midtier", gziphandler.GzipHandler(http.HandlerFunc(midTier)))

	// initialize routes - UI tier
	http.Handle("/static/", gziphandler.GzipHandler(http.StripPrefix("/static/", http.FileServer(http.Dir(*staticPath)))))
	http.Handle("/query", gziphandler.GzipHandler(http.HandlerFunc(jsonQuery)))
	http.Handle("/", gziphandler.GzipHandler(http.HandlerFunc(ui)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second, // Time to read the request
		WriteTimeout: 10 * time.Second, // Time to write the response
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, os.Kill)
	go func(ctx context.Context) {
		done := ctx.Done()
		select {
		case <-done:
		case sig := <-stop:
			log.Print("Received signal ", sig.String())
			d := time.Second * 5
			if sig == os.Kill {
				d = time.Second * 15
			}
			wait, cancel := context.WithTimeout(ctx, d)
			defer cancel()
			err := server.Shutdown(wait)
			if err != nil {
				log.Print(err)
			}
		}
	}(context.Background())

	log.Printf(appName+" starting on port %d", *port)

	// listen for requests and serve responses.
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Print(appName + " shutting down")
}
