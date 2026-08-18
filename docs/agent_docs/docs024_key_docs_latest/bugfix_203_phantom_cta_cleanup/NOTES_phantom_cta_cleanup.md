# NOTES — bugfix 203 cleanup (append-only, newest at the bottom)

## 2026-08-06 — session start, claim, and the two cheap wins

- Triage: 206/207/205 (the newest opens) all actively claimed by other live sessions —
  verified in their transcripts (edits + task lists), not just `who-owns.py`, which is
  lagging by design. 203 unclaimed; took it. Claim committed `dfdcdecd2`.
- Council verdict for the source fix (`880a405a6`, corr `42eda9a5`): **APPROVED r1**,
  3 objectors, 5 advisory objections, none high. Read in full from
  `diagnosis_artifacts.body`. First query failed — guessed a `content` column instead
  of reading the schema; `\d diagnosis_artifacts` first, then it was `body`.
- Liveness: proven by ancestry against 197's pod-proven commit (see RUNBOOK). Then a
  fresh roll to **v1.0.1261** landed mid-session (confirmed both pods) — built from
  later HEAD, so carries the fix a fortiori.
- The census re-run and everything downstream still pending at this note.

Key discovery, pre-empting the council's own class-audit ask: `contextToMap`'s
DEFAULT VALUES map (component_library.go ~1136–1147 at `880a405a6`) still fabricates
`primary_cta_url: /contact.html` / `secondary_cta_url: /about.html`, and the alias
block copies `cta_url → primary_cta_url`. The 203 fix removed the `cta_url` default
but the `primary_*` family keeps the class alive on the regex-fallback path.
[UNMEASURED at this note]: whether any live template consumes `primary_cta_url`, and
what the regex renderer ships for an absent key (bug_historian M2 warns: possibly
literal `{{.field}}` text, which is WHY those defaults exist).

## 2026-08-06 (later) — P1 measurements: the class audit is done, and its answer is "inert"

**F1 — the bug file's census SQL cannot execute as written.** It joins
`sites s ON s.id=pc.site_id`, but `page_components` has **no `site_id` column**
(`\d page_components`; the join must go through `p.site_id`). So the recorded "13 rows"
was produced by some *other* query than the one written down. Logged in WRONG_CALLS.

**F2 — census drift on the original signature: 13 → 4** (cta_text present, cta_url
absent, an anchor in stored html). Of the 4, `robot-hands.com/how-to-specify-a-gripper`
now renders a correct first href (`/tools/gripper-safety-factor-calculator/index.html`).

**F3 — the original symptom page has self-healed THROUGH THE FRAMEWORK.**
`dartsonline.com/news/index.html` rerendered 2026-08-06 20:07:14Z on the fixed binary:
all three components carry no `/contact.html` at all and no `cta_url` key. This is the
cleanup path working, on the exact page that started the bug. [MEASURED]

**F4 — the class audit the council asked for (bug_historian M1) is COMPLETE, and exactly
two members remain**, both in `contextToMap`'s DEFAULT VALUES map:
`primary_cta_url → "/contact.html"` (line 1138) and `secondary_cta_url → "/about.html"`
(1140). Every other fabricated default in the file is cosmetic — colours, company_name,
and the `cta_text → "Get Started"` label. A fabricated colour cannot mislead a visitor
about where a click goes; a fabricated href can. That distinction is the class boundary.

**F5 — both remaining members are INERT in shipped data, and the check could have come
out otherwise** (the disconfirmability test, not just a marker):
- **39** stored components consume `.primary_cta_url` in their template AND carry a
  `/contact.html` anchor. **0** of those 39 lack an authored `primary_cta_url`. The
  population was non-empty, so the zero is a finding, not an empty-set artefact.
- **0** rows carry an `/about.html` anchor with no authored `secondary_cta_url`.
- Blindness closed: **43** `page_components` rows (3.4% of 1,247) have a NULL
  `component_id` and were silently dropped by my template join — checked separately,
  **0** of them carry a contact anchor.

**F6 — bug_historian's M2 objection is CONFIRMED IN CODE, so the naive fix is a
regression.** `renderGoStyleSubstitutions` (line ~1730) returns `match` — the literal
`{{.field}}` — when a key is absent. So deleting a URL default on the regex path ships
literal template syntax *inside an href*. Proven live: `idea.uk/tools/ab-test-calculator`
stores a literal `{{.section_heading}}` (1 row fleet-wide). **Do not delete 1138/1140
without first changing what an absent key renders as.**

**F7 — `RenderTemplateWithValidation` is dead code**: no non-test callers anywhere in
`platform/`. So `contextToMap` is reached ONLY through the regex fallback at
`component_library.go:989`, when `executeGoTemplate` errors.

**F8 — [UNMEASURED, and I nearly recorded it as measured] the fallback's firing rate.**
`kubectl logs --since=24h | grep -c "using regex fallback"` returned 0 on both pods — but
the pods started 19:54Z and I ran it at 20:19Z, so that was a **25-minute** window, not
24 hours. It is worthless as evidence of absence. Logged in WRONG_CALLS. The durable
substitute is F6's whole-population placeholder scan.

**F9 — the real backlog is the queues, not the 13.** `cta_names_unknown_destination`:
**123** rows at `needs_human_review`; `unresolved_cta`: **26** at `needs_human_review`
(latest 08-04/08-05). That is candidate 2's population and it dwarfs the census.

**F10 — the genuinely MISMATCHED survivors**, by extracting each anchor's own label
rather than trusting `cta_text` (label promises an action, href goes to contact):
`finetuning.uk/guides/tool-ai-data-risk-checker-guide` "Run the Risk Checker" ·
`leopardessconsulting.co.uk/who-we-help` "Score your process" ·
`robot-hands.com/how-to-specify-a-gripper` "Run MatchMatrix" ·
`finetuning.uk/about` "How We Work". Plus **4** leopardess blog heroes labelled
"Get Started" — the *fabricated* label default — pointing at contact (destination
plausible, button never authored). The other ~13 contact anchors are legitimately
contact-directed ("book a discovery call", "Talk to us", "Start an enquiry") and mostly
sit in `article-body`/`generic-text-block`, whose templates consume **no** cta_url key —
so they are not this bug's mechanism at all. **`/contact.html` exists on all 7 sites**,
so none of these is a broken link; the defect is label↔target mismatch.

### Consequence for the code half

No code change is warranted this session, and that is a result rather than a shortfall:
the class has two members, both inert (F5), and the only cheap fix available is a
regression (F6). The door-closing shape is recorded in the PLAN as P2 — it is a change
to what the regex path does with an absent key, or deletion of that path outright (F7
makes the latter thinkable), and either is a guarantee change on shared rendering
plumbing that needs its own measurement and council round. Not something to bolt onto a
cleanup session.

## 2026-08-06 (later still) — the detector's premise is REFUTED, and my own trailer misstep

**F11 — `bugs_open/203` candidate 3 is WRONG, and this is worth correcting in place.**
The bug file says `check_misdirected_cta` "clearly isn't running frequently/broadly enough
to have caught 11 of 13". Measured: `cta_names_unknown_destination` holds **123** items
across **10** distinct sites, filed as recently as 08-05 (robot-hands 29, ai-agent-orch 22,
leopardess 20, finetuning 18, gaswholesalers 12, idea.uk 8, vonc 7, fundamentallyai 4,
relojistas 2, gamesdesign 1). The detector runs, broadly, and files plenty. **Under-running
is not the defect.**

**F12 — the actual defect is that nothing drains what it files.** All 123 sit at
`needs_human_review` with `handler_agent = ''` (empty string, the column's default — not a
handler that failed, a handler never assigned). That is `bugs_open/083`'s class
("detected findings never reach a handler"), arriving on this queue. Candidate 3 should be
re-aimed from "make the detector run more" to "give its output a handler", which is a
different bug and probably not this file's to fix.

**F13 — but the detector genuinely missed my four mismatches, for a structural reason.**
Searching its items for the four labels returns exactly one row, and it is a *different*
page (`idea.uk/about`, info-card-grid, "How we work →", filed 07-18). Reading the
predicate: `ctaClassifyAnchor` (`check_misdirected_cta.go:164`) **returns early unless
`bestPageMatch(tokens, pages)` finds a real page whose tokens match the anchor text.** So
an anchor is only ever flagged when the checker can already name a better destination. A
CTA whose promise matches no page in the index — or whose match is filtered out of the
index upstream — is invisible to it by construction, not by scheduling. Note also
`bugs_open/185` ("detectors select deployed and miss 28 live pages"), which would bite the
same check independently. **Not chased further here: that is the detector's own bug, and
185/083 are other files' territory.**

**F14 — my own misstep, logged rather than quietly dropped.** I put
`Council-Reviewed: 42eda9a5` on commit `eba83792e`, which is **docs-only**. The verdict I
read is real and I read it in full, so it is not a fabricated verdict — but it reviewed the
*code* change (`880a405a6`), and the gate refuses docs submissions client-side, so a
docs commit can never have a verdict of its own. The trailer's whole purpose is an exact
commit↔verdict join for the `098` coverage report, and I have just fed it a join that
credits a prose commit with a code review. Forward-only, so `eba83792e` keeps the trailer
and this note is the correction. **The rule I should have followed: a docs commit needs no
trailer at all** — `Council-Reviewed:` belongs on the commit carrying the reviewed code,
and `880a405a6` (which carried `Council-Submitted:`) is the one the report resolves
automatically once the verdict turned approved. Added to WRONG_CALLS.

## 2026-08-07 — the gate measurement lands, and the cleanup route is NOT what the handoff assumed

**F15 — the owed measurement (D4's gate) is now TAKEN, with both controls.** Yesterday's
attempt was worthless because the pods were 25 minutes old (F8). Today the same two pods
have **5h24m** uptime (started 2026-08-06T19:54Z, measured 2026-08-07T01:18Z), same image
v1.0.1261. Over that window, `"using regex fallback"` appears **0 times on both replicas**.
Two controls make that zero mean something, and I checked them *before* believing it:
- **rendering happened**: 11 `page_components` rows have `updated_at` inside the window
  (54 in 24h), so the render path was exercised;
- **Warn reaches me**: 34 `"level":"warn"` lines in the same window, and the fallback logs
  at Warn — so the level is not being swallowed.
`"RenderTemplate"` appears 0 times, which is consistent: the surrounding lines are Debug.
Combined with yesterday's durable population bound (1 literal-`{{.` row in 1,247 stored
components, and the Go path cannot produce that literal because it strips `<no value>` to
empty), the regex fallback is **rare-to-never in practice**. 11 renders is a small
denominator — it cannot separate "never" from "1 in 50" — so the honest statement is a
bound, not a zero.

**Consequence: D4's two candidates are both LOW priority, confirmed.** Neither member of the
class can fire on a path that does not run. This does *not* license deleting lines
1138/1140 — F6's reason still stands, and now there is even less reward for the risk.

**F16 — the worklist is STABLE overnight.** All 8 rows still present, `updated_at`
unchanged (07-25 → 08-05). Nobody else touched 203 overnight (`git log` on the bug file,
the lane dir and `component_library.go` since 21:00Z: empty).

**F17 — the real targets EXIST, so a bare rerender is the WRONG repair.** This is the
finding that overturns the handoff's own method note:
- "Run the Risk Checker" → `finetuning.uk/tools/tool-ai-data-risk-checker.html` **exists**
- "Run MatchMatrix" → `robot-hands.com/tools/matchmatrix/index.html` (+ `/matchmatrix.html`) **exists**
- "Score your process" → `leopardessconsulting.co.uk/tools/process-automation-scorer/index.html` **exists**
A rerender on the fixed binary would apply correct-or-absent and **delete** all three
buttons, on pages whose whole purpose is to send the reader to that tool. Honest, and a
usefulness regression. The four "Get Started" blog heroes are the opposite case — label and
destination both fabricated, nothing to preserve.

**F18 — and the resolver CANNOT be run on its own, which is why the handoff's D3 was
unbuildable as written.** `internal-link-resolver` is a **pure function**: its workflow is
`resolve_links → complete`, returning `sections_ready`/`unresolved` to its caller, writing
nothing. Fleet-wide, exactly **one** agent references it — `page-content-writer`, as
`spawn_link_resolver → resolve_links (call_agent) → prepare_link_context →
build_render_context`. So link resolution is a **content-writing-time** step, not a repair
one. "Re-run resolve_internal_links for the page, then rerender" names an operation the
framework does not expose.

**F19 — and there is no existing repair path for a MISDIRECTED link, only for a DEAD one.**
`component_link_repair.go` / `repairSectionLinks` is dead-internal-link repair — its own
header says so, and it explicitly does **not** touch `content_data`. Our hrefs point at
`/contact.html`, which **exists on all 7 sites**, so that machinery correctly no-ops on
every one of the 8. Nothing in the estate re-points a link that is live but wrong.

**F20 — a warning about F12's "just give the queue a handler".** Listing all 18
`cta_names_unknown_destination` items for finetuning.uk (filed 08-03): they are dominated by
**correct** CTAs — "Get in Touch", "Talk to Us", "Start a Conversation" — flagged under
"lands in an excluded area (contact/legal/about)". A "Get in Touch" button pointing at
`/contact.html` is exactly right. So the queue's precision is poor, `affected_url` is
**empty on all 18**, and a handler that auto-applied `suggested_target` would **re-break
correct buttons at scale**. Yesterday's F12 ("give its output a handler") is therefore
premature as stated: precision first, then a handler. Recording this against my own earlier
note rather than leaving it to be inherited.

**F21 — CORRECTION to my own README claim of an hour ago, caught before acting on it.**
I wrote "I'd do this one without asking, it's clearly right" about rerendering the four
"Get Started" blog heroes. Then I sized it: those pages are **8–9 days stale** (oldest
components 07-29/07-30) and there have been **244 commits to
`platform/orchestration/actions/` since 2026-07-29**. A rerender does not apply *my* fix —
it applies all 244 changes' worth of behaviour to a live customer page, which is the
`a-stale-page-holds-every-improvement-since-it-rendered` hazard exactly, now with a number
on it.

**And the reward is the smallest of the eight.** These four are the *least* harmful rows in
the worklist: the label is fabricated ("Get Started") but `/contact.html` is a **plausible**
destination, so today's visible defect is a generic button, not a promise broken. The three
tool CTAs are the ones that actually mislead — and they are the ones a rerender would make
worse by deleting (F17).

**So the risk/reward inverts what I asserted:** the least harmful instances carry the
largest unaudited blast radius, and the most harmful ones cannot be fixed by the cheap route
at all. **Nothing dispatched.** PLAN D7's recommendation is amended accordingly: option (1)
is no longer "do it unasked" — it wants the same two-page canary-and-diff discipline as any
other rerender on this estate, and it is not urgent enough to spend that on ahead of the
owner's call on the three that matter.

## 2026-08-07 — route 2 AUTHORISED by the owner, and the canary is dispatched

Chassis rolled to **v1.0.1262** overnight (both pods, started 05:47Z).

**Pre-flight, all four checks passed before anything was queued:**
- `load_current_section_content` **is wired into the live `page-build-handler` workflow**
  (`load_current_section_content → spawn_content_writer`), read from `agent_definitions`.
- The **binary carries it**, both replicas, with a negative control:
  `load_current_section_content`=2, `edit_live`=4, `zzz_no_such_symbol_203`=**0**. So this is
  the pipeline being proven, not my spelling.
- The channel has **already completed real work**: 7 `content_rewrite` rows fleet-wide with
  `spec.mode='edit_live'` and status `complete`. My earlier "maturity unverified" caveat is
  discharged.
- **The defect is live on the SERVED page**, not merely stored — fetched
  `https://finetuning.uk/guides/tool-ai-data-risk-checker-guide.html` (29,396 bytes) and it
  carries `<a href="/contact.html" class="btn btn-primary">Run the Risk Checker`.

**Canary chosen deliberately: `finetuning.uk/guides/tool-ai-data-risk-checker-guide.html`.**
It is the strongest of the four mismatches (a tool CTA whose target provably exists) and, at
**2 days stale** (rendered 08-05), it carries far less rerender blast radius than the 8–9-day
blog pages F21 warned about. Before-state pinned in
`SQL_2026-08-07_canary_cta_repair_finetuning_risk_checker.sql`: hero 2908 B /
`dd767cfb…`, article-body 12440 B / `b958d624…`, call-to-action 2443 B / `b2d1e81a…`; the
hero's `content_data` holds `cta_text` and **no** `cta_url`, confirming the href was
fabricated at render time and is not recoverable from stored data.

**Dispatched 08:09:46Z** — `content_rewrite` / `mode=edit_live` / `status=triaged` /
`handler_agent=page-build-handler`, then `build-dispatch-loop` fired at the site by hand
(it is scheduler-driven with a fixed `system.internal` input, so it never fires for a real
site on its own). `kcat -P` exit 0 was **not** treated as evidence: verified at the DB —
item `claimed` by `build-dispatch-loop` at 08:09:46.18Z, child orchestration
`c5a254b8` at `spawn_content_writer` by 08:10:07Z, then `call_content_writer`, then a
research-agent spawn at 08:10:58Z. The chain is running.

**The instruction is deliberately narrow**: set the hero's `cta_url` to
`/tools/tool-ai-data-risk-checker.html` "exactly as written", keep the existing label, and
change no prose — with an `acceptance_test` naming the other two slots as needing to be
unchanged. The URL is derived from a real `pages` row, and the framework writes the copy;
naming an exact URL in `suggestion` follows `create_tool_cross_link_items.go`'s own
precedent, so this is not hand-authored content.

### The full repair worklist, with ids, verified targets and STALENESS (2026-08-07)

Every target below is a real `pages` row, checked live. Staleness matters as much as the
target does — F21's lesson — so it is a column, and it sets the order.

| # | site / page | slot | label | verified target | last built | stale |
|---|---|---|---|---|---|---|
| 1 | finetuning.uk `/guides/tool-ai-data-risk-checker-guide.html`<br>`856e2b44-49e1-4abb-a1eb-13df784d1f32` | hero | Run the Risk Checker | `/tools/tool-ai-data-risk-checker.html` | 08-05 | **2d** ← CANARY |
| 2 | finetuning.uk `/about.html`<br>`c0c68034-469f-420c-90bd-d3c0fc0e13d2` | content-block-about | How We Work | `/how-we-work.html` | 08-03 | 4d |
| 3 | robot-hands.com `/how-to-specify-a-gripper.html`<br>`5a385981-c2fd-4edb-bc4d-927b93177281` | hero | Run MatchMatrix | `/tools/matchmatrix/index.html` | 08-02 | 5d |
| 4 | leopardessconsulting.co.uk `/who-we-help.html`<br>`3e480330-d2b3-4d08-951a-a4e4804a90da` | hero | Score your process | `/tools/process-automation-scorer/index.html` | 07-25 | **13d** ⚠ |

Site ids: finetuning.uk `1368e337-dd1d-4799-bbb3-8221a1b79bcc` · robot-hands.com
`00ff3af5-dad8-4770-9f70-3edc267a3c92` · leopardessconsulting.co.uk
`4851f6fc-71cf-4160-a270-e03d6d3e0732`.

**Order: 1 → 2 → 3, and 4 LAST and separately.** Row 4 is 13 days stale, so an `edit_live`
pass over it carries the largest unaudited blast radius of the four (F21) — it should not
ride along on a batch, and it wants its own before/after diff and a look at the served page.
**Nothing is dispatched beyond the canary until the canary verifies**; that is what a canary
is for, and batching them would forfeit it.

The four `leopardessconsulting.co.uk` "Get Started" blog heroes are **NOT** in this table and
are still parked per F21: fabricated label, plausible destination, 8–9 days stale — the worst
risk/reward ratio of the eight.

## 2026-08-07 08:15Z — THE CANARY EARNED ITS KEEP: the resolver mis-assigns CTA targets, and the writer cannot set a cta_url at all

Read live out of the in-flight canary's `collected_data` (orchestration
`a9e8e280-f937-4a3c-bb32-8949e6c07101`, step `resolve_links` → `response.sections_ready`,
hero section). I enumerated the section's keys rather than path-reading for `cta_url`,
which is the only reason I found it — a top-level `s->>'cta_url'` returns NULL and would
have read as "the resolver did nothing":

```
hero.resolved_data = {
  "cta_url":                    "/tools/password-entropy.html",
  "cta_target_title":           "Password Strength Physics",
  "secondary_cta_url":          "/tools/tool-ai-data-risk-checker.html",
  "secondary_cta_target_title": "AI Data Risk Checker | Tools",
  "background_image":           "/assets/images/hero.jpg"
}
hero.llm_fields = ["subheadline", "secondary_cta", "cta_text", "headline"]
```

**F22 — the resolver SWAPPED the two CTAs.** The primary button is labelled
**"Run the Risk Checker"** and it resolved to the **password-entropy** tool; the secondary
is labelled **"Speak to Us About Data Privacy"** and it got the **risk checker**. The
correct target for the primary was available and the resolver had it in hand — it put it in
the other slot. So this page's CTA is about to stop pointing at `/contact.html` and start
pointing at a password-strength calculator, which is **worse**: `/contact.html` was at least
a generic plausible destination, whereas this is a non-sequitur that looks deliberate.

**F23 — and this is the structural half: `cta_url` is NOT a writer field.** `llm_fields`
lists only `subheadline`, `secondary_cta`, `cta_text`, `headline`. URLs live in
`resolved_data`, which the **resolver** owns. Consequences, both important:
1. **My work item's instruction was unobeyable.** I told the writer to "set the hero's
   `cta_url` to /tools/tool-ai-data-risk-checker.html — use that URL exactly as written".
   The writer cannot write that field at all. The `create_tool_cross_link_items` precedent I
   copied works because it asks for a link **inside prose** (an LLM field); a *structural*
   CTA URL is a different thing wearing the same name, and I did not check before copying.
2. **`bugs_open/203`'s candidate 1 is not achievable by asking the writer either** — "resolve
   the real target from the CTA text and set a real `cta_url`" is resolver work, not writer
   work, on every page with a structural CTA.

**So route 2 does not fix this class.** It replaces a fabricated destination with a
mis-resolved one. The defect to fix is the resolver's slot assignment.

**[UNMEASURED] how wide F22 is.** One page, two CTAs, observed once. It could be a
label-matching failure, a greedy/ordered assignment that ignores the label, or specific to a
site with many similarly-named tools (finetuning.uk has 8 `page_type='tool'` rows). **Do not
generalise from this single observation** — it wants the resolver's assignment code read and
a fleet census of `resolved_data.cta_target_title` against the adjacent `cta_text`. That
census is the natural next step and is cheap, because both values are persisted per run.

## 2026-08-07 08:17Z — CANARY RESULT: honest page, preserved prose, and the button GONE not re-aimed

Work item `complete` at 08:16:40Z, `attempt_count=0`, no error. Verified at all three layers.

**F24 — what actually changed, measured against the pinned before-state:**

| slot | before | after | verdict |
|---|---|---|---|
| hero | 2908 B `dd767cfb` | 2836 B `ab78a790` | changed — **all anchors removed** |
| article-body | 12440 B `b958d624` | 12440 B `b958d624` | **byte-identical** |
| call-to-action | 2443 B `b2d1e81a` | 2443 B `b2d1e81a` | **byte-identical** |

**Served page** (`https://finetuning.uk/guides/tool-ai-data-risk-checker-guide.html`,
29,396 → 29,324 B):
- `<a href="/contact.html" class="btn btn-primary">Run the Risk Checker` — **GONE**. ✅
- `"Run the Risk Checker"` no longer appears anywhere on the page (grep = 0).
- `password-entropy` = **0** — the resolver's mis-aimed URL **never reached the page**. ✅
  Note this was luck, not a control: nothing stopped it except `cta_url` being absent from
  `content_data`, which is the same accident that removed the button.
- The `/contact.html` links that remain are chrome (nav "Contact", header "Get Started"), not
  this defect.

**So the scorecard: two of three goals.** The page stopped lying, and `edit_live` genuinely
protected the prose — both other sections byte-identical, and the hero's headline and
subheadline unchanged, which is exactly what the acceptance test asked for and a real
endorsement of 178's channel. **But the button was deleted, not re-aimed** — the same visible
outcome a bare rerender would have produced, for the cost of an LLM pass over three sections.

**Small content loss to be honest about:** the hero's `secondary_cta` went from
`"Speak to Us About Data Privacy"` to `""`. `cta_text` was retained
(`"Run the Risk Checker"`) but renders nothing, because the template guard needs both text
and URL.

**F25 — and I did NOT get to blame persistence, because I checked.** My first instinct was
"resolved URLs are never persisted, so every rerender loses them" — a tidy explanation for
the whole 203 class. It is **false**: fleet-wide, **129 of 1,247** components DO carry
`cta_url` in `content_data` (and 90 carry `primary_cta_url`). So persistence happens; it did
not happen *here*. **[UNMEASURED] why** — the plausible reading is that `cta_url` is absent
from `llm_fields`, so the writer never emits it and the save has nothing to store, while
`resolved_data` only ever feeds the render context. That is a hypothesis, not a finding, and
the 129 rows are the counter-example that has to be explained before anyone acts on it.

### Consequence: STOP the per-page dispatches (worklist rows 2–4 NOT dispatched)

Route 2 does not achieve the stated goal, so repeating it three more times would spend three
LLM passes to delete three more buttons — including on a 13-day-stale page. **The defect to
fix is upstream: the resolver's slot assignment (F22), and the `resolved_data` → `content_data`
gap (F25).** Rows 2–4 stay parked with their ids and verified targets in the worklist table
above, ready for whoever fixes that.

## 2026-08-08 — F22 answered from the code, and confirmed live on a second and third site

Picked this bug up (`who-owns.py` correctly named this lane; contributing here rather than
competing). Answered the 08-07 handoff's "start here": **is the target chosen by matching the
CTA's own label, or by position/order? — Read `chooseCTATargets`
(`resolve_internal_links_action.go:319-350`) end to end: it takes `pageName` only to exclude a
page's own URL from its candidate list (line 326, "don't point a page's hero at itself") — that
is the ENTIRE use of anything page-specific. Every other line ranks candidates by
`(NavOrder, Name)` and returns `ordered[0]`/`ordered[1]` as primary/secondary. `cta_text` /
`secondary_cta` — the button's own label — is never read by this function or anywhere in its
call chain. So F22 is not a bug in matching logic gone wrong; **there is no matching logic**.
Confirms the file's own doc comment (line 22-24): "agent boundary lets this be upgraded (LLM
intent-matching...) without changing callers" — that upgrade has not happened. v1 fills
primary/secondary with the site's top nav-ranked hubs, full stop, regardless of what the
button claims to do.

**This makes F22 systemic-by-construction, not [UNMEASURED]-width.** Whether it's *visible* on
a given page depends on whether that page's CTA label happens to be specific enough to name a
destination the position-based pick disagrees with — generic labels ("Learn More") never
expose it; named ones always can. So the right question isn't "how wide" but "how many live
CTAs are named specifically enough to be checkable" — and I did that census the cheap way (F25's
129 `page_components` rows that already persist `cta_url`, i.e. **live, published pages** —
better ground truth than ephemeral `orchestration_states` history, which turned out to hold
only the resolver's OWN output snapshot, pre-writer, so `cta_text` isn't even present alongside
it there — a dead end for this specific census, noted so nobody re-tries it).

**Confirmed on real, currently-published pages, at least 2 sites beyond finetuning.uk:**

| page | slot | label | got | should be (label-implied) |
|---|---|---|---|---|
| `/guides/ai-agent-roi-estimator-guide.html` | hero primary | "Calculate Your ROI" | `/tools/password-entropy.html` | an ROI estimator tool |
| `/privacy.html` | hero primary | "Read the policy" | `/tools/password-entropy.html` | the privacy page itself, or nothing |
| `/gripper-catalog.html` | hero secondary | "Read MatchMatrix methodology" | `/tools/gripper-payload-calculator/index.html` | a methodology page |

`password-entropy.html` recurring as the wrong target across unrelated pages/sites is the same
shape as the finetuning.uk canary (F22) — not a coincidence of one page, a property of whichever
site has that tool ranked early by `(NavOrder, Name)`. Query: `page_components.content_data ?
'cta_url'` (129 rows fleet-wide, F25's set) — cheap to re-run, no orchestration history needed.

**What this means for the open questions:**
- F22 is answered: **not a bug to patch, a capability that doesn't exist yet** (label/intent
  matching). D7 option 3 ("build the missing capability") was already naming this correctly;
  it is not a smaller job than that framing suggested.
- F25 is separately still open — WHY 129 rows persist `cta_url` in `content_data` when the
  resolver's contract writes to `resolved_data` only. **[UNMEASURED, not attempted this pass]**:
  worth checking whether these 129 predate `resolve_internal_links` entirely (an older writer
  generation path that DID let the LLM emit a URL directly) rather than being resolver output
  that somehow persisted — the sample above shows plausible-looking real tool URLs, which reads
  more like "an LLM was once allowed to write this field" than "the resolver's output leaked
  into content_data by accident." Whoever picks up F25 next: check `created_at` on these 129
  against `880a405a6`'s deploy and against when `cta_url` was removed from any `llm_fields` list,
  if it ever was one.
- **Recommendation, not yet actioned**: this is now clearly a scoped feature addition (resolver
  gains label-aware/intent-based hub selection), not a bugfix — architecture-adjacent per this
  repo's own "guarantee change on a shared mechanism" bar, since it changes what every
  CTA-bearing page's resolved link is allowed to depend on. Did not start building it or dispatch
  further pages this pass — flagging the fork (build the capability vs. accept D7's narrower
  per-page options) as a decision worth the owner's or this lane's explicit call before more
  session-hours go in either direction.

## 2026-08-08 (later) — re-checked the worklist for a "bare rerender is safe" subset: there isn't one

Owner asked (via a different session) to clean up the remaining sites, scoped down to a
**safe subset only** — bare `page-rerender`, no LLM edits, nothing this lane flagged as
risky or parked. Checked every row against that bar rather than assuming the original 13
still all exist.

**3 of the original census have self-healed on their own** via ordinary site activity in the
3 days since filing (rerender on the now-fixed binary regenerates from `content_data`, so a
page that got rebuilt for any unrelated reason stopped lying for free): dartsonline.com/news
(fixed directly, prior session), and — newly confirmed now —
`leopardessconsulting.co.uk/who-we-help.html` (component id changed, hero rewritten entirely,
`updated_at` today, no anchor) is no longer phantom; not the same content as when the
worklist recorded it ("Score your process" is gone, replaced by "Find the work worth
automating"). Component ids drift on this estate (multiple sites' pages get rewritten
independently of this bug); re-verify by id before trusting a row, not by url alone.

**What's actually left, checked live just now, is exactly the two buckets this lane already
marked unsafe — nothing new, nothing safer:**
- **finetuning.uk/about.html** (`content-block-about`, label "How We Work") and
  **robot-hands.com/how-to-specify-a-gripper.html** (hero, "Run MatchMatrix") are still
  phantom. Both are real tool CTAs with a verified live target (worklist rows 2 and 3) — a
  bare rerender **deletes** an achievable correct button rather than fixing it (F17). Not a
  safe-subset candidate; the unsafe part is the outcome, not the staleness.
- **The 4 leopardessconsulting.co.uk "Get Started" blog heroes** are still phantom, still
  parked. Re-measured the blast radius F21 costed: **237 commits** to
  `platform/orchestration/actions/` since these pages last rendered (07-29/07-30), measured
  today — same order of magnitude as F21's 244, i.e. **the risk has not shrunk**, only aged
  three more days. No new information makes these safer to touch today than on 08-07.

**So: zero rows currently qualify as "phantom link, bare rerender clearly improves it, nothing
this lane flagged against it."** The set that's left IS the set that's flagged. Declining to
dispatch anything this pass rather than force a match to the brief — noted here so the next
session doesn't re-run this same census expecting a different answer without new information
(the resolver capability question from the entry above, or an owner sign-off on the two
flagged buckets, are the actual unlocks).

## 2026-08-08 (later) — the resolver capability is BUILT: label-aware CTA matching, calibrated against the live fleet

Answers this file's own open fork ("build the capability vs. accept D7's narrower per-page
options") by building the capability. Full design in
`architecture_review/` scoping (this session's plan, not yet a numbered RFC file — see
"what's owed" below) and the commits themselves; this entry is the evidence trail.

**What shipped, all behind mutation-proven tests:**
1. `check_misdirected_cta.go`'s existing token-overlap matcher (`ctaTokens`/`bestPageMatch`,
   proven at audit time since this lane's own earlier work) extracted to
   `datahelpers.LabelTokens`/`BestLabelMatch`/`NewLabelMatchCandidate` — one definition, not two.
   Behaviour-preserving: that file's own existing tests pass unchanged post-extraction.
2. `resolve_internal_links_action.go`'s `setCTAField` now tries a label match FIRST when the
   page has a currently-published label for this slot (a new small query,
   `loadExistingSectionContentData`, keyed on site+page name — content_data is otherwise empty
   at resolve time, confirmed by tracing the actual pipeline: the writer's prompt never even
   receives `resolved_data`/`cta_target_title`, so a "the writer will see the resolved target"
   design intent in this file's own doc comment was dead code, not a real coordination path).
   Falls back to today's positional pick unchanged when no label exists yet or it matches
   nothing real.
3. `applyCTARecompute` (the actual repair path `check_misdirected_cta`'s own
   `cta_links_stale` work item triggers) gets the same matcher. **This was a live, separate bug
   in the platform's own remediation loop**, found while tracing this fix: the function's
   old "authored link to a real, sensible destination — keep it" guard accepted ANY valid,
   non-excluded, non-self URL — including a misdirected-but-otherwise-fine-looking one, which
   is EXACTLY what triggers `cta_links_stale` in the first place. So the detector's own repair
   action could not fix what it was invoked to fix; it silently kept the wrong link and the
   next discovery pass re-flagged the same page forever. Now: a label match that disagrees
   with the currently-stored URL wins over the old "keep if valid" guard.

**Calibrated against the real shipping function over the live fleet before submitting**
(`CALIBRATION_2026-08-08_label_match_report.txt`, this dir — `cmd/ctacalibrate`, a throwaway
harness importing the actual `datahelpers` package via a `kubectl port-forward` to
`postgres-clients`, deleted before commit, not part of the platform build):
- 1,251 labelled CTA fields examined fleet-wide (the six `ctaFieldNames` components).
- 634 label-matched a real candidate (interactive/hub pages only — same pool
  `chooseCTATargets` itself offers; an early pass over ALL active pages produced a
  materially larger, unfaithful number and was corrected before trusting it — worth recording
  as its own small lesson: calibrate against the SAME candidate pool the shipped code uses, not
  "every real page", or the measurement answers a different question than the one asked).
- 315 would be newly resolved where nothing was stored before (a pure fill-in, no override).
- **162 would OVERRIDE an existing, different, valid stored URL — the risk-bearing case.**
  Spot-checked the full 162: dominated by clear improvements (exact tool-name matches — "Open
  Drop Rate Tuner" → the Drop Rate Tuner tool; "Read Guides" → the actual guides index), a
  handful of plausible-but-looser token matches (a "architecture" token sending a CTA to a
  complexity-estimator tool rather than a prose page — defensible, not obviously wrong), and
  one class of genuine false positive **found and fixed before shipping**: interrogatives
  ("what", "how", "why"...) were not in the stopword list, so "See What We Build" token-matched
  a page titled "What It Costs to Work With Us" on "what" alone. Added to
  `datahelpers.LabelStopwords`; re-calibration confirmed the specific case is gone and the
  matched-count dropped 662→634 (28 fewer spurious matches fleet-wide from this one
  word-class fix).
- **[UNMEASURED, noted rather than chased]**: a second, rarer collision class exists —
  a common preposition inside a label ("...about your use case") can coincide with a page
  literally named "About". Seen once in the pre-fix sample, absent from the post-fix run, but
  not independently isolated as fixed-by-the-same-change or just no-longer-triggered by a
  different removed match. Not adding "about" to the stopword list speculatively — that word
  legitimately appears in real distinctive labels too ("Learn about our process") and
  over-narrowing the stopword list is how a detector goes quiet on what it exists to catch
  (LANDMINES: narrowing past an invented false positive can make a rule inert). Flagged for
  whoever reviews the submission to look at directly rather than pre-emptively patched.

**What's owed, not done this pass:** the scoping plan itself recommends this go through
architecture review (a shared-mechanism guarantee change, not a point fix) alongside the
normal council gate — not yet submitted as of this entry. Two things named as explicitly OUT
of this round, for whoever picks them up: feeding `cta_target_title` into the content-writer's
own prompt (closes the coordination gap from the writer's side too, separate council
footprint), and a `repairSectionsBeforePersist` arm (alongside the existing
`RepairContentDataLinks`, LNK-028) so the 2 real-tool-CTA pages and 4 parked "Get Started"
heroes from this lane's own earlier cleanup pass could self-heal on their next ordinary save
without a per-page dispatch.

## 2026-08-09 — the submission never actually ran, and this lane did not cause it

Checked the verdict the next session-day rather than assuming silence meant pending. It
wasn't pending: `orchestration_states` for `258e4ed7-55a2-4280-a919-2713363c8b89`'s run showed
`COMPLETED | complete_invalid`, 16+ hours old. **Read `__step_error`, not `error`** (the
landmine already on file for this exact shape) — `review_editquality` got a real Anthropic API
400: `"Your credit balance is too low to access the Anthropic API."` **Not a defect in this
submission.** A different lane (`finetuning_uk_service`, commit `bc6c99cff`) independently hit
and documented the same fleet-wide outage the same evening (first credit failure 18:25:48Z,
zero successful anthropic `llm_call_log` rows after that until credits were topped up) — owner
already notified by that lane; nothing new to escalate here.

**Confirmed restored before resubmitting, not assumed**: `max(created_at) FROM llm_call_log
WHERE provider='anthropic' AND success=true` = 2026-08-09 14:58Z, minutes before this entry,
with a run of consecutive successes immediately before it — the fleet is genuinely processing
LLM calls again, not just accepting a request that will fail downstream.

**Resubmitted** on the SAME trail, per this repo's own resubmission pattern for exactly this
case (proven by the finetuning lane the same evening): `RESUBMIT_CORR=258e4ed7-55a2-4280-a919-2713363c8b89`
against the original, unedited submission JSON (still on disk in this session's scratch dir —
not rebuilt, per the standing instruction not to reconstruct a submission from memory). New
run orchestration `e1c497e0-2be0-4a1a-821c-446166404451`, dispatched cleanly (other real
orchestrations progressing normally in the same queue snapshot, a second confirmation the
fleet is healthy). **Verdict not yet read as of this entry** — next session: query
`diagnosis_artifacts` for `kind='council_report'` keyed on the submission correlation above,
per the standing verdict-reading instructions in the HANDOFF.

## 2026-08-09 (cold-start continuation) — verdict APPROVED, fix confirmed LIVE at the pod, and row 2 self-answered its own canary question (negatively, and structurally, not a fluke)

**Verdict**: `diagnosis_artifacts` for correlation `258e4ed7-55a2-4280-a919-2713363c8b89` /
`kind='council_report'` returned exactly one row — `decision='approved'`,
`created_at=2026-08-09 15:08:09Z`. Per the standing rule, no `Council-Reviewed:` amend on
`bd6e3320c`/`465e45531` — `098`'s coverage report resolves this automatically.

**Live-verified at the pod, not the tag**: both `agent-chassis` replicas run
`docker.io/aqls/agent-chassis:v1.0.1274` (already rolled by another session as part of a
larger fleet release — 17 services' kustomizations were mid-edit in the working tree when
this session picked the branch up). Positive-control grep, both replicas:
`strings /app/agent-chassis | grep -c BestLabelMatch` → 2/2, and
`grep -c loadExistingSectionContentData` → 6/6 (the pipeline-works control). **The fix is live.**

**Row 2 (finetuning.uk `/about.html`, `content-block-about`, label "How We Work") got
auto-repaired by the platform's own `cta_links_stale` remediation loop at 15:34:17Z — 26
minutes after the verdict landed, and NOT dispatched by this session** (orchestration
`7c02cb09-0524-40c4-a12f-066047a4af36`, `initial_request_data.input_data.spec.check =
'misdirected_cta'`). Result: `cta_label` is still "How We Work" but `cta_url` is now
`/tools/password-entropy.html` (`cta_target_title: "Password Strength Physics"`) — **not**
this row's own verified target, `/how-we-work.html`.

This is NOT a fluke or a matcher miss — it's structural, confirmed by reading the code rather
than assumed: `chooseCTATargets`/`candidatesFromHubs` (`resolve_internal_links_action.go:139-149,
338-356`) only ever offers two pools as label-match candidates — `loadInteractivePages`
(`page_type='tool'`/game) and `loadContentHubs` (`page_type='section-index'`). Checked live:
`how-we-work` is `page_type='content'` — in neither pool, so it was **never a reachable
resolver output**, positional or label-matched, before or after this fix. `password-entropy`
is `page_type='tool'`, i.e. a legitimate member of the candidate pool the old positional pick
already drew from — this fix changed *which* tool got picked when a label exists, it did not
(and structurally cannot, as built) reach a plain content page as a CTA target.

**This confirms rather than contradicts the 08-08 conclusion** that row 2 has no safe
automated repair path — it just got exercised live instead of staying hypothetical. It also
means the HANDOFF's step-2 hope ("a full rebuild ... may now produce the right link directly")
is **false for row 2 specifically**, and by the same reasoning **false for any row whose
verified target is `page_type='content'`**. **Row 3 (robot-hands.com, "Run MatchMatrix" →
`/tools/matchmatrix/index.html`) is different and still worth the canary**: checked live,
that target is `page_type='tool'` (site's page is literally named `tool-matchmatrix`) — inside
the candidate pool, so the label match has a real shot. Not yet dispatched or auto-repaired
as of this entry (page `updated_at` still 08-04, pre-fix). **The 4 leopardessconsulting.co.uk
"Get Started" heroes need the same page_type check on their own verified targets before
assuming either outcome** — not done this entry, do it before dispatching any of them.

**Not yet decided**: whether to manually dispatch a `cta_links_stale`-style rerender at row 3
as the deliberate canary, or wait and see if the same automated loop reaches it on its own
(as it just did, unprompted, for row 2). Left for the next step in this session or the next
session — see HANDOFF for the decision point.

## 2026-08-09 (same session, later) — row 3 canary DISPATCHED, mixed result: the primary CTA is fixed, and it EXPOSED a real scoring-priority bug in the matcher itself

This session asked the user to choose between dispatching now, waiting for the automatic
loop, or stopping; the user chose to dispatch now. Mechanism: the existing detector-created
`page_rerender`/`misdirected_cta` item for this exact page already existed
(`fab86424-0078-469b-b355-76c6a625b67e`, `status='detected'`, `approval_mode='auto'`), and its
own `suggested_target` already named `/tools/matchmatrix/index.html` — so no new item was
authored, just promoted: `UPDATE site_work_items SET status='triaged' WHERE id=...`, satisfying
`load_work_item_actions.go:650-652`'s dispatch predicate. Fired `build-dispatch-loop` at
robot-hands.com's site (RUNBOOK's kcat worked example) — claimed in <15s, orchestration
`107418a7-4412-456d-885c-f5534ec75866` → child `33c23067-6f31-4c45-ab4d-46201d3d79db`,
COMPLETED in ~15s total.

**The good part, and it's real**: `hero`'s "Run MatchMatrix" now resolves to
`/tools/matchmatrix/index.html` with `cta_target_title: "Run MatchMatrix | Gripper Selection
Tool | Robot-Hands.com"` — exactly row 3's verified target. **This is the first live,
end-to-end proof that the shipped fix does what it was built to do**, on the actual repair
path (`applyCTARecompute`), not just the calibration harness.

**The bad part, found by checking every field the same rerender touched, not just the one
being tested** — `call-to-action`'s `secondary_cta` ("Browse the Gripper Catalog") resolved to
`secondary_cta_url: /tools/gripper-cycle-time-estimator/index.html`, **not** a catalog page.
robot-hands.com has a real, valid candidate for this label:
`gripper-catalog-index`, `page_type='section-index'`, url `/gripper-catalog/index.html` — a
2-token overlap ("gripper", "catalog") against the label's tokens. The estimator page it
actually got is a 1-token overlap ("gripper" only). **This is not a fluke — it's the matcher's
own documented tie-break rule, read literally**: `datahelpers.BestLabelMatch`
(`platform/orchestration/datahelpers/label_match.go:133-136`) checks
`c.Interactive && !bestPtr.Interactive` **before** comparing overlap counts, so ANY
`page_type='tool'` candidate with just 1 overlapping token beats ANY hub candidate regardless
of how much better the hub's overlap is. The doc comment even states the intent plainly —
"interactive (tool/game) candidates beat non-interactive ones, **then** higher token
overlap" — but "then" here means *only among candidates of the same category*, which silently
demotes overlap-quality to a tie-break that can never fire across categories. **Calibration
likely saw this and mis-filed it**: the 08-08 entry's spot-check of the 162 override cases
noted "a handful of plausible-but-looser token matches" as acceptable noise — this is probably
that class, just not diagnosed as a category-priority bug at the time.

**Not yet fixed. Not yet reverted on this page either** — this session stopped to write it up
rather than patch the shipped, council-approved matcher unilaterally. Two shapes of fix, for
whoever picks this up: (a) compare overlap count first, fall back to interactive-preference
only on a genuine tie (matches the doc comment's stated intent); (b) same as (a) but weight
interactive candidates rather than hard-gate them. **This should go back through the council
gate as its own small round** (it's the same shared mechanism, same guarantee-affecting class
as the original 203 follow-on) — not silently patched into the last commit, and not bundled
into whatever ships row 3's remaining slots.

**Consequence for the rollout plan (HANDOFF step 2, "test on ONE page before assuming for all
six")**: the canary answered its own question two ways at once — yes, the matcher can produce
the correct primary-CTA fix live; and no, it is not yet safe to assume every field a rebuild
touches comes out right, because the priority bug above can silently downgrade an
already-correct or better-candidate secondary/tertiary CTA on ANY page with both a tool and a
hub candidate in play — which is most of this fleet. **Recommend holding the remaining
five pages (2 leftover + 4 leopardess heroes) until the priority-order fix ships and is
verified**, rather than treating this canary as a clean pass.

## 2026-08-10 — the priority-order fix is written, tested, mutation-proven, committed, and submitted to council

User chose "fix + resubmit to council" over the other two options offered (hold-only, or
just revert the one bad field). Shipped as shape (a) from the entry above: overlap count
compared first, interactive-vs-non-interactive only breaks a genuine tie — this is also
exactly what `TestBestLabelMatch`'s own pre-existing comment already said was intended
("Interactive pages beat content pages on **equal-strength** matches"), so the fix aligns
implementation to already-documented intent rather than introducing a new policy.

**Commit**: `3bc0486d7` — `platform/orchestration/datahelpers/label_match.go` +
`label_match_test.go` only (clean commit-scope block, nothing else swept in). Two new
tests: `TestBestLabelMatchOverlapBeatsCategory` (reproduces the live robot-hands.com case
directly, generalised) and `TestBestLabelMatchInteractiveTiesBreakToInteractive` (guards
the genuine-tie case so the fix doesn't overcorrect into always preferring hubs).
**Mutation-proven**: `git stash push` the fix, re-ran the new tests against the pre-fix
comparator — `TestBestLabelMatchOverlapBeatsCategory` FAILED as expected (returned the
tool page, not the hub), `TestBestLabelMatchInteractiveTiesBreakToInteractive` passed both
ways (it's guarding tie-break behaviour both versions already got right, not the bug
itself). Full package tests, `go vet`, `gofmt` all clean. Also re-ran the CTA-adjacent
tests in `platform/orchestration/actions` and `discovery_checks` (the detector and the two
write-time call sites) — all pass unchanged, including the one existing test whose name
sounds like it could conflict (`interactive_page_preferred_as_suggested_target` — checked
it: both its candidates have equal overlap count, so it was never exercising the buggy
cross-category comparison and needed no change).

**Submitted to the council gate**: `SUBMISSION_CORR=6cb8c72b-0abc-4eb6-b4d2-4cbf01eed515`,
run orchestration `76b19b7e-3127-41a6-a1ad-b32efcad5f9c`. Verdict not yet read as of this
entry — check per the standing pattern:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='6cb8c72b-0abc-4eb6-b4d2-4cbf01eed515' AND kind='council_report'
ORDER BY created_at;
```

**Process slip, logged in full at `WRONG_CALLS.md` 2026-08-10**: commit `3bc0486d7` was
made BEFORE the submission existed, so it carries **no `Council-Submitted:` trailer at
all** — not even the honest-but-unverified one. `098`'s report resolves purely by
scanning each commit's own message; there is no path-based or cross-commit fallback, and
forward-only forbids amending the message now. **So this commit will read as UNREVIEWED
in `098` forever, even once the verdict lands approved** — this NOTES entry (and the
SUBMISSION_CORR above) is the only durable link between the two. If the verdict comes
back APPROVED: do not try to retrofit `Council-Reviewed:` onto `3bc0486d7` — there is
nothing to amend it with. Just record the approval here and treat the fix as reviewed in
substance, unreviewed in `098`'s bookkeeping.

**Still held**: the remaining five pages (2 leftover real-tool-CTA + 4 leopardess
heroes) stay undispatched until this verdict lands and, separately, until someone
decides whether the already-wrong robot-hands.com `secondary_cta_url` (still pointing at
the cycle-time estimator as of this entry) gets a follow-up rerender once the fix is
live — it will not self-correct; the priority bug lives in the matcher, not in this one
page's stored data, but a fix landing does not retroactively touch pages that already
rendered under the old code.

## 2026-08-10 (later) — verdict APPROVED (07:49:20Z), but shipping it hit the whole-fleet-release constraint, not a build failure

Verdict confirmed approved. Went to build+roll per the HANDOFF's own step 2 and stopped
before touching anything live: `agent-chassis`'s pods are ALREADY running
`v1.0.1277` (`startTime` 2026-08-09 21:34-35Z), which **predates commit `3bc0486d7`**
(2026-08-10 07:43:58+01:00, i.e. 06:43:58Z) by ~10 hours — so that image does not, and
cannot, carry this fix.

**Checked before building anything, and this is the load-bearing finding**: the
working tree has an UNCOMMITTED bump of `makefile`'s `IMAGE_TAG` (committed value
`v1.0.1275` → working tree `v1.0.1277`) plus all 17 services' kustomization overlays
bumped to matching tags, uncommitted — and the live pods already match that
uncommitted `v1.0.1277`. That is the signature of an in-flight or just-finished
whole-fleet release, not this workstream's own doing. Per this estate's own established
correction (2026-08-03, memory `releases-are-whole-fleet-make-release`): a solo
`build-agent-chassis` + single-overlay deploy at its own tag **fragments the fleet** —
every other service's overlay would then point at a tag agent-chassis alone was never
built at. **Did not bump `IMAGE_TAG` or touch any kustomization file.** Building and
pushing an image at a NEW tag would itself be harmless (an orphaned pushed-but-
undeployed tag costs nothing), but doing so right now, mid someone else's uncommitted
release-in-progress, risks exactly the collision this rule exists to prevent — so this
session stopped and surfaced it to the user rather than guessing.

**Consequence**: step 2 of the HANDOFF ("build+roll agent-chassis yourself") was written
before this constraint was rechecked and should be read as superseded — the correct
next action is coordinating with the owner/an active release, not a solo build. Left
open for whoever continues this: once a fleet release lands AFTER this commit
(`3bc0486d7`, or later), verify at the pod exactly as planned (positive-control the
live behaviour via the robot-hands.com re-test in step 3, since this specific fix has
no new symbol to `strings`-grep — it only reorders an existing comparator).

## 2026-08-10 (evening) — PROVEN LIVE: the priority fix produces the correct hub match on the exact page that exposed the bug

Owner ran a fresh fleet build; `agent-chassis` pods came up on **`v1.0.1280`**,
`startTime` 2026-08-10 15:44:46Z / 15:45:06Z — **after** commit `3bc0486d7`
(2026-08-10 06:43:58Z), so the image can carry the fix. Tag+timing is not proof, and
this fix adds no new symbol to `strings`-grep (it only reorders an existing
comparator), so the verification is **behavioural**, at the artefact:

**Re-dispatched the same page** (`robot-hands.com/how-to-specify-a-gripper`) by cloning
its own `misdirected_cta` work item under a fresh `item_key` (`..._retest_160006`,
id `043f2f0b-721d-43d4-9740-b47f56dcfe20`) — canonical row shape copied from the
existing row via `INSERT ... SELECT`, so no hand-authored spec — then fired
`build-dispatch-loop` at the site. Completed in ~5 min.

**Result, `page_components` at 2026-08-10 16:01:08Z — the defect is GONE:**

| slot | label | resolved url | verdict |
|---|---|---|---|
| `call-to-action` | Browse the Gripper Catalog | **`/gripper-catalog/index.html`** | ✅ was `/tools/gripper-cycle-time-estimator/index.html` |
| `call-to-action` | Run MatchMatrix | `/tools/matchmatrix/index.html` | ✅ held (control — the already-correct value did not regress) |
| `hero` | Browse Gripper Catalog | **`/gripper-catalog/index.html`** | ✅ a SECOND instance, not previously examined |
| `hero` | Run MatchMatrix | `/tools/matchmatrix/index.html` | ✅ held |

**Both halves matter**: the 2-token hub match now beats the 1-token tool match (the
fix), AND the two already-correct tool matches were not flipped to hubs by it (the
overcorrection guard `TestBestLabelMatchInteractiveTiesBreakToInteractive` asserts in
unit form — here it is confirmed on live data). The `hero` slot's own
`secondary_cta` is a bonus finding: it had the same defect and was never in any
worklist, which is direct evidence the bug's blast radius was wider than the one field
this lane happened to look at — see the census question below.

**Now genuinely closed**: bugfix 203's worklist row 3 (both its CTAs), and the
priority-order defect the canary exposed. `3bc0486d7` is approved
(`SUBMISSION_CORR=6cb8c72b-0abc-4eb6-b4d2-4cbf01eed515`), live, and proven.

## 2026-08-10 (evening, later) — CORRECTION: "the pages heal for free" was WRONG, and the manual alternative turns out to be actively dangerous

> **CORRECTED 2026-08-10.** Earlier this same day I recommended, in this file's D2
> framing and to the owner directly, *"do nothing — each of those pages fixes itself the
> next time it's rebuilt for any reason."* I cited as evidence that 3 of the original 13
> pages self-healed. **The evidence was real and the inference was wrong.** Those three
> healed while the discovery/improvement schedulers were RUNNING. **The owner caught it:
> "the improvement loop is switched off."** I had never checked — the whole recommendation
> rested on a mechanism I assumed was live because I had watched its output weeks earlier.
> This is the `zero-adoption-means-read-the-mechanism` / "a silent mechanism is usually
> UNDRIVEN" lesson, and I walked straight into it while writing a recommendation.

**Measured `scheduled_tasks`, 2026-08-10 ~18:47Z** — the healing chain is broken at two
of its three stages:

| stage | task | agent | `enabled` |
|---|---|---|---|
| 1. detect | `site-discovery-rotation-completeness` (hosts `misdirected_cta`) | completeness-discovery-agent | **f** |
| 1. detect | `site-discovery-rotation-quality` / `-design` | quality-/design-discovery-agent | **f** |
| 2. triage (`detected`→`triaged`) | `improvement-sweep` | improvement-loop | **f** (last completed 2026-08-09 15:03Z) |
| 3. dispatch | `build-pipeline-trigger` | build-pipeline-trigger | **t**, 120s |

`TriageDetectedItemsAction`'s own header confirms the topology: discovery writes
`status='detected'`, the dispatch loop only selects `status IN ('triaged','approved')`,
and the improvement loop is what bridges them. With stages 1–2 off, **nothing files new
findings and nothing promotes the 192 existing `detected` ones** — so a page's wrong CTA
persists indefinitely unless a human drives it. Note stage 3 being ON is what let this
lane's own two manual dispatches work: promoting one row by hand was sufficient, which is
exactly why the loop being off was invisible to me.

**But the obvious remedy — promote the queue and let it heal — is the dangerous one, and
this is the more important finding.** Read `applyCTARecompute`'s keep-it guard
(`rerender_page_sections_action.go:686-691`) against `areasExcludedFromCTA`
(`resolve_internal_links_action.go:72-74` = `{about, contact, privacy, terms, legal}`):

```go
if hasCurrent && current != "" &&
    validPages.Contains(current) &&
    !ctaExcludedDestination(current) &&            // <-- /contact.html fails HERE
    NormalizePagePath(current) != NormalizePagePath(pageURL) {
    return // authored link to a real, sensible destination — keep it
}
```

A CTA already pointing at `/contact.html` **can never take the keep branch**. If its
label also matches no candidate — the normal case for genuine contact-button copy, since
`get`/`in`/`us`/`to` are all stopwords, so "Get in Touch" and "Talk to Us" reduce to
`[touch]` / `[talk]` and match nothing — it falls through to the positional pick, and a
**correct contact link is replaced by the site's top-ranked tool or hub.**

**This is not a defect in the code; it is the code doing bug 203's original job.** The
whole point of 203 was that `/contact.html` was a *fabricated* fallback that needed
recomputing. The problem is that a fabricated `/contact.html` and an authored one are
**byte-identical in `content_data`** — nothing distinguishes them — so the repair cannot
target one without hitting the other. That is the real reason the `misdirected_cta` queue
cannot be drained blindly, and it is a sharper statement of this lane's 2026-08-07
finding (which blamed the detector's *suggestions*; the damage actually comes from the
*recompute's fallback*, which fires regardless of what the suggestion said).

**Blast radius, measured 2026-08-10** (not estimated — the query is in the LANDMINES
entry added today): **24 CTAs fleet-wide currently point at an excluded area**, 7 of them
written during the 08-09→08-10 window when the priority bug was live. Those 24 are what a
bulk promotion would put at risk. Recorded as a LANDMINE (footprint
`applyCTARecompute` / `misdirected_cta` / `areasExcludedFromCTA`) and synced to
`doc_notes`, because it fires on a *future* session that reaches for the queue with no
symptom in front of it — which is precisely what I was about to do.

**Consequence for D2**: "let them heal for free" is not available, and "run the checkers
manually" is only safe with a per-page look at the label. Safe manual healing is
therefore: re-run **detection** (completeness-discovery-agent — it only files findings,
it changes no page), then repair **selectively**, excluding any component whose current
url is in an excluded area unless a human has confirmed the label is not genuine contact
copy. Do NOT run `TriageDetectedItemsAction` over a site to achieve this: it promotes
every `detected` row for that site with no type filter (its own header says so).

## 2026-08-11 — ran detection manually as agreed; it HALTED the repair rollout by exposing a third matcher defect

Owner: *"file a bug for that landmine it needs to be fixed. please go ahead with your repair
list."* Both done — `bugs_open/248` filed for the landmine, and the repair list was started
and then **deliberately stopped at the canary**, which is the whole point of having one.

**Bug 248 filed with a mechanical repro.** Took the package's own
`TestApplyCTARecomputeFallsBackWhenLabelGeneric` and changed exactly ONE variable, the
stored URL:
```
CONTROL  stored=/tools/password-entropy.html label="Get in Touch" -> resolved=map[]
CASE     stored=/contact.html                label="Get in Touch" -> resolved=map[cta_url:/tools/tool-ai-data-risk-checker.html …]
```
Control kept, case clobbered. The repro test was **run and then deleted rather than
committed** — it asserts a defect, and a passing test that enshrines wrong behaviour is how
a bug becomes a spec. Also added to `016b` §9 (four transferable rules) and the §10 index.
Prior art found: `bugs_closed/023:405-410` recorded this exact exclusion in its *benign*
direction ("makes some correct pairings unreachable"), named the right fix ("an
authored-intent escape hatch"), and closed without building it — the destructive direction
went unnoticed for three weeks.

**Detection, step 1 of the repair plan — dispatched ONE site first, robot-hands.com,
deliberately chosen as a positive control** because I had just verified its CTAs were
correct, so a clean result was the expected answer. Envelope: direct `orchestrate` at
`completeness-discovery-agent` with `site_id`/`domain` (the scheduled task's own `pre_query`
picks one site per 7-day rotation, so it cannot be used to target). Completed in ~3s.

**It did not come back clean — 16 pages flagged, including the control page.** Two things
came out of chasing that, and the second one stopped the rollout:

1. **Zero work items were created.** The fresh findings deduped against the existing open
   rows via `ON CONFLICT DO NOTHING` (`insertPageRerenderItem`). So the 192 `detected`
   `misdirected_cta` rows are **not stale** — they are the current findings, already filed.
   Re-running detection adds nothing while they stay open; there was never a fresh list to
   build.
2. **The control page's 3 findings are FALSE POSITIVES with a new, distinct mechanism —
   now `bugs_open/253`.** The anchors read *"Gripper Safety Factor Calculator"* → 
   `/tools/gripper-safety-factor-calculator/index.html`, which is the correct active tool
   page at exactly that URL; the detector wants to re-point them at the **payload**
   calculator. Reproduced with the real live rows:
   ```
   label tokens: [gripper safety factor calculator]
     gripper-payload-calculator              interactive=false overlap=4
     tool-gripper-payload-calculator         interactive=true  overlap=4
     tool-gripper-safety-factor-calculator   interactive=true  overlap=4
   BestLabelMatch -> tool-gripper-payload-calculator
   ```
   All three tie at 4, and `c.Name < bestPtr.Name` hands it to payload on the alphabet. The
   payload page earns its 4 from a **nav_label** that happens to read *"…Validate Capacity
   with Safety Factor…"*. `BestLabelMatch` counts how many of the LABEL's tokens a candidate
   holds and never how much of the CANDIDATE that is, so a verbose nav_label competes for
   every label on the site.

**Why this halts the repair rollout rather than just annoying us**: the repair recomputes
with this same function, so acting on those findings would have rewritten three *correct*
links to the wrong tool. **Detection is safe (it writes no pages); promotion and dispatch
are not, and are stopped** pending 248 and 253.

> **CORRECTION to this file's 2026-08-10 entry.** I wrote there that the priority fix meant
> "the matcher can produce the correct primary-CTA fix live". True, and too broad as I used
> it — I then reasoned as if the matcher were now *trustworthy*. It is not: it had (and has)
> two further defects, one of which I had already filed and one I had not yet found. The
> 08-10 verification was sound for the case it tested and did not license the generalisation
> I drew from it. **Three defects in one 40-line function, found only by exercising it on
> real pages — the unit tests pass on all three.**

**2026-08-11 — bugs_open/253 fix shipped, commit `f1819861f`.** `BestLabelMatch` now ranks
identity-token overlap (label tokens present in a candidate's own `name`/`title`) ahead of
total overlap (name+title+`nav_label`, the old and only signal); interactive-preference and
the alphabetical `Name` tie-break are unchanged. Council-Submitted:
`ccef36de-6757-4777-91db-37864b018622` (submitted before commit, per the standing rule;
verdict not yet read).

**Wrong turn, worth keeping**: the first attempt at this fix added a fourth ranking key —
smaller token-set size wins a tie, meant to replace the alphabetical final tie-break with
something that at least looks principled ("the more precisely-named candidate should win a
tie"). It passed its own unit tests and the first calibration pass looked like an
improvement. Only a second, harder look at the live numbers caught it: on
gaswholesalers.com it flipped all 9 already-correct CTA labels away from the right
calculator, because "Break-Even" tokenises to two words where the correct page's own title
uses the unhyphenated "breakeven" — one stray hyphen made the wrong candidate's token set one
entry smaller, and the new key handed it the win. The same shape recurred on
vetcomparison.uk (a shared generic word, "cma") and robot-hands.com (the word "run" sitting
in an unrelated page's own title, absorbing labels like "Run Payload Calculation"). The key
was dropped and the shipped ranking keeps the original alphabetical `Name` as the final
tie-break, unchanged from before this fix — identity-overlap is the only new key. Full
numbers and the regression traces:
`docs/agent_docs/docs024_key_docs_latest/bugfix_203_phantom_cta_cleanup/CALIBRATION_2026-08-11_label_match_identity_report.txt`.

Calibration (both candidate pools the shipping code actually uses, 784-row set): detector
pool 784 examined / 697 matched / 347 newly-resolved / OVERRIDDEN 208->205, 18 changed picks
(2.3%), each inspected individually; resolver pool 784/401/196/44, OVERRIDDEN unchanged
at 44, 0 changed picks (its `contentHub` candidates have no `nav_label`, so identity==total
there structurally — this fix cannot move that pool).

`bugs_open/253` stays OPEN — fixed-and-committed is not fixed-and-live until the next
fleet roll, per the standing bar. `bugs_open/248` (the recompute-destroys-authored-links
defect in the repair path's excluded-area handling) is untouched by this fix and remains the
other blocker on draining the `misdirected_cta` work-item queue.

**2026-08-12 — fleet rolled; 253 confirmed FIXED AND LIVE and CLOSED
(`bugs_closed/253_HANDOFF_...md`, commit `d9664e4dc`).** Two direct checks, neither inferred
from the roll announcement itself: (1) build provenance — the startup log line had scrolled
out of `--tail=3000` on both live pods, so fell back to probing `/proc/1/exe` directly:
extracted every 40-hex substring in one bounded `grep -aoE` (not a blind single-string
discovery grep — that lands on Go's internal digit tables and returns the same wrong answer
on every service, per the standing landmine) and cross-checked each candidate against real
commit hashes with `git cat-file -e`, since a spurious binary-table match cannot also be a
real, existing commit in this repo. Found `da5a7eb8f`, confirmed a descendant of the fix
commit `f1819861f` via `merge-base --is-ancestor`. (2) Live control — re-ran
`completeness-discovery-agent` on robot-hands.com directly (the scheduled rotation can't be
targeted); `how-to-specify-a-gripper` now carries zero `misdirected_cta` findings, down from
the 3 that opened this bug. 17 `misdirected_cta` findings remain fleet-wide on that run, but
all a different, unrelated site-content ambiguity (matchmatrix/matchmatrix-methodology/
how-it-works naming) — not a 253 recurrence, not chased further here.

First council run **failed on infrastructure, not content**: the top-level `error` column
(not `collected_data`, which held nothing — a second landmine-shaped trap, distinct from the
`__step_error`-vs-`error` one already on file, since here `__step_error` was itself empty and
the real message lived in the table's own `error` text column) read `reaper: stale
EXECUTING_STEP for >4h; step=review_bug_historian` — one seat hung and was reaped, no verdict
ever produced. Resubmitted under the same trail correlation
(`ccef36de-6757-4777-91db-37864b018622`, new run `101c11d9-50d7-4499-98b2-894138213094`);
verdict still pending when this entry was written — check it before treating this lane as
fully clear, closing the bug file does not retire that obligation.
---

## CONTRIB 2026-08-18 from `bugfix_248_authored_cta_destinations` — your CTA guarantee changed, commit `53a8d3c1d`

Telling you rather than only measuring that nothing broke (the 2026-07-29 owner ruling: a
shared mechanism's other consumers must be told). **Nothing in this lane's files was edited.**

**What changed for you.** `areasExcludedFromCTA` ({about, contact, privacy, terms, legal}) was
answering three questions with one answer. It now answers only the first:

1. a FRESH positional pick still never lands on a utility page — **unchanged**;
2. an ALREADY-STORED valid utility destination is now **KEPT** by both writers
   (`applyCTARecompute` and, newly, `setCTAField`), via `storedCTADestinationIsAuthored`;
3. the `misdirected_cta` check's "lands in an excluded area" arm still emits its finding but
   **no longer files a `cta_names_unknown_destination` work item**.

**The label-match branch is untouched and still runs first**, so a stored contact url whose
label names a real page is still recomputed — 248's verification bar #2, pinned by tests on
both writers.

**One thing you may have believed that was never true:** `candidatesFromHubs`' doc comment
claimed its inputs arrive pre-filtered by `chooseCTATargets`' `rank()`. `rank()` filters a
local copy and never mutates its inputs, and both call sites passed the raw loader output.
It now really does filter, which is what makes the derived-provenance invariant exact.

**⚠ The landmine this creates for anyone widening the candidate set** (register **LNK-033**):
the keep-branches rest on "no resolver path can emit a utility-area url". Widen
`loadContentHubs`/`loadInteractivePages`, drop `candidatesFromHubs`' filter, or add a
utility-area schema `fallback` to a `ctaFieldNames` component, and both writers start freezing
the resolver's own output — with the detector arm that would have noticed now demoted. That is
filed as `bugs_open/308`, which is routed here and must land with recorded provenance, not
before it.
