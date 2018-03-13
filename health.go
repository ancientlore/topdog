package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/concur/go-health"
)

var (
	errNotDirectory = errors.New("Static path is not a directory")
)

var healthCheck = health.Tester{
	Log: func(testName, messageText, errorText string) {
		log.Print(testName+": "+messageText, ": ", errorText)
	},
	Tests: health.TestFuncs{
		"staticFiles": func(ctx context.Context) error {
			fi, err := os.Stat(*staticPath)
			if err != nil {
				return err
			}
			if fi.IsDir() != true {
				return errNotDirectory
			}
			filesToCheck := []string{"grim-reaper.png", "dog.png", "jquery.min.js", "dog.css", "index.html", "jquery-rotate.min.js"}
			for _, f := range filesToCheck {
				_, err = os.Stat(filepath.Join(*staticPath, f))
				if err != nil {
					return err
				}
			}
			for _, dog := range dogs {
				_, err = os.Stat(filepath.Join(*staticPath, dog+".png"))
				if err != nil {
					return err
				}
			}
			return nil
		},
	},
}
