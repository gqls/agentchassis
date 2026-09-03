# HANDOFF — bugfix 417/420 lane — 2026-09-03, continue here

**Supersedes `HANDOFF_2026-09-02_continue_here.md`** (kept; do not delete — its §5 owner decisions
are unchanged and still the substance of this lane).

**One lane, two bugs, both class fixes SHIPPED, APPROVED and LIVE.** 417's disconfirmations are now
6/6 reached and 5/6 eye-checked. The work since the last handoff was mostly *other people's* bugs
found while verifying this one, plus one live repair.

**Bug files (resolve by SLUG — both numbers are ambiguous):**
- `bugs_open/417_HANDOFF_2026-08-31_planner_logo_exemplar_licenses_a_wordmark_it_never_names_so_the_image_model_invents_a_brand.md`
- `bugs_open/420_HANDOFF_2026-08-31_order_intake_publishes_the_billing_email_as_the_sites_public_contact_and_registers_it_as_a_renderable_claim.md`
  ⚠ **420 is ambiguous** — the other 420 is the negation gate's prose walker. **417 is not.**

**Working docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/` and
`…/bugfix_420_contact_consent/`. The **RUNBOOK is the valuable artefact from this session** — five
new sections, every one of them a thing that cost real time.

---

## 1. FLEET STATE — verified at the artefact this morning, with controls

**Both services rolled 2026-09-03 08:57–08:58Z to `v1.0.1356`.**

| claim | how it was checked | result |
|---|---|---|
| adapter build | `build provenance` log line | `7bf1ff674` |
| **424's guard fix `fcbe6071c` is LIVE** | `git merge-base --is-ancestor fcbe6071c 7bf1ff674` | **YES** |
| ↳ control (must be YES) | `6440ec968` (original matting fix) | YES |
| ↳ control (must be NO) | `20d595bb0` (committed after the stamp) | NO |
| **b2322a203 (magenta contradiction fix) is LIVE** | chassis binary probe | **PRESENT** |
| ↳ positive control | `Render a text-free mark` | PRESENT |
| ↳ removed-string control | `resolveLogoIntent` | absent |
| ↳ impossible control | nonsense string | absent |

**So for the first time BOTH halves are live together: a prompt with no self-contradiction, and a
guard that can actually see a failed matte.**

⚠ **Neither has been exercised.** **Zero** matte runs and **zero** guard refusals since the roll;
no logo asset written since 08:57. The clean test everyone has been waiting for **has not happened
yet** — see §3 item 1.

---

## 2. WHAT THIS SESSION DID

**417 verification COMPLETE.** Disconfirmation A (did the clause reach the generation?) **6/6**;
disconfirmation C (did the model obey it?) **6/6 eye-checked, all clean** — boxingonline, advertise,
designblog, seotools, websitepromotion and **gamedesign.uk** (a woven lattice mark; captured
2026-09-03 09:40Z, minutes before the 424 lane regenerated it). Zero lettering, single composition,
no invented brand on any. **421's two-panel shape did not recur in six.**
⚠ **That last one was nearly lost.** The 424 lane flagged it was about to reset those three sites'
logo items; a regeneration UPSERTs the row and the old artefact is gone. **A peer's fix can destroy
the evidence you still owe — capture it when you hear the plan, not when you get to it.**

**A live repair, at the owner's instruction.** `websitepromotion.co.uk` had a logo asset and a text
header. Chrome re-rendered, then 11 hand-filed assemble-mode `page_rerender` items — **25/25 served
pages now carry `<img class="logo-img">`**, invented-path control 404, logo file 200.

**Three bugs contributed to other lanes, none of them mine to fix:**
- **`bugs_open/424`** — their matte's first production run failed and their fail-closed guard scored
  it **1.000**. `BorderKeyed` counted BFS reachability, not final alpha. They fixed it within the
  hour (`fcbe6071c`), mutation-proved it, and it is now live. I then replayed their **new**
  statistic against four real stored artefacts: **4/4 correct** (websitepromotion 0.9993 PASS;
  designblog / seotools / gamedesign 0.0000 REFUSED).
- **`bugs_open/433`** — the extension question answered: **12 of 12** logo source objects are JPEG
  under `.png` keys, and **their fix candidate 2 would write a confidently wrong value into 910
  rows** (it proposes propagating `uploadImage`'s hard-coded `"image/png"`).
- **`bugs_open/421`** — no recurrence in 5 eye-checks. Still unowned.

**Four LANDMINES entries and three WRONG_CALLS rows**, listed in §5.

---

## 3. WHAT IS LEFT, ORDERED

1. **⭐ THE CLEAN TEST — trigger ONE logo regeneration and look at it.** Both fixes are live and
   unexercised; every other open question here depends on it. The subject should be one of the
   three sites currently serving a broken logo (§4). Expected: a real alpha channel, no coral veil,
   no magenta halo, and — if it fails — a **refusal** in the adapter log rather than a stored
   artefact. **Verify at the bytes, not the status:** RUNBOOK §"Fetch a generated asset's BYTES".
   ⚠ **This dispatches work at another lane's sites — the 424 lane said it is theirs to fire
   post-roll. Coordinate; do not just do it.**
2. ~~Eye-check `gamedesign.uk`~~ **DONE 2026-09-03 09:40Z — clean.** 417 is now 6/6 on both
   disconfirmations.
3. **The fence decision** (417) — still deliberately not taken. 5/5 clean is now the evidence base.
4. **417 and 420 stay OPEN** per the fixed-AND-live bar; 420 also on its §C residual.
5. **Watch the 424 lane's rerun.** They reset `designblog` / `seotools` / `gamedesign`'s
   `needs_imagery:site:-:logo` items to `triaged` (owner-authorised) so the queue retries them under
   the fixed guard. **Their result is 417 evidence too** — but ⚠ **do NOT read `assets.updated_at`
   to decide whether a regeneration happened.** gamedesign's row was bumped 2026-09-03 00:55:58Z
   with **no regeneration behind it** — proven by re-fetching the object and getting a byte-identical
   md5 to the copy pulled ~12 h earlier. Use the **storage key's date directory** (§5 item 4).

---

## 4. LIVE DAMAGE STILL ON THE FLEET `[MEASURED 2026-09-03 09:30Z]`

Three sites serve a logo with **0.0% transparent pixels** — the ground survives as an opaque veil:

| domain | served logo | transparent |
|---|---|---|
| designblog.co.uk | 200, 400×218 | **0.0%** ❌ |
| seotools.co.uk | 200, 400×218 | **0.0%** ❌ |
| gamedesign.uk | 200, 400×400 | **0.0%** ❌ |
| websitepromotion.co.uk | 200, 400×218 | 84.3% ✅ |

All three were produced by the pre-fix guard. **A regeneration now should come out right** — which
is why item 1 above is both the fix and the test.

### 4a. A CANDIDATE BUG, fully evidenced and deliberately NOT filed — pick this up or hand it off

**websitepromotion's "good" logo is still not good, and nothing in the estate can see it.**

- `[MEASURED 2026-09-03]` The mark is pale blue/lavender on a white header: **median contrast
  1.43:1**, darkest 5% **1.70:1**, against the WCAG non-text floor of **3.0:1**. Present,
  transparent, and close to invisible. Plus a **0.69%** magenta halo from despill.
- `[MEASURED 2026-09-03 — upgraded from [UNVERIFIED]]` **No check covers this.**
  `request_render_audit_action.go:4` states its remit as *"text contrast against"* its background;
  `write_render_audit_findings_action.go:12-13` files a `contrast_failure` keyed
  `contrast_failure:<page-path>#<class>` and routes it to **`css-patch-agent`** — i.e. an element's
  CSS, not an image's own luminance. Images are handled **broken / not-broken**
  (`attributed broken images → undeployed_asset`), and `over_image` findings are *"counted,
  deliberately not filed"* (`registry.go:1257`). **So a logo that is perfectly rendered, correctly
  deployed, and illegible passes every gate.**
- **No prior art.** Nothing in `bugs_open/` or `bugs_closed/` covers logo legibility (the grep hits
  are 417 itself, 322's head block, 231's spec defaults, 131's og-card — all different).
- **No owner.** 424's lane has explicitly scoped it out — *"transparency mechanism vs mark
  legibility against arbitrary backgrounds are genuinely different problems"* — and agreed to hand
  it off explicitly rather than absorb it.

**Why it is not filed here:** the owner asked for a handoff, not a new bug, and I said I would check
ownership before filing. The check is done and the answer is "nobody". **Filing it is now a short
job** — the measurement, the three code citations and the negative prior-art search are all above,
which satisfies the 2026-07-31 ruling's *"stated equivalent first-hand verification"* escape hatch
without a `090` run.
⚠ **Do not let this wait on `bugs_open/424` closing.** That is the exact shape of MEMORY's
*"closing a bug does NOT retract the deferrals pointing AT it"* — 20 days lost last time.

---

## 5. TRAPS THIS SESSION HIT — all four are in LANDMINES, read before repeating the work

1. **A hand-built `orchestrate` publish that omits the ORCHESTRATION headers is refused before any
   state row exists.** Receipt says PUBLISHED, exit 0, and **nothing** appears in
   `orchestration_states` — indistinguishable from queue latency, for which the estate's standing
   advice is "wait, do not re-fire". Cost ~50 minutes. Measured: `action`+`message_type`+
   `from_agent_*`+`request_id`+`responses_topic` → no row; **+`client_id`** → still no row;
   **+`message_id`, `orchestration_id`, `orchestration_name`, `step_name`** → row in **~10 s**.
   ⚠ I did not isolate which of the last four is load-bearing `[UNMEASURED]`.
   **The demand control that tells refusal from latency:** are top-level dispatches landing at all?
   (154 were, in the hour mine were absent.) Canonical header set:
   `platform/orchestration/actions/dispatch_verifiers.go:154-168`.
2. **`kubectl` prints UTC; `git log --date=format:` prints the commit's `+01:00` and shows no
   offset.** One hour, in the direction that manufactures a near-miss. ⚠ `TZ=UTC git log
   --date=format:` does **NOT** fix it. Use ancestry (`merge-base`), which has no clock.
3. **The `.png` in an image key is a lie** — 12/12 logo source objects are JPEG, and a JPEG cannot
   hold alpha at all. Read magic bytes before diagnosing anything upstream of the format.
4. **A logo regeneration UPSERTS the asset row**, so `created_at` is the ORIGINAL date and
   `updated_at` is bumped by any write. **The work-item trail is also incomplete** (boxingonline's
   regeneration filed no item). The instrument that caught all of them, sound by construction:
   the **storage key's date directory** (`dynamic_adapter.go:717` mints a fresh uuid under
   `<YYYYMMDD>/`).

**WRONG_CALLS rows added:** the work-item trail one (I recommended an instrument blind to half the
population); the `publish_target` one (I invented a serving rule from one failed probe — **it is
NOT true that only `publish_project` sites are served**); the timezone one.

---

## 6. THE DECISIONS THAT ARE STILL THE OWNER'S — unchanged, see the 09-02 handoff §5

`docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/HANDOFF_2026-09-02_continue_here.md`

1. **RFC_058, the identity model** — the big one; lane recommends option **B** with the subject
   derived by the register.
2. **The 420 §C residual** — does the narrow ruling extend to *derived* contacts? 28 specs carry one.
3. **Ordering cannot reopen until the intake chat asks the contact question** — box-side, no session
   here can close it.
4. **`bugs_open/421` still has no owner.**

Plus, new from this session: **the logo-contrast gap in §4** needs a decision on whether to file.

---

## 7. IF YOU READ ONE THING

**Every real defect this session found was found by opening the artefact, and every one of them had
a green status sitting next to it.** 424's guard reported `border_keyed=1.000` on an unusable image;
a publish reported PUBLISHED on a message that was refused; a header component said `rendered` while
the served page showed something else; and a logo that is genuinely present measures 1.43:1 against
its own background. **The census proves the instruction arrived. Only looking proves it worked.**
