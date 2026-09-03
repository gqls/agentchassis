# NOTES — site-design-planner (running record, append-only, newest at bottom)

## 2026-09-02 — lane opened, survey done, no code touched yet

Session was renamed "site design planner" by the owner, with the instruction to
pick up the thread if one exists or take responsibility for it if not.

**Checked for an existing thread first.** `MEMORY_workstreams.md` has no entry
for it; no `docs024_key_docs_latest/site_design_planner*` directory existed;
`bugs_open/`/`bugs_closed/` grep for `site-design-planner`/`resolve_composition`/
`install_site_composition` returns only `bugs_open/113` (a real hit, but owned by
the `brochure_component_library` thread, not a dedicated site-design-planner
lane) and `bugs_closed/291` (a passing mention — 291 is about a different phantom
handler, `hitl-review`, hit incidentally on one of this mechanism's own work
items — see below). `who-owns.py` returns no match for `site-design-planner`,
`needs_composition`, or any of the three domains below. **Conclusion: no active
thread. Took responsibility per the owner's fallback instruction.**

**Read the concept register** (`design-composition.md`, DES-001 through DES-062,
freeze date 2026-07-13) to get the mechanism's shape and history rather than
guessing from code cold. Cross-checked live: `agent_definitions` row for
`type='site-design-planner'` exists, `status='active'`, single version.

**Read `bugs_open/113` in full** (huge file, 2026-07-27 → 2026-08-12, six
sessions' worth of corrections-on-corrections). It is the most complete recent
account of how this mechanism actually behaves under load, including two
self-corrected wrong attributions in the same session (both instructive — see
PLAN §1). Its tail names the fix that closed the "no re-resolve" platform gap
(`allow_reinstall`, per-request, chassis v1.0.1290) — checked live in RUNBOOK §3
is copied straight from that file's worked example.

**Queried live `site_work_items`** for anything in this mechanism's item types
still open. Found three (PLAN §2). Checked each one's actual current state
rather than trusting status alone:
- `adversecreditmortgage.co.uk`'s `needs_composition` looked like a live stuck
  item at first read (`unresolved`, empty spec/result). Read `load_work_item_actions.go`
  around the two-strike anti-churn logic and realised `unresolved` at
  `attempt_count=0` means "this key already failed twice recently", not "this
  attempt failed" — the row itself never ran. Then checked `site_specs` and
  found `resolved_composition` was written 2026-08-25, after this item's last
  update — the site got composed some other way and this row is stale. Then
  pulled the site's full work-item history and found ~230 other unresolved rows
  from an unrelated Anthropic billing outage on 2026-08-25/27. **This item is not
  a composition-mechanism defect** — recorded so nobody re-derives the same dead
  end.
- `loancalculator.co.uk`'s `needs_composition` is a deliberate 2026-08-12 park,
  and that domain already has its own active session (`loancalculator` in
  `ListAgents`). Not this thread's to un-park.
- `ai-agent-orchestration.com`'s `needs_new_layout_candidate` is the one
  genuinely open, in-scope question — see PLAN §3. **Not yet investigated
  further this session.**

**Cross-session coordination.** Mid-investigation, `bugs_open/427` messaged
asking whether this thread touches `build-site-planner`/`write_site_plan_action.go`/
`validate_site_plan` (they're starting on `bugs_open/428`, a build-site-planner
bug, and the name similarity worried them). Replied: no overlap, this thread is
the composition-resolution agent only. Worth recording because the two agent
names (`site-design-planner` vs `build-site-planner`) are close enough to
collide in a skim, and apparently already have, at least once, in a cross-session
check rather than a wasted edit.

**What's NOT done yet:** the one real open item (§3 in the PLAN) — whether
`ai-agent-orchestration.com` now has real classification tags and could resolve
to something other than `brochure-formal`. That's the natural next step if this
thread continues.

## 2026-09-02 (continued) — traced the "why", fixed it, filed it, submitted it

Went to check whether `ai-agent-orchestration.com` now has real classification
tags (the natural next step §3 named). Queried its current `classification` spec
— **still no `industry_tags`, refreshed as recently as today 04:59Z** by
`evaluate_news_feed`. Read that action's code first rather than assuming it
clobbered something: it deep-merges (`feed_news_recommendation_action.go:334`),
so it isn't the cause — it just re-stamps whatever shape was already current.

Read `resolve_composition_layout_action.go`'s `extractClassificationTags` next,
expecting it to match the shared `readClassificationFromContext` helper (the one
`resolve_composition_helpers.go`'s own doc comment says handles exactly this
case). **It doesn't.** `extractClassificationTags` reads only
`classData["category"]`/`classData["industry_tags"]`, no identity fallback.
Grepped `readClassificationFromContext`'s actual callers: `install_site_composition`,
`resolve_composition_typography`, `resolve_composition_palette` — three of the
four composition resolvers. The layout resolver, the one that decides between 18
named layouts rather than blending values, was the one NOT using it.

Measured the live population before writing anything: 4 sites currently have a
current `classification` spec with no `industry_tags` at all (`finetuning.uk`,
`leopardessconsulting.co.uk` — legacy classifier shape, 2026-04-18; `gaswholesalers.com`,
`ai-agent-orchestration.com` — both re-stamped by `evaluate_news_feed` today, same
underlying legacy shape carried forward). `ai-agent-orchestration.com`'s `identity`
spec has real data (`industry: "Technology Services"`, present since 2026-05-01)
that the layout resolver never consulted — this is exactly why its
`needs_new_layout_candidate` item fired on 2026-08-12 with `site_tags: []` and has
sat since.

**Fixed** (`bd8e45aba`): swapped the private extraction for the shared helper,
deleted the now-dead 58-line function (confirmed one caller, no other refs, `go
build ./...` clean after removal). Added a sqlmock regression test reproducing
`ai-agent-orchestration.com`'s exact spec shapes
(`resolve_composition_layout_action_test.go`) — asserts `site_tags` non-empty and
`is_fallback` false against a mocked layouts table; without the fix it would hit
the zero-terms short-circuit in `resolveLayoutByTags` and never reach that table
at all.

**Concurrent-tree collision, handled cleanly.** Mid-edit, the "theme kits" session
committed a `loadSiteThemeKitDefaults` short-circuit into the SAME function, just
below my call site — I only found out because a `Read` came back with a
system-reminder saying the file had changed on disk. Did not revert or fight it;
re-read, confirmed it built on top of my already-changed code cleanly (it
references the `category`/`industryTags` variable names my fix introduced), and
my test still passed against it. `scripts/verify-head-builds.sh` then caught a
real but not-mine gap — HEAD didn't build because their symbol was only in the
working tree — so I messaged them directly rather than guessing at their
implementation. They committed within minutes; HEAD builds clean now
(`0902039c0`). Worth recording as a worked example of the "another session may
commit between your add and your commit" warning in CLAUDE.md actually firing,
live, mid-task.

**Submitted to council** (`bd469ba1-228e-443e-a04d-6a577a210e5d`,
`Council-Submitted:` trailer — verdict pending, not yet read). **Filed
`bugs_open/431`** (case file, kept open per the fixed-AND-live bar — this is code
only, not yet in a rolled chassis). **Added a 016b §9 pattern entry** — the
transferable shape is "a sibling of a shared helper's callers can diverge
silently; only the input shape the majority case never exercises exposes it."

**Deliberately not touched:** `ai-agent-orchestration.com`'s stuck queue item
itself. The fix makes a future re-resolve for it (and the other 3 affected sites)
see real signal for the first time; whether to actually trigger that re-resolve,
and what it picks, is a decision for that site's own thread/owner, not something
to do or predict from here.

## 2026-09-02 (later still) — council verdict read: APPROVED. Still not live

Checked the correlation (`bd469ba1-228e-443e-a04d-6a577a210e5d`) —
`metadata->>'decision' = 'approved'`, read 2026-09-02 12:33Z. Per the standing
rule, did **not** amend the commit to add `Council-Reviewed:` (forward-only
forbids the amend, and `098_REPORT` credits the correlation automatically once
it runs). Updated `bugs_open/431`'s status line to say so.

**Checked whether it's actually live before believing "approved" means
"shipped."** `service_binary_capabilities` says the deployed chassis is built
from `a2732c7207d…`; `git merge-base --is-ancestor bd8e45aba a2732c720…` is
**false** — the deployed build predates this fix, it has not rolled. Left as-is
deliberately: the working tree currently carries a large set of modified
`deployments/kustomize/.../kustomization.yaml` files (visible in `git status`,
not mine), which reads as a fleet-wide release already in preparation by
someone else. Building or deploying this one service myself right now risks
colliding with that — matches the standing "releases are WHOLE-FLEET, owner
runs make release" practice. Left for the next coordinated roll; `bugs_open/431`
states plainly it is not verified live yet.

**Notified the two other affected sites that have active sessions** —
`finetuning` and `leopardess` (for `finetuning.uk` and
`leopardessconsulting.co.uk`, the two legacy-classifier-shape sites from the
census) — a short FYI, not a request for action, matching the "contribute the
measurement, don't act unilaterally" norm this whole file has followed
throughout. `ai-agent-orchestration.com` has no active session right now
(checked `ListAgents`) so nobody to notify there beyond the bug file itself and
the memory workstream entry.

**Nothing further actionable in this thread right now** without either (a) a
fleet roll landing this fix, or (b) new site-design-planner-scoped work
appearing in the queue (checked — still the same 3 items, unchanged since this
morning). Next session picking this up should re-check both before assuming
anything has moved.

## 2026-09-02 (later) — `bugs_open/438`, a sibling structural bug in the same mechanism, flagged by `theme kits`

Not this thread's find — `theme kits` filed it and messaged directly because
`resolve_composition_pallette_action.go` is a file this thread owns (part of
the four composition resolvers, same family as `bugs_open/431`). **Verified
independently rather than accepted on report** (three checks, all confirmed):

- `extractPaletteSignal`'s rung 1 doc comment and code both say
  `mission.preferred_palette` — read the function directly, `:224-236`.
- `082_submit_domain_unified.sh:143` sends the brief under `mission_brief`, not
  `mission` — grepped the script.
- Live `resolved_composition.lineage.palette_source` distribution:
  `design_intent_values` 30, `archetype_default` 1, **`mission_hint` absent
  entirely** (not just zero-count — no row) — queried directly, matches their
  claim exactly.
- Same shape confirmed in `resolve_composition_typography_action.go:213-215`
  (`mission.preferred_typography`), the fourth sibling resolver, on my own
  check — not in their file but worth recording since it's the same mechanism.

**The claim:** `domain-submitter`'s `persist_mission` step reads
`input_data.mission` (nothing ever sends that key) and only its error-fallback
step, `persist_mission_brief`, actually runs — writing a DIFFERENT aspect
(`mission_brief`) that the palette/typography cascades never read. So the
"most authoritative" rung in both resolvers has never fired, fleet-wide, ever.
Silent — every site still gets a palette from rung 2, so nothing looks broken.

**Why I'm not acting on it, even though it's "my" file:** unlike 431, none of
the three ranked fix candidates in the bug file are low-risk. Candidate 1 (the
leading one — repoint `persist_mission` at the key that's actually sent) is a
live-on-apply config change that would immediately start OVERWRITING
`gamedesign.uk`'s hand-seeded `mission.preferred_palette` row on its next
rebuild — a lever that lane is actively relying on precisely because nothing
currently touches that aspect. Candidate 2 widens the resolver's read surface
with an unguaranteed shape. Candidate 3 (retire the rung, admit it's dead) is a
policy call about whether human-specified palette preference should be a real
capability at all — not mine to decide unilaterally either. **Replied
acknowledging the finding, confirmed independent verification, declined to pick
a candidate without the gamedesign.uk lane and/or the owner weighing in given
the named cross-lane consequence.**

**Update from `theme kits`, same day:** the typography measurement is now in
438 itself (`typography_source`: `fingerprint_font_family_match` 30,
`fallback_sans_modern` 1, `mission_hint` 0 — same shape as palette, both rungs
confirmed dead). **§6's overwrite hazard is CORRECTED, not just restated** —
`write_site_spec` deep-merges, so candidate 1 would NOT actually clobber
`gamedesign.uk`'s hand-seeded row as originally warned; the real (narrower)
hazard is any writer that supersedes-without-merging, which exists elsewhere in
the tree. My caution above is therefore partly stale — worth knowing, not
worth chasing down and rewriting mid-file, since a live discriminating test is
already running: `gamedesign.uk` is the first site ever to reach composition
with a populated `mission.preferred_palette` row, and whether it resolves via
`mission_hint` settles whether 438's whole mechanism account is right or needs
reopening. Not acting until that lands — they said they'll report the result.

**Second update, same day — scope grew, spot-checked both new claims
independently before recording them:**
- `082_submit_domain_unified.sh` sends no `roadmap` key at all
  (`grep -c roadmap` → **0**, checked myself) — so `persist_roadmap` and
  `persist_roadmap_brief` are dead on the only path that runs them, not merely
  under-firing. Live rows: `roadmap` 1 site, `roadmap_brief` 4, against
  `mission_brief` 22.
- A demand control now exists: `agent_error_log` rows for
  `missing required fields: [spec_data]`, 30-day window — re-ran their query
  myself, exact match: `persist_mission` 16/12 sites, `persist_roadmap` 16/12,
  `persist_roadmap_brief` 14/11, `persist_mission_brief` 6/3. This is the
  before/after meter a fix should be judged against (should go to zero on the
  next fresh submit).
- Correction to the mechanism: `error_step` chains are a linear continuation,
  not a designed fallback pair — `persist_mission_brief` "rescuing" the brief
  is incidental ordering, not intended pairing. So deleting the dead
  `persist_roadmap*` steps has no pairing to preserve.
- Net effect on scope: what was "repoint one step" may now be "repoint one,
  delete two" — a materially different, and if anything simpler, fix than
  either of us was looking at. **Still not acting** — same gate as before
  (`gamedesign.uk`'s `palette_source`), confirmed by them directly against the
  live pipeline (`needs_strategy` triaged, composition not yet reached).

## 2026-09-02 (later) — `gamedesign.uk`'s test landed: diagnosis CONFIRMED. Then I found candidate 1 is wrong anyway

`palette_source = mission_hint`, doubly discriminated (lineage string AND the
landed hex colours matched the hand-seeded values byte-for-byte, not the
classifier's near-identical rung-2 values). 438's mechanism account is right.
`theme kits` handed the decision to me explicitly ("your call from here").

**Before picking a candidate, traced what "repoint" would actually do — and it
doesn't do what the file's own §7 verification step claims.** `extractPaletteSignal`
checks `mission["preferred_palette"]` as a structured map; `082` only ever sends
free text (`--mission`/`--mission-file` → `{"text": "..."}"` under `mission_brief`
— grepped the whole script, no structured alternative exists). So repointing
`persist_mission` to read `mission_brief` would write a bare string into `mission`,
never satisfying the map check — candidate 1 fixes nothing for ordinary `082`
submissions.

**Found the actual working producer while checking who the "1 pre-existing
`mission` site" was** (438 never named it): `vonc.com`, via a bespoke
"Tier 3" script (`080_submit_vonc.sh`) that sends a genuinely structured
`input_data.mission` object — the shape `persist_mission` was built for. That
script sends BOTH `mission` and `mission_brief` in one payload, so candidate 1
would also silently stop capturing the rich `mission` object from any future
Tier-3-style submission, for zero gain on the `082` side. **Candidate 1 is not
a smaller fix than advertised — it's the wrong fix**, and only candidate 3
(retire the rung, or build `082` a real structured-preference input, which is
a design question, not a bugfix) survives this correction intact.

**Contributed the correction into `bugs_open/438` directly** (my own file, same
collaborative convention the file already uses across three lanes) rather than
raising it only here — it changes the shape of a decision another lane is
about to make, not just a fact for my own record. Not implementing anything
myself; flagged to `theme kits`.

## 2026-09-02 (later still) — `bugs_open/438` CLOSED OUT to an owner decision

`theme kits` verified the §6c correction independently and extended it:
candidate 2 dies to the identical fact (its own doc comment already said so —
"`mission_brief`... not guaranteed to carry a `preferred_palette` map", written
and then not acted on). §5 struck and pointed at §6c rather than duplicated.

**Final shape of the bug, agreed by both lanes:** not a misrouted read, not a
mis-pointed write — **the standard build path (`082`) has no way to express a
structured design preference at all.** The "most authoritative" rung was never
broken; nothing has ever been able to reach it. Two candidates survive:
retire the rung (say plainly this capability doesn't exist), or build it
(give `082` a real structured-preference input — a CLI flag or file, an actual
design change, not a config repoint). **Neither lane is choosing between
those** — both explicitly declined, and `theme kits` signed off with "nothing
further from me on 438 unless the owner picks a direction." Surfacing to the
owner directly rather than leaving it to be found in the bug file.

## 2026-09-02 (later) — owner critique routed via `designblog.co.uk`: composition sameness, filed as `bugs_open/445`

Owner reviewed a freshly-live site and found it "exactly the same as all the
other sites"; routed to five design-side threads including this one, asking
specifically whether the sameness is library breadth, matcher behaviour, or
chrome (outside this mechanism). ACKed, then measured rather than guessed.

**Finding: three sibling remakes (`designblog.co.uk`, `advertise.co.uk`,
`websitepromotion.co.uk`) all resolved to `magazine-grid`.** Traced the actual
mechanism (`resolveLayoutByTags` scoring, read line-by-line) rather than
inferring: of 18 layouts, exactly one (`magazine-grid`) fits "professional
content hub" — the only other editorial-category layout (`soft-editorial`) is
built for lifestyle/wellness blogs, wrong register entirely. All three sites
are genuinely the same underlying shape — a content hub whose core offering is
embedded interactive tools — and the library has no archetype for that shape.
**This is a real library-breadth gap, not a matcher defect**: checked and ruled
out two more specific theories first (the classifier literally tagging sites
"magazine-grid" — real prompt-hygiene issue, but doesn't actually score against
anything; the matcher ignoring differentiating signal — it evaluated genuinely
different candidate shortlists per site). Secondary, less certain finding: 9 of
18 layouts fleet-wide have never been chosen for any live site, 3 layouts cover
73% of deployed sites — flagged as worth its own look, not claimed as the same
mechanism.

**Chrome (identical header/footer nav) confirmed OUT of this mechanism**: all
three sites' `style_collections.header_component_id`/`footer_component_id` are
NULL, meaning `link_site_components` (a different action, different owning
thread — `components`, per the critique's own routing) never linked them, so
they're serving a hardcoded fallback rather than picking from the library's
actual header/footer variety. Not investigated further — correctly routed
elsewhere already.

Filed `bugs_open/445` with the full evidence and three ranked fix candidates
(build the missing archetype; fix the classifier's few-shot examples on their
own merits; investigate the 9 never-chosen layouts as a separate piece of
work). Replied to `designblog.co.uk` with the same read. Not implementing the
new layout myself — a real design task, 18 more remakes queued, worth doing
right rather than fast.

**Corroborated independently — `theme kits` measured the same fleet
concentration (73% on three layouts, 9/18 never chosen) before my message
arrived, unprompted.** Two separate measurements agree. Their nav-side finding
(`ChromeSlotFunction()` hardcodes slot→function, 10 chrome-eligible headers
unused, only 6/40 collections pin `header_component_id`) matches §5's
NULL-linkage finding from a different angle — `components`' territory
confirmed by three lanes now, not just asserted by one. Who designs the new
archetype is going to the owner as a priority call, not assigned by either
session. `designblog.co.uk` closed the loop: "nothing further needed" — no
reply owed.

## 2026-09-02 (later) — `vetcomparison` asked for a design-pass view before touching anything

Owner wants vetcomparison.uk's homepage "a bit better designed", content
frozen. Asked this lane specifically, with a full brief including a parked
`palette_contrast` capability gap (`d6da17b4`: accent `#10b981` used as ink on
`#f8fafc`, 2.42:1 vs 3.0:1 needed) and a hard sequencing constraint —
`bugs_open/357`'s migration 701 (retyping the homepage's comparison tool out of
a mislabelled `hero` row) MD5-census-guards this exact page and aborts on any
concurrent write.

**Recommended NOT re-resolving the composition** — `industry-hub` is the right
fit, palette/imagery were deliberately pinned 2026-08-24, a fresh resolve risks
the colour-churn landmine for no gain.

**Diagnosed the contrast gap precisely rather than repeating their summary.**
Checked the served stylesheet: `--color-accent-text`/`--color-accent-ink` are
already correctly derived (`#0f172a`, would pass). The bug is the `latest-news`
component's own embedded CSS (`.news-card-title a:hover`) hardcoding raw
`--color-accent` as ink instead of the pre-derived text slot — a one-line
template fix, not a palette defect. **Fleet-relevant**: this component
(`content_components` id `77dafa26…`) is live on **8** deployed sites, so the
fix belongs at the shared template, not per-site.

**Not touching it.** Even a shared-template edit could conceivably fall inside
357's write guard and I don't know its exact scope — recommended it wait for
their remainder batch rather than risk an abort for a fix this small. Replied
with the full diagnosis; did not implement.

> **CORRECTED same day, hours later — the priority fix above was imprecise, and
> the correction came from a peer's critique, not my own re-check.** The
> `offer analyser benefit analyser visual designer` (vigilant-designer) lane
> critiqued `vetcomparison.uk/index` directly and found `--color-accent` is
> "never applied to anything" on the live page — checked this myself before
> passing it on, since it directly contradicts the fix I'd just recommended.
> **Their central point holds; their absolute "zero times" was one selector too
> strong.** `.news-more-link` (my second flagged selector) is genuinely dead —
> the class never appears in the rendered body, only inside repeated `<style>`
> blocks. But `.news-card-title a:hover` **is** real: 3 genuine news headline
> links exist in the served HTML, so hovering one does render accent-as-ink at
> 2.42:1. Momentary and hover-gated, not the priority I'd made it. The higher-
> value reframe, which I now agree with over my own original recommendation: the
> palette promises three colours and the page visibly delivers two (blue +
> red-for-dates); decide to actually use the accent or drop it, rather than
> patch a contrast ratio nobody sees at rest.
>
> **Also surfaced, more consequential than anything either of us had flagged:**
> `--color-primary` (#2563eb) as ink has only **0.44 of contrast headroom**
> (4.94:1 vs 4.5 required) and is used **statically and widely** — nav active
> state, category strip, more (checked: `color: var(--color-primary)` appears
> in multiple non-hover selectors). Real, live exposure on something
> load-bearing, unlike the accent issue. This is the actual constraint on any
> future palette move for this site, not the parked accent findings.
>
> Sent a correction to both lanes rather than let the imprecise version stand
> uncorrected in either thread. **Transferable lesson for this lane**: a CSS
> variable's presence in the stylesheet, or even a plausible-sounding selector
> match, is not evidence it reaches rendered content — check the served HTML
> for the class, not just the CSS for the property. I did this correctly for
> the earlier `resolve_composition_layout` work (431) by reading code before
> asserting; I skipped the equivalent step here (checking the served DOM before
> naming a priority fix) and a peer's independent check caught it.

**Final reconciled picture, three lanes converging same day:**
- The 3 parked `contrast_failure` items are **real and separate from the accent
  question entirely** — a muted grey `rgb(107,124,133)` at 4.10–4.14:1 inside
  the comparison tool's own markup, which `bugs_open/357`'s migration 701 is
  about to adopt as a component. Fix is one edit to the adopted component's
  `html_template`, **owed after 701, not before** — a fixing pre-701 would abort
  their census. `vetcomparison` has already told the 357 lane to expect it.
- The accent is **vestigial, not absent** (vigilant-designer self-corrected
  twice, found its own regex missed every usage with a CSS custom-property
  fallback — logged as a pattern, third such miss that lane had that week):
  one live decorative `::before` mark, two hover-only rules, three dead. "Blue
  and red with a decorative green tick" is the accurate description. Also
  caught: the accent's CSS fallback value is amber (`#d97706`) — if the
  variable were ever unset, the page would render a different colour identity
  entirely. Not urgent, worth knowing.
- `--color-primary` (0.44 contrast headroom, static, load-bearing) remains the
  sharpest live finding and should lead any combined write-up over the accent
  question.
- Additional parked specs from `vetcomparison`'s own audit (not independently
  verified by this lane): a contact-page button at 3.77:1, a spacing-token
  mismatch on `info-card-grid`, and 5 `needs_design_review` items including one
  that brushes the content freeze (a CTA button carrying a full paragraph —
  restructuring is fair game, rewording is not, per the owner's own words).
- `vetcomparison` is coordinating the full pass and holds sequencing; this
  lane's part is done — composition stays as-is, the "honest colour use"
  decision and any imagery investment go to the owner as flavoured calls, nothing
  implemented from here.

## 2026-09-02 (later still) — owner ruled on both; gave a real composition-level plan

Owner ruled: use the accent deliberately (don't drop it); build per-section
imagery for the homepage (vetcomparison = first live exercise of IMG-075).
Both explicitly routed back to this lane for a composition-level view.

**Accent: found a genuine, low-risk home for it, already inside the layout's
own vocabulary.** `industry-hub`'s CSS template has a dedicated "independence
claim" slot (`--color-independence-bg`/`-border`, the layout's own comment:
"visible, not buried"), currently **unset in vetcomparison's palette** —
falling to the layout's hardcoded blue default, which is the exact "third
blue" the needs_design_review item flagged. Recommended repointing those two
palette slots to accent-derived values: solves the accent-use ruling, the
third-blue review item, and fits the site's actual semantics (an
"independently reviewed" trust badge in green, not blue-on-blue) — three
things from one palette edit. A palette-value change + re-render, not a
component/structure write, so it shouldn't touch what 357 is guarding — flagged
that as needing their confirmation, not asserted as safe.

**Imagery: checked the layout has NO existing image-slot scaffold at all** —
only generic `img{}` and hero-specific CSS. So this is greenfield CSS work at
the layout layer too, not just `site_assets`/`site_plan_imagery` data-plumbing
— whoever runs IMG-075's first exercise will need a companion layout treatment
designed (placement, sizing, how it sits against the existing sections).
Offered to design it once the decision is final; did not build it
speculatively.

Plan only, nothing implemented — matches `vetcomparison`'s own "plan now,
write after" framing, and everything still sequences behind 357.

## 2026-09-02 (later still) — 701 closed, unblocked, and this lane EXECUTED the accent batch

`vetcomparison` signalled 701 applied and closed (proven in production —
survived an organic rebuild byte-identically), design pass unblocked, and
handed this lane the write: palette values + the amber-fallback fix, batched,
then a rerender with the load-bearing `reason='template_changed'` (016b §10
row 404 — the wrong/missing reason silently re-ships stale stored bytes and
reads as success; checked this row before doing anything).

**Read before write, both rows, confirmed exact current state first:**
`palettes` (`palette-vetcomparison-uk`, 8 keys, neither `independence_bg` nor
`independence_border` set — matches the earlier finding) and `content_components`
(`latest-news`, confirmed **6** occurrences of `#d97706`, not 5 as the earlier
proposal doc miscounted — corrected here).

**Executed, both writes verified by their own RETURNING clause:**
1. `palettes.colours` merged (`||`) with `{"independence_bg":"#ecfdf5",
   "independence_border":"#10b981"}` — additive, nothing else touched.
2. `content_components.html_template` (`latest-news`): all 6 `#d97706` → `#10b981`,
   verified 0 remaining amber, 6 accent occurrences post-edit.

**Queued two work items, both status `triaged` (bypassing the
detected→triaged promotion gap bug 113 already documented), both verified with
a `DO`/`RAISE` block per the "a SELECT block cannot stop a COMMIT" rule:**
- `needs_design` → `webdesign-agent`, to regenerate `styles.css` so the
  `.independence-banner` (confirmed live in the served HTML, 6 occurrences,
  not just defined in CSS) picks up the new specialised slots.
- `page_rerender` (`reason='template_changed'`) for the index page, to
  regenerate the stored `latest-news` section HTML with the corrected fallback.

**One real risk considered and accepted, not ignored:** `needs_design` regenerates
core palette slots via a fresh per-run LLM guess (bug 113's own documented
behaviour). Mitigant: `independence_bg`/`independence_border` are SPECIALISED
slots, which the merge-authority rule gives to composition, not the LLM — and
this site's `design_intent.palette.reference_values` is explicitly pinned
("preserved verbatim" per the original brief), which steers the LLM guess back
to the same core values every time. Flagged explicitly to `vetcomparison`
rather than assumed silently — they already planned to check "primary
untouched at 4.94:1" as part of their own verification, which is exactly the
check that would catch this if the mitigant failed.

Not touched: the 3 grey contrast fixes in the adopted tool component (their
call, needs a browser-check for runtime-injected rows first per the vigilant-
designer's caution) and the imagery work (explicitly sequenced after this batch
lands clean). `vetcomparison` verifies at the artefact; not re-checked from
here.

## 2026-09-02 (later still) — verification found one miss (the head chrome), and chasing it surfaced a bigger, standing hazard. STOPPED, did not touch it

`vetcomparison` verified the batch: core palette survived byte-identical
(the bug-113 LLM-reroll risk did not bite — though they noted the pin I cited
as the mitigant was actually retired by the owner the same day, so it was the
ONLY guard, not a belt-and-braces one; worth remembering for the next
`needs_design` run), palette write and component fix both confirmed live. **One
miss**: the independence banner still served the old blue — traced to
`site_components` slot `head`, which inlines a snapshot of the stylesheet
(`{{.theme_css}}` in its source template — confirmed clean, zero hardcoded
literals) that was captured 2026-08-27, before my fix, and never refreshed.

**Investigated before touching it, per their own caution about a prior
incident, and found something bigger than a stale-cache fix.** This site
carries **two open, 3-week-old `chrome_divergence_overwritten` items**
(`needs_human_review`, both 2026-08-11) — the platform's hand-patch-preservation
guard (`bugs_open/226`, live since 08-09) caught a chrome rebuild overwriting
hand-patched `header` content TWICE, archived it (2952/3094 bytes,
`site_component_history`), and is waiting on a human decision (restore via
config carriage, or dismiss) that nobody has made. Separately, a `needs_rerender`
item for this site's chrome (all three slots together) has fired every few
hours since 08-27 and gone `unresolved` every time for 5 straight days — not
diagnosed, error field empty on the rows checked. **A fresh instance of that
same item fired at 21:39:16 today — almost certainly triggered by my own
`needs_design` run — and was sitting `triaged`, unclaimed, at the moment I
found this.**

**Did not cancel it, did not refresh `head` myself, did not investigate
further.** A blanket chrome refresh touches all three slots, would very likely
re-trigger the header divergence a third time, and this is squarely the kind
of standing, cross-cutting, someone-must-decide situation bug 226's own text
says explicitly is "the queue's [call], not this lane's." Flagged clearly and
urgently to `vetcomparison` — including the live, unclaimed item — and left the
decision with them rather than acting on a mechanism I only just found and
don't own.

**Transferable for this lane:** a "just refresh the stale cache" fix can be
sitting on top of an entirely separate, already-flagged, unresolved hazard —
checking site_work_items for the SLOT/SITE before triggering any chrome
action is now a standing habit, not a one-off. Also: my earlier safety
reasoning for the `needs_design` risk (citing a pinned `design_intent` as the
mitigant) needs a note added — pins can be retired without this lane knowing,
so "cite the pin" is not durable safety, only what held on THIS run.

## 2026-09-03 — resolved benignly. Thread closed, one small dormant residue left named, not chased

The held item ran (21:51Z) and did no new damage: the guard's design meant
"nothing was re-archived", the green callout landed in the head chrome (theme_css
correctly picked up the palette write), and 16 per-page reranders were minted to
propagate it fleet-side. `vetcomparison` deliberately left the header hand-patch
decision for the owner rather than solving it as a side effect, and left the
5-day `unresolved` backlog undiagnosed rather than theorised — both good calls,
worth naming as the right shape of restraint, not just recording the outcome.

**One dormant residue, offered as optional, declined.** The served styles.css
(checked directly, regenerated 21:44Z — after both my writes) still carries 6
`#d97706` occurrences, not 0. So the amber fix I made to `content_components.
html_template` did not fully propagate into the generated stylesheet at
generation time — best unconfirmed guess, recorded as a guess: the CSS-bundler
reads stale per-page `rendered_html` snapshots rather than the live template,
which would mean full clearance needs all 8 `latest-news`-carrying sites to
rerender, not just this one. **Not investigated further** — zero live-rendering
impact (the CSS variable is always defined, so the fallback never fires),
explicitly framed as low-priority by the other lane, and chasing the
CSS-bundler's actual read path is a real, separate investigation this session
chose not to open today. Named precisely so it can be picked up later without
re-deriving the starting point.

**This closes the vetcomparison design-pass thread for now.** Composition
stayed untouched throughout (the original recommendation held); the accent got
real, deliberate use; the imagery work is next, sequenced after this batch,
not started.

## 2026-09-03 — reopened: the independence-banner ELEMENT never renders at all. I made the exact mistake I'd already caught once this thread

`vetcomparison`'s own follow-through found it: 0 `.independence-banner`
elements on any live page (checked 7), 0 `page_components`/`site_components`
rows for it. **My earlier "6 occurrences" check for this exact selector
(2026-09-02) counted CSS rule definitions, not rendered markup — the identical
error the vigilant-designer's regex made on the same page, and that I myself
caught and corrected for `.news-more-link` a few hours earlier in this same
thread.** I did not re-apply that lesson to my own next check on the same
page. Worth being blunt about in this file rather than quietly folding it into
"corrected" language: **checking a selector's occurrence count in a served
file is not checking whether it renders — that has now bitten this thread
twice on one page, once caught by a peer both times.**

**Consequence:** the owner's "use the green deliberately" ruling is still
unsatisfied — the whole accent batch gave the palette/CSS infrastructure real
values with nothing consuming them. Investigated the offered fallback path
(a "claimed" chip on the directory page) before agreeing with it: the
component template only has a NEGATIVE badge (`{{if not .is_claimed}}`) — no
positive branch exists, so this isn't "restyle an existing element" either, it
needs new markup, same as a banner. **Corrected that premise back rather than
letting it stand.** The one path that's genuinely free of new content: widen
the existing decorative `::before` heading-mark (currently scoped to
`.latest-news-section .section-heading` only) to the other section headings —
pure CSS, checked its exact selector scope, ready to draft if wanted. Left the
choice with `vetcomparison`/the owner rather than picking for them — both
substantial options touch the content freeze regardless of which one looks
smaller.

## 2026-09-03 — owner approved the claimed-chip (checkmark, no word); implemented

Read the current `directory-listing` template fresh (diffed against the
2026-09-02 copy — unchanged, safe to build on) before editing, per the
read-before-write discipline this whole thread has kept. Added the positive
`{{if .is_claimed}}` branch beside the existing negative one, unchanged
copy on the negative side. Glyph is a bare `✓`, `role="img"
aria-label="Claimed listing"` (metadata per the owner's own ruling, not page
copy). **Deliberately reused the independence_bg/independence_border tokens**
from the earlier (still-unplaced) banner work rather than inventing new ones —
same trust-signal family, and it's the pattern the owner pointed at. Glyph
colour is the accent-ink token, never raw `#10b981` (would fail 3:1... actually
non-text is 3:1, but the owner's own instruction named the ink token
explicitly, so honoured it as written rather than relying on the numbers
alone).

**Applied via a scripted, verified diff, not a hand SQL replace**: python
generated the new template from an exact-match, uniqueness-checked splice
(both anchors confirmed `count==1` before writing), MD5-verified the fetched
copy matched the live DB row byte-for-byte (a trailing-newline artefact from
`psql -A` output was the only difference, confirmed and accounted for, not
ignored), applied via dollar-quoted SQL to sidestep any escaping risk on the
glyph/quotes, verified all four expected substrings present post-write with a
single query (`~`/`LIKE`, not another eyeball diff).

**Rerender queued, but found and left alone a pre-existing sibling item worth
recording:** `directory-index` already had a `page_rerender` item from this
morning's `claimed-first-ordering` rollout (`rerender-pages`, **no `reason`
key at all** — takes the assemble/re-staple branch, not the regenerating one).
Reasoned through whether this races my new `template_changed` item: it
doesn't, because assemble only re-stitches ALREADY-STORED section HTML rather
than regenerating anything, so whichever order the two run in, the reason-less
one is a harmless no-op and mine is the one that actually matters. Did not
cancel it — no need to, and cancelling a work item I don't own outright felt
like the wrong kind of intervention for a genuinely harmless collision.

Reported to `vetcomparison` for their artefact verification, as agreed
throughout this thread.

## 2026-09-03 — `bugs_open/445` handed off cleanly to a new owner

A new session ("bugs_open/445") picked up the layout-library-gap bug, checked
ownership first rather than assuming, and confirmed 445 is parked here (my own
NOTES/commits already said so). Handed off with two honest answers rather than
guesses: confirmed `git status`/`log` clean on both files they're about to
touch (`resolve_composition_layout_action.go`, `fork_theme_composition.go`) —
nothing of mine in flight; and on whether the weak-positive-match case was
DELIBERATELY excluded from `queueLayoutCandidateReview`'s trigger — said
plainly **I don't know**, having actually checked (grepped the register and
the code comments near `IsFallback`/`IsSchemeMismatch`, found nothing), rather
than assert an answer to sound useful.

**Their reframe is worth naming here since it's stronger than my own bug's fix
candidate 1**: `queueLayoutCandidateReview` only fires on total mismatch or a
scheme gap — a positive-but-weak match (which is what produced 445's
convergence on `magazine-grid`) trips neither condition, so the mechanism that
exists specifically to surface "library is missing something" cannot see the
shape of gap 445 found. `[their MEASURED, 2026-09-03]`: one `needs_new_layout_candidate`
row, ever, across 38 resolved sites, and it's the zero-tags case (bug 431's own
territory), not a weak-match case. If confirmed, this is the better fix to
prioritise over 445's one-archetype proposal — routed there now, not chased
here.

## 2026-09-03 — misrouted `build-site-planner` finding, redirected

`gamedesign.uk` sent a genuinely good finding (the planner declining to plan
blog-post pages, citing a nonexistent "editorial producer") to this lane by
name-collision — the exact `site-design-planner`/`build-site-planner` mix-up
this thread flagged with `bugs_open/427` on day one and has held the line on
since. Not this mechanism, no code overlap. Redirected to `gamedesign.uk`
plainly, and forwarded the finding directly to the live `bugs_open/427`
session rather than leaving it to be found — it reads as the same shape as
their own `428` (the planner inventing authority to decline a page type),
just a different page type. Not investigated further here.

## 2026-09-03 — discriminator read: `template_changed` does NOT re-resolve query-backed sections

`vetcomparison`'s clean experiment answered the guardian's council advisory
from earlier: the claimed-chip rerender completed, redeployed, bytes changed
— and served **0 chips**, alphabetical order, 60 unclaimed. The NEW template
rendered over the OLD, build-time-resolved snapshot. **Confirmed: a
`page_rerender` with `reason='template_changed'` regenerates a page's HTML
from its component templates, but a QUERY-BACKED section (`directory-listing`
resolves `query.business_directory` at BUILD time into stored `content_data`)
keeps whatever it resolved at the last real build — a template edit alone
cannot refresh the underlying data, only how it's painted.** My template fix
itself is correct and shipped fine; the visible fix needs the OTHER route
(`needs_page` → `directory-build-handler`, which re-resolves at build), now
in flight on their side.

**Worth carrying here, not just in their lane**, because `directory-listing`
now exists on multiple sites and any future site-design-planner-adjacent
component edit could hit the identical trap: **before recommending a
`template_changed` rerender for a component fix, check whether the
component's template reads from a `query.*` source — if it does, the reason
you want is whatever triggers a real rebuild/resolve, not a template-only
rerender, however correct the template edit itself is.** This is the query-
backed-section analogue of the earlier lesson in this file about the AMBER
fallback needing all instances to REBUILD, not just re-render — same shape,
one level more specific.
