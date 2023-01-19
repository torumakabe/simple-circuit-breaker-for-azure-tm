package main

import (
	"net/http"
	"os"

	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/functions/breaker"
	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/logger"
)

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/breaker", breaker.HandleBreaker)
	logger.Infof("About to listen on %v. Go to https://127.0.0.1%v/", listenAddr, listenAddr)
	logger.Fatal("Stop server: %v", http.ListenAndServe(listenAddr, nil))
}
