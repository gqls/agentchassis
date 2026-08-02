# NOTES — bugs_open/174 seed_scope relay

Append-only, newest at the bottom. The missteps are the point.

---

## 2026-08-02 — picking it up

Checked ownership before starting: `scripts/who-owns.py 174` → no owning
workstream; last touched by `e699ab60a` (the `bugfix_164` lane closing 164 and
filing this as a by-product). Grepped the live `.jsonl` transcripts for the two
sessions active today — one on `bugs_open/151`, one on `bugs_open/138`, neither
on 174. `who-owns.py` reads commits and is blind to a session mid-fix, hence the
transcript grep as well.

Re-verified the bug is still real against live config before doing anything: the
loop's `call_handler` mapping had 10 keys, no `seed_scope`, no `runtime_page`.
Still valid.

## The correction that defines this lane: the ticket's fix candidate 1 is insufficient

The ticket says to add `seed_scope?: claimed.seed_scope` to the input_mapping.
Reading `claim_item` showed **`claimed.seed_scope` does not exist** — the
`RETURNING` clause is itself an allow-list and projects nine spec keys, not
including it. So candidate 1 alone maps from a path resolving to nothing, and the
key being optional, `ResolveInputMapping` drops it in silence. **A fix for a
silent-drop bug that silently drops.**

Then a third gate, found by following the type rather than the key:
`QueryDatabaseAction` stringifies every `[]byte` a column scan returns
(`database_actions.go`), and `ExtractStringListHelper` returned nil for a string.
The 090 trigger's own comment had already written this down —
*"ExtractStringListHelper takes []interface{} or []string only; a bare "a,b"
string yields nil and the seed is ignored"* — which is the strongest possible
evidence, sitting in the repo the whole time.

## Misstep 1 — I nearly built a detector that was blind to its own bug

Fix candidate 2 asks for a lockstep check. I wrote the obvious general version:
*every `call_agent` must forward every key its callee declares*. **Measured it
before believing it**: 31 findings from 75 resolvable call sites. Spot-checked
one (`pageflow-builder.apply_site_design` → `webdesign-agent` omits
`site_context`) and it was legitimate — the callee has `else_step:
load_site_context` and loads it itself.

Tightened to "the callee also READS `input_data.<key>`": 3 findings. Better, and
**still wrong**, because it cannot distinguish "the caller dropped it" from "the
caller never had it".

Then the actual catch: **both versions returned 3 findings that did not include
174.** `call_handler` resolves its callee through `agent_type_field:
claimed.handler_agent` — a runtime value — so a static resolver skips it. I was
about to ship a check that would pass for ever on the bug that motivated it.
This is exactly `narrowing-a-detector-can-make-it-inert`, and the only reason I
caught it is that I ran the tightened version and looked at *which* 3 it found
rather than at the count.

**The cheap check that would have caught it sooner:** run any new detector
against the state that motivated it, first, and require it to FIRE. I eventually
did this properly — see below — but I designed two versions before testing that.

## Misstep 2 — `min(jsonb)` does not exist, and the migration failed on first run

The migration's final assertion did `SELECT count(*), min(default_config #> ...)`.
There is no `min(jsonb)` aggregate. It failed mid-transaction. **The transaction
rolled back cleanly and live config was untouched** — verified before retrying,
rather than assuming `ON_ERROR_STOP` had done its job. Split into a count
statement and a value statement.

## Misstep 3 — I gave the council an inflated blast-radius figure

My submission said *"14 live `query_database` steps project json/jsonb"* as the
reason for not fixing `QueryDatabaseAction`. That number came from a loose regex
over the whole query text, which also matched `->>` **text casts** and `->` inside
**WHERE predicates** — neither of which is output.

> **CORRECTED 2026-08-02: the real figure is ONE, and it is the projection this
> fix added.** Re-measured by extracting only `-> '...' AS <alias>` projections,
> then reading all three ambiguous queries in full rather than trusting the
> tightened filter — `model-directory-trigger` and `content-feed-trigger` use the
> arrow in a JOIN predicate terminating in `->>::boolean`, and
> `claims-auditor.load_evidence_facts` projects `string_agg(...)`, which is text.

The error was in the cautious direction (it made the deferral look *more*
necessary), and the council approved anyway. It was still an unmeasured number
stated as a measurement, in the one place a reviewer was most likely to lean on
it. Logged in `WRONG_CALLS.md`.

The corrected figure is *better* news: the deferred `QueryDatabaseAction` fix has
**zero currently-affected consumers**, so the trap is genuinely prospective — a
landmine for the next author, not a live defect.

## Misstep 4 — my first live run of the detector matched NOTHING, and said so

`validation.WalkSteps` qualifies every step path (`steps.call_handler`), while the
registry names steps as an author writes them (`call_handler`). Equality matched
nothing, so the registered relay was checked **zero** times.

It did not report clean. It reported one **unmatched registry entry**, which is a
category I had added ten minutes earlier for exactly this reason and did not
expect to need immediately. That is the entire argument for the category: an
assertion that stopped running must be louder than a finding, not quieter.

Fixed with a `stepName`/`lookupStep` normalisation, and pinned by
`TestFindRelayGaps_UnmatchedRegistryEntryIsReported`.

## Proving the work, rather than passing

- **Go tests by MUTATION**, not by passing: removed the JSON-text arms of
  `ExtractStringListHelper` and confirmed the failures — `scope_source` came back
  `code_results` instead of `seed`, the `[]byte` subtest failed, and the
  bundle-note test failed. Restored, green. The negative-control tests correctly
  kept passing under the mutation, since they do not depend on that arm.
- **Detector by FIRING**: rebuilt the pre-fix config from migration 289's own
  snapshot (`agent_definitions_backup`) and ran the real binary over it. Exit 1,
  naming `seed_scope` and `runtime_page` in **both** `not_projected` and
  `not_forwarded`. Exit 0 against live. Exit 2 on an empty export.

## What the council caught (corr `081d98b3`, APPROVED round 1, 6 advisory)

Four were answered with work:

1. **`reuse_agent` (medium)** — "no search for an existing JSON-list decode
   helper before adding `stringListFromJSON`". **Right.** `SafeUnmarshal` was
   already in the same package, doing exactly "attempt an optional parse without
   propagating an error". Rewired to use it.
2. **`bug_historian` (medium)** — "a LANDMINES entry is the minimum acceptable
   mitigation and is not yet a committed deliverable", plus a request for a
   census of the other query_database sites. Both done — and the census is what
   corrected misstep 3 above.
3. **`guardian` (medium)** — "name the other consumers and confirm their
   behaviour on a newly-non-nil result, don't just assert it safe". **Measured**
   across live config: the three `prefer_domains` values are already JSON arrays
   (unaffected); `training-launcher`'s `checkpoint_keys`/`checkpoint_urls`/`keys`
   are field-path strings (`"ckpt_keys.checkpoint_keys"`) which do not start with
   `[` and so still return nil. **No live config changes answer.**
4. **`prior_art_librarian` (low)** — "`runtime_page` has no consumer" was asserted
   by prose grep with no check attached. Re-ran it unfiltered over the whole
   repo: 8 hits, **all of them in files this lane wrote**. Zero pre-existing
   producers or consumers. Claim was right; it now has its check.

Two were noted and not acted on: `editquality` observing that the already-applied
migration is historical context rather than a pending edit (true), and
`tooling_provenance` asking for a `doc_notes` entry for the next lane touching
the diagnose loop — which is what this directory is.

## Concurrency: two collisions with other sessions in one hour

- **`cmd/config-key-audit/main.go`** was being edited by another session
  (RFC 006 / `--single-owner-actions`) while I needed a two-line dispatch in it.
  Committing it would have taken their work as a same-file passenger — and at
  that moment their `singleowner.go` was still **untracked**, so it would have
  **broken the build at HEAD**, which every `make build-*` compiles from. Left it
  alone; `relaygaps.go` decodes the export independently instead of widening
  their `liveAgent` struct. Verified HEAD compiles from `git archive HEAD` after
  each of my commits.
- **Migration number 285** was taken by another session eight minutes after mine
  landed. Their commit subject names "285" and mine did not, so I renumbered
  **mine** to 289 — `git mv` with **both** paths named on the commit, verified at
  HEAD with `git ls-tree` (not `ls`, which cannot tell you: the file is gone from
  disk either way).

## Disclosure — my docs commit carried another session's LANDMINES edit

`b34717c18` reports "2 lines removed from LANDMINES.md" in the pattern check. I
only appended. The two lines are the `storage.DeployedWebPath` entry's old
`footprint` and `source` lines, **rewritten in the working tree by the
`bugs_open/168` lane** while I was working: they added a "Reading the SOURCE at
HEAD" note, widened the footprint, and recorded that
`TestDeployedWebPathCannotExpressBrandHeadPaths` had fired and been inverted.

Checked before assuming: the new lines are a strict superset of the old ones.
**Nothing was lost** — the pattern check reads a replacement as a deletion, which
is the correct conservative default for an append-only ledger.

But it *is* a same-file passenger, and my commit message did not say so. This is
the case CLAUDE.md names as unpreventable — "if two sessions edit one file,
whoever commits takes both edits, and no hook can prevent that". Recording it
here because the commit message cannot be amended (forward-only). The `168`
lane's LANDMINES improvement is in `b34717c18` under my message.
