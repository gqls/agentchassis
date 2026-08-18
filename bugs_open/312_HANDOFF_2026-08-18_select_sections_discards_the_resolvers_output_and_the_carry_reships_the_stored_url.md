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
