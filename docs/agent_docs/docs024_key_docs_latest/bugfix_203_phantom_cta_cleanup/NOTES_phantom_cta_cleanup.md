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
