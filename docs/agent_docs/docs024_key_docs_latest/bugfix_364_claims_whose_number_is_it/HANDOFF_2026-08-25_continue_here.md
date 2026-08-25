# HANDOFF 2026-08-25 — `bugs_open/364`: the claims layer cannot tell WHOSE number it is

**Read this first, then `bugs_open/364` §5a–§6g.** Everything below is either measured and dated, or
explicitly marked as not.

**Lane status: the shipped work is DONE, LIVE and VERIFIED. Three things remain, none of them
blocking, and one of them is a human decision rather than a task.** See §5.

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

### 5.1 End-to-end confirmation (small, waiting on demand rather than on work)

Nothing of an affected page type has rebuilt since the roll. Either wait for one, or create the
demand deliberately. **I did not trigger a rebuild myself**: it regenerates copy on a live customer
site, which is outward-facing and content-destructive, and it was not needed to establish the
logic. If you do it, the recipe and the flag-then-trigger steps are in `bugs_open/364` §6e; check
the **build stamp**, not a later success, because the writer regenerates copy each attempt and a
page can pass for unrelated reasons.

### 5.2 ⚠ Two false items sit in the human queue and will NOT self-close — this is a human ruling

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

### 5.3 Phase 2 / RFC_053 — needs an architecture decision, not code yet

Open question for a human. Not blocking anything; the interim holds and its cost is pinned.

## 6. Spin-offs filed by this lane — both real, both UNOWNED

- **`bugs_open/386`** — refreshing a **counting fact** turns every page that already rendered the
  previous value into an "unregistered claim" (fundamentallyai renders `11513`, register now holds
  `11646`). Framework-wide and periodic. **`bugs_open/380` has since confirmed it interacts with the
  new hourly claims-audit rotation as an AMPLIFIER** and recorded it against CLM-027.
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
