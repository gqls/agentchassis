# 268 — a `content_rewrite` drops the CTA destination keys, and 214 call-to-action buttons are missing from live pages fleet-wide

**Filed:** 2026-08-12 · **Lane:** `ai_site_selling_automation` · **Severity: high** —
the primary conversion control is absent from 214 components across 19 live
customer-facing sites, and it fails silently in every instrument we have.
**Class:** structural (shared component schema + the regeneration write path).

> **STATUS: CLOSED 2026-08-14 — fixed AND live AND repaired; see §12.** Fix
> `8f899cc8d` live since `v1.0.1298` (canary-proven on a live regeneration,
> `carried_fields` on the plan items); all 10 history-recoverable rows
> restored and re-rendered; permanence proven by a second rewrite. The ~194
> remaining label-without-URL rows are the `unresolved_cta` never-resolved
> class — a SEPARATE deliverable awaiting an owner decision (§12), not this
> bug. (The banner below is the original filing state, kept for the record.)
> The webdesign.uk lock question (keep or lift, now the fix protects the
> rows) is also with the owner.

---

## 1. The symptom

A component that carries a button LABEL but no button URL renders **no anchor at
all**. Both shared templates gate the anchor on the URL rather than the label:

```
hero            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}" …
call-to-action  {{if and .primary_cta .primary_cta_url}}<a href="{{.primary_cta_url}}" …
```

So the failure produces: no error, no missing prose, no shortened byte count, a
clean claims scan, and a page that looks finished. **The call to action is
simply not there.**

## 2. Fleet exposure, measured 2026-08-12 ~20:45Z

```sql
SELECT s.domain,
       count(*) FILTER (WHERE pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta') AS has_label,
       count(*) FILTER (WHERE (pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta')
                          AND NOT (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')) AS label_no_url
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.slot_name IN ('hero','call-to-action') AND p.status='active'
GROUP BY s.domain ORDER BY 3 DESC;
```

**216 components across 19 sites carry a label with no URL. 214 of them render
zero anchors.** Worst affected: finetuning.uk 31, ai-agent-orchestration.com 29,
gamesdesign.co.uk 21 (of 28), mortgagecalculator.co.uk 21 (of 25),
relojistas.com 17 (of 22), dartsonline.com 15, fundamentallyai.com 15.

The two shared components are `hero` (`23f95f00-…`, 20 sites, 276 instances) and
`call-to-action` (`0197e8d7-…`, 20 sites, 237 instances).

## 3. Reproduction, with a control — this is the load-bearing evidence

A `content_rewrite` (`mode=edit_live`) was dispatched against five webdesign.uk
pages at 20:2xZ. The URL keys were present in `content_data` BEFORE it ran,
because this lane had restored them earlier the same afternoon.

| | before | after |
|---|---|---|
| components carrying `cta_url` / `primary_cta_url` | 7 | **0** |
| site-wide `href="` count | 28 | **13** |

**The control: `contact/hero` was NOT part of the rewrite. It kept its keys
(`cta_url` + `secondary_cta_url`) and both its links.** Every component that WAS
rewritten went to `0|0|0`. Same site, same components, same schema, same run —
the only difference is whether the page was regenerated.

Raw before/after in
`docs024_key_docs_latest/ai_site_selling_automation/` scratch notes and the
NOTES entry of 2026-08-12 (close).

## 4. Mechanism — CANDIDATE ONLY, and one reading already refuted

The fields are declared in `content_components.input_schema` as
`{"type":"url","source":"renderer","required":false,"on_missing":"skip_field"}`.

The candidate reading: `sourceResolver.resolve` short-circuits that source —
`if source == "" || source == "llm" || source == "renderer" || source == "static" { return nil, true }`
— returning **found=true with a nil value**. The field is therefore never
*missing*, `handleMissingField` never runs, and so `carryStored` (the
`bugs_open/238` carry, PBP-039) never runs either: the carry protects fields
that FAIL to resolve, and this class always "succeeds" with nothing.
`plan_sections` then `continue`s on the renderer/static branch writing only a
declared `fallback`, and these declare none. `save_page_sections` replaces
`content_data` wholesale, so the key is gone.

> **⚠ `090` run `97ef39f0-19df-4935-834d-c80514fbc43e` REFUTED this.** Its
> citations are `content_data` rows carrying `"cta_url": "/contact.html"` —
> **the values this lane had restored sixteen minutes before the run started.**
> It measured a repaired system and correctly reported nothing missing. The
> refutation is therefore not decisive either, and the reproduction in §3 (which
> post-dates it, and has a control) is the better evidence. **A re-run is owed,
> authored against `page_component_history` for the 16:37–17:23 window, with the
> symptom stating plainly that the live rows were repaired at 17:23.**
>
> A second reading of mine was refuted outright: I assumed the URL keys were
> **undeclared** in the schema. They are declared, all four, with
> `source: renderer`. A field outside the carry's reach and a field absent from
> the schema look identical from the symptom.

Also note **238 §8 is stale**: it says both fix halves are "inert until the
fleet next rolls". They have rolled — `agent-chassis v1.0.1291` was built from
`da5a7eb8f`, and `git merge-base --is-ancestor d26c26a9a da5a7eb8f` passes with
controls both ways. So the carry is LIVE and this still happened.

## 5. What does NOT fix it

- **Restoring `content_data` alone.** Proven twice on 2026-08-12: a
  `page_rerender` dispatched *after* the keys were restored still rendered no
  buttons.
- **A fallback on the shared schema.** `/contact.html` is wrong for sites with
  no such page; it would ship broken links to 19 other sites.

## 6. Fix candidates, ordered by what closes the door

1. **Take the URL fields out of the renderer short-circuit** so a failed
   resolution reaches `handleMissingField`, and the existing 238 carry protects
   them. Config-only (`content_components.input_schema`), live immediately, no
   roll — but it touches a component used by 20 sites, so it is architecture
   scope and wants the council gate. **Verify the carry actually fires before
   assuming this is sufficient.**
2. **Make `resolve_internal_links` resolve these CTAs**, which is the platform's
   intended mechanism — it already files `unresolved_cta` items saying "no real
   page destination", so it knows it failed. Go change; inert until a roll.
3. **Render-time fallback in the templates** — gate the anchor on the LABEL and
   default the href to the site's contact page. Cheapest, but bakes a
   site-shaped assumption into a shared template.
4. **Component lock** (what webdesign.uk has now) — site-scoped tourniquet.
   Freezes the copy in the locked components; blocked changes are at least
   recorded as `lock_blocked_change` items rather than lost silently.

## 7. How to verify any fix

Diff the invariant as a matched pair, which is the check that caught this:

```sql
SELECT p.name, pc.slot_name,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')) AS links
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND p.status='active' ORDER BY p.name, pc.position;
```

Take it before the rewrite and after. **Five other checks were green throughout
the original incident** — the claims scan, byte deltas, a retired-term grep, the
served-artefact fetch, and a link gate armed on the one page that happened not
to be affected. None of them can see a missing `href`.

## 8. Relations

`bugs_open/238` (the parent key-loss family; §9 there carries this case and §8
of it is stale) · `bugs_open/229` (what a hand-patched `rendered_html` re-arms —
the repair here had to do exactly that) · `bugs_open/058` (the component lock
this leans on) · `bugs_open/178` (`mode=edit_live`, without which the rewrite
also guts the prose) · `WRONG_CALLS.md` 2026-08-12 (why the five green checks
missed it) · `LANDMINES.md`, the `save_page_sections` REPLACES entry.


---

## 10. FIX WORKSTREAM OPENED 2026-08-13

Cold-start for the fleet fix:
`docs/agent_docs/docs024_key_docs_latest/bugfix_268_cta_buttons_fleet/HANDOFF_2026-08-13_start_here.md`
(also in MEMORY_workstreams). Contributions to the evidence go HERE; the fix
thread's own working docs go in that directory. Check `who-owns.py 268` and
live transcripts before routing work at it.

## 11. Fix lane contributions (2026-08-14)

- **Census moved: 216/19 → 217/20** (§2 query re-run 2026-08-13 ~14:1xZ) —
  the leak is active; one more component and one more site since filing.
- **§4's candidate mechanism: right outcome, wrong route** (code-read at HEAD,
  090 pending). `carryStored` never runs for these fields — but not because
  `resolve` returns `(nil, true)`: the renderer/static branch at
  `plan_sections_action.go:2361-2369` `continue`s BEFORE `resolver.resolve`
  is ever called; the `:622-626` short-circuit is a second, redundant guard.
  Had the early branch not existed, `(nil, true)` fails `found && value != nil`
  and falls into `handleMissingField` naturally. And the §6 caution about
  `carryStored`'s guard is inverted: it excludes only `""`/`"llm"` — a
  renderer field REACHING it would be carried. The defect is that it is never
  reached.
- **090 re-run in flight** (authored per §3 of the fleet-fix handoff): intake
  `95df3483-0291-48d3-992f-6453b5e8324f`, run correlation
  `38e53a03-ddcd-46c6-8533-d48510747758`. Fix design + owner's default-ON
  decision recorded in the fleet-fix lane's PLAN/NOTES.

### §11.1 (2026-08-14, later) — the 090 broke instead of ruling; verification substituted first-hand, and the fleet count splits into two classes

**The 090 produced NO verdict either way.** The diagnose-agent's `verdict`
step died on iteration 5 with `response truncated: stop_reason=max_tokens
(output_tokens=32000 reached the configured cap)` — bundles had grown
77KB→141KB over five iterations; `max_attempts=1` so the item is terminally
`failed` (orchestrations `bdf69bc1`/`661a2b55`/`a02b83dd`). The loud failure
is `bugs_closed/076`'s guard working as designed; not re-filed as a new bug on
one occurrence — if a second verdict-step truncation appears, file it then.

**Per the 2026-07-31 owner ruling, the substituted first-hand verification,
stated plainly:**
1. **In-vitro reproduction of the mechanism** — a test fixture declaring a
   `source:"renderer"` url field beside a stored row (the first such fixture
   in the repo): against the UNFIXED code the key is absent from
   `resolved_data` (the early branch `continue`s before `carryStored`);
   against the one-line fix it is carried. Mutation-verified in both
   directions (`plan_sections_renderer_carry_test.go`).
2. **Independent second read** — a separate-context trace with citations:
   the branch predates both `handleMissingField` and the carry
   (`abf1c308a`, 2026-03-09, the file's founding commit), so it is an
   optimisation that became a bypass, not a guard.
3. **History, read with the trigger's semantics** — the archive trigger
   stores OLD (prosrc read), so a `delete` archive is the pre-rewrite state:
   webdesign.uk `index/hero` delete-archives at **16:50:19** and **20:34:58**
   both carry `cta_text,cta_url,hero_url,secondary_cta,secondary_cta_url` —
   keys present going INTO each rewrite, gone from the stored replacement
   (the §3 live measurements), untouched on the control page.

**The fleet count is TWO populations, and §2's census cannot tell them
apart** [MEASURED 2026-08-14, split query in the fleet-fix RUNBOOK]:
of 217 label-without-URL components (20 sites), only **10** ever had a URL in
any archived generation — the regeneration-loss class, all lost 2026-08-11/12
(ai-agent-orchestration news/hero; dartsonline grip-styles + index, both
slots; idea.uk tool-funding-fit + tool-patent-check heroes; vonc.com
archetypes, both slots; webdesign.uk index/call-to-action — lost 08-11 13:43,
BEFORE the repair lane's baseline, so it survived the 08-12 repair and sits
LOCKED). **74** have archived generations none of which ever held a URL, and
**133** have no archived generation at all [INDETERMINATE: first-generation
rows never regenerated, or history orphaned by a page identity change]. So
the honest attribution: the *mechanism* is proven and was actively deleting
resolved URLs (webdesign.uk's 7+1, plus these 10); but **most of the 217 are
the `unresolved_cta` never-had-a-destination class** — `resolve_internal_links`
filed items for them and rendered no anchor, correctly. **The repair
deliverable therefore splits**: ~10 rows recover from history; the ~200
others are a destination-resolution problem (fix candidate 2 territory /
unresolved_cta backlog), NOT a history restore, and no amount of re-rendering
will conjure URLs they never had.

> **CORRECTED 2026-08-14 (repair session):** this section said webdesign.uk
> `index/call-to-action` "sits LOCKED". **It is not locked.** The live lock
> map shows the 08-12 repair locked `index/hero` (a row it repaired) and
> seven other hero/call-to-action rows; `index/call-to-action` was lost
> 08-11 13:43 — before that lane's baseline — so it was in neither their
> repair set nor their lock sweep. Caught by reading `lock_type`/`locked_at`
> on the actual rows before drafting the restore. All 10 recoverable rows
> are unlocked; no unlock step is needed for the repair.
> Split re-measured 2026-08-14 ~16:50Z: **10 ever-held / 73 never-held /
> 134 no-history, of 217** (one never-held row moved to no-history since
> §11.1 — a page identity change, the fleet moving; the 10 are unchanged).

## 12. CLOSED 2026-08-14 — fixed, live, canary-proven, repaired, permanence-proven

- **Fix live:** `8f899cc8d` (carry inside the renderer/static branch), council
  APPROVED round 1 (`e6c1e4eb…`), live since `v1.0.1298`; re-verified on
  `v1.0.1299` (stamp `6f8efa158`, both replicas, binary probe + controls).
- **Canary (the §7 verification, run for real):** `edit_live` rewrite of
  dartsonline.com/beginners — prose rewritten, every url key survived, hrefs
  identical, site-wide invariant diff unchanged, live page redeployed. Route
  discriminated: plan items record `carried_fields` for exactly the CTA
  destination fields; `structural_misses` empty. The same operation pre-fix
  is §3's reproduction, which deleted them.
- **Repair:** all 10 ever-held rows restored from `page_component_history`
  (`SQL_2026-08-14_restore_cta_urls_10_rows.sql`; every target URL verified
  live first) and re-rendered (`reason=section_data_resolved`, 7 pages, 7/7
  complete). Live pages spot-checked serving the restored anchors.
- **Permanence (fix+repair compose):** a SECOND `edit_live` rewrite on the
  freshly repaired dartsonline/index — restored keys survived, carried again
  (`8183390d…`), live page redeployed 18:52:21Z. Repair-then-fix was the
  original trap; fix-then-repair holds.
- **Census after:** 194 label-without-URL / 21 sites (was 217), and the
  ever-held-a-URL bucket is **ZERO** — the regeneration-loss class is empty.
- **The remaining ~194 are NOT this bug** and do not block closure: they are
  the `unresolved_cta` never-had-a-destination class (§11.1). Scoped
  2026-08-14: the `unresolved_cta` queue holds only 71 items across 6 sites
  (28 open `needs_human_review`) against ~194 damaged rows across 21 sites —
  most rows have never even been queued for a destination decision. OWNER
  DECISION PENDING: re-run resolution per site / accept label-only / new
  lane (fleet-fix handoff §3 options). Recorded in the lane's
  `README_where_we_are.md`.
- Two operational notes recorded on the way: both `content_rewrite` runs'
  work items read `failed` on a `deploy_page` RESULT-DELIVERY failure while
  the work succeeded and deployed (contributed to `bugs_open/217`, which
  owns that seam); and the canary/repair items carry SYNTHETIC backdated
  `created_at` values (queue-position, lane NOTES 2026-08-14) — do not read
  those timestamps as filing dates.

### §12 addendum (2026-08-15) — owner rulings executed; the unresolved_cta re-run is done

Owner (2026-08-15, in chat): re-run resolution per site; lift webdesign.uk's
8 emergency locks.

- **Locks lifted**: `ai_site_selling_automation/SQL_2026-08-15_unlock_cta_components.sql`
  — 8 off, verified; sibling chat-input-box lock untouched.
- **Resolution re-run (cta_links_stale re-renders, 126 pages / 21 sites,
  item_keys `ctaresolve_268_%`): census 194 → 11 label-without-URL rows
  (21 → 4 sites), 183 rows resolved.** Canary site dartsonline verified as
  a matched pair first (11/11 resolved, untouched rows byte-identical);
  248-class clobber exposure measured ZERO pre-flight; three bounded
  recompute anomalies contributed into `bugs_open/248`.
- **The 11-row residue, each with its reason** (list with labels in the
  lane NOTES 2026-08-15): 1× aao/services blocked by the claims floor
  (banned claim "70+ agent…" in stored copy — 149 C1's guard, filed its
  own item; needs a copy fix first) · 10× tool/game-page heroes whose
  label names the page they sit on (gamesdesign ×3, mortgagecalculator ×6,
  vetcomparison/index) — a self-target the resolver refuses by design;
  these want an in-page anchor or a copy change, i.e. a content decision,
  not resolution. v1.0.1300 (stamp `a2a691213`) verified carrying the fix
  before any of this ran.

### §12 addendum 2 (2026-08-15, close) — the residue is dispositioned; the lane is finished

Owner rulings on the addendum-1 residue, executed and live-verified:
**D1** aao/services reworded by the framework ("Look through a case study
first" — no count, ban class avoided entirely), claims floor passed, CTAs
then resolved by cta_links_stale. **D2** the 10 self-target heroes split by
evidence: 1 in-page anchor (jelly-invaders `#gameCanvas`), 7 vestigial
labels deleted (proven render-neutral first), 2 KEPT — their label keys are
live tool-UI text (`SQL_2026-08-15_d2_selftarget_heroes.sql`). **D3** the
target-diversity content pass is commissioned as its own lane
(`cta_target_content_pass/`). Residual census rows belonging to THIS bug's
mechanism: **zero** (the label-without-url census remains a live fleet
number that counts other causes, including other lanes' mid-build pages —
split by history before reading it, per 016b §9).

---

## CONTRIBUTION from the `bugfix_271` lane, 2026-08-16 — YOUR CANARY HAS BEEN RE-RUN

Not a finding about this bug; a courtesy so you do not discover it as an anomaly.

`bugs_closed/271` (`spec.content_guidance` had no reader — a rewrite brief written
under that key never reached any writer prompt) shipped its fix in chassis
**v1.0.1304** and, at the owner's explicit instruction, **all 25 non-terminal work
items whose brief lived only in that dead key were re-triaged** on 2026-08-16.

**One of the 25 is this lane's own canary** — the `edit_live` proof on
dartsonline (`CANARY bugs_open/268: edit_live rewrite of beginners — proves
renderer…`), which had been sitting in `failed`. It is now back in `triaged` and
will re-run.

Two things that matter for how you read its result:

1. **It failed at `deploy_page` with "workflow completed but its result could not
   be delivered to the parent"** — the spawn→call handshake race, not a content
   failure. Its page had most likely deployed fine while the item recorded
   `failed`, so treat the old `failed` status as unreliable evidence about what
   your canary actually proved.
2. **Its re-run is NOT the same experiment as the original.** Before v1.0.1304 the
   item's `content_guidance` reached nothing, so the rewrite was steered by
   `writer_block` and the existing page alone. From now on that brief DOES reach
   the writer prompt (under `## Rewrite Guidance`). If the re-run behaves
   differently from your recorded baseline, the changed variable is the brief
   arriving — not the renderer/CTA mechanism this bug is about.

Each re-triaged row's `error` column records its prior status and the reason.
Pre-state for reversal, if you want your canary put back exactly as it was:
`docs/agent_docs/docs024_key_docs_latest/bugfix_271_content_guidance/RETRIAGE_2026-08-16_pre_state.psv`.
