# HANDOFF 2026-08-25b — `364` is CLOSED. What shipped, what is still true, and what is NOT fixed

**The bug file moved to `bugs_closed/364_HANDOFF_2026-08-22_a_clock_time_in_page_copy_is_read_as_an_unregistered_business_claim.md`.**
Its §5a–§6n hold the full account; this file is the read-out.

**Supersedes `HANDOFF_2026-08-25_continue_here.md`**, which was written while the work was open and
whose §5 "what is left" list is now spent. Read this one.

---

## 1. What the bug was, in plain terms

The platform refuses to publish a page carrying a number that looks like a boast it cannot back up.
The check decided by asking *"do business-ish words sit near this number"* — and on a site **about**
AI agents, words like "agents" and "uptime" are in every sentence. So it could not tell **our** claim
("we run 170 agents") from **somebody else's** statistic ("65% of Fortune 500 use agents"), and it
refused real pages over the second kind. The filed symptom was narrower — a clock time, "2am", read
as the number 2.

## 2. What shipped, all of it live and artefact-proven

| commit | what | council |
|---|---|---|
| `ebe8f4323` | the original clock-time fix (`am`/`pm`) — was already live, the bug file said otherwise | — |
| `a9002793b` | **interim**: three tracker page types added to `editorialPageTypes` | `b8df25dc` APPROVED r1 |
| `0f9f7f3ff` | `businessClaimContextRe` gains the plural `orchestrations?` | same round |
| `52958897f` + `fa0b513f1` | **Phase 2**: the scan decides at COMPONENT grain; the three page types come back OUT | `3ed2b792` APPROVED r1 |
| `35f452a0e` | **pre-roll fix**: a digit inside an acronym (`A2A`, `W3C`) is not a claim | `64d852b4` submitted |

**Live on chassis `v1.0.1339`** (rolled 2026-08-25 19:07Z). Proven two ways, and the second is the
one that matters:

- binary probe on **both** replicas — `thirdPartyDataComponents` 1, `normaliseSurfaceKey` 2,
  `isASCIILetter` 1, **negative control `isASCIILetterZZZ` 0**, positive control `news-index` 11;
- **behavioural demand control** — a synthetic unsupported claim injected into three slots of the
  live `protocol-tracker` copy is **FLAGGED in `hero`**, **FLAGGED in `call-to-action`**, and
  **SILENT in `protocol-tracker-listing`**. That is the design, observed rather than inferred.

**Effect measured on live copy:** ai-agent-orchestration.com went **36 → 16** findings, zero on any
tracker page, with all 16 genuine first-person claims retained. Corpus-level mutation proof: empty
the component map and exactly the original 36/20 return.

## 3. ⚠ What is NOT fixed — read this before assuming the class is closed

1. **The lexical gate is still an allow-list OF NOUNS**, with the same unbounded-miss property as the
   unit allow-list it sits beside — and a miss there is **silent**, where a false positive is loud.
   Proof it bites: the gate carried `orchestration` singular while live copy said "orchestration**s**",
   so *"We run over 1,600 orchestrations a day"* had **never been scanned at all**. That instance is
   fixed; the class is not. A plural, synonym or hyphenation it has not met is a silent miss.
2. **A new tracker/directory component must be DECLARED** in `thirdPartyDataComponents`, verified
   against live `slot_name` (not `content_components.function` — 106 of 2,033 rows differ). An
   undeclared one is scanned, which is the safe direction but means the false positives return for
   that component until someone measures and adds it.
3. **`RFC_053` stays open on one forward question**: whether that declaration should move from a Go
   map into a `content_components` column as the list grows. Not a live defect.
4. **Rolling-window `gte` facts can refuse an honest page** — `bugs_open/386`'s class, and the reason
   a page can fail for a reason that has nothing to do with this bug. Check the counter with any
   rebuild of these pages.

## 4. The two spin-offs this lane filed — both OWNED, both live work

- **`bugs_open/386`** — refreshing a counting fact convicts every page that rendered the previous
  value. Owned, council-approved slice built. ⚠ Its §4b carries **two corrections to my own recorded
  ruling**: my monotonicity premise was false (a reaping counter fell 1,267 → 1,051, and the
  migration saying so predated my claim by a month), and the example figure I quoted went stale in a
  day.
- **`bugs_open/387`** — owned. ⚠ **My filing's headline was REFUTED**: I reported three pages
  "deployed and 404"; they serve fine at `.html`, and the extensionless form 404s for every page on
  that site. My invented-URL control shared the defect with my claim. What survived is the `NNN+`
  placeholder, which that lane traced to literal prompt text and has fixed.

## 5. Instrument findings worth more than this bug

- **`service_binary_capabilities` was STALE for the final roll.** Its newest `agent-chassis` row
  named a commit containing no Phase 2 while the binary plainly carried it; a second lane dated the
  same build from that table and got a commit predating the acronym fix. **Probe the artefact; treat
  that table as a hint.**
- **A council APPROVAL never validates a measured number.** A review seat **executes nothing** — not
  even read-only SQL (seven config keys, no query key; the schema is TEXT in its prompt). *A
  submission can make its reasoning checkable; nothing it can do makes its evidence checked.*
- **The inert window between commit and roll is a free check, and it caught a real regression here** —
  Phase 2 was approved and correct and would still have refused `protocol-tracker`'s build over the
  digit in `A2A`.
- **`git diff --numstat <file>` before committing.** I swept a peer's whole in-flight mechanism into
  a commit about a landmine — 98 lines where I expected 20 — after being warned in writing that we
  were both in that file. The commit-scope block cannot catch it: it lists files, so it finds a path
  that went *missing*, never one that arrived *fatter*.

## 6. Where everything is

- Bug: `bugs_closed/364_HANDOFF_2026-08-22_a_clock_time_in_page_copy_is_read_as_an_unregistered_business_claim.md`
- RFC: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_053_*.md`
- Register: **CLM-016** in `docs/agent_docs/docs026_concept_register/register/claims-verification.md`
- Landmines: the tracker blind-spot entry is **RETIRED** (kept as a record); live entries remain for
  the acronym class, the silent psql truncation, the `git log` attribution trap, the council-approval
  limit, and rolling-window `gte` facts
- Wrong calls: `WRONG_CALLS.md`, 2026-08-24 and 2026-08-25 — **sixteen entries**, most of them mine
- Code: `platform/orchestration/datahelpers/claims.go` (`thirdPartyDataComponents`, `ClaimSurface`,
  `isExcludedNumber`, `businessClaimContextRe`), `claims_surface_test.go`,
  `platform/orchestration/actions/validate_page_content_surface_sections.go`

## 6b. ⚠ IF A TRACKER PAGE REFUSES AFTER THIS ROLL — where it goes, now that BOTH lanes are closed

`bugs_open/387` and this lane agreed a disposition rule while both were open, and it routes by the
`CONTENT_VALIDATION_BLOCKER_DETAIL` type. **Both lanes are now closed, so "goes to X first" points at
nobody** — which is exactly the dangling-pointer failure this estate keeps paying for. Written here
so the rule survives its authors:

| finding type on a tracker page | what it means | where to start |
|---|---|---|
| `unregistered_number` on **`hero`** or **`call-to-action`** | **NEW since `v1.0.1339`.** Those slots were silenced wholesale before this roll; Phase 2 restored the scan to them. Either a genuine unsupported first-person claim (the layer working), or a shape component grain has newly exposed. | this file §3, then `bugs_closed/364` §5b and §6m. **Check the register's rolling counter first** — `bugs_open/386`'s class can refuse an honest page for an unrelated reason |
| `unregistered_number` on a **`*-listing`** slot | a declared third-party component should be exempt — so either the component is NOT in `thirdPartyDataComponents`, or its `slot_name` differs from the declared key | `claims.go`, the membership query beside the map. **Verify against live `slot_name`, not `content_components.function`** — 106 of 2,033 rows differ |
| `placeholder_text` with a numeric stand-in `Value` | `bugs_closed/387`'s detector working as designed — a stand-in reached the copy and was refused rather than published | `bugs_closed/387` |

**The general rule, which is the durable half:** a refusal on those pages is no longer evidence that
either bug regressed. Both mechanisms are live and both are *supposed* to fire. Read the
`BLOCKER_DETAIL` type before concluding anything, and do not reopen either bug on the strength of a
refusal alone.

## 7. The one thing still owed

**Read the acronym fix's council verdict** — `64d852b4-bd70-43a6-9c7a-82202ea5688f`. The code is
already live, so a REVISE would mean acting on a shipped change, not holding one back:

```sql
SELECT metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='64d852b4-bd70-43a6-9c7a-82202ea5688f' AND kind='council_report';
```

Nothing else. The lane is closed.

**`bugs_open/387` closed the same evening** — same roll, source fix proven, detector and CLM-029 carry
live, probed at the binary on both replicas (their pre-roll probe read 0 and post-roll reads 1, so the
discriminator flipped rather than merely agreeing). Two residuals survive that close and are written
in its file: a refresher-survival check on the 611 block, and the disposition rule now recorded in §6b
above.
