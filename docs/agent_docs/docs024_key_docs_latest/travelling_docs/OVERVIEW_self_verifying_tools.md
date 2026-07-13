# Self-Documenting, Self-Verifying Tools — an overview

*A plain-language tour of the mechanism, written to be talked about. Last updated 2026-07-12.*

---

## The one-sentence version

Every interactive tool our platform builds now carries its own **living
specification and change-history in the database**, and the platform can
**drive that tool in a real browser to check it still works** — writing down
the verdict, and filing its own fix ticket when it doesn't. No human in the
loop.

---

## The problem it solves

The platform builds websites autonomously, including interactive tools —
calculators, simulators, games. Two failure modes kept biting:

1. **Lost intent.** An agent builds a tool, makes deliberate design choices,
   then moves on. Weeks later another agent "fixes" the tool and quietly
   undoes those choices, because nothing recorded *why* they were made. The
   reasoning lived only in a conversation that's gone.

2. **"Deployed" ≠ "works".** A tool can pass every structural check — valid
   HTML, a `<script>` block, no truncation — and still be broken in ways only
   a running browser reveals: a slider that does nothing, a chart line pinned
   to the wrong axis, a console full of errors. We watched this happen *twice*
   on the same game: the original build introduced two behavioural bugs, and a
   later automated repair faithfully reproduced them — both times passing every
   check we had.

The fix is two mechanisms that work together: **travelling docs** (so intent
and history travel with the tool) and a **verification ladder** (so "works" is
something we test, not assume).

---

## Mechanism 1 — Travelling docs (PLAN + NOTES)

Every tool gets two documents, stored in Postgres, **written by the agents
themselves** as a byproduct of building and fixing:

- **A PLAN** — what the tool is for, how it's delivered, its deliberate
  decisions ("do not re-fix this"), and — crucially — its **acceptance
  criteria**: a machine-readable list of what *working* means for this specific
  tool. The PLAN is versioned: updates supersede the previous version, never
  overwrite it, so the full history of intent is preserved.

- **A NOTES stream** — an append-only log of every fix, diagnosis, and dead
  end, each entry tagged by category. When an agent is about to touch a tool,
  it **loads the PLAN and NOTES first**, so it builds on prior decisions
  instead of re-deriving lost context.

The point: the documentation *travels with the subject* and is *maintained by
the machine*. A human never has to write or update it, and it can never drift
away from the code, because the same agents that change the tool write the docs.

**Proven in production:** the tool-generator now writes a complete PLAN
(criteria included) at the moment it creates a tool — unaided. The fix agents
each append a NOTES entry after a successful repair. Both were demonstrated
end-to-end on live tools.

---

## Mechanism 2 — The verification ladder

Acceptance criteria in a PLAN are only useful if something checks them. We
check in tiers, cheap-to-expensive, each catching a different class of problem:

| Tier | What it checks | How | Cost |
|---|---|---|---|
| **0 — Generation** | Output integrity — is the generated code complete and untruncated? | Marker + balance checks at build time | Free |
| **1 — Structural** | Is it deployed, does it have a script/style, is it mobile-ready, self-contained? | SQL + string checks over stored HTML | Free |
| **2 — Static acceptance** | Do the tool's own criteria selectors/assets actually exist in the deployed page? | Fetch the page, check each criterion's **anchor** (its leftmost id/class) against the HTML | Cheap |
| **4 — Behavioural** | Does it actually *work* in a browser? | Drive the deployed page in **headless Chromium** and evaluate the criteria for real | Real (a browser) |

*(Tier 3 — an LLM code-review of the deployed tool — sits between 2 and 4 and
predates this work.)*

The two tiers built in this arc are worth understanding:

**Tier 2 — the anchor rule.** A criterion might say "assert `#tableWrap tr`
exists". The rows are built by JavaScript at runtime, so they're *not* in the
static HTML — but the container `#tableWrap` is. Tier 2 validates the
**anchor** (`#tableWrap`), never the whole path: a real-but-empty container
passes (Tier 4 will assert the rows for real), while an **invented** selector
that exists nowhere fails. This is what catches an agent that hallucinates a
selector — which happened, and was caught, on the first live sweep. *Static
checks confirm, never refute.*

**Tier 4 — the real browser.** A dedicated service (the "browser-runner
adapter") launches headless Chromium, loads the deployed page, waits for the
JavaScript to render, and evaluates the criteria against the *live DOM*: does
the page return HTTP 200, do the required elements actually exist after render,
is the console free of errors? This is the only tier that can assert
`#tableWrap tr` → *"20 rows present"* — the thing no static analysis can know.
It's also the only tier that can see the tools we adopted from other sites,
which exist only as deployed pages.

---

## How it all comes together — the autonomous loop

Put the two mechanisms together and the platform can maintain its own tools:

```
   build/recreate a tool
        │
        ▼
   write its PLAN (criteria included)  ── travelling docs
        │
        ▼
   deploy to the live site
        │
        ▼
   tool-acceptance-agent:
     load the PLAN's criteria
        → drive the live page in headless Chromium   ── the browser-runner
        → judge each criterion
        │
        ├── all pass → write an "acceptance-run" NOTE   (done)
        │
        └── something failed → write an "acceptance-fail" NOTE
                              + file an "improve this tool" ticket,
                                carrying the failing criteria
                                     │
                                     ▼
                              a fix agent loads the PLAN + NOTES,
                              fixes the tool, appends a NOTE, redeploys
                                     │
                                     └── re-run acceptance …
```

Every arrow is machine-driven. The criteria come from a document the machine
wrote; the browser run is triggered by an agent; the verdict is written back
into the same travelling docs; a failure becomes a work item that the fix
pipeline already knows how to process; and the fixer loads the history before
it touches anything — so it won't undo a deliberate decision.

**This is the whole idea:** a tool that documents its own definition of
working, tests itself against it in a real browser, and repairs itself when it
falls short — while keeping a written record of every decision along the way.

---

## What's actually proven (not just designed)

- **Tools write their own PLANs** — the generator created a tool and documented
  it, criteria and all, unaided.
- **Agents write their own fix NOTES** — a broken game was recreated through the
  system with both real bugs fixed, leaving the first machine-authored fix
  notes.
- **The static checker runs live** — its first sweep of a real site correctly
  caught a hallucinated selector and a delivery-mechanism mismatch, filing
  tickets for both.
- **The browser tier is self-driving** — an agent drove a live tool in headless
  Chromium against its PLAN's criteria and wrote the first machine-authored
  acceptance verdict, with no human involved.

Along the way the same mechanism surfaced and fixed several of its own
blind spots — an agent that trusted a page's visible label over a buried
requirement, a prompt that asserted a delivery mechanism the pipeline never
actually performed, an infinite loop that had been misdiagnosed as a memory
leak. Each became a documented correction rather than lost tribal knowledge.

---

## Design principles worth repeating

- **Criteria describe delivered reality, not aspiration.** If the pipeline
  ships inline JavaScript, the criteria say so; a "someday we'll extract it"
  belongs in a roadmap, not in a check that fails every run.
- **Static checks confirm, never refute.** A runtime-built element passes on
  its anchor; only a truly-absent anchor fails. This is what makes the cheap
  tier trustworthy.
- **"Passed the checks" is not "works" until the top tier says so.** The whole
  ladder exists because that gap is real — we watched it ship bugs twice.
- **The machine maintains the docs.** Humans don't write or update travelling
  docs; the agents that change a thing write its history. That's the only way
  documentation stays true.

---

## Where it goes next

- **Close the loop on failure, live** — demonstrate a genuine acceptance failure
  flowing into a fix ticket and back through repair.
- **Make it continuous** — fire acceptance automatically after every
  tool creation, recreation, and repair, plus a periodic sweep.
- **Deeper behavioural checks** — mobile profiles, and *interaction* tests that
  assert a tool actually calculates the right answer, not just that it boots.

---

*Where the detail lives: `RUNBOOK_travelling_docs.md` (operating manual + status
tracker), `PLAN_travelling_docs.md` (the design spec),
`PLAN_tool_acceptance_runner.md` (the browser-runner contract),
`RUNNING_NOTES_travelling_docs.md` (the chronological build log).*
