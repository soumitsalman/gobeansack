package main

import (
	"os"
	"strconv"

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
	dbpath, ok := os.LookupEnv("DB_PATH")
	if !ok {
		dbpath = DEFAULT_DB_PATH
	}
	dim, err := strconv.Atoi(os.Getenv("VECTOR_DIM"))
	if err != nil {
		dim = DEFAULT_VECTOR_DIM
	}
	// Get cluster epsilon from env or use default
	cluster_eps, err := strconv.ParseFloat(os.Getenv("CLUSTER_EPS"), 64)
	if err != nil {
		cluster_eps = DEFAULT_CLUSTER_EPS
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// initialize database
	init, err := os.ReadFile("./factory/init.sql")
	noerror(err, "READ SQL ERROR")
	ds := NewBeansack(dbpath, string(init), dim, cluster_eps)
	defer ds.Close()

	r := setupRoutes(ds)
	noerror(r.Run(":"+port), "SERVER ERROR")
}
