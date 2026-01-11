Here is the 11-agent plan, re-summarized to reflect your production architecture.

Our plan maps perfectly onto your existing system. Your "Agent Group" is a high-level orchestrator agent (like your `Multi-Section Website Builder`), and our 11 agents are simply new rows in your `agent_definitions` table.

The core Go `Run` and `executeWorkflow` logic **does not change**. We will implement our plan by:
1.  **Defining** new `agent_definitions` and their `workflow` JSONs.
2.  **Implementing** new Go `action` functions (like your `GenerateImageAction`) that act as Kafka producers.
3.  **Creating** new adapter services (like your `dynamic_adapter.go`) that consume from Kafka topics, perform the work, and send a response.

---

### 🧠 Group 1: The "Strategy & Content" Group (The Orchestrator)

This group is composed of high-level orchestrators. These agents are defined almost entirely by their `orchestration_workflow` JSON, which uses `spawn_agent` and `call_agent` to manage other agents.

* **Agent 10: The Strategist**
    * **Definition:** An `agent_definitions` row (e.g., `type: "strategist-build-new-site"`) that is a pure orchestrator.
    * **Workflow:** Its `orchestration_workflow` is a complex, multi-step plan. For example, it would first `call_agent` on the `site-profiler` (Agent 1), then `call_agent` on the `capture-bot` (Agent 2), passing the results to subsequent agents. This is identical to how your `Multi-Section Website Builder` calls the `hero_writer`, `features_writer`, etc..

* **Agent 11: The Content Infuser**
    * **Definition:** This is your existing agent group (e.g., `content-creator-hero`, `content-creator-features`).
    * **Workflow:** It's triggered by the `Strategist` (Agent 10) or `Architect` (Agent 9). Its workflow, as seen in your `content-creator-hero` example, can *also* be an orchestration (calling a `content-researcher` sub-agent). This nested orchestration is a key part of our plan.

---

### 🗄️ Group 2: The "Library & Storage" Group (The Hub)

This group is responsible for the *persistence* of our library. Its "actions" will be new, built-in Go functions that talk directly to your Postgres Vector DB, not external adapters.

* **Agent 7: The Librarian**
    * **Definition:** A worker agent (`type: "librarian"`).
    * **Workflow:** Its workflow is triggered at the end of an ingestion pipeline. It will:
        1.  Receive a payload of all extracted data (HTML, CSS, tags, S3 paths).
        2.  Call a new `run_clip_embed` action (which sends a Kafka message to a CLIP adapter) to get the vector.
        3.  Call a new `db_insert_component` action to write the final, aggregated JSON row to your Postgres Vector DB.
    * **New Go Actions Required:**
        * `InsertComponentAction`: A new Go function in your `actions` package (like `loadAgentDefinitionForImageAction`) that connects to your PGVector DB (`pgxpool.Pool`) and runs an `INSERT` command.
        * `QueryComponentAction`: A Go function that runs `SELECT` queries against the PGVector DB (used by the Generation Group).

---

### 📥 Group 3: The "Design Ingestion" Group (The Analyzers)

This group contains specialist agents. Most will be "worker" agents whose workflow consists of a single "action" that calls a Kafka adapter.

* **Agent 0: The Prospector**
    * **Definition:** A worker agent (`type: "prospector"`).
    * **Workflow:** Calls a new `http_get_links` action. This can be a simple, built-in Go function (like `loadAgentDefinitionForImageAction`) that uses `net/http` and a parser, with no Kafka adapter needed.

* **Agent 1: The Site Profiler**
    * **Definition:** A worker agent (`type: "site-profiler"`).
    * **Workflow:** A two-step workflow:
        1.  `"action": "http_get_text"` (a new, simple Go action).
        2.  `"action": "execute_llm_prompt"` (an *existing* action you already have) to classify the text.

* **Agent 2: The Capture Bot**
    * **Definition:** A worker agent (`type: "capture-bot"`).
    * **Workflow:** A single-step workflow that calls our new `run_playwright_capture` action. This Go action (`CaptureSiteAction`) will be a Kafka producer that sends a message to `system.adapter.playwright.requests`, precisely like your `GenerateImageAction` sends to `system.adapter.image-generator.requests`.

* **Agent 3: The Layout & Labeling Agent**
    * **Definition:** An orchestrator agent (`type: "layout-labeler"`).
    * **Workflow:**
        1.  Calls `run_cv_layout_cut` (a new Go action that sends a Kafka message to a Python/OpenCV adapter).
        2.  Receives the array of layout blocks.
        3.  Uses your system's `loop` logic (a Kafka "fan-out") to call the `run_llava_label` action for each block.
        4.  `run_llava_label` is a new Go action that sends a Kafka message to our `vlm-adapter` (the Go proxy for ThunderCompute).
        5.  Waits for all responses (Kafka "fan-in") before proceeding.

* **Agent 4: The Component Generator**
    * **Definition:** A worker agent (`type: "component-generator"`).
    * **Workflow:** Calls the `run_vlm_generate_code` action. This re-uses the `vlm-adapter` (Go proxy) but sends a different prompt for "screenshot-to-code."

* **Agent 5: The Style Extractor**
    * **Definition:** A worker agent (`type: "style-extractor"`).
    * **Workflow:** Calls the `run_playwright_get_style` action. This *re-uses* the `playwright-adapter` (Python consumer). The Kafka message body will specify `"action": "get_style"` so the adapter knows which function to run.

* **Agent 6: The Behavior Extractor**
    * **Definition:** A worker agent (`type: "behavior-extractor"`).
    * **Workflow:** Calls the `run_llm_refactor_code` action. This can use your existing `execute_llm_prompt` action or a new action that calls a self-hosted CodeLlama adapter via Kafka.

---

### 🏗️ Group 4: The "Generation" Group (The Builders)

This group consists of orchestrators that *read* from the library using the new `db_query_component` action.

* **Agent 8: The Publisher**
    * **Definition:** An orchestrator agent (`type: "design-site-publisher"`).
    * **Workflow:**
        1.  Calls `db_query_component` to get, for example, all `color_palettes`.
        2.  Calls `execute_llm_prompt` to format this data into a blog post.
        3.  (Future) Calls a `wordpress-post` adapter to auto-publish.

* **Agent 9: The Architect**
    * **Definition:** An orchestrator agent (`type: "template-architect"`).
    * **Workflow:**
        1.  Calls `db_query_component` to find the best `layout_map` for a given prompt (e.g., `"vertical": "e-commerce"`).
        2.  Uses the `loop` (Kafka fan-out) to call `db_query_component` for each section of the layout to find the best *components*.
        3.  Calls a final `aggregate` action (just like your `aggregate_webpage`) to stitch the HTML/CSS/JS together.
        4.  Returns the S3 path to the finished, un-populated template.