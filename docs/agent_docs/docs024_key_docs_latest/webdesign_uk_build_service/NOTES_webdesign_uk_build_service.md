# NOTES — webdesign.uk build service

Running record, append-only, **newest at the bottom**. Evidence, commands, what
the system actually said, and every misstep.

---

## 2026-07-28 — session 1: grounding the plan

Owner asked for thinking and planning only, and for prior discussion to be found
first. Both done before a line of the plan was written.

### Prior art found (the plan is not new — it was recorded)

`webdesign_couk/PLAN_2026-07-27b_buying_design.md` §8 already records this
direction in the owner's own words, including *"we stand up a copy chassis in its
own cluster with its own database"*, and explicitly says **recorded, not started**.
It also already did two of the checks I was about to repeat: that
`cmd/tools-api`/`internal/tools-api` exist with exactly one live endpoint, and
that `idea_uk_vm_site/` is the precedent that has run to a completed paid
transaction.

**Misstep avoided by looking:** I would otherwise have written §8's content again
as though it were a finding. The prior-art search cost about four minutes.

Also relevant and read: `vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md`
(the estate this box would join, and the owner's pull-only ruling),
`webdesign_couk/PLAN_2026-07-27_phase2_buyer_track.md` (superseded, but its
no-figures rail survives), `stripe/PLAN_stripe_billing_integration.md` (May, and
superseded in practice by idea.uk's working `billing.go`).

### What is actually built — checked, not assumed

**tools-api holds no chassis coupling.** `cmd/tools-api/main.go` is config → DB
pool → gin router. Grep for `kafka|Kafka` across `cmd/tools-api/` and
`internal/tools-api/`: **no matches.** Routes are `/health` plus
`/api/v1/tools/gauntlet/{round,position,defend}` with CORS → rate-limit →
input-cap (`internal/tools-api/api/server.go:32-58`).

**core-manager does have an HTTP → pipeline seam**, and it is admin-only:

```
internal/core-manager/api/server.go:108   apiV1.Use(middleware.AuthMiddleware(authConfig))
internal/core-manager/api/server.go:134   adminGroup.Use(middleware.AdminOnly())
internal/core-manager/api/server.go:227   pipelineGroup.POST("/:name/trigger", ...)
```

So "we have no system to directly call the chassis" is right **for untrusted
callers**, and slightly wrong in general — the seam exists for the admin
dashboard. Worth knowing, because P4 may be able to reuse its handler rather than
invent one.

**Site creation is a human writing SQL.** oufe.com was created by
`oufe/SEED_2026-07-25_oufe_site_and_specs.sql` (site row + specs, applied out of
band with `psql -f`, deliberately *not* through the migration runner). Live check:

```sql
SELECT aspect, is_current, created_at::date FROM site_specs
 WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com') ORDER BY created_at;
-- 12 aspects seeded 2026-07-25: design_intent, mission_brief, identity,
-- roadmap_brief, vertical_landscape, strategy, briefing, resolved_composition,
-- imagery_style_guide, classification, content_direction, submission
```

**Build throughput, for the "next day or so" promise:**

```sql
SELECT s.domain, count(p.id) pages, min(p.created_at)::date, max(p.created_at)::date
  FROM sites s LEFT JOIN pages p ON p.site_id=s.id
 WHERE s.domain IN ('webdesign.co.uk','oufe.com','fundamentallyai.com','relojistas.com','idea.uk')
 GROUP BY 1;
-- oufe.com             8 pages  2026-07-25 → 2026-07-28
-- fundamentallyai.com 13 pages  2026-07-20 → 2026-07-25
-- relojistas.com      20 pages  2026-07-16 → 2026-07-28
-- idea.uk             21 pages  2026-06-21 → 2026-07-28
-- webdesign.co.uk     99 pages  2026-07-25 → 2026-07-27   (PORTED, not generated)
```

**[Caveat, and it matters]** those spans are *elapsed calendar days on a
human-driven build*, not machine time. They are evidence that "next day or so" is
achievable with a human in the loop; they are **not** a measurement of how long
the pipeline takes. Do not quote them as build times.

**Deploy path.** Default repo `sites` → GitHub Action → B2; per-site override to
`vm-sites` resolved by `resolveGitRepoNameDB` (`git_repo_resolution_test.go`
documents the failure it prevents: a VM site's artefacts silently landing in the
default repo). Only idea.uk and relojistas carry `github_repo='vm-sites'`; only
relojistas has a non-empty `deploy_config`.

**Cost is measurable.** `llm_call_log`: 45,205 rows, 2026-03-25 → 2026-07-28. So
the cost of one full site build is a query, not an estimate. This is the single
cheapest thing that could be done before pricing anything.

### Fleet size — the finding I did not go looking for

```sql
SELECT count(*) total, count(*) FILTER (WHERE status='deployed') deployed,
       count(*) FILTER (WHERE domain LIKE 'pool-%') pools FROM sites;
-- total 32 | deployed 14 | pools 17
```

The buyer-track positioning says *"about a thousand sites"*
(`webdesign_couk/README_where_we_are.md:405`;
`SUMMARY_2026-07-28_what_the_news_feed_taught_us.md:21`). Traced back, ~1,000
enters the record as a **target** in the scale arguments
(`OPEN_THREADS_RESTART_LIST.md:344`, `robot_hands_gripper_dossier/NOTES…:636` —
*"9 of the 296 exist for 2 of ~1,000 sites"*, *"at 1,600 domains"*), where it is
used legitimately to argue about per-site Go actions. It has drifted from target
to present-tense claim in outward copy.

It may be true of **domains owned** rather than **sites built**. Written up as
PLAN §12 for an owner decision, because it is already in outward-facing prose and
webdesign.uk would be selling on it.

### webdesign.uk DNS

```
dig +short webdesign.uk A      → (nothing)
dig +short webdesign.uk NS     → (nothing)
dig +short webdesign.co.uk A   → 172.67.192.20 / 104.21.92.109  (Cloudflare)
```

Owner confirmed mid-session: **"I haven't pointed the dns yet."** So the empty
result is expected, not a sign the domain is unregistered.

### An asset with no documentation

The cluster runs a **`wireguard` deployment** (`linuxserver/wireguard:latest`,
manifests under `deployments/kustomize/services/wireguard/`). Grep across
`docs024_key_docs_latest/` returns **no mention of it in any workstream**.
**[UNVERIFIED — I do not know what it is for or whether it is in use.]** If it is
a working private transport to the VM estate it could change the P4 design. Noted,
not pursued; it is a P4 question, not a P0 one.

### Two dead ends, recorded

- **`internal/auth-service/subscription/`** looked like reusable billing — it has
  `StripeCustomerID`, `StripeSubscriptionID`, tiers, quotas. It is a **dormant
  skeleton for the original per-seat SaaS**, not for selling sites:
  `SELECT count(*) FROM subscriptions;` → `ERROR: relation "subscriptions" does
  not exist` in `clients_db`. Do not plan reuse around it. idea.uk's `billing.go`
  is the live path and it is one-off payments, which is what this product needs.
- **`pool-*.internal` sites** (17 rows, `status='pool'`) looked like they might be
  pre-built inventory a customer could be assigned. `content_data` on
  `pool-web-tech.internal` is `{}` — they are empty shells. Not investigated
  further; not load-bearing for this plan.

### Misstep

Queried `SELECT type, name FROM agent_definitions` → `ERROR: column "name" does
not exist`. The column is `display_name`. Cost one round trip; the fix is in the
RUNBOOK so the next thread does not repeat it. This is the "schema first: `\d
<table>` before writing SQL" rule in CLAUDE.md, skipped because the query felt
too small to check.

---

## 2026-07-28 — two corrections to this file's own first entry, both from checking a memory line

### 1. idea.uk has not sold anything to a stranger

I wrote that idea.uk *"has taken real money"* and *"survived a real sale"*. The
transaction was real; **the buyer was the owner**. Genuine external buyers: still
zero. The order that looked external on 28 July was a test, and the thread that
inferred otherwise had already recorded its own correction:

```
idea_uk_vm_site/HANDOFF_RESUME_idea_uk_vm_site.md:17
> Genuine external buyers: still ZERO.** One order on 28 July looked external and was a …
idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md:2764
> genuine external buyer. I inferred "external" from two pieces of circumstantial evidence — an IP …
```

**What caught it:** reading `MEMORY_workstreams.md` to add this lane's entry. Its
idea.uk line carries the correction, so the check cost one grep. **What would have
caught it earlier:** reading the lane's own `HANDOFF_RESUME` — which I listed as
prior art and did not open, because I had already got what I wanted from the PLAN
and the EVIDENCE doc. *A workstream's resume doc is where its corrections live;
the plan is where its intentions live.*

**Why it matters more than a wording fix.** The claim as written implied demand,
and demand is the entire question this product faces. Corrected in PLAN §2a, where
it now argues **for** the phasing rather than decorating it: idea.uk is complete,
verified, working and has sold nothing to a stranger. That is the estate's most
recent expensive lesson and it says *build the shop first*.

### 2. The unit cost is a range, and it tracks artefact length

I cited `EVIDENCE_2026-07-27_ai_unit_economics.md` — correctly, as a floor. But it
has since been superseded by a complete measurement:

```
idea_uk_vm_site/HANDOFF_RESUME_idea_uk_vm_site.md:42-43
> ~92% of spend, so cost tracks report *length* — 26,264 chars here vs 13,227 on 07-27, and roughly
> double the cost. **Quote "~$1.20–$1.45 depending on length", never one number.**
```

**Output tokens are ~92% of spend**, so cost is a function of how much artefact
gets produced. Transferred to PLAN §7 with the consequence spelled out, because
for a *website* it bites much harder than for a report: a 5-page site and a
40-page site are not one product at one price. **Either the deliverable is capped
or the price is a range.** That is a product decision that has arrived from a
measurement, which is the right direction of travel.

**The general shape of both misses is the same, and it is the one in `WRONG_CALLS`
from earlier today:** I read the documents that stated the position and not the
one that recorded the corrections. `EVIDENCE_…` is even named as though it settles
the matter — and it did say clearly that it was a floor, which I repeated. Being
faithful to a superseded source is still being wrong.

---

## 2026-07-29 — session 2: Fable 5, the offer rulings, and one landmine imported

### "We're using Fable 5" — measured before acting on it

The owner asked for the plan to be re-checked "now that we're using Fable 5". The
first check was whether the *fleet* is:

```sql
SELECT model, count(*), max(created_at)::timestamp(0) latest FROM llm_call_log
 WHERE created_at > now() - interval '4 days' GROUP BY 1 ORDER BY 3 DESC;
-- claude-sonnet-5    1468  2026-07-28 22:19:56
-- mistral-small3.1     85  2026-07-28 21:25:19
-- gemini-pro-latest    33  2026-07-28 20:30:37
-- claude-sonnet-4-6   311  2026-07-28 19:57:58
-- claude-haiku-4-5      1  2026-07-27 11:06:13
-- claude-opus-4-6       2  2026-07-25 18:55:13
```

**Zero `claude-fable-5` rows.** So the phrase covered two different things — the
session's own model, and an intention for the builds — and only the second is a
change to plan. Clarified with the owner: **both**, and the builds should be
planned on Fable. Written up as PLAN §7b.

**Why the distinction was worth four minutes:** DB model config is live
immediately (CLAUDE.md), so "we're on Fable" could have been read as "the lanes
already run it" and the three pre-flight checks below would have been skipped as
redundant.

### Fable 5 facts, from the skill on the day (not from memory)

Read via the `claude-api` skill, 2026-07-29. The rows that bear on this plan:

- **$10 / $50 per MTok** — 2× Opus 5 ($5/$25), ~5× Sonnet 5's introductory
  $2/$10. Sonnet 5's intro rate **ends 2026-08-31** → $3/$15, which raises the
  fleet's baseline ~50% on its dominant model.
- **Thinking is always on.** `thinking: {type:"disabled"}` and
  `{type:"enabled", budget_tokens:N}` both return 400.
- **`temperature` / `top_p` / `top_k` all return 400.**
- **Requires ≥30-day data retention** — a ZDR org gets `400
  invalid_request_error` on *every* request, with a payload that is otherwise
  perfectly valid. This is the one that would burn an afternoon: the error names
  the request, not the org setting.
- **Minutes-long turns are normal**, and **`stop_reason: "refusal"`** must be
  handled before reading content.

**Consequence recorded in §7b:** a model swap is **not** a config-only change if
the chassis call layer passes any of those params — and it demonstrably sets
params (all 16 council seats set `max_tokens=8000`). The P0 grep is therefore a
prerequisite, not a nicety, and it is ordered *before* any lane is pointed at
Fable precisely because config is live on write.

### Landmine imported from `bugs_open/139` — a per-IP limiter that never was

Picked up from `MEMORY_workstreams.md` (refreshed 07-29) while adding this lane's
entry. The gauntlet/island lane found that its per-IP rate limiter keyed on a
**constant**: `client_ip_hash = sha256("172.18.0.1")` — the docker gateway — in
**83 of 83 rows**. Every visitor shared one bucket.

Why it transfers directly: **we are about to build the same shape** — a Go
service behind a reverse proxy behind Cloudflare — and our per-IP limit is the
control bounding §8's spend faucet.

- The real address is in **`CF-Connecting-IP` only**: Caddy overwrites
  `X-Forwarded-For`, Cloudflare strips `X-Real-IP`.
- **`platform/httpguard` does not fix it** — its rightmost-XFF fallback lands on
  the same constant. It reads as a fix. (Noted against P0's httpguard item too,
  which was about SSRF: *two different questions, same file.*)
- **One test machine cannot detect it.** Your own traffic yields one value
  whether the key works or not. The discriminating check is
  `count(DISTINCT <ip key>) > 1` **from two networks**.

Written into PLAN §5.1 as a block-quoted landmine rather than a table cell,
because the "it reads as a fix" part is the whole content.

### Owner rulings this session

Recorded in the PLAN where each bites; summarised for the record:

| ruling | where |
|---|---|
| Full sites, high quality-based price — supersedes cap-or-tiers | §7 (struck through, kept) |
| Full money-back guarantee, **acceptance-gated on the preview** | §7a |
| Corrections carry a fee; **changes paid, our defects free** | §7a |
| Builds on `claude-fable-5` | §7b |
| Preview host = a different, shorter domain, TBS | §6 |
| Thousand-sites figure accepted as-is | §12 (closed) |

**The one place I sharpened rather than transcribed** is the fee boundary. The
owner said "any and every correction can carry a fee"; asked directly, he chose
changes-paid/defects-free. Recorded because the *reason* is load-bearing and
easy to lose: it makes §5.3a's fabrication controls revenue-protective — every
invented detail that ships is a free repair we owe. That is a much stronger
argument for those controls than the reputational one they were filed under.

### Not done, deliberately

No council run: the gate refuses docs client-side (`097_TRIGGER…:116`, scope is
`platform/`, `internal/`, `pkg/`). P4/P5 are still the first submissions.

---

## 2026-07-29 (later) — two of three Fable-5 P0 checks closed; SSRF confirmed unbuilt

### The call-layer grep, done rather than deferred

Read `platform/aiservice/anthropic.go:89-113` directly rather than trusting the
plan's own instruction to "grep for it later":

- **`temperature` is already unconditionally dropped** — a standing guard,
  predating this workstream, with its own comment: *"Claude Opus 4.7+ returns a
  400 for any non-default temperature… the Anthropic client simply ignores it."*
  Nothing to fix.
- **`budget_tokens` → `thinking:{enabled,...}` only if `ai_service.budget_tokens`
  is set** in the calling agent's config. No unconditional send.
- `rag_actions.go`'s `top_k` is retrieval-k (an unrelated concept — read the call
  site to be sure, since the string match alone would have been a false alarm).

```sql
SELECT type FROM agent_definitions
 WHERE is_active AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false
   AND default_config::text LIKE '%budget_tokens%';
-- council-gate, fix-proposer only
```

Neither is a candidate for the build lane. **Consequence: a fresh agent pointed
at `claude-fable-5`, with no `budget_tokens` in its own config, is safe at the
code layer today.** Confirmed against the `claude-api` skill directly (not
memory) the same session: omitting `thinking` or `{type:"adaptive"}` runs
adaptive; `{type:"disabled"}` 400s (Fable-specific — Opus 4.8/4.7 accept
`disabled`); `{enabled,budget_tokens}` 400s; non-default
`temperature`/`top_p`/`top_k` all 400.

### SSRF — read `platform/httpguard` in full; it does not cover this

The plan's §8 risk 1 said "check whether `platform/httpguard` already does this
before writing a second one." It does not, and its own package doc says so:
*"the platform's ONE set of **inbound**-abuse primitives for public HTTP
endpoints."* Grepped `IsPrivate`/`IsLoopback`/`169.254`/SSRF across
`platform/httpguard/` and `internal/adapters/webscrape/` (the live scrape
path a domain-intake flow would reuse): **zero hits either place.** This is a
real, unbuilt gap, not a lookup miss — the first real code this workstream will
need to write is an outbound-fetch guard, not a config change.

Landmine appended (`platform/httpguard is INBOUND-abuse only`) because the
package is genuinely well-built and the name invites exactly this assumption —
reading it is satisfying, which is what makes the gap easy to miss.

### Still open

Org data-retention level (account/console setting, not in the tree) and one
measured Fable-5 build cost from `llm_call_log`.

---

## 2026-07-31 — first live Fable-5 measurement, and a misstep in getting there

### The misstep, recorded before the result

First `kubectl exec -i "$POD" -- sh -c 'cat > /tmp/fable_probe.json'` with the
payload piped via `<` reported success, and a **separate** follow-up
`kubectl exec "$POD" -- sh -c 'wc -c /tmp/fable_probe.json'` read **0 bytes**.
Combining both into one exec call (`cat > file && wc -c file`) showed 329 bytes
written correctly. Root cause not fully isolated — plausibly stdin buffering
across two independent `exec -i` sessions rather than the two pods (there are
2 replicas, but I pinned the pod name explicitly both times). **The fix that
matters: write-then-verify in the SAME exec call**, not two. Not yet promoted
to a landmine — one occurrence, cause not pinned down cleanly enough to state
as a rule others should trust.

### The probe

No Anthropic client in this shell (`env | grep ANTHROPIC` empty, no `ant` CLI).
`agent-chassis` pod carries `ANTHROPIC_API_KEY` (confirmed via `printenv`
before using it), and its image has only BusyBox `wget` — no `curl`, no `jq`
(checked with `which`, not assumed). Sent one real request, `model:
"claude-fable-5"`, no `thinking` key, a 120–150 word "About Us" copy prompt,
`max_tokens: 4096`.

```
HTTP/1.1 200 OK
usage: input_tokens=69, output_tokens=282 (thinking_tokens=25 of that),
       cache_creation=0, cache_read=0
stop_reason: end_turn, stop_details: null
```

**Cost:** 69 × $10/MTok + 282 × $50/MTok = $0.000690 + $0.014100 = **$0.01479**.

### What this closes, and what it does not

- **Closes PLAN §7b item 1** (org data retention ≥30 days): `200`, not the
  retention-configured `400 invalid_request_error`. The org passes.
- **Confirms, live rather than from the skill doc:** the exact $10/$50 per-MTok
  rate; `stop_details: null` on a non-refusal turn; a `thinking` block present
  with empty `"thinking":""` text when `display` is left at its default
  (`"omitted"`) — the reasoning happened (25 tokens, billed) and the text was
  withheld, exactly as documented, now seen on the wire rather than read in
  the doc.
- **Does NOT close item 3** (measure one real build). This is one short
  paragraph, not one page, and nowhere near one site. Written up in PLAN §7c
  with the heading doing the work: *"a FLOOR, not a build cost"* — the exact
  phrasing this workstream already burned itself on once this week with
  idea.uk's $0.641 figure. Marked before the number, not after, this time.

### Housekeeping

Removed `/tmp/fable_probe.json`, `/tmp/fable_resp.json`, `/tmp/fable_headers.txt`
from the pod after reading the response — a production pod's `/tmp` is not a
scratch space to leave things in, even though it would clear on a restart.

---

## 2026-07-31 (later) — the SSRF guard: built, wired into a real live hole, submitted

### Where the fetch actually happens

Traced the webscrape adapter fully before writing anything. `a.httpClient` in
`internal/adapters/webscrape/adapter.go` is shared between two very different
uses:

1. Calls to `api.firecrawl.dev` (the scraping provider's own fixed, trusted
   API) — fine as-is, host never varies.
2. `downloadImage`, called from two places (`adapter.go:596` for a
   screenshot URL, `:679` for **image URLs taken from the `images` array
   in the scraped page's own parsed content**) — the real SSRF surface,
   because that page belongs to whatever domain a customer submitted.

`browser-runner-adapter` was also checked (`run_checks_action.go:688`,
`page.Goto`) — Playwright, a full headless browser navigating the URL
directly. A Go `http.Transport` guard cannot see a browser's own DNS/connect
calls at all, so this is explicitly a different, unaddressed problem — recorded
in the bug file and the register entry rather than silently left implied-fixed.

### Measured, not assumed, before filing or submitting

- `grep -rliE "ssrf|169\.254|metadata"  bugs_open/ bugs_closed/` → nothing.
  Not a duplicate.
- Blast radius, for the council submission's `grounded_in` (per the owner
  ruling: measure it yourself, don't ask the reviewer to):
  ```sql
  SELECT type FROM agent_definitions WHERE is_active AND deleted_at IS NULL
   AND (default_config::text LIKE '%scrape_web%' OR LIKE '%batch_webscrape%');
  -- 10 distinct types: grounded-explainer, vet-practice-verifier,
  --   directory-researcher, research-agent, evidence-researcher, council-gate,
  --   feature-designer, adoption-researcher, fix-proposer,
  --   domain-research-classifier
  ```
  Also checked `orchestration_states` for the same string — 15 rows, but
  spanning only 2026-07-30→31. **Did not cite this as the blast-radius
  figure** — it's a retention-window artefact (the exact landmine already in
  memory: "every history table is on a retention clock"), and the 10-agent
  count is the durable, structural measure.

### The package itself

`platform/fetchguard`, deliberately a **sibling** to `httpguard`, not folded
into it — `httpguard`'s own package doc scopes itself to "inbound-abuse
primitives" explicitly, and adding an outbound guard to that package would
make its own header wrong (the exact trap the earlier landmine entry
describes, one level up). Core design: `Transport.DialContext` resolves the
target itself and checks the *specific address about to be dialed* — not a
pre-resolved hostname — closing the DNS-rebinding check-then-connect gap.
Redirects re-dial through the same transport, so a redirect to a private
target is caught automatically, with no separate redirect-target inspection
to maintain.

### The self-correction, caught by my own test before it shipped

Wrote a comment claiming `ip.Unmap()` was "the exact bypass" needed to
classify `::ffff:169.254.169.254` correctly. Wrote the test meant to *prove*
that claim; it disproved it instead — `netip`'s classifiers already handle
4-in-6 addresses with no unmap step. Fixed the comment to state the true,
smaller reason (`Unmap()` only affects `ip.String()`'s readability in error
messages), renamed the test accordingly, logged in `WRONG_CALLS.md`. The
transferable point, recorded there: **a security-rationale comment with no
prior art to contradict it has no external referee** — the only thing that
catches it is writing the test that tries to *prove* the claim, not just
tests the code that assumes it.

### Test design bug caught the same way

First draft of the redirect-cap test used a live `httptest.Server` (binds
127.0.0.1) as the redirect target — which the private-IP check refuses on
hop 1, before the redirect count could ever climb high enough to exercise
`MaxRedirects`. The test would have "passed" while proving the wrong thing
entirely. Fixed by unit-testing `checkRedirect` directly against fake
requests (no network), plus a best-effort live version gated on the
environment actually having a publicly-routable-shaped interface (it doesn't,
in this sandbox — both interface-dependent tests skip here, honestly, rather
than being deleted or faked).

### Verification before commit

- `go build`, `go vet`, `go test -race` on both touched packages — clean.
- `git archive HEAD | tar -x` into a scratch dir, rebuilt and retested from
  the **committed tree alone** (not the dirty working copy) — clean. Per the
  shared-tree rule: a green local build is not a green HEAD.
- Council submitted (`097_TRIGGER…`), corr `41bbaca4-25f1-45da-a2c1-28a246a5d07a`,
  queue busy (5 in flight, no ETA printed). Committed immediately with
  `Council-Submitted:` rather than holding the fix uncommitted for the verdict.

---

## 2026-07-31 (later) — P4 planned, and its premise tested live rather than assumed

### The finding that resized P4

`build_queue` + `seed_build_queue` + `build-pipeline-trigger` already implement
the whole queue → site → work-item → build chain, and the trigger is **enabled
and firing every 120s in production**. The chain is idle, not absent: 2 rows in
`build_queue`'s entire history, most recent 2026-03-22. So P4 is one action +
one scheduled task that put a row in a table — not a dispatch pipeline.

### The test, and why it needed a guard

Seeded work items get `status='triaged'`, which is exactly what
`find_dispatchable_site` selects — so a naive test row would have started a
**real site build** (model spend, real pages) in the same trigger run. Read
that query properly before inserting anything:

```sql
WHERE wi.status IN ('triaged','approved')
  AND NOT EXISTS (SELECT 1 FROM site_work_items active
                   WHERE active.site_id = wi.site_id AND active.status='claimed')
```

The `NOT EXISTS` is correlated per `site_id`, so a single `claimed` sentinel
work item makes **one** site undispatchable without touching any other. Test
therefore: pre-create the site + one `claimed` sentinel, then insert the
`build_queue` row. Domain used the RFC 2606 reserved `.invalid` TLD so an
accidental fetch could not reach a real third party.

**Result — passed.** `queued` 15:00:38Z → trigger fired 15:05:34Z → `seeded`,
with a `needs_domain_research` / `triaged` / `domain-research-classifier` work
item whose spec carried the `objective` from my `direction` jsonb intact. That
last detail is the real proof: it shows the documented `direction` contract is
live behaviour, not just a comment. **0 orchestrations ever referenced the test
domain** — the guard held, no build ran. Cleaned up to exact baseline
(`build_queue` 2, `sites` 33, zero leftovers).

### The confound that nearly produced a false negative

A fresh chassis rolled **during** the test (`v1.0.1214` → `v1.0.1215`). The
trigger stopped firing for ~5 minutes and the row sat `queued` — indistinguishable
from "the pipeline is broken" or "my row is malformed". It was neither:
CLAUDE.md's ~300s post-restart dispatch rule. Pod age had been checked *before*
starting (5h58m, safe) precisely because the owner had mentioned a build was
coming, and re-checking it when the trigger went quiet (2m50s) diagnosed it in
seconds.

**Without that check this would have been written up as "build_queue is
broken" — and P4's entire plan discarded on a false negative.** Recorded as a
landmine (`A chassis roll makes a scheduled task look BROKEN for ~5 minutes`).

### Owner rulings this session

- **Poll interval: 15 minutes.**
- **Repeat domains: reject and alert a human.** Identical re-collection →
  `ON CONFLICT DO NOTHING`; but an order for a domain whose row is already
  `seeded` must be surfaced for a human, never dropped. **This makes the
  collector's conflict clause stateful** — it must read the existing row's
  status rather than blindly `DO NOTHING`, and that is the most important
  single piece of logic in P4.

---

## 2026-07-31 (later) — DNS measured, and a three-day-old claim of mine falsified

Owner asked what he needs to do for webdesign.uk's DNS and the price, and named
**`ugg2.com`** as the preview host ("set up in cloudflare now"). Before answering
I ran `dig`. I should have run it on 2026-07-28.

### The claim I had been repeating, and its refutation

The PLAN, the RUNBOOK and the memory file all said **"DNS for webdesign.uk not
pointed yet (owner)"**, sourced from the owner's own remark on 07-28 (*"I haven't
pointed the dns yet"*). It was carried for three days and re-stated in three
places without a check.

```
$ dig +short NS webdesign.uk
alexis.ns.cloudflare.com.
leah.ns.cloudflare.com.
$ dig +short A webdesign.uk
172.67.223.216
104.21.54.51
```

**Delegated to Cloudflare and proxied** (those are CF anycast addresses; there are
AAAA records too). Whether it changed after 07-28 or the original remark meant
something narrower, I cannot tell and **did not need to** — the point is that
three days of derived planning rested on an unchecked second-hand claim about
infrastructure, checkable in one second. `WRONG_CALLS.md`.

### The finding that came free with it

```
$ curl -sS https://webdesign.uk
{
  "error": "B2 returned error",
  "objectKey": "webdesign.uk/index.html",
  "status": 404,
  ...<Code>NoSuchKey</Code>...
}
```

**PLAN §6(i) carried `[UNVERIFIED — I have not read how a domain maps to a bucket
prefix]`. The live 404 states the mapping outright: `<host>/index.html`.** The
question had been marked unverified because it looked like it required reading
the deploy path; it required visiting the URL.

Cross-check, because a mapping could be table-driven rather than host-derived:

```sql
SELECT domain FROM sites WHERE domain ILIKE '%webdesign%';
-- webdesign.co.uk    (one row; NO webdesign.uk row)
```

**The Worker built an object key for a hostname that has no site row at all** —
so it derives the key from the request host rather than consulting a list. That
is the property §6a's route (iii) needs. Marked `[INFERRED]` in the PLAN, not
measured: **one sample, and the Worker source is not in this repo**
(`grep -rn 'objectKey\|B2 returned error'` → no match anywhere). The
`test.ugg2.com` probe settles it.

### ugg2.com — delegated, not serving

```
$ dig +short NS ugg2.com     -> alexis/leah.ns.cloudflare.com.   (ours)
$ dig +short A  ugg2.com     -> 199.59.243.228                   (registrar parking)
$ curl -m 12 http://ugg2.com -> curl: (28) Connection timed out
```

No AAAA, no MX, no `www`, no wildcard. The A record is **not** a Cloudflare
address, so the record is **grey (DNS-only)** — Cloudflare holds the zone but is
not in the request path. The DNS-01/wildcard property PLAN §6 required is
satisfied by the NS delegation; the serving half is entirely unbuilt.

### The cross-lane finding: idea.uk is NOT behind Cloudflare

Checking what "our" VM pattern actually looks like, to answer whether
webdesign.uk should be proxied:

```
$ dig +short NS idea.uk
oxygen.ns.hetzner.com.  helium.ns.hetzner.de.  hydrogen.ns.hetzner.com.
$ dig +short A idea.uk
116.203.204.115                      # the VM itself, bare
$ curl -sS -o /dev/null -D- https://idea.uk | grep -i '^server\|cf-ray'
server: nginx/1.28.3 (Ubuntu)        # no cf-ray
```

**`RUNBOOK_idea_uk_vm_site.md:12` says `DNS (Cloudflare) → the VM`, and its §4a
opens *"idea.uk is behind Cloudflare"* and builds a remediation on it** — headed
*"Restore the real client IP — nothing else works until this is done"*. That
section also asks the reader to *"first confirm whether the DNS record is
actually proxied (orange) or DNS-only (grey)"*. **Answered: neither. Cloudflare
is not in idea.uk's path at all.** Consequences, in the order that matters:

1. Its `limit_req` on `$binary_remote_addr` is **already correct** — nginx sees
   real client addresses. §4a's premise is false and the work is unnecessary.
2. **More importantly, there is no WAF, no Turnstile and no DDoS layer in front
   of a live money-taking box**, because §4a assumed Cloudflare was available as
   "the blocking layer" and it is not. That is the finding worth acting on, and
   it belongs to that lane, not this one.
3. `RUNBOOK:435`'s *"purge the Cloudflare cache for idea.uk"* is a **no-op** —
   someone debugging a stale page would purge a cache that isn't there and
   conclude the purge had failed.

Relojistas, by contrast, **is** proxied (`server: cloudflare`, CF anycast A
records). **So the two VM sites do not share a front-end shape**, which the
estate docs treat as one pattern. Filed as a landmine + a dated correction in
that lane's runbook; not rewritten, per the who-owns rule.

**Transfer to this lane:** the `CF-Connecting-IP` landmine that PLAN §5.1 and the
P1 handoff both lead with applies to *relojistas'* shape, not idea.uk's. Copying
idea.uk's `setup.sh` gets you a rate limiter that is correct **only for as long
as nothing is proxied in front of it** — and P1 will put something in front of it.
The failure is silent and reads as a working limiter either way.

### On the price

Recorded in PLAN §7d with the correction to §11 item 8: **the price was never
blocked on the Fable-5 build measurement**, and saying it was smuggled cost-plus
back into a decision that had explicitly rejected cost-plus. Recommendation
£1,200; the measurement is still owed for margin and for P2's free-tier cap,
where per-unit × volume *is* the whole story.

### Second sample, same session — the inference splits in two

```
$ curl -sS https://webdesign.co.uk/no-such-page-xyz.html
{"error":"B2 returned error","objectKey":"webdesign.co.uk/no-such-page-xyz.html", ...}
$ dig +short A foo.webdesign.co.uk   -> (empty)
$ dig +short A foo.vonc.com          -> (empty)
```

A different host, and the **path** passes through too — so the key is
`<host>/<path>`, on two hosts, one of which the database has never heard of.

**But running the second sample is what showed me I had been conflating two
claims.** "The key is host-derived" (well supported now) is **not** "a wildcard
host would reach the Worker" (untested, and the half route (iii) actually rests
on). Both samples have their own explicit DNS record; nothing I can observe from
outside says a `*.ugg2.com` record would match the Worker's route pattern, which
is Cloudflare-side config I cannot read. And **no zone on the account has a
wildcard record today**, so there is no precedent to reason from either.

Recorded in PLAN §6a as (A) and (B) rather than one confidence level. **The
strong half was making the weak half feel measured** — the trap in
`two-blind-checks-agree-with-each-other`, except here the two checks agreed
about a proposition neither of them tested.

---

## 2026-07-31 (later still) — owner pointed both domains at idea.uk's live box

Owner added `A 116.203.204.115` for **both** `webdesign.uk` and `ugg2.com` and
removed the Worker routes. `116.203.204.115` is **idea.uk's live, earning VM**.

### Measured immediately

```
$ for d in webdesign.uk ugg2.com idea.uk; do curl -sS https://$d | md5sum; done
cf4c46c2b4e0...   cf4c46c2b4e0...   cf4c46c2b4e0...      # byte-identical
$ curl -sS https://webdesign.uk | grep -o '<title>[^<]*</title>'
<title>idea.uk — Where You Take an Idea Seriously</title>
```

Both new domains are **proxied** now (CF anycast A records, `server: cloudflare`),
with the origin being idea.uk's box, which has one vhost — so every hostname gets
idea.uk.

**The app does not validate `Host`.** `grep -n 'r\.Host' main.go service.go
billing.go` → **no match**; redirect targets come from the configured
`PublicBaseURL`, not the request. So the page served on `webdesign.uk` is fully
functional and **a real, payable order can be created from it**, with the customer
bounced to idea.uk partway through checkout. **Not tested — firing it would create
a live order and a real Stripe session.** Stated as reachable, not as exercised.

### Origin probe — what the box actually does

```
$ curl -o /dev/null -D- -H 'Host: webdesign.uk' http://116.203.204.115/
HTTP/1.1 301   Server: nginx/1.28.3 (Ubuntu)   Location: https://webdesign.uk/

$ openssl s_client -connect 116.203.204.115:443 -servername webdesign.uk | openssl x509 -noout -subject -ext subjectAltName
subject=CN=idea.uk
X509v3 Subject Alternative Name: DNS:idea.uk          # idea.uk ONLY — not even www
```

So the origin presents a certificate that **does not match** `webdesign.uk`, and
Cloudflare returns **200** regardless. **That proves the zone SSL mode is "Full",
not "Full (strict)"** — the CF→origin leg is encrypted but unauthenticated. It is
not a guess and it did not need dashboard access: strict mode would have failed
the handshake. Fix when a real origin exists: a free **Cloudflare Origin CA cert**
(15-year, wildcard, no renewal) + Full (strict).

**Also noted:** idea.uk's nginx serves `Strict-Transport-Security: max-age=31536000;
includeSubDomains` — which is now being served **for ugg2.com**, pinning every
future `*.ugg2.com` preview to HTTPS-only for a year in any browser that visits it.
Harmless (previews will be HTTPS) but it removes the click-through on a cert error.

### The thing this changes in the PLAN

**Being behind Cloudflare deletes §6(ii)'s hardest task.** Universal SSL covers a
**proxied** `*.ugg2.com` at the edge, so the wildcard certificate, the DNS-01
challenge, the scoped API token and the renewal timer are all unnecessary. That
machinery was designed for a **direct-to-VM** world, which is the world idea.uk
lives in — and I had carried it forward without noticing the premise had changed
when the preview host turned out to be a Cloudflare zone. **A design decision
inherits the assumptions of the estate it was copied from; re-check them when the
estate changes.**

### Recommendation recorded: do NOT host P1 on idea.uk's box

Two reasons, the first of which is our own tooling:

1. **`setup.sh` has box-takeover semantics (`ufw --force reset`)** and
   `RUNBOOK_idea_uk_vm_site.md:16` already says *never point it at the live
   idea.uk box*. Co-hosting puts P1's provisioning one careless run away from
   resetting a live earning box's firewall.
2. P1 is an **anonymous LLM chat** — a spend faucet taking hostile input — on the
   same disk as `/etc/idea/idea.env` (Stripe keys) and `orders.json`. This is §4.2's
   blast-radius argument, arriving earlier than P3 because the DNS made it concrete.

---
**2026-08-02 (from the idea_uk_vm_site lane, cross-lane courtesy note):** the
owner had us create a CF zone for **webzy.uk** (`aeddc60d…`, pending, NS
alexis/leah, one proxied A → 199.59.243.228 copied from webdesign.co.uk's
pattern) as a new statically-hosted-sites member. NS change is at **GoDaddy**
(webzy.uk is GODADDY-tag, ns17/ns18.domaincontrol.com — NOT DESIGNCONSULT, so
Nominet EPP cannot move it), still pending owner action. Zone details in
idea_uk_vm_site/RUNNING_NOTES §X.38. Worker route/B2 wiring not touched — that
is your lane's machinery.
