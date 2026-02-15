// =============================================================================
// Aplicação de Exemplo - Pipeline Hardening PoC
// meluansantos/secure-pipeline-poc
// =============================================================================
// Esta é uma aplicação Go mínima para demonstração do pipeline hardenado.
// =============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	Timestamp string `json:"timestamp"`
}

type InfoResponse struct {
	App         string `json:"app"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
	Hardened    bool   `json:"hardened"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		GoVersion: runtime.Version(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	response := InfoResponse{
		App:         "Pipeline Hardening PoC",
		Description: "Demonstração de pipeline CI/CD hardenado com gVisor, Falco, SBOM e Cosign",
		Author:      "meluansantos",
		Repository:  "https://github.com/meluansantos/secure-pipeline-poc",
		Hardened:    true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pipeline-hardening-poc\n\nGET /health  -> status do servidor\nGET /info    -> metadata da app\n")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-health" {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/info", infoHandler)

	log.Printf("listening on :%s", port)
	log.Printf("version=%s commit=%s", version, gitCommit)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
