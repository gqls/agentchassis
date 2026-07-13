# Where it all fits — a consolidation map (2026-06-04)

You asked to step back and see the whole shape. Here it is: five layers, what
each is, what's built vs not, and how they connect. The thing to hold onto is
that these are **one stack**, not separate projects — each layer is a customer
of the one below it.

---

## The one-paragraph version

The chassis builds websites. The idea engine decides *what's worth building*.
Today they're separate: the chassis builds static content sites, and the idea
engine runs on its own (for us internally, and soon as idea.uk for strangers).
The near future joins them — the idea engine becomes a *planning input* to the
chassis, so when you hand the platform a domain name it can decide what tools,
content, and features that site should have, then build them. The far future
closes the last gap: some of those tools are *backend services* (not static
files), and the platform can't yet provision the servers they need — that's the
VM-deployment piece you don't do yet. The Thunder adapter is the seed of it.

---

## The five layers

```
┌─────────────────────────────────────────────────────────────────────┐
│ LAYER 5  Automated backend deployment onto VMs        [FUTURE — seed] │
│          provision a server, deploy a service, wire it up             │
│          (Thunder adapter is the beginning of this)                   │
├─────────────────────────────────────────────────────────────────────┤
│ LAYER 4  Tool-rich site building for ANY domain          [FUTURE]     │
│          give a domain → find best-in-vertical examples → plan and    │
│          build a site with targeted design, content, tools, news,     │
│          interactive graphics  (the ORIGINAL problem statement)       │
├─────────────────────────────────────────────────────────────────────┤
│ LAYER 3  The suggested tools — built for real          [IN PROGRESS]  │
│          turn a recommendation into a working product                 │
│          first one: SFI26 Diff Alerts (chassis-native, scheduled)     │
├─────────────────────────────────────────────────────────────────────┤
│ LAYER 2  idea.uk as a product                          [IN PROGRESS]  │
│          £29 report + free audience-check taster, real-door delivery  │
├─────────────────────────────────────────────────────────────────────┤
│ LAYER 1  The idea-generation engine                    [BUILT]        │
│          audience → generate → cut → verify → score → rank (+ Risk)   │
│          runs internally (CLI) and behind idea.uk                     │
├─────────────────────────────────────────────────────────────────────┤
│ LAYER 0  The chassis                                   [EXISTS]       │
│          Go / Kafka / Postgres; builds static multi-page sites:       │
│          classify → plan → build components → deploy (git→Actions→B2)  │
└─────────────────────────────────────────────────────────────────────┘
```

Read it bottom-up: each layer uses the one below.

---

## Layer 0 — the chassis (exists, in production)

What you already have. A Go/Kafka/Postgres agent platform that builds static
multi-page websites: classify a domain, plan the site (`site_specs`), build
components, deploy via git → GitHub Actions → Backblaze B2. Every agent is an
orchestrator; workflows stay thin; logic lives in Go actions. It already has the
actions the higher layers reuse: `execute_llm_prompt`, `web_search`,
`spawn_agent`/`spawn_group`, the HITL set (`request_human_input`,
`await_approval`, …), `read_site_spec`/`write_site_spec`, the tool pipeline
(`create_tool_component`, `deploy_tool_to_site`), `send_notification`, state and
memory.

This is the substrate. Nothing above replaces it; everything above runs on it.

---

## Layer 1 — the idea-generation engine (built, standalone)

The method: challenge the audience → generate across four lenses → cut against
the free alternative (different model) → verify survivors with web search →
score on five fitness factors → rank, **plus the Risk column** that gates and
flags operator exposure.

Two ways it runs:
- **Internal** — `go run . internal <domain> <audience> <assets>`. For our own
  domains. No billing.
- **External** — the idea.uk service. Strangers pay; the same engine runs.

Status: the engine works against live APIs (just upgraded to Opus 4.8 + adaptive
thinking + web_search v2 + caching; three API bugs found and fixed during
validation). The service plumbing is tested (24/24), billing is wired. **Not
deployed yet.**

Important: this is currently a **standalone Go service**, not chassis-native.
The chassis-native version (the method expressed as an agent + workflow reusing
Layer 0's actions) is real but deferred — it's Phase D in the runbook, and it
needs a schema pass first. The standalone form was the right MVP and is
sale-ready.

---

## Layer 2 — idea.uk as a product (in progress)

idea.uk is Layer 1's external face: a £29 verified-idea report, with a free
30-second audience-check taster as the hook, and (planned) a real-door streaming
page so the buyer watches the work happen. This is the **first thing to go
live** — it validates that strangers will pay, and it generates a stream of
candidate tools (Layer 3 inputs) as a side effect.

Status: page rewritten in plain English with the taster widget wired to a live
`/audience-check` endpoint (built this week). Runbook Phase A is the pre-launch
checklist (latest-LLM upgrade ✅, taster endpoint ✅, streaming page, refund
endpoint, T&Cs solicitor review, PII quote, Stripe live mode). Phase B is deploy
+ first 10 operator-reviewed orders.

Hosting reality (worth keeping straight): the **page** is serverless (static on
B2, like every other site). The **service** is *not* serverless and can't be —
it's a minutes-long multi-LLM job with a billing webhook, so it's a small
always-on container. "Static front + small back end." This distinction is the
hinge that Layer 5 eventually automates.

---

## Layer 3 — the suggested tools, built for real (in progress)

idea.uk *recommends* tools. Turning a recommendation into the best-in-vertical
product is a separate, harder build — and it's the more valuable half, because
most consultants stop at "here's the opportunity" and never build the thing.
This is the "build bridge" that gives the whole enterprise its moat.

First vertical tool: **SFI26 Diff Alerts** (UK farm advisors — the audience the
method keeps surfacing). It replaced the SFI single-farm assessment, which the
Risk column flagged as too high-exposure for a first build (a wrong number could
cost a farmer £5–50k; Diff Alerts only reports what changed, advisor decides).

Crucially, **the vertical tools are chassis-native, not standalone** — opposite
to idea.uk. They're recurring, hold per-user state (client portfolios), run on
schedules, and reuse Layer 0's actions, all of which the chassis is built for.
So:

- **idea.uk** = standalone service (serves strangers across all verticals, one
  generic engine, one-off paid).
- **vertical tools** = chassis-native agents + workflows (recurring,
  per-user-state, vertical-specific, reuse chassis actions).

Both share the same method DNA (multi-step, multi-model, verified, cited); their
plumbing differs because their shapes differ.

Status: scoped in runbook Phase C (method spec, versioned corpus, diff engine,
operational shape, first 8 weekly digests). Not built yet.

---

## Layer 4 — tool-rich site building for any domain (future)

This is the **original problem statement**: hand the platform a domain name, and
it finds the best examples in that vertical, interrogates them, and builds a
sophisticated site with targeted design, content, pertinent tools, blog/news
from the wider world, and interactive graphs — not by copying, but by
understanding *why* the good examples work.

How the pieces connect here:
- **The idea engine (Layer 1) becomes a planning input to the chassis (Layer
  0).** When a new domain enters the build pipeline, the engine answers "what
  tools and content would make this site genuinely useful for its audience?"
  and writes those into the `site_specs` (`site_plan` aspect, items that start
  `blocked` and flip to `planned` once the capability exists).
- **The chassis builds what the engine planned.** Static content and
  client-side tool widgets it can already build and deploy today.
- "**Apply the tool to other domains, which then creates the tools**" is this
  layer: the same tool-building capability, pointed at an arbitrary domain via
  its site spec, producing the right tools for that site automatically.

The honest boundary: a site that only needs **static content + client-side
widgets** (calculators, that kind of thing) is fully within reach of today's
chassis once Layer 1 feeds it the plan. A site that needs a **backend service**
(anything like the idea engine or Diff Alerts) hits the Layer 5 gap.

Status: not built. This is where Layers 1–3 are heading once each is proven.

---

## Layer 5 — automated backend deployment onto VMs (future; the gap you named)

Today the platform deploys **static files to B2**. It does **not** provision and
deploy **persistent backend services** — and the richest tools need exactly
that. This is the "deployment of backend functionality onto VMs that we don't
yet do."

The shape of the gap, concretely:
- **What works today:** static site → git → GitHub Actions → B2. Serverless,
  zero servers to manage.
- **What doesn't exist yet:** "this site needs a backend service" → provision a
  VM (or a container host) → deploy the service → wire DNS/TLS/secrets →
  health-check → connect the static front to it. Today that's the manual work
  we'd do by hand to put idea.uk or Diff Alerts live (runbook Phase B's B1).

The **Thunder adapter is the seed of this layer** — it already provisions
compute (the `thunder-adapter`, the provisioning workflow, the
`prepare_artefact_url`/storage plumbing in the project). Generalising it from
"provision a GPU/VM for X" to "provision and deploy a backend service that a
built site depends on" is the path from where Thunder is now to Layer 5.

Two kinds of "tool" and two kinds of "deploy" make the gap precise:
- **Tools:** (a) static client-side widgets — deploy as files, no backend;
  (b) backend services — need persistent compute.
- **Deploy:** (a) static → B2 (done); (b) backend service → VM (the gap).

Status: not built; Thunder adapter is the groundwork.

---

## So where are you right now?

- **Layer 0** is live and carrying real sites.
- **Layer 1** is built and just validated against the newest models (bugs found
  and fixed). One clean end-to-end run away from "proven."
- **Layer 2 (idea.uk)** is the immediate focus — Phase A nearly done, then
  deploy. This is the first revenue and the first real-world signal.
- **Layer 3 (vertical tools)** is scoped and waiting; SFI26 Diff Alerts is the
  first, chassis-native, after idea.uk is running.
- **Layer 4** is the original vision — it switches on once the engine (1) is
  trusted enough to *plan* sites, not just advise humans.
- **Layer 5** is the deepest gap and the biggest future build; the Thunder
  adapter is where it starts.

The natural order is the order you're already in: **prove Layer 1, ship Layer 2,
build Layer 3 once, then generalise into Layer 4, and grow Layer 5 from Thunder
when tool-rich sites actually need backends deployed at scale.** Each step earns
the right to the next; none of them is wasted if a later one changes shape.

---

## How this maps to the docs you already have

- Method (with Risk): `idea_uk_method_v0.md`
- Engine + service architecture, hosting, Stripe: `idea_uk_architecture_and_deployment.md`
- Build sequence, Phases A–D: `DEVELOPMENT_RUNBOOK.md`  (Layer 2 = Phase A/B, Layer 3 = Phase C, chassis-native Layer 1 = Phase D)
- Liability framework + the Risk-column first-line filter: `LIABILITY_AND_TERMS.md`
- Cost / pricing / self-host / white-label analysis: `idea_uk_open_discussion.md`
- Running journal (cross-session memory): `running_notes.md`
- Debugging-guide lessons incl. the API-shape disciplines: `016_debugging_guide_v2_32.md`

Layers 4 and 5 don't have their own design docs yet — they're vision, not plan.
When you're ready to make either concrete, that's a focused planning pass of its
own (and Layer 5 would start by reading the Thunder adapter design docs already
in the project).
