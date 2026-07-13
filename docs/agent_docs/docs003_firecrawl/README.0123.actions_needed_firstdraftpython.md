This is the perfect set of questions to bridge our plan with your production system. Your Go-based architecture is ideal for this.

Here’s a breakdown of how we'll handle these new adapters and actions.

-----

## 1\. 🤝 Go to Python Adapter: The Microservice Pattern

You are 100% correct, we will not be running Python inside your Go agent containers. That would be a nightmare for dependency management.

The solution is the **Adapter Microservice Pattern**, which you're already halfway to using. Your Go `executeStep` function will make a simple, internal HTTP (or gRPC) call to a new, dedicated Python service that we'll deploy inside your K8s cluster.

Here is the step-by-step for creating the `run_playwright_capture` adapter:

1.  **Create the Python "Worker" Service (e.g., `playwright-adapter`)**

    * We'll write a simple Python app using a lightweight web server like **FastAPI** or **Flask**.
    * This app will have one endpoint, e.g., `POST /capture`.
    * This endpoint will receive a JSON body: `{"url": "...", "s3_bucket": "..."}`.
    * It will then execute the Playwright logic, capture the screenshot/DOM, and upload them directly to your Backblaze S3 bucket.
    * It returns a JSON response with the S3 paths: `{"screenshot_path": "...", "dom_path": "..."}`.

2.  **Containerize the Python Service**

    * We'll create a `Dockerfile` for this FastAPI app.
    * This `Dockerfile` will install Python, `pip install playwright fastapi uvicorn boto3`, and run `playwright install` (to get the browsers).
    * We build and push this image (e.g., `my-registry/playwright-adapter:v1`) to your image repository.

3.  **Deploy the Adapter in Kubernetes**

    * We create a standard K8s `Deployment` and `Service` (`playwright-adapter-service`). This service runs the container from Step 2. It doesn't need a GPU.

4.  **Update Your Go `executeStep` Function**

    * You'll add a new `case` to your `switch` statement in Go:
      ```go
      // In agent.go -> executeStep
      case "run_playwright_capture":
          // 1. Resolve config (e.g., get the URL)
          targetURL := resolveContext(runContext, step.Config["url"])
          
          // 2. Build the JSON request for our Python service
          reqBody, _ := json.Marshal(map[string]string{"url": targetURL})
          
          // 3. Make the internal K8s HTTP request
          // We'll get this URL from a config map or env var
          adapterURL := "http://playwright-adapter-service.default.svc.cluster.local:8000/capture"
          resp, err := http.Post(adapterURL, "application/json", bytes.NewBuffer(reqBody))
          
          // 4. Handle response, parse the resulting JSON, and save to runContext
          var result map[string]interface{}
          json.NewDecoder(resp.Body).Decode(&result)
          saveToContext(runContext, step.Name, result)
      ```

This pattern is clean, scalable, and keeps your Go orchestrator lightweight. We will use this *exact same pattern* for **all** our Python-based actions (LLaVA, CodeLlama, XY-Cut, etc.).

-----

## 2\. 🌩️ External LLaVA (GPU-as-a-Service)

This is excellent news. Using an external provider like "ThunderCompute" (or Replicate, Anyscale, etc.) is *simpler* than managing your own GPU nodes.

**This is just another adapter.** We don't need to change the `AgentWorkflow` JSON I proposed at all. We just change what our adapter *does*.

1.  **Provider Setup:** You'll provision a LLaVA model on your GPU provider. They will give you two things:

    * A public **API Endpoint** (e.g., `https://api.thundercompute.com/v1/llava-generate`)
    * A **Secret API Key** (e.g., `tc_sk_...`)

2.  **Create the `llava-adapter-service` (in Go)**

    * We'll create another small Go microservice (a new `Deployment` in K8s). This service is a **secure proxy and translator**.
    * Its job is to receive the *internal* request from your agents and turn it into the *external* request for ThunderCompute.
    * It will have its own endpoint: `POST /api/v1/generate`.

3.  **The Translation Logic:**

    * Your `layout-labeler` agent (Agent 3) calls `http://llava-adapter-service/api/v1/generate` (internally) with the `ai_service` config: `{"provider": "my-llava", "model": "llava:latest", ...}`.
    * The **`llava-adapter-service`** receives this.
    * It checks the `provider` name.
    * It loads the *real* `THUNDERCOMPUTE_API_KEY` from its own K8s environment variables (as a `Secret`).
    * It reformats the JSON to match ThunderCompute's API spec.
    * It makes the **external, authenticated** call to `https://api.thundercompute.com/...`
    * It gets the response and passes it back to the `layout-labeler` agent.

**Why this is the perfect architecture:**

* **Security:** Your `THUNDERCOMPUTE_API_KEY` is safely stored in *one* adapter's environment, not known by any of the 100+ agent definitions.
* **Modularity:** If you ever get a better deal from "Replicate" or decide to bring a GPU *inside* your K8s cluster, you **only update this one adapter service**. All your `AgentWorkflow` JSONs remain unchanged.

-----

## 3\. 📝 All New Actions We Need to Think About

Here is the complete list of new `action` types (adapters) we'll need to add to your Go `executeStep` function to implement the full 11-agent plan.

### Design Ingestion Group

* **`http_get_text`** (for Agent 1: Profiler)
    * **Adapter:** A simple, built-in Go `http.Get` call with basic text parsing.
* **`run_playwright_capture`** (for Agent 2: Capture Bot)
    * **Adapter:** Go -\> `http.Post` -\> **`playwright-adapter-service` (Python)**
* **`run_cv_layout_cut`** (for Agent 3: Layout Labeler)
    * **Adapter:** Go -\> `http.Post` -\> **`cv-adapter-service` (Python/OpenCV)**
* **`run_llava_label`** (for Agent 3: Layout Labeler)
    * **Adapter:** Go -\> `http.Post` -\> **`llava-adapter-service` (Go Proxy)** -\> External GPU API
* **`run_vlm_generate_code`** (for Agent 4: Component Generator)
    * **Adapter:** Go -\> `http.Post` -\> (This can re-use the `llava-adapter-service` if the provider is the same, just with a different model and prompt).
* **`run_playwright_get_style`** (for Agent 5: Style Extractor)
    * **Adapter:** Go -\> `http.Post` -\> **`playwright-adapter-service` (Python)** (a different endpoint, e.g., `/get-style`).
* **`run_llm_refactor_code`** (for Agent 6: Behavior Extractor)
    * **Adapter:** Go -\> `http.Post` -\> (This can call your existing Claude/Gemini API, or a new **`codellama-adapter-service`** for an open-source model).

### Library & Generation Groups

* **`db_insert_component`** (for Agent 7: Librarian)
    * **Adapter:** Internal Go function. This action handler will have the `*sql.DB` connection to your Postgres Vector DB and will run the `INSERT` query.
* **`db_query_component`** (for Agent 9: Architect & Agent 10: Strategist)
    * **Adapter:** Internal Go function. This action handler will run the `SELECT ... ORDER BY (embedding <=> $1)` query against your Vector DB.
* **`request_human_approval`** (for HITL)
    * **Adapter:** Internal Go function. As we discussed, this action will write to your `human_approval_queue` table (or produce a Kafka message) and set the `agent_run` status to `pending_human_input`.


--

This is perfect. The `image_actions.go` and `dynamic_adapter.go` files are the exact "Rosetta Stone" we need.

Your existing architecture is a **Kafka-based Asynchronous Adapter Pattern**. It's modular, scalable, and exactly what we should use. My previous suggestions for HTTP/REST adapters were just an assumption; this Kafka-based approach is far superior for your system.

You are **100% correct**. We will not change your Go orchestration code. We will simply:

1.  Create new **Go action files** (like `image_actions.go`) for each new capability.
2.  Create new **Python adapters** (like `dynamic_adapter.go`) that listen on new Kafka topics.

This is the most modular and maintainable path. Here is the detailed breakdown.

-----

### 1\. 🐍 How to Build Your Python Adapters (e.g., Playwright)

We will precisely follow the pattern in your `image_actions.go` and `dynamic_adapter.go` files.

#### Step 1: The Go Action (The "Requestor")

We create a new file, `playwright_actions.go`, in your Go `agent-chassis/internal/actions/` package. This function's *only* job is to format and send a Kafka message.

```go
// in playwright_actions.go
package actions

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    // ... other imports
)

// This tells the orchestrator to wait for the response
type CaptureSiteResult struct {
    Success       bool   `json:"success"`
    RequestID     string `json:"request_id"`
    TopicSentTo   string `json:"topic_sent_to"`
    AwaitResponse bool   `json:"await_response"` // This will be true
}

func CaptureSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
    params.Logger.Info("Executing CaptureSiteAction", zap.String("step_name", params.ExecutionContext.StepName))

    // 1. Get data from the workflow
    // The workflow JSON will pass the URL in the config
    urlToCapture, _ := params.StepConfig.Config["url"].(string)
    if urlToCapture == "" {
        return nil, fmt.Errorf("url is required for CaptureSiteAction")
    }

    // 2. Define our topics
    adapterRequestTopic := "system.adapter.playwright.requests"
    // The orchestrator will listen on this dynamic topic for the response
    myResponsesTopic := params.ExecutionContext.ResponsesTopic

    // 3. Build the message body for the Python adapter
    requestBody := map[string]interface{}{
        "action": "capture_site",
        "data": map[string]interface{}{
            "url":              urlToCapture,
            "s3_bucket":      "my-backblaze-bucket", // Or get from env/config
            "s3_path_prefix": fmt.Sprintf("captures/%s", params.ExecutionContext.CorrelationID),
        },
        "reply_to_topic": myResponsesTopic,
    }

    // 4. Build the full Kafka message (just like in image_actions.go)
    newRequestID := uuid.NewString()
    adapterRequest := map[string]interface{}{
        "headers": map[string]interface{}{
            "correlation_id":   params.ExecutionContext.CorrelationID,
            "orchestration_id": params.ExecutionContext.OrchestrationID,
            "step_name":        params.ExecutionContext.StepName,
            "request_id":       newRequestID,
            "message_type":     "request",
            "responses_topic":  myResponsesTopic,
            "sender_agent_type": params.ExecutionContext.Sender.AgentType,
            // ... all other required headers
        },
        "body": requestBody,
    }

    // 5. Send the message
    headers := datahelpers.BuildKafkaHeaders(adapterRequest["headers"]) // Assuming you have a helper
    messageBytes, _ := json.Marshal(adapterRequest)
    
    if err := params.Producer.ProduceWithValidation(
        ctx,
        adapterRequestTopic,
        headers,
        []byte(params.ExecutionContext.CorrelationID),
        messageBytes,
    ); err != nil {
        return nil, fmt.Errorf("failed to send to playwright adapter: %w", err)
    }

    // 6. Tell the orchestrator to wait
    result := CaptureSiteResult{
        Success:       true,
        RequestID:     newRequestID,
        TopicSentTo:   adapterRequestTopic,
        AwaitResponse: true,
    }
    
    return result, nil
}
```

#### Step 2: The Python Adapter (The "Worker")

This is a *new K8s deployment*. It's a Python script that mimics `dynamic_adapter.go`.

* **Dockerfile:** Installs `python`, `kafka-python`, `playwright`, and `boto3` (for Backblaze).
* **main.py (Simplified):**

<!-- end list -->

```python
import os
import json
from kafka import KafkaConsumer, KafkaProducer
from playwright.sync_api import sync_playwright
# ... import boto3, etc.

KAFKA_BROKERS = os.environ.get("KAFKA_BROKERS", "kafka:9092")
REQUEST_TOPIC = "system.adapter.playwright.requests"

consumer = KafkaConsumer(
    REQUEST_TOPIC,
    bootstrap_servers=KAFKA_BROKERS,
    auto_offset_reset='earliest',
    group_id="playwright-adapter-group"
)

producer = KafkaProducer(bootstrap_servers=KAFKA_BROKERS)

def capture_site(url, s3_bucket, s3_path_prefix):
    # This is the actual work
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        page.goto(url)
        
        # Capture screenshot and DOM
        ss_bytes = page.screenshot(full_page=True)
        dom_content = page.content()
        browser.close()
        
        # Upload to S3 (Backblaze)
        ss_path = f"{s3_path_prefix}/desktop.png"
        dom_path = f"{s3_path_prefix}/page.html"
        
        # ... boto3 upload logic ...
        
        return {"screenshot_path": ss_path, "dom_path": dom_path}

for message in consumer:
    try:
        data = json.loads(message.value.decode('utf-8'))
        headers_map = {k: v.decode('utf-8') for (k, v) in message.headers}
        
        reply_to_topic = data.get("body", {}).get("reply_to_topic")
        if not reply_to_topic:
            continue # Can't reply

        # Do the work
        url = data["body"]["data"]["url"]
        bucket = data["body"]["data"]["s3_bucket"]
        prefix = data["body"]["data"]["s3_path_prefix"]
        
        result_data = capture_site(url, bucket, prefix)
        
        # Build the response message
        response_message = {
            "headers": {
                "correlation_id": headers_map.get("correlation_id"),
                "in_response_to_request_id": headers_map.get("request_id"),
                "message_type": "response"
            },
            "body": {
                "success": True,
                "data": result_data
            }
        }
        
        # Send the response back
        producer.send(reply_to_topic, json.dumps(response_message).encode('utf-8'))
        
    except Exception as e:
        print(f"Failed to process message: {e}")
        # Send an error response back (omitted for brevity)
```

-----

### 2\. 🌩️ How to Build Your External LLaVA Adapter

This is even simpler because the adapter *doesn't run Python*. It's a **Go-based proxy** that listens on Kafka and makes an HTTP request. We follow the *exact same pattern*.

#### Step 1: The Go Action (`vlm_actions.go`)

This will be almost identical to `CaptureSiteAction`, but it will send a different payload (prompt, image) to a different topic.

* **Action:** `LabelLayoutAction(...)`
* **Sends to Kafka Topic:** `system.adapter.vlm.requests`
* **Payload:** `{"action": "llava_label", "data": {"image_b64": "...", "prompt": "..."}, "reply_to_topic": "..."}`
* **Returns:** `AwaitResponse: true`

#### Step 2: The `vlm_adapter` (in Go)

This is a new K8s deployment, but it's a **Go binary** (like `dynamic_adapter.go`).

* It consumes messages from `system.adapter.vlm.requests`.
* When it gets a message, it:
    1.  Loads the `THUNDERCOMPUTE_API_KEY` from its K8s `Secret`.
    2.  Parses the message body to get the `image_b64` and `prompt`.
    3.  Builds the JSON request for ThunderCompute's HTTP API.
    4.  Makes the `http.Post("https://api.thundercompute.com/...", ...)` call.
    5.  Gets the JSON response from the external API.
    6.  Finds the `reply_to_topic` from the original Kafka message.
    7.  Produces the result back to that topic.

This pattern is perfect. It's modular, secure (the API key is isolated), and reuses your existing Kafka architecture.

-----

### 3\. 📝 The Complete List of New Actions & Adapters

Here is the full list of new `action` types we'll need to add to your Go `agent-chassis` and the corresponding adapter services we'll need to build.

| `action` Name (in Workflow JSON) | Go Action Function (in `agent-chassis`) | Adapter Service (New K8s Deployment) | Tech Stack |
| :--- | :--- | :--- | :--- |
| `http_get_text` | `GetTextAction` | **None** (Built-in Go) | Go `net/http` |
| `run_playwright_capture` | `CaptureSiteAction` | **`playwright-adapter`** | Python + Playwright + Kafka |
| `run_playwright_get_style` | `GetStyleAction` | **(Re-uses `playwright-adapter`)** | Python + Playwright + Kafka |
| `run_cv_layout_cut` | `LayoutCutAction` | **`cv-adapter`** | Python + OpenCV + Kafka |
| `run_llava_label` | `LabelLayoutAction` | **`vlm-adapter`** (Go Proxy) | Go + Kafka + HTTP |
| `run_vlm_generate_code`| `GenerateCodeAction` | **(Re-uses `vlm-adapter`)** | Go + Kafka + HTTP |
| `run_llm_refactor_code` | `RefactorCodeAction` | **`codellama-adapter`** (Go Proxy) | Go + Kafka + HTTP |
| `db_insert_component` | `InsertComponentAction` | **None** (Built-in Go) | Go `pgx` (to Vector DB) |
| `db_query_component` | `QueryComponentAction` | **None** (Built-in Go) | Go `pgx` (to Vector DB) |
| `request_human_approval`| `RequestApprovalAction` | **None** (Built-in Go) | Go `kafka.Producer` |

-----

### 4\. 💡 A Note on Your "Idle Adapters" Concern

You mentioned: *"I'm not sure quite how we stop having lots of these adapters always running..."*

This is a valid concern. The solution is **KEDA (Kubernetes Event-driven Autoscaling)**.

* KEDA is an open-source tool that plugs into K8s.
* You can tell KEDA: "Watch the Kafka topic `system.adapter.playwright.requests`."
* If the number of unread messages (lag) is 0, KEDA will automatically scale your `playwright-adapter` deployment down to **0 pods**.
* The moment a message (a new job) arrives on the topic, KEDA will see the lag go to 1 and *instantly* scale the deployment up to 1 (or more) pods to handle the work.

This gives you a serverless, event-driven architecture that only uses resources (and costs money) when there is actual work to be done. We can absolutely sort this out later, but your architecture is tailor-made for it.


