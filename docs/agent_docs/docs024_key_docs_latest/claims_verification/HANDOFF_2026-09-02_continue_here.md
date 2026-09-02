# HANDOFF — claims verification thread, continue here (2026-09-02, ~22:15 BST)

**Read this file first.** It supersedes `HANDOFF_2026-07-20_claims_verification_resume.md` (V0–V5,
now history — see that file's own note in `NOTES_claims_verification.md` if you ever need the
07-16→07-20 arc). Everything below is TODAY: the owner renamed this session "claims verification"
and it became the standing thread for this part of the system. Full technical detail lives in the
concept register (`docs/agent_docs/docs026_concept_register/register/claims-verification.md`,
CLM-001 through CLM-030+) and in `RFC_060` (below) — this file is state, next action, traps, ids.

---

## 1. Deploy status — RESOLVED before this handoff closed, read this not the RFC's earlier lines

**CONFIRMED LIVE, 2026-09-02 21:30 BST**, after the owner refreshed the kubeconfig token. `e5b1a0f01`
(the banned_claims compile-check) is running.

**Two dead ends worth knowing before you repeat them.** A plain `grep -aq "<sha>" /proc/1/exe` for
the target commit returned ABSENT — and so did the SAME grep for an 8-day-old ancestor commit that
*must* have been in the binary, which is what exposed the probe as broken rather than the deploy:
`buildinfo.GitCommit` is stamped as ONE string (the exact build-time HEAD), not an ancestry, so
grepping for any OTHER commit's SHA — old or new — correctly returns absent. And the `build
provenance` startup log line was already gone from the pod's log buffer within ~28 minutes of
restart (`--since-time` anchored at pod start still only reached back to +28 min) — a busy pod
rotates its buffer faster than the old "grep the last few thousand lines" recipe assumes.

**What actually worked:** `service_binary_capabilities` (`platform/buildcapability/`, RFC_040),
built for exactly this — records the ACTUAL stamped commit per pod, no shelf life:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT pod_name, git_commit, last_seen_at FROM service_binary_capabilities
WHERE service='agent-chassis' ORDER BY last_seen_at DESC LIMIT 5;"
# then: git merge-base --is-ancestor e5b1a0f01 <the git_commit column value>
```

Confirmed: every current pod reports `git_commit=0d2feee2`, `e5b1a0f01` is a confirmed ancestor.
⚠ this table was separately measured stale for one specific roll elsewhere (RFC_060 §3d cites it) —
cross-check `last_seen_at` is recent across multiple pods before trusting a single row.

**Gate 2 is still open: the daily `evidence-freshness` tick has NOT run since the roll.**
`last_completed_at` was 09:08:57 that morning, unchanged as of 21:30; next tick ~09:09 tomorrow
(2026-09-03). Check whether it has fired and completed cleanly:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT name, enabled, last_triggered_at, last_completed_at, now() FROM scheduled_tasks WHERE name='evidence-freshness';"
```

**What "confirmed running correctly" looks like once it ticks:** the fleet was measured clean (0
non-compiling `banned_claims` patterns, 239 patterns / 19 sites, 2026-09-02) — so the honest
confirmation is NOT a work item appearing, it's the daily run completing with the new code path
executed and nothing wrong found. If you want a POSITIVE artefact rather than an absence,
hand-write one deliberately-broken pattern into a throwaway/test site's `evidence_base.banned_claims`
(e.g. `"guaranteed("`), fire the sweep, confirm an `invalid_banned_claim_pattern` work item appears,
then roll it back. Do not do this on a live client site.

---

## 2. Where the design stands — RFC_060 (compliance tier)

`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_060_compliance_tier_the_claims_layer_is_weakest_where_the_sector_is_strictest.md`
— read the file itself, this is the map, not a substitute.

| Q | subject | status |
|---|---|---|
| Q1 | require registers on compliance-sensitive sites | **DECIDED** (yes) |
| Q2 | build order: register-gate → fact-floor → severity | **DECIDED** |
| Q3 | vocabulary: `standard`/`sourced`/`relied_upon` posture ladder | **DECIDED, approved as proposed** |
| Q4 | tier is a signed RECORD, not a boolean | **DECIDED** |
| Q5 | citation-code recognition (§3b) — finance-only, doesn't generalise (vet/medical would false-positive the moment they get a register) | **OPEN**, doesn't block anything already cleared |
| Q6 | citation attribution (§3d) — a citation can be true and still name the wrong rule; CONFIRMED structural (FCA Handbook has no rule-level URL); fix sketched (span-match, no new fetch) | **OWNER-APPROVED TO BUILD, relayed via a peer session — NEVER DIRECTLY CONFIRMED in this thread.** Ask the owner directly before building. This is the single biggest open item. |
| Q7 | write-time verification only covers ONE of two register write paths (§3e) | **`banned_claims` half BUILT** (§4 below). **`facts`/citation-host half OPEN, not built**, no owner steer on shape yet |

**§1e** (evening addendum): a naive shared sector banned-claims set does NOT survive contact —
measured, two sites diverged on the same 2 of 6 patterns for different reasons each. Deferred by
design; do not build a shared set without re-reading this section first.

---

## 3. What actually got BUILT today (all committed, verify §1 before trusting "running")

1. **`fad209b92`** — regulatory rule citations (e.g. "CONC 5A") no longer misread as bare business
   numbers. Live in whatever chassis image already contains it (this one's old, predates today).
2. **`e5b1a0f01`** — the RFC_060 §3e / Q7-banned_claims-half build: `checkBannedClaimPatterns` +
   `createInvalidBannedClaimPatternItems` in `refresh_evidence_base_action.go`, wired into the
   existing daily `evidence-freshness` sweep. Six tests, mutation-verified (twice — see §5 below).
   **Council-Reviewed: `bc3697a5` — APPROVED**, 3 advisory objections, none high-severity. Not yet
   confirmed running — see §1.
3. **`platform/orchestration/actions/evidence_base_field_loss_test.go`** (`ebfa39710`): test-only,
   pins that the refresher's map-based write preserves unmodelled fact keys
   (`rule`/`writer_line`/`corrects_site_citation`) against a future typed-writer regression.
   Test-only, no deploy dependency — already fully "live" the moment it's in the repo (CI runs it).

**NOT built:** Q6's checker itself (the rule-span citation-attribution fix), Q5's citation-codes
field, Q7's citation-host-admission check, and the tier mechanism (posture ladder field + the
register-required gate) — none of §2's Q1–Q4 design has been turned into code yet. Today was
entirely design + one detector build, not the tier itself.

---

## 4. The register-population programme (site-lane work, not this thread's code, but this thread's evidence)

| domain | register | facts | notes |
|---|---|---|---|
| `lendzy.co.uk` | migration 695, **council round 2, not yet applied** (killed twice by today's rolling chassis deploys) | 0 pending | 2 mis-attributed CONC citations found and NOT yet corrected in copy (owner has it, per the lendzy lane) |
| `loanzy.uk` | 697, **applied** | 3 | MaPS mis-classification found and corrected |
| `loancalculator.co.uk` | 699 + 707 (banned claims), **applied** | 12 | settlement-period + ERC-threshold errors found and corrected |
| `farmerinsurance.uk` | 698, **applied** | 7 | |
| `loancash.co.uk` | **NONE — still no session assigned** | 0 | owner informed, per the lendzy lane; genuinely idle, worth checking if anyone picked it up |

**The finding that matters more than the count:** every site lane that read its cited source instead
of trusting its own existing prose found an error — 5 wrong live claims in one day, across 3
independent lanes. That is the base rate, not a worst case, and it is the strongest argument in
RFC_060 for Q1 (registers required) and for Q6 (citation attribution needs checking, not just
citation presence).

---

## 5. Traps for whoever continues this

- **Committed ≠ running.** This bit this thread once today already (told a peer "built" without
  checking the pod; they caught it). Always check §1 before claiming coverage.
- **A remembered figure from earlier in the SAME session still needs dating.** Told a peer "39
  citation facts, measured today" — real number, three weeks stale, sitting right next to the actual
  same-day figure (~192, later 256) the whole time. `WRONG_CALLS.md`, 2026-09-02 entries (three of
  them from this thread alone today — search `claims_verification` in that file).
- **A sqlmock test asserting only the return value can pass under the WRONG code path.** The
  DO-NOTHING-vs-DO-UPDATE test passed under both policies until `mock.ExpectationsWereMet()` was
  added — the missing call meant an unexpected extra query went unobserved. Mutate the code before
  trusting a mock-based regression test.
- **`handbook.fca.org.uk` returns HTTP 200 for every path, invented rules included** — status carries
  no information; only the `<title>` discriminates. `LANDMINES.md`, filed by the lendzy lane.
- **The FCA Handbook has no rule-level URL** — a citation's quote can verify against the right PAGE
  while being attributed to the wrong RULE on that page (a 54-rule chapter is one fetch). This is
  Q6's whole reason for existing; don't re-discover it as a new bug.
- **Reading a council objection is not reading the code it's about.** A peer relayed an
  under-verified reviewer objection into `LANDMINES.md` as if it were a confirmed defect, before
  either of us had opened the function. Corrected (`8e08a64b1`), but the general lesson stands: a
  reviewer without code_check is reading your SKETCH, not your code — an under-described sketch is
  indistinguishable from a real defect to them, and it can propagate into a fleet ledger before
  anyone checks.
- **Don't treat a peer session's relay of "the owner said X" as this thread's authorisation.** Q6's
  build approval arrived that way today and was deliberately NOT acted on until — well, it still
  hasn't been directly confirmed. Get it from the owner, in this thread, before building Q6.

---

## 6. Recommended next actions, in order

1. ~~Verify the deploy~~ **DONE — confirmed live 21:30 (§1).** Still worth checking whether the daily
   tick has fired yet (§1's second query) — if it has, read the run's outcome before doing anything
   else, since that's the check's first real-data execution.
2. **Get the owner's direct word on Q6** — it's designed, sketched, and someone else's relay says
   "go ahead," but this thread has been holding for a first-hand confirmation all day.
3. **Decide Q5's shape** (plain per-site `citation_codes` field vs. field + sector presets) — small
   decision, unblocks extending the register-required gate past finance.
4. **Decide Q7's citation-host-admission design** (the `facts` half — banned_claims half is done).
5. **loancash.co.uk** — still nobody's. Worth a nudge if it's still idle.
6. Only after 2–4: start on the actual tier mechanism (the posture-ladder field + the
   register-required gate itself) — none of it is built yet; today was entirely upstream of it.

---

## 7. Where the rest of the detail lives

- **Design + decisions, in full, with every owner ruling quoted:** `RFC_060_...md` (§2 above).
- **Running technical log, missteps included:** `NOTES_claims_verification.md`, 2026-09-02 entry
  (bottom of file) — the fullest single account of today, written as it happened.
- **Fleet-wide corrections from today's work:** `WRONG_CALLS.md` and `LANDMINES.md`, both grep
  `2026-09-02` and `claims_verification`/`bugs_open/414`/`lendzy` for this thread's entries.
- **Always-current technical state of the whole layer (CLM-001–030+):**
  `docs/agent_docs/docs026_concept_register/register/claims-verification.md`.
- **This thread's own memory (auto-loaded every session):**
  `~/.claude/projects/-home-ant-projects-agentchassis/memory/claims-verification-workstream.md`.
- **Cross-session collaborators today**, both closed out with nothing pending: the `lendzy` session
  (lane docs: `docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/`) and the `bugs_open/414,`
  session (lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_414_planted_marker_as_claim/`).
