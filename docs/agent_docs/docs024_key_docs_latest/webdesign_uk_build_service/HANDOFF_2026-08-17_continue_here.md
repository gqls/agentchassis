# HANDOFF 2026-08-17 — PLAN_2026-08-11 is 6/6 DONE and the 285 acceptance PASSED; the lane is now ON HOLD awaiting the owner's new site plan — SUPERSEDES HANDOFF_2026-08-16

**Start here cold.** Read order: this file → `REQUIREMENTS_2026-08-17_owner_brief_pending_new_plan.md`
(the owner's words, verbatim, and what each means here) → NOTES 2026-08-17 entries →
`bugs_open/299` → the owner's new plan when it lands.

## 0. State in one paragraph

**PLAN_2026-08-11 ("chat box as a framework capability") is complete, 6 of 6.** Step 5
(deployer backend half, TL-043 + the shared `sitechat` binary) and step 6
(tool-suggester cites the approved EXPERIENCE_PLAN, TL-046 / migration 449) both
landed and are recorded. The `bugs_open/285` lock-blind-assembler defect this lane
diagnosed was fixed by the separate lane, went live on `v1.0.1305`, and **this lane's
five-step acceptance PASSED on the real page** (2026-08-17 12:15–12:21Z); `a4cd5dc8`
is answered and closed; the bug file is now in `bugs_closed/`. **The chat-box lock is
STILL ON** — it comes off only on the owner's word. The lane is now **ON HOLD**: the
owner has briefed a new direction (chat onto the home page, prouder positioning,
corrected payment terms, a full site rewrite) and said explicitly *"Don't go ahead
until I've finalised that plan"*, which he is doing in the session
`webdesign live web builder project` (`d10f1acc-1627-4729-b660-93d6e84911e3`).

## 1. DO NOT START — the hold, and what it covers

Everything in `REQUIREMENTS_2026-08-17_owner_brief_pending_new_plan.md` is held:
the chat move, the exclusions/positioning copy, the payment-terms correction, and the
site rewrite. **Also do not surgically patch `bugs_open/299`** (below) — it lives in a
section the rewrite regenerates. What is NOT held: reading, measuring, and anything
that does not change this site.

**The one binding constraint on HOW, when the hold lifts** (owner, 2026-08-17, and it
matches the 2026-08-04 standing ruling): do it **through the framework — spec and
planner, the way a client edit would arrive** — not by surgical `UPDATE` of
`rendered_html`. This lane's habitual direct-SQL fix is the wrong tool for this job by
his instruction.

## 2. Owner decisions OUTSTANDING

1. **The new plan itself** (other session) — gates everything in §1.
2. **Why the payment sentence is wrong.** *"You see it first on a private preview
   link, and you pay in full only once you are happy with it."* The page is a faithful
   rendering of the live `evidence_base` facts (`payment_after_approval`, `no_refund`),
   so **the register must change before any page does** — correcting the copy alone is
   undone by the next rebuild (the `bugs_open/161` mechanism: the register causes the
   claim and then vouches for it). The true terms are owner copy: **ask, do not infer.**
3. **Does the chat-box lock come off**, and does `contact` keep a chat box once one is
   on `index`?
4. **The 2026-08-17 rebuild's copy changes** (now live): the hero gained `£149`, lost
   "we usually reply within a day or two", and `hero` + `contact-info` now BOTH open
   with "Get in touch". Backup retained (see §4) — restore, keep, or let the rewrite
   settle it.
5. Standing: contact email `webdesign@contactforsales.com` (domain mismatch, open
   `content_rewrite` `a8d6f440`); Stripe webhook hostname; Stripe keys via terraform.

## 3. ⚠ FLEET FACT that changes what "deployed" means right now

**The chassis is NOT running current HEAD, despite a fresh deploy.** Measured
2026-08-17 ~12:30Z with a well-formed probe (positive control `6a782274b` MATCH,
plausible-fake sha absent, current HEAD absent):

- pods are new (~92 min old, new replicaset hash) but the image tag is **still
  `v1.0.1305`** and the running binary is **still commit `6a782274b`** — the same
  binary as this morning;
- **203 commits are in HEAD but not in that binary**, including a lot of `platform/`
  code from several lanes.

This is the documented same-tag trap (`IMAGE_TAG` must be bumped per build, or the
node serves its cached image). **So: any Go change committed today by anyone is
INERT**, and "a fresh build was deployed" is not evidence your code is live. Config
(`agent_definitions`, migrations) IS live immediately — which is why step 6 (449)
works today and TL-043 needed the earlier roll. ⚠ My first probe of this used
**40 zeros as a negative control and it MATCHED** — a degenerate control proves
nothing; use a plausible fake sha.

## 4. What is LIVE and mine, verified 2026-08-17

- **Step 6 / TL-046** — migration `449` applied + verified in-transaction:
  `load_experience_plans` (query_database, no params) between `load_library_tools` and
  `suggest_tools`; `experience_plans` in `input_fields` (the half that makes the
  template variable non-empty); prompt 3,471→4,098 bytes; `experience_plan` citation
  field on each suggestion. **Runtime proof still OWED** — tool-suggester has 0 runs in
  7 days, so nothing has rendered the block. TL-046's verify-later has the check WITH
  a positive control (must contain `site-chat-intake`, must NOT read "None on file.").
- **Step 5 / TL-043** — live on `v1.0.1305`, council APPROVED (`55cda19b`), the
  predicate lockstep test in place. Box: one `sitechat` binary + `sitechat@<domain>`
  template unit; webdesign runs on it, bound `127.0.0.1:8081`.
- **285 acceptance** — all five steps PASS. Step 1's evidence is the strong form: the
  run's own `spec_sections` recorded `locked_merge_count: 1`,
  `locked_sections_merged: ["chat-input-box"]` on the **tier-1** source, and
  `section_facts` came back `[null,null,null]` so the alignment obligation held 3-for-3.
- **Backups retained** (scratchpad, copy them somewhere durable if you need them past
  this session): `285_backup_contact.json` (all three contact components,
  `rendered_html` + `content_data`, pre-rebuild), `285_contact_before.html`.

## 5. Bugs and watch items

- **`bugs_open/299` (NEW, owner-reported)** — the home page CTA names the Website Brief
  Starter and its href **dials the phone**; `tel:` URI malformed; section written AFTER
  the 268 fleet fix, so the PRODUCER question comes first. **File, do not patch.**
  Contains the control that stops a false pass (nav/footer link the tool correctly, so
  a page-wide grep for the URL passes while the button stays broken).
- **`bugs_open/243` (contributed)** — the Anthropic cap stops the BUILD QUEUE by an
  unrecorded path: `claim_work_item` gates every claim fleet-wide on
  `ai_endpoint_health`, re-probed hourly, so one sampled 400 wedged all dispatch
  **11:09:53 → 12:10:18 (60m25s)** with no error row and every liveness check green.
  Falsifier resolved in-file. Not fixed by me: `check_endpoint_health_action.go` is
  dirty under another session. Three fix shapes recorded, ordered by what closes the door.
- Watch: `bugs_open/282` (tool-resolver eats tool-level names — still the co-requisite
  if the chat box is ever DEPLOYED by the deployer rather than placed by hand),
  `bugs_open/275` (LIMIT 30 hides 38/68 tools — unclaimed).

## 6. Falsifiers / re-check before trusting this file

- The owner's new plan may have landed — read it before anything in §1.
- The chassis may have had a REAL rebuild since: re-probe with a positive control,
  do not trust the tag (`v1.0.1305` has now covered two different deploy events).
- `tool-suggester` may have run: if so, TL-046's runtime proof is available — take it.
- The chat-box lock state, and whether `contact`'s copy was restored or kept.
- Whether another lane has taken `bugs_open/299`'s producer question.
