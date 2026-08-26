# NOTES — bugs_open/404, a reason vocabulary whose readers disagree

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-26 — lane opened; ownership checked, defect re-verified at both readers

### Why this bug, and what the sweep ruled out

Third bug of this session, after `bugs_open/359` and `bugs_open/407`. The ownership sweep ran
over 39 candidate numbers and found only **three** with no ACTIVE owning workstream — 338, 356
and 404. Recorded so the next sweep does not re-walk them:

| candidate | why not |
|---|---|
| `338` | voice-gate density rules on a single sentence — genuinely open and unowned; **a real candidate, left for next** |
| `356` | fixed in the tree and awaiting a roll; its remaining work is **17 separate routing gaps, each needing its own judgement** — a programme, not a bug fix |
| `404` | taken |

`scripts/who-owns.py 404` → no ACTIVE lane. Filed 2026-08-25 by the `bugs_open/384` lane, which
found it while ANSWERING a council objection — the `prior_art_librarian` seat asked why migration
615 hand-rolled a fan-out instead of reusing the shared per-page re-render creator, and reading
the shared creator to answer showed that reusing it would have shipped 40 assemble-only items and
no visible change. **The hand-rolled version was right for the wrong reason.**

### The defect, verified at BOTH readers today rather than taken from the file

The live gate, read from `agent_definitions` `[MEASURED 2026-08-26]`:

```
check_rerender_mode.condition:
  input_data.spec.reason == 'image_landed'
   OR ... == 'section_data_resolved'
   OR ... == 'cta_links_stale'
   OR ... == 'template_changed'
   OR ... == 'literal_markdown'
then_step: rerender_sections     else_step: render_page     <- assemble
```

The Go reader, `create_rerender_items_action.go:~219-235`, verbatim:

```go
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
stampReason := scoped || reason == "cta_links_stale"
keyReason := ""
if stampReason { keyReason = reason }     // otherwise the item is ASSEMBLE-ONLY
```

**Five values in the gate, three in Go.** `template_changed` and `literal_markdown` were added on
the SAME DAY — 2026-08-18, migrations 460 and 473, by different lanes — and neither touched Go.
So "the next vocabulary addition will repeat this" is not a prediction; it has already happened
twice in parallel within one day.

### The property that makes this dangerous rather than untidy

**Every reader that does not know a value fails toward `assemble`** — which re-ships the stored
HTML verbatim, completes green, and changes nothing. Checked at each reader: the gate's
`else_step` is `render_page`; the Go reader's unknown reason leaves `keyReason` empty so the item
carries no reason at all. A vocabulary whose readers failed toward *re-resolve* would announce
itself — you would get too many re-renders and notice. Failing toward assemble means the estate's
own preferred, safe, cheap mode is also its silent-failure mode.

### ⚠ THE EXPOSURE IS LATENT, AND THE BUG FILE PROVES IT RATHER THAN ASSUMING IT

The file carries **three dated self-corrections**, and the second and third are the instructive
part: the filer's own inference that 471 reason-bearing items were shipping silently was **wrong**,
and they went and checked. Every live producer stamps `spec.reason` in its own INSERT, so the gate
sees it and routes correctly; **not one item ever reached the stale Go reader.** Verified across
live AND archive: of 17,285 `page_rerender` items from that path, 203 carry a reason and all 203
are `section_data_resolved` — which Go knows.

**So this is a trap for the next author, not live damage.** Anyone planning this must quote the
471 as an URGENCY argument (the reason is heavily used via paths that bypass the shared action, so
a future author routing through it is likely) and never as a damage claim.

The third correction is worth reading twice on its own account: the filer's control —
*"6,428 items, 3 carry a reason"* — was a LIVE-WINDOW undercount, because closing a row archives
it out of `site_work_items`. The real figure is 203 of 17,285. And the `bugs_open/410` lane had
already relayed the wrong number into their own rationale, then re-ran the ORIGINAL query
independently, got it to the digit, and recorded that as first-hand confirmation — **they verified
the number by making the same population error, and the exact agreement made it more convincing.**
The lesson recorded there is the one to carry: *re-derive the POPULATION — which tables, which
window — not only the arithmetic over someone else's choice of table.*

### The machinery that already exists for this class

`platform/livespec` is `bugs_open/363`'s answer to exactly this problem: a Go guard that asserts a
property of a live DB object by reading the MIGRATION FILE cannot work, because a migration is
append-only history frozen by its checksum while the live object keeps moving. So livespec is the
**declaration of what a live object should contain, in a file that is allowed to change**, with
both legs live since 2026-08-23 — Go guards compare Go against the declaration, and a daily
auditor (`config-key-audit --live-declaration-drift`, 07:00 UTC) compares the declaration against
the live object through each entry's `ProbeSQL`. `Kind` already includes `workflow`, and
`ClaimedItemTimeoutExclusions` is the worked precedent for a Go list generating the fragment a
declaration asserts.

That is where this fix belongs, and it is why fix candidate 0 (a parity test) and candidate 1 (one
definition) are the same change here rather than two.

### ⚠ Adjacent, not ours

`platform/livespec` is **RED at HEAD** on `TestNoNewMigrationFileReadersOutsideTheAllowList`,
failing on the 405 lane's committed `write_audit_findings_origin_test.go` (`ffa1707b3`). Clean in
the working tree, so it is a committed-HEAD failure. Run this lane's tests by name so that
does not mask the result.

---

## 2026-08-26 (later) — the vocabulary is SIXTEEN values, and the drift is already realised at the OTHER reader

### The census, taken over live AND archive because that is what the file's own §c is about

`spec->>'reason'` on `item_type='page_rerender'`, `site_work_items` UNION `site_work_items_archive`,
`[MEASURED 2026-08-26]`:

```
<none>                          18165   17844 of them via the shared creator — CORRECT; a
                                        site-wide refresh IS supposed to be assemble-only
cta_links_stale                  1905 ✓ in gate
section_data_resolved            1428 ✓ in gate
template_changed                  390 ✓ in gate
verbatim_adoption_deploy           86 ✗ NOT IN GATE
light_palette_chrome_replaced      13 ✗ NOT IN GATE   first seen 2026-08-25 — ONGOING
"migration 415 repointed .article-body__content a…" 11 ✗ free prose
image_landed                        6 ✓ in gate
meta_description_corrected          4 ✗ NOT IN GATE
"the 20:2x rewrite deployed these pages before…"     4 ✗ free prose
"bugs_open/238: the £149 rewrite dropped…"           4 ✗ free prose
legal_page_publish                  3 ✗ NOT IN GATE
"section_edit a007f0ff complete + tool-list removed" 1 ✗ free prose
listing_stale                       1 ✗ NOT IN GATE   first seen 2026-08-24 — ONGOING
m2_rebuild_safety_proof             1 ✗ NOT IN GATE
claims_corrected                    1 ✗ NOT IN GATE
```

### The drift is REALISED at the gate, not only latent at the Go reader

**129 `page_rerender` items carry a reason the live gate does not know. All 129 were handled by
`page-rerender` — the gate's own agent — and 96 COMPLETED.** By the gate's own structure
(`else_step: render_page`) every one took the assemble branch.

The bug file's corrections establish, correctly and by measurement, that **zero** items ever
reached the stale **Go** reader, so that arm is latent. But it bounds the *gate*-side instance at
"7 historical `literal_markdown` items… not chased here". **It is 129, across eleven distinct
reason values, and two of those values first appeared in the last two days.** Same asymmetry,
same silent direction, live and ongoing.

### ⚠ WHAT I COULD NOT ESTABLISH — the discriminator FAILED and I am recording that, not hiding it

I tried to convert *"took the assemble branch"* into *"therefore shipped nothing"*. The
`migration 415` cohort was the strongest candidate because its own reason text says *"this page
still serves the raw rule"* — a checkable claim. Result: 1 of 3 components on **every one of the
11 pages** now carries `--color-primary-ink`, **including the pages whose items were CANCELLED**.

The control and the treatment agree, so the marker arrived by some other route and its presence
is not attributable to these items. **The cohort proves nothing either way.**

So the honest split, and it must travel with the number:

| claim | status |
|---|---|
| 129 items carry a reason the gate does not know | **MEASURED** |
| all 129 went to `page-rerender`; 96 completed | **MEASURED** |
| every one took the assemble branch | **MEASURED**, by the gate's own structure |
| therefore they shipped nothing | **NOT ESTABLISHED** |

⚠ The temptation at this point is to try the next cohort, and the next, until one agrees. That is
how the estate's worst measurements get made. **A cohort is only evidence if a marker can be
ATTRIBUTED to it** — which needs pages that did NOT get the item as a control, and here the
control had the marker too.

### The structural finding, which is the real design input

**`spec.reason` is TWO FIELDS WEARING ONE NAME.** Four of the sixteen values are free prose —
whole sentences, a `£` sign, a bug reference, an operator's note to themselves. Humans are using
`reason` as an ANNOTATION while the gate uses it as a ROUTING KEY.

Three consequences, and I think they decide the fix's shape:

1. **A parity test over "the five" is not enough.** It keeps Go and the gate in step and leaves
   the sixteenth free-text value silently assembling, for ever.
2. **The single definition must also answer "what happens to a reason nobody declared?"** Given
   the fail-toward-assemble asymmetry, the safe answer is probably not "assemble silently". An
   unknown routing key that completes green is this bug in one sentence.
3. **The vocabulary spans at least THREE item types**: `template_changed` also appears on **65
   `section_edit`** items, and `literal_markdown` appears ONLY on `item_type='literal_markdown'`
   items — **never on a `page_rerender` item at all**. So the gate's fifth value may not be
   exercised through this path; check before assuming it is.
