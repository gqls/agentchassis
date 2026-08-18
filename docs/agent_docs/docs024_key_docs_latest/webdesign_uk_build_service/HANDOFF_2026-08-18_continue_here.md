# HANDOFF 2026-08-18 — the new commercial terms are LIVE and proven at the bot; the ONE open thread is placing the chat box on the home page — SUPERSEDES HANDOFF_2026-08-17

> **⚠ §2's OPEN THREAD IS CLOSED (2026-08-18 ~10:37Z, recorded by the
> site_delivery_and_editor session at the owner's direction — full record in
> NOTES, two dated 2026-08-18 entries).** The chat box IS on the served index
> and the payment-first/ZIP terms ARE live on preview.webdesign.uk (run
> `ea12d8c9`, verified at the served page, old copy gone). The blocker hunt
> recipe below is STALE: validation issues are now PERSISTED
> (`agent_error_log.error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`, live on
> v1.0.1308) — query the table, don't grep pods. The 09:57Z failure was
> `unregistered_stat` ("1 day" vs a build_duration fact with no numeric
> value); fixed by `SQL_2026-08-18_attest_build_duration_numeric.sql` (this
> dir). Still open for this lane: the two flags in the NOTES 10:37Z entry
> (no-refunds sentence gone from served index; index `rebuild_policy` check),
> plus §3's owner decisions and §4's prompt-maker TODO.

**Start here cold.** Read order: this file → `DECISION_2026-08-17_reasons_for_the_no_refund_position.md`
(the owner's decisions + the lead he still owes) → `TERMS_2026-08-17_new_commercial_position_impact.md`
→ NOTES 2026-08-17 (night) → `bugs_open/299`.

## 0a. UPDATE 2026-08-18 midday — THE CHAT IS LIVE ON THE HOME PAGE; the blocker is identified and fixed

- **LIVE, verified at the artefact** (`preview.webdesign.uk/index.html`, served 10:35Z):
  component order `hero, brief-explanation, chat-input-box, call-to-action` — the owner's
  requested placement. 26 hrefs (none lost). **Zero** occurrences of the retired
  pay-after-approval claim.
- **The blocker was `unregistered_stat`**: the writer turned the HEDGED `build_duration`
  ("usually ready the next day") into a hard stat `"1 day"` at
  `brief-explanation.stat_2_value`. Fixed at the spec
  (`SQL_2026-08-18_voice_assistant_and_stat_guard.sql`) — **not** by attesting 1 as a number,
  which would convert the owner's hedge into a promise he has ruled against.
- **The retrieval recipe (I wasted two rebuilds on this):** the failing step's output is NOT
  persisted on the orchestration, but the action DOES write the issues to `agent_error_log`.
  Query by `context ? 'issues'`, **never** by `domain`:
  `SELECT jsonb_pretty(context) FROM agent_error_log WHERE occurred_at > now() - interval '30 minutes' AND context ? 'issues' ORDER BY occurred_at DESC LIMIT 1;`
- **Owner's VOICE BRIEF is live in writer_block**: helpful assistant, not marketing bot —
  state it, then the next step, the order to do it in, or who can help; the six
  `third_party_options` services are HELP, not just exclusions.
- **Cross-lane:** the other thread is `docs024_key_docs_latest/site_delivery_and_editor/`
  (its SESSION is unreachable from this machine — absent from all 36 peers, `SendMessage`
  refused). Correspondence left as a file **in their directory**:
  `NOTE_2026-08-18_from_webdesign_uk_lane_terms_changed.md`.
- **⚠ TWO ATTESTED CLAIMS ARE AHEAD OF THE MECHANISM** (flagged to that lane, needs a ruling):
  "a ZIP to keep" — the presign defaults to **7 days** (`zip_deliverable_action.go`,
  `expiry_minutes: 10080`); and "a preview link … about a month" — **no month-long preview
  serving found at all**. The live bot makes both promises in pre-sale. Either Phase 4 builds
  them or the owner re-words.
- **Milestone read-out written:** `SUMMARY_2026-08-18_webdesign_uk_build_service.md`.

## 0. State in one paragraph (as at 08-18 morning; §0a supersedes on the chat placement)

PLAN_2026-08-11 is **complete, 6/6**. The owner's new commercial position is **applied and
verified live**: payment before build, no customer preview before paying, ZIP to keep plus a
~1-month preview as an added benefit, no refunds justified against the deal. The chat-box
lock is **OFF** (safely — the plan carries the section now). The chat is **in the home page's
plan and in `pages.sections`**, but the rebuild that would place the component **failed at
`validate_content` ("1 blockers")** and the home page is unchanged. That is the one open
thread. The owner still owes a decision on the page LEAD and the example sites that make it
credible.

## 1. What is LIVE, and how it was proven (not inferred)

| thing | state | proof |
|---|---|---|
| Commercial terms | **LIVE** | live bot: *"No. Payment comes first, then you get the finished site as a ZIP file and a preview link… stays live for about a month."* (it promised the opposite an hour before) |
| `billing_settings.payment_timing` | **`upfront`** | moved in the SAME transaction as the facts — it is read by auth-service (`repository.go:247`); copy and system must not disagree |
| Register facts | 15, three changed | `SQL_2026-08-17b_terms_pay_before_build.sql`; jsonb surgery, other twelve carried through, fact count asserted in a DO/RAISE block |
| Chat-box lock | **OFF** | `SQL_2026-08-17_plan_carries_chat_then_unlock.sql` — plan row FIRST, then unlock (unlocking alone would have deleted it) |
| Bot audience | not businesses-only | fixed in TWO places: `SITE_DESCRIPTION` (box env) AND `promptConduct` (compiled Go). Verified: a running-club enquiry is asked what the club is called |
| PLAN step 5 / TL-043 | live `v1.0.1305`+ | council APPROVED `55cda19b` |
| PLAN step 6 / TL-046 | live (config, mig 449) | **runtime proof still OWED** — tool-suggester has 0 runs; TL-046's verify-later has the check WITH a positive control |
| Chassis | `v1.0.1308` | probe with controls: negative absent, HEAD absent (HEAD has moved on) |

## 2. THE ONE OPEN THREAD — place the chat on the home page

**What works:** `chat-input-box` is in the CURRENT plan for `index` at ordering 2 (after
`brief-explanation`, the price block), and `pages.sections` reads
`["hero","brief-explanation","chat-input-box","call-to-action"]`. `plan_sections` marks all
four **ready**, including the tool.

**What fails:** the rebuild dies at `validate_content` — `__step_error` =
*"content validation failed: **1 blockers, 0 errors**"*. `page_components` for index are
UNCHANGED (2026-08-16), so **the page is intact**; the gate refuses before any write.

**Why the blocker was not identifiable, and the fix for that:** the failing step's output is
NOT persisted (`valid`/`issues`/`blockers` all null on the run — the error path runs instead
of the output field being written), and `agent_error_log` has no row. **But the action LOGS
each issue**: `validate_page_content.go:412`, `logger.Warn("ValidatePageContentAction: issue", …)`.
So the recipe is: re-dispatch, then
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=20000 --since=25m \
  | grep -a "ValidatePageContentAction: issue"
```
**A watcher doing exactly this was left running on 2026-08-18** against item `762f09c0`
(reset to `triaged`); its output is
`<scratchpad>/blocker_hunt.log`. If that scratchpad is gone, re-run the two steps above.

**The blocker categories that can stop it** (`validate_page_content.go`): `placeholder_text`,
`unrendered_template`, `unrendered_template_block`, `cross_site_domain`, `cross_site_company`,
`placeholder_email`, `banned_claim`, `meta_commentary`. **Leading hypothesis, UNCONFIRMED:**
the tool section is in the list but has no `page_components` row on index, so the assembled
page carries an empty/placeholder section → a `placeholder`/`template` blocker rather than a
claims one. **Do not act on that until the log confirms it** — the copy that failed was never
saved, so guessing would change wording to satisfy a rule that may not be the one complaining.

**If it IS the missing component:** the framework route to attach an existing tool to an
existing page is `create_tool_component`'s **`adopt_existing_page`** arm (register **TL-044**,
another lane, `bugs_open/286`) — NOT `deploy_tool_to_site`, which creates a tool PAGE at
`/tools/…`. Read TL-044 before using it.

**⚠ Known interaction, found in another site's rows:** `save_page_sections` refuses a page
whose `rebuild_policy='owned'` (*"tool/widget-owned: a generic section save would clobber
it"*). Once the chat IS on index, index may become owned and generic rebuilds of it refused.
Design the placement with that in mind.

## 3. Owner decisions outstanding

1. **The page LEAD.** Proposal **F — "show the work, promise nothing"**: *you can see exactly
   what you get: real sites built with this system, and the exact prompt that produced each
   one.* Rationale, and why it survives the cheaper sibling brand, in `DECISION_…`. He also
   owes the **example sites on his own domains** — under pay-first these are the trust
   mechanism that replaces the preview, and `dda32da9` ("no portfolio or case studies",
   already FAILED) is now load-bearing rather than cosmetic.
2. **`bugs_open/299`** — the home-page CTA names the Website Brief Starter and its href DIALS
   THE PHONE. Filed, deliberately NOT patched (the rewrite regenerates that section). The
   producer question survives the rewrite: the section was written 2026-08-16, AFTER the 268
   fleet fix, so something still generates it.
3. **Contact email** `webdesign@contactforsales.com` (domain mismatch, open item `a8d6f440`);
   Stripe webhook hostname; Stripe keys via terraform.

## 4. TODO recorded at the owner's request

**A prompt maker** — help a visitor write a really good prompt for the system. His own
preference, and the better route: **make the EXISTING chat box do it** rather than build a
second widget (it already has the site's facts, the four abuse controls and a live
deployment). The change is to `promptConduct` (compiled behaviour, not facts). Not started;
interacts with the `website-brief-starter` tool, which already covers adjacent ground.

## 5. Traps this lane has now paid for (all filed, all cheap to re-hit)

- **A "fresh build" can ship NO new code** — a same-tag rebuild serves the cached image (203
  commits inert on 2026-08-17). Probe the binary with a **positive control and a negative
  that can actually be absent** — 40 zeros MATCHES every binary and proves nothing (LANDMINES).
- **The claims gate reads a bare "no" as an intensifier** — *"there is no refund"* scans as a
  refund PROMISE and blocks the page. Always *"we do not offer refunds"* (writer_block says so).
- **The register is the authority, not the page.** Correcting page copy alone is undone by the
  next rebuild; and the bot renders the register, so a register fix reaches it in ~5 minutes —
  **restart the service to beat the cache when verifying**.
- **Unlocking a section can delete it** if the plan does not carry it (the lock was the only
  thing merging it into the list). Plan first, then unlock.
- **A backup file can contain an ERROR instead of a backup** — check its CONTENT, not its
  existence (a `psql` error wrote itself into one here; a row-count check caught it).
- **A watcher can fire on the wrong signal** — mine triggered on the section-list change,
  which happens early, and reported a mid-run snapshot as if it were the result.

## 6. Falsifiers

- The blocker may already be identified in `blocker_hunt.log` — read it before re-running.
- The owner's plan session may have landed a fuller brief; check before rewriting anything.
- `pages.sections` for index may have been reset by another rebuild — re-read it.
- The chassis rolls often: re-probe with controls, never trust the tag.
