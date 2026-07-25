# HANDOFF — image provider routing (`bugs_closed/011`), resume point 2026-07-20

**Start a new chat from this file for the provider-routing thread.** Self-contained:
assume no prior context. Sibling entry point for the wider workstream is
`HANDOFF_imagery_best_in_class.md`; this one covers only who-generates-which-image
and what 011 left open.

> ## ✅ THE BUG IS CLOSED — 2026-07-25. Read this before the rest of the file.
> `011` moved to **`bugs_closed/011_HANDOFF_2026-07-18_hero_kind_routes_to_a_model_that_cannot_render_text.md`**;
> its **§8 is the closure evidence chain** and supersedes anything below that reads
> as outstanding. In short: R1 live since `v1.0.1139` and proven fleet-wide by a
> **zero** — 60 active SDXL assets remain, **none generated on or after 2026-07-18**.
> The council residual is live and **APPROVED at round 9** (corr `e996bf0a`, 8/8
> approve, 2026-07-24 20:33Z), and §4 item 1's deferred live-fire proof **fired**:
> an `UNROUTED_IMAGE_KIND` row for kind `scratch_unrouted_011`, 2026-07-24 20:45Z.
> **§4 items 2/3/4 (R2/R3/R4) are no longer tracked here** — they were never routing
> defects and now live in **`features_open/022`** (legibility guard + the 60-asset
> sweep) and **`features_open/023`** (evidence-based figures; the SVG boundary).
>
> **What is still worth reading below:** §3 (how routing works — unchanged and
> current), §6 (the traps, all still live), §7 (key files). **What is stale:** the
> "NOT proven end-to-end" paragraphs in §2 and the owed items in §4.
>
> **One live artefact the closure deliberately left alone:** the gibberish SDXL hero
> is still the header on `leopardessconsulting.co.uk/how-it-works.html`. Stale
> artefact, not a reproduction — regenerating needs no code change. Owner-flagged;
> in `features_open/022`.

---

## 1. Where we are in one paragraph

`bugs_open/011` R1 is **fixed, deployed and verified in the running binaries**, but
**not yet exercised by a single real generation**. The defect was never really "hero
images use the wrong model" — that was the third symptom of a mechanism: provider
selection was a hand-maintained `switch` whose `default:` branch silently chose the
weaker provider, so a kind nobody routed was indistinguishable at runtime from one
deliberately placed there. That switch is now an enumerable table plus a pure
function, and an unrouted kind is detected and named in the log. `hero` — the
fleet's largest kind — joined the routed set, and sites can now override the choice
in **config rather than code**. R2/R3/R4 of 011 are untouched and still open.

## 2. State: what is true right now

**LIVE on `v1.0.1139`** (both services; pods started 2026-07-20 07:35). Verified
against the running binaries, never the tag, using **log-message strings** — the
Docker build (`-a -installsuffix cgo`, alpine) does **not** retain `case` values, and
grepping for those produces a false negative that reads exactly like a stale deploy:

```bash
# adapter — the routing + the unmigrated-kind guard
kubectl exec -n ai-persona-system <image-generator-adapter-pod> -- \
  sh -c 'strings /app/image-generator-adapter | grep -c "UNROUTED KIND"'      # → 1
kubectl exec -n ai-persona-system <image-generator-adapter-pod> -- \
  sh -c 'strings /app/image-generator-adapter | grep -c "routed_kinds"'       # → 1
# chassis — the action layer that resolves the site preference
kubectl exec -n ai-persona-system <agent-chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "site provider preference applied"'  # → 1
```

> **UPDATED 2026-07-20 ~11:50Z — PROVEN end-to-end.** Eight assets generated on
> dartsonline.com (`5fe8785b`) 10:41–10:56Z: 1 `hero` + 7 `icon`, all
> `origin_model = 'banana/gemini-3-pro-image-preview'`. Adapter log (replica `-pl6jc`)
> shows `"msg":"generateImage: dispatching to provider","kind":"hero","provider":"banana"`
> at `dynamic_adapter.go:569`; no `UNROUTED KIND` on either replica. Trap: the adapter
> runs **two replicas** and only one carried traffic — grep both before reading an empty
> log as "nothing generated". Full evidence in `bugs_open/011` §6.

**NOT proven end-to-end** *(morning state — superseded above)*. Zero assets have
generated since the roll
(`SELECT count(*) FROM assets WHERE created_at > '2026-07-20 07:35'` → 0),
consistent with the owner's tool-imagery HOLD (see the memory / `bugs_open/020`).
So no hero has actually travelled the new path. **The first hero generated after the
hold lifts is the real proof**, and the check is one query:

```sql
SELECT asset_key, origin_model, created_at FROM assets
 WHERE asset_key LIKE 'hero%' ORDER BY created_at DESC LIMIT 5;
-- expect origin_model = 'banana/gemini-3-pro-image-preview' on anything new
```

**Council verdict was REVISE, not APPROVED** (`e996bf0a-4cdd-40fa-8ff0-1f1a76c3d181`,
three rounds). No `Council-Reviewed:` trailer was claimed — doing so on a
non-approved submission registers as MISMATCH in the 098 coverage report. The
residual objection is real and is item 1 of §4 below.

**Commits:** `6896ce22e` (the fix) · `dbff03308` (working docs + missteps) ·
`743195773` (live verification, and correcting 028's stale caveat).

## 3. How routing works now (the 60-second model)

`internal/adapters/imagegenerator/routing.go` — **new file, this is the source of truth**:

```go
var kindProviderRouting = map[string]string{
    "icon": banana, "logo": banana, "illustration": banana, "infographic": banana,
    "sprite_sheet": banana, "content_hero": banana, "hero": banana,
}
func routeProvider(kind, providerHint string) routingDecision  // pure, unit-tested
```

Precedence, highest first:

1. **`provider_hint`** — the site's own preference. Resolved in the action layer by
   `providerForKind` from `imagery_style_guide.provider` (guide-level **or** per-kind
   under `kinds.<kind>.provider`), shipped to the adapter as data. Values: `"banana"`
   | `"stability"`. Anything else → warn, fall through. **The adapter has no DB
   handle**, which is *why* routing was hardcoded in the first place; resolving in
   the action and sending the answer as data is what makes the decision site-owned.
2. **The kind table** above.
3. **Stability**, as fallback — and if a *non-empty* kind reached it, that is flagged
   `UnmigratedKind` and logged as `UNROUTED KIND` with the valid set listed. An
   **empty** kind does not warn: legacy callers predating the field are a documented
   Stability path, and a warning that fires constantly is one nobody reads.

Override semantics mirror `avoidForKind` exactly: **a per-kind override replaces the
guide-level value wholesale, including when empty** — a site pinning one provider may
deliberately want one kind left on the fleet default.

To put a site's heroes back on SDXL, no code change — one spec edit:
```sql
-- site_specs, aspect 'imagery_style_guide', is_current = true
-- guide-level:      {"provider": "stability", ...}
-- or per-kind only: {"kinds": {"hero": {"provider": "stability"}}, ...}
```

## 4. What is left — in priority order

> **STATUS 2026-07-20 evening — item 1 is BUILT and LIVE; see `bugs_open/011` §7 for the
> full account.** The mechanism below was implemented exactly as prescribed (adapter
> reports → chassis persists to `agent_error_log`), verified on both replicas of both
> services, and reached prod via an owner sweep (`bca5d8255`, v1.0.1140) while the
> council verdict was still REVISE — **no `Council-Reviewed` trailer is claimed.**
> Round 8 (corr `e996bf0a`, orchestration `49512359`) was queued at ~18:10Z and is the
> outstanding item. Two council-imposed constraints must not be undone: the **sender
> allowlist** (guardian hard veto — an unconditional parse in shared dispatch is a
> fleet-wide ungoverned bus) and the **absent ≠ malformed ≠ partly-dropped** distinction
> (bug_historian, twice). Live-fire proof is deliberately still owed.

**1. ~~The council residual: `UnmigratedKind` is a log line, not a record.~~ ✅ DONE AND
LIVE on `v1.0.1140` (2026-07-20 18:58 BST) — built by another thread, not this one.**
The adapter now reports `reported_conditions` in its success response and the chassis
coordinator persists each to `agent_error_log` via `persistReportedConditions`
(`coordinator.go:2258`, commit `8ec9e2ab8`) — exactly the *adapter reports → action
persists* shape prescribed below. Verified in the deployed binaries: the three condition
codes on **both** adapter replicas, and the persister's own log strings on the chassis.
**Full account and the two design constraints that must not be undone (the sender
allowlist, and absent ≠ malformed ≠ partly-dropped) are in `bugs_open/011` §7** — read
that before touching it; its own council review was still running at round 8.
The paragraph below is retained because the trap it names is what that implementation
had to avoid, and remains the trap for anyone changing it.

*Original statement of the residual:*
`bug_historian` objected (high → medium across rounds) that detection living only in
process logs still depends on someone tailing the right pod, which this repo's own
history says is unreliable. The platform already has the right shape:
`agent_error_log(severity, resolved, work_item_id, context)`, or a `site_work_items`
row for anything that should demand action.
> **The trap, and it is the whole reason this was not just done:** the adapter has
> **no DB handle at all** (`grep 'sql.DB\|pgxpool' internal/adapters/imagegenerator/dynamic_adapter.go`
> → nothing), so persisting from there means giving an adapter service a database
> dependency. The tempting shortcut is to detect it in the **action** layer, which
> does have a DB — **do not**: the action and the adapter are separate services on
> separate images, so the action would be predicting a routing table that may not
> match the one actually deployed. That is the dedup-index↔Go-list drift class this
> platform has already been bitten by. Correct shape: **adapter reports the condition
> in its response → the action, which has the DB and the orchestration context,
> persists it.** A queued diagnosis item already covers this drift shape:
> `5db192c5-4f3b-4140-9e77-e9a65548bb06` (`awaiting_diagnosis`).

**2. `bugs_open/011` R2 — a text-legibility guard before publish.** Still open and
arguably now the most valuable of the three: the capable model still misspells inside
images (the owner's own map rendered "REPRETITIVE"), and **nothing in the pipeline
reads rendered text**. Generation reports success. Shape: OCR/vision pass after
generation → flag misspellings and any number not in the request → work item → HITL,
the same check→work-item→HITL shape as the claims and voice gates. Never auto-publish
an image whose text failed.

**3. `bugs_open/011` R3 — infographic numbers from `site_specs.evidence_base`**, so an
infographic structurally cannot state an unverified figure. Ties into the
claims-verification layer.

**4. `bugs_open/011` R4 — keep code-rendered SVG for exact data.** Generated
infographics are good enough for *explanatory* graphics now; they remain wrong for
charts whose values must be exact, selectable, translatable and screen-reader
accessible.

**5. Cost — ANSWERED, and the change is APPROVED (owner, 2026-07-20).** Costed from
list prices against real volume: **~14× more per hero, but trivial in absolute terms —
about +$5/month, with the fleet's whole image bill under $15/month.**
Gemini 3 Pro Image **$0.134**/1K–2K image (Google official); SDXL at our exact hero
config (1024², 30 steps) **~$0.0094**. Real volume: **40 heroes / 108 images in July**
(up from 15/46 in June, 8/22 in May). One-off backlog: **89 planned heroes** = ~$11.93
if a sweep drains them. Full table, sources and caveats in `bugs_open/011` §6.
> **The lever, if volume grows:** the **Batch API is exactly half price ($0.067)** and
> our image pipeline is *already fully asynchronous* — Kafka, work items, hand-fired
> sweeps, nothing waits on an image interactively — so the ≤24h turnaround costs us
> nothing we use. **Unverified:** whether `banana/api` and our provider wrapper can
> submit batch jobs at all. That is the highest-value cost work left, and only worth
> doing if volume climbs.
>
> **Watch the growth, not the total** (8 → 15 → 40 heroes/month). At 10× today's
> volume the bill is ~$145/month and batch stops being optional.
>
> Two things remain genuinely unmeasured: **latency** (the adapter's 120s timeout was
> tuned around SDXL's 30–60s; Gemini's is untested), and **actual spend** — the
> platform records **no cost data at all** for image generation (`llm_call_log` covers
> text only), so per-image cost is unknowable from our own data. A cost column on
> `assets`, or reusing `llm_call_log` for image calls, would fix that cheaply.

## 5. What this fix armed — owned by other threads, do not duplicate

- **`bugs_open/028` — every `avoid` list is inert.** Banana discards negative prompts
  outright (`banana/provider.go:18`), and `avoid`'s only destination is the negative
  prompt. **011 R1 did not cause this — it extended it to the largest kind**, since
  `hero`'s `kindDefaults` NegativePrompt genuinely did work on the SDXL path. 028's
  caveat that hero "may still be reaching SDXL" was made false by our deploy and has
  been corrected in place with the evidence. **That thread has since shipped
  `32f2d51e2`** (Banana honours NegativePrompt by folding it into the positive
  prompt), inert until an image roll. Do not "fix" this by routing heroes back to
  SDXL — that trades brand anchoring and legible text for a negative prompt.
- **`bugs_open/027` §4b — the 200-char direction cap.** `maxImageryDirectionInPrompt`
  carries an in-code note that it is sized for *"the only generation backend
  (Stability hosted SDXL)"* and its 77-token CLIP wall, listing Banana at *"~1000+
  char effective; cap could be raised significantly"*, deferred *"until provider
  routing lands"*. **Provider routing has now landed**, so the deferral is due — the
  cap is calibrated for a provider no declared kind uses. A council submission for
  that is already in flight (`786ae8edd`).

## 6. Traps for whoever picks this up

- **The two services must ship together.** The action layer sends `provider_hint`;
  the adapter consumes it. Shipping one is a half-live change. (016b §9, "One image
  tag, two services, different vintages".)
- **Pod-grep is a POSITIVE test only.** Grep log-message strings, not `case` values —
  see §2. A miss on a `case` value proves nothing.
- **Do not keyword-infer the provider from `design_intent.imagery_direction`.** It is
  the obvious reading of R1 and it was tried against all 11 live values and rejected:
  it misfires on at least three. Site `9ec3b9ee` reads *"Minimal photography. Prefer
  abstract geometric constructions…"*, `1244516d` reads *"Photography and
  illustration should be minimal…"* — both contain "photography" while intending the
  opposite, so substring matching misroutes them **silently**, reproducing the very
  bug class being fixed.
- **`imagery_style_guide` and `generate_image_actions.go` are contested ground.**
  `bugs_open/027`, `028` and the imagery-5 thread all work these files. Re-read
  both bugs and `git log` on the files before editing.
- **Absence of `orchestration_state_audit` rows means queued OR dropped**, not
  dropped. A council run of mine was queued ~80 minutes; I read it as a dropped
  dispatch and resubmitted, costing a duplicate run. Check
  `diagnosis_artifacts` for a `fix_plan` on the correlation first, and check consumer
  lag (`kafka-consumer-groups.sh --describe --group generic-requests-group`) — a
  frozen offset with a growing end is a backlog, not a break.

## 7. Key files

| file | what |
|---|---|
| `internal/adapters/imagegenerator/routing.go` | **the routing table + `routeProvider`** — source of truth |
| `internal/adapters/imagegenerator/routing_test.go` | 6 tests; the unmigrated-kind guard is the one that must never regress quietly |
| `internal/adapters/imagegenerator/dynamic_adapter.go` | `generateImage` binds the decision to provider instances; `ImageRequestData.ProviderHint` is the wire field |
| `platform/orchestration/actions/imagery_style_guide.go` | `providerForKind` + the `Provider` spec field |
| `platform/orchestration/actions/generate_image_actions.go` | resolves the hint, emits `provider_hint`; also holds `kindDefaults` and the 200-char cap |
| `bugs_open/011_HANDOFF_…` §6 | the full account: what shipped, what is owed, the verification table |
| `016b` §9 "A dispatch table's `default:` branch is a silent bug factory" | the transferable pattern |

## 8. Process notes from this thread

- The **diagnosis loop should have been filed before asserting the mechanism**, not
  after — CLAUDE.md's "Diagnosis before debugging" was corrected on 2026-07-19 to make
  that the default for exactly this class of claim (a mechanism, a structural
  property, a fleet-wide behaviour change). Filed retroactively as `f6e6a732`; it
  gathered 5 evidence bundles, completed, and produced **no verdict**, leaving the
  claim neither validated nor refuted by the loop. It paid off sideways: the 028
  thread cited the resulting work item as justification for *not* duplicating the
  routing list in the action layer.
- The same mechanism was found **independently by another thread hours earlier** as
  `bugs_open/027` (`directionAppliesToKind`'s `default: return true`). Two threads,
  one flaw, one day — which is corroboration, but also a reminder to grep
  `/bugs_open/` **and `/bugs_closed/`** and the diagnosis queue before writing up a
  pattern. There is also a **"five-place checklist for a new kind"** in
  `HANDOFF_imagery_best_in_class.md` that overlaps `kindProviderRouting`; whoever next
  adds a kind should reconcile the two.
