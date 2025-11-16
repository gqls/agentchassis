Here is a complete, independent summary of the project objectives and operational plan.

### **Core Mission**

The primary objective is to build and manage a **large-scale portfolio of high-performing websites**. The goal is not just to build the sites, but to create a **"Pragmatic Evolution Engine"** that systematically improves their effectiveness (e.g., traffic, conversions, revenue) over time.

The ultimate asset this system builds is an **"Internal Library of Effectiveness"**—a proprietary, first-party database of which designs, behavioral models, and components are *statistically proven* to work.

---

### **Phase 1: MVP Site Generation (The "Build" Loop)**

This is the "pragmatic-first" approach to building a functional site from Day 1, solving the "cold start" and "backlog" problems.

* **First-Principles Strategy:** Site generation begins with a *first-principles* behavioral model (e.g., AIDA, PAS), not scraped data.
* **The `Build Plan`:** A `Chief Strategist` agent generates a simple, generic JSON `Build Plan` that outlines the *functional* sections of a page (e.g., `{"sections": ["problem_statement", "agitation", "solution_provider"]}`).
* **Intelligent Fallback Logic:** A `Site Architect` agent builds the site using a 3-tiered logic to find components in the `in-house` library:
    * **P1 (Perfect Match):** Find an `in-house` component matching the exact function (e.g., `function: "problem_statement"`).
    * **P2 (Good Match):** Find any `in-house` component with a similar purpose.
    * **P3 (Base Fallback):** If no match exists, use a `generic-text-block`. The site *always* gets built.
* **Decoupled Content:** A `data-function` attribute in the HTML acts as a **"shared contract"**. This allows the `Site Architect` to build the *empty container* and the `Content Pipeline` to *independently* fill it with the correct, purpose-driven copy.

---

### **Phase 2: Evidence Gathering (The "Learn" Loop)**

This is the "evidence-gathering" loop, which functions as an **"Idea Generator,"** not a "Fact-Finder."

* **Acknowledge "Messy Data":** All scraped data is treated as "messy, high-correlation" *ideas*, not as "truth." The system assumes it will find "cargo cults" (correlation, not causation).
* **Learn from "Winners":** A `Prospector` agent will use external metrics (e.g., Ahrefs API) to find high-traffic, high-authority sites to analyze. This filters the "noise" from low-performing sites.
* **Behavioral "Scorecard":** A `Pattern Deconstructor` agent will analyze a target site from the perspective of *multiple* behavioral models (AIDA, PAS, Cialdini) to create a "scorecard" of evidence.
* **Prioritized Test Backlog:** A `Librarian` agent stores this "messy evidence" and generates a "Hypothesis Priority List" (e.g., "P1: 'Comparison tables' are highly correlated with success. This is a high-priority test.").

---

### **Phase 3: Portfolio Evolution (The "Test" Loop)**

This is the "scientific method" loop for turning "correlation" into *causation*.

* **Controlled Evolutionary Cohorts:** To test hypotheses, the system runs large-scale, **single-variable A/B tests** across defined "explore" cohorts of the site portfolio (e.g., 50% of sites get `lists` vs. 50% get `comparison-tables`).
* **Statistical Validity:** This controlled approach avoids the "Attribution Black Hole" of a chaotic 90%-churn model. It provides clean, measurable data on a single change.
* **SEO & Brand Stability:** Content and layout are evolved on different tracks.
    * **Content:** Added/pruned continuously (e.g., "add 3 new pages/day") to build SEO authority.
    * **Layout:** Changed only in controlled, cohort-based tests to maintain user trust and brand stability.

---

### **Phase 4: Optimization & Operations**

This phase details how the portfolio is managed, scaled, and optimized long-term.

* **Site-Specific Optimization (No Monoculture):** This is the "exploit" strategy. When a test (e.g., `comparison-tables`) wins *on average*, the change is **only** applied to the *individual sites* where it *actually improved performance*. Sites where it failed will keep their original component. This avoids a "monoculture" and optimizes each site for its own niche.
* **Manifest-Driven Operations:** Every site is managed via a `manifest.json` file, tracking its unique component "genes" and build history.
* **Git-Based Deployment:** A `git_adapter` commits all new builds and edits to a **dedicated Git repository for each domain**.
* **External Edit Handling:** A `git_hook_adapter` (webhook) detects external *human* commits to a repo, flags that site's manifest as `desynchronized`, and queues it for a human-in-the-loop (HITL) review.
* **Client Handoff Strategy:** The system will include "Exporter" agents (e.g., a `WordPress Formatter`) to transpile site content (as an `.xml` file) and styles (as an `.sql` file) for easy handoff to clients on common platforms.

===

Here's a breakdown of the four MVP agents we've designed, their responsibilities, their code structure, and the messages they pass.

-----

## 1\. Agent: `chief-strategist`

* **Responsibility:** The "Thinker." Its one job is to take a high-level goal (like `domain`, `objective`, and `model`) and use an LLM to create a simple, "first-principles" JSON `Build Plan` that dictates the functional sections of the site.
* **Code Structure:** This agent is simple and requires **no new custom Go actions**. Its logic is entirely defined by its `agent_definitions` workflow, which just calls the `execute_llm_prompt` action you already have.
    * **File:** `agent_definitions.sql` (This is just an `INSERT` statement)
    * **SQL (as defined previously):**
      ```sql
      INSERT INTO agent_definitions (id, type, display_name, description, category, default_config, ...)
      VALUES (
        gen_random_uuid(),
        'chief-strategist',
        'Chief Strategist Agent',
        'Creates a "first-principles" Build Plan (e.g., AIDA, PAS) from a simple objective.',
        'generation',
        '{
          "workflow": {
            "start_step": "generate_build_plan",
            "steps": {
              "generate_build_plan": {
                "action": "execute_llm_prompt",
                "config": {
                  "prompt_template": "You are a Chief Marketing Strategist... Create a simple JSON ''Build Plan'' based on the ''{{.model}}'' model... Example: {\"sections\": [\"problem_statement\", \"agitation\", \"solution_provider\"]}",
                  "input_fields": ["input_data.domain", "input_data.objective", "input_data.model"],
                  "ai_service": { ... },
                  "output_field": "build_plan_json" 
                },
                "next_step": "complete"
              },
              "complete": {"action": "complete_workflow"}
            }
          }, ...
        }'::jsonb, ...
      );
      ```
* **Messaging (Kafka Payloads):**
    * **Input (from `mvp-site-builder`):** The `call_agent` message contains the initial user request.
      ```json
      {
        "action": "call_agent",
        "input_data": {
          "domain": "boxing-tickets.com",
          "objective": "affiliate-sales",
          "model": "PAS"
        }
      }
      ```
    * **Output (to `mvp-site-builder`):** The `complete_workflow` message contains the new `Build Plan`.
      ```json
      {
        "status": "success",
        "results": {
          "build_plan_json": "{\"sections\": [\"problem_statement\", \"agitation\", \"solution_provider\"]}"
        }
      }
      ```

-----

## 2\. Agent: `site-component-architect`

* **Responsibility:** The "Builder." Its job is to take the `Build Plan` from the strategist, connect to your Postgres DB, and use the **`AssembleFromLibraryAction`** (with its "Intelligent Fallback" logic) to build the empty, semantically-tagged HTML template.
* **Code Structure:** This agent requires **one new custom Go action**.
    * **File:** `internal/backend/agent-chassis/platform/orchestration/actions/site_architect_actions.go` (A new file).
    * **Action (as defined previously):** `AssembleFromLibraryAction(ctx context.Context, params ActionParams) (interface{}, error)`
    * **Registration:** This action must be registered in your `agent-chassis/main.go`'s `actionRegistry`.
    * **Agent Definition:** The `agent_definitions` SQL we wrote, which defines a simple workflow that just calls this one action.
* **Messaging (Kafka Payloads):**
    * **Input (from `mvp-site-builder`):** Contains the output from the previous agent.
      ```json
      {
        "action": "call_agent",
        "input_data": { ... }, // Original input
        "collected_data": {
          "build_plan_data": {
            "build_plan_json": "{\"sections\": [\"problem_statement\", \"agitation\", \"solution_provider\"]}"
          }
        }
      }
      ```
    * **Output (to `mvp-site-builder`):** The empty template and the "shopping list" for the content creator.
      ```json
      {
        "status": "success",
        "results": {
          "stitched_html_template": "<div data-function=\"problem_statement\">...</div>\n<div data-function=\"agitation\">...</div>\n...",
          "content_requirements": {
            "component_abc123": {"headline": "string"},
            "component_def456": {"subheading": "string", "point1": "string"},
            ...
          }
        }
      }
      ```

-----

## 3\. Agent: `content-creator`

* **Responsibility:** The "Writer." Its job is to take the `stitched_html_template` and `content_requirements` from the architect. It must then fill the empty, tagged slots with copy that matches the *function* of each component.
* **Code Structure:** You already have this agent. We will need to **enhance its internal workflow**.
    * **File:** `internal/agents/contentcreator/agent.go` (and new `actions`)
    * **New Action:** You'll need an "enhanced HTML action" as you said, let's call it `populate_html_action`. This action will:
        1.  Parse the `content_requirements` to see what's needed (e.g., a `headline` for `component_abc123`).
        2.  Call an LLM (using `execute_llm_prompt`) *for each component* to generate the specific copy (e.g., "Write a 'problem\_statement' headline...").
        3.  Use a Go template or string replacement to inject this new copy into the `stitched_html_template`.
* **Messaging (Kafka Payloads):**
    * **Input (from `mvp-site-builder`):** Contains the architect's output.
      ```json
      {
        "action": "call_agent",
        "input_data": { ... }, // Original input
        "collected_data": {
          "template_data": {
            "stitched_html_template": "<div ...></div>...",
            "content_requirements": { ... }
          }
        }
      }
      ```
    * **Output (to `mvp-site-builder`):** The final, populated HTML.
      ```json
      {
        "status": "success",
        "results": {
          "final_html": "<div data-function=\"problem_statement\"><h1>Sold out again?</h1></div>..."
        }
      }
      ```

-----

## 4\. Agent: `deployer-agent`

* **Responsibility:** The "Publisher." Its job is to take the `final_html` and the `domain`, and call the standalone **`git-adapter`** to commit this file to a new, domain-specific GitHub repo.
* **Code Structure:** This agent requires **one new custom Go action** to call the adapter.
    * **File:** `internal/backend/agent-chassis/platform/orchestration/actions/deployer_actions.go` (A new file).
    * **Action (as defined previously):** `GitCommitAction(ctx context.Context, params ActionParams) (interface{}, error)`. This action constructs the detailed `adapterRequest` and sends it to the Kafka topic.
    * **Registration:** This action must also be registered in your `agent-chassis/main.go`'s `actionRegistry`.
    * **Standalone Adapter:** We also need the separate **`git-adapter` microservice** (code provided previously) running and listening on `system.adapter.git.requests`.
* **Messaging (Kafka Payloads):**
    * **Input (from `mvp-site-builder`):** Contains the final HTML.
      ```json
      {
        "action": "call_agent",
        "input_data": { "domain": "boxing-tickets.com", ... }, // Original input
        "collected_data": {
          "final_site_data": {
            "final_html": "<html>...</html>"
          }
        }
      }
      ```
    * **Output (to `mvp-site-builder`):** The response from the `git-adapter`.
      ```json
      {
        "status": "success",
        "results": {
          "success": true,
          "repo_url": "https://github.com/your-org/boxing-tickets.com"
        }
      }
      ```
      

====


Phase,Title,Core Objective,Key Agents & Components
"Phase 1: The ""Build"" Loop (MVP)","""Pragmatic-First"" Execution","To build a strategically-sound, ""v1"" website immediately, solving the ""cold start"" problem.","• chief-strategist: Applies ""first-principles"" theory (e.g., PAS, AIDA) to create a simple Build Plan (e.g., [""problem"", ""agitate""]).• site-component-architect: Uses ""Intelligent Fallback Logic"" (P1/P3) to assemble empty HTML from our small, in-house component library.• content_components (DB): Our ""In-House Forge"" of ""Lego blocks,"" tagged by function (e.g., function: ""problem_statement"").• deployer-agent: Pushes the final site to a new, dedicated Git repo."
"Phase 2: The ""Learn"" Loop","The ""Idea Generator""","To analyze successful ""winner"" sites and generate a prioritized backlog of testable ideas (hypotheses).","• Prospector: Uses external APIs (Ahrefs, etc.) to find high-traffic ""winner"" sites to analyze.• Capture Bot: Gets the dom.html, screenshot.png, and the crucial layout_map.json (the ""Rosetta Stone"" that links code to pixels).• Pattern Deconstructor (VLM/LLM): The ""Inference Engine."" It interrogates a site's DOM and screenshots to create a ""scorecard"" of hypothesized functions (e.g., ""This component has a 9/10 confidence of being 'Cialdini: Social Proof'"").• Librarian: A statistical engine that stores this ""messy evidence"" and creates a ""Hypothesis Priority List"" (P1-P5) of test ideas."
"Phase 3: The ""Evolve"" Loop","The ""Testing Engine""",To turn the correlations from Phase 2 into causation (proven data) by running live experiments on our portfolio.,"• Chief Strategist (Evolve Mode): Initiates a ""Cohort Test"" (e.g., ""Test comparison-tables vs. lists on 50% of the 'loser' cohort"").• Site Architect (Evolve Mode): Executes the test, applying the ""Test A"" component to one group and ""Test B"" to another.• ""Controlled Cohorts"": We use large-scale, single-variable A/B tests, not ""90% chaos changes."" This solves the ""Attribution"" and ""SEO Destabilization"" problems.• ""Site-Specific Optimization"": The final, critical refinement. A winning test is only applied to the individual sites where it actually won, avoiding a ""monoculture."""
"Phase 4: ""Ops & Handoff""","The ""Backend & Scalability"" Loop","To manage the full lifecycle of the 1,000-site portfolio, including maintenance, external edits, and client handoffs.","• Site Manifest: A manifest.json in each Git repo tracks the site's unique ""winning genes"" (the components it uses).• git_hook_adapter: A webhook that detects human commits, flags the manifest as desynchronized, and queues a HITL review.• ""Handoff Agents"" (WordPress Formatter, etc.): Exporter agents that transpile our Build Plan and content into formats for external systems (e.g., WordPress xml/sql or a Headless CMS json)."
Cross-Phase Groups,"The ""Specialists""",The human- and AI-powered groups that provide specialized services to all other loops.,"• The ""In-House Forge"" (Our Developers): The team that builds the in-house ""Lego blocks"" (components, backend templates) and stores them in the content_components table. They use the ""suggestions"" from Phase 2 as benchmarks.• The ""Content Pipeline"" (Agent 11): A ""master agent"" that manages:   • ""Copywriters with Character"": Persona-driven LLM prompts to create content with a specific voice and style (e.g., ""enthusiastic boxing fan"").   • The ""Purifier Agent"": A sub-system for fact-checking, copyright, and plagiarism audits."

