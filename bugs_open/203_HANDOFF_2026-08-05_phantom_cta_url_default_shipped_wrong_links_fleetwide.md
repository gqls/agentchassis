# 203 — an unconditional `/contact.html` default in the shared CTA-rendering helper shipped mismatched links fleet-wide; fixed at source, existing instances not yet cleaned up

**Filed 2026-08-05.** **Status: fix committed and unit-tested; NOT yet built/deployed; NOT
yet council-reviewed; 13 existing live instances are NOT yet cleaned up.** This file covers
all three.

**Escape hatch, per the 2026-07-31 owner ruling:** `090` was not run. Substituting first-hand
verification instead: a code-reading agent traced the exact two call sites with file:line
citations, I read both functions myself and confirmed the mechanism, wrote two tests that
reproduce the live bug's exact preconditions, and confirmed by mutation that reverting the
fix makes both tests fail with the live symptom's exact string (`cta_url = "/contact.html"`).
That is the same standard of evidence `090` would have produced, obtained directly because
the cause was a two-line, greppable default rather than something needing DB-side induction.

## The symptom that started this

`https://dartsonline.com/news/index.html`'s hero rendered:
> "Find Your Setup. Follow the Tour. ... **Read the tungsten percentage guide**"
> — linking to `/contact.html`.

## The mechanism

`platform/orchestration/actions/component_library.go`, two near-identical functions,
`contextToMap` (string-map, the regex-fallback render path) and `contextToInterfaceMap`
(the live path — called from `RenderTemplate`, `render_site_components_action.go:971`, for
every hero/call-to-action/archetype-grid/archetype-combinations/gauntlet-cta/content-block-
about component), each contained:

```go
result["cta_url"] = defaultString(ctx.CTAUrl, "/contact.html")
```

`ctx.CTAUrl` (the struct field) is empty for essentially every section-level component,
because per-section CTA text/links are LLM-authored into `ContentData`, not the struct.
`resolve_internal_links_action.go`'s `setCTAField` (lines 281–299) does the real work: write
a validated `cta_url` into `ContentData` when a real target exists, or leave the field ABSENT
and raise an `unresolved_cta` work item when it doesn't — by design, "gated template renders
no button" (that file's own comment, line 22).

The bug: `ContentData` is merged into `result` **after** the line above, and the merge only
overwrites a key it actually contains. `cta_text` almost always IS in `ContentData` (the LLM
wrote it), so it gets overwritten with the real value. `cta_url` is NOT in `ContentData` when
unresolved, so the earlier fake default survives untouched. The template's own guard —
`{{if and .cta_text .cta_url}}` — is written correctly and would have hidden the button, but
it never sees an empty `cta_url`; it sees the fabricated one, which is truthy.

This exact defect class was already named and fixed once, but only for one caller:
`render_site_components_action.go:851-853` calls the `/contact.html` default "the fossil of
the phantom-CTA bug LNK-007" and applies a "correct-or-absent" guard (LNK-005) — but only
inside `renderAndStoreSiteComponent`, the site-chrome (header/footer) path. `component_library.go`'s
`contextToMap`/`contextToInterfaceMap` — used by every OTHER CTA-bearing component — were
never touched.

## The fix (committed, tested, not yet built into an image)

Both lines changed to `result["cta_url"] = ctx.CTAUrl` (no fallback) — correct-or-absent,
consistent with LNK-005's already-decided principle, just applied where it was missing.
Non-regression: a genuinely resolved `ctx.CTAUrl` still passes through unchanged (nothing
reads the removed default deliberately — see below).

Three new tests, `component_library_cta_url_fallback_test.go`:
- `TestContextToInterfaceMapLeavesCTAUrlAbsentWhenUnresolved` — reproduces the exact live
  precondition (`ctx.CTAUrl=""`, `ContentData={"cta_text": "..."}`), asserts `cta_url==""`.
- `TestContextToMapLeavesCTAUrlAbsentWhenUnresolved` — same, for the fallback path (bugs_open/109:
  the two maps must not diverge on what they default).
- `TestContextToInterfaceMapStillPassesThroughAResolvedCTAUrl` — non-regression.

**Mutation-proven**: reverted both lines to the old default, re-ran — both new tests failed
with `cta_url = "/contact.html"`, the live symptom exactly. Reverted back to the fix; full
`go build ./...`, `go vet ./platform/...` (one pre-existing, unrelated unreachable-code
warning in `load_component_library_actions.go:207` — not touched by this change), and the
entire `platform/orchestration/actions/...` test suite (including `discovery_checks`) pass
clean.

**Not yet done: image rebuild + roll.** Per CLAUDE.md, Go changes are inert until built and
deployed. Whoever picks this up: `make build-agent-chassis`, bump `IMAGE_TAG`, roll, pod-grep
verify (positive control: the new default's absence — grep for the literal `"/contact.html"`
string count should drop by exactly 2 occurrences fleet-wide in the binary; a cleaner positive
control is checking `cta_url` is empty on a fresh render of a page with an intentionally
unresolved CTA, e.g. re-triggering `resolve_internal_links` + a page-content-writer pass
against a test page and reading the result).

## Fleet-wide census of ALREADY-SHIPPED instances (as of 2026-08-05, will drift)

```sql
SELECT s.domain, p.url, pc.content_data->>'cta_text' AS cta_text
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=pc.site_id
WHERE pc.content_data ? 'cta_text' AND NOT (pc.content_data ? 'cta_url')
  AND pc.rendered_html ~ 'href="[^"]*"[^>]*>[^<]*</a>';
```
13 rows, 7 sites — the CTA text and the actual `/contact.html`-pointing anchor, for each:

| site | page | CTA text shown |
|---|---|---|
| robot-hands.com | /index.html | Run MatchMatrix |
| robot-hands.com | /how-to-specify-a-gripper.html | Run MatchMatrix |
| fundamentallyai.com | /index.html | Talk to us |
| fundamentallyai.com | /model-fine-tuning.html | Talk to us |
| fundamentallyai.com | /multi-agent-review-council.html | See the decision record |
| fundamentallyai.com | /guides/tool-model-approach-selector-guide.html | Work through the decision guide |
| fundamentallyai.com | /blog/self-correction-leopardessconsulting.html | Read the audit record |
| ai-agent-orchestration.com | /index.html | Book a Technical Discovery Call |
| leopardessconsulting.co.uk | /index.html | Talk to us about your system |
| leopardessconsulting.co.uk | /who-we-help.html | Score your process |
| dartsonline.com | /news/index.html | Read the tungsten percentage guide |
| gaswholesalers.com | /index.html | Review fleet fuel services |
| finetuning.uk | /guides/tool-ai-data-risk-checker-guide.html | Run the Risk Checker |

**This session fixed the SOURCE, which stops NEW instances. It does not touch these 13** —
this is exactly this codebase's own recorded lesson (`dartsonline_traffic/README_where_we_are.md`,
2026-08-03 entry): "when we fix something that generates files, only future files get fixed,
and nobody ever counts the ones already out there." Each of these 13 needs either a targeted
content correction (set a real `cta_url`, or drop the mismatched CTA) and a page rerender, OR
routing through the existing `unresolved_cta`/`cta_names_unknown_destination` machinery.

**The downstream backstop is partially, not reliably, catching these.** `check_misdirected_cta.go`
(discovery_checks) has a `cta_names_unknown_destination` arm meant to catch exactly this shape
on a deployed-HTML audit pass. Checked live: it HAS fired for 2 of the 13
(`robot-hands.com/index.html` 2026-08-04, `gaswholesalers.com/index.html` 2026-08-03) — both
sitting `needs_human_review`, unresolved, not yet acted on. The other 11 have no matching
work item as of this filing — either the check hasn't swept those sites recently, or (for
`dartsonline.com/news/index.html`, written 2026-08-05) the content postdates any prior sweep.
This backstop is a detector, not a fix, and its own queue isn't being drained either.

## Fix candidates for the 13, ordered by effort

1. **Cheapest, narrowest**: for each row, resolve the real target from the CTA text (most are
   obvious — "Run MatchMatrix" is almost certainly the tool page, "Talk to us" the contact
   page, which may even make `/contact.html` CORRECT for some of these by coincidence — check
   per-row, don't assume all 13 are wrong) and either set a real `cta_url` in `content_data` or
   remove the CTA fields entirely, then `page-rerender` (assemble-only, matching the pattern in
   `single-page-deploy-bypasses-stalled-queue` memory / `049b_deploy_single_page.sh`).
2. **Systematic**: drain the existing `cta_names_unknown_destination`/`unresolved_cta` queues
   fleet-wide (currently sitting idle at `needs_human_review`) rather than hand-fixing row by
   row — would also catch instances outside this exact 13-row signature.
3. **Detection improvement**: `check_misdirected_cta` clearly isn't running frequently/broadly
   enough to have caught 11 of 13 before this filing — worth checking its actual schedule
   coverage, separately from this bug.

## Taken up 2026-08-06 (session "bugfix 206" — renamed in intent to 203)

A follow-on session has taken the remaining work: verdict read for the source fix's council
round (`Council-Submitted: 42eda9a5` on `880a405a6`), live-image verification, census re-run,
cleanup of the shipped instances through the framework (resolver + rerender, not hand-set
URLs), detector-coverage measurement, and the `primary_cta_url`/`secondary_cta_url` defaults
in `contextToMap` (same class, still at HEAD as of this note — under investigation).
Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_203_phantom_cta_cleanup/`.

## Related

- `render_site_components_action.go:782-856` / `drop_dead_url_controls.go` — LNK-005, the
  sibling fix this one extends from chrome to every other CTA-bearing component.
- `resolve_internal_links_action.go` — the resolver whose correct "leave unresolved" behaviour
  this bug was quietly defeating.
- `platform/orchestration/actions/discovery_checks/check_misdirected_cta.go` — the downstream
  backstop, evidenced above as under-running, not this bug's fix.
- `dartsonline_traffic/README_where_we_are.md` (2026-08-03 entry) — the "only future files get
  fixed" lesson this bug is a fresh instance of.
