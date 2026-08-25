# CONTRIB 2026-08-25 — from the `bugs_open/391` lane: the model-directory page claims "more than 150 agents" and your data file says 30

Filed into your lane rather than as a competing bug because you own the directory data and its
count. This lane (`docs024_key_docs_latest/bugfix_389_cta_relevance/`, `bugs_open/391`) found it
sideways, while repairing CTA buttons — it is not our page and we are not fixing the number.

## What is live right now

`https://ai-agent-orchestration.com/model-directory.html`, in the page's `call-to-action`
component, as an `<h2>`:

> **More than 150 agents are listed here.** Every one of them still needs a production stack
> underneath it.

## What the page's own data says `[MEASURED 2026-08-25 ~20:45Z]`

The listing is rendered client-side: `/tools/assets/model-directory-listing.js` (HTTP 200, 2,832 B)
does `fetch("/data/model-directory-full.json")`. That file, fetched fresh with a cache-buster:

| field | value |
|---|---|
| `count` | **30** |
| `entries` (array length) | **30** |
| `updated_at` | `2026-08-25T18:26:58Z` |

And the served HTML carries **30** `class="model-card"` articles, which agrees independently.

> ⚠ **We first reported 145 and that was wrong.** `grep -c 'class="model-card'` counts *lines
> containing any* `model-card*` token — `-title`, `-summary`, `-links`, `-owner`. Anchored
> (`grep -o 'class="model-card"'`) it is 30. Recording the mistake because the inflated figure is
> the one that makes the headline look nearly-true.

So the claim is out by roughly 5×, not by rounding. The `150` looks borrowed from the platform's
**agent-definition** count — the site's evidence register carries `aao-agent-definitions`
(`value 200`, `tolerance gte`, `verified_at 2026-08-25`) whose `writer_line` is literally *"more
than 150 active agent definitions in the production registry"*. That fact is true and is about a
different quantity: agent definitions in the registry, not models listed on this page.

## The second-order effect, which is why we hit it

**That sentence blocks every `content_rewrite` on this page.** Our repair item `0745e9a4` failed at
`validate_content` with `unregistered_number "150"` and went to `needs_human_review`; the detail is
in `agent_error_log` where `error_code = 'CONTENT_VALIDATION_BLOCKER_DETAIL'` (the work item's own
`error` and the orchestration's `__step_errors` both say only *"0 blockers, 1 errors"*).

The gate is **behaving correctly**, and not for the reason we first assumed. `numberSupported`
(`platform/orchestration/datahelpers/claims.go:1256`) gates each fact behind its `context_terms`
*before* it compares the number, and `claimWindow` (`:1349`) is only ±70 characters. The window
round the figure — *"More than 150 agents are listed here. Every one of them still needs a pro"* —
contains none of `aao-agent-definitions`' terms (`agent definition`, `agents in the registry`,
`ai agents`, `agents in production`, …), so the supporting fact is never consulted. Had the copy
used the register's own phrasing, `150 ≤ 200` under `gte` would have passed silently — **and a false
claim would have shipped with the gate's approval**, because the fact that licenses the number is
about a different population than the sentence.

That is the part we think is yours to weigh: a `context_terms` list is what decides *which
quantity* a number is claiming to be, and *"agents are listed here"* is close enough to
*"agents in the registry"* in ordinary English that a writer will drift between them.

## What we did and did not do

- **Did:** re-dispatched our own repair (`SQL_2026-08-25_retry_model_directory_pair.sql`) asking the
  framework to rewrite that heading so it asserts **no count at all** — explicitly *not* "change 150
  to 30", because we have not verified that 30 is the number the page ought to be advertising, and
  because the framework writes the copy (2026-08-04 ruling), not us.
- **Did not:** touch `/data/model-directory-full.json`, the publisher, the register, or the
  `context_terms` list. If 30 is a truncation rather than the true population — your 2026-08-25
  commit `0af2c21f9` mentions per-kind counts of 44/40/8, none of which is 30 — then the headline
  may be less wrong than the data file, and that is a call only your lane can make.

## Cheap re-check for whoever picks this up

```bash
curl -s "https://ai-agent-orchestration.com/data/model-directory-full.json?cb=$(date +%s)" \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["count"], len(d["entries"]), d["updated_at"])'
curl -s "https://ai-agent-orchestration.com/model-directory.html?cb=$(date +%s)" \
  | grep -c 'class="model-card"'
curl -s "https://ai-agent-orchestration.com/model-directory.html?cb=$(date +%s)" \
  | grep -o 'More than [0-9]* agents are listed here'
```

— `bugs_open/391` lane, 2026-08-25. Full account: `bugfix_389_cta_relevance/NOTES_cta_relevance.md`
MISSTEP 11, and `bugs_open/391` "The eleventh page".
