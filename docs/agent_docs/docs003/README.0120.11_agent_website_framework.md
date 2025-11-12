Here's a detailed re-summary of our 11-agent plan, restructured around your "Agent Groups" concept to emphasize their independent responsibilities.

This architecture is a **service-oriented model**, where each group is a "capability" that can be called by a central orchestrator. This directly enables the flexible ordering you're looking for.

-----

### The 4-Group Architecture

Our framework is now split into four distinct, independently responsible "Agent Groups."

1.  **Strategy & Content Group (The "Brain"):** The orchestrator. It holds the business logic, decides *what* to do and *why*, and infuses final products with content.
2.  **Library & Storage Group (The "Hub"):** The central persistence layer. Its *only* responsibility is to safely store, index, and retrieve data.
3.  **Design Ingestion Group (The "Analyzers"):** A "read-only" service for the web. Its responsibility is to deconstruct a target website into abstract, structured data.
4.  **Generation Group (The "Builders"):** A "read-only" service for our library. Its responsibility is to build *new* assets by querying the Hub.

-----

### 🧠 Group 1: The "Strategy & Content" Group (The Orchestrator)

This is your master agent group. It's the "client" that consumes the services of all other groups.

* **Agent 10: The Strategist**

    * **Role:** The main "brain" of the entire operation.
    * **Responsibility:** To initiate and orchestrate complex, multi-step workflows. It's the *only* agent that understands the high-level "why" of a task.
    * **Flexible Workflows (as you suggested):**
        * **Flow A (Analysis-First):** "I have `boxing-tickets.com`. I want to analyze it." -\> The Strategist calls the `Design Ingestion Group` and tells it to run an analysis.
        * **Flow B (Build-First):** "I have a new domain, `premiumfightnight.com`. What should I build?" -\> The Strategist queries the `Library Group` to find patterns for "fight" sites, then calls the `Generation Group` to build a template.
        * **Flow C (Targeted Ingestion):** "My library is weak on 'e-commerce checkout' pages." -\> The Strategist calls the `Prospector Agent` with the task: "Find me 10 sites with `vertical: 'e-commerce'`."

* **Agent 11: The Content Infuser**

    * **Role:** Your existing agent group, integrated as a specialist service.
    * **Responsibility:** To populate a *semantically-labeled* HTML template (from the `Generation Group`) with new, original content (copy, images).
    * **Interface:** It looks for `data-semantic-purpose` and `data-funnel-stage` attributes in the HTML to know *which* of its sub-agents (e.g., "hero-copy-writer," "trust-builder-writer") to activate.

-----

### 🗄️ Group 2: The "Library & Storage" Group (The Hub)

This group is the "state" of your system. It's a persistent, high-availability service.

* **Agent 7: The Librarian**
    * **Role:** The single, "write-only" API gateway to your databases.
    * **Responsibility:** To receive synthesized data packets from the `Ingestion Group`, process them (e.g., run CLIP embedding), and safely store them. It ensures data integrity.
    * **Storage:**
        * **S3 (Backblaze):** Stores all binary assets (screenshots, JS modules, `rrweb` recordings).
        * **Postgres (Vector DB):** Stores all queryable metadata, code, and embeddings.
    * **Example Component Schema (a row in Postgres):**
      ```json
      {
        "component_id": "bx-001-hero",
        "source_site": "boxing-tickets.com",
        "site_goal": "e-commerce",
        "layout_purpose": "hero",
        "funnel_stage": "attention",
        "html_clean": "<section class='hero-001'>...</section>",
        "css_clean": ".hero-001 { background: var(--primary-color); ... }",
        "design_tokens": {"--primary-color": "#CC0000", ...},
        "behavior_module_s3_path": "/behaviors/scroll-fade-in.js",
        "screenshot_s3_path": "/screenshots/bx-001-hero.png",
        "clip_embedding_vector": [0.12, 0.45, ..., -0.23]
      }
      ```

-----

### 📥 Group 3: The "Design Ingestion" Group (The Analyzers)

This is a complete, standalone "service" that the `Strategist` can call. Its *only* responsibility is to deconstruct a target URL.

* **Agent 0: The Prospector**

    * **Role:** Finds new target URLs.
    * **Model:** Deterministic (`requests`, `BeautifulSoup`).

* **Agent 1: The Site Profiler**

    * **Role:** Classifies the site's high-level goal (the "wealth of useful data").
    * **Model:** Text Classification (e.g., **BERT**, or **Gemini/Claude** for the MVP).

* **Agent 2: The Capture "Good Bot"**

    * **Role:** Ethically captures all raw assets (screenshots, DOM, `rrweb`).
    * **Model:** Deterministic (**Playwright**).

* **Agent 3: The Layout & Labeling Agent**

    * **Role:** Creates the *semantic wireframe* from the screenshot.
    * **Model:** **Recursive XY-Cut** (deterministic algorithm) + **LLaVA** (multimodal LLM).

* **Agent 4: The Component Generator**

    * **Role:** Generates clean, semantic HTML/CSS *structure* with variables.
    * **Model:** **VLM (Screenshot-to-Code)** (e.g., `HuggingFaceM4/VLM_WebSight_finetuned`).

* **Agent 5: The Style Extractor**

    * **Role:** Extracts the *real design tokens* (colors, fonts).
    * **Model:** Deterministic (**Playwright** `getComputedStyle`).

* **Agent 6: The Behavior Extractor**

    * **Role:** Refactors messy JS into clean, reusable vanilla JS modules.
    * **Model:** **CodeLlama** (or **Gemini/Claude** for the MVP).

-----

### 🏗️ Group 4: The "Generation" Group (The Builders)

This group is "read-only" from the library. It's a flexible set of tools for building new things.

* **Agent 8: The Publisher**

    * **Role:** A family of agents that create *public-facing* assets from your library.
    * **Responsibility:**
        * **Publisher A (Design Site):** Builds your "Dribbble-like" site of mood boards, colorways, etc.
        * **Publisher B (Case Study Bot):** Generates blog posts like "A Semantic Analysis of 5 Top E-commerce Heros."

* **Agent 9: The Architect**

    * **Role:** The internal "template builder."
    * **Responsibility:** Queries the library to find and assemble a *new, empty, multi-page template* based on a prompt from the `Strategist`.
    * **Example Query:** `SELECT * FROM components WHERE layout_purpose = 'hero' ORDER BY (clip_embedding <=> [vector_for_'rustic_brewery']) LIMIT 1`
    * **Output:** A set of semantically-labeled HTML/CSS/JS files, ready for the `Content Infuser (Agent 11)`.