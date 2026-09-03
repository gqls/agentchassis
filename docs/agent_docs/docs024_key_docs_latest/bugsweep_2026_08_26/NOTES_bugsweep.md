# NOTES — bug sweep lane (append-only, newest at the BOTTOM)

Technical log: what was tried, what the system actually said, and every misstep.
Plain-English history is `README_where_we_are.md`; current state is the newest HANDOFF.

---

## 2026-09-02 — `bugs_open/338`, the voice gate's density rules on a single value

Picked up from `HANDOFF_2026-09-02_continue_here.md` §6 ("next bug: 338, still unowned")
while the 404 round-4 council verdict was outstanding.

### Ownership, checked before touching anything
`scripts/who-owns.py 338` prints **OWNED**, exactly as the handoff predicted — and the
commits driving that verdict are the 08-26/09-02 handoffs' own. No owning workstream
directory. `git log --since=2026-08-20` on both fix sites: **no commits**. Tree check
(who-owns is blind to uncommitted sessions): both target files WERE dirty, which looked
like a session mid-fix and was not — `git diff` showed a Go 1.19 **gofmt doc-comment
sweep** (blank-line removal; `''` → `”` typographic quotes in a comment), part of a
15-file formatting pass across `platform/orchestration/actions/`. Not semantic WIP.
Declared as a same-file passenger in the commit message rather than reverted.

### Re-measuring before writing code
- Blank `meta_description` census: `leopardessconsulting.co.uk` and `oufe.com` have
  **1** blank active page each — still exactly the two sites §3 predicts.
- Voice-gate census re-run: still **9** enabled gates, still **7** with the length checks
  disabled. **The bug file's §3 table omits a column**: those seven also set
  `em_dash_per_1000_words: 100000`. That mattered (below).
- ⚠ The blank leopardess page is now `case-study-automated-intelligence-pipeline`, **not**
  the page §2 quotes. The failure RECURS against new pages; it is not one stuck row.
- `orchestration_states` returns nothing for these — aged out, the same trap §3 of the
  handoff records for the 404 verdict. The refusal text is only readable as quoted history.

### MISSTEP 1 — I nearly implemented the bug file's own §4 remedy, which was wrong
§4 says to drop `em_dash_density` "as a density rule" and hand-roll a flat "contains an em
dash" test. I started to, then did the arithmetic: the rule is `emDashes/totalWords*1000`,
a rate over **words**, so one em dash in 20 words scores **50.0** against a default trip
of 3, and a single em dash trips the default at any length below **333 words**. It already
means "contains an em dash" for every single-value field. **And the flat replacement would
have been worse than redundant** — it would have ignored site config and re-gated the seven
sites that had switched the rule off, which is precisely the "relaxing/tightening a checker
to fit the content" error §4 itself warns about two paragraphs earlier.

So the axis is not content-vs-density. It is **what the signal is a rate OF**: a rate over
words reduces correctly at n=1; a count per page (`triad_density` 4, `negation_density` 12)
or a share over sentences (`long_sentences`) does not. Corrected in the bug file and in the
LANDMINES entry, both visibly rather than by editing away.

### MISSTEP 2 — the bug file's check list was already stale, and so was the landmine's
§4 enumerates the check names and omits **`negation_density`**, added by `bugs_open/305`.
Correct when written, wrong by birthday. The landmine's remedy ("filter to
`check == "banned_phrase"`") was narrower still — it would have dropped `strawman`,
`flourish_ending` **and** the em-dash rate, three working rules.
**Consequence for the design:** the fix does not hand-keep a list. `voiceCheckKinds` is
exhaustive over all 8 emitted names and `TestEveryVoiceCheckIsClassified` reads the
`Check:` emission sites straight out of `voicetells.go`, failing on any unclassified check
or any stale map entry. The map's literals are `"name": VoiceCheck…` and the emissions are
`Check: "name"`, so the test cannot match the map and vouch for its own completeness.

### MISSTEP 3 — I wrote "ENUMERATED rather than asserted" about a grep I had not run
In the council rationale, arguing this was not architecture-scope. Running it corrected me
twice in four items: a **doc comment** at `save_page_meta_description_action.go:55` counted
as a call site, and **`cmd/voicescan/main.go:103` missed entirely**. Logged in
`WRONG_CALLS.md`; the stale comment was repointed in the same commit.

### Verification actually run
Both arms of §6 induced. Every corpus-only case asserts the check fires on the PAGE path
first (`CONTROL FAILED` fatal), so a drop cannot pass because the input never tripped it.
Gate fixture is the **live** leopardess `site_specs` row, not composed — a fixture invented
to suit the fix exercises its own rule.

Four mutations, each killing exactly the expected tests, green on restore:

| mutation | killed |
|---|---|
| drop `negation_density` from the map | `TestEveryVoiceCheckIsClassified` + `CorpusOnly/negation_density` |
| emit `Check: "brand_new_check"` unclassified | `TestEveryVoiceCheckIsClassified` + `CorpusOnly/triad_density` |
| remove the filter (return `ScanVoice` raw) | `GoodLongDescription` + all four `CorpusOnly` subtests |
| drop everything (gate removed) | `BannedPhraseStillRefused` + `EmDashRuleTravels` |

Committed `425398a01`, 3 files, one area. `scripts/verify-head-builds.sh` → **OK at HEAD**.
Council **SUBMITTED** `106802fc-ad14-4beb-b622-147c3a0ab982` (admission dry-run first, free;
`operation` must be `modify|add|remove|config_change` — `create` is rejected). Registered as
**CQ-035**.

### Adjacent reds found, neither mine, neither touched
- `TestNoHandSpelledTombstonePredicate` **fails at committed HEAD** on
  `check_unrendered_page_imagery.go:156/197/202/207` (`a87746b77`, the 114/IMG-077 lane).
  Verified via `git show HEAD:` — same four lines, file clean in the tree. This is a
  SECOND instance of the handoff §5 class (`platform/livespec` red at HEAD).
- The pre-commit hook reported **"optional-key parity: NOT CHECKED (the tree does not
  build)"** — that is the working TREE with every session's WIP, not HEAD. HEAD builds.
- `_RELOCK` still an unclassified migration suffix (handoff §5, unchanged).

### Owed
Read the council verdict. Then, **after the next chassis roll**, the artefact check: the two
blank pages fill. 338 stays OPEN until then — a Go change is inert until an image rolls.

### Council verdict — APPROVED, and it still found a real hole
`106802fc-ad14-4beb-b622-147c3a0ab982`, 18:25Z. 9 approve, 2 advisory objections, none
high-severity. `architecture` recorded `ARCHITECTURE_SIGNAL: point_fix | DEFLECTIONS: 0`
and confirmed the RFC_022 shape. Several seats specifically credited the grep-corrected
call-site enumeration — the misstep-3 correction was read as discipline, not as noise,
which is an argument for recording missteps in the submission rather than tidying them.

**`bug_historian` [medium] was RIGHT and is now `bugs_open/442`.** Its ask: confirm a
failure-surfacing mechanism exists downstream of `metaDescriptionFailsCopyGates`'s
non-empty return, or add one. Confirmed **absent** — the action returns a nil error, the
live workflow has no conditional on `save_result` anywhere, `continue_on_error` is moot
because nothing errors. My 338 fix narrows how often the gate fires; it does nothing about
a true refusal being invisible.

⚠ **Investigating it made it sharper than the objection knew.** The workflow's
`result_message` — the only surface telling a person how to read the outcome — names four
refusal reasons and omits all three copy-gate ones (`voice_tell`, `banned_claim`,
`voice_gate_unreadable`), i.e. exactly the ones that caused 338. **Third stale-by-addition
enumeration in one task.**

`editquality` [medium] worried the partial diff orphaned `blocks` and would not compile.
Answered by fact, not argument: `blocks` still goes to `checkBannedClaims`, and
`verify-head-builds.sh` says OK at HEAD. A fair objection from a partial diff.

### MISSTEP 4 — I cited a grep in bug 442 that does not return the number beside it
Wrote *"the action returns **7** distinct reasons (`grep -n '"reason":' …`)"*. The number is
right; **the cited command returns 4.** The four are literals in the result map; the three
that matter are returned as bare strings from the gate helper. So the citation would have
led the next reader to the very number the bug file exists to correct, and confirmed it.
Caught by running it before committing — ~90 minutes after logging misstep 3, which is the
same fault. Second row in `WRONG_CALLS.md`; the pair's shared check is **paste the output,
not the command's name**. ⚠ And the asymmetry was itself the finding: the obvious single
grep finds exactly the four already documented and reports the list as complete.

### Another session swept my LANDMINES edit into their commit
`381529d5a` (19:22:48) landed 86 seconds before mine (19:24:14) and carried my LANDMINES
correction with it, so my own docs commit shows 4 files, not the 5 I named. **Nothing
lost** — all three markers verified present in HEAD, forward-only holds. Recorded because
CLAUDE.md warns about exactly this and it is easy to misread as a failed edit: the file was
clean and `git show HEAD:` found my text, while `git log -- LANDMINES.md` named someone
else's commit.

### CQ-035's index row shipped in a SECOND commit
The entry named the category file by pathspec and not `000_concept_index.md`, so it landed
entry-without-row — the exact case the index header already narrates twice (WFA-022,
TL-049) and the `pattern-check` advisory caught it at commit time. Row added in
`6413ad1f4`: 2,052 → 2,053, CQ-035 confirmed unclaimed first, present as both row and entry.
⚠ `WII-035` is a **duplicate row id** (two near-identical rows, lines 415/416) — verified
present at HEAD before my edit, so another lane's; left unfixed rather than picking which
description survives.

### 404's round-4 verdict landed while this ran — APPROVED
`f2e4ac2a-…`, 16:33:30Z: *"approved with 3 advisory objection(s) — none high-severity"*
(`editquality`, `bug_historian`, `debug_historian` objecting, all advisory). That discharges
the handoff §6 item 1. The r3/r4 commits already carry `Council-Submitted:`, so 098 credits
them at report time with no amend. Reading those three objections is the 404 lane's, not
mine — but note this round and mine BOTH approved while finding real defects.

---

## 2026-09-03 — the roll landed, and 338's acceptance test turned out to be unfalsifiable

`agent-chassis` on `v1.0.1356`, pods started 08:58:07Z.

### Provenance was unavailable by BOTH sanctioned routes, each for a recorded reason
1. **Log stamp rotated inside 18 minutes.** LANDMINES' own precheck: at 09:16Z the two pods'
   first log lines were 09:16Z and 09:10Z, so the 08:58 startup line was already gone. The
   entry's time-limited case, not "unstamped".
2. ⚠ **`grep 'build provenance'` on this service matches LANDMINE TEXT about build
   provenance** — the chassis logs whole council/diagnosis payloads, so the recipe greps up
   its own documentation. My first attempt returned 1.9MB. Already a landmine; live today.
3. **The release was applied from an UNCOMMITTED overlay bump.** `git show HEAD:` on the
   chassis overlay → `v1.0.1353`; tree and pods → `v1.0.1356`. The build point is not in git
   history, so no ancestry check exists. `make release` bumps makefile + overlays in the
   working tree; nobody committed it.

### MISSTEP 5 — my must-be-absent control passed because the command ERRORED
`$C` (deploy commit) was empty, so both `git merge-base --is-ancestor` calls printed usage
errors and exited non-zero — and my `&& / ||` turned each failure into its designed
"negative" branch. Output read `ANCESTOR: no` + `CONTROL OK`. **A control keyed on a
NON-ZERO EXIT cannot discriminate, because every failure mode of the command is non-zero.**
Full row in `WRONG_CALLS.md`. I was one paste from writing "my fix did not ship" into a
handoff.

### THE FINDING THAT MATTERS — §6 can never come out either way
338 §6 says: watch the two blank pages fill. They cannot. `load_pages_missing_meta` gates on
`page_visible_text_len(p.id) > 200`, and they measure **0** (with **0** eligible components)
and **124**. The backfiller never selects them → the gate is never consulted → the fix
cannot move them. A session running §6 after the roll would see "still blank" and conclude
the fix failed.

Fleet-wide `[MEASURED 2026-09-03]`, **with a demand control**: all **37** remaining blank
active pages (11 sites) fail the gate, **zero** selectable, averaging **8** chars of visible
text. Control — the **1,164** pages that DO have a description average **4,401** chars with
**1,137 (97.7%)** clearing it. So the instrument is sound; every remaining blank is a
near-empty page.

**Why §6 was right when written and is wrong now:** on 08-20 the blanks WERE gate refusals.
Migration `501`'s ≤20-word instruction has since filled that population. What is left is a
different population, blank for an unrelated reason. **A measurement can go stale by having
its subject REPLACED, not just by drifting** — the population was renewed under a
predicate that no longer selects for the cause.

⚠ **And `501` makes the fix largely inert for this caller anyway** — ≤20 words clears the
trip of 22, so the natural exercise will not occur. The fix's value is the seam and the
lower-threshold case, both real, neither observable today.

### Knock-on: 320's headline is 3.1%, not 55.7%, and its residual changed KIND
Updated in place with today's dated figures. What is left there is a **coverage floor**, not
a writing failure — and whether a near-empty page should carry a description at all is an
owner question, not a backlog.

---

## 2026-09-03 (later) — `bugs_open/442`, the unowned item the handoff left: candidate 3 shipped, and two things the bug file had wrong or missing

Picked up 442 because §5 of `HANDOFF_2026-09-03_continue_here.md` names it as the lane's only
unowned leftover. `scripts/who-owns.py 442` returns this lane, so there was nobody to collide with.

### What shipped
Migration `728_meta_description_backfill_result_message_names_the_copy_gates.sql` (+ `_ROLLBACK`),
commit `5a8728db9`, **applied and recorded**, council `2ed33c57-b49a-4b1b-ad1e-7e23ce6c477a`
submitted (`Council-Submitted:`, verdict owed a read). Config-only: no image, no roll.

The old `result_message` told a reader that a refusal *"carries a named reason (empty_candidate /
candidate_looks_internal / candidate_too_long / already_has_description)"*. Seven exist. The three
missing are the copy-gate ones. Verified at the live row afterwards, independently of the
migration's own verify block, with a must-be-absent control (`THIS_REASON_DOES_NOT_EXIST` → `f`).

I did **not** just extend the list to seven. §4 of the bug file points out that these lists rot by
ADDITION and that the 2026-08-22 date-your-counts ruling has no equivalent for enumerations — so a
seven-item list is the same defect one birthday later. The new message names the seven, splits them
by what they ask of a reader, and then says it is a copy, where the authoritative set lives, and
that finding it takes two greps.

### MISSTEP 6 — a mutation that "should have failed" passed, and my first reading was wrong
Rehearsing 728 I mutated the `UPDATE` to also write `task_workflow`, expected the positive control
to abort, and it committed. I briefly took that as the control being broken. It is not: the control
compares `default_config`, which is the only column the `UPDATE` writes. The mutation was outside
what the control claims. Corrected mutation (a second `jsonb_set` **inside** `default_config`)
aborted at once. The trap worth naming is the next move I nearly made — widening the control until
the out-of-scope mutation fails, which would have the migration assert things its own `UPDATE`
cannot cause. `WRONG_CALLS.md` row written.

### MISSTEP 7 — and this one was in the bug file, in my own hand, for a day
442 §2 said `orchestration_states` *"returns zero rows carrying a `save_result.reason` fleet-wide —
the rows age out, so even the field that does exist is not readable after the fact"*, and §6 turned
that into *"do not verify at `orchestration_states`"*.

The zero is real. The reason is invented. `[MEASURED 2026-09-03]` the table holds 9,277 rows over a
**~26-hour** window (oldest `2026-09-02 09:41`); five backfiller runs survive in it; **all five
carry a `save_result`** — 5/5 `updated:true`, **0/5** carrying a `reason`. The action only writes
`reason` when it refuses, so a window with no refusal returns zero either way. **Two sufficient
causes, one measured.** The predicate NAMES the event, so it cannot separate "no record" from "no
event"; the demand control is one column — is `save_result` present at all.

So §6 was telling the next session not to look in the only place the evidence is. Corrected in
§9b; landmine written, footprinted on `orchestration_states`/`save_result`/`__step_error`, because
there is no tell — both answers are `(0 rows)` and the retention story is the more interesting one,
so it is the one that gets written down.

### The finding that resizes candidate 1 — there are TWO silent paths, not one
The writer step's own prompt says, live and verbatim: *"omit that page entirely rather than
inventing one. Returning fewer entries than you were given is a correct answer."* And nothing
compares `pages_missing_meta.count` with `jsonb_array_length(written.result.descriptions)` —
`check_has_pages` reads one, `backfill_loop` reads the other, `complete` prints a message.

A page the model drops therefore leaves **less** trace than a gate refusal: no `save_result` at
all, because the loop never reaches the action for it. Candidate 1 files from **inside** that
action, so it structurally cannot see this path. Recorded as a fifth candidate (compare the two
counts) which covers both and files nothing into a queue.

⚠ And I cannot say whether it fires: all five surviving runs were `offered 1 / written 1`, and five
single-page runs rule out **no** omission rate below roughly 45%. `0 of 5` is not evidence.

This is 016b §9's 2026-08-24 entry (*"a loop over the model's ANSWER cannot account for what the
answer OMITS"*) with a **second cause** — the prompt instructing the omission rather than a
`max_tokens` ceiling causing it — so that entry's `max_tokens` census comes back clean while the
class still applies. Extended in place rather than filed again (`652f32f74`).

### The volume objection to candidate 1, measured, and it holds
`voice_tells` work items, counting the archive as well as the live table: **66** in
`needs_human_review` against **5** ever complete (3 live + 2 archived), nothing filed since
2026-08-27. Filing gate refusals there relocates the silence. That is the open owner question in
handoff §0.4 and I have not pre-empted it.

### Damage today is ZERO, demand-controlled
Active pages: **37** blank (avg 8 chars visible text), **0** of them clearing the backfiller's
`page_visible_text_len > 200` gate; **1,171** described (avg 4,381), **1,137** clearing it. So no
page is currently both eligible and blank — both silent paths are latent, not costing anything.
That is an argument about priority, not about whether the mechanism is broken.

---

## 2026-09-03 (later still) — the owner ruled "make them loud", and the design turned on one query

### The question I had to answer before building
§5 candidate 1 says "file a work item"; §9c said the obvious queue is a graveyard (66 parked, 5
ever closed). The owner ruled build it anyway. So the real question was not *whether* but *where*,
and it was answerable in one query: **is the graveyard about humans, or about the shape of the
row?**

`[MEASURED 2026-09-03, site_work_items UNION site_work_items_archive]` items **WITH** a
`handler_agent`: 56,315, **83%** complete. Items with **NO** handler: 6,699, **17%** complete, 989
parked. And `voice_tells` is **69 rows, every one `handler_agent = ''`**.

**It is the shape of the row.** Filing at `needs_human_review` with no handler would have looked
like a fix and been one more row in the 17%. I had been about to treat "the queue is a graveyard"
as a fact about people; it is a fact about whether anything is pointed at the row.

### What got built
Commit `776511e70`. Migration `734` seeds `meta-description-repair` (live now); the Go files
`meta_description_refused` at it (inert until the roll). The repair agent re-asks with the refusal
**quoted back** and saves through the **same gated action** — the reason the first attempt failed
is the one thing the hourly backfiller never had. Second refusal parks at `needs_human_review`
with both attempts on the row.

Config was applied BEFORE the Go was committed, deliberately: a refusal filed at a handler that
does not exist is demoted to `deferred` by `writeWorkItem`'s registration probe. Safe, never a
livelock (078), but a parked row nobody asked for.

### MISSTEP 8 — my own negative test was vacuous, and its comment said it could not be
The arm asserting §6's "a clean candidate must file nothing" registered **no** sqlmock
expectations and asserted `ExpectationsWereMet() == nil`. That is nil **unconditionally** —
`ExpectationsWereMet` reports UNFULFILLED expectations, and there were none. Deleting the
classifier's early return, so every reason filed, left the test **green**: the unexpected
`BeginTx` just returned an error to the code, which logged it and returned.

The comment I had written said it "cannot pass by accident".

Fixed by **inverting** the assertion: register `ExpectBegin()` and require it to be UNFULFILLED.
Now red for all six reasons. **Only running the mutation found this** — and the arm I got wrong
was the NEGATIVE one, which is always the easier one to write vacuously. That is §6's own warning
("induce both arms, or the test proves nothing") landing on the person who wrote it.

### MISSTEP 9 — I ran a "mutation" that never executed and would have reported green
To avoid dirtying the tree I staged mutation 5 in the scratchpad and passed
`--with /abs/path/scratch.go` to `verify-head-builds.sh`. It exited **2** — "could not run the
check" — because the overlay wants a repo-relative path. Had I been grepping only for `FAIL` I
would have seen none and recorded the mutation as **not killed** (or worse, skimmed it as fine).
**Exit 2 is not exit 0 and is not exit 1; a mutation harness needs its own exit-code check**, for
the same reason the 09-03 `merge-base` misstep did: a control keyed on the wrong signal cannot
discriminate. Re-run in-tree afterwards: RED, and the file restored byte-identical.

### The tree does not compile, and it is not mine
Another session has an **untracked** `criteria_value_assertions.go` whose `itoa` collides with the
one in `provocation_gate_action_test.go` at HEAD. So `go test ./platform/orchestration/actions/`
fails in this tree and proves nothing about HEAD. Everything above was run through
`scripts/verify-head-builds.sh --test --with …`. Worth remembering that the working tree being red
is not evidence your own change is red — and that the reverse (green tree, red HEAD) is the
dangerous direction the script exists for.

### Stated gaps rather than surprises
No verifier for `meta_description_refused` (five build guards + a live claimed-item-timeout
migration merged with other lanes — a named follow-up). `voice_gate_unreadable` still silent, and
correctly so for a rewrite handler, but it is a residual. §9d's writer-omission path still open.
And nothing has exercised any of it: zero pages are currently both blank and eligible, so there is
no page to refuse — read "no rows" with that demand control or it reads as failure.

---

## 2026-09-03 (post-roll) — it shipped on the SECOND roll, and the council found one more thing

### The Go is live, and the near-miss is the lesson
`v1.0.1359`, pods 13:28:18Z / 13:28:43Z. Verified at the binary on **both** pods with a
present-control (`candidate_looks_internal`) and an absent-control.

⚠ **It did NOT ship in `v1.0.1358`**, which rolled at 12:06Z — *after* both my commits were
already ancestors of HEAD. The build was cut from an earlier HEAD. I had told the owner the
pending build would carry it; it did not, and I only knew because I probed instead of inferring.
**`git merge-base --is-ancestor` answers "is my commit in the source", not "is my code in the
binary".** Two rolls, one negative and one positive, both measured with the same controls — which
is as close to a clean experiment as this ever gets.

⚠ And the prescribed `kubectl logs | grep 'build provenance'` **failed exactly as its own landmine
predicts on this service** — it matched the chassis's logged landmine corpus *about* build
provenance and returned pages of unrelated text. Second confirmation in one day.

### The roll killed the council round, and the tell was fleet-wide
Round 2 was at `review_guidelines` at 12:05:35Z; pods replaced 12:06:47Z. `[MEASURED]` **7 of 11**
in-flight orchestrations idle >15 min in the same window, worst 24m52s — so it was the roll, not
my submission. That distinction is the runbook's own discriminator (one sick run vs a sick fleet)
and it stopped me editing a submission that had nothing wrong with it.

### Round 2 verdict: REVISE, gated by prior_art [HIGH] — and it was a documentation defect
*"Rationale claims the new filing 'uses the shared writeWorkItem door' … but the sketch calls
`insertWorkItem(...)` — a different symbol."* Right that I had given the reviewer no way to connect
them. Wrong that they differ: `insertWorkItem` is a **two-line wrapper** over `writeWorkItem` with
`dropOnConflict`. Every probe runs.

**This is the same fault in its THIRD consecutive round.** r1's prior_art asked for the greps'
output; r2's asked for three symbols; r3's gating objection is the same shape again. Four
`WRONG_CALLS` rows. **The claims were all TRUE every time** — which is precisely why it keeps
happening and why it matters: a reviewer cannot distinguish a verified claim from an unverified
one, so "I checked" is worth nothing unless the output is on the page. I had written that exact
remedy into `WRONG_CALLS.md` myself that morning and then not applied it to the next two
submissions.

### What the round produced besides the objection
- **`bugs_open/464` FILED** — `bug_historian`'s `MISSING` asked, by name, that the four unread
  copy-gate call sites become a numbered bug before the register entry closes as done. Right: my
  audit was a grep intersection and I had said so, but a stated limitation inside a submission is
  invisible six months later. A bug file is not.
- **`reuse_agent` [low], measured:** eight `*ItemKey(` builders in the actions package, all
  package-private, all different signatures. No shared convention to reuse — the same answer as
  the parkers, and I had assumed it rather than checked it.
- **Nine seats approved cleanly**, including `guardian` and `architecture` — the two that objected
  in round 1 — so r1's fixes held.

### Where it stands
Round 3 submitted with every answer as OUTPUT rather than assertion. Config and Go both live.
**Nothing has exercised any of it**: zero pages are both blank and past the `>200` gate, so no page
can be refused. The first `meta_description_refused` row is the acceptance evidence, and until one
exists this is live-and-unproven — which is exactly what `bugs_open/338` taught this lane not to
paper over.
