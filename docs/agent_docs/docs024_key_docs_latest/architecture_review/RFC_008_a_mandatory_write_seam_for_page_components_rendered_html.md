# RFC 008 — a mandatory write seam for `page_components.rendered_html`, and why an advisory lint is the wrong ceiling

**Filed:** 2026-08-02 by the `bugfix_136_sibling_link_repair` lane, discharging the
`bug_historian` seat's explicit instruction in council round 1 of correlation
`0275f9c2-035f-4c9e-8a50-83836dfeffd9` (APPROVED, 5 advisory objections, none high):

> *"Flagging as architecture-level for a human: candidate 3 (mandatory write-seam for
> `page_components.rendered_html`) should be tracked as a real ticket, not left implicit
> in a lint rule's silence."*

**Status:** ~~open. Not urgent — measured traffic on the newly-guarded paths is zero — and
deliberately NOT actioned by the filing lane, because taking it inside a bug patch is
precisely what CLAUDE.md's platform-seam ruling forbids.~~
**ANSWERED 2026-08-22 — OWNER RULED: no mandatory seam. Closed as answered, jointly with
`RFC_042`.** Decision record at the end of this file; the two-writer recommendation became
`bugs_open/362`.

## What shipped, and what it did not close

`bugs_open/136` (section-editor slug) closed two of the LLM-authored writers of
`page_components.rendered_html` by giving them a shared single-component repair seam
(`repairComponentHTMLBeforePersist`, LNK-027), and added
`check_unrepaired_component_write` to `scripts/pattern-check.py` so a new writer
announces itself at the moment of the edit.

That lint is **advisory and diff-scoped**. It fires only on files a commit touches. A
writer that already exists and is never edited again is never examined; a writer added by
a session that ignores the advisory output ships unrepaired with a green build.

## Four seats said the same thing from different angles, which is the signal

| seat | severity | what it said |
|---|---|---|
| `bug_historian` | medium | the identical shape is already closed twice (`bugs_closed/021` "durable write guard covers one path only") and still open once (`bugs_open/093` "one guarded call site, rerender unchecked"); an advisory lint "does not stop the next silent writer from shipping unrepaired `rendered_html` with a green build" |
| `editquality` | low | the guard "fires on changed files, not a repo-wide sweep", so the writer-set drift the bug asked to fix is "only partially delivered for writers that never get touched again" |
| `guardian` | low | the rule "produces fleet-wide noise with no allow-list entry for those writers until someone acts on it" |
| `architecture` | low | "if report pages or section-editor usage starts growing, this should convert to the mandatory candidate-3 gate rather than accumulate more advisory call sites" |

`bug_historian` also recorded what it could not check, and it is the sharpest line in the
round: *"Whether `pattern-check.py`'s findings are enforced in CI (block PR) or merely
surfaced/read by nobody — several landmines in this corpus show 'advisory' checks
routinely go unread."*

## The proposal, and the argument against it that must be answered

**Proposal.** One `persistComponentHTML(ctx, params, target, html, opts)` that every writer
of that column must call, making an unrepaired write *unrepresentable* rather than merely
detected. Ten SQL sites across nine files today.

**The argument against, stated fairly, because it is strong.** `pattern-check.py` is
advisory *by deliberate design*, and its header says why: `.githooks/pre-commit` warns
that "a stray non-zero exit here stops the whole fleet committing", several sessions share
this tree, and a false positive that blocks is a fleet-wide outage. A mandatory Go seam is
a different mechanism from a blocking lint and does not inherit that objection — but it
inherits a worse one: **two of the ten writers must NOT repair.** `adopt_verbatim.go`
stores the crawled document verbatim and a sha256 of it; the colour fixers rewrite colours
in existing markup. A mandatory seam therefore needs an explicit opt-out parameter, and an
opt-out parameter is a lint allow-list wearing a type signature. The gain is that it is
*visible in the call*, not in a Python dict — real, but smaller than it first looks.

## What would settle it, and it is a measurement, not an argument

Both sides are currently arguing from principle. The question that decides it is
empirical: **does the advisory check actually get read?**

- Instrument it: how many commits in the next N days carry a
  `unrepaired-component-write` finding, and how many of those are followed by a fix or an
  allow-list entry rather than nothing?
- The same question generalises past this rule, which is why it is worth doing once
  properly: `pattern-check.py` now hosts several advisory rules and **nobody has ever
  measured whether any of them changes behaviour.** If the answer is "findings are
  ignored", that is an argument for mandatory seams across the board and a much bigger
  finding than this RFC. If the answer is "findings are acted on", the cheap mechanism is
  the right one and this RFC should be closed as answered.

## Recommendation

1. **Do not build the mandatory seam yet.** Measure the advisory channel first — it is the
   load-bearing assumption under this and every other rule in that script.
2. **Do close the two known-unguarded writers** (`create_tool_component_action.go`,
   `deploy_tool_action.go`) on their own merits, in the tool lane that owns those files.
   That is a bug fix, not architecture: 7 of the 35 live unresolved internal hrefs sit in
   tool-shaped slots.
3. **Revisit this RFC if either trigger fires:** the advisory measurement shows findings
   going unread, or `page_type='report'` / section-editor traffic stops being zero.

---

## DECISION — 2026-08-22, OWNER RULED: no mandatory seam; answered jointly with RFC_042

**The ruling.** Do not build `persistComponentHTML`. `rendered_html` takes the **same posture the
owner ruled for its sibling column the same day** (`RFC_042` option (c)): *detect and attribute, do
not refuse*. This RFC is **closed as answered** — not deferred, not parked pending measurement.

**Why the answer is now available when it was not on 2026-08-02**, in the owner's terms: the thing
the seam was for — knowing that every write to this table is accounted for — has largely been built
by other means since this file was written, and built in the detect-not-refuse shape.

1. **`bugs_closed/229` gave the column an archive** (migration `357`, live and proven 2026-08-19):
   every overwrite and delete keeps a pre-image, four writers stamp `rendered_html_digest` in the
   same statement, and hand-patched divergence raises work items automatically. A write that
   destroys artefact-only content is now *visible after the fact*, which is what "mandatory" was
   trying to buy in advance.
2. **`bugs_open/355`'s A1 made writes self-attributing** (shipped 2026-08-22): every write site
   stamps `application_name`, so the archive names the caller rather than a socket. The census the
   seam would have made unnecessary is now simply *possible*.
3. **The sibling column's detector found ZERO losses across the uncarried writers** against a
   demand control that found 72 where losses were known — so the guard's population, on the one
   column where it has been measured, is empty. Building the same guard on this column would be
   sized from inference, which is the failure mode `WRONG_CALLS` 2026-07-25 names.

**And the strongest argument in this file survives the ruling intact**, so it is restated rather
than dropped: **two of the ten writers must NOT repair**, and *"an opt-out parameter is a lint
allow-list wearing a type signature."* The estate now has that allow-list in
`scripts/pattern-check.py`'s `COMPONENT_WRITE_ALLOWED` **with reasons rather than exemptions** —
and, notably, with the two open cases deliberately left OUT of it so the check keeps firing on
them. That is the seam's discipline without the seam's cost.

### What this ruling does NOT license

- **It is not a finding that advisory checks work.** §"What would settle it" asked for a
  measurement — *do `pattern-check.py`'s findings actually get read?* — **and that measurement was
  never taken.** The ruling was made without it, deliberately and with that stated. The informal
  evidence available on the day points the *wrong* way and is recorded here rather than buried:
  this RFC's own recommendation 2 sat undone for twenty days, and the blocker that justified the
  delay (`bugs_open/180`) closed on the day it was filed while the landmine telling readers to wait
  for it stayed unchanged. ~~**The generalised form of that question now lives in `bugs_open/358`
  candidate B2** (no new finding code ships without a reader; a lint tying new writers to the
  seam). Whoever takes B2 answers this RFC's open question too — build it once, for both.~~
  **CORRECTED same day: B2 is built and does NOT answer this** — see the correction under reopen
  trigger 2. The two questions look identical and are not: B2's subject (`error_code` values) is
  *rows in a table*; this RFC's subject (advisory lint findings) is *terminal output that persists
  nowhere*. ~~**The measurement remains untaken and is nobody's work item today**, and its
  prerequisite — a durable record for advisory findings — is named in that correction.~~
  **TAKEN AND ANSWERED the same day** (`scripts/audit-advisory-findings.py`, DBG-076): the findings
  are recomputable from git, so the durable record named as a prerequisite was never needed. Result
  and its breakdown under trigger 2. **This section's claim — that the RFC was ruled without its
  decisive measurement — was true for about six hours and is now historical.**
- **It does not close the class.** A future writer of `rendered_html` inherits nothing but an
  advisory check on files it happens to touch. That is a *stated* residual, not an oversight.

### The reopen triggers — named, so this is a decision and not a shrug

Reopen this RFC (or escalate straight to the seam) if **any** fires:

1. **The archive shows an unrepaired or destructive write in production** — now answerable, because
   A1 names the writer. A single confirmed instance outside the allow-listed set is enough.
2. ~~**`bugs_open/358` B2's measurement shows advisory findings going unread**, i.e. commits carrying
   an `unrepaired-component-write` finding that are followed by neither a fix nor an allow-list
   entry. That was this file's own decisive question.~~
   > **⚠ CORRECTED 2026-08-22, hours after this record was written — THIS TRIGGER WAS UNARMABLE AS
   > DRAFTED, and saying so is the point of a trigger list.** It routed at `bugs_open/358` B2, which
   > has since been **built** (`cmd/config-key-audit --finding-codes`, register DBG-075) — and B2
   > answers a *different* question. B2 is **DB-authoritative**: it asks which `error_code` values in
   > `agent_error_log` have a declared reader, and it can do that because the codes are *rows*.
   > **`pattern-check.py`'s advisory findings are not rows.** They are printed to a terminal at
   > commit time and leave no durable record anywhere, so "was this finding read?" has nothing to
   > query — and no amount of B2 will produce it. Reported by the 358 lane, verified here.
   >
   > ~~**So the trigger, restated honestly: it is NOT ARMED, and arming it has a prerequisite** — give
   > advisory findings a durable record (one row per finding at emit time, the shape every daily
   > check in this estate already uses), then measure fix-or-allow-list follow-through against it.
   > Until someone does that, **this RFC's own decisive question remains unanswerable**, and the
   > ruling stands on triggers 1, 3 and 4, which are armed today.~~
   >
   > ### ✅ ARMED AND MEASURED, 2026-08-22 — and the prerequisite turned out not to be needed
   >
   > **`scripts/audit-advisory-findings.py` (register DBG-076) answers it, and no durable row at
   > emit time was required after all.** The findings are RECOMPUTABLE — `pattern-check.py --commit
   > <sha>` already replays any past commit — so the record is reconstructed from git across the
   > whole of history rather than accumulated forward from a write inside the fleet's commit hook.
   > That reads history instead of waiting for it, and puts nothing in the hook.
   >
   > **THE ANSWER, `[MEASURED 2026-08-22]`, 5,000 commits (14 days), baseline pinned:**
   >
   > | | |
   > |---|---|
   > | findings replayed | 226 |
   > | decidable (check reads state, file still exists) | 37 |
   > | **acted on** (fixed or allow-listed) | **8** |
   > | unacted (condition still true) | 29 |
   >
   > **8/37 = 22% overall — but that headline is the wrong statistic, and the breakdown is the
   > finding.** ONE rule, `logged-model-output`, supplies **25 of the 29** unacted. Remove it and
   > follow-through is **8 of 12 = 67%**. `logged-model-output` alone is **0 of 25**.
   >
   > **So the advisory channel is NOT ignored — it works for most rules and is ignored for one.**
   > That single rule flags logging an LLM response verbatim, a *privacy* concern whose premise is
   > that the prompt contains what a VISITOR wrote; it is firing across `platform/orchestration`,
   > which builds sites and serves no visitor prose. A rule whose premise does not hold in the
   > package it fires on **should** be ignored, so this is evidence about that rule's scoping, not
   > about the estate's discipline. **It needs a ruling (narrow its gate, or accept and allow-list
   > the package with a reason) — it is not evidence for mandatory seams.**
   >
   > **The rule THIS RFC is actually about scores 3 of 3.** `unrepaired-component-write` fired on
   > three writers of `rendered_html` and all three were fixed on 2026-08-22 (`bugs_open/362`) —
   > including one, `create_tool_component_regenerate.go`, that this RFC's own ten-writer census
   > could not have known about because the file was **born 2026-08-19**, seventeen days later.
   >
   > **Verdict on this trigger: it does NOT fire.** The advisory channel changes behaviour on the
   > rule this RFC cared about, at 100% over the measured window. The no-seam ruling stands, now on
   > evidence rather than on the absence of it.
   >
   > ⚠ Two limits, stated because the number is quotable and will be quoted. (a) Only 8 of 21 checks
   > are state-evaluable at all; the other 13 read the diff, so 98 findings are **undecidable** and
   > this measures a slice, not the whole channel. (b) `acted` still under-counts: a finding that
   > first fired before the window and was fixed inside it is never replayed.
   >
   > ⚠ Note what nearly happened: this file would have sat waiting for ever on a measurement nobody
   > was building, while reading as though it had coverage — the exact class `bugs_open/362` §2 was
   > filed about, authored by the same session inside a day. **A reopen trigger that names another
   > lane's work is a dependency, and it needs the same "is it still true?" check as any other.**
3. **A third writer arrives that must not repair** — two is a considered allow-list; three is a
   vocabulary, and a vocabulary belongs in a type signature after all.
4. `page_type='report'` or section-editor traffic stops being zero (the original §Recommendation 3
   trigger, unchanged and still live).

### Actions taken under this ruling

- **Recommendation 2 executed as filed**: the two known-unguarded writers
  (`create_tool_component_action.go`, `deploy_tool_action.go`) are now `bugs_open/362`, routed to
  the `webdesign_tool_rebuilds` lane that owns both files. ⚠ **The wiring has been unblocked since
  2026-08-02** — `bugs_open/180` (repair destroys JS-built anchors) was fixed the same day it was
  filed by LNK-029's span-aware repair, and both the landmine and the register still said "wait for
  180" twenty days later. Both corrected in the same commit as this decision.
- **No council submission**: this file, `bugs_open/362`, the register and `LANDMINES.md` are prose,
  which `scripts/council-scope.sh` refuses client-side. The ruling commissions no platform code. The
  `362` fix itself is platform code and is council-gated when its lane takes it.

## Related

- `bugs_open/362` — the two unrepaired tool writers, the executed half of this RFC
- `RFC_042` — the sibling column, ruled the same day and in the same posture (its §6 carries the
  joint answer)
- `bugs_closed/180` / LNK-029 — the blocker that cleared, and the reason `362` is now actionable
- `bugs_open/358` — where this RFC's unanswered measurement question went
- `bugs_open/136` (section-editor slug) and `docs024_key_docs_latest/bugfix_136_sibling_link_repair/`
- LNK-027 (the seam that shipped), LNK-024, LNK-023
- `bugs_closed/021`, `bugs_open/093` — the same "one guarded call site" family, cited by
  the seat that asked for this RFC
- RFC 007 — the same shape one subsystem over: a guard multiplied because the structural
  fix was out of scope for the lane that hit it
