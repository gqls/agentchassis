# HANDOFF — 2026-08-14 (b), fresh chat starts here: batch-8 fully closed, `bugs_open/248`'s code fix is LIVE and PROVEN on gaswholesalers; mortgagecalculator's hero is a separate, tangled sub-problem

**Supersedes `HANDOFF_2026-08-12_continue_here.md`.** That handoff's whole job (the batch-8
tail, `tool-bayesian-ranking` + `tool-llm-cost-calculator`) is **done** — both items closed
the same session window, independently, by two different sessions, with no collision. This
file exists because a large amount of *different* work happened afterward, on
`bugs_open/248`, and it deserves its own cold-start rather than inheriting a stale story.

**Update, same day, after a second fresh roll:** re-verified against `v1.0.1299`
(`6f8efa158…`) — no regression, `930ace3bd` still an ancestor, gaswholesalers' logo still
200. Went to repeat the same proof on mortgagecalculator's hero and found its situation is
NOT a simple re-trigger — see the new §2b.

## 1. State (verified 2026-08-14, re-verified same day against a second roll)

- **Batch-8: CLOSED.** Nothing left on it. If a fresh session is tempted to re-check it,
  don't — `CHECK_naming_contract.sh` already reflects both landings (0 broken class).
- **Fleet: `v1.0.1299`** (was `v1.0.1298` earlier the same day — the fleet rolled again
  mid-session; re-verify yet again by the time you read this, it moves fast on this tree).
  Chassis + browser-runner-adapter confirmed on the SAME commit
  (`6f8efa158ea3365bea79eec0de0283041ed54842` at the second check;
  `bc39e7bf547e9d5db07c92085be85c6874654774` at the first) via the binary-probe-with-controls
  method both times (chassis's own provenance log line had already scrolled out of even a
  `--since-time` pull from pod start, both times — a busy-service rotation issue, not a
  time-window one; fell back to cross-checking chassis's binary for browser-runner's own
  known commit, positive AND negative control, both as expected, both times). **`930ace3bd`
  (the `bugs_open/248` Go fix) is confirmed `git merge-base --is-ancestor` of BOTH builds —
  the code half of 248 is LIVE and has not regressed across a fleet roll.**
- **`bugs_open/248` (the placeholder-filename bug — NOT the same-numbered CTA bug, see §5
  below) has been through FOUR council rounds, all REVISE, each answering the last with real
  evidence rather than argument:**
  - **R1** (`bugfix_153`... no — filed by this lane, corr `7f0c1535…`): gating HIGH from
    `editquality` — `assetRowIdentity`'s `asset_id` recovery can fall through to the
    landmined aggressive recursive search (`findFieldRecursive`) when a caller doesn't map
    it explicitly. Traced and confirmed true for `image-build-handler`'s
    `call_asset_deployer` step. Fixed with **migration 401** (one optional config key,
    `asset_id? -> asset_stored.asset_id`, live, no roll needed).
  - **R2**: `editquality` again — the SAME gap exists on the OTHER caller
    (`check_undeployed_assets` → `build-dispatch-loop` → `asset-deployer`, the repair path),
    asserted "safe" in R1's own resubmission without being traced. Fixed with **migration
    402**, mirroring an ALREADY-REVIEWED precedent (**migration 380**, `bugs_open/231`) that
    fixed the identical nested-spec shape for a different field (`purpose`) on the SAME
    shared `build-dispatch-loop` mapping. Blast radius measured first: exactly one
    `(item_type, handler_agent)` pair fleet-wide carries `asset_id` in its spec —
    `(undeployed_asset, asset-deployer)`, 267 rows. R2 also cited two landmines that didn't
    hold up on inspection: one is explicitly **RETIRED** (2026-08-06, `bugs_open/155`), one
    describes a genuinely open but **different** bug on a **different** branch
    (`bugs_open/235`, `image-build-handler`'s brand-update path, not the one 248 touches).
  - **R3**: `guardian` — two HIGH, both concrete and checkable, both answered CLEAN on
    inspection rather than argued away: (1) whether `build-dispatch-loop`'s live
    `process_item` step actually executes under `sub_workflow` (migration 402's target) or
    the landmined `substeps` shape instead, where only one nesting runs — checked the live
    row directly (`substeps=false`, `sub_workflow=true`) and re-ran the fleet-wide `substeps`
    census fresh (still **0**, matching the 2026-08-08 measurement); (2) whether either
    target agent type is one of four known to carry two active rows (version ambiguity) —
    checked, both have exactly one active row each.
  - **R4** (current, corr `7f0c1535…`, run `8a36998f-7188-4cda-9a44-f8a99f71e6a0`,
    `complete_revise`, **completed**): `bug_historian` gated, but the objection is
    **qualitatively different** from R1-R3 — it is an **architecture-scope question**, not a
    defect in the specific edits: *"should `ExtractActionInputs`'s fallback strategies ever
    run for a field with no explicit `input_fields` entry, full stop?"* Its own text says
    "Not a veto… should be flagged to a human independent of this round's disposition."
    **This is the point to stop resubmitting** — CLAUDE.md's own council-gate section is
    explicit that a scope-shaped objection "is not answered by resubmitting with better
    measurements… route the seam to architecture review on its own merits, and let a human
    break it." Ran the two concrete checks R4 also proposed (both fine — no third live
    caller of `asset-deployer` exists beyond the two already patched, confirmed by checking a
    plausible-looking false-positive `render-audit-agent` match and finding it was just
    descriptive prose, not a real dispatch edge; `asset-deployer`'s `input_contract` doesn't
    even list `asset_id`, so edit 5's "required field" framing doesn't hold as stated).
- **Council-Submitted, not yet reviewed-approved.** No commit here needs
  `Council-Reviewed:` — none of this session's own commits (`956bf19c6`, `f5386b8f9`,
  `278b104a0`, `14307d38d`) touch `platform/`/`internal/`/`pkg/`; the actual platform-code
  commit is `930ace3bd`, made by an earlier session, which already carries
  `Council-Submitted: 7f0c1535…`. **Do not write `Council-Reviewed:` on it** — no round has
  approved yet.
- **The code fix is live; the artefacts it fixes are NOT yet repaired.** Confirmed live:
  `curl https://gaswholesalers.com/assets/images/logo.png` and
  `curl https://mortgagecalculator.co.uk/assets/images/hero.jpg` both still **404** as of
  this writing — the fix prevents *future* placeholder deploys, it does not retroactively
  repair the ~146 rows / 15 sites already committed under the wrong name (the bug file's own
  words: "it will not fall on its own"). **A live verification was in progress when this
  handoff was written** — see §2.

## 2. RESOLVED before this handoff was finished writing — the fix is proven, not just live

Promoted the existing `unresolved` `undeployed_asset` work item for gaswholesalers.com's
logo (`edff6d42-9c5d-4777-af27-be7c6d558f74`, asset `b99c5355-4b3a-430c-9294-56482726be34`)
to `triaged` at `2026-08-14 13:55:48Z`, specifically to test whether the NOW-LIVE fix
deploys `logo.png` correctly this time (previous attempts, pre-fix, produced
`input-data.asset-key.jpg` and, on one hand-corrected retry, `logo.jpg` — both wrong).

**It worked, first try, exactly right.** Claimed by `build-dispatch-loop` in 19s, complete in
53s. `deploy_result`: `file_path: "/assets/images/logo.png"`,
`commit_message: "Deploy logo image for gaswholesalers.com"` — the correct extension AND the
correct purpose name, both of which every prior attempt on this exact asset got wrong.
**Verified at the served artefact, not the status**:
`curl https://gaswholesalers.com/assets/images/logo.png` → **HTTP 200, 42,211 bytes**. The
symptom named at the top of this bug file ("→ 404… four months") is resolved.

**This is real, end-to-end, disconfirming-capable proof** — the run could have reproduced
the old bug (it had, twice before, on this exact asset) and didn't. Recorded in NOTES and the
bug file.

## 2b. Went to repeat the proof on mortgagecalculator — found a SEPARATE, tangled situation, did not fix it

`mortgagecalculator.co.uk/assets/images/hero.jpg` is still **404**. Looked for the same kind
of stalled, re-triggerable work item gaswholesalers had, and the situation is genuinely
different, not just "hasn't been re-tried yet":

- The homepage's own `background-image` really does reference `/assets/images/hero.jpg`
  (checked at the served page, not assumed).
- The matching asset row — `purpose='hero'`, `asset_key='hero'` (the plain, site-wide
  homepage hero, as opposed to the per-page variants below) — has **NO active row at all**.
  The two most recent attempts (`9e94250d…`, updated 2026-08-12; `d6ead260…`, 2026-08-11) are
  both `status='superseded'`, and one earlier one (`477838e3…`) is `status='rejected'`.
- This is DIFFERENT from the per-page heroes on the same site (`asset_key='hero_about'`,
  `'hero_contact'`, etc.), which ARE active and presumably serving fine — those use distinct
  per-slot keys that don't collide the way the bare `hero` purpose apparently does.
- Every `needs_hero_image`/`undeployed_asset` item on this site for plain `purpose='hero'`
  is already `complete` or `cancelled` — none sitting `unresolved`/`triaged` ready to
  promote the way gaswholesalers' was. Matches the bug file's own 08-12 contribution:
  "`needs_hero_image` has been filed five times here (3 cancelled, 2 complete) and
  `image_url_404:hero.jpg` has been blocked since 2026-08-05."

**Did not create a new work item or force anything** — that would be a heavier, more
speculative action than promoting an already-stalled one, and this site's history (repeated
supersede/reject/cancel cycles on exactly this asset) suggests there may be a second,
different mechanism in play here beyond the plain placeholder-filename bug 248 already
fixed — worth understanding BEFORE dispatching another attempt, not after. Whoever picks
this up next should start by reading why the two most recent `hero` (plain) generations
were marked `superseded` rather than assuming a fresh dispatch will behave like
gaswholesalers' did.

## 3. What's actually left on `bugs_open/248`

1. **Route R4's architecture question to a human / RFC, don't resubmit a 5th time.** The
   question — should `ExtractActionInputs`'s Strategy 4 (aggressive recursive search) ever
   run for a field with no explicit `input_fields` entry, or should an unmapped field always
   be a hard refusal — is bigger than this bug and shapes every future caller of the shared
   mechanism, not just these two. This is exactly `architecture_review/`'s job, not another
   round of evidence-gathering on the same submission.
2. **`mortgagecalculator.co.uk/assets/images/hero.jpg`** — investigated, NOT fixed (see
   §2b). No stalled item to promote; the plain `hero` asset_key has no active row and a
   history of superseded/rejected generations. Understand why BEFORE dispatching a fresh
   attempt — this may be a second, distinct mechanism layered on top of 248's own defect,
   not just "the same bug, one site behind."
3. **Design the backlog-drain job.** ~146 rows / 15 sites already committed under the
   placeholder filename will not repair themselves. Nobody has designed this yet. §2 proves
   the corrected deploy path works, so the drain is mechanically "for each affected asset
   row, re-trigger its repair item the same way §2 did" — but the **framework rule stands:
   no hand-placed artefacts**, and re-triggering 146 real deploys is exactly the kind of
   bulk, higher-blast-radius action that deserves its own careful plan (batching, rate,
   which sites are `owned` vs `generic` rebuild policy) rather than a one-off script
   improvised at the end of a session.
4. **`bugs_open/256`** (mobile screenshot exceeding the vision API's 8000px cap, found as a
   side effect of the batch-8 work) — still open, still explicitly **not this lane's job**;
   belongs to whoever owns `run_checks_action.go`'s capture path.

## 4. What's explicitly NOT this lane's job

- The email/mailer/contact-form work from the 08-12/08-13 handoffs turned out to be a dead
  end for THIS lane: `bugs_open/228` (contact-block delivery) is already fixed and live
  (mailto branch, proven end-to-end); the fence gap that looked open was already closed by
  the time it was checked (a stale note, corrected in NOTES rather than re-fixed). The
  `platform/mailer` (PUB-003) work — a real `SendMail` timeout bug on the port this estate
  must use — is genuine but is **architecture-scope** (a new http receipt endpoint decision,
  an SES credential only the owner can provision) and was correctly NOT picked up as a quick
  task. If a fresh session wants it, it needs its own scoping, not a continuation here.
- `bugs_open/256` — see above.
- `bugs_open/235` (brand-update branch hardcodes `purpose:'hero'`) — real, open, but a
  DIFFERENT branch than 248 touches; not this lane's to fold in.

## 5. Landmine this session hit and is worth repeating

**Bug numbers collide, and 248 is one of them.** There are TWO unrelated
`bugs_open/248_HANDOFF_*.md` files: this one
(`..._undeployed_asset_repair_deploys_every_asset_as_a_hero_under_a_placeholder_name.md`) and
a completely separate CTA-link bug
(`..._cta_recompute_clobbers_authored_contact_links.md`), owned by a different lane
(`bugfix_203_phantom_cta_cleanup`). Resolve by filename, not by number, every time — a bare
"248" in conversation or a commit message is ambiguous.

**A `timeout N` wrapper on a script that publishes via `kubectl run --rm` can be killed
mid-flight without a clean answer about whether the message sent.** Happened this session
re-running the council trigger just to see truncated output — cost a few minutes of
uncertainty (resolved by checking `orchestration_states` for the correlation: only one row
existed, so no duplicate landed) and left a stray `kcat-cgate-*` pod that needed a manual
`kubectl delete`. Don't `timeout` a submission script to peek at output; let it finish.

## 6. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK.
2. Pod-grep chassis + browser-runner for the CURRENT build (method: chassis's provenance
   line scrolls out fast on this fleet — cross-check via a known-good sibling service's
   commit + a bogus negative control, per §1, rather than a blind binary grep). The fleet
   rolled twice in one day already this session — assume it has moved again.
3. `who-owns.py 248` (both files will show — resolve by filename) and grep live transcripts
   before touching the drain-job design, the R4 architecture question, or mortgagecalculator's
   hero (§2b).
4. If picking up the architecture question: that's `architecture_review/`'s process
   (RFC-shaped), not a normal task — read `docs/agent_docs/docs024_key_docs_latest/
   architecture_review/` for the current convention before starting one.
5. If picking up mortgagecalculator's hero (§2b): read why the two most recent `hero`
   generations were `superseded` before dispatching anything new.
