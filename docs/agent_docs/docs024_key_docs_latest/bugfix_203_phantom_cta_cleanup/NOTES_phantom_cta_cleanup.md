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
