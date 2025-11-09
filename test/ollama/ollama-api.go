package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// This matches your app's config format
type AiServiceConfig struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	ApiKeyEnvVar string `json:"api_key_env_var"`
}

// This is the format Ollama's API expects
type OllamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"` // For LLaVA, we send base64 encoded images
	Stream bool     `json:"stream"`
}

// This is what our main Go agent will send to this Adapter
type AdapterRequest struct {
	Config   AiServiceConfig `json:"ai_service"`
	Prompt   string          `json:"prompt"`
	ImageB64 string          `json:"image_b64"` // A single base64 image for simplicity
}

// The internal URL for our Ollama service from Step 1
// var ollamaInternalURL = "http://ollama-service.ai-services:11434/api/generate"
var ollamaInternalURL = "http://localhost:11434"

func main() {
	http.HandleFunc("/api/v1/generate", handleGenerate)
	log.Println("Starting AI Adapter Service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req AdapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if this request is for our self-hosted service
	if req.Config.Provider != "my-llava" {
		http.Error(w, "Unsupported provider: "+req.Config.Provider, http.StatusBadRequest)
		return
	}

	// --- Authentication ---
	// Here, we check the API key sent from your agent
	// This is a simple shared secret
	internalApiKey := os.Getenv("MY_INTERNAL_API_KEY")
	clientApiKey := os.Getenv(req.Config.ApiKeyEnvVar)

	if internalApiKey == "" || clientApiKey != internalApiKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// --- 1. Translate our format TO Ollama's format ---
	ollamaReq := OllamaRequest{
		Model:  req.Config.Model, // e.g., "llava:latest"
		Prompt: req.Prompt,
		Stream: false,
	}
	// If the request includes an image, add it.
	if req.ImageB64 != "" {
		ollamaReq.Images = []string{req.ImageB64}
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// --- 2. Forward the request to the internal Ollama service ---
	resp, err := http.Post(ollamaInternalURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// --- 3. Stream the response back to the original caller ---
	// This just forwards the raw JSON response from Ollama.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
