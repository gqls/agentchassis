# NOTES — bugfix 213, verifier/producer join

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-10 — session 1: selection, the wrong bug, the fix, and two things the tree did to me

### I picked the wrong bug first, and the way I picked it was the mistake

Started from the task "find the next bug in `bugs_open/` that isn't being worked on".
Ran `scripts/who-owns.py` across the high numbers, found `214` (imagery scope_refs
never validated) with one filing commit and no fixing commits, no workstream
directory, and nothing in `git log` touching its four cited code paths since it was
filed. Four checks, all agreeing.

**All four read the same surface: what has been committed.** The owning session's
work was entirely in the working tree. I found out ~1 hour in, by accident — a
comment in `write_site_plan_action.go:192` cited `bugs_open/214` and a date of
"2026-08-10", which is today, on a file with no commit from today. `git status` on
that one path then showed `M` with 116 added lines, plus three untracked files
(`write_site_plan_imagery_scope.go` and two test files). The owner had converged on
materially the fix I was designing, including a lock-transfer transition I had not
thought about.

Full write-up in `WRONG_CALLS.md`. The check that would have caught it in one
command, before any investigation: `git status --short <the bug's cited code paths>`.
My session-start `git status` snapshot was ~1h stale, and CLAUDE.md says in as many
words that it goes stale within minutes — I had read that line and trusted the
snapshot anyway, because `who-owns.py` had already told me what I wanted to hear.

**The hour was not entirely wasted.** My census went wider than 214's and found the
filed defect is substantially larger than recorded: the file measures section scope
only (5 orphans), but page scope has **28 unresolvable refs of 162**, its consumer
join is `scope_ref = $page` *exactly* rather than the section join's tolerant `LIKE`,
and **19 of 22 current-plan orphans have active generated assets**. Proven at the
artefact: gamesdesign.co.uk's about page carries two `<img src="/assets/images/hero.jpg">`
that **404**, while the commissioned `hero-about.jpg` sits deployed and serving at
202,259 B. Contributed to `bugs_open/214` as a clearly-marked section rather than
opening a rival account.

### Re-selection, with the working-tree check FIRST this time

Ruled out in order: `085` (both paths already verified live; only verification owed),
`093` (not a code task — blocked on `083`'s missing scheduler, and `230`'s fix is
already on its way), `211` (owned by the active 122 lane), and everything in 178–240
that `who-owns` showed as fixed/live. Landed on **`213`**, and checked
`git status --short` on its code paths *before* reading a line of it. Clean.

### 213 re-verified on pickup, and it has worsened

The bug file's §4 query, re-run:

```sql
SELECT status,
       count(*) FILTER (WHERE spec->>'audit_source' = 'design-audit') AS producer_b,
       count(*) FILTER (WHERE spec->>'audit_source' IS NULL)          AS producer_a,
       count(*) AS total
FROM site_work_items WHERE handler_agent='color-variable-fixer' GROUP BY 1;
```

`complete`: B=**11** (was 7 on 2026-08-07), A=2. `detected`: B=0, A=3. `unresolved`:
B=0, A=5. **Zero producer-B items have ever failed to close; all 8 that did are
producer A's.** Four more closed clean in the three days the file sat open.

Fleet scope, across all 11 registered verifier item_types:
`hardcoded_section_colors` is the **only** two-producer verified type today. A
disconfirmable measurement — the same query returns 2 where the shape exists and 1
everywhere else.

### The pre-flight that chose the scope predicate, and could have failed

`Grades` keys on `spec.check`. That is only sound if it actually partitions the two
producers, so it was measured *before* being designed in:

```sql
SELECT status, (spec ? 'check') AS has_check_key, spec->>'audit_source', count(*)
FROM site_work_items WHERE item_type='hardcoded_section_colors' GROUP BY 1,2,3;
```

All **10** producer-A rows carry `spec.check`; **0** of 11 producer-B rows do. Clean
partition, no overlap. `git show 62a79c8ac` confirms the key has been written since
the check's first commit, so no historical A row is wrongly disclaimed.

### MISSTEP, caught by the landmine file before it cost anything

While measuring 214 I first resolved page names against `site_plan_sections` and got
27/4 unresolvable. Wrong table: a page with no sections looks like a missing page.
`site_plan_pages` is the one the consumers use, and gives 28/4. I caught this only
because the LANDMINES entry on `DeployedWebPath` had already trained me to ask "which
table does the consumer actually read?" — the same lesson, one door over. Had I not,
I would have reported a number built on a definition no consumer uses.

### MISSTEP, mine, and it inverted a claim I was about to write

Before checking, I assumed the visible symptom of an orphaned hero was "the page
falls back to the site brand hero" — a cosmetic degradation. I nearly wrote that.
The actual served page requests `/assets/images/hero.jpg`, which **404s**. The
symptom is a broken image, not a fallback. I only found out because I checked the
URL instead of reasoning about the fallback chain in `plan_sections_action.go`, which
*does* have a brand-hero fallback and which I had just finished reading. **Reading
the fallback chain told me what the code would do; only the HTTP request told me what
the page does.**

### The test: what it proves, and what it does NOT

I wrote `TestDesignAuditRoutesAreNotGradedBySomeoneElsesVerifier`, ran it, and it
passed. That is worth nothing on its own — this repo's `WRONG_CALLS.md` carries two
separate entries from two lanes about test suites that passed while the fix was
unwired. So I mutated:

| route | scope test | result |
|---|---|---|
| old `hardcoded_section_colors` | none | **RED** — true pre-fix HEAD, names the defect exactly |
| old | present | GREEN |
| new `dark_section_audit` | none | GREEN |
| new | present | GREEN — as shipped |

**The honest reading, which I nearly did not write down:** reverting Half A *alone*
did **not** turn it red, because Half B independently closes this route. So the test
measures "at least one half holds", not "both halves hold", and a future session
could delete Half A without this test noticing. That is defence in depth and it is
also a gap in the guard — recorded here rather than left to look like one test
guarding one line, because the difference only shows up in the mutation matrix and
nobody re-derives a mutation matrix.

Deliberately, the test **never executes a verifier's predicate**. The bug file's §6
warns that re-running the verifier will pass again for the same correct reason; a
test built that way would be green for the wrong reason forever.

### The working tree broke under me, twice, from two other sessions

1. **A non-compiling passenger.** Midway through, `go build ./...` started failing on
   `save_page_sections_action.go` and an untracked `save_sections_decision_gate.go` —
   neither mine, both another session's in-flight work. My own tests had passed
   minutes earlier. Verified my change properly by extracting `git archive HEAD` into
   a scratch tree, copying in **only my seven files**, and building there: clean, and
   the full `./platform/orchestration/actions/...` suite green. A green build in this
   working tree would not have been evidence either way.

2. **A same-file passenger took one of my files.** My commit named seven paths and
   landed **six**. `verifier_coverage_test.go` was absent *and* clean. It had been
   swept into `d644723b8` ("RFC_015 round 1 revisions") by another session — found
   with `git log -S "dark_section_audit"`. This is the documented landmine: a pathspec
   commit protects you from other sessions' files, not from another session taking
   yours out of a file you both touched. **Nothing was lost** — the content at HEAD is
   exactly my two edits, verified by grep (4 occurrences, correct text) — and
   forward-only holds, so it stays as it is. Noted here because the commit message for
   `2d151c41f` describes seven changes and `git show` will only ever show six.

### Council

Submitted before the verdict, correlation **`c9c7c83f-d706-48b0-b433-55de51d88f9f`**,
committed as `2d151c41f` with `Council-Submitted:` rather than `Council-Reviewed:` —
the trailer gate blocked an earlier attempt that used the literal `pending`, correctly,
since a non-UUID resolves to nothing in the 098 join and forward-only forbids an amend.

Two submission-schema gotchas cost a round trip each and are now in the RUNBOOK:
`.plan.summary` is required and is not the same field as `rationale`; and the
operation enum is `modify|add|remove|config_change` — `create` is rejected, a new
file is `add`.

### Still owed

- Read the verdict on `c9c7c83f` and act on it (the code is already on the shared
  branch, so a REVISE is a follow-up commit, not a hold).
- The WII-013 concept-register entry — required by RFC_010 narrowing 1, which permits
  N producers on one item_type *provided* the producer set and key shape are written
  down. Neither was, and 213 §5.1 calls that the cost.
- Migration `374`, **after** the roll: defensively re-type any still-open producer-B
  rows (0 today, but audits run between commit and roll).
- Grade the 11 closed producer-B items against their own `acceptance_test`. Two
  verdicts pre-exist: gamesdesign confirmed FALSE at the served artefact; relojistas
  measured CLEAN in an independent baseline. A `complete` count is not a
  false-complete count.
- Pod-grep after the next roll. Discriminating marker: `verifier_scope_mismatch`
  (expect 0 before, 1 after). Positive control: `verification_unavailable`, live
  since RFC_017, must read 1 in the same exec.

---

## 2026-08-10 (late) — council APPROVED round 1, and the four advisory objections answered by measurement

**Verdict: APPROVED, round 1**, corr `c9c7c83f-d706-48b0-b433-55de51d88f9f`.
14 seats fired, **0 unreadable**, 3 abstained, `gated_by_truncation: false`,
"approved with 4 advisory objection(s) — none high-severity". Committed
`5d482297e` with `Council-Reviewed:`.

> **TRAP, and I nearly recorded another lane's verdict as my own.** CLAUDE.md's
> documented query — `SELECT body FROM doc_notes WHERE categories ? 'council-gate'
> ORDER BY created_at DESC LIMIT 1` — is **correlation-blind**. It returned a REVISE
> for `cb547e0a`, a different session's `save_page_sections` decision-gate round that
> landed between my submission and my read. On a tree this busy, `LIMIT 1` on a
> shared table is a coin toss. Key the read on your own correlation:
> `... FROM doc_notes WHERE body LIKE '%<corr-prefix>%'`, or read
> `diagnosis_artifacts WHERE correlation_id=<yours> AND kind='council_report'`.
> Added to the RUNBOOK.

Also worth knowing: the report `metadata` shape differs between rounds — mine had
`reviewers` (an integer count) and `unreadable`, the other lane's had `reviews` (the
array). The objections live in `diagnosis_artifacts.body`, not in `metadata`, and not
in the `doc_notes` note, which stops after the plan summary.

### The four, and what I did about each

**1. `guardian`, medium — the shared branch was untested. Right, and now closed.**
The scope gate sits in `verifyBeforeComplete`, the single choke point every
completion passes across every pipeline; both of my tests targeted only the fixed
route, so "nil `Grades` leaves the other ten byte-identical" was asserted by reading
the diff. Added `TestOnlyTheOptedInVerifierCarriesAScopeTest`: exactly one item_type
is licensed to carry a scope test. **Mutation-proven** — giving `empty_section` a
`Grades` turns it red naming that type; restored clean afterwards. Stated honestly in
the test's own comment that it proves the branch is *not taken* for the other ten, not
that the nil branch cannot panic (it is a plain nil check on a func field with no
`target.Spec` access before it, so there is no assertion to slip).

**2. `prior_art_librarian`, medium — other consumers keyed on the literal item_type.
This was a check I genuinely had not run, and the seat was right to demand it.**
It came back clean, three ways:
- **Go**: no consumer keys on `"hardcoded_section_colors"` as an *item_type* outside
  the check and its own tests. The three other hits are a CHECK-name list
  (`discovery_checks_registration_test.go:60`) and two doc comments.
- **`reviewRevalidators`** (`revalidate_review_queue_action.go:169`): 6 entries —
  `unresolved_cta`, `required_fields_missing`, `needs_section_data`, `needs_page`,
  `voice_tells`, `claims_unverified`. Neither of my types is among them.
- **Claim-timeout exclusion** (`sql_for_agents/220:170`): holds
  `hardcoded_section_colors` (correct — it has a verifier) and **not**
  `dark_section_audit` (correct — it has none). Unchanged, as planned, and
  `TestRegisteredVerifiersMatchClaimTimeoutExclusion` enforces both directions.
- No `"dark_section"` literal exists outside `write_audit_findings_action.go`.

**3. `guardian`, medium — in-flight rows at deploy time.** Real concern: once Half B
ships, an open producer-B row under the OLD type lacks `spec.check`, so it would be
disclaimed and blocked. **Measured: ZERO producer-B rows are open under the old
type** (`status NOT IN ('complete','failed','rejected','wont_fix')` returns nothing).
So **migration 374 would be an empty migration and must not be shipped as one** —
re-check after the roll and skip it if still 0. The handoff says so.

**4. `architecture`, medium, explicitly non-blocking** — and it is the sharpest thing
in the round, so it is worth quoting rather than paraphrasing: *"`VerifierPolicy.Grades`
is opt-in — the NEXT converging producer on any of the other 10 verified item_types
reproduces this exact bug unless a human remembers to write a Grades function, which
is precisely the discipline that already failed once here."* Its verdict on scope was
`ARCHITECTURE_SIGNAL: insufficient` while agreeing no fresh RFC is needed, because
this is "a correct application of a governance decision already made". The follow-on
it asks for — a periodic check flagging a verified item_type accumulating rows with
more than one spec-shape/`audit_source` and no `Grades` — is recorded in the handoff
as the real closure of the *class*. I have not built it.

Two seats (`editquality`, `prior_art_librarian`) asked whether
`RegisterVerifierWithPolicy` and `VerifierPolicy` pre-exist, since edits 1 and 3
build on them rather than create them. They do (`verifiers.go`, shipped with
WII-011/RFC_017) — which is why the edits connect. A submission-legibility miss on my
part rather than a design fault: I cited the struct's doc comment as precedent but
never stated plainly that the struct already existed.

`guidelines` flagged that WII-013 was asserted in the rationale but not listed as an
edit. Fair — it shipped in `3c72619fc`, a separate commit from the code, which the
submission did not say.

---

## 2026-08-11 — LIVE on v1.0.1284, both replicas. Deployed is not exercised.

Pod-grep, both replicas, one exec each:

| needle | 6j5xn | rvrdg | what it proves |
|---|---|---|---|
| `verifier_scope_mismatch` | 1 | 1 | Half B (the gate) is in the binary |
| `dark_section_audit` | 1 | 1 | Half A (the item_type split) is in the binary |
| `verification_unavailable` *(positive control)* | 1 | 1 | the grep works and this is the binary I think it is |

**The needles discriminate, checked rather than assumed.** `git grep` at `2d151c41f^`:
`verifier_scope_mismatch` = **0 files**; `dark_section_audit` = 1 file, and that one is
`verifier_coverage_test.go` — a `_test.go`, which Go does **not** compile into the
production binary, so its presence in the pod can only come from
`write_audit_findings_action.go`. (That pre-fix occurrence exists only because my
coverage-guard edits were swept into `d644723b8` before I committed.)

**No negative control exists, and I am not claiming one.** The change is purely
additive — it removes no string — so there is nothing that should read 0 after the
roll. `bugs_open/153`'s discipline is satisfied by the positive control instead; the
honest statement is "the binary contains strings that only this change introduces",
not "a removed string is gone".

### DEPLOYED ≠ EXERCISED, and the difference matters here

```sql
SELECT count(*) FROM site_work_items WHERE result->'_verification'->>'status'='out_of_scope';
-- 0
```

Nothing has been disclaimed yet. That is expected rather than worrying — the gate only
fires when a `hardcoded_section_colors` item **without** `spec.check` reaches
completion, and no design-audit item has completed since the roll. **The behavioural
half is unproven and this file stays open until it fires.** WII-011 made exactly this
mistake's mirror image (deployment proven, behaviour asserted) and had to correct
itself the next day; not repeating it.

### Migration 374: measured post-roll, and the answer is DO NOT SHIP IT

In-flight producer-B rows under the old item_type: **0**. The guardian seat raised this
as a state-transition side effect; the population is empty, so the migration would be
an empty migration. Skipped deliberately, not forgotten.

### The 11, enumerated for grading (9 sites)

dartsonline.com ×2 (08-09) · finetuning.uk (08-09) · fundamentallyai.com (08-05) ·
gamesdesign.co.uk (08-03) · gaswholesalers.com (08-04, **no page_id**) ·
leopardessconsulting.co.uk (08-08) · relojistas.com (08-04) · vonc.com (08-03) ·
webdesign.co.uk ×2 (08-04, 08-08).

All 11 carry a real, mechanical `acceptance_test` — computed background-colour,
contrast ratio, absence of an inline `<style>`. **They are not uniform**, and that
matters for the routing question below: gamesdesign's is an already-`var()` fallback
(outside `ReplaceHardcodedColors`' remit, which is *why* it passed), but several others
name inline `style` attributes and `rgba(0,0,0` literals, which may well be **inside**
it. So "the fixer cannot repair these" is true of the worked instance and
**[UNVERIFIED] as a generalisation** — do not assume it of all 11 without checking each.
