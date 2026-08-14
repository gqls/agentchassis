# HANDOFF — 2026-08-14, fresh chat starts here: batch-8 fully closed, `bugs_open/248`'s code fix is LIVE on the fleet, its council review is stuck on a genuine architecture question

**Supersedes `HANDOFF_2026-08-12_continue_here.md`.** That handoff's whole job (the batch-8
tail, `tool-bayesian-ranking` + `tool-llm-cost-calculator`) is **done** — both items closed
the same session window, independently, by two different sessions, with no collision. This
file exists because a large amount of *different* work happened afterward, on
`bugs_open/248`, and it deserves its own cold-start rather than inheriting a stale story.

## 1. State (verified 2026-08-14 ~14:00Z)

- **Batch-8: CLOSED.** Nothing left on it. If a fresh session is tempted to re-check it,
  don't — `CHECK_naming_contract.sh` already reflects both landings (0 broken class).
- **Fleet: `v1.0.1298`.** Chassis + browser-runner-adapter confirmed on the SAME commit
  (`bc39e7bf547e9d5db07c92085be85c6874654774`) via the binary-probe-with-controls method
  (chassis's own provenance log line had already scrolled out of even a `--since-time` pull
  from pod start — a busy-service rotation issue, not a time-window one; fell back to
  cross-checking chassis's binary for browser-runner's own known commit, positive AND
  negative control, both as expected). **`930ace3bd` (the `bugs_open/248` Go fix) is
  confirmed `git merge-base --is-ancestor` of this build — the code half of 248 is LIVE.**
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
bug file. `mortgagecalculator.co.uk/assets/images/hero.jpg` was NOT re-tested this session —
same mechanism, different site, reasonable to expect the same result, but say so as an
expectation, not a re-verified fact, until someone actually triggers and checks it (its own
`needs_hero_image` items would need the equivalent promotion/re-dispatch).

## 3. What's actually left on `bugs_open/248`

1. **Route R4's architecture question to a human / RFC, don't resubmit a 5th time.** The
   question — should `ExtractActionInputs`'s Strategy 4 (aggressive recursive search) ever
   run for a field with no explicit `input_fields` entry, or should an unmapped field always
   be a hard refusal — is bigger than this bug and shapes every future caller of the shared
   mechanism, not just these two. This is exactly `architecture_review/`'s job, not another
   round of evidence-gathering on the same submission.
2. **`mortgagecalculator.co.uk/assets/images/hero.jpg`** — same mechanism, not yet
   re-tested. Find its own `needs_hero_image`/`undeployed_asset` item(s) and repeat §2's
   promotion-and-watch, then verify at the artefact.
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
2. **Read §2 first** and resolve the in-flight verification before anything else.
3. Pod-grep chassis + browser-runner for the CURRENT build (method: chassis's provenance
   line scrolls out fast on this fleet — cross-check via a known-good sibling service's
   commit + a bogus negative control, per §1, rather than a blind binary grep).
4. `who-owns.py 248` (both files will show — resolve by filename) and grep live transcripts
   before touching either the drain-job design or the R4 architecture question.
5. If picking up the architecture question: that's `architecture_review/`'s process
   (RFC-shaped), not a normal task — read `docs/agent_docs/docs024_key_docs_latest/
   architecture_review/` for the current convention before starting one.
