# RFC 008 — a mandatory write seam for `page_components.rendered_html`, and why an advisory lint is the wrong ceiling

**Filed:** 2026-08-02 by the `bugfix_136_sibling_link_repair` lane, discharging the
`bug_historian` seat's explicit instruction in council round 1 of correlation
`0275f9c2-035f-4c9e-8a50-83836dfeffd9` (APPROVED, 5 advisory objections, none high):

> *"Flagging as architecture-level for a human: candidate 3 (mandatory write-seam for
> `page_components.rendered_html`) should be tracked as a real ticket, not left implicit
> in a lint rule's silence."*

**Status:** open. Not urgent — measured traffic on the newly-guarded paths is zero — and
deliberately NOT actioned by the filing lane, because taking it inside a bug patch is
precisely what CLAUDE.md's platform-seam ruling forbids.

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

## Related

- `bugs_open/136` (section-editor slug) and `docs024_key_docs_latest/bugfix_136_sibling_link_repair/`
- LNK-027 (the seam that shipped), LNK-024, LNK-023
- `bugs_closed/021`, `bugs_open/093` — the same "one guarded call site" family, cited by
  the seat that asked for this RFC
- RFC 007 — the same shape one subsystem over: a guard multiplied because the structural
  fix was out of scope for the lane that hit it
