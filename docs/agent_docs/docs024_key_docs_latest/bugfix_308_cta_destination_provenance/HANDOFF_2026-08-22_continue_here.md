# HANDOFF — `bugs_open/308`, CTA destination provenance. Continue here.

**Written 2026-08-22 ~19:00Z.** Lane dir:
`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`

Read this, then `NOTES_cta_destination_provenance.md` (the technical log, newest at the bottom)
and `PLAN_2026-08-22_cta_destination_provenance.md` (the design). `README_where_we_are.md` is the
owner's plain-prose log.

---

## 1. State in one paragraph

**Phase A is written, committed, council-reviewed twice (REVISE both times, round 3 in flight),
and LIVE on the chassis as of the 2026-08-22 15:10Z roll — verified at the binary on both
replicas.** It changes no behaviour on its own; that is deliberate. It is the record that makes
Phase B (the actual repair of the 200 findings) safe. **Phase B has NOT started.**

## 2. What Phase A did

Replaced LNK-033's **derived** provenance (*"the resolver cannot produce a utility url, so a
stored one was authored"*) with a **recorded** one: `page_components.content_data.__cta_minted`,
shaped `{url_field: minted_url}`. A stored link is authored when it is a valid page **and** the
record does not name its current value.

**Value-bound is the whole design.** It records *which* url, not a boolean. That makes 8 of
RFC_042's 9 `content_data` writers correct with zero changes, and is the only form that survives
the section-editor MERGE — a presence-bound stamp would licence the recompute to clobber a
human's edit, which is `bugs_open/248` again.

Registered as **LNK-035**; **LNK-033 amended visibly** (its invariant is retired by owner ruling,
not quietly broken).

Commits: `288ce3e7a` (code + register), `cbef38e7a` (round 3 + dated counts),
`577cae3ca` / `3484c978b` (docs). Landmine + WRONG_CALLS rows were swept into other sessions'
commits (`e7bf70cc9`, `62291fa66`, `ce3ca376d`) — content is in HEAD, `git blame` misattributes.

## 3. THE NEXT ACTION — the 0→N proof, and nothing is proven until it moves

**BASELINE RECORDED 2026-08-22 after the roll: `0` stamped rows of `1866` rows with
`content_data`.**

```sql
SELECT count(*) FILTER (WHERE content_data ? '__cta_minted') AS stamped,
       count(*) FILTER (WHERE content_data IS NOT NULL)      AS with_content_data
FROM page_components;
```

To move it you must induce **BOTH** writers — they are independent and one proves nothing about
the other:

1. a **full page build** (exercises `setCTAField`), and
2. a **`cta_links_stale` rerender** (exercises `applyCTARecompute`).

Then the negative control: non-CTA components must stay stamp-free.

⚠ **The detector has not run since 2026-08-19.** The 200-finding census returns ~200 whether
anything works or not, so it is NOT a measure of this fix. Any verification must induce a
discovery run and then read a **served page**.

## 4. ⚠ BLOCKED — the estate's LLM budget is exhausted until 2026-09-01 00:00 UTC

**Round 4 was dispatched and NEVER REVIEWED.** It completed at terminal step `complete_invalid`
with no `council_report` row, which reads exactly like "the gate rejected my submission". It is
not. The real cause is in `__step_error`:

> `execute_llm_prompt` … **HTTP 400** `invalid_request_error`:
> *"You have reached your specified API usage limits. You will regain access on 2026-09-01 at
> 00:00 UTC."*

**Do NOT resubmit before that date** — a retry cannot succeed and each one re-runs the seats that
did answer. `DRY_RUN=1` will keep passing, because it validates locally and never calls a model.

**It is NOT confined to this lane** [MEASURED 18:15:39–18:26:25Z]: 7 failed steps across 5
orchestrations, and the failing steps include **`call_content_writer`** — live site content
generation. 95 orchestrations still reached COMPLETED in that hour so it is not a total outage,
but zero have completed carrying `__usage_output_tokens` since the last failure. `[UNVERIFIED]`
whether the block is model-specific (the error names `claude-sonnet-5`) or account-wide.

**This does not hold Phase A.** The council is advisory and cannot block a commit; Phase A is
committed, live, and carries `Council-Submitted:`, which asserts nothing and can never become a
false claim. Rounds 1–3 were substantive (round 3 approved 10–3) and every objection is answered
in the tree. What is missing is the final verdict, not the review.

## 4b. Council history — rounds 1-3 all REVISE

`SUBMISSION_CORR = e4336931-487b-4db3-b4dc-a4b128b3566c`

```sql
SELECT iteration, created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id='e4336931-487b-4db3-b4dc-a4b128b3566c' AND kind='council_report'
ORDER BY created_at;
```

⚠ **Two traps, both hit here.** (a) Never read the verdict via CLAUDE.md's
`doc_notes ORDER BY created_at DESC LIMIT 1` — with ~40 live sessions it returns whoever finished
last. (b) **A resubmit writes another `iteration = 0` row**, so a watcher keyed on `iteration > 0`
waits for ever. Key on the correlation and count rows, or compare `created_at`.

**Rounds 1, 2 and 3 were ALL REVISE, every one `decided_by` a gating objection from
`editquality`, and all three on ONE class of defect: my hand-written sketch drifting from code
that was already correct.** Round 3 approved 10–3. The work was never the problem; the submission
was. **Fixed structurally in round 4 — sketches are now generated from the committed diff with
`git show`**, which immediately caught a fourth instance before dispatch (a naive line truncation
cut `SeedCTAMinted`'s body, reproducing round 2's objection inside the fix for it). Full account
in `WRONG_CALLS.md`. **If round 4 also revises, read the objection carefully before changing
code** — three times running the code was right.

Round 3's other seats found two things worth keeping: **four** discovery checks raw-iterate
`content_data` (as of 2026-08-22), and `check_literal_markdown`'s `walkContentDataStrings`
**skips every key beginning `_`** as platform metadata — which is both a second precedent for the
convention and the thing that stops `__cta_minted` convicting itself, since `__` is markdown bold.
And the "no generic per-field provenance helper" absence claim survives only on the *broader*
grep: adding `fieldOrigin|valueOrigin` returns 16, all `sectionsMetadataFieldOrigin` (the origin
of a config key, not of a content value).

**The best objection across three rounds** was guardian's round-2 HIGH: does the key perturb
`content_hash` / divergence fleet-wide? It named a blast radius I had not considered. Measured: it
does not — `pages.content_hash` is the sha256 of the **committed page bytes**
(`v3_site_actions.go:1057`), `page_components.content_hash` has **zero** Go writers, and
`rendered_html_digest` is md5 of rendered HTML. All hash rendered output, and across **341** live
templates (counted 2026-08-22) there are **zero** root-scope map ranges, so the key cannot render.

## 5. Phase B — NOT started. Design constraints that are already settled.

1. **Widen at `candidatesFromHubs`, NEVER at the loaders.** `loadContentHubs` /
   `loadInteractivePages` have **three** non-test callers as of 2026-08-22 — the third is
   `render_site_components_action.go:182-190`, the **site header CTA fallback**, which no CTA bug
   file or register entry named. Widening at the loaders silently re-picks every site's header
   button, and `site_components` holds **0** `cta_url` keys across all **24** header rows, so no
   `content_data` diff could see it.
2. **Rewrite the asymmetry pin** `TestFreshPickRefusesUtilityWhileStoredUtilityIsKept` — Phase A
   left it byte-identical on purpose (308's own bar #2: editing it without provenance landing
   means the fix is wrong). Provenance has now landed, so Phase B may rewrite it.
3. **Recalibrate stopwords with a measured report, not an assumption.** `about` was deliberately
   kept OUT of `LabelStopwords` in the 2026-08-08 calibration, so "…about your use case" will
   match an About page the moment About becomes a candidate. Known case to explain: "how we work"
   → `/about.html` (n=12) is an alphabetical **tie** loss today.
4. **Make the repair consume the detection.** `suggested_target` has **no consumer** anywhere in
   the tree — the detector computes the answer, writes it down, and the repairer re-derives it
   from a narrower list. Preferred shape: one shared `LoadCTALabelUniverse` + move
   `ctaClassifyAnchor` to `datahelpers`, plus a completion verifier
   (`VerifyMisdirectedCTAResolved`) that re-runs the detector's own predicate before a
   `page_rerender` may complete — that is what turns the 112 "complete and unchanged" outcomes
   into refusals. Do **not** have the repair execute the stored `suggested_target`: a work item's
   spec is data written by an earlier binary.
5. **Migration `555_requeue_misdirected_cta_stock.sql`** — Phase C only, and only AFTER the Phase
   B image is stamp-verified per service. A status flip is live instantly; flipping under the old
   binary re-runs the broken repair and burns strikes. `[UNVERIFIED]` that `unresolved` is
   non-dispatchable and `detected` is — read the dispatch loop's status predicate first.

## 6. Landmines this lane added or hit

- **NEW:** *"The envelope decode DROPS every `__`-prefixed key…"* — `isEnvelopeMarkerKey` is
  `strings.HasPrefix(k, "__")` with no allow-list. Bounding it by today's population guards
  nothing. Design test: **decide which direction your marker's ABSENCE fails in, and make absence
  the safe reading.** `__cta_minted` passes deliberately (absent ⇒ authored ⇒ frozen, never ⇒
  recomputable).
- **Extended:** the cwd-persistence entry gained a **second form** — the same trap that makes
  `ls`/`find` report a false ABSENCE also makes `cd reldir && cat > f <<EOF` discard a heredoc
  while a trailing `echo` prints success (a false PRESENCE). It cost this lane's RUNBOOK.

## 7. My own errors, so you do not repeat them (full text in `WRONG_CALLS.md`)

- **Named the wrong fix site.** I said PBP-039's carry must carry the stamp. Wrong: `setCTAField`
  reads a **fresh DB row**, not the carry's output. *A claim that data flows A→B is not checkable
  at A — it is checkable at B's call site.*
- **Wrote "(verified)" on a mutation I had not run** — inside the edit correcting a *different*
  unverified claim. Being mid-correction felt like care and wasn't.
- **Asserted `NormalizePagePath` equates `/contact.html` with `/contact/index.html`** in three
  places before running it once. It does not.
- **Two defects mutation caught that reasoning did not:** my first version would have caused the
  very freeze it prevents (shallow merge dropping a sibling slot's record), and the guard I wrote
  for it was **deletable with every test in the repo still green** — so both seed calls now live
  *inside* the two writers. A second helper turned out to be dead and was deleted.

## 8. Working-tree hazards

- `platform/orchestration/actions/flag_page_image_rebuild_action.go` was mid-edit by another
  session and **did not compile** (`emitContentCardDerive` undefined). All Phase A testing ran in
  a scratch tree from `git archive HEAD` + my files. Re-check before trusting a local `go test`.
- `who-owns.py 308` says the `cta_target_content_pass` lane owns it; that lane was told (CONTRIB
  in its NOTES 2026-08-22) and its phase-1 open question is answered by this design. It is a
  *content* pass, not a competitor.
