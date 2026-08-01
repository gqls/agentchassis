# PLAN — bugs_open/170, the style-collection chrome pin has no eligibility check

**Opened:** 2026-08-01. **Predecessors:** `bugs_closed/118` (assignment predicate),
`bugs_closed/166` (the repair), `bugs_closed/167` (the build-path pool lookup).
All three are closed and live on `v1.0.1225`.

## What 170 was filed as, and what it actually is

Filed (2026-07-31, by the 167 lane) as: *the pin is dereferenced by id with no
predicate, so three deployed sites render a deactivated header.* That is true and
it re-verified on 2026-08-01. But the filing describes the pin as a **read** path
only — "a fourth path that 118's census did not include" — and the census framing
is what hid the more serious half.

**The pin is also a WRITE source, and it is the one that matters.**
`LinkSiteComponentsAction` (`link_site_components_action.go:79-122`) reads
`scol.header_component_id` / `scol.footer_component_id` and calls
`ResolveChromeComponent` — the eligible-only pool lookup that 118 installed —
**only when the pin is NULL**:

```go
if !headerCompID.Valid {                        // ← the ONLY gate
    if comp, eligible, err := ResolveChromeComponent(...); err == nil && eligible {
```

A pin that is present is used unconditionally and handed to `relinkSiteComponent`
(`site_component_lock_guard.go:162-181`), which upserts it into
`site_components.component_id` **and** sets `rendered_html = NULL,
build_status = 'pending'`.

So the ordering is: **118 gave chrome one predicate; the pin bypasses it; 166's
repair writes the correct component into the same column the pin will overwrite.**

### Measured 2026-08-01 — the two stores now disagree, in writing

| domain | `site_components` assignment (after 166) | `style_collections` pin | pin `is_active` |
|---|---|---|---|
| ai-agent-orchestration.com | `header-theme-chrome` / `footer-theme-chrome` | `header-professional-dark` / `footer-4-column` | **false / false** |
| finetuning.uk | `header-theme-chrome` / `footer-theme-chrome` | `header-professional-dark` / `footer-4-column` | **false / false** |
| gaswholesalers.com | `header-theme-chrome` / `footer-theme-chrome` | `header-professional-dark` / `footer-4-column` | **false / false** |
| leopardessconsulting.co.uk | `header-leopardess` (own fork, correct) / `footer-theme-chrome` | `header-leopardess` / `footer-4-column` | true / **false** |

Every assignment in that table is correct and every pin but one is not. The
assignments were repaired on 2026-07-31 (`site_components.updated_at`); the pins
were not touched, because nothing in the platform has ever looked at them.

**The footer half is wider than the filing records.** 170 says "three deployed
sites"; that is the header count. Four collections pin `footer-4-column`
(`is_active=false`), *including* leopardess, whose header pin is the one
legitimate row. So: 3 sites on a dead header, **4** on a dead footer.

### Is the revert live or latent? — latent, and stated as such

`site-component-linker` is the only live agent carrying `link_site_components`
(one row, `is_active=true`). It is the wired `HandlerAgent` for two discovery
checks (`check_component_standards.go:114,316`) and for `header_footer` audit
findings (`write_audit_findings_action.go:96`), so it is dispatchable today.
**[MEASURED]** No run appears in `orchestration_states`, whose whole retention is
2026-07-13 → 2026-08-01 (3,309 rows) — so it has not run in 19 days. The claim is
therefore **"armed and revertible", not "reverting now"**, and the fix is worth
doing because the repair that shipped on `v1.0.1225` is not durable while the pin
stands, not because pages are changing under us this week.

## The decision, and why it is not the owner call 170 was waiting for

170 held candidate 1 back because it "changes served markup on three live sites"
and wanted an owner ruling. Re-reading it against what 166 actually shipped, that
framing does not survive:

- The destination is not a new choice. Guarding the pin makes those sites fall
  through to `ResolveChromeComponent`, which returns `header-theme-chrome` /
  `footer-theme-chrome` — **exactly the components 166 already moved the same
  sites' assignments to, with council approval, and which are live today.**
- So the change does not decide anything new about how those sites look. It makes
  the pin path agree with the assignment path, which is already the answer.

That is a materially different question from "shall we restyle three clients'
sites", and it is the reason this lane implements candidate 1 rather than
stopping at the flag-only 1b. **Recorded as a judgement, not a licence:** the
markup does change on the next build, and that is stated in the council
submission rather than buried.

## The predicate — reused, not invented

The 167 lane wrote `chromePinEligibleSQL` in council round 1 and deleted it in
round 3 along with the reporter it fed (`2605d3f92`). The *reporter* was rejected
by four seats; **the predicate was never the objection**, and it came with a live
measurement over all four pins that is the reason it is right:

```
chromePinEligibleSQL = is_active AND component_level IN ('site','header','footer','head')
```

It deliberately **omits `forked_from IS NULL`**, which `chromeEligibleSQL` (the
POOL predicate) carries. That asymmetry is load-bearing and measured: a fork is
illegitimate as a library *default* (118's `header-leopardess` finding) and is the
entire point of a *pin*. Copying the pool predicate would flag the single correct
pin in the fleet and is the exact mistake the 167 lane's table caught.

`architecture` (round 2) objected to that predicate as "a **second** bespoke
eligibility predicate". The answer this lane gives — and must give to the council —
is that a predicate feeding one `zap.Logger` call on one path is bespoke, whereas
**one named predicate that every pin consumer dereferences through** is the
opposite: it is what turns two hand-written pin reads into one. The test for
"bespoke" is how many callers it serves, and this one serves all of them.

## The fix, in four edits

1. **`chromePinEligibleSQL`** — restore the 167 lane's predicate, named and
   commented with the fork asymmetry.
2. **`GetChromePinComponent`** — the ONE way to dereference a chrome pin. Returns
   `(component, eligible, error)`, matching `ResolveChromeComponent`'s shape so
   the two read alike at a call site. Not a change to `GetComponentByID`: that is
   a general by-id fetch used by `RenderComponentAction` for arbitrary page
   sections (`v3_site_actions.go:1722`), where a chrome predicate would be wrong.
3. **Both consumers route through it.**
   - `RenderHeader` / `RenderFooter`: ineligible pin ⇒ fall through to the
     existing `ResolveChromeComponent` branch (167's fixed pool path).
   - `link_site_components`: ineligible pin ⇒ same fall-through the NULL case
     already takes. This is the edit that stops the pin re-breaking 118 and 166.
4. **`deactivated_site_components` sees pins** (170's candidate 1b) so the
   condition is durable and queryable rather than only prevented.

Edits 1–3 make the bad state unrepresentable; edit 4 makes an existing bad row
visible. Both are wanted: prevention does not repair the four rows already there.

### Ordering, and the ownership hazard on edit 4

`discovery_checks/` is the checker-layer lane's subsystem (`bugs_open/149`) and
`verifier_coverage_test.go` was dirty in another session's tree on 2026-07-31.
Edit 4 is therefore **committed separately from edits 1–3** so that a conflict
there cannot hold up the prevention half.

## Decisions taken

- **Retirement vs eligibility.** 166's `repointRetiredChromeSlot` deliberately
  asks `NOT cc.is_active` and is tested against reusing `chromeEligibleSQL`
  (`chrome_selection_test.go:547`). This lane does **not** touch that function and
  does not widen it. The pin predicate adds `component_level` on top of
  `is_active` because a pin, unlike an assignment repoint, chooses what may be
  *served* on every page of every site on the collection.
- **`header_home_component_id` is left alone.** Third pin column on
  `style_collections`, **[MEASURED]** 0 of 14 collections populate it and no Go
  consumer reads it (the `StyleCollection` struct does not even model it). Noted
  as a dormant seam, not fixed — fixing an unread column is inventing a consumer.
- **No render-path reporting.** 167 round 2 settled this: the render path cannot
  repair, has no reader, and fires unboundedly. Detection belongs in the discovery
  check. This lane does not relitigate it.
