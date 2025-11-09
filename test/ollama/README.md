kubectl --kubeconfig=/home/ant/.kube/gpu_config_sanjose001 apply -f  https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/main/deployments/static/nvidia-device-plugin.yml
kubectl --kubeconfig=/home/ant/.kube/gpu_config_sanjose001 get pods -A | grep nvidia
kube-system            nvidia-device-plugin-daemonset-lpp4z       0/1     ContainerCreating   0          13s



You absolutely can. This is a very common and powerful pattern: using a lightweight Go agent for orchestration while calling a dedicated Python service for the heavy AI lifting.

Your `ai_service` variable is the perfect starting point. We will build a small Go service that acts as an **"Adapter"** or **"Shim"**. This service will:

1.  Receive your *exact* `ai_service` JSON format.
2.  Translate that request into the format an open-source model server (like Ollama) understands.
3.  Forward the request to the model, get the response, and send it back to your main Go agent.

Since you're new to GPUs in Kubernetes, we'll go slowly. The best "first step" is to use **Ollama**, as it's the simplest way to get a production-grade API for LLaVA.

Here is the 3-step plan to get this working.

-----

### Step 1: Deploy a Model Server (Ollama) in Kubernetes

First, we need the "engine" that runs LLaVA on the GPU. Ollama is the perfect tool for this. It's a single container that serves models (like LLaVA) with a simple REST API.

**Prerequisite: The K8s GPU Driver**
Before you do anything, your Kubernetes cluster *must* have the **NVIDIA Device Plugin** installed. This plugin is what tells K8s that a node has a GPU and allows a `Deployment` to request it. Your cloud provider likely has a one-click install for this.

**1. The Deployment (The App)**
This `yaml` file tells K8s to run the Ollama container and, critically, to *give it one GPU*.

```yaml
# ollama-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-server
  namespace: ai-services
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ollama-server
  template:
    metadata:
      labels:
        app: ollama-server
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
        resources:
          limits:
            nvidia.com/gpu: 1 # This is the magic line that requests 1 GPU
        volumeMounts:
        - name: ollama-models
          mountPath: /root/.ollama # This stores your models on a persistent volume
      volumes:
      - name: ollama-models
        persistentVolumeClaim:
          claimName: ollama-models-pvc # You must create this PVC
```

**2. The Service (The Internal API Endpoint)**
This `yaml` file gives the `Deployment` a stable, internal-only name (`ollama-service.ai-services.svc.cluster.local`) so our other agents can find it.

```yaml
# ollama-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: ollama-service
  namespace: ai-services
spec:
  type: ClusterIP # Only reachable inside the K8s cluster
  selector:
    app: ollama-server
  ports:
  - protocol: TCP
    port: 11434
    targetPort: 11434
```

Once you `kubectl apply` these, you will have an internal API at `http://ollama-service.ai-services:11434` that runs LLaVA on a GPU.

-----

### Step 2: Create Your "Adapter" Service (in Go)

This is the new service that will match your `ai_service` format. It's a lightweight Go program that acts as a translator.

Its job is to receive your request, transform it, and call the `ollama-service` we just deployed.

```go
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
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	ApiKeyEnvVar  string `json:"api_key_env_var"`
}

// This is the format Ollama's API expects
type OllamaRequest struct {
	Model   string   `json:"model"`
	Prompt  string   `json:"prompt"`
	Images  []string `json:"images"` // For LLaVA, we send base64 encoded images
	Stream  bool     `json:"stream"`
}

// This is what our main Go agent will send to this Adapter
type AdapterRequest struct {
	Config  AiServiceConfig `json:"ai_service"`
	Prompt  string          `json:"prompt"`
	ImageB64 string         `json:"image_b64"` // A single base64 image for simplicity
}

// The internal URL for our Ollama service from Step 1
var ollamaInternalURL = "http://ollama-service.ai-services:11434/api/generate"

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
```

You would deploy this Go program as its own `Deployment` and `Service` (e.g., `ai-adapter-service`) inside your K8s cluster.

-----

### Step 3: How Your Agents Use It

Now, your Go agents can treat your self-hosted LLaVA *exactly* like they treat Anthropic, just by changing the config.

**Your Old Config:**

```json
"ai_service":{
  "provider":"anthropic",
  "model":"claude-haiku-4-5-20251001",
  "api_key_env_var":"ANTHROPIC_API_KEY"
}
```

**Your New Config (for LLaVA):**

```json
"ai_service":{
  "provider":"my-llava",
  "model":"llava:latest",
  "api_key_env_var":"MY_INTERNAL_API_KEY"
}
```

Your Go agent code doesn't change at all. It just sends its request to the `ai-adapter-service`'s endpoint (`/api/v1/generate`) instead of Anthropic's. The adapter handles all the translation.

-----

### Going Slowly: Your First Steps

Before you touch Kubernetes, let's get this working on your local machine.

1.  **Install Docker Desktop:** Make sure it's configured to use your local GPU.
2.  **Run Ollama Locally:** Open a terminal and run:
    ```bash
    # This command downloads and runs Ollama, mapping the port, and giving it GPU access
    docker run -d --gpus=all -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
    ```
3.  **Pull LLaVA:** Once the container is running, execute a command inside it to download the LLaVA model:
    ```bash
    docker exec -it ollama ollama pull llava
    ```
4.  **Test with `curl`:** From your *host machine's* terminal, you can now send a raw API request. This proves the server is working.
    ```bash
    curl http://localhost:11434/api/generate -d '{
      "model": "llava",
      "prompt": "What is in this image?",
      "images": ["/9j/4AAQSk... (your base64 image string) ..."],
      "stream": false
    }'
    ```
5.  **Run Your Adapter Locally:** Run the Go "Adapter" service from Step 2 on your local machine (e.g., `go run main.go`). Change the `ollamaInternalURL` to `http://localhost:11434/api/generate`.

Once you've completed these 5 steps, you'll have the *entire* system working on your local machine. Moving it to Kubernetes is then just a matter of translating your `docker run` commands into the `.yaml` files from Step 1.