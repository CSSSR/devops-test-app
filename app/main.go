package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/logrusorgru/aurora"
)

type appConfig struct {
	Port               string
	ImagePath          string
	MutexWaitingPeriod time.Duration
	MutexTTL           time.Duration
	RedisHost          string
	RedisPort          string
	RedisDB            int
}

var (
	config appConfig
	ctx    = context.Background()
	rdb    *redis.Client
)

// Init application configuration
func init() {
	config.Port = getEnv("PORT", "3000")
	config.ImagePath = getEnv("IMAGE_PATH", "tmp-image")

	config.RedisHost = getEnv("REDIS_HOST", "localhost")
	config.RedisPort = getEnv("REDIS_PORT", "6379")
	config.RedisDB = mustAtoi(getEnv("REDIS_DB", "0"))

	config.MutexWaitingPeriod = mustDuration(getEnv("MUTEX-WAITING-PERIOD", "10s"))
	config.MutexTTL = mustDuration(getEnv("MUTEX-TTL", "30s"))

	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		DB:   config.RedisDB,
	})

	log.Println(aurora.Green("Service configuration loaded successfully"))
}

// Utility functions for env variables
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustAtoi(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Failed to parse integer: %v", err)
	}
	return val
}

func mustDuration(s string) time.Duration {
	val, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("Failed to parse duration: %v", err)
	}
	return val
}

// Acquire Mutex
func acquireMutex(key string, ttl time.Duration) bool {
	resp, err := rdb.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		log.Println(aurora.Red("Error acquiring mutex: "), err)
	}
	return resp
}

// Release Mutex
func releaseMutex(key string) {
	rdb.Del(ctx, key)
}

// Background Mutex TTL updater
func keepMutexAlive(key string, ttl time.Duration, stopChan chan struct{}) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rdb.Expire(ctx, key, ttl)
		case <-stopChan:
			return
		}
	}
}

// Handler: Upload file
func uploadFile(w http.ResponseWriter, r *http.Request) {
	mutexKey := "image-upload-mutex"
	start := time.Now()

	// Wait for mutex
	for !acquireMutex(mutexKey, config.MutexTTL) {
		log.Println(aurora.Yellow("Mutex locked, waiting..."))
		time.Sleep(1 * time.Second)

		if time.Since(start) > config.MutexWaitingPeriod {
			http.Error(w, "Timeout waiting for mutex", http.StatusRequestTimeout)
			return
		}
	}
	defer releaseMutex(mutexKey)

	// Start background process to update TTL
	stopChan := make(chan struct{})
	go keepMutexAlive(mutexKey, config.MutexTTL, stopChan)
	defer close(stopChan)

	// File processing
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Save image to Redis
	imageKey := "uploaded-image"
	err = rdb.Set(ctx, imageKey, fileBytes, 0).Err()
	if err != nil {
		http.Error(w, "Error saving image to Redis", http.StatusInternalServerError)
		return
	}

	log.Println(aurora.Cyan("Successfully uploaded file to Redis"))
	fmt.Fprintf(w, "Successfully uploaded file to Redis\n")
}

// Handler: Serve image
func image(w http.ResponseWriter, r *http.Request) {
	imageKey := "uploaded-image"
	imageData, err := rdb.Get(ctx, imageKey).Bytes()
	if err != nil {
		http.Error(w, "No image found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(imageData)
}

// Handler: Metrics endpoint
func metrics(w http.ResponseWriter, r *http.Request) {
	metrics := fmt.Sprintf("uptime: %s\nmutex: %s", time.Since(startTime), "ok")
	fmt.Fprintf(w, metrics)
}

// Liveness and Readiness probes
func livenessProbe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

var startTime = time.Now()

func main() {
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/image", image)
	http.HandleFunc("/healthz/liveness", livenessProbe)
	http.HandleFunc("/healthz/readiness", readinessProbe)
	http.HandleFunc("/metrics", metrics)

	log.Println(aurora.Green("Starting server on port " + config.Port))
	err := http.ListenAndServe(":"+config.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
