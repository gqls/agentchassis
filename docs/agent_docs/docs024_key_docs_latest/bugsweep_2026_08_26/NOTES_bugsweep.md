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
