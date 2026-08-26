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
