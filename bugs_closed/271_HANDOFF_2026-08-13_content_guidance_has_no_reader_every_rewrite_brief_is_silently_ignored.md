# 271 — `spec.content_guidance` has no reader: every `content_rewrite` brief in the fleet is silently ignored

**Filed:** 2026-08-13 · **Lane:** `ai_site_selling_automation` · **Severity: high** —
the field every operator and the platform's own gap-planner use to tell the
writer WHAT TO CHANGE is write-only. The work happens anyway, steered by
nothing but `writer_block` and the existing page, and reports `complete`.
**Class:** structural (a dead channel that looks like the steering wheel).

> **STATUS: OPEN — but FIXED IN CODE 2026-08-15 (`9a7d23c49`), awaiting a chassis
> roll. START AT §9**, which carries the fix, the post-roll checklist that closes
> this file, and a correction to §2/§5 below.
>
> **§2 and §5 are superseded in one respect — read §9 before acting on them.**
> They were written without knowing that the channel ALREADY WORKS under the
> other spelling (`spec.suggestion` → `rewrite_guidance` → the writer prompt),
> which is the fleet convention. So the shipped fix aliases the dead spelling
> onto the live one at the dispatch loader; it is NOT §5's candidate 1
> (injecting into the section brief at `plan_sections`), and §9 says why that
> would have been the worse change.
>
> Original verification: two independent code reads eleven days apart
> plus rendered-prompt probes across six LLM calls (§3) — stated per the
> 2026-07-31 ruling as the substitute for a 090 run; a 090 remains cheap to add
> if a fixing thread wants the independent pass.

## 1. The symptom, as an operator experiences it

You file a `content_rewrite` item with careful `content_guidance` ("state X,
remove Y, add Z"). The item completes; the page is rewritten fluently; the
specific asks are randomly honoured or not. Whether they are honoured depends
entirely on whether `evidence_base.writer_block` happens to say the same thing
— which is why the failure is intermittent enough to have survived: briefs
that RESTATE the register appear to work.

Measured cost in one lane, 2026-08-12/13: **four LLM rewrite rounds** burned
trying to get six service names into one FAQ answer, three of them spent
"fixing" real-but-incidental register defects the rounds appeared to expose.
The guidance never reached any writer prompt in any round.

## 2. The mechanism

`content_guidance` is written into `site_work_items.spec` by four emitters —
`apply_gap_plan_action.go:209` (the platform's own content-gap pipeline),
`tool_content_item.go:170`, `create_tool_component_action.go:448`,
`deploy_tool_action.go:575` — and **read by nothing on the work-item path**.
The only read of the string anywhere (`apply_gap_plan_action.go:178`) takes it
from a gap PLAN, not from an item spec, and only to copy it into the next
spec. `page-build-handler → plan_sections → generate_content` builds the
writer's prompt from `writer_block` + the section's stored content + the
section brief; the item's spec never contributes prose.

## 3. Evidence

- **Grep at HEAD (2026-08-13):** `grep -rn content_guidance --include=*.go
  platform/ internal/` → the four writers above, zero readers. (Spelling
  caveat acknowledged; closed by the probes below.)
- **Independent prior read, 2026-08-02:** `bugs_closed/177` close-out footnote:
  *"written by all four tool emit sites and read by NO handler on the
  work-item path... dead weight even on satisfiable items."* Same conclusion,
  different session, different motivation — and it sank as a footnote.
- **Rendered-prompt probes (the artefact, not the code):** rounds 3 and 4 of
  the webdesign.uk FAQ job, six `page-content-writer` calls
  (`llm_call_log` 1e65e476/9c7735b8/089e055f and f0fd40b4/b45ed6de/df9f43b5):
  six distinctive guidance phrases ("One job", "Change no other answer",
  "grouped by purpose", "Extend the FAQ answer", …) — **absent from every
  prompt**, while `writer_block` content is present verbatim in all of them.
- **The discriminating case:** round 5 moved the same instruction INTO
  `writer_block` and changed nothing else about the item (its guidance says
  only "this field has no reader"). See item `gapfill5_faq` for the outcome.

## 4. Blast radius

Every `content_rewrite` filed with a brief, fleet-wide, forever — including
every `add_to_page` item the platform's own `content-gap-planner` emits (its
LLM-composed `content_guidance` is the entire point of that pipeline). The
gap pipeline "works" to the extent the gap planner's asks coincide with what
the register and the section structure already imply.

## 5. Fix candidates, ordered by what closes the door

1. **Wire the field in**: `plan_sections` (or the generate_content prompt
   assembly) injects `spec.content_guidance` into the section brief for every
   section the item touches. One insertion point; makes four existing writers
   and every historical brief meaningful. Council: it changes what a shared
   pipeline GUARANTEES about items already in flight — submit before/alongside.
2. **Kill the field**: delete it from all four writers, document
   `writer_block` / section data as the only steering channels. Honest, but
   forfeits per-item steering entirely — the gap pipeline loses its purpose.
3. Do nothing but document — rejected by writing this file: a steering wheel
   that is connected only sometimes is worse than either connecting it or
   removing it.

## 6. Verification for any fix

File a `content_rewrite` whose guidance demands a sentinel phrase that appears
NOWHERE in the register or existing copy, on a canary page. Grep the rendered
prompt (`llm_call_log.prompt_rendered`) for the sentinel, then the artefact.
Today both greps return nothing; after fix 1 both must return it. Run the
negative control too: an empty-guidance item must not regress.

## 7. Relations

`bugs_closed/177` (the footnote that saw it first) · `bugs_open/268` §the
four rounds (the incident that measured the cost) · the memory lessons
`writes-the-field-is-not-reads-the-field` and `a-doc-comment-is-not-an-
enforcement-mechanism` (this is both at once: a field whose NAME is the
documentation) · `webdesign_uk_build_service` NOTES 2026-08-09 ("facts[] is
bookkeeping; writer_block is the wire") — the same law, one layer down.


---

## 8. THE DISCRIMINATING CASE RAN — the fix-1 direction is CONFIRMED at the artefact (2026-08-13 22:1xZ)

`gapfill5_faq`: identical item shape, `content_guidance` deliberately inert
("this field has no reader"), the instruction moved into `writer_block` as an
imperative (`SQL_2026-08-13c`). **Outcome: all six names present in the
component and on the served page** (deployed 22:19:40Z; verified cache-busted
the next morning, buttons and terms intact) — after four rounds where the same
ask, carried only by `content_guidance`, produced nothing.

One channel dead, the other alive, same writer, same page, same day. §6's
verification recipe stands for whoever wires the field in; note the item's own
status reads `failed` while the page deployed fine (the spawn→call handshake
race, fourth sighting in this lane — verify at the artefact, not the status).

---

## 9. FIXED IN CODE 2026-08-15 (commit `9a7d23c49`) — NOT LIVE until a chassis roll carries it

**Status: still OPEN.** The bar is fixed AND live; this is Go, and Go is inert
until an image ships. The post-roll checklist at the end of this section is what
closes it.

### The fix is NOT §5's candidate 1, and the reason is a fact this file did not have

§5 was written without knowing that **a working steering channel already exists
under a different spelling, and that it is the fleet convention.** Verified end
to end against live `agent_definitions` on 2026-08-15:

```
site_work_items.spec
  → LoadWorkItemsAction parses spec to a map      (load_work_item_actions.go ~:740)
  → build-dispatch-loop:  "spec": "current_item.spec"      [inside a loop sub_workflow]
  → page-build-handler:   "rewrite_guidance?": "input_data.spec.suggestion"
  → page-content-writer:  {{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT: incorporate
                          this into the content){{end}}
```

So the defect is not "a field with no reader" in the abstract — it is **two
spellings of one channel, one of them wired.** `[MEASURED]` 2026-08-15:

| item_type | rows | guidance | suggestion | **guidance-only** |
|---|---|---|---|---|
| content_rewrite | 231 | 56 | 175 | **56** |
| needs_content_page | 113 | 34 | 56 | **34** |

No row carries both keys, and these two types are the only holders of a
non-empty `content_guidance` fleet-wide; every other item type
(`contrast_failure` 226/226, `needs_design_review` 65, `cta_improvement` 51)
uses `suggestion` exclusively.

That reframes the fix. Injecting guidance into the section brief at
`plan_sections` (§5 candidate 1) would have built a SECOND spelling-specific
reader one layer below a channel that already works — two live-but-different
paths is worse than one dead one — and it would have meant committing into
`plan_sections_action.go` while another session had uncommitted work on its
symbols. It also contradicts that file's own stated design at :1560 ("Briefs
apply at content-write time").

### What shipped

1. **`aliasGuidanceIntoSuggestion`** (`load_work_item_actions.go`, beside
   `setRoutingField`) — fills `spec.suggestion` from `spec.content_guidance`
   when suggestion is absent, at the ONE choke point every dispatched item
   passes through. **In-memory only; the DB row is never written.** Writes at
   most that one key, never over an existing value, never from an empty or
   non-string guidance, and never materialises `""` (absent-vs-empty is
   load-bearing: an optional mapping path that RESOLVES to empty is forwarded,
   a MISSING one is skipped). This is the half that closes the door — it covers
   producers no source scan can see: config-driven `create_work_item` steps and
   operator SQL.
2. **The four emitters now write `suggestion`** (`apply_gap_plan_action.go`,
   `tool_content_item.go`, `create_tool_component_action.go`,
   `deploy_tool_action.go`), with a **ratchet test** banning the dead key's
   return. Hygiene, not the door-closer. `apply_gap_plan_action.go:178` still
   reads `content_guidance` **from the gap PLAN** — that is content-gap-planner's
   LLM output contract and is deliberately untouched; the ratchet's own fixture
   asserts it does not convict that read.
3. **The "spec map is NEVER mutated" invariant (`:586`) is NARROWED in the
   commit, not silently broken.** Routing/id keys keep the old rule. The stated
   discriminating test: `suggestion` is a **prose** channel — its only live
   readers (`page-build-handler`'s optional mapping, plus prompt lines in
   `content-gap-planner` and `css-patch-agent`, enumerated by query over every
   active agent row) render it into a prompt, and none gates scope, routing or
   any branch. `component_id` did gate scope, which is why backfilling THAT key
   is still forbidden and `TestSetRoutingField_NeverMutatesSpec` still guards it.

Registered as **WDS-016** in `docs026_concept_register/register/work-dispatch.md`.
Council: `Council-Submitted: b24608e8-4fb1-4028-9512-86af2ef788b7`.

### No opt-in switch, and the measurement is the reason

The 2026-08-02 §2 ruling asks for new authority on a shared seam to ship
default-OFF. `[MEASURED]` the loader dispatches only `status IN
('triaged','approved')` (`:651`), and the 90 guidance-only rows are **all
terminal or parked** — complete 62, needs_human_review 14, failed 11, cancelled
2, wont_fix 1, **zero dispatchable**. Nothing in flight changes on roll day. A
default-OFF switch here would have re-created the dead channel behind a flag.
What *does* change: any of the 25 failed/needs_human_review rows an operator
re-triages will act on its brief for the first time — which is what re-triaging
one means.

### Tests are mutation-proven, not merely green

`load_work_items_guidance_alias_test.go`. Both mutations run this session:
deleting the **call site** fails ONLY the end-to-end test (a helper with no
callers looks exactly like a finished refactor, so the unit tests cannot be the
whole guard); removing the **precedence check** fails the never-overwrite and
non-string tests.

### POST-ROLL CHECKLIST (whoever sees the next chassis roll)

1. Confirm the stamp **per service**, not per fleet:
   `kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'`,
   then `git merge-base --is-ancestor 9a7d23c49 <stamp>`. An empty grep means
   "scrolled", not "unstamped" — fall back to the `/proc/1/exe` probe with a
   present-AND-absent control pair.
2. **Baseline first** (this is what makes the post-fix result disconfirmable):
   `SELECT count(*) FROM llm_call_log WHERE agent_type='page-content-writer' AND prompt_rendered LIKE '%<sentinel>%';`
   → must be 0 before the canary runs.
3. File a canary `content_rewrite` on a canary page with the brief in the
   **DEAD spelling only** (`content_guidance`, no `suggestion`) and a sentinel
   phrase that greps zero in the register and existing copy. This discriminates
   the ALIAS specifically, not the emitter rename.
4. After it runs: the sentinel appears in `llm_call_log.prompt_rendered`
   alongside `## Rewrite Guidance`, and then in the served page (cache-busted).
   Per §8, expect the item's own status may read `failed` while the page
   deployed fine — verify at the artefact, not the status.
5. **Negative control in the same window:** an item with neither key must gain
   no `## Rewrite Guidance` heading.
6. Then: this file moves to `bugs_closed/`, and WDS-016's status goes
   built → deployed.

### ⚠ THE 2026-08-15 EVENING ROLL DID **NOT** CARRY THIS FIX — measured, not assumed

A fresh chassis build **was** deployed on 2026-08-15 and it is **v1.0.1303**,
which **predates this fix by about two hours**. Do not read "a roll happened" as
"the fix shipped" — that is the fleet's oldest landmine and this is a live
instance of it.

The measurement (2026-08-15, taken because the startup `build provenance` line
had already scrolled out of `--tail=400` on this busy service, which means "not
in range", never "unstamped"):

```
commit 9a7d23c49518dcfaaf42854416674f5353024ffb   2026-08-15T21:42:42+01:00  (20:42:42 UTC)
agent-chassis-584b6fcf-9mtqd  started=2026-08-15T18:45:33Z  image=…agent-chassis:v1.0.1303
agent-chassis-584b6fcf-gz5bt  started=2026-08-15T18:45:58Z  image=…agent-chassis:v1.0.1303
deploy/agent-chassis .spec…image = v1.0.1303 · "successfully rolled out" (nothing pending)
```

Pods that started **1h57m before the commit existed** cannot contain it, and the
deployment is settled rather than mid-rollout, so no newer image is on its way.
`IMAGE_TAG` must be bumped and a **new** build cut (v1.0.1304+) from a HEAD that
contains `9a7d23c49`; releases are whole-fleet and the owner runs `make release`.

**So the §9 POST-ROLL CHECKLIST above is still pending in full** — it applies to
the NEXT roll, not this one. Re-establish the stamp first:
`git merge-base --is-ancestor 9a7d23c49 <stamp>` per service. If the provenance
line has scrolled again, probe the binary for the known sha with a
present-AND-absent control pair — never a discovery grep for "some 40-hex
string", which matches Go's internal digit table and returns the same wrong
answer on every service.

---

## 10. ✅ VERIFIED LIVE 2026-08-16 on chassis **v1.0.1304** — CLOSING

**Fixed AND live**, both halves measured at the artefact rather than inferred.

### The binary carries it (symbol probe, both controls in one breath)

The `build provenance` startup line was gone from retained logs — absent even
from `--since-time=<pod start>`, i.e. rotated, not missing (the chassis does emit
it, `cmd/agent-chassis/main.go:53`). So:

| probe | result |
|---|---|
| `aliasGuidanceIntoSuggestion` (subject) | **PRESENT** on both pods |
| `setRoutingField` (positive control, weeks old) | PRESENT |
| `aliasGuidanceIntoSuggestionZZZ` (negative control) | ABSENT |

> ⚠ **A discovery grep gave the OPPOSITE answer first.** Grepping the logs for
> `(commit|git_sha|revision)…[0-9a-f]{7,40}` returned `commit a85ad401`, and
> `git merge-base` then said the fix was NOT in the build. `a85ad401` is the
> **code-index snapshot** (2026-08-12) quoted inside the landmine-verifier's
> verdict prose, which the chassis logs as CONTENT. Anchor a provenance read to
> the emitting line (`"msg":"build provenance"`), never to any hex-looking token
> in the stream.

### The canary — a brief that could ONLY have arrived via the alias

`pool-energy-utilities.internal` (unserved internal pool site, 0 deployed pages,
quiet 6 days), page `faq`, spec carrying **`content_guidance` and no
`suggestion`** (asserted at insert), sentinel `heliotrope kettledrum`.
Baseline: 1 writer call in the prior 3h, **0** sentinel hits.

| measure | positive canary | negative control |
|---|---|---|
| page-content-writer calls | 2 | 2 |
| prompts with the sentinel | **2 / 2** | 0 / 2 |
| prompts with `## Rewrite Guidance` | **2 / 2** | **0 / 2** |
| stored `page_components` with the phrase | **2** | — |
| item final status | `complete`, no error | — |

The control (same site, page `about`, **neither** key) proves the heading is not
unconditional — it appears only when a brief is present. Control prompts are
identified by their own opening line (`… section of About Us | Pool Energy
Utilities`), not by a time window, because 25 guidance-carrying items were
re-triaged into that same window minutes later and would have contaminated a
window-scoped count.

### The 90 historical rows

All 25 non-terminal ones (10 `failed`, 15 `needs_human_review`) were
**re-triaged 2026-08-16 at the owner's explicit instruction**, after the fix was
proven — each row's `error` records its prior status and why. The 9 `failed`
rows had died at `deploy_page` with the spawn→call handshake error, so their
pages very likely deployed while the item read `failed`; they are being re-run
now **with their briefs actually reaching the writer** for the first time.
Pre-state for reversal: `bugfix_271_content_guidance/RETRIAGE_2026-08-16_pre_state.psv`.
The remaining 65 are `complete`/`cancelled`/`wont_fix` and were left alone.

> **Owner decision, recorded because it overrides a gate:** flipping the 15
> `needs_human_review` rows bypasses a human-review gate, several parked since
> April, and touches nine live sites owned by other lanes (6 of the 10 failures
> are webdesign.uk/.co.uk). That was put to the owner with those three risks
> named and the answer was "all 25". One re-triaged row is another lane's own
> canary (`bugs_open/268`'s `edit_live` proof on dartsonline) — that lane should
> know its canary has been re-run.

**Register `WDS-016` promoted `built` → `deployed`.** This file moves to
`bugs_closed/`.
