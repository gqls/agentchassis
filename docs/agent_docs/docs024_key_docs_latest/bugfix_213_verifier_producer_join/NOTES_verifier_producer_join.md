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

---

## 2026-08-11 (later) — CLAUDE.md banned the method I verified with, so I re-verified. The reading held, and got stronger.

CLAUDE.md was rewritten today (§"Building & deploying images"). Two changes hit this
lane's verification directly, and I only saw them because the file was edited under me:

1. **`strings … | grep -c` is now forbidden** — "the old recipe that stood here
   produced three confidently wrong readings in one day". That is exactly what I used.
2. **`v1.0.1284` shipped THREE revisions under one tag** (`bugs_open/249`). The release
   pinning fix has not rolled, so a tag can straddle other sessions' commits. "Read the
   stamp of the service you actually mean" — a tag is not a commit.

So my "proven live" claim rested on a discredited method and a tag that does not
identify a revision. I re-ran it properly rather than defending it.

**The sanctioned log route returned nothing** — `kubectl logs -l app=agent-chassis`
had **13 lines** and no `build provenance` line. Not a missing stamp: chassis log
retention is seconds (the standing log-measurement landmine), so a startup line is long
gone. Worth knowing, because the new CLAUDE.md recipe is the *first* thing it tells you
to try and it will silently return nothing on this service.

**The binary probe, both replicas, with BOTH controls** (`grep -aq … /proc/1/exe`, no
`2>/dev/null`, no discovery grep for "some 40-hex string" — both named traps avoided):

| needle | 6j5xn | rvrdg | role |
|---|---|---|---|
| `verifier_scope_mismatch` | PRESENT | PRESENT | Half B |
| `does not grade this item` | PRESENT | PRESENT | Half B's operator message |
| `verification_unavailable` | PRESENT | PRESENT | **positive** control (live since RFC_017) |
| `zzz_this_string_must_never_exist_213` | ABSENT | ABSENT | **negative** control |

> **CORRECTION to my own entry above, and it corrects in the STRONGER direction.**
> The 2026-08-11 entry says *"No negative control exists, and I am not claiming one …
> the change is purely additive — it removes no string."* That reasoning was sound for
> a *removed-string* control and wrong as a general claim: a control only has to be a
> string that **must not be there**, and an invented one does that job perfectly. I had
> talked myself out of a control that cost one line. The probe now demonstrably returns
> ABSENT when it should, so PRESENT means something.

**Net: the conclusion is unchanged and now rests on the sanctioned method with a
two-sided control.** Both halves are in the running binary on both replicas. What I got
wrong was the method and the confidence, not the answer — and I would not have caught
either if the guidance file had not changed under me, which is its own argument for
re-reading it rather than trusting a session-start snapshot.

---

## 2026-08-11 (evening) — D3 BUILT: `verifier-remit-check`, the class detector

Picked up at D3 (the handoff's "start here — self-contained"). Built as a daily
CronJob Go binary rather than a query somebody has to remember to run.

### The design, and the three axes that look right until you measure them

The question: *does any item_type with a REGISTERED VERIFIER carry rows from more
than one producer shape, while its verifier declares no remit?*

The owner's ruling says key on the **spec shape**. The obvious implementation —
count distinct top-level key-sets per item_type — is **wrong, and the live data
says so** [MEASURED 2026-08-11]:

| item_type | distinct key-sets | true producers |
|---|---|---|
| `hardcoded_section_colors` | 5 | 2 |
| `empty_section` | 2 | 1 |
| `literal_markdown` | 2 | 1 |
| `page_canonical_collision` | 2 | 1 |
| `truncated_component` | 2 | 1 |

Four false positives out of nine. Producers add and drop optional keys over their
life (`original_pipeline`, `out_of_remit`, `intact_version_number`), so the count
measures *spec revisions*, not producers. **Clustering is the load-bearing half of
the design, and nothing in the ruling said so — the data did.**

The clustering rule is the **overlap coefficient**, `|a∩b| / min(|a|,|b|)`, not
Jaccard. Also decided by data: `page_canonical_collision`'s two real shapes are 11
keys and 3 sharing 2 — J = **0.167**, which invents a second producer, against an
overlap of **0.667**. A small shape almost contained in a large one is a variant
of it, not a rival to it. The threshold (0.5) is **not tuned**: every same-producer
pair in the fleet overlaps ≥0.667 and the one genuine cross-producer pair overlaps
**0.000**, so 0.5 sits in the middle of an empty band. If that band ever closes,
the threshold stops being defensible — which is a thing to check, not to assume.

Three axes measured and rejected, each of which would have looked authoritative:

- **`created_by`** — 2–3 distinct values on `empty_section`, `literal_markdown`,
  `hardcoded_section_colors`. Fires on single-producer types. (The owner's ruling
  said this; this is the independent re-measurement, not a restatement.)
- **the `source` COLUMN** — reads 2 on `page_canonical_collision`, one producer.
- **the VALUE of `spec.check`** (a council suggestion in spirit) — reads 2 on the
  same type, whose probe rows omit the key. Its *presence* already participates
  through the key-set, which is where it legitimately discriminates.

### The zero had to be able to show its working

Today's finding set is **empty**, and that is the mechanism working:
`hardcoded_section_colors` still has 2 producer families and is suppressed only
because WII-013 registered its `Grades`. A bare "0 findings" is exactly the shape
016b §9 warns about, so the report **always names the suppressed types**, and
`--ignore-remit` (a diagnostic mode that refuses to write) re-runs the census with
the suppression off. Live, that produces the real finding — 9 rows with no
`audit_source` since 2026-04-08 against 8 `design-audit` rows since 2026-08-04,
exit 1. **The detector can fire, demonstrated against production data, without
mutating source or writing a row.**

### The mutation matrix (six, recorded rather than summarised)

| mutation | result |
|---|---|
| threshold `0.5 → 0.0` | **RED** — the same-label test |
| threshold `0.5 → 0.7` | **RED** — `page_canonical_collision` splits at 0.667 |
| `audit_source` axis removed | **RED** — *but only the constructed test; the live-census test stays GREEN* |
| `Finding()` ignores the declared remit | **RED** |
| `Finding()` fires on one family | **RED** |
| coverage-guard entry removed | **RED** — "has NO verifier and is NOT an acknowledged gap" |

The third row is the honest one and worth reading twice: **the live data cannot
pin the `audit_source` axis**, because today's only two-producer type also happens
to be shape-disjoint. So the key-set clustering could have been deleted entirely
with every real-data test still green. That is what
`TestDisjointShapesUnderOneLabelAreTwoProducers` is for, and its comment says in
as many words that it is CONSTRUCTED rather than observed.

A seventh, silent one worth recording: removing `verifier_remit_gap` from
`liveItemTypes` while keeping the gap entry is **GREEN**. The two entries are only
jointly meaningful — the ratchet is what makes the map entry load-bearing — so
"tidying" one of them later would be silent.

### The write path, proven against the live schema without writing

```
BEGIN; <the exact INSERT>; <the exact close-out UPDATE>; ROLLBACK;
```
Insert accepted (no 42P10 against the partial dedup index — the standing lockstep
trap), a second identical insert swallowed by `idx_swi_dedup` (`INSERT 0 0`), the
UPDATE closing exactly one row with `result.closed_by` set, and the rollback
leaving zero rows.

### MISSTEP — an inverted grep that nearly became a council argument

To answer "is any discovery check fleet-scoped?", I ran
`grep -Lq "SiteID" check_*.go && echo` and read the 20 files it printed as *checks
with no site scope*. **`-L` and `-q` together is just `-q`**: it printed the files
that DO reference SiteID, i.e. the exact opposite set. The corrected run
(`grep -rLE`, no `-q`) returns **one** file, and that one is a helpers file, not a
check. I caught it because "20 of ~50 checks ignore the site" contradicted
everything else I had read about the framework — a plausibility check, not a
process. Had it survived, I would have told the council the framework was
routinely fleet-scoped and used it to justify the opposite conclusion to the one
the evidence supports. **Shell flags that combine into a different question fail
silently and produce a confident answer** (the sibling of the estate's other
silent-shell traps).

### Council round 1: REVISE, and it was worth the round

Gated by `debug_historian [high]`: *"the plan adds only base/cronjob.yaml with no
overlay pinning newTag … for a brand-new service this risk is worse, not better"*.
The overlay **did** exist — written before the verdict — but the submission had not
listed it as an edit, so the seat could only see what I showed it. Same class for
`prior_art_librarian [high]`, which flagged `RegisteredVerifierItemTypes` as a
symbol the plan invoked but never added: it pre-exists at `verifiers.go:177`, and
the binary had been built and run against live data before the objection arrived.
**Both were submission-legibility failures, not design failures — and the fix is to
list what you built, not to argue.**

Two objections found real things:

- `guardian [medium]` — *"no existing 'deferred' status is confirmed live on this
  table … a new status value can leak into a queue it was never meant to enter"*.
  I had asserted safety by analogy. Measured instead: **`deferred` carries 316 rows
  across 14 item_types**, of which **only 15** also carry an empty `handler_agent`.
  So the status is old news, and — the part I did not know — **`deferred` alone is
  not undispatchable; only the PAIR is.** That is now a LANDMINE entry, because
  reading `status='deferred'` as "nothing will touch this" is wrong for 95% of the
  rows carrying it.
- `reuse_agent [medium]` — *"why not a `discovery_checks/check_*.go`"*. Fair, and it
  wanted evidence rather than precedent. Three checkable reasons, all in the code
  now: invocation is gated on live agent config (`enabledChecks` — a check file is
  inert until `agent_definitions` names it, the exact inert-by-omission failure this
  class of detector exists to avoid); retraction is site-scoped
  (`resolveWorkItems(…, dctx.SiteID, …)`); dedup is `(site_id, item_key)`. A
  fleet-level answer has no site to be scoped to.

`editquality [medium]` named the residual precisely — two producers sharing a
label with ≥50% shape overlap merge silently — and it cannot be closed by any
row-shaped test. It is now written at the clustering function rather than only in a
risks block.

### A live finding for D1/D2, noticed while measuring something else

`dark_section_audit` already has **14 rows, all created 2026-08-11, 13 of them
already `complete`** — the rotation re-detected within a day of the roll, exactly
as D2 predicted. But the type still has **no verifier**, so those 13 closed
**ungraded**. D2's stated dependency on D1 is no longer hypothetical: the machine
is now re-finding these defects and losing them again on a ≤7-day cycle, 13 items
per cycle at the current rate. That is the strongest argument yet for D1, and it
was not visible when the ruling was taken.

### Council round 2: APPROVED, 8 advisory objections — checked, not waved through

`decided_by: approved with 8 advisory objection(s) — none high-severity`,
14 seats, 3 abstained, `gated_by_truncation: false`. What each turned out to be:

**One changed the shipped code.** `constitution [medium]`: the census filtered
`item_type IN (…)` by interpolating a regex-validated list, "the classic workaround
for parameterization, not parameterization itself". It was right, and the fix is
better than parameterising: the filter was **redundant** — `assess()` already
iterates the registry — so the census now covers every item_type and the SQL is a
**constant with no parameters and no concatenation at all**. Measured cost: ~150
census rows instead of 17, over ~6.5k work items, no perceptible change in runtime.
A test now pins the constant (re-adding "just one" interpolated filter is how it
would come back), and a new refusal was added with it: an EMPTY census exits 2,
because `site_work_items` is never empty and a broken read must not read as a clean
fleet.

**Four were already right in the code, and are recorded so nobody re-derives them:**

- `editquality [medium]` — "the plan never states the item_key's granularity; a
  coarse key would collapse distinct per-type findings into one row". The key is
  `verifier-remit:<item_type>`, one per subject type, and `fileFinding` is called per
  assessment. Plan-legibility gap, not a defect.
- `guardian [medium]` — "is `system.internal` a real row in `sites`, or a fabricated
  UUID? a dangling FK on insert". **CHECKED: it is a real row** (`domain
  system.internal`, `status system`) and `site_work_items_site_id_fkey` really is a
  FK to `sites(id)` — which the ROLLBACK probe had already proven by accepting the
  insert.
- `guardian [low]` — "does a `scheduled_tasks` pre_query treat deferred+empty
  handler as anomalous?". **CHECKED: no.** Only two of the 30 pre_queries mention
  either column: `feasibility-recheck` selects `status='blocked'` and requires an
  `agent_definitions` row matching `handler_agent` (two reasons it cannot see this
  row), and `claimed-item-timeout` works on claimed rows.
- `bug_historian [medium]` — `bugs_closed/078`, a NULL `handler_agent` silently
  livelocking the build dispatcher. **Two independent reasons it cannot recur here,
  and the first is structural:** `handler_agent` is `NOT NULL DEFAULT ''`, so 078's
  exact shape (a *Scan into string* failure on SQL NULL) is unrepresentable; and 078
  bit rows at `status IN ('triaged','approved')`, which `deferred` is not.

**One had exact prior art, and the prior art is the answer.** `guidelines [medium]`
claimed a WORK-ITEM DEDUP rule of "DELETE+INSERT, not ON CONFLICT". The same seat
raised the same objection against the `bugs_open/208` lane, which answered it by
induction: a **bare** `ON CONFLICT DO NOTHING` names no conflict target, so there is
no partial-index inference and no 42P10 — that hazard belongs to the *targeted*
`ON CONFLICT … WHERE` form `insertWorkItem` uses. This check uses the bare form, and
`triageInsertNeedsDiagnosis` — the existing producer of the sibling fleet-level type
— uses it too. Ironically the objection fired *because* the submission mentioned
proving "no 42P10": naming a hazard you avoided reads as using the risky shape.

**Two are a standing signal rather than an objection to this change.**
`reuse_agent`, `architecture` and `tooling_provenance` all say the same thing: this
is the **fifth** standalone Go CronJob meta-check (component-render-check,
single-owner-carriers-check, component-fallback-check, shared-output-fields-check),
each re-implementing doc_notes-on-every-run, dedup and retraction. Three seats, two
rounds, and my own risks block called it "recorded not actioned". Recording it a
third time would be the same non-answer, so it is now filed as a proposal in
`architecture_review/` naming the five and what they duplicate.

**One is a limit of the reviewer, not of the change.** `improvement_guardian
[medium]`: RUNNER OWNS INSERTION is bypassed, so `insertWorkItem`'s Go-level
two-strike anti-churn does not apply. True. Accepted deliberately: two-strike
suppresses an item that keeps being re-filed after terminal closes, and this check
files at most one row per verified item_type per day, deduped, retracted only on a
positive observation. The churn ceiling is one row per type — but it IS a real
difference from the framework's contract and belongs in the register entry, not in
a shrug.

### Deployed, and proven at the artefact

`make build-… IMAGE_TAG=v1.0.1288` (committed HEAD) → push → `deploy-…` (CronJob
created, `25 7 * * *` UTC) → `verifier-remit-check-now`. The Job completed and its
pod printed the same census my terminal run produced — **and with no "READ-ONLY RUN"
banner, which is how the report says `PG_CLIENTS_HOST` was wired and the direct
route was taken.** The artefact:

```
SELECT created_at, source, left(body,160) FROM doc_notes WHERE source='verifier-remit-check';
-- 2026-08-11 18:45:16+00 | verifier-remit-check | "Verified item_types evaluated: **12**. Findings: **0**.
--                                                  Multi-producer types answered by a declared remit: **1**."
SELECT count(*) FROM site_work_items WHERE item_type='verifier_remit_gap';  -- 0, correctly
```

Re-rolled at **v1.0.1289** after the constitution-seat change, because a same-tag
rebuild ships the node's cached binary.

---

## 2026-08-12 (afternoon) — D1 task 1: the remit measurement, and what it turned up

Full write-up with the evidence is the `CONTRIBUTION 2026-08-12 (afternoon)` section of
`bugs_open/213`. This is the working log: what I ran, what went wrong on the way, and the
three things I nearly got wrong.

### The probe

`git archive HEAD` into the scratchpad (RUNBOOK recipe — the live tree has four other
sessions' `.go` edits in it), then a throwaway `cmd/remitprobe/main.go` inside **that
extracted copy** importing `checks.ReplaceHardcodedColors` directly. **It is not in the
repo and must not be** — it is a measurement, run once, with no reason to become a
sixth standalone Go checker (`RFC_024` exists because we already have five). If the
measurement needs repeating, re-extract and re-write it; the durable artefact is the
number and the method, both here. Bodies exported from the live DB as
`replace(encode(convert_to(col,'UTF8'),'base64'), E'\n','')` so nothing is mangled by psql
or by the shell; one `id|domain|layer|key|b64` line per body. The binary decodes, applies
the transform, prints CHANGED/UNCHANGED, and for UNCHANGED prints *why* — style-block
count, inline-`style` count, and the colour vocabulary actually present in the CSS.

Result: **0 changed out of 61** (16 named rendered, 16 named templates, 23 swept rendered,
6 swept templates). Query and table in the bug file.

### Misstep 1 — I nearly reported a 0 with no control, on a lane whose whole subject is bad greens

The first run came back 0/61 and my instinct was that it was obviously right (the fixer
had already swept these sites). That is the `a-post-fix-zero-needs-a-demand-control`
shape exactly. I added two positive controls and one negative, pushed through the
*identical* psql→base64→decode→transform path:

```
<style>.hero{background:#1a2b3c;}</style>                      → CHANGED  ✓
<style>.cta{background:linear-gradient(135deg,#1e40af,#1e3a8a);}</style> → CHANGED  ✓
<style>.hero{background:#ffffff;}</style>                      → UNCHANGED ✓
```

Only then is the 0 a zero that looked. **Cost: one minute. Do not skip it.**

### Misstep 2 — I "found" a contradiction in my own output that was not there

I read the diagnostic line *"the action's own SQL row filter does not match this body"* as
appearing on rows the SQL filter had just selected, called it a mirror-regex bug, and went
to debug it. It appears only on the `named_*` rows, which is correct and is in fact the
point — the components these items name are bodies the fixer would **never even SELECT**.
`grep -cP` on the decoded body (3 matches) and `od -c` on the bytes settled it in one
command. Two lessons, and the second is the real one: read the output rows you are
accusing before you accuse them; and a diagnostic that is *right* can still cost you ten
minutes if you skim it.

### Misstep 3 — my first slot resolution silently lost 4 of 16 targets

`spec.affected_component` is prose, not a key. My first pass stripped a trailing
`-section` and joined on `page_components.slot_name`; 4 of 16 resolved to nothing and I
briefly took that as "the item names a section that does not exist". It does not:

- `cta-section` → the slot is `call-to-action` (3 sites). Not a suffix rule at all.
- `features-section, differentiators-section` (leopardess, ONE item naming two) →
  `features` **and** `differentiators-section` — the second slot is *literally* named with
  the suffix, so stripping it breaks the one case stripping was invented for.

A join that quietly drops a quarter of its targets and returns rows for the rest looks
exactly like a join that worked. Now in `LANDMINES.md` with the footprint
`site_work_items.spec->>'affected_component'`.

### The three findings, in the order they arrived

1. **0/61** — the re-routing option is dead, and dead for a reason that generalises: the
   transform's entire output alphabet is `var(--color-primary)` and `var(--color-secondary)`
   and every one of the 15 items asks for `--section-*` / `--color-cta-*`. Vocabulary, not
   regexes (`2210aaeea` said this for 077 and it transfers verbatim).
2. **A false completion, caught by the re-detection loop.** `finetuning.uk` closed
   `complete` on 08-11 with the handler's own `total_fixed: 0` in the row, nothing on the
   page changed either side of it, and the audit re-filed the same item_key on 08-12. This
   is the first completed re-detection cycle for the type and it contradicts the completion
   it followed. It is also the strongest available answer to the coverage guard's stated
   posture for this type — the re-detection *does* fire here, unlike `contrast_failure`.
3. **The candidate verifier cannot be written as specified.** All 15 `acceptance_test`
   values read: 10 name a computed property, 2 contain clauses no probe can assess ("no
   visible seam", "visibly … or equivalent"), and the two filings of the same defect carry
   differently-worded tests. `criteria_check over acceptance_test` is a producer-side
   contract change, not a verifier.

### Cross-lane: 122 has decoupled, and the 08-12 handoff's "one round decides both" is stale

The `bugfix_122` lane costed the same fork on the same afternoon and reached option (4),
retraction on the discovery path — its handoff §3 banner says the standing objection kills
its options (1), (2) *and* (3), and §4 records that it has withdrawn "the 226 unpark when
213 closes" in our file. So D1 is our decision alone. Their option (4) transfers here and
is **cheaper here**: our re-detection loop demonstrably fires, our dedup key is page-level
so no per-section identity is needed, and `resolveWorkItems` already lives in the
producer's own package. Same precondition though — the audit must report which pages it
examined.

### Not chased, deliberately

10 of the 14 completions carry a payload that is not the handler's (a design-system spec
for 9, an unrelated child-page triage decision for 1). Recorded in the bug file as an
OBSERVATION with the mechanism marked NOT ESTABLISHED. It bounds fix candidate (2) to 4 of
14 rows, so whoever takes D1 needs to know it, but guessing at its cause here would be the
exact move this lane exists to punish.

---

## 2026-08-13 — gate 1b built and committed; and why option (2) is NOT ready to build

### What shipped (committed `96c53bc18`, INERT until the next chassis roll)

Completion gate 1b — `platform/orchestration/actions/complete_work_item_no_change.go`.
An opt-in, per-`item_type` gate that refuses to stamp `complete` when the handler's own
result payload reports it changed nothing. `dark_section_audit` is the only type on the
roster. Every other type takes a map miss and is byte-identically unaffected.

**The design turned on one fact that killed the obvious implementation.** I started to
write this as a verifier and it cannot be one: `VerifyTarget` carries the **spec**, not
the result, and `load_work_item_actions.go:871` reads the handler's report as an ACTION
INPUT which is marshalled into `site_work_items.result` at `:918` — *after* the gates run.
A verifier querying that column would have graded **the row's previous value** and looked
like it worked. The question "did the handler change anything?" can only be asked beside
`handlerReportedFailure`, which reads the same payload at the same moment for the same
reason. That is where it now lives, as gate 1b.

Three properties worth keeping in mind if you touch it:

- **No verifier is registered**, so `RegisteredVerifierItemTypes`, the coverage guard and
  the `sql_for_agents/220` claim-timeout exclusion are all untouched. That is a real
  advantage of doing this at gate 1b rather than as a verifier — no registry lockstep to
  keep.
- **The third arm ABSTAINS rather than guessing.** If the declared counters cannot be
  resolved, the item completes (an unreadable payload is not evidence of a no-op) and the
  abstention is recorded to `agent_error_log`. This arm is *live today*, not defensive
  boilerplate: 10 of the 14 completed rows carry a payload that is not this handler's.
  Instrumenting that split is what §D of the bug file said to do instead of theorising.
- **`lookupNumericPath` accepts `float64|float32|int|int32|int64|json.Number`** and that
  breadth is load-bearing, not defensive. A missing arm reads "counter absent" for a
  counter that is present and zero — i.e. reports *unknown shape* where the data supports
  *block*. Both numeric arms are mutation-proven below for exactly that reason.

### Mutation matrix (one at a time, and the first attempt was worthless)

| mutation | result |
|---|---|
| delete the `dark_section_audit` roster entry | **RED** |
| remove the any-non-zero early return | **RED** |
| break the `json.Number` arm | **RED**, on exactly that case |
| break the `int` arm | **RED**, on exactly that case |
| restored | GREEN |

**The misstep, and it is the same class the RUNBOOK already warns about.** My first attempt
at the `json.Number` mutation *deleted* the case arm — which made `encoding/json` an unused
import, so the package failed to **compile**. `FAIL` appeared, I nearly recorded it as a
pass, and it proved nothing at all: a build error is not a test detecting anything. Redone
so the arm still compiles but returns `(0, false)`, it goes red on precisely the one case.
**A mutation must leave the program buildable or it tests the compiler, not your test.**

**NOT PROVEN, stated because nothing else will say it:** no test asserts that
`verifyBeforeComplete` actually *calls* the gate. The wiring needs a `*sql.DB` for the
item_type lookup, and a source-scanning test would make comments load-bearing (own
landmine). The behavioural check owed after the roll is in the RUNBOOK.

### The council submission did NOT dispatch — kubeconfig expiry

Authored, validated and fired. The script printed
`SUBMISSION_CORR=4c2028f6-c3c6-4dbf-9113-5ebc8705c7b2`, then its `kubectl -n kafka run …
kcat -P` publish (line 170) returned **`Unauthorized`**. The token has expired fleet-wide;
`kubectl get pods` fails the same way. **So no council run exists**, and the printed
correlation names nothing.

I committed with **no trailer at all**. `Council-Submitted:` asserts that a submission was
made; this one was not, so writing it would be the same false claim the `Council-Reviewed:`
rule forbids, one rung down. The payload is ready at
`scratchpad/213_d1_gate1b_submission.json` — re-fire it verbatim once the owner refreshes
the token, and record the *new* correlation here.

Two schema notes for whoever re-fires (neither is in the RUNBOOK's list):
`.plan.risks` must be a **STRING**, not an array — the script refuses an array outright
(`must be a STRING (Go: string) … join the risks into one prose block`). And the
correlation is printed **before** the publish, so **a printed `SUBMISSION_CORR` is not
evidence of a dispatch.**

### Option (2), the discovery-path retraction: a SECOND precondition, specific to this type

The 08-12 note recorded that 122's option (4) transfers here and is "cheaper here than
there". The first half stands; the second half was **too quick, and I am correcting it
before anyone builds on it.**

`WriteAuditFindingsAction` (`:509-545`) takes its findings from an **LLM response**
(`audit_result.result`, parsed out of JSON, ````json` fences and all). It records no set of
pages examined; `loadSitePages` loads the site's page inventory for *classification*, which
is a different thing.

So this type has the 122 blocker **and one 122 does not have**:

> `contrast_failure` retracts on the silence of a **measurement** — a browser computed a
> contrast ratio and the bad pairing is gone. `dark_section_audit` would retract on the
> silence of an **LLM**. A model that does not mention a defect on run N+1 has not
> established that the defect is gone; it may simply not have said so this time.

Direct evidence that this audit's output is not stable text: the two `finetuning.uk`
filings of the **same defect on the same component**, one day apart, carry differently
worded `description` and `acceptance_test` values. That proves the *wording* varies. It
does **not** prove the *finding set* varies, and I am not asserting that it does.

**[UNVERIFIED — needs the DB, blocked on the token]** On 08-11 the audit filed 14 items
across 14 sites; on 08-12 exactly **one** was re-filed, although nothing repaired any of
the other 13 (0 of 61 bodies changed). Two candidate explanations and I cannot presently
separate them: the rotation visited only one site that day (likely — the quality rotation
runs ~3 site passes/day), or the audit's finding set is unstable. **Separating them is the
first task of option (2)**, and the measurement is cheap: group the `dark_section_audit`
rows by `batch_id` and `site_id` over consecutive runs, and check whether a site the audit
demonstrably re-visited re-filed its finding. If the finding set turns out to be unstable
on an unchanged site, retraction here would close real defects on model variance, and
option (2) needs a different design — not a `pages_audited` list.
