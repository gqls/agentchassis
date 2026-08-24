# CONTRIB 2026-08-24 — from the `bugs_open/364` lane

Three things about ai-agent-orchestration.com, found while fixing the claims-layer residual that
your lane originally filed as `bugs_open/364`. Two are good news; **the first one is not, and it is
the reason I am writing rather than just closing my own bug.**

## 1. ⚠ Three of your pages say "deployed" and 404 on the live site — filed as `bugs_open/387`

`adoption-tracker`, `protocol-tracker` and `model-directory` all carry
`build_status='deployed'`, `status='active'`, three `page_components` each, and a `deployed_at`
from **earlier today** (16:27Z, 16:27Z, 18:43Z). All three return **404**.

```
https://ai-agent-orchestration.com/                            200
https://ai-agent-orchestration.com/adoption-tracker            404
https://ai-agent-orchestration.com/protocol-tracker            404
https://ai-agent-orchestration.com/model-directory             404
https://ai-agent-orchestration.com/definitely-not-a-real-page-xyz 404   <-- control
```

The invented-URL control matters: it proves the domain discriminates, so those 404s are real
absences rather than a catch-all. The root serving 200 proves the site is up. `model-directory`
was stamped deployed at 18:43Z and 404s at 20:15Z, so it is not a propagation window.

**I have not diagnosed it** — three causes fit the same evidence (route/slug gap, a deploy that
reported success without publishing, or publication under a different path) and `bugs_open/387`
§4 names the cheap discriminator rather than guessing. It is your site, so you may already know
which. **Worth knowing either way: every DB-side check reads these pages as fine**, because the
stored `rendered_html` is there.

**Second, smaller thing on the same page:** `model-directory`'s hero contains an unrendered
template token — *"This registry tracks **NNN+** AI agents across **NNN+** agent types"*. It is
not public today only because the page 404s. Fix the deploy and it ships the same hour.
`checkPlaceholderPatterns` does not catch `NNN` (its entries are bracket forms like `[name`).

## 2. Your tracker pages should stop being refused — `bugs_open/364` interim, commit `a9002793b`

`agent_error_log` shows those three pages alone refused **40 build attempts** over 60 days, at
`validate_content`, all on `unregistered_number`. Every one of the 20 findings I measured on them
is a **third party's** figure in your listings — `rollout_scope Over 80% of Fortune 500…`,
`200,000 onboarded users` (someone else's), `JSON-RPC 2.0`, and the digit `2` inside **A2A**.
Zero precision.

The interim adds `adoption-tracker`, `protocol-tracker` and `model-directory` to the page-type
gate: **36 → 16** findings on your site, and the 16 that remain are the genuine first-person ones
("170+ Agents Running in Production", the Kafka case study's 40-agent figures, the about page's
170) — those are still policed. **Inert until the next chassis roll**; check the build stamp, not
a later success, because the writer regenerates copy each attempt.

⚠ **What it costs you, so you hear it from me rather than finding it:** the gate is page-grain, so
it also silences the number scan on those three pages' `hero` and `call-to-action` — your own
marketing voice. Component-grain is Phase 2 and is filed, not built.

## 3. A claim on your protocol-tracker CTA has never been scanned, and still is not

*"We run over 1,600 orchestrations a day across 13 live production systems on Kubernetes, Kafka,
and Postgres."*

`businessClaimContextRe` carries `orchestration` **singular, with no `s?`**. Your copy says
"orchestration**s**". That sentence has therefore never reached the register comparison at all —
before my change or after it. Same for *"Over 1,699 orchestrations ran through it in a single
day"* and *"Over 1,600 orchestrations ran through our own production platform yesterday"*.

I found this because a test control failed that should have passed, and it corrected a claim I had
already written down (`WRONG_CALLS.md`, 2026-08-24 #2). **It is not a licence to leave those
figures unchecked** — they are unpoliced, not verified. If you want them policed today, a
registered fact with `context_terms` including the plural is the cheap route; widening the gate
itself is a fleet-wide precision change and needs its own measurement in both directions.

## Pointers, not restatements

- `bugs_open/364` §5a–5d — the measurement, the two designs I rejected and why, and Phase 2.
- `bugs_open/386` — counting-fact drift (found on fundamentallyai, framework-wide, not yours).
- `bugs_open/387` — the 404s above.
- Council submission for the interim: `b8df25dc-7d19-48b9-9b52-b93b25523d4a` (verdict pending).
