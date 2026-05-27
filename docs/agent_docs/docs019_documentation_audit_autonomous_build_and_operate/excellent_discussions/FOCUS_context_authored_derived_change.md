# FOCUS — Context as One Substrate: Authored, Derived, and the Change Layer Between

**Status:** exploratory, active discussion. Captures the thread that arose from "how do I make my development workflow better," which opened into a more general question: documentation vs state, and whether the written/not-written split is even the right one. Companions: the doc-tree, routing, adoption, and curation FOCUS docs (those are about *authored* standards; this reframes where they sit in a larger picture).

---

## 1. The unifying insight

Everything used to ground a reasoning step is retrievable evidence with a different shape: a focus doc, a Go source file, a stack trace, a row in the components table, the original mission statement. To the LLM doing a piece of work these are not different in kind. The **written-vs-other distinction is not the one that matters** — taking that seriously as the organising principle, rather than treating docs as primary and everything else as adjuncts, is the move.

---

## 2. The distinction that does matter: authored vs derived

The atomic-breakdown lens cuts differently here than it did for standards. A standard is *authored* — written once, owned, maintained, with a lifecycle. Most of the rest is *derived* — emitted as a side effect of the system running, true by virtue of being the actual state, not by virtue of being curated.

| | **Authored** | **Derived** |
|---|---|---|
| Examples | mission, focus docs, handoffs, rules, architecture notes | logs, errors, traces, the components table, the specs the pipeline wrote, DB results |
| Owner / lifecycle | yes — someone is responsible; needs maintenance | none — no owner, no maintenance |
| Can it be "wrong"? | yes — it can drift from reality | no — only current or superseded |
| Atom | a rule with an owner and a lifecycle | an observation with a timestamp and no owner, true at that moment |

**Source code sits on the line.** It is authored, but it is also the ground truth of what the system does — simultaneously a document and a state readout. Forcing authored and derived into one model breaks precisely on the owner/lifecycle axis, which is why this is the load-bearing distinction.

---

## 3. The change layer between (diffs of code and logs)

Refinement: the *change* in code and the *change* in logs sits **between state and documentation** — so the picture is three-way, not binary: **documentation / changed-state / state.**

A diff is *derived* (auto-generated, true by being the actual delta) but *narrative* (it tells the story of what changed and implies why). It has the truth-by-being-actual property of derived data and the explanatory shape of authored docs. The change layer is **where state becomes legible as narrative without anyone authoring it** — the most compressed honest account of "what happened." This makes diffs a privileged object: they are the natural audit and learning surface (see the dynamic-creation stresses, §9), because a created or modified artifact *is* a diff against prior state.

---

## 4. Two staleness modes, two fixes

Authored and derived go stale in different ways, so the fixes differ.

- **Authored drift** — the handoff said the dispatcher works one way, the code changed, the handoff now lies. *Fix:* keep the authored layer **thin and pointer-rich** — point at the derived layer instead of paraphrasing it. "Retry logic is in this function" stays true longer than a paraphrase of what the function does, because the paraphrase duplicates ground truth that then moves.
- **Derived snapshot-staleness** — the log was true an hour ago. *Fix:* **fetch at reasoning time**, not paste-time — read the current log, current schema, current source, not a copy from chat-start.

---

## 5. Bridge to the workflow problem (the main focus)

The current pain is mostly **derived-staleness wearing a context-limit disguise.** The loop today: paste fresh code + a focus doc (both accurate at paste time) → the chat burns context → the code moves underneath while we talk → the snapshot we reason over quietly diverges from the repo while the window fills with the history of how we got here → eventually restart.

The deepest version of "make my workflow better" is not a bigger window or better docs. It is shifting from **paste a snapshot, reason until it rots** to **reason against fetched-on-demand current state** — context reassembled fresh each step from the thin authored layer (intent, pointers) plus the derived layer pulled live (current source, current schema, recent errors, relevant component rows). Same references-not-copies thread that has run through the whole design: a chat full of pasted code holds copies; a chat that fetches the file when it needs it holds references.

---

## 6. Where the two routes unify

Infra vs web-work unify, but **not** at "a website is code, a workflow is SQL, an action is Go" — that is true and it is the surface. They unify at the **context layer beneath**: every one of those tasks is the LLM reasoning over a thin authored layer that frames intent plus a thick derived layer that is the live truth of code, data, and runtime. The vertical (infra vs web) changes *which* authored frame and *which* slices of derived state get pulled; the machinery — assemble thin authored context, fetch relevant current state, reason, act, let the result become new derived state — is identical. Get the machinery right once and both routes ride it.

---

## 7. Two directions from here

1. **Taxonomy** — catalogue all the information types (authored, derived, change-layer) and how each atomises. The classification project.
2. **Retrieval/assembly machinery** — the mechanism that lets a development step pull the right *thin-authored-plus-live-derived* bundle on demand. This is the thing that would actually fix the paste-and-rot workflow problem, and is closer to the stated main focus.

Related but different next moves. The second is the higher-leverage one for the workflow goal.

---

## 8. Open / next: long-term stresses (under discussion)

The forward case being discussed: actions and workflows **dynamically created**, appearing as a change in state, hopefully rule-following but possibly not via a full documented path — and whether the task is to reconstruct the documented route from the result, or something else. The stresses this places on the authored/derived model (provenance gaps, reconstructed-rationale as a third epistemic category, the reversed arrow of artifacts→docs, fluke-codification, conformance-vs-route, outcome-grounded justification) are being worked through in-thread and will be folded in.
