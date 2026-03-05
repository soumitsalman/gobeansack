package main

import (
	"context"
	"os"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	bs "github.com/soumitsalman/gobeansack/beansack"
	r "github.com/soumitsalman/gobeansack/router"
)

func main() {
	_ = godotenv.Load()
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connStr := getEnv("PG_CONNECTION_STRING", "", true)
	db := bs.NewPGSack(ctx, connStr)
	defer db.Close()

	api := r.NewRouter(&r.Configuration{
		DB:      db,
		APIKeys: parseAPIKeys(os.Getenv("API_KEYS")),
	})

	port := getEnv("PORT", "8080", false)
	addr := "0.0.0.0:" + port
	log.WithField("addr", addr).Info("Routes Initialized. Server starting...")
	bs.NoError(api.Run(addr), "server error")
}

func getEnv(name, fallback string, must_exist bool) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	if must_exist {
		log.Fatalf("%s is required\n", name)
	}
	return fallback
}

func parseAPIKeys(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	result := map[string]string{}
	for _, pair := range strings.Split(raw, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		header := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if header != "" && value != "" {
			result[header] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
