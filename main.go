package main

import (
	"os"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_DB_PATH     = ""
	DEFAULT_VECTOR_DIM  = 384
	DEFAULT_CLUSTER_EPS = 0.43
)

func main() {
	// Load configuration from environment variables
	godotenv.Load(".env")
	// Read the configuration parameters
	catalog_path, ok := os.LookupEnv("CATALOG_PATH")
	if !ok {
		catalog_path = DEFAULT_DB_PATH
	}
	storage_path, ok := os.LookupEnv("STORAGE_PATH")
	if !ok {
		storage_path = DEFAULT_DB_PATH
	}
	ds := NewReadonlyBeansack(catalog_path, storage_path)
	defer ds.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := setupRoutes(ds)
	noerror(r.Run(":"+port), "SERVER ERROR")
}
