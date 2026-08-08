# NOTES — 117 chrome staleness reference

Append-only, newest at the bottom.

---

## 2026-08-07 — picking the bug, and the ownership check

Surveyed `bugs_open/` (49 files). Cross-referenced every open bug number against
the 27 session transcripts modified in the previous 5 hours. Eight open bugs had
no mention in any live session: 033, 107, 113, 116, 117, 126, 161, 181.

Ran `scripts/who-owns.py` on all eight. Six were disqualified on reading:

| bug | why not |
|---|---|
| 033 | drain ran 07-27; residual blocked on open owner decisions D2/D3 |
| 107 | commit `1a6a1b7e7` says "the bug is owned by vigilant_designer Phase 4" |
| 113 | fix live (`3096a55a6`); owning lane is `dartsonline_traffic` |
| 116 | owner ruled option 4 on 08-06 (`31f423d56`) — owner-gated |
| 126 | CLOSED + LIVE v1.0.1254; back in `bugs_open/` by owner direction |
| 181 | CLOSED + LIVE v1.0.1259; back in `bugs_open/` by owner direction |

**Note the trap in the last two:** `bugs_open/` now contains finished bugs by
owner ruling (2026-08-06), so *being in `bugs_open/` is no longer evidence a bug
is unfixed.* Two of my eight candidates were already done. Read the file.

117 survived: structural, unowned, last touched 2026-07-27.

## 2026-08-07 — the bug is still live, verified at the code and at the artefact

Code path unchanged from what 117 describes:
- `getSiteComponents` at `rerender_single_page_action.go:662` still does
  `SELECT slot_name, rendered_html FROM site_components WHERE site_id=$1 AND
  rendered_html IS NOT NULL AND rendered_html != ''`.
- `assemblePage` (same file, :352) still assembles `head + header + sections +
  footer` from it.
- Re-confirmed after the 2026-08-08 chassis build: both still present.

Artefact check on relojistas.com (117's own verification recipe):
```
$ curl -s https://relojistas.com/index.html | grep -o '<h4>[^<]*</h4>'
<h4>Quick Links</h4>
<h4>Explore</h4>
```
Current `footer-theme-chrome` template yields `Quick Links / Explore / Contact`,
with Contact gated on `{{if or .email .phone}}`. relojistas has neither, so the
served page matches the **fixed** state. **relojistas was repaired** — as 117
itself says it was mitigated on that site. The general gap is what remains.

## 2026-08-07 — the measurement 117 asked for, and what it found

117 fix candidate 2 carries `[UNMEASURED] fleet-wide — run it before designing
anything`. Ran it. `site_components` holds 57 rows: 19 sites × {head, header,
footer}, all 57 with non-empty `rendered_html`, 3 with `component_id IS NULL`
(all three of loanandmortgagecalculator.co.uk — no provenance at all).

Then the decisive query — the existing detector's condition cross-tabbed against
"is the assigned `content_components` row newer than the stored chrome", over
the 53 rows with a non-null `component_id`:

| truly_stale | detector_fires | rows | reading |
|---|---|---|---|
| t | t | 3 | fires, but for an unrelated reason |
| t | **f** | **1** | **false negative — oufe.com/footer, exactly 117's mechanism, unseen** |
| **f** | **t** | **36** | **false positives** |
| f | f | 14 | — |

**36 of the 39 firings are unrelated to chrome drift.** The check is not
dormant — `site_work_items` with `item_key LIKE 'stale\_sc\_%'` shows 7 complete
per slot and 3 detected per slot, most recent 2026-08-06, handler
`rerender-pages`. So it runs, it drains, and it is mostly answering a different
question.

The measurement **could have come out otherwise and did** — all four cells are
populated. A detector that agreed with the real signal would have put 0 in both
off-diagonal cells.

Also measured: **17 of 19 `head` rows point at an `is_active = false`
component.** That half is already covered by the sibling
`deactivated_site_components` check in the same file (extended by
`bugs_open/170` to the `style_collections` pin) — do not duplicate it.

## 2026-08-07 — MISSTEP 1: I used a data-gated block as a version discriminator

To prove the 4 timestamp-stale footers actually serve different HTML, I diffed
the stored footers of lendzy.co.uk (rendered 3 minutes after the 08-02 template
change) and oufe.com (rendered 07-31) and found exactly one distinguishing
token: `class="footer-compliance"`. I was about to use its absence on oufe as
proof that oufe serves the old template.

**It proves nothing.** Reading the template around the string:
```
{{if .compliance_lines}}
<div class="footer-compliance">…
```
It is gated on **site data**, not on template version. lendzy is a loan site with
compliance lines; oufe has none. Both would render identically from either
template version.

**What caught it:** reading the template context around the literal before
trusting it. **The cheap check:** before using any literal as a version
discriminator, read its surrounding template for a conditional gate — a token
inside `{{if …}}` discriminates data, never versions.

## 2026-08-07 — and the honest consequence: "4 stale footers" is a TIMESTAMP claim

Having lost that discriminator, I tested **every** `class="…"` literal in
`footer-theme-chrome` against all 16 footers rendered from it, split by the
08-02 22:33 template-change time. **No class literal splits them.** The 08-02
edit changed no class name.

So I have **not** established that a re-render would change those four footers'
output. `[NOT ESTABLISHED]` — "4 stale footers" means "rendered from an older
template revision", not "serving demonstrably wrong HTML".

This is not a weakness in the case; it is the strongest argument in it. A
timestamp comparison cannot distinguish "stale and it matters" from "stale and
it is a no-op", which is why 36 false positives drain real rebuild work. The fix
should answer **"would a re-render change anything?"** — which is a fingerprint
question, not a timestamp question.

## 2026-08-07 — MISSTEP 2: my own proposed fix reproduced the defect

My first instinct for the corrected reference was
`GREATEST(content_components.updated_at, site_nav_items.updated_at,
sites.updated_at) > site_components.updated_at`. I ran it before writing it up.
**It marks essentially every row stale**, because `sites.updated_at` moves for
any touch of the site row — 6 of the 19 sites had it move within the last 48
hours for reasons having nothing to do with chrome.

A wider timestamp is not a better signal; it is the same defect with a bigger
numerator. **What caught it:** running the candidate predicate as a query before
proposing it. **The cheap check:** any candidate detector predicate gets run
against live data, and if it flags ~100% or ~0%, it is measuring the wrong thing.

## 2026-08-07 — MISSTEP 3: I reported a code change that had not happened

Re-validating after the chassis build, I grepped with a double-quoted pattern
containing `$1`:
```bash
grep -n "UPDATE content_components SET html_template = \$1 WHERE id = \$2" …
```
It returned nothing and I said the line "has changed under me". It had not — the
shell ate the escaping. Re-grepping without the literal `$` found it intact at
`fix_harcoded_colours_action.go:180`.

**The cheap check:** single-quote any grep pattern containing `$`. Exit 0 with
zero matches is the silent-failure shape this estate keeps paying for.

## 2026-08-07 — the second, independent defect: a writer that moves no timestamp

`fixTemplateColors` (`fix_harcoded_colours_action.go:180`):
```go
UPDATE content_components SET html_template = $1 WHERE id = $2
```
No `updated_at`. And its selection query (same file, ~:145-160) explicitly
includes chrome:
```sql
EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id = $1 AND sc.component_id = cc.id)
```
Census of the other writers (`grep -rn 'UPDATE content_components' --include=*.go`,
excluding tests): `fix_component_template_action.go:497,974`,
`fix_nav_link_templates_action.go:166`, `update_component_html_action.go:235`,
`fix_forced_text_colours_action.go:283`, `component_selector.go:135`,
`rename_tool_identity_action.go:91`, `tool_admin_handlers.go:200` — **all set
`updated_at`.** `store_generated_component_action.go:466` sets it as part of a
multi-column update. So `fix_harcoded_colours_action.go:180` is the **single**
outlier, which makes it cheap to fix and easy to regress.

## 2026-08-07 — 090 diagnosis filed and run

Filed per the owner ruling of 2026-07-31 (a `bugs_open/` file asserting a
cross-cutting root cause is not "filed" until it has been through the loop).

- intake correlation `9366c2c5-412e-498c-9431-c45a37dd8411`
- run correlation `0001d9ee-c0ad-4ef2-9304-57e1b4757ec8`
- work item reached `status='complete'` at 2026-08-07 08:54:52Z
- 5 bundle iterations persisted in `diagnosis_artifacts`; final bundle metadata
  `{"truncated": false, "symbol_count": 12, "symbols_in_scope": 13,
  "symbols_unreadable": 1}` — so it read the code rather than failing early, and
  its gathered evidence includes exactly the `site_components` ↔
  `content_components` ↔ `stale_sc_%` join the hypothesis names.

**The verdict itself I could not retrieve.** No `doc_notes` row joins to either
correlation; `diagnosis_artifacts` holds only `kind='bundle'` rows, none
containing a `VERDICT` line. This matches the defect recorded in commit
`0252b3cae` ("the verdict itself was computed then thrown away").
**So: the diagnosis ran, and I am NOT claiming it confirmed anything.**
`[UNVERIFIED]` — the hypothesis stands on my own first-hand measurement above,
which is what the owner ruling's named escape hatch permits, and I am saying so
rather than omitting it.

## 2026-08-08 — where the work stopped

The fable planning agent was cut off by a session limit partway through reading
`render_site_components_action.go`. **No implementation plan exists yet**, and
no code has been changed. Nothing is committed to `platform/`. See
`HANDOFF_2026-08-08_continue_here.md`.
