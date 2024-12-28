package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/logrusorgru/aurora/v4"
)

type appConfig struct {
	Port      string
	ImagePath string
}

var config appConfig

func init() {
	config.Port = os.Getenv("PORT")
	if config.Port == "" {
		config.Port = "3000"
	}

	config.ImagePath = os.Getenv("IMAGE_PATH")
	if config.ImagePath == "" {
		config.ImagePath = "tmp-image"
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("image")
	if err != nil {
		log.Fatalf("Error Retrieving the File %v", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading file %v", err)
	}

	os.WriteFile(config.ImagePath, fileBytes, os.FileMode(0600))

	log.Println(aurora.Cyan("Successfully Uploaded File"))
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func image(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, config.ImagePath)
}

func livenessProbe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func main() {
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/image", image)
	http.HandleFunc("/healthz/liveness", livenessProbe)
	http.HandleFunc("/healthz/readiness", readinessProbe)

	err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
