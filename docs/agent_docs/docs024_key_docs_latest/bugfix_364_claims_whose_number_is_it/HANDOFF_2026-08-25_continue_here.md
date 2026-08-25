# HANDOFF 2026-08-25 — `bugs_open/364`: the claims layer cannot tell WHOSE number it is

**Read this first, then `bugs_open/364` §5a–§6g.** Everything below is either measured and dated, or
explicitly marked as not.

**Lane status (updated 2026-08-25, after the owner's rulings): the interim is LIVE and VERIFIED;
Phase 2 is BUILT and awaiting the next roll; the two false queue items are CANCELLED. One thing is
genuinely outstanding — an end-to-end build of an affected page on the new binary — and the
`bugs_open/387` lane is going to hand it to us. See §5.**

---

## 1. What this bug actually is (one paragraph, no jargon)

The platform refuses to publish a page if it contains a number that looks like a boast the business
cannot back up. The check decides by asking "do business-ish words sit near this number" — and on a
site *about* AI agents, words like "agents", "uptime" and "verified" are in every sentence. So it
cannot tell **our** claim ("we run 170 agents") from **somebody else's** statistic ("65% of Fortune
500 use agents"), and it refused real pages over the second kind. The original bug was narrower — a
clock time, "2am", read as the number 2 — and that fix shipped. The residual, which is what this
lane worked, is the general defect underneath it.

## 2. What shipped, and the proof it is live

| commit | what | state |
|---|---|---|
| `ebe8f4323` (pre-existing) | `am\|pm` added to `unitSuffixRe` — the original clock-time fix | LIVE (was already, despite the bug file saying otherwise) |
| `a9002793b` | **interim**: `adoption-tracker`, `protocol-tracker`, `model-directory` added to `editorialPageTypes` | **LIVE v1.0.1337** |
| `0f9f7f3ff` | **fast-follow**: `businessClaimContextRe` gains the plural `orchestrations?` | **LIVE v1.0.1337** |
| `0fe414745`, `afdd163e5`, `69c7b8b1b` | docs, bug files, verdict record, RFC_053 | committed |

**Verified at the artefact 2026-08-25**, chassis rolled 09:27:48Z, build commit `635f2d32`:
binary probe reads 4 / 1 / **0** / 11 — the third is the **negative control** (pre-fix regex form
absent) and it is what makes the probe a real test rather than a promiscuous grep. `merge-base`
puts both commits inside the build. `git diff 635f2d32 HEAD -- claims.go` is empty, so a locally
built `cmd/claimscan` runs the fleet's exact code.

**Result, post-roll, export asserted 118/118:** ai-agent-orchestration.com **36 → 16** findings,
**zero** on any tracker/directory page, and all 16 genuine first-person claims retained.

Council: `b8df25dc-7d19-48b9-9b52-b93b25523d4a` — **APPROVED round 1**, 13 seats. Two objections
changed the code (the plural fix, and RFC_053 being filed for real); one was refuted by enumeration.
Full table with what each cost: `bugs_open/364` §6d.

## 3. ⚠ The three traps in reading this lane's own evidence

Whoever continues will be tempted by all three. They are the reason the numbers above are worded
the way they are.

1. **"0 refusals since the roll" is NOT yet evidence the fix works.** All 6 pages built since the
   roll are `guide`/`tool`/`blog-post` — types already excluded *before* this change. The zero
   measures **absence of demand**, not success. The end-to-end confirmation is genuinely still
   outstanding (§5.1).
2. **The census can only see copy that PASSED the gate.** A refusal writes nothing, so refused copy
   was never stored. Every `claimscan` figure in this lane systematically **under-counts** the
   problem. The damage figure to quote is `agent_error_log`'s: **70 refused build attempts, 23
   pages, 8 sites, 60 days** `[MEASURED 2026-08-24]`.
3. **The psql export truncates silently at exit 0** — observed 101 of 115 rows, findings 36 vs 26,
   no error either time. **Assert `wc -l` against `count(*)` with the same predicate on every run.**
   The short read is monotone in the *safe-looking* direction, which is exactly what a session
   widening an exclusion list hopes to see. `LANDMINES.md` carries it.

## 4. The mechanism, and why the shipped fix is deliberately an interim

`businessClaimContextRe` is a **lexical** gate. It is an allow-list of nouns, with the same
unbounded-miss property as `unitSuffixRe`'s allow-list of units — and a miss there is **silent**,
where a false positive is loud. Proof it bites: the gate carried `orchestration` singular while the
live CTA said "orchestration**s**", so *"We run over 1,600 orchestrations a day across 13 live
production systems"* had **never been scanned at all**. That is now fixed; the class is not.

The interim gates at **page** grain, and the content is mixed at **component** grain: a tracker page
carries a third-party listing *and* a first-person marketing hero and CTA. So the fix knowingly
silences the hero and CTA too — it fails the second half of `editorialPageTypes`' own membership bar
("a body that is never marketing"), which is the exact ground `section-index` was refused on.
`TestTrackerPagesGiveUpTheirFirstPersonClaims` pins that cost and **must be INVERTED, not deleted**,
when Phase 2 lands. `LANDMINES.md` has the blind spot.

**Phase 2 is `RFC_053`** (`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_053_the_claims_number_scan_is_gated_at_page_grain_and_the_content_it_must_judge_is_mixed_at_component_grain.md`).
It carries the consumer enumeration, the two designs already refuted, and the one hard call site.

## 5. What is left — three items, and only one is a task for an engineer

### 5.1 End-to-end confirmation — THE ONE THING STILL OPEN

Nothing of an affected page type has rebuilt on the new binary. The owner cleared a deliberate
rebuild on **any site that is not finance and not being worked on by another thread** — and
measured, **the only site on the fleet carrying those page types (or the word "orchestrations") is
ai-agent-orchestration.com**, which `bugs_open/387` is actively working. So the permission does not
reach a site that could exercise the fix, and I stood down rather than collide.

**`bugs_open/387` has agreed to hand it over**: their plan rebuilds `model-directory` and both
trackers after their `writer_block` fix, all starting after the roll, and they will message the
orchestration id, start time and `unregistered_number` count for each. **Predicted: zero.**

⚠ **Check the clock on any confirmation, including one a peer hands you.** They first offered a
clean `model-directory` build (orchestration `3384ae13`, 0 error rows) — it ran at 06:26Z and the
roll landed at **09:27:24Z**, so it predates the binary by three hours and proves nothing. The same
arithmetic clears the 08-24 tracker *failures* of saying anything against the fix. `WRONG_CALLS.md`
2026-08-25 #6.

### 5.1a FIRST GENUINE POST-ROLL BUILD — and it is a WEAK datum, now measured rather than assumed

`bugs_open/387` delivered it 2026-08-25, dated properly: `model-directory` rebuilt via the scheduled
12:25Z refresh — page-build-handler `f846d061`, writer `9e5f9c48`, rerender `43dbd635`, all
COMPLETED, page re-stamped deployed **12:41:25Z**. Build start **12:38Z**, comfortably after the
**09:27:24Z** roll, so it genuinely ran on `v1.0.1337`. `unregistered_number` rows in the window:
**0**. Rolling counter read in the same breath per §5.1c's protocol: **7,281** (clear).

**They flagged it as weak because `model-directory` was already page-type-gated on the live binary. I
tested WHY rather than accepting the caveat, and the test confirms them:**

- the current copy, judged as the live binary judges it → **0 findings** (expected — page-type gated);
- **the same copy with the page type NOT gated** (i.e. pre-interim behaviour, page_type rewritten to
  `content`) → **also 0 findings**.

So this build would have passed **with or without the fix**. The writer simply produced copy with no
third-party figures this time — precisely the regeneration non-determinism `bugs_open/364` §5 warns
about. **It confirms nothing.**

⚠ **My zero needed its own control and got one.** The sed was verified to have rewritten all three
rows' `page_type`, and the identical trick on `adoption-tracker`'s current copy returns **17
findings** — so the ungated scan certainly fires when there is something to find. Without that
positive control the model-directory zero would have been indistinguishable from a broken probe.

**The strong datum is still the two trackers**, and this is now quantified: adoption-tracker's stored
copy carries **17** ungated findings today. When it rebuilds post-roll, a clean pass IS discriminating
— unlike this one. 387 will send those the same way.

### 5.1b Phase 2 is BUILT but INERT — do not read the commit as the state

Commits `52958897f` (mechanism) + `fa0b513f1` (round-1 rework) + `a3a4597e6` (the record).
**Council APPROVED round 1** — `3ed2b792` — but *approved is not finished*: three of its five
advisory objections were right and changed the code. The worst found a **real silent coverage hole**
(a section whose HTML could not be resolved was dropped from the scan while the page still read as
scanned — on the only gate that refuses a page, and my own `len(out)==0` guard could not see it
because it fired only when *every* section failed). Three seats independently caught me hand-parsing
`sections_metadata` when `extractSectionsFromMetadata` already was the canonical reader. And the
guardian called the block-equivalence sample thin, so it was re-run over **775 pages / 2,042
components** — the whole fleet, export asserted 2042/2042 — with zero differences. Full table:
`bugs_open/364` §6i.

⚠ **One correction from that round a future author needs:** `slot_name` is **not** always
`content_components.function` — 106 of 2,033 live rows differ (`prose-0` vs `ported-prose`,
`call_to_action` vs `call-to-action`, `FAQ Section` vs `faq`). Keying on the slot is right, because
it is what the call sites hold, and a mismatch fails safe — but it is a **silent no-op**, so a new
member of `thirdPartyDataComponents` must be checked against live `slot_name`. The query is beside
the map.

It is Go, so **everything the LANDMINES entry says about the scan being silent on three whole page
types is still true of the running fleet** until the next roll. Prove it at the artefact when you
come to it:
`grep -a -c -F 'adoption-tracker-listing' /proc/1/exe` on a chassis pod (≥1 = Phase 2 live), or
`merge-base --is-ancestor 52958897f <service_binary_capabilities.git_commit>`.

### 5.1c ⚠ NEW, LIVE, and not caused by this lane — a rolling-window fact can refuse a build

`aao-orchestrations` is a **rolling window, not a total**: 35 snapshots ranging **1,494–7,281**,
17 falls, **below 1,600 on 3 of them** — while the site publishes *"We run over 1,600 orchestrations
a day"*. Replaying that low through the scan turns 16 findings into **20**, all four new ones at
`error`, which **refuses the page build**.

**One of the four (`enterprise-reference-deployment`) is exposed on the CURRENT binary**, so this is
live today with nothing from this lane. Phase 2 surfaces the other three by scanning the tracker
pages' own hero/CTA again — the fix working, not breaking: those claims were always unsupported on a
low day and merely un-scanned.

**Before anyone rebuilds those pages, read the counter and record it with the run** — a dip makes the
rebuild fail for a reason unrelated to the fix, and it looks exactly like the fix failing:

```sql
SELECT f->>'value' FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND ss.aspect='evidence_base'
  AND ss.is_current AND f->>'id'='aao-orchestrations';   -- 7,281 on 2026-08-25; clear above ~1,700
```

⚠ **And the fix is NOT what I first wrote — corrected by the `386` lane, twice over.** It is not
their history-retention work: `historySupports` is **exact-match only**, so with the fact at 1,494 a
page saying 1,600 stays convicted, and building it the way I described (support anything ≤ any value
ever held) would pin support at the all-time maximum for ever — worse than today.

**The real remedy needs no code and is already in the register.** `aao-orchestrations`' `writer_line`
says *"over a thousand orchestrations a day"*, and **a floor of 1,000 sits below the historic low of
1,494**, so it is safe on every day on record. The published copy says 1,600 / 1,699 / 1,834 — all
above that low, so it deviates from the register's own instruction. **This is the owner's "state a
floor" ruling applied with the floor set too high; the rule is that the floor must sit below
`lowest`, not below today's value.**

Full account: `bugs_open/364` §6j and an armed `LANDMINES.md` entry (both carry the correction).
**Deliberately not fixed here**: bringing that copy down to its floor, or narrowing the fact's broad
`context_terms`, are changes to a customer site owned by another lane. Flagged to `bugs_open/387`,
`bugs_open/386` and the owner.

### 5.2 ✅ DONE — the two false queue items were ruled false by the owner and cancelled

Owner ruling 2026-08-25. `4405fb38` (adoption-tracker, 8 findings) and `2f8f67dd` (protocol-tracker,
3) are `cancelled`, with the full reasoning on each row's `resolution_path` so it travels with the
row. The four genuine siblings on content pages were verified untouched as the control. The
pre-cancel snapshot — all 11 matched tokens, every one a third party's figure — is preserved in
`bugs_open/364` §6h, because closing archives the row out of `site_work_items`.

**The integrity control was not bypassed**: `armGateClaimsStillPresent` refused to auto-close and
escalated to a human, which is the designed path. The original analysis is kept below because the
mechanism is what matters for the next detector fix.

### 5.2b Why they could not self-close (kept — this is the durable part)

| item id | page | findings | verdict it will now get |
|---|---|---|---|
| `4405fb38-0201-463a-bd2a-40698bed9db7` | adoption-tracker | 8 | `armGateClaimsStillPresent` |
| `2f8f67dd-07b1-4907-a6a1-d7b2bd86fcf4` | protocol-tracker | 3 | `armGateClaimsStillPresent` |

Open since **2026-08-09**. I nearly predicted the revalidator would close them; it will not, and it
is **right not to**. Its claim-granular gate compares the cited token against the slot's current
text: the tokens are still there because the copy never changed — only the standard did. Its own
words: *"the standard moved, not the copy; a claim that stopped being flagged while its words are
untouched has not been addressed."* That is a deliberate integrity control (council round 5,
`compliance` HIGH) stopping a detector change from silently disposing of items on a HITL-terminal
type. **Do not "fix" it.** Someone has to rule these two were always false and cancel them.

Their siblings `ddc90e58` (about, 6) and `962da5c9` (case-studies, 1) are on content pages and are
**genuine** — they stay open on merit. ⚠ Record the ids before closing anything: closing archives
the row out of `site_work_items`, so a later census cannot see what it succeeded at.

### 5.3 RFC_053 — mostly ANSWERED by building it; one question left for the architecture track

The owner ruled "finish the jobs properly" and Phase 2 is built (§5.1b). What remains open is
question 1's second half only: whether the declaration should move out of a Go map and into a
`content_components` column once the list grows, and what the membership bar is when the author is
not also the person measuring.

## 6. Spin-offs filed by this lane — both real, both UNOWNED

- **`bugs_open/386`** — refreshing a **counting fact** turns every page that already rendered the
  previous value into an "unregistered claim" (fundamentallyai renders `11513`; the register held
  `11646` when measured `[MEASURED 2026-08-24]` and it climbs). Framework-wide and periodic.
  **OWNED since 2026-08-25**, lane at `docs/agent_docs/docs024_key_docs_latest/bugfix_386_counting_fact_drift/`.
  ⚠ **Two things I wrote about it are corrected in that bug's §4b, both by the owning lane:** my
  monotonicity premise is FALSE (a reaping counter fell 1,267 → 1,051, and the migration saying so
  predates my claim by a month), and the `4068` I quoted read `7281` a day later — the bug ate my own
  warning.
  ⚠ **And retract one thing I relayed here unverified**: I wrote that `bugs_open/380` "confirmed"
  the rotation amplifies it "and recorded it against CLM-027". The **discriminator that entry
  describes cannot fire on the Go path** — `nearest_fact_id` has **ZERO Go readers repo-wide**
  (verified 2026-08-25); it exists only in the auditor LLM's output JSON
  (`claims_verification/SEED_claims_auditor.sql:70`). At the build gate it is undecidable in
  principle anyway, because the gate runs *now* against the *current* register, so both timestamps
  are today. The 386 lane is correcting CLM-027 and the 380 handoff. **I passed a peer's claim on
  without checking it, in a handoff — exactly what this file warns other people not to do.**
- **`bugs_open/387`** — ⚠ **MY HEADLINE WAS REFUTED 2026-08-25 by the lane that took it.** I filed
  "deployed and 404". The pages **serve fine** at `/adoption-tracker.html` etc.; the extensionless
  form 404s for *every* page on that site, `/about` included, because the worker does not resolve
  slashless paths. **My invented-URL control shared the defect with my claim** — both extensionless
  — so it proved the domain discriminates and nothing about the URL form. The missing control was an
  untouched page at the same form. What survives, and is worse: the `NNN+` placeholder in
  `model-directory`'s hero **is public right now**, and the 387 lane has traced it to literal prompt
  text (migration 557 told the writer to phrase it as "NNN+ AI agents", with no substitution
  machinery behind it — 137 writer calls carried the instruction, 14 copied `NNN` verbatim).

## 7. Where everything is

- Bug: `bugs_open/364_HANDOFF_2026-08-22_a_clock_time_in_page_copy_is_read_as_an_unregistered_business_claim.md` (§5a–§6g)
- RFC: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_053_*.md`
- Register: CLM-016 in `docs/agent_docs/docs026_concept_register/register/claims-verification.md`
- Landmines: two entries in `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` — the interim's
  blind spot, and the silent export truncation
- Wrong calls: `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md`, 2026-08-24, four of mine
- Lane contribution to the site owner: `docs/agent_docs/docs024_key_docs_latest/site_ai_agent_orchestration/CONTRIB_2026-08-24_from_the_364_lane_*.md`
- Code: `platform/orchestration/datahelpers/claims.go` (`editorialPageTypes`, `businessClaimContextRe`), `claims_surface_test.go`

## 8. Coordination already done

`bugs_open/380` (same file) — confirmed no overlap, exchanged findings, it recorded 386 against its
rotation; committed `c9cd817d9` / `ff9c55cb6`. `305 negation gate` — confirmed `claims.go` untouched
by it, since ended. Neither is owed anything further.

## 9. My own missteps, so you do not repeat them

Four in `WRONG_CALLS.md`. The one that matters: **I explained an absence without testing the
explanation** ("a `gte` fact vouches for those figures") and it reached an owner-approved plan before
the test refuted it. An explanation for why something does NOT fire is a hypothesis wearing a
finding's voice. The others: grepping a tool's output for a needle I had never seen it print (false
zero on a 44-finding corpus); calling a stored defect "live public damage" without curling the URL
(it 404s — which produced `bugs_open/387`); and trusting a silently truncated export.
