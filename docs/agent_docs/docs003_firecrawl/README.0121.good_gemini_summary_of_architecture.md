This is a fantastic, robust framework. Thank you for sharing the code. I've done a deep analysis, and the good news is that your architecture is not only compatible with our 11-agent plan, it's *ideally suited* for it.

Your "Agent Group" concept is already implemented in your code as a **versioned, workflow-driven agent**. Our "Agent Groups" are simply `Agent` definitions in your database, and their "pipelines" are just `AgentWorkflow` definitions.

This is a powerful, flexible, and auditable system. The "step execution" (`executeStep` in your `agent.go`) is the "TODO" plug-in point where all our models will live.

Here is the detailed, point-by-point analysis and comparison.

-----

### Detailed Analysis of Your Existing Framework

Based on your `production_code-all_context.txt` file, I've identified the following core concepts:

* **Core Unit: The `Agent` (in `models.go`)**

    * This is the "definition" of an agent. It's defined by its `Name` and `Version`.
    * It has a defined `InputSchema` and `OutputSchema`, which is excellent for creating data contracts between agents.

* **The "Brain": The `AgentWorkflow` (in `models.go`)**

    * This directly confirms your point: **"each versioned agent has its own workflow saved in the database."**
    * An agent's entire logic is just a `Definition` (a JSON document) stored in the `agent_workflows` table.
    * This is incredibly flexible. To update an agent's logic, you just update this JSON.

* **The Orchestrator: The `AgentService` (in `agent.go`)**

    * This is the Go service that runs the agents. Its `Run` function is the main entry point.
    * Its `executeWorkflow` function is the "meta-orchestrator." It fetches the agent's JSON workflow and iterates through the defined "steps."
    * This confirms your second point: **"every agent is an orchestrator."** An agent's workflow *is* an orchestration plan. This also means a workflow step can, in theory, trigger *another* `AgentRun` (a workflow-of-workflows).

* **The "Memory": The `AgentStorage` (in `models.go`)**

    * This confirms your first point: **"each agent can have its own version and storage."**
    * This is a generic, key-value JSON store in your Postgres DB.
    * It's perfectly namespaced by `AgentID`, `Version`, `AgentRunID`, and even `Client` and `Job`. This is ideal for an agent's "scratchpad" (e.g., "for this run, I found these 50 URLs") without polluting a central database.

-----

### Point-by-Point Comparison (Our Plan vs. Your Framework)

Here is how our 11-agent plan maps directly onto your existing Go structures.

#### 1\. Agent Definition and Responsibility

* **Our Plan:** We defined 11 conceptual agents (Prospector, Profiler, Capture Bot, Layout Labeler, Component Generator, Style Extractor, Behavior Extractor, Librarian, Publisher, Architect, Strategist).
* **Your Framework:** These 11 agents are not 11 different *binaries*. They are 11 new **rows** in your `agents` table.
* **Merge Strategy:**
    * We will run a script to `INSERT` our 11 agents into your `agents` table.
    * **Example Row:**
      ```sql
      INSERT INTO agents (name, description, version, input_schema, output_schema)
      VALUES (
        'Profiler',
        'Agent 1: Classifies a site''s high-level goal and vertical.',
        '1.0',
        '{"type": "object", "properties": {"url": {"type": "string"}}}',
        '{"type": "object", "properties": {"site_goal": {"type": "string"}, "vertical": {"type": "string"}}}'
      );
      ```

#### 2\. Orchestration and Workflows

* **Our Plan:** We have "pipelines" and a master "Strategist" (Agent 10) that orchestrates the others.
* **Your Framework:** This is *perfectly* handled by your `AgentWorkflow` model. An agent's complexity is defined by its workflow.
* **Merge Strategy:**
    * **Simple Agents (like `Profiler`):** The `AgentWorkflow` for `Profiler:1.0` will be a simple, one-step JSON. This step is the "TODO" in your `executeStep` function.
      ```json
      {
        "steps": [
          {
            "name": "run_profiler_model",
            "task_type": "python_service_call",
            "endpoint": "http://python-profiler-svc:8000/classify",
            "input": "{ \"url\": $.agent_run.input.url }"
          }
        ]
      }
      ```
    * **Orchestrator Agents (like `Strategist`):** The `AgentWorkflow` for `Strategist:1.0` will be a complex, multi-step JSON that calls *other* agents, just as we discussed.
      ```json
      {
        "steps": [
          {
            "name": "find_targets",
            "task_type": "run_agent",
            "agent_name": "Prospector",
            "agent_version": "1.0",
            "input": "{ \"query\": $.agent_run.input.query }"
          },
          {
            "name": "analyze_targets",
            "task_type": "loop",
            "over": "$.steps.find_targets.output.urls",
            "run": [
              {
                "task_type": "run_agent",
                "agent_name": "Profiler",
                "agent_version": "1.0",
                "input": "{ \"url\": $.loop.item }"
              }
            ]
          }
        ]
      }
      ```

#### 3\. Storage, State, and "The Librarian"

* **Our Plan:** We have a central "Librarian" (Agent 7) that writes to a central Postgres Vector DB.
* **Your Framework:** You have `AgentStorage` for *per-run, per-agent* "scratchpad" memory.
* **This is the most important distinction.**
* **Comparison & Merge Strategy:**
    * Your `AgentStorage` is **kept as-is**. It's perfect for transient data (e.g., "Agent 2: Capture Bot" uses `saveToStorage` to save the S3 path for its screenshot, so that "Agent 3: Layout Labeler" can find it in the *same run*).
    * Our "Librarian" (Agent 7) is **still required** and is implemented as a *new, standalone Agent* (`Name: "Librarian"`).
    * **The Librarian's Job:** Its `AgentWorkflow` is triggered after an ingestion run. Its job is to:
        1.  Use `loadFromStorage` to get all the transient data from the previous agents' runs.
        2.  Connect to the *main, shared* **Postgres Vector DB** (which is a separate database, *not* the `AgentStorage` K-V store).
        3.  Format and `INSERT` the final, permanent component row (the one with the vector embedding, clean CSS, etc.).
    * This gives us the best of both worlds: your auditable, run-specific storage *and* our permanent, shared library.

-----

### Implementation Plan: Starting Points

This analysis shows your framework is the perfect "chassis" for our plan. The immediate task is to implement the "TODO" in your `executeStep` function.

1.  **Modify `executeStep`:** We need to teach your Go `AgentService` how to execute the new `task_type`s we defined.

    * `task_type: "run_agent"`: This would make a new `agentSvc.Run()` call to nest workflows.
    * `task_type: "python_service_call"`: This is the most important. It will be a simple Go `http.Post` call. This is how your Go orchestrator will "talk" to the Python ML models (LLaVA, CodeLlama, etc.) which we will deploy as separate microservices in your K8s cluster.

2.  **MVP Agent Deployment (Our "Reduced" Plan):**

    * We will start by building **Agent 3 (Layout & Labeler)**.
    * We will deploy **LLaVA** as a containerized Python service (e.g., using FastAPI or Ollama) in your K8s cluster.
    * We will register `Agent: "LayoutLabeler", Version: "1.0"` in your DB.
    * We will write its `AgentWorkflow` JSON (a one-step workflow: `task_type: "python_service_call", "endpoint": "http://llava-svc:8000/label"`).
    * We will use your **existing Claude/Gemini APIs** for all other models (Profiler, Generator) by creating a `task_type: "gemini_api_call"`.

This gives us a clear path: we use your Go framework for *orchestration and state* and Python services for *ML model execution*, with your Go `AgentService` as the glue.