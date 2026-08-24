# HANDOFF — vigilant designer + offer analyser (2026-08-24b)

**COLD-START = this file + `features_open/030` §10 (the v2 batch) + register `CLM-024` + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-24_continue_here.md`**, which is still correct about `335` and the
v2 batch's contents; everything it says about v2(d) being unbuilt is now out of date.

> **Re-run every liveness claim here before acting.** This branch takes hundreds of commits a day.
> Verify against `git archive <resolved-sha>` — never the working tree (another lane is often
> mid-edit and it may not compile: on 2026-08-24 `platform/livespec/livespec.go` was dirty and broke
> `cmd/config-key-audit`'s test build) and never the moving name `HEAD` (HEAD moved from `47d9d9198`
> to `35832a9fa` inside one session here).

## The one-line state

> **v2(d) is BUILT, TESTED and NOT LIVE.** A finding can now carry a refute-only, machine-checkable
> half of its own acceptance test; the gate that decides whether one may be stored is
> `verify_acceptance_predicates`. **Three things are owed before it does anything: an image roll, a
> hand-applied `_HOLD` migration, and a council submission that is written but undispatched.**

## What is DONE

| | state |
|---|---|
| the gate action + exported evaluator (`verify_acceptance_predicates_action.go`) | **built**, 26 tests, package green, proven against committed HEAD `35832a9fa` |
| `write_audit_findings` passthrough (2 optional fields, absent-by-default) | **built**, plus the two-decode-path lockstep test that was missing |
| `digitCardinalSpans` extracted from CLM-023's gate | **built** — one implementation of the letter-adjacency rule, old tests unchanged and green |
| registry entry | **built** |
| migration `601_offer_analyser_acceptance_predicates_HOLD.sql` (+ `_ROLLBACK`) | **written, NOT applied** |
| register `CLM-024` + index row | **written** (same commit as the seam, per the 2026-07-28 ruling) |
| LANDMINE — `pages.in_header` is not the rendered nav | **appended**, sync/dispatch OWED (see below) |
| council submission | **written + `DRY_RUN` admitted, NOT dispatched** — `SUBMISSION_2026-08-24_v2d_acceptance_predicates.json` in this directory |

## The three things owed, in order

### 1. Confirm the action is in the running binary, THEN apply 601 by hand

⚠ **The chassis build the owner started on 2026-08-24 afternoon does NOT carry this** — `make build-*`
takes **committed HEAD**, and this work was uncommitted when that build began. So the roll that
matters is the first one after commit `<this commit>`.

Probe the **capability**, not the commit, with a control in the same breath, on **every replica**:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- grep -aq "verify_acceptance_predicates"      /proc/1/exe && echo PRESENT-ok
kubectl -n ai-persona-system exec "$POD" -- grep -aq "verify_acceptance_predicates_NOPE" /proc/1/exe && echo CONTROL-FAILED
```

Expect `PRESENT-ok` and no second line. Then apply 601 **by hand** (`--apply` would sweep other
lanes' pending files) and follow the file's own three steps — they are written to be followed, and
step 2 is *"what did I break?"*, not *"did it work?"*.

⚠ **Applying it before the roll does not degrade — it breaks the agent.** A step naming an
unregistered action makes the workflow validator reject the WHOLE workflow ("requires a topic").

### 2. Fire the council submission

`./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/SUBMISSION_2026-08-24_v2d_acceptance_predicates.json`

`DRY_RUN=1` already passed every client-side validation and the scope admission check (6 edits,
29,342 plan bytes). It was not dispatched because the **kubeconfig token expired mid-session**
(fleet-wide `Unauthorized`) — and firing a publish through a dead connection is how you end up with a
printed correlation id and no dispatch behind it. Save `SUBMISSION_CORR`, budget ~30 minutes, and
record the verdict in `features_open/030` §10 and in `CLM-024`.

### 3. Sync the LANDMINE (needs cluster access)

`./scripts/landmines-verify-dispatch.sh` — **not** `landmines-sync.py --apply`, which consumes the
"new entry" status so the verifier never checks it. Owed for the `pages.in_header` entry appended
2026-08-24.

## What this feature does, in three sentences, because the shape is the whole design

A finding may carry an optional `acceptance_predicate` — one of **three** text shapes
(`text_absent`, `text_present` with `min`, `text_order`) over **`pages.meta_description` or
`pages.title`** of one named page, and nothing else. **It can only ever REFUTE**: satisfying it means
"not refuted", never "the test is met", so there is no green tick to be false — which is the answer to
the trap `features_open/030` §10 states against this feature. **And it must already refute the page at
emission or the gate discards it** (recording why under `acceptance_predicate_rejected`), because that
is the only property of a predicate checkable at the moment it is written, and it is what excludes the
vacuous predicate that would pass for ever.

## The measurement that justified it — re-check it, it is a moving target

`[MEASURED 2026-08-24, live DB + `curl`]` **three** work items marked `complete` (page rebuilt,
deployed, `commit_sha` in `result`) have their own stated acceptance test refuted by one line over the
exact field the test names. Worked case: webdesign.co.uk/index, item 08-24 10:13Z→complete 10:34Z,
test *"the meta description must state the zero-data or zero-account promise **before** any catalogue
count"*, served meta `Sixty-three browser tools for web design and development. No account, no
upload, …` — and the page was re-deployed at 12:12Z, **after** the item closed, still violating.

⚠ **A fix elsewhere could repair any of these at any time, so re-run before quoting the number.** The
query is in NOTES 2026-08-24 session 2 §1; `site_work_items` has **no `audit_source` column** (it is
`spec->>'audit_source'`, and the column form errors rather than returning zero).

## Watch-outs this session added

- **⚠ `pages.in_header` IS NOT THE RENDERED NAV.** 13 rows vs 7 served destinations on leopardess. A
  column-based nav check "found" a fourth false green and `curl` refuted it — **my own finding, before
  it reached a document.** Nav is therefore OUT of the predicate vocabulary, the reason is in the
  action's file header, and the full trap is in `LANDMINES.md`. `pages.rendered_header` is not the
  escape route: it is `''` on all 35 active pages of robot-hands.com.
- **⚠ An empty findings array is not an unresolvable findings path.** A *recognised* empty list is the
  auditor saying "nothing is wrong here" and it ARMS silence retraction (`parseAuditFindings`' third
  return, `bugs_open/213` D1 half two). The gate omits the `findings` key entirely when its input does
  not resolve; do not "tidy" that into returning `[]`.
- **⚠ `write_audit_findings` decodes findings TWICE** — struct tags for a JSON string, a hand-written
  map for a native list — and the native list is the shape an upstream ACTION hands over. A struct
  field without a matching line in `findingsFromList` is dropped in silence. Held now by
  `TestFindingsFromListPopulatesEveryTaggedField`.
- **⚠ A verdict that quotes the NEEDLE is not evidence.** The first cut said `"$cardinal" appears at
  0`; a reader cannot check that against anything. It quotes the matched text in the page's own case
  now (`"Sixty-three" appears at 0`).
- **⚠ Zero adoption on the first run is POSSIBLE AND ACCEPTABLE.** The key is deliberately absent from
  the prompt's OUTPUT skeleton (every key there reads as required), so the model is invited in prose
  only. `kept: 0, rejected: 0` means *the model wrote none* — it does **not** mean the gate failed, and
  it is **not** evidence the gate works either (a run with no predicates passes it trivially: the same
  mistake that cost this lane two days on 537's enforcement arm).

## The rest of the v2 batch — and why this did NOT batch, which was a deliberate departure

§10 says *"batch them: each on its own would cost a migration plus a live LLM re-proof"*. v2(d)
shipped alone because it is the only one of the four with a **hard ordering constraint**: it needs Go,
so its migration must be held for an image roll, and folding the config-only items into a held file
would hold them for no reason.

**The batch's real goal — ONE re-proof — is still available, and this is the recommendation:** write
**602** for v2(a) + v2(b) + v2(c) (config-only, no Go), apply **601 and 602 in the same window** once
the image has rolled, and then fire **one** B4 run. Details and traps for each remain in
`features_open/030` §10 and in `HANDOFF_2026-08-24_continue_here.md`, unchanged:
- **v2(a)** bounded head-of-hero excerpt per page — ⚠ GROWS the surface; re-run the truncation check on
  webdesign.co.uk (104 pages, `__truncated` absent, 08-15 baseline is v1's) afterwards. **It also
  widens what a predicate could address** — body-text shapes are excluded today precisely because the
  surface carries no content.
- **v2(b)** attribution in the `why` clauses — partly done by 537 (points only); re-read the live
  prompt before rebuilding.
- **v2(c)** `primary_model` in the degraded arm's field list — LATENT, no live instance; must not be
  the reason to open the batch, and do not fix it by letting the model *infer* one.

## Residual, stated plainly

**Nothing automated reads `acceptance_predicate` yet, so today this prevents no false green** — it
makes one machine-checkable. The consumer is a COMPLETION-time check ("the handler reported success
— does the item's own predicate still refute?") beside `complete_work_item_no_change.go`, whose own
comment records that grading `acceptance_test` needs *"a producer-side contract change first"*. This
is that change, for one producer; the evaluator is exported with no `ActionParams` or DB dependency so
that consumer is a call, not a rewrite. It is deliberately unbuilt here: it changes live completion
semantics for a handler other lanes own. **Whoever builds it inherits CLM-023's residual in the same
shape — the requirement belongs in the ACTION, not in one call site.**

## Who owns what nearby

Unchanged from 08-24a: the **leopardess lane** holds five of this lane's findings at
`needs_human_review` — coordinate before firing B4 at that site. `bugs_open/333` belongs to the 301
lane. `copy_quality_two_stage` + the LMC lane still work loanandmortgagecalculator.co.uk. The
**`bugfix_308_cta_destination_provenance` lane** has routed the undecidable-CTA question to this agent
(`CONTRIB_2026-08-24`, owner ruling in `RFC_047` §10) and its own honest read is *"after your v2
batch, not before"* — nothing is blocked on us.
