# 084 — Owned/ported pages get no browser verification, and nothing asserts a script URL actually loads

**Filed** 2026-07-26. **Substantially rewritten the same day — see the correction
box.** Class: coverage gap in an existing tier (not a missing capability).

---

> ## CORRECTED 2026-07-26, before anyone acted on this
>
> **The first version of this file was wrong, and wrong in the most misleading
> way a bug report can be: it asserted a universal negative without grepping.**
> It said —
>
> > *"There is no point in the pipeline where 'this page's JavaScript works' is
> > asserted. Every check we have is a check of presence… and none is a check of
> > integrity."*
>
> That is false. The platform has **a whole verification ladder built for exactly
> this**, and its top tier drives the deployed page in **real headless Chromium**:
>
> - `internal/adapters/browserrunner/run_checks_action.go` launches Chromium via
>   `playwright-go`, waits for JS to render, performs real `fill`/`click`/`select`
>   steps, asserts elements exist and match after interaction, and captures
>   `console.error` **and uncaught page errors** through
>   `page.OnConsole` / `page.OnPageError`. Desktop *and* mobile emulation
>   profiles. Screenshots of failures to B2.
> - It is **live in production at `v1.0.1167`**
>   (`deployments/kustomize/services/browser-runner-adapter/overlays/production/uk_001/`).
> - It is made continuous by the `tool_acceptance_due` discovery check, and the
>   ladder is documented in
>   `docs/agent_docs/docs024_key_docs_latest/travelling_docs/OVERVIEW_self_verifying_tools.md`
>   — whose Tier-4 row reads *"Does it actually **work** in a browser?"*
>
> There is also a live `dead_controls` detector
> (`discovery_checks/check_dead_controls.go`), a `truncated_component` check that
> catches unterminated `<script>` blocks, a `tool_health` check that fails a tool
> with no `<script>` at all, and a render-time `chrome_dead_control` guard.
>
> **What caught it:** the owner asked "consult the docs for where we have
> discussed and built mechanisms to check whether the JS functions as expected —
> how does this compare to your diagnosis?" One question; the whole premise gone.
>
> **The cheap check I skipped** is the one CLAUDE.md already mandates — *"Grep
> before you file. `/bugs_open/` and the workstream dirs first"*. A grep for
> `dead_control`, `browser-runner` or `acceptance` would have found the ladder in
> seconds. I had even *read* the memory line naming the dead-controls detector as
> live, and filed anyway.
>
> **Why the error was easy to make and is worth guarding against:** I generalised
> from *my* population (owned, ported, non-tool pages) to *the platform*. Those
> pages genuinely have no browser coverage — but that is a **coverage boundary**,
> not an absent capability, and the two call for completely different work. The
> first version would have sent someone to build a headless-browser tier that has
> been running in production for weeks.
>
> Logged in `WRONG_CALLS.md`.

---

## The real, narrowed finding

A page can publish, render, report `complete`, and be completely dead — every
control inert — **if it falls outside Tier 4's coverage boundary**, which most
pages do.

### 1. Tier 4 only ever runs against `component_level = 'tool'` pages with declared criteria

`discovery_checks/check_tool_acceptance_due.go:51` — `AND cc.component_level = 'tool'`.
And `request_browser_run` no-op-skips when the PLAN carries no ` ```criteria` `
fence.

So the browser never visits:

- **owned pages** (`rebuild_policy='owned'`) — the entire webdesign.co.uk port,
  97 pages, ~63 of them interactive tools whose behaviour is inline JS;
- **chrome** (`site_components`) — including any search widget;
- **ordinary sections**;
- **`js_snippets` bundles**.

That population is not marginal. It is most of what the fleet serves.

### 2. Nothing fetches a `<script src>` and asserts it returns 200

The closest existing check is Tier 2's `asset_loads`, and it stops one step short
— `check_tool_acceptance.go:375`:

```go
} else if strings.Contains(html, ch.Path) {
    pass(ch.ID, "asset path referenced: "+ch.Path)
```

**String presence in the HTML, never a request.** A `<script src>` pointing at a
404 passes this check. It would surface at Tier 4 only *indirectly*, as a console
error, and only for a tool that declares `no_console_errors`.

This is exactly the shape of `bugs_closed/041` (chrome `js_content` published to
a path nothing loaded) and of the gauntlet misstep recorded in
`gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md:105` — *"the HTML is right +
status COMPLETED is NOT proof the tool works"* when behaviour lives in a JS asset.

### 3. Nothing detects a handler attribute with no backing script — on our own sites

`extract_interactive_fingerprint_action.go` parses `onclick`/`oninput`/
`addEventListener` and `<script src>` — but it runs over **crawled third-party
HTML during adoption**, produces a brief for an LLM, and makes no judgement. Its
own note says so: *"regex check on our deployed tool output, not on crawled
source. Different direction."*

`check_dead_controls.go` is explicit that this is out of its scope:

> *"`<button>` with no handler is NOT judged statically (JS binds at runtime); the
> post-hydration equivalent lives in the Tier-4 browser tier."*

…and the Tier-4 replacement is **not built** — it is T5.1 in
`experience_loop/RUNBOOK_experience_loop.md:343`, *"Post-hydration dead-control
assertion (closing T2.3's Tier-4 gap)"*.

So the `<button>`-with-no-handler class is currently owned by nobody: descoped at
Tier 2 by design, and unbuilt at Tier 4.

### 4. No script-parity gate at any generic write site

Any writer that transforms existing HTML can drop scripts silently. The one that
did (`cmd/webdesignport`) now has `checkScriptParity`, but that guard lives in
that tool alone.

---

## Root cause, correctly stated

Not "we have no integrity checks" — we have a good one. It is that
**behavioural verification is opt-in and narrowly scoped**: a page is browser-
tested only if it is a `tool`-level component *and* someone wrote acceptance
criteria for it. Everything else is verified by presence alone, and JavaScript is
uniquely exposed to presence-only checking because its absence changes nothing
visible until a human clicks something.

---

## Fix candidates (all extend existing machinery — none is new)

**1. Make `asset_loads` actually load.** Replace the `strings.Contains` with a
HEAD/GET of the resolved URL and assert 200. Roughly a ten-line change in
`check_tool_acceptance.go`, and it closes the `041` class permanently rather than
by census.

**2. A `script_assets_resolve` discovery check.** For each deployed page, extract
every `<script src>` and request it; flag non-200. No browser needed, cheap
enough to run fleet-wide, and it covers owned pages, chrome and snippets — the
population Tier 4 cannot reach. This is the single highest-value item here.

**3. Widen Tier 4 past `component_level='tool'`.** The runner already does
everything needed; the gate is one predicate. An owned tool page could carry
criteria in `content_data` rather than a PLAN fence. **For webdesign.co.uk
specifically this is the shortest path to real coverage of ~16 canvas/clipboard
tools that otherwise need manual clicking.**

**4. Build T5.1** (post-hydration dead-control assertion) to give the
`<button>`-with-no-handler class an owner.

**5. Script parity at generic write sites**, generalising `checkScriptParity`.

## How to verify any of it

```bash
curl -s https://<domain>/<page> \
  | grep -oP '(?<=<script src=")[^"]+' \
  | while read -r s; do
      printf '%s %s\n' "$s" "$(curl -s -o /dev/null -w '%{http_code}' "https://<domain>${s}")"
    done
```

For webdesign.co.uk on 2026-07-26 every script returns 200 —
`/tools/assets/webdesign-couk-header.js` and each tool's sibling engine.

## Prior art — read before building anything

- `travelling_docs/OVERVIEW_self_verifying_tools.md` — the Tier 0/1/2/4 ladder.
- `internal/adapters/browserrunner/run_checks_action.go` — the browser tier.
- `discovery_checks/check_dead_controls.go`, `check_truncated_component.go`,
  `check_tool_health.go`, `check_tool_acceptance.go`.
- `experience_loop/RUNBOOK_experience_loop.md` T2.3 and T5.1.
- `bugs_closed/041`, `/046`, `/023`, `/024`; `bugs_open/033` (detection with no
  consumer), `/071` (findings discarded).

---

## 2026-08-05 — candidate 2 BUILT and council-APPROVED; candidate 3 was already done; candidate 1 is RFC-scope and must not be attempted here

Worked by `docs024_key_docs_latest/bugfix_084_asset_reference_resolution/`
(PLAN / RUNBOOK / NOTES). **Status: still OPEN** — the check is built, committed
and approved, and is **inert until a chassis image carrying it rolls and it is
enabled**. `bugs_closed`'s bar is fixed AND live.

### Two corrections to the text above, both stale rather than wrong-at-the-time

> **CORRECTED 2026-08-05 — §2 cites `check_tool_acceptance.go:375`.** The
> `strings.Contains` is at **`:433-440`** today. The code is otherwise unchanged
> from the filing: `asset_loads` still passes on presence and never requests.

> **CORRECTED 2026-08-05 — §1 is SUPERSEDED. "Tier 4 only ever runs against
> `component_level = 'tool'` pages" was true on 2026-07-26 and is not true now.**
> `tool_eligibility.go` widened both Tier 2 and Tier 4 past that predicate on
> 2026-07-29, registered as **TL-033**, making 71 components newly eligible
> (webdesign.co.uk 63, gamesdesign 6, vonc 2). **So fix candidate 3 is DONE.**
> What remains of it is that coverage is still *criteria*-gated — exactly 1 of
> those 71 carried a PLAN — which is why the new check needs no criteria at all.
> `bugs_open/146 (ported_tool_pages_are_outside_every_acceptance_tier)` is stale
> in the same way and should be re-read before anyone works it.

### Fix candidate 1 is OUT OF BOUNDS for a bug patch — already ruled, twice

Making `asset_loads` actually load **changes what the type means for every
document that uses it**, which the owner ruling of 2026-07-29 makes RFC-scope. It
was argued and settled in a council round months before this session:

> *"the true reason is that fixing it would change what `asset_loads` means for
> 63 live documents, which the owner's 2026-07-29 ruling makes RFC-scope"*
> — `experience_register/harvest/entries/CC-001_feed-driven-teaser-list.json:255`

It would also flip `asset_loads` from `experienceStaticConfirming` to
`experienceStaticRefuting` (`static_attribute_checks.go:95-112`), which is
verbatim the RFC_002 trigger, and would fail `TestEveryStaticCheckTypeIsClassified`.
The platform's chosen resolution was to **re-type the network claim to Tier 4**
(`NOTES_experience_register.md:978`) — where nothing yet drives a browser for the
register, so the capability was given up rather than smuggled in.

> **Ground the figure before you re-argue it.** Measured 2026-08-05: `doc_plans
> WHERE is_current` has **6 plans / 8 `asset_loads` occurrences**. The "63"
> counted every superseded version (66 plans / 153 occurrences all-history). The
> ruling does not depend on the count — RFC-scope is triggered by a change to
> what the mechanism GUARANTEES — but an RFC should argue from 8, not 63.

**Owed:** an RFC in `architecture_review/` proposing the `asset_loads` upgrade on
its own merits. Not filed by this session.

### What was built: `asset_reference_404`

`platform/orchestration/actions/discovery_checks/check_asset_reference_404.go`
(+ tests), commits `e526a5196` and `9496f2cbd`. Concept register **IMP-051**.
Council **APPROVED round 1**, correlation
`3675ec9a-4a8b-4c3f-b3b4-cc95498584c8`, 5 advisory objections, none high.

Resolves every `<script src>` and `<link rel~=stylesheet>` on deployed
`page_components` and on `site_components` chrome to an absolute URL and probes
it. **Only 404/410 file, and a candidate 404 is confirmed by a second request
first**; 401/403/429, all 5xx, timeouts and transport errors are logged skips.
That asymmetry is the design, not timidity: a prior Python sweep of this exact
population had all 63 webdesign.co.uk requests refused by Cloudflare, and a
"non-200 is a finding" rule would have filed 63 false 404s on a site whose
scripts all serve 200. Flag-only; retracts through `CheckResult.Resolved` on a
positive re-observation only.

### THE MEASUREMENT THAT SHOULD SHAPE HOW ANYONE READS THIS: the population is ZERO

All 541 deployed pages fetched and **DOM-parsed** (not regexed — see below):
854 `<script src>` elements, **96 distinct referenced assets, 96 of 96 returning
200**. Negative control confirmed on four sites: a missing asset returns a real
404, so a status check discriminates. It simply has nothing to find.

**This is a regression guard, not a repair.** The class has bitten
(`bugs_closed/041`; `cmd/webdesignport` carries `checkScriptParity` because ~60
of 63 tools nearly shipped as dead markup, caught by luck) — but nobody should
read a future "0 findings" from this check as proof it works. That is exactly
what a silently broken check reports. The only evidence it bites is the mutation
table and the induce-and-revert procedure in the lane's RUNBOOK.

> **A misstep worth inheriting.** The first measurement regexed `rendered_html`
> for `<script[^>]+src=` and produced one "live 404": a `<script src="...">`
> that turned out to be **a comment inside a tool's own JavaScript** describing a
> regex. A regex cannot tell an element from a mention of one, and tool pages —
> this bug's whole population — are the pages most likely to talk about HTML
> inside JS. `WRONG_CALLS.md`, 2026-08-05.

### The objection this session could NOT answer, left standing for a human

The `bug_historian` seat, [medium], advisory:

> *"the platform keeps adding new detected-but-unhandled work item types (071,
> 079, 083, and now this one) faster than it drains them, and this plan
> explicitly chooses to extend that population rather than resolve it, citing an
> owner ruling as cover. That ruling may be sound, but a human should confirm it
> still holds given the growing size of the undrained pile."*

Recorded rather than argued away. Flag-only is defensible per the owner ruling of
2026-08-02 §1 and the repair genuinely is a judgement no generator can make — but
the seat is right that this is another occupant of an open systemic gap.

### What is still owed before this can close

1. **A roll.** Verified 2026-08-05 21:50Z that chassis **v1.0.1254 does NOT carry
   it** — pod-grep `asset_reference_404` = 0 and `agentchassis-discovery` = 0,
   with `image_url_404` = 7 as the positive control and `asset_reference_405` = 0
   as the negative. The build predates the commit; that is expected, not a
   failure.
2. **Enablement, in order.** Pod-grep first, then add `asset_reference_404` to
   `design-discovery-agent`'s `checks` array, and update `liveConfiguredChecks`
   in `discovery_checks_registration_test.go` in the same commit. An unregistered
   check name **fails the step** (`bugs_open/149` B4). SQL and gotchas in the
   lane's RUNBOOK.
3. **Prove it bites in production** — induce one broken reference on a page we
   own, assert exactly one item keyed on the resolved URL, then revert and assert
   the retraction.
4. **Candidates 4 (T5.1 post-hydration dead-control assertion) and 5 (generalise
   `checkScriptParity`) are untouched** and are why this file cannot simply be
   closed once the check is live. Candidate 5 overlaps `bugs_open/178`/`198`.
