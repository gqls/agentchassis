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

---

## 2026-08-13 (evening, token restored) — the stability measurement, and the two ways I got it wrong first

### The result [MEASURED 2026-08-13]

**The design audit re-reported the colour defect on 7 of 7 post-closure re-visits, across
4 sites. Zero silences.**

```sql
-- pairs a CLOSED colour finding with a LATER audit day on the same site, and asks
-- whether the colour finding came back. Matched on site + page, NOT on item_key.
WITH colour AS (
  SELECT site_id, spec->>'page_name' AS pg, item_type, created_at, completed_at
  FROM site_work_items
  WHERE spec->>'audit_source'='design-audit'
    AND item_type IN ('dark_section_audit','hardcoded_section_colors')
), visit_days AS (
  SELECT DISTINCT site_id, created_at::date AS d
  FROM site_work_items WHERE spec->>'audit_source'='design-audit'
)
SELECT s.domain, c.completed_at::date, v.d,
       EXISTS (SELECT 1 FROM colour r WHERE r.site_id=c.site_id
                AND COALESCE(r.pg,'')=COALESCE(c.pg,'') AND r.created_at::date >= v.d)
FROM colour c JOIN visit_days v ON v.site_id=c.site_id AND v.d > c.completed_at::date
JOIN sites s ON s.id=c.site_id WHERE c.completed_at IS NOT NULL ORDER BY 1,2;
-- dartsonline 08-09→08-11 (×2) · finetuning 08-09→08-11, 08-09→08-12, 08-11→08-12
-- · leopardess 08-08→08-11 · webdesign.co.uk 08-08→08-11 — all TRUE
```

The dedup confound is handled: every earlier row is `complete`, so a re-file was not
suppressed by `idx_swi_dedup`. And these are findings we independently know were **never
repaired** — 0 of 61 bodies change under the handler's literal transform — so a silence
would have been a true miss, not a true retraction.

### What 7 of 7 actually licenses, which is less than it looks

**Zero misses in 7 bounds the miss rate at ~35% (95% upper), not at zero.** For a single
silent run to be adequate evidence at a ≤5% miss rate you would need roughly **60**
consecutive clean re-detections. So:

> The detector is **not** unstable in the way I feared on 08-13 morning — the design
> question is not blocked. But 7 observations do **not** license retracting a finding on
> ONE quiet audit. Item 2 must require **N consecutive silences**, or pair the silence with
> a deterministic check. A single-silence design would be building on a bound of 35%.

This supersedes the 08-13 morning note, which said the measurement had to be run before
the design could proceed. It has been run; the design proceeds, with the trigger changed
from "the audit did not re-report it" to "the audit did not re-report it N times running".

### MISSTEP 1 — my widened query said 62% silence and it was junk

First attempt widened to the whole `design-audit` producer and returned **209 re-visit
pairs, 80 refiled, 129 silent**. I nearly wrote that down as "the auditor forgets 62% of
the time". It is an artefact, twice over:

- **Batch granularity.** Each site gets a burst of 2–3 `batch_id`s within ~2 minutes
  (finetuning: 13:21:00, 13:21:55, 13:23:10). Those are steps of one pass, not three
  visits, and the later ones are different audit steps that would never re-file the first
  one's finding. Treating each as a "visit" multiplied every finding by 2–3.
- **Only 5 sites and 46 distinct findings** produced those 209 pairs. A big-looking N built
  from a small population.

Collapsing visits to distinct DAYS is what makes the number mean anything.

### MISSTEP 2 — the fixed query then told me the OPPOSITE of the truth for my own item type

The day-granularity version, matched on `item_key`, reported:

```
 hardcoded_section_colors |  6 revisit_days |  0 refiled |  6 went_silent
 dark_section_audit       |  1              |  1         |  0
```

Read literally: the colour finding goes silent every single time. **That is exactly
backwards**, and the cause is my own lane's Half A: the fix RENAMED this producer's output
from `hardcoded_section_colors` to `dark_section_audit`, and `item_key` embeds the
item_type. So a re-file under the new name cannot match the old key, and a perfect
re-detection reads as a silence.

**The lesson, which is more general than this bug: a join key that contains a value your
own change renamed will read as absence, and absence is the finding.** I caught it only
because I had the by-hand pairing from the earlier query to contradict it. Matching on the
DEFECT (site + page + either type) is what the question actually asks.

### Recorded, not resolved: other design-audit types DO go silent

At day granularity, matched on item_key (sound for types this lane did not rename):
`needs_content_planning` 0 refiled of 5, `tone_shift` 0 of 2, `cta_improvement` 6 of 11,
`responsive_fix` 1 of 3, against `needs_design_review` 12 of 15 and `content_rewrite` 6 of 8.

**I cannot separate "genuinely repaired" from "not re-reported" for any of those**, because
unlike the colour findings I have no independent evidence that the defect survived. So this
is not evidence of instability — it is evidence that **the 7-of-7 result is specific to the
colour findings and must not be generalised to the producer.** Anyone extending retraction
to another design-audit type owes the same measurement for that type.

### And then the precondition failed anyway — `page_name` is prose too [MEASURED 2026-08-13]

Stability was the blocker I expected. It cleared. The one I did **not** expect is that the
field retraction would have to be scoped by cannot carry the weight:

```sql
SELECT COALESCE(NULLIF(spec->>'page_name',''),'<none>'), count(*), count(DISTINCT site_id)
FROM site_work_items WHERE spec->>'audit_source'='design-audit' GROUP BY 1 ORDER BY 2 DESC;
--  index 140/19 · about 29/15 · contact 28/13 · all 17/9 · global 8/4 · services 7/4
--  · "all pages" 5/3 · sitewide 4/1 · "index / about" 3/1 · "index, services" 2/2
--  · site-wide 2/2 · "case-study-multi-agent…, case-study-news-pipeline, case-study-…" 3/1
```

`page_name` is **free prose**, exactly like `affected_component` (yesterday's landmine) and
exactly like `acceptance_test`. `all`, `global`, `sitewide`, `site-wide` and `all pages`
are five spellings of one idea; `index / about` and `index, services` are two pages in one
string; one row carries three case-study slugs comma-joined.

Three consequences, in order of how much they cost:

1. **Page-scoped retraction cannot be built on this field.** "The audit examined page P and
   did not report a dark-section defect" is unanswerable when P may be the string `all`.
2. **The dedup key is already degraded by it.** `item_key` is
   `{source}_{item_type}_{page_name}_{site}` (`write_audit_findings_action.go:291`), so the
   SAME defect described as `index` on Monday and `sitewide` on Tuesday mints two different
   keys, both open, neither deduping the other. Not hypothetical — the spellings above are
   from live rows.
3. **It bounds my own 7-of-7 result**, and I am saying so rather than leaving it implied: I
   matched on `COALESCE(page_name,'')` equality. Every colour finding happens to carry
   `index`, so the match is sound *for those rows*. It would not be sound for a producer-wide
   version of the same query, and anyone widening it must resolve the spellings first.

**So item 2's blocker is NOT the 122 lane's.** Theirs is "the audit does not report which
pages it examined" — a missing field, fixable by adding one. Ours is "the page identifier it
DOES report is prose", which is a producer-side contract change or a different scope.

**The design that survives this, and it is what I would submit:** scope retraction at the
**SITE**, not the page, for this item type only — *the audit visited this site N consecutive
times and reported no dark-section defect → retract*. It never has to resolve a page name.
It is coarser, and the coarseness is bounded by fact rather than hope: all 15 live
`dark_section_audit` findings name `index`, and `index` is the most-audited page in the
fleet (140 findings across 19 sites), so a site visit that examines nothing is not a case
that arises. With N ≥ 3 the miss-rate bound from 7-of-7 (~35% per run) compounds to under
5%, which is the number that licenses the retraction — **that is why N is 3 and not 1, and
it should be stated that way in the submission rather than asserted as a safety margin.**

Not built. It is a second shared-seam change and needs its own council round; gate 1b's is
still in flight, and stacking a second submission on an unread verdict is how a lane ends up
defending two designs at once.

---

## 2026-08-14 (afternoon) — gate 1b PROVEN IN PRODUCTION, and the sweep re-enabled at the owner's instruction

### The behavioural proof, which no test could give

Owner authorised both actions. **They are not the same action**, and the sweep would not have
delivered the proof: `improvement-sweep`'s `pre_query` takes **one site per fire**
(`LIMIT 1`, `ORDER BY sites.updated_at ASC NULLS FIRST`, 900s interval), so
`mortgagecalculator.co.uk` might have waited days. The controlled route is the dispatcher that
was already running: **`build-pipeline-trigger`, enabled at 60s**, whose `pre_query` fires on any
site holding a `triaged` + `pipeline='build'` item with attempts remaining.

So: promote the ONE waiting row `detected` → `triaged` (id-scoped UPDATE, `AND status='detected'`
so it cannot fire twice) and let the existing dispatcher claim it.

```
14:09:25Z  promoted to triaged
14:11:26Z  triaged | attempt 1 | _verification.status = handler_reported_no_change
           error: "completion blocked: the handler reported it changed nothing, so this cannot
                   be a repair (bugs_open/213 D1): handler reported 0 changes at
                   response.fix_result.total_fixed and response.text_color_result.total_fixed — …"
           result.response.fix_result.total_fixed        = 0
           result.response.text_color_result.total_fixed = 0
14:13:04Z  attempt 2
14:14:41Z  failed  | attempt 3  ← TERMINAL
```

**Both directions, in one window.** The gate fired on the BLOCK path (not the abstain path — the
counters were present and zero, which is what the 0-of-61 transform measurement predicted), and
the retry cycle **terminated** at `failed` after the three permitted attempts rather than
churning. That termination was a claim in my council risks block; it is now measured, not
asserted. Three dispatches, ~3m15s wall clock. **Before the gate this row would have been
stamped `complete`** — the finetuning.uk shape exactly.

The error text carries the roster's `Why` verbatim, so an operator reading the item's error
column gets the reason and the measurement, not a code.

### I doubted my own cost claim, checked it, and it held

The 6-minute dispatch window showed **7 LLM calls** against a baseline of **5–9 per HOUR** — so
on its face my council submission's claim ("pure SQL and regex with no LLM spend") looked false.
Attributed: **6 `council-gate` (another lane's round) + 1 `content-gap-planner`, and ZERO from
`color-variable-fixer` or `improvement-loop`.** The claim holds.

Worth keeping: **a raw count in a window would have incriminated my dispatches.** The
attribution query is what settled it, and the trap is the mirror image of the usual one — not "a
count proves the damage" but *a count proves damage that was somebody else's*. In a fleet this
busy, any before/after count over a shared surface needs a `GROUP BY agent_type` before it means
anything.

### The sweep: what one fire actually did

`enabled=true` at 14:15Z; it was two days overdue and fired within seconds. First fire:
**one site**, 13 findings across 13 item types, 1 claimed, 1 triaged, and **zero attributable
LLM spend**. One of the 13 is a new `dark_section_audit`, so a 17th row now exists and will meet
gate 1b on a later fire.

⚠ **That is ONE FIRE, not a rate, and it must not be quoted as a cost measurement.** The 122
lane's 3.2x was measured over 5h29m across many sites, and the LLM-heavy work is in the audit
agents this fire did not reach. Their lane has been told the switch is back on — appended to
their own handoff, with the baseline figures and the note that their dated 08-16 ramp
measurement now has a second driven mechanism in the window.

### Also measured, and it moves the "up to 15 items will fail" risk from prediction to fact

`dark_section_audit` now: 17 rows · 1 gate-blocked · 1 `failed` · **0 abstain records**. The
abstain arm has not yet fired, which is the one prediction in WII-017's landmine still
outstanding — 10 of 14 historical rows carried an unreadable payload, so it should appear as more
of these dispatch. Watch `agent_error_log` under `NO_CHANGE_GATE_UNREADABLE_RESULT`.

> ### ⚠ CORRECTED 2026-08-14, ~15 minutes after writing the section above — "zero attributable LLM spend" was MEASURED TOO EARLY and is FALSE
>
> The claim above, that the sweep's first fire cost no attributable LLM spend, was true when I
> ran the query at **+1 minute** and false by **+5 minutes**. **The sweep dispatches
> asynchronously and its agents write `llm_call_log` on completion**, so a cost query run
> immediately after a fire measures the dispatch, not the work.
>
> **The corrected figure, with the control that makes it attribution rather than coincidence:**
>
> ```sql
> SELECT agent_type,
>        count(*) FILTER (WHERE created_at BETWEEN '2026-08-14 10:00:00+00' AND '2026-08-14 14:15:00+00') AS calls_4h_before,
>        count(*) FILTER (WHERE created_at > '2026-08-14 14:15:00+00')                                    AS calls_since_sweep,
>        sum(input_tokens) FILTER (WHERE created_at > '2026-08-14 14:15:00+00')                            AS in_tok_since
> FROM llm_call_log GROUP BY 1 ORDER BY 3 DESC;
> ```
>
> | agent_type | 4h before | since sweep | input tokens since |
> |---|---|---|---|
> | `content-quality-auditor` | **0** | 2 | 3,850 |
> | `site-review-agent` | **0** | 1 | 6,807 |
> | `visual-design-auditor` | **0** | 1 | 2,321 |
> | `tool-acceptance-agent` | **0** | 1 | (null) |
> | `content-gap-planner` | 24 | 1 | 4,488 — **already active, NOT attributable** |
> | `council-gate` | 16 | 5 | 21,930 — **another lane's rounds, NOT attributable** |
>
> **Four agent types that were completely idle for the preceding four hours started within five
> minutes of the switch.** That zero-before column is the control, and it is what turns this
> from "calls happened" into "the sweep caused them". Attributable to the first site pass:
> **~12,978 input tokens across 5 calls.**
>
> **What that does and does not license.** Naively, ~13k input tokens per site pass at one pass
> per 15 minutes ≈ **~52k input tokens/hour** sustained, against a measured baseline of
> 22k–101k/h — so it roughly doubles a quiet fleet's floor and sits far below the **806k/h** the
> 122 lane measured while driving it. But this is **one pass**, the spend may not all have
> landed even now, and their 3.2x was measured over 5h29m with more sites due. **It is still not
> a rate.** The honest summary is: the first pass is priced, the hourly figure is an
> extrapolation from n=1, and only a few hours of running will settle it.
>
> **The transferable rule, and it is the second instrument failure in this lane in two days:**
> **an async mechanism's cost cannot be measured at t+1 minute.** Wait for the work to land, or
> measure a window that closes after the dispatched agents complete. And keep the before-column
> in the same query — without it, `council-gate`'s 5 calls and `content-gap-planner`'s 24 prior
> calls would have been read as sweep spend, which is the same misattribution in the other
> direction. Logged in `WRONG_CALLS.md`.

### The REAL sweep cost, 2h20m in — 6.0x baseline, and my n=1 extrapolation was 6.5x too LOW

[MEASURED 2026-08-14 16:35Z, 2h20m of sweep running, attribution by the zero-before control]

| | input tokens/hour |
|---|---|
| baseline (4h15m before the switch) | **56,480** |
| since the switch | **339,457** |
| **ratio** | **6.0x** |

My earlier extrapolation from the first pass said ~52k/h. The real figure is **339k/h**. The
extrapolation was not merely imprecise, it was **6.5x low**, and the reason is instructive: the
first pass I priced had not yet reached the expensive stage. Attributable spend, zero calls in
the prior 4h15m in every case:

```
page-content-writer       57 calls   541,676 in_tok   <- 80% of the attributable total
site-review-agent         10          38,296
content-quality-auditor   20          16,963
webdesign-agent            8          15,610
component-template-fixer  10          15,215
visual-design-auditor     10          13,552
feed-triage / tool-improver / tool-auditor / tool-acceptance / med-price-collector  (small)
```

`council-gate` (20 calls) and `content-gap-planner` (15) were active before and are NOT
attributable — the before-column is what separates them.

**`page-content-writer` is the cost.** The sweep is not just auditing; it is triggering page
content rewrites, and those dominate everything else by 14x. **This is also worse than the 122
lane's measured 3.2x**, which they took over 5h29m — so it is not that today is unusually
expensive, it is that this is a higher multiple than the figure the owner was given when they
authorised the switch.

**Three lessons, and the middle one is the expensive one:**
1. **A first-pass price is a lower bound on a pipeline, not an estimate of it.** The cheap
   stages complete first and log first. Anything staged will read low early, and the error is
   unbounded — here 6.5x.
2. **I gave the owner a number (3.2x, from another lane's doc) as the basis for a decision, and
   the real number for the same action was 6.0x.** The 3.2x was measured on a different day
   against a different baseline. **A cost ratio does not transfer between days** — the
   denominator moves. Quote the ratio AND both absolute figures, or quote neither.
3. **The dominant cost was invisible in the first pass and obvious in the census.** One
   `GROUP BY agent_type ORDER BY input_tokens DESC` found it immediately. Do that before
   extrapolating anything.

### Sweep OFF at 16:41:46Z, owner's decision on the corrected number. Final accounting

Ran **14:15:23Z → 16:41:46Z, 2h26m**. Whole-window fleet total: **164 calls, 807,704 input
tokens, 198,912 output tokens ≈ 331k input tokens/h**, against a 56,480/h baseline. Owner was
given the 6.0x figure and chose to switch it off; that reverses my own enable, not a decision of
theirs made on good information — the 3.2x I had quoted was another lane's ratio against a
~248k/h baseline and never transferred to a 56k/h one.

⚠ **Switching the scheduler off does NOT stop work already dispatched.** At the moment of the
UPDATE there were **1 `claimed` work item and 3 orchestrations still EXECUTING_STEP**. They
finish on their own and their spend lands afterwards, so a cost query run immediately after the
switch understates the window — the same lag that made my first-pass figure 6.5x low, in the
other direction. Anyone verifying the stop should check `claimed` count and running
orchestrations, not just `enabled=false`.

### What the sweep bought, since it was not free

Not nothing, and worth recording so the spend is not written off as pure waste:

- **Gate 1b's BLOCK arm proven** on real traffic beyond the single deliberate dispatch: 3 items
  blocked (2 `failed`, 1 `triaged`).
- **Gate 1b's ABSTAIN arm proven — the one untested arm in WII-017.** 4 items completed
  unblocked, and **all 4 have a matching `agent_error_log` row**, timestamps agreeing to
  milliseconds (14:24:46.072 gate record vs 14:24:46.079 completion). **A 1:1 accounting with no
  silent holes**, which is the property that mattered.
- **The §D payload split reproduced live, 3:1**, matching the historical 9:1 — so it is
  systematic, not a historical artefact, and the instrument recorded the shapes:
  `color_scheme design_notes spacing typography` (3) and
  `add_to_page approach new_page not_actionable reasoning retype_existing update_spec` (1).
- **Enough evidence to file the §D diagnosis properly**, which had sat NOT ESTABLISHED for days.

### §D: filed to the diagnosis loop, and the trigger has the SAME trap as the council one

`090` filed, **RUN_CORRELATION_ID `266be67d-a6e1-4afc-8fc1-84b553b2ea82`** (use this, not the
intake correlation, for artifacts). Prior art checked first: the `needs_diagnosis` queue was
empty and neither bug directory carried the mechanism.

What I could establish first-hand and put in the symptom: neither `color-variable-fixer` **nor**
`build-dispatch-loop` declares a `complete_work_item` step in
`agent_definitions.default_config->'workflow'->'steps'`, yet `agent_error_log.agent_type` on
every abstain row is `build-dispatch-loop`. So the site that binds the `result` input is
unidentified — which is exactly the "the cause is not where the symptom is" shape the loop exists
for. I stopped there rather than guessing.

⚠ **NEW TRAP, and it is the third instance of this shape in two days: the `090` trigger prints a
correlation and THEN fails.** My first attempt died on `invalid input syntax for type json`
because the symptom text contained **double quotes** (`inputs.GetMap("result")`) which the script
does not escape. It had already printed `SAVE: CORRELATION_ID=…`. **No work item was written.**
Same lesson as the council trigger's `Unauthorized`: **a printed correlation is not evidence of a
filing.** Keep double quotes out of a `090` symptom, and check the item exists before trusting
the id.

### THE FLEET'S LLM CAPABILITY IS DOWN UNTIL 2026-09-01 — and it is NOT the sweep's fault

Found while chasing why my `090` verdict step failed. `llm_call_log.success` by hour:
**14:00 → 98 ok / 2 failed · 15:00 → 49 / 13 · 16:00 → 0 / 17.** Every call now fails with

```
status 400 invalid_request_error: "You have reached your specified API usage limits.
                                   You will regain access on 2026-09-01 at 00:00 UTC."
```

**I nearly reported that my sweep caused it.** The timing invited it — the cap bit at ~15:36Z,
80 minutes into a burst I had enabled and measured at 6.0x baseline. **It is wrong by two orders
of magnitude.** The cap is **MONTHLY**, and August had already consumed **~221.9M input tokens**
by the 14th:

```
08-09  42.7M   08-08  33.2M   08-10  26.2M   08-04  24.9M   08-02  20.9M   08-03  15.9M
...    08-13   1.7M   08-14   2.1M  (daily spend had FALLEN an order of magnitude)
```

The sweep's entire attributable contribution was **~0.71M ≈ 0.3% of the month.** It did not
cause this and did not meaningfully bring it forward.

**The rule, and it is the fourth instrument lesson of this lane in three days: a period-scoped
limit needs the PERIOD's denominator.** A day's spend — however dramatic, however recent, however
much it was your own doing — is not evidence about a month's cap. The dramatic local number and
the guilty conscience point the same way, which is exactly when to go and get the denominator.

**Two consequences for this lane:**

1. **The `090` on §D cannot complete.** It built a 54,805-char evidence bundle and died at the
   `verdict` step; the item is back at `triaged` and `diagnose-pipeline-trigger` will retry it
   every 60s, **failing identically each time**. Correlation
   `266be67d-a6e1-4afc-8fc1-84b553b2ea82` stands, and the bundle is worth reading by hand — the
   evidence was gathered before the wall.
2. **Gate 1b is unaffected and remains fully proven.** It is pure Go with no LLM call, which is
   now an accidental virtue: it is one of the few completion-path checks that still works. Its
   proof was completed at 16:35Z, before the wall mattered.

⚠ **Do not read any LLM-backed check's "0 findings" as clean until 09-01.** That includes
`verifier-remit-check`'s daily report, the council gate, and every audit rotation. Filed as a
fleet-wide entry in `LANDMINES.md`.

> ### ✅ CORRECTED 2026-08-15 — THE WALL LASTED ~90 MINUTES, NOT TWO WEEKS, AND THIS SECTION'S HEADING WAS FALSE WITHIN THE HOUR IT WAS WRITTEN
>
> [MEASURED 2026-08-15 09:30Z] `llm_call_log` by hour, capped-call counts:
> **08-14 15:00 → 11 capped · 16:00 → 17 capped · 17:00 → 0 capped, 24 successes.**
> Zero capped calls in every hour since; the 8 hours to 09:00Z today are **0 failed of 131**.
> The entire outage was **~15:36Z → ~17:05Z on 2026-08-14**. The owner raised the cap.
>
> **So this section's own §"two consequences" is retired:** the `090` on §D did not merely
> fail once, it became **re-runnable ninety minutes later, and nobody re-ran it for a day.**
> Re-fired 2026-08-15 as run correlation `adecf408-1e60-4293-8b22-351ddbb52a08`. And the
> council gate was never unavailable either — this lane's half-two submission ran
> **13 seats to APPROVED in 11 minutes** on 2026-08-15.
>
> **What I got wrong, and it is not the diagnosis — that was excellent.** The attribution work
> in this section is right and worth keeping: the cap is monthly, August had already burned
> ~221.9M input tokens by the 14th, and the sweep's ~0.71M was 0.3% of it. **The error is
> that I took the vendor's stated reinstatement date as a FACT ABOUT THE FUTURE** and wrote
> it into a heading, a warning, and a fleet-wide file. `LANDMINES.md` had **already recorded
> this exact correction twice** — the 2026-07-31 cap stated 08-01 and the 2026-08-10 cap
> stated 2026-09-01 and cleared in ~2–3 hours, both times because a human raised it, and the
> entry says in terms *"the stated reset is the vendor's worst case, not a forecast"* and
> *"re-run the absence-of-success query for the current hour before concluding you are
> blocked"*. **I hit the third recurrence and repeated the mistake the second one had already
> been written up to prevent.** Reading the landmine file for the thing you are ABOUT to
> touch is the habit; I only grepped it for the cap's *signature*, not for its *duration*.
>
> **The check, which takes one query and settles it:** verify a lift on the SUCCESS side, never
> the failure side — failures simply stop appearing whether or not capability returned.
> ```sql
> SELECT date_trunc('hour',created_at) hr, count(*) FILTER (WHERE success) ok,
>        count(*) FILTER (WHERE error_message LIKE '%usage limits%') capped
> FROM llm_call_log WHERE created_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
> ```
> Logged in `WRONG_CALLS.md`. The fleet-wide `LANDMINES.md` entry was already struck through
> and corrected by another thread on 08-14; **this lane's copy is what stayed wrong for a day**,
> which is its own lesson: correcting the shared file does not correct the doc that quoted it.

---

## 2026-08-15 — D1 HALF TWO BUILT, 213 AND 216 CLOSED ON THE OWNER'S RULING

Owner instruction: *"213 half two can proceed here, we can close 213 and 216."* All three done.
Commits: `a620912f5` (half two + shared helper + WII-016 migration + register), `d103dfcea`
(216 closed), `0c467cea3` (213 closed), `0d40f25ad` (2 landmines + 1 wrong call),
`e3d61d7d4` (council objections actioned), `dbe29bbd6` (register verdict update).

### The design was decided by measurement, and two of the three constraints came from data I went looking for rather than data I had

- **SITE-scoped, not page-scoped** — `spec.page_name` is free prose. [MEASURED] live values in
  the item_keys include `index`, `all` and `all pages`; the last two were filed **on the same
  day**, 08-14. Nothing resolves that to a page.
- **Silence is SITE-level, not per-key** — and this is the one I nearly got wrong. The obvious
  design counts silences per item_key. [MEASURED] the `audit_source` literal was **renamed
  `design-audit` → `visual-design-audit` between 08-12 and 08-13**, and it is embedded in the
  key: gamesdesign.co.uk holds two rows for the same defect on the same page under two keys.
  A per-key rule would have read that single rename as **fifteen defects being fixed at once**.
  Found by listing the rows before designing, not by reasoning about the key format.
- **N = 3** — arithmetic from the lane's existing 7-of-7 measurement, unchanged: the 95% upper
  bound on the per-run miss rate is `1-0.05^(1/7)` ≈ 0.35, so N=1 → ~35%, N=2 → ~12%,
  N=3 → ~4.2%, the first under 5%.

### The near-miss that was NOT in the design, and would have been invisible

The silence streak has to live somewhere. `site_work_items.result` is the obvious home. Before
writing the UPDATE I went looking for readers of `updated_at` on that table — and found the
**`stale-work-item-reaper`**, enabled, hourly, parking `triaged` rows whose `updated_at` is
older than 48h. `site_work_items` carries `trg_site_work_items_updated_at` (BEFORE UPDATE), so
**any** write bumps that column and no column list avoids it. With the sweep at 900s, a streak
write would have made every `triaged` dark-section row **permanently unreapable** — the exact
queue deadlock that reaper exists to prevent, and migration `237`'s own header predicted it in
the abstract as a risk *of that migration*, where nobody touching `result` would ever read it.

**The damage would have been an ABSENCE — a park that never happens.** No error, no bad value,
nothing to inspect; and [MEASURED 2026-07-27, which is why 237 could not be tested] there is
essentially no `triaged` backlog at any moment because the claimer drains it in ~2 minutes, so
there is no population in which to notice. Remedy is `workItemInFlightStatuses` = {`triaged`,
`claimed`}. **Two reasons support it and they are not equal** — "the pipeline is carrying the
item" is arguable, "the reaper keys on this column and the trigger bumps it" is mechanical.
The comment states the mechanical one as load-bearing, or the next reader relaxes the rule on
the strength of the softer one. Filed fleet-wide in `LANDMINES.md`.

### THE MISSTEP: I wrote a test that proved the mock, not the guard

`TestAuditRetraction_UnrecognisedReplyIsNotSilence` set no `ExpectBegin` and asserted no
retraction came back, with a comment saying the transaction must never open. It passed. **It
passes identically with the guard deleted**: the code opens the transaction, sqlmock refuses
the unexpected call, the action logs the error and returns no retraction key, and the
assertion is satisfied by the mock's refusal instead of the code's guard. Two opposite worlds,
one identical pass — and the five other tests in the same file all discriminated properly,
which is why it did not stand out.

**Caught by the mutation matrix and by nothing else.** Not by review, not by re-reading it.
Fixed by testing the FUNCTION directly with a mock carrying no expectations, so a stray call
becomes an error RETURN the assertion can see; paired with the opposite-direction test so
"guard always on" is caught too. Logged in `WRONG_CALLS.md` and `LANDMINES.md`.

Two more results from the same run, both worth keeping:

- **A mutation that passes may have hit a guard in SERIES.** Removing WII-016's
  `len(PagesAudited)==0` short-circuit failed nothing, because a second `len(audited)==0`
  return sits behind it. That was true before my migration too. The fix is not to delete the
  guard but to pass the real condition down to the shared helper, so each can be mutated
  singly — which is why `observed` is now `len(payload.Summary.PagesAudited) > 0` and not `true`.
- **When a mutation genuinely is not caught, delete the code.** The streak's own `audit_source`
  re-check was unreachable by construction (the caller's spec scope already guarantees it).
  I removed it rather than writing a test to justify it. Untested defence reads as protection
  and is not.

Matrix discipline used throughout, per this lane's earlier lessons: gated on **exit status**
(not a `grep '^\s+--- FAIL'`, which sees only subtests), an **unmutated control**, and a
**byte-identity check** of every mutated file afterwards. 9 mutations, 8 caught.

### The council round: 11 minutes, APPROVED, and three objections were checkable

Corr `54e3b698-3d18-4dd1-9d6f-badec7e331fa`. Dispatch **verified in `orchestration_states`
before waiting** — this lane has been bitten three times by a printed correlation that named
nothing. 13 seats, 5 advisory objections, none high.

- **`editquality` asked the question I should have asked myself:** the producer-scope guard
  skips rows whose `spec.audit_source` differs from the run's literal — so does it **strand**
  the 4 rows this whole change exists to release? I had disclosed the rename as a *risk* and
  never checked it against the population. [MEASURED] all 4 carry `visual-design-audit`, and
  the live literal is hardcoded in the auditor's own `query_database` step as exactly that.
  They match; the guard releases them. **It could have come out the other way.**
- **`bug_historian` on truncation:** a well-formed but truncated reply parses as
  recognised-and-empty and would advance a streak wrongly. [MEASURED] `visual-design-auditor`:
  **4,088 calls, 0 at `max_tokens`**, max output 1,869 against a 4,000 cap. Real risk, zero
  occurrences, ~2.1x headroom — recorded so a cap or prompt change re-raises it.
- **`guardian` found a real hole in the evidence.** Hoisting classification above the filters
  changed the loop **all six** producers run through, and my tests pinned only
  `dark_section_audit`. Added two regression tests on a non-gated type (`cta` →
  `cta_improvement`). That objection earned its round.

### And a stale claim of this lane's own, corrected today

The 08-14 section above declared the fleet's LLM capability **down until 2026-09-01**. It was
false within the hour: [MEASURED] the cap bit for ~90 minutes and cleared at ~17:05Z on 08-14.
`LANDMINES.md` had already recorded that exact correction **twice**, for two earlier caps, and
says in terms that the stated reset is the vendor's worst case rather than a forecast. Full
correction inline above. **Correcting the shared file does not correct the doc that quoted it** —
the fleet entry was struck through on 08-14 and this lane's copy stayed wrong for a day.

---

## 2026-08-15 (later) — half two EXERCISED on production traffic, WITHOUT re-enabling a carrier

Owner asked to either flip `improvement-sweep` on and off again, or trigger the audits by hand.
Took the second. Getting there corrected two claims this lane had been repeating.

### ⚠ CORRECTION — the carrier claim in every handoff of this lane is WRONG

Stated everywhere, including `HANDOFF_2026-08-15`: *"`improvement-sweep` and
`site-discovery-rotation-design` … are the only carriers that dispatch this audit."*
[VERIFIED first-hand] `site-discovery-rotation-design` drives `design-discovery-agent`, whose
live workflow has **three steps** (`ensure_site_record, run_checks, complete`) and which does
**not** appear among the agents whose config contains `write_audit_findings`:

```sql
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%write_audit_findings%'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- brief-fidelity-auditor, content-quality-auditor, council-gate, council-gate-036scratch,
-- fix-proposer, offer-analyser, site-review-agent, visual-design-auditor
```

The real chain is `improvement-sweep → improvement-loop → call_design_audit →
design-audit-agent → call_visual_auditor → visual-design-auditor → write_findings`.
There is exactly **ONE** carrier, not two, and the second-named one could never have exercised
this code at all.

**It would also have done nothing if enabled.** [MEASURED 13:41Z] its `pre_query` requires a
rotation stamp older than 7 days; the oldest live stamp is `robot-hands.com` at **6d04h**, so
the SELECT half returns **0 rows**. Enabling it would have fired, selected nothing, dispatched
nothing — and looked exactly like a working test that found nothing to do.

### Why the enable/disable route was refused

`improvement-sweep` is not detection-only. Its own description: *"Discovery agents find issues,
triage promotes them, **dispatch fixes them**."* [MEASURED] `build-pipeline-trigger` is
**enabled, 60s interval, last fired 13:45:12Z**, claiming any row at `status='triaged' AND
pipeline='build'`. The sweep's `pre_query` is `ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1`,
so *which* live site receives edits is not controllable from the switch.

### What was actually run

`visual-design-auditor` dispatched directly at gamesdesign.co.uk, payload-in-COMMAND form with
the `PUBLISH_OK` receipt (the heredoc form silently drops ~4 of 5). Command in RUNBOOK §9.
`corr=5b8c3944-920b-4525-87df-355095f6eb55`, orchestration **COMPLETED**.

Safety established BEFORE firing, not after:
- `write_audit_findings_action.go:787` inserts with a **hardcoded `'detected'` literal**.
- `build-pipeline-trigger`'s predicate is `status='triaged'` → `detected` rows are inert.
- `design-audit-agent`'s live config has **no triage step**. `081b_…robot_hands.sh`'s comment
  *"step 4: triage_detected_items → promotes detected → triaged"* is **STALE**; triage now
  lives in `improvement-loop` as `triage_findings`.
- Binary probe on `v1.0.1302` (both replicas rolled 11:28Z, after the handoff's 1301): three
  needles, single pass — `re-audited this site on` **present**, control
  `NO_CHANGE_GATE_UNREADABLE_RESULT` **present**, nonsense needle **absent**. The
  `build provenance` line was already out of `--tail=3000` on ~2h-old pods, as warned.

### The result — first live execution of WII-018

```json
"retraction": { "dark_section_audit": {
    "silent": false, "candidates": 2, "streaks_bumped": 0,
    "streaks_reset": 0, "retracted": 0, "retracted_parked": 0, "skipped_in_flight": 0 }},
"audit_source": "visual-design-audit", "total_findings": 5, "items_created": 4
```

- **The safety arm is proven on production traffic.** The producer spoke
  (`classification_stats` carried `dark_section_audit: 1`), so `silent=false` and nothing
  retracted. This is the outcome the handoff insists is CORRECT, not a failure.
- **The `failed` row was not written at all.** `70416d75-…` still carries `retraction` null and
  `updated_at = 2026-08-14 15:05:31.904145+00`, unchanged. That proves the `toReset` branch's
  *"only write when there is something to clear"* guard **live** — a spurious write there would
  have bumped `updated_at`, the exact hazard the file's header is built around.

### `candidates: 2` — retraction runs AFTER the insert, so a run sees its own new row

I predicted 1. The run filed a NEW `dark_section_audit` row (`9ae28ee3-…`, `detected`) whose
`item_key` is **byte-identical** to the existing `failed` row's:
`visual-design-audit_dark_section_audit_index_e33263f4-74f8-494f-b191-546845dbbddf`.
Dedup did not suppress it because `failed` is terminal for `idx_swi_dedup`. Both rows are then
non-closed and in-scope, so both count.

**Self-consistent and bounded at 2**: a run that files a dark-section finding is by construction
not silent about dark sections, so it can never bump the streak on the row it just created; and
a second `detected` duplicate cannot be filed while the first is open. Recorded because
`candidates` will read 2 rather than 1 per affected site, and that is **not** a defect.

### What is still NOT proven live, stated plainly

- ~~**The bump arm** (streak climbing) — needs a run that is silent about a site holding an open
  row. All four `failed` sites are genuinely unrepaired, so no live run can be silent yet.~~
  > **CORRECTED 2026-08-15, ~50 minutes later — THE BUMP ARM FIRED, and this claim was FALSE
  > when I wrote it.** See the batch-2 section below. What caught it: the owner overruling my
  > recommendation to stop at one site. I had recommended against running the other three on
  > the grounds that they would "produce the identical result", and the very next site produced
  > the single most valuable result of the day. **The claim inherited "they are genuinely
  > unrepaired" from the handoff without re-checking it** — and it had gone stale that same
  > afternoon, when another lane re-rendered one of the four sites at 13:44–13:53Z.
- **The retraction arm** (N=3 → close) — needs three consecutive such runs.
- **The `audit_source` scope guard is still indistinguishable from the closed-status guard on
  live data.** [MEASURED] every `design-audit`-sourced row fleet-wide is `complete`, so it is
  excluded by status before source is ever consulted. It is structural and unit-tested; it has
  still never had to do work in production.

### Side effects, stated

4 new `detected` items on gamesdesign.co.uk (batch `a53c7b41-…`): `dark_section_audit`,
`needs_design_review`, `responsive_fix`, `spacing_fix`. All inert while nothing triages. Fleet
`dark_section_audit` population moves 19 complete / 4 failed → **19 / 4 / 1 detected**. One
visual LLM audit's spend. **No live site content changed.**

---

## 2026-08-15 batch 2 — the other three sites, and THE BUMP ARM FIRED

Owner chose to run the remaining three despite my recommendation to stop at one. That was the
right call and my recommendation was wrong; the reasoning is recorded in `WRONG_CALLS.md`.

| site | corr | silent | candidates | bumped | created |
|---|---|---|---|---|---|
| mortgagecalculator.co.uk | `a62e2cd7` | **true** | 1 | **1** | 0 |
| oufe.com | `81538a11` | false | 2 | 0 | 4 |
| webdesign.uk | `d21a326c` | false | 2 | 0 | 5 |

### The silent run, and what it wrote

`mortgagecalculator.co.uk` returned 5 findings, `classification_stats` =
`{spacing_fix: 1, responsive_fix: 1, needs_design_review: 3}` — **no `dark_section_audit`**. So
`shapeRecognised` was true (it parsed a full findings list) and `observedItemTypes` lacked the
gated type. That is the honest definition of silence, not a failure to read the instrument.

The row `6fe8a0fc-…` now carries, and remains `failed`:
```json
{"silent_runs": 1, "audit_source": "visual-design-audit",
 "last_silent_at": "2026-08-15 14:05:37.321502+00"}
```

**`candidates: 1` here vs `2` on the three noisy sites is the same rule seen from both sides**,
and it confirms the batch-1 explanation exactly: a run that files a dark-section row counts its
own new row (2); a silent run files nothing, so only the pre-existing `failed` row is a
candidate (1). Nothing else needed to change for that to come out right.

### The silence is CORROBORATED AT THE ARTEFACT, not taken on the LLM's word

A single silence is weak by design (N=3 exists because the per-run miss rate is bounded at only
~35%), so the interesting question is whether this one was a true repair or a miss. It is
neither of the obvious answers. **The ticket's premise does not hold on the served page.**

Ticket `6fe8a0fc-…` (filed 2026-08-13) reads: *"CTA section uses
`var(--color-cta-bg, var(--color-primary))` which resolves to the gold accent (#b59230) as its
background … if `--color-cta-ink` is undefined the text will inherit the body dark colour
(#334155) … creating very poor contrast."*

[MEASURED 2026-08-15 14:2xZ, `curl` of the live site + its stylesheet]
`https://mortgagecalculator.co.uk/assets/css/styles.css` declares:
```css
--color-cta-bg:   #e9e2d3;   /* light cream */
--color-cta-text: #1a1a1a;   /* near-black */
--color-primary:  #b59230;   /* the gold the ticket feared */
```
`--color-cta-bg` **is defined**, so the `var(…, var(--color-primary))` fallback is
**unreachable** — the CTA is near-black on cream, contrast ≈ **13.5:1**. `--color-cta-ink` does
not appear anywhere in the HTML or the CSS (0 occurrences), so the whole conditional the ticket
is built on never evaluates.

This is textbook `[[a-css-fallback-is-present-and-inoperative]]`: the auditor read the *source*,
reasoned about what would happen *if* the fallback fired, and filed a ticket on a branch the
served page never takes. **So the silence is correct, and half two's first live streak is
accumulating against a ticket that arguably should never have been open.** That is the mechanism
doing exactly its job — an honest exit for an item the handler could never have "repaired",
because there was nothing to repair.

⚠ **What I could NOT establish:** whether `--color-cta-bg` was already declared on 08-13 (making
the original ticket a false positive from the start) or was added by the five `page_rerender`
items another lane completed at 13:44–13:53Z, ~12 minutes before this audit. I have no
pre-render copy of the CSS. Both stories end in the same correct behaviour here, but they are
different facts and I am not asserting either.

### ⚠ A REAL GAP IN THE N=3 RULE, found by nearly exercising it

`ConsecutiveSilences: 3` counts **runs, not independent observations, and nothing enforces a
minimum interval between them.** Three manual dispatches inside one minute would satisfy it and
retract the ticket. The measurement that licensed N=3 was *"7 of 7 post-closure re-visits across
4 sites"* — re-visits spread over **days**, against page states that had had time to change.
Three reads of an unchanged page minutes apart are heavily correlated, so they do not carry the
independence the ~4.2% figure assumes.

On the natural cadence this never bites (the sweep round-robins the estate, so a given site is
re-audited rarely). It bites the moment anyone does what I just did — drive the streak by hand.
**Do not walk a ticket to retraction with back-to-back manual runs and then cite the 4.2%.**
Candidate fix if this ever matters: gate the bump on `last_silent_at` being older than some
interval, so a burst counts once. Not filed as a bug — the mechanism is correct on its intended
driver, and this is a hazard of the manual lever §9 introduces.

> **UPDATE, run 2 below: this worry was real but I had its DIRECTION wrong.** I feared two rapid
> runs would be *too correlated* to count as independent evidence. Run 2, forty minutes later on
> a byte-identical page, returned the **opposite verdict**. The observations are far noisier than
> I assumed, not more correlated. The interval gate above would have made things *worse* — it
> would have suppressed the very re-observation that caught the miss.

---

## 2026-08-15 batch 3 — RUN 2 REVERSED IT, and this is the best result of the day

Owner approved driving mortgagecalculator to N=3. Run 2 (`277b8644`, 14:45:05Z) returned:

```json
{"silent": false, "candidates": 2, "streaks_reset": 1, "streaks_bumped": 0, "retracted": 0}
classification_stats: {spacing_fix: 1, responsive_fix: 1, dark_section_audit: 1, needs_design_review: 2}
```

**The streak was RESET from 1 to 0.** The row `6fe8a0fc-…` is `failed` again with no `retraction`
key at all. **I stopped here and did not fire run 3** — see the last section.

### Same page. Same bytes. Forty minutes. Opposite answers.

This is the first **direct live observation of the auditor's per-run miss rate**, and every
alternative explanation was excluded before claiming it:

- **The page did not change.** [MEASURED] `curl` of the homepage after run 1 and again after run
  2: both 36,807 bytes, `cmp -s` **identical**. Not "looks the same" — byte-identical.
- **Nothing repaired it in between.** [MEASURED] the only `site_work_items` rows on that site
  with `updated_at` between 14:00 and 14:46 are the two run 2 itself wrote at 14:45:05, plus
  `page_rerender` rows another lane **triaged at 14:45:55** — *after* run 2 had already read the
  page.
- **The defect run 2 found was present during run 1.** `--hero-btn-ink` appears **3 times** in
  the page I fetched *before* run 2 ran.

So run 1 did not observe a repaired site; **run 1 missed a defect that was there.**

### Why this vindicates N=3 rather than embarrassing it

N=3's justification was *"7 of 7 post-closure re-visits re-reported a known-unrepaired defect"* —
a sample that bounded the per-run miss rate at ~35% while **observing zero misses**. Today we
observed one directly, on the second same-page pair anyone has run.

**Had `ConsecutiveSilences` been 1, run 1 would have retracted the CTA ticket at 14:05, and run 2
would have filed a fresh dark-section finding on the same site forty minutes later.** The
safeguard did precisely the job it was designed for, and the reset arm — `streaks_reset: 1`,
never previously executed in production — is what did it.

**Three arms of WII-018 are now proven live in one afternoon:** the refusal arm (batch 1), the
bump arm (batch 2), and the reset arm (batch 3). Only the retraction arm remains unexecuted, and
it should stay that way until it happens honestly.

### ⚠ CORRECTION — my "the silence is CORROBORATED at the artefact" claim was TOO BROAD

Batch 2 above says the CSS check corroborated the silence. **It did not, and the error is worth
naming precisely because it looks like diligence.**

What I verified: the *CTA ticket's* premise is void — `--color-cta-bg` is declared `#e9e2d3`, so
the `var(…, var(--color-primary))` fallback is unreachable and the CTA is ~13.5:1. That part
stands and is still true.

What silence actually asserts: **that this producer saw no dark-section defect anywhere on the
SITE.** The rule is site-scoped by deliberate design — the file's own header says *"Any
dark-section finding, under any spelling, on any page, keeps every one of that site's tickets
alive."* So I checked one ticket's claim and treated it as evidence for a **site-wide** one.

Run 2 found a dark-section finding in the **hero**, not the CTA, and reset the streak on the CTA
ticket. That is the site-scoping working exactly as documented — and it is also the exact hole in
my verification. **The narrower check felt like strong evidence and was addressing a different
proposition from the one the mechanism makes.** Logged in `WRONG_CALLS.md`.

⚠ Not resolved, and it does not change any of the above: whether run 2's hero finding is a true
contrast defect or a hardcoding complaint (`--hero-btn-ink: #0F1115` is *declared*, and used as
`color: var(--hero-btn-ink, var(--color-primary))`; I did not establish the button's background).
Either way it is a `dark_section_audit` classification and correctly resets the streak.

### Why I stopped at run 2 instead of firing run 3

The owner's approval was for "two more runs to close it", on the shared premise that the site was
silent. **Run 2 falsified that premise.** With the streak back at 0, reaching N=3 now means
firing until three silences happen to line up — which, at an observed miss rate of roughly one in
two on this page, is running the experiment until it produces the answer I already said I wanted.
That is not a proof of the retraction arm; it is selection. The retraction arm stays unproven,
and that is the honest state.

---

## 2026-08-15 follow-up — WHO ELSE reads an ABSENT finding as evidence? (answer: nobody at risk)

The miss rate above only matters if something else in the estate depends on it. Audited that
directly rather than assuming. **The answer is reassuring, which is itself the useful result:
half two is the only consumer exposed, and it is the one already protected.**

**1. WII-016, the render audit's retraction — SAFE, and for a structural reason.**
[VERIFIED at source, `write_render_audit_findings_action.go:540-595`] its `stillFailing` set is
built from `payload.Contrast`, i.e. **pairings the render-audit adapter actually measured**, not
from an LLM's report. It resolves per pairing, scopes to pages in `audited` (anything else →
`retractionOutOfScope`), and its availability argument is the real expression
`len(payload.Summary.PagesAudited) > 0`. So it retracts on **N=1 and is right to**: half two's own
header states the rule — *"Never 1 unless the observation is a MEASUREMENT rather than an LLM's
report."* WII-016 is on the correct side of that line and does not inherit today's finding.

**2. Half two is the only absence-keyed LLM consumer**, and N=3 is what covers it — now vindicated
by direct observation rather than by the 7-of-7 sample.

**3. ⚠ CORRECTION — this lane's "two-strike rule suppresses re-detection" claim is NOT confirmed
as stated.** `SUMMARY_2026-08-15` says *"a two-strike rule actively suppresses re-detection of
faults that keep coming back."* I repeated it to the owner today before checking it. What the code
actually does:
- [VERIFIED] exhausting attempts sets **`failed`**, not blocked
  (`complete_work_item_verification.go:385`, `WHEN attempt_count + 1 >= max_attempts THEN 'failed'`).
- [VERIFIED, and watched live today] **`failed` does NOT suppress re-detection** — `failed` is
  terminal for `idx_swi_dedup`, which is exactly why batch 1 filed a fresh `detected` row beside
  the `failed` one under a byte-identical `item_key`.
- The real suppressor is a **different mechanism with a different trigger**: the blocked filter in
  `write_audit_findings_action.go:793`, which fires on `status='blocked'` — and blocking comes
  from an unroutable handler (`claim_work_item_action.go:162`), not from repeated failure.

Not corrected in the summary itself (summaries are never overwritten); recorded here, which is
where corrections to this lane's claims live.

**4. Out of scope for this lane, and handed on rather than worked.** The blocked filter turns out
to be blind to both category and producer — `capability_gap` always carries `PageID: nil`, so the
filter's `($3::uuid IS NULL OR …)` clause always collapses to "any blocked row of this item_type
on this site", and all 18 live blocks were filed by the discovery/remit path rather than by
`write_audit_findings`. Armed on 14 of ~22 sites; **whether it has ever fired is unknown**, and
the counter that would say so has existed for one day across 9 runs. `who-owns.py 279` says that
territory is owned and active, so it went into `bugs_open/279` as a marked contribution
(`1aa53be77`), explicitly not as a competing bug and not as work routed at them.

**Consequence for the expensive follow-up:** re-measuring the auditor's miss rate with more trials
is what `write_audit_findings_retraction.go`'s own header nominates as the way to move N — but
since nothing outside half two depends on that number, it is **not urgent**, and the natural
cadence supplies the trials for free once a carrier is back on. Deliberately not spent today.

### Side effects, batch 2

9 new `detected` items across oufe.com (4) and webdesign.uk (5); **0 on
mortgagecalculator.co.uk**, where all 5 findings deduped onto existing rows. Fleet
`dark_section_audit` population: **19 complete / 4 failed / 3 detected**. The four `failed` rows
are all still `failed` — none retracted, which remains the correct outcome for the three noisy
sites and is correct-so-far (1 of 3) for the silent one. Three visual LLM audits' spend. **No
live site content changed by any of this.**
