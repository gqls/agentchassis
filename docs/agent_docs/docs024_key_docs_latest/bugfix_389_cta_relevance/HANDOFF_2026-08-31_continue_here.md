# HANDOFF — 2026-08-31. **START HERE.** `bugs_open/391` — **ONE OWNER DECISION IS THE ONLY THING BETWEEN THIS LANE AND CLOSURE**

> **⚠ SUPERSEDED 2026-09-02 by `HANDOFF_2026-09-02_continue_here.md`.** The decision it was blocked on has been taken and applied.

**Supersedes `HANDOFF_2026-08-26_continue_here.md`.** Everything below was re-measured 2026-08-31
~15:0xZ; the 08-26 figures had gone stale by more than half.

> **Chassis `v1.0.1349`** (image tag, matching the `uk_001` overlay at HEAD). ⚠ **Do not verify the
> chassis sha by grepping the binary** — the all-zeros control MATCHES on this image, so both answers
> are worthless. Read the tag.
> **Nothing behaviour-changing has landed in the resolver, the rerender action, or the retraction
> since this lane's proof** — `git log --since='2026-08-26'` over those four files returns exactly one
> commit (`b1190467c`, additive). **The proven mechanism still holds.**
> **Nobody has worked this lane since 2026-08-26.**

---

## 1. ⭐ THE FINDING THAT CHANGES THE PLAN: **archiving drains the site by itself**

`[MEASURED 2026-08-31]` The remaining field population fell **41 → 25 with zero dispatch from this
lane** over five days. The drop is entirely on the two **archived** sites.

**Why, proven rather than assumed:** all three fields still pointing at the tool on archived sites
sit on components whose `updated_at` **predates their site's archive** —
`finetuning/ai-guides` (written 08-24, archived 08-26), `finetuning/guides/llm-cost-calculator-guide`
(08-17), `leopardess/blog/can-you-trust-ai-with-your-data` (15:04, archived 15:25 the same day).
**Every component rewritten *after* the archive has already moved off the tool on its own**, because
any rerender — fired for any reason — recomputes the CTA, and archiving is what releases KEEP #2.

**⇒ Step 3b does not need a 60-item dispatch.** Archive, and ordinary fleet churn does most of it.
That is the cheapest path and it is already proven in production.

## 2. ⛔ BUT THE DRAIN CUTS BOTH WAYS — and this is the decision

The recompute sends a field to whichever **tool** ranks first. Correct when the copy is about a tool;
**wrong when the copy asks the reader to get in touch.** A button reading *"Write to
leopardess@contactforsales.com"* that opens an ROI estimator is worse than an off-topic link, because
the copy makes a promise the destination cannot keep **in kind** — and it looks plausible, so nobody
reports it.

**The drain has already produced one such case on the archived sites since 08-26.** It is not
hypothetical and it is not finished.

`[RE-MEASURED 2026-08-31]` of the **25** fields still pointing at the tool on active pages:

| site | fields | contact-intent labels | no label |
|---|---|---|---|
| **ai-agent-orchestration.com** (NOT archived) | **22** | **19** | 2 |
| finetuning.uk (archived) | 2 | 1 | 0 |
| leopardessconsulting.co.uk (archived) | 1 | 0 | 1 |
| **total** | **25** | **20** | 3 |

**20 of 25 are contact-intent, and 19 of those sit on the one site not yet archived.** So archiving
`ai-agent-orchestration.com` today would start manufacturing wrong-kind buttons there, passively,
exactly as it has begun doing on the other two.

**⇒ THE DECISION (owner's):** what should a contact-intent CTA resolve to?
1. **`/contact.html`** — a `content_data` write per field, no LLM, immediate, and it is what the copy
   already promises. Cheapest and most conservative.
2. **Rewrite the copy** so the button is about the tool it points at — the framework's own job, but it
   is an LLM pass over ~20 fields and it changes voice on live pages.
3. **Per-site** — plausible if some sites want a contact CTA and others want a tool.

**Do not archive `ai-agent-orchestration.com` before this is decided.** The other two sites are
already archived, so their 3 stragglers are the standing evidence either way.

## 3. WHAT IS LEFT BEFORE THE LANE CAN CLOSE

| # | item | blocked by | effort |
|---|---|---|---|
| 1 | **Decide §2** (20 contact-intent fields) | **owner** | one decision |
| 2 | Apply that decision to the 20 | 1 | SQL, or one LLM pass |
| 3 | **Archive `ai-agent-orchestration.com`** | 1 | one guarded `UPDATE` |
| 4 | Let the drain finish, or nudge the ~5 stragglers | 3 | `cta_links_stale` rerenders, no LLM |
| 5 | **Retract** the three pages (`page-retraction` agent) | 4 — it refuses while any editorial link remains | one dispatch per site |
| 6 | **Refresh the aiao footer** — `nav-link-fixer`, then propagate in **assemble mode** (`page-rerender`, **no** `spec.reason`) | 5 | one run |
| 7 | Final sweep: zero `password-entropy` at the served bytes on all three sites, footer included | 6 | curl |

**Items 2–7 are mechanical and could be one working session.** Item 1 is the whole critical path.

**NOT required for closure — spin it out:** owner decision 3, the platform lever (the ranking has no
relevance input, so the next site inherits the same bug). That is architecture-scope, needs RFC_022
and a consumer enumeration, and holding this lane open for it just keeps a done thing open. **File it
as its own lane when 1–7 land.**

## 4. ⚠ Do NOT use the 399 audit as the census for §2

`CTA_LABEL_MISMATCH` is live and recording (**218 findings** as of 2026-08-31, 19 on these three
sites), and it looks like the right instrument. It is not. `[MEASURED 2026-08-31]` of all 218
findings, **exactly 1** is a contact-intent label pointing at a tool — because `JudgeCTALabel` is a
**page-identity** test and *"Write to leopardess@…"* names no page, so the judge has no opinion.
Its recorded verdicts are `no_opinion/ambiguous` **179**, `contradicts` **35**, `no_opinion/(none)`
**4**.

**Use the regex census in §2** (its query is in NOTES, dated), and hang any gate off
`SilenceNamesNothing` (`datahelpers/cta_label_agreement.go`, live at HEAD) — **in this lane's gate**,
calling `JudgeCTALabel` first and adding the intent arm on top. Do not widen the judge: that is the
drift RFC_047 §9 forbids, and the 399 lane built the reason-code seam specifically so we would not.

⚠ **Query the config key `audit_cta_label_agreement`, never the Go filename `cta_label_audit`** — the
filename census returns **false on all four writers, armed or not**, and reads as "the migration never
applied". `LANDMINES.md` (`cd6cb3cc5`) has it, with the control: **expect two true, two false** while
`645` is held (still held as of 2026-08-31).

## 5. SEQUENCING, if §3b is ever dispatched as a batch

`page-rerender` is one of only **two** armed writers, so a burst **dominates** the audit record rather
than sitting in it. **Name the window in NOTES with timestamps before starting**, and ping the
`bugs_open/399` lane — they have asked for it and will fold it in so the next reader need not
reverse-engineer a spike. Any rate read during that window is meaningless twice over.

## 6. Current state, verified 2026-08-31 ~15:0xZ

- `nav_order` 900 on all three tool pages — holding.
- **Label-locked fields: 0 fleet-wide** (was 20 of 80). Step 2 remains complete.
- finetuning.uk **archived**, leopardess **archived**, ai-agent-orchestration.com **active**.
- **All three tool pages still serve 200** — archiving freezes without unpublishing. Nothing is dead.
- Work items: 28 complete, 2 cancelled, **0 open**.
- Everything so far is reversible by setting `status` back to `'active'`.

## 7. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 391` · **chassis image tag**
(not a binary grep) · the §2 census (**re-run it — the count moves on its own**) · §6 state · then §3.
