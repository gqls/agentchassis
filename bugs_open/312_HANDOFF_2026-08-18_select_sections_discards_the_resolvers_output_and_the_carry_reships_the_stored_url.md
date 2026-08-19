# 312 — `page-content-writer`'s `select_sections` reads the resolver's output at a path that does not exist, so every fresh build discards link resolution and the carry re-ships the stored URLs

**Filed 2026-08-18** by the `bugfix_299_cta_dials_phone` lane, found while answering
`bugs_open/299`'s producer question ("why does a chassis carrying the 268 fix still emit
this?"). **Class: structural** — a config path on the shared writer workflow, fleet-wide.

**090 substitution, declared per the OWNER RULING of 2026-07-31:** no diagnosis run was
fired. The substitution is first-hand and disconfirmable: the config string and the live
response shape were both read from the running system; ONE orchestration was traced
end-to-end with the discarded value and the used value both present in its own
`collected_data`; and the negative control (does ANY retained run match the configured
path?) could have come out otherwise and came out 0.

## The defect in one line

`select_sections` (an `extract_fields` step on `page-content-writer`) tries
`resolved_links.response.link_resolution.sections_ready` first — but the resolver's
response carries the resolution object **directly** under `response`
(`{unresolved, sections_ready}`, no `link_resolution` level). Path 1 never matches; the
**silent** fallback (`input_data.section_plan.sections_ready`) feeds the render the
**pre-resolver** section plan.

## Evidence (all live, 2026-08-18)

**One traced build, both sides in its own row** — orch `05e3839d-8e18-4935-9c7e-3c6d741665d6`
(page-content-writer, webdesign.uk `index`, 10:25–10:33Z; child resolver
`a907e946-5e43-401a-941b-5ca10cf19ac8`):

| where | `call-to-action.resolved_data` |
|---|---|
| `resolved_links.response.sections_ready` (resolver wrote) | `primary_cta_url` AND `secondary_cta_url` = `/tools/website-brief-starter/index.html`, plus both `*_target_title`s |
| `sections_for_render.sections_ready` (render used) | `primary_cta_url = /contact.html`, `secondary_cta_url = tel:+44 (0) 7934 524 911` |

The used values are the PBP-039/268 carry of the stored row. The saved component
(10:31:54Z) and the served page both carry the tel: — `bugs_open/299`'s button.

**The control (0/150):**

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'link_resolution') AS path1_would_hit,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'sections_ready') AS real_shape
FROM orchestration_states WHERE collected_data ? 'resolved_links';
-- 150 | 0 | 149    (retention window starts 2026-08-17 — claim the window, not "ever")
```

**The live config** (`agent_definitions`, type `page-content-writer`, step
`select_sections`): fields.sections_ready =
`["resolved_links.response.link_resolution.sections_ready",
"input_data.section_plan.sections_ready", "section_plan.sections_ready"]`,
required = `["sections_ready"]`, description — with unintended irony — *"Use
resolver-augmented sections, falling back to the original plan"*.

## Why every instrument is quiet

- The `required` opt-in (from `bugs_closed/192`) checks **presence, not provenance** — the
  fallback resolves, so the step succeeds.
- `ExtractFieldsAction` logs a Warn per missed **target**, not per missed path — a target
  that resolves on path 2 draws no log at all.
- The build completes, the page deploys, and the carried values are yesterday's — which
  LOOKS like "the rewrite kept the links", i.e. exactly what the 268 carry is praised for.

## Consequences

1. The resolver's label-match machinery (bugs 203-follow-on, `f1819861f`/253 ranking) has
   been **inert on every fresh build** in the retained window. Its fixes were real and
   tested — on a path whose output is discarded.
2. `bugs_open/299`'s button survives every rebuild: right answer computed, thrown away,
   carry re-ships the stored `tel:`.
3. `unresolved_cta` items keep filing from the child (the child runs fine) — the queue is
   truthful about the child and silent about the discard.

## ⚠ THE INTERLOCK — do not "just fix the path" (this is the trap for the fixing thread)

The dead wiring is currently an **accidental safety device**. The traced run shows the
resolver's positional pick would have replaced the authored "Get in touch" →
`/contact.html` primary with the brief-starter tool — `setCTAField` has **no keep branch
at all** (`bugs_open/248`'s finding, and their NOTES hold a CONFIRMED production clobber
via the sibling rerender path: finetuning.uk/services, 08-17 19:11Z). Fixing this path
while the binary lacks the keep branches arms that clobber **on every fresh build,
fleet-wide**. Config is live immediately; the keeps are Go.

**Required order:** keep branches (non-page: the 299 lane; utility-page: the 248 lane)
committed → image rolled → pod-verified → THEN a migration correcting path 1 to
`resolved_links.response.sections_ready` (ship it `_HOLD`-named until the verification).

## Fix candidates, ordered by what closes the door

1. **Correct path 1** to `resolved_links.response.sections_ready` (one string in
   `agent_definitions`) — under the interlock above. Fixes the discard.
2. **Make the fallback loud**: `extract_fields` logs (or records) WHICH path satisfied
   each target, so a shape change upstream surfaces as a diff in the log rather than a
   silent downgrade. The structural half — a guard on `storeActionResult` envelopes — is
   already routed to `RFC_012` by the 192 lane; do not fork it here.
3. **A lockstep test** pinning the writer config's path against the resolver's actual
   response shape (the two live in different artefacts and drifted invisibly — this is
   the dedup-index/Go-list lockstep class).

## How to verify

After the interlock is satisfied and the path corrected: a fresh page-content-writer run
where `sections_for_render...resolved_data` **equals** `resolved_links.response...`
for a CTA section (they differed on `05e3839d`), AND the authored `/contact.html` primary
**survived** — the second half is the control that the keeps, not luck, made it safe.

## Pointers

- `bugs_open/299` (motivating case) · `bugs_open/248` (the keep half for page-scheme
  authored links, owned by `bugfix_248_authored_cta_destinations`) · `bugs_closed/192`
  (the fallback chain's previous failure — that fix made the fallback tolerant; this bug
  is the fallback being tolerant of the wrong thing) · `bugs_closed/268` (the carry that
  re-ships the stored values).
- Working docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_299_cta_dials_phone/`.

---

## 2026-08-18 (same day, later) — this is a RECURRENCE of LNK-014, and the register called the shot

Register archaeology (`docs026_concept_register/register/link-management.md`):

- **LNK-014** records this exact defect firing in JUNE, in the OPPOSITE direction: then, the
  `call_agent` envelope nested the reply at `resolved_links.response.link_resolution.sections_ready`
  and `select_sections` read the top-level `resolved_links.sections_ready` → null → silent
  fallback → schema-phantom CTAs shipped. Fixed 2026-06-26 by a one-line jsonb_set repointing
  the config TO `response.link_resolution.…` — the very path that is wrong today.
- **LNK-014's own open follow-up asked for the change that re-broke it:** "the resolver
  returns its whole echoed collected_data with empty final_result (should return a lean
  `{sections_ready, unresolved}`)". The lean return exists now — today's measured response IS
  `{unresolved, sections_ready}` directly under `response` — so [INFERRED, culprit commit not
  yet pinned] the lean-return follow-up landed without the config repoint, and the same
  fallback masked the same class a second time.
- **LNK-013 named the mechanism in advance:** "Designed so resolver failure is byte-identical
  to prior behaviour — which later proved double-edged: the fallback silently masked a path
  mismatch for two weeks." It has now done so twice, in both directions of the same seam.

Consequences for the fix candidates above: candidate 1's target path
(`resolved_links.response.sections_ready`) is confirmed against both the measured shape and
this history; candidates 2 (loud fallback) and 3 (lockstep test binding the config path to
the resolver's actual response shape) are no longer nice-to-haves — a seam that has failed
twice, silently, in both directions, earns its tripwire. The June fix also supplies the
repair mechanics (jsonb_set on the agent_definitions row) and the verification posture.
The INTERLOCK above is unchanged and still binds.

---

## 2026-08-19 — CONFIRMED at fleet scale, with a sharper instrument than this file originally used

Two things happened: a `090` diagnosis run came back **UNVERIFIABLE**, and re-doing its work
by hand produced a much better control than the single trace above.

### The 090 run did not refute this — it could not SEE it

Run correlation `d1434dd5-4c5c-4097-9223-be8aca0dcd69` (intake `a26efb3c`, `FORCE=1` with the
reason recorded in the lane NOTES — the coverage refusal listed 30+ items on
`webdesign.uk/index`, none of them this seam). Two iterations, then
`status=UNVERIFIABLE`, `stopped_by=scope-not-narrowing`, "Hand to a human with the full
trail; do NOT auto-conclude."

Its own account of why: the `data_request` meant to fetch `select_sections` "returned the
plan_sections step instead (truncated before reaching select_sections)"; no fetched
orchestration row carried the two structures side by side ("truncated before any such
structure"); and the `agent_error_log` rows it did get were `validate_content` blockers and a
`save_sections` SECTION SHRINK REFUSED, which it correctly reported as neither confirming nor
refuting. **A truncation in the loop's own evidence fetch, on a hypothesis whose evidence is
large nested jsonb.** Recorded plainly because an UNVERIFIABLE verdict is neither a REFUTED
one (which would be a success) nor a CONFIRMED one, and it would be dishonest to file it as
either. Per the 2026-07-31 owner ruling, first-hand verification was substituted for each of
the three gaps it named:

1. **The live config, re-read** — unchanged from §Evidence above: `select_sections`
   `fields.sections_ready` still leads with
   `resolved_links.response.link_resolution.sections_ready`, `required = ["sections_ready"]`.
2. **The negative control, re-measured on a rolled-forward window** (08-18→08-19): **48** runs
   carry `resolved_links`; **0** match the configured path; **48/48** carry the lean shape.
   The window has moved since the 0/150 measurement and the answer has not.
3. **The same-run comparison** — see below.

### The sharper instrument: `*_target_title` is minted ONLY by the resolver

The original trace compared URLs. That works when the stored value is wrong, and **silently
scores a run as healthy when the carried value happens to already be right**. Of the 48
retained runs, 18 have the two sides byte-identical — a url-diff calls all 18 fine, and it is
not measuring anything on them.

The resolver also writes `*_target_title` companions, and **nothing else on this path mints
them**. So their absence downstream is the discard itself, visible on every run regardless of
whether the urls agree. Worked example, fresh post-roll run `01b5ba83` (2026-08-19 18:33Z,
ai-agent-orchestration.com): resolver wrote `cta_url`, `secondary_cta_url`, `cta_target_title`
("Password Strength Physics") and `secondary_cta_target_title`; `sections_for_render` carried
the two urls and **neither title**. The urls agreed — by coincidence — and the titles did not
survive.

Fleet-wide over the retained window:

| | runs |
|---|---|
| carry both structures | 48 |
| resolver minted `*_target_title` on a CTA section | 26 |
| …titles SURVIVED into `sections_for_render` | **0** |
| …titles DISCARDED | **26** |
| the two sides byte-identical | 18 |
| the two sides DIFFER | 30 |

**26 of 26, no survivors.** Add this to §How to verify: after 477 the titles must appear
downstream. It is a strictly better post-fix assertion than a url match, for the same reason
it is a better pre-fix one — a url match can be satisfied by the carry doing nothing.

⚠ **Measurement trap, paid for once (WRONG_CALLS 2026-08-19):** the first cut of that table
cast the whole `sections_for_render` blob to text and `LIKE`d it, which matched
`_target_title` elsewhere in the structure and returned **31 minted / 31 survived / 0
discarded** — the exact opposite, and the comfortable answer. Anchor the cast to
`jsonb_path_query_array(…,'$.sections_ready[*].resolved_data')` on BOTH sides, never to the
container.

### Interlock status

Both keep halves are now LIVE (chassis v1.0.1316, capability-probed on both pods with a
negative control — `LANDMINES.md` records why commit ancestry was not the check available).
That satisfies 477's stated precondition. **477 is still HELD**, on the canary and the owner's
decision, and `RFC_040` (DRAFT) proposes making this class of interlock mechanical rather than
prose — raised out of the council objection that the enforcement point for this very bug's
ordering constraint is documentation in three places and code in none.
