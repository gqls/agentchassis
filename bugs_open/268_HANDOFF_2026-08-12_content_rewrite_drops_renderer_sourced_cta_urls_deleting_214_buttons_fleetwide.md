# 268 — a `content_rewrite` drops the CTA destination keys, and 214 call-to-action buttons are missing from live pages fleet-wide

**Filed:** 2026-08-12 · **Lane:** `ai_site_selling_automation` · **Severity: high** —
the primary conversion control is absent from 214 components across 19 live
customer-facing sites, and it fails silently in every instrument we have.
**Class:** structural (shared component schema + the regeneration write path).

> **STATUS: OPEN.** webdesign.uk is protected by a site-scoped component lock
> (`SQL_2026-08-12k`), which is a tourniquet and not a fix. The other 18 sites
> are unprotected. The mechanism is NOT established — see §4, including a
> refuted hypothesis of mine.

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
