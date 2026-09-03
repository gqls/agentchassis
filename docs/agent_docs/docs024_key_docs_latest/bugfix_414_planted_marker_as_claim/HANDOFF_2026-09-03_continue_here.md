# HANDOFF 2026-09-03 — `bugfix_414_planted_marker_as_claim`

> ## ⚠ SUPERSEDED — read `HANDOFF_2026-09-03b_continue_here.md` instead
>
> Written 10:34 BST 2026-09-03 and **stale in three rows within the hour** (the demand control was
> already planted by its owner; `farmerinsurance`'s gap had been discharged the previous evening;
> and "the one genuine gap" was one of thirteen). Its own §3/§4 carry those corrections inline, and
> NOTES §(q) records the staleness as a finding. **This file is kept for the trajectory, not for
> state** — everything current, plus `loancash.co.uk`'s register, the 13:28 build, the vetcomparison
> question and the three outstanding owner decisions, is in the `-03b` handoff.

**Supersedes** `HANDOFF_2026-09-02_continue_here.md` (which supersedes `HANDOFF_2026-08-31…`).
Read this one; the older two are kept for the trajectory, not for state.

---

## 0. THE ONE-LINE STATE

**This lane is CLOSEABLE, and there are NO owner decisions outstanding.** The bug is closed, its
framework fix is live and verified at the artefact, and RFC_060 — which this lane spun out — was
**fully ruled by the owner on 2026-09-03**. Everything still moving is a BUILD owned by the
**claims-verification lane**, not a decision and not this lane's work.

The only thing anyone should still *do* on this lane's account is one verification, described in
§3, which belongs to the lane that owns the code.

---

## 1. WHAT THIS LANE WAS, AND WHAT IT DELIVERED

**The bug (`bugs_closed/414`).** A 2026-08-02 shadow experiment planted a tripwire in
`lendzy.co.uk`'s `site_specs.content_direction`: *"include the exact phrase: checked against the FCA
handbook, rule by rule."* The content writer obeyed. The sentence was then **served** on a finance
site as an unverifiable claim of regulatory diligence — and the estate's own audit fleet read it
back off the page and filed a `content_rewrite` canonising it as *"the site's core differentiator"*.
A tripwire did not merely leak; the improvement machinery adopted it as the site's identity.

**Delivered, all live:**

- **The claim is gone** from `content_data`, from `rendered_html`, and from both served bodies —
  verified by curl with an invented-URL 404 control (a parked domain 200s every path).
- **The spec sources are clean fleet-wide.** The original strip missed a *paraphrase* the
  `domain-strategist` had already copied into the `strategy` aspect — the bug file's claim that
  *"regeneration can no longer re-plant the phrase"* was **false when written**, and that
  refutation is the lane's most transferable finding.
- **Three framework changes shipped** (`v1.0.1349` / `v1.0.1354`), so the class is caught rather
  than the instance: the completeness-of-verification patterns widened in `claims_global.go`
  (refusing set, blocker); a documentary-diligence pattern added to `claims_practice.go`
  (warning-only, per the owner's 2026-08-24 ruling that the practice family stays out of the
  refusing union); and a **spec-side detector** (`cmd/brief-negation-check/specclaims.go`,
  `spec_supplies_claim`, register **CLM-030**) so a spec can no longer lawfully order a page to say
  what the page gate would refuse.
- **A regulatory-citation exemption** (`fad209b92`, council `1dd3d298`) so a site quoting its own
  regulator is not convicted for it — deliberately case-sensitive, requiring the code to be
  immediately followed by a digit, two-letter codes excluded, so a bare `FCA` never exempts.

---

## 2. WHAT HAPPENED SINCE THE LAST HANDOFF (2026-09-02 evening → 2026-09-03)

### 2a. A landmine was telling site authors the opposite of the truth — corrected, with a census

Raised by the `loancalculator_couk` lane in passing. The `banned_claims` escaping entry ended
*"A pattern that fails to compile is caught. A pattern that compiles WRONG is not."*

**That is true only of the FLEET-WIDE set**, which is authored in Go and pinned by
`TestEveryGlobalPatternIsAValidRegex` (`claims_global_test.go:376`) — **a CI test, which by
construction cannot see a pattern arriving as DATA.** A per-site pattern from `evidence_base` JSON
hits the identical silent fallback at `claims.go:348`: no logger, no error path; the admin door
counts `banned_claims` and guards against emptying the set but **never compiles one**; a migration
cannot compile a regex Postgres never parses. So **both** halves are silent for exactly the
population that entry footprints. Corrected in `4f1ca1384`.

`[MEASURED 2026-09-02]` **239 live per-site patterns across 19 sites: 0 non-compiling, 0
doubled-backslash**, with two deliberately broken controls firing in the same run. A clean baseline,
established the day five finance sites gained hand-authored sets — **and stale by ADDITION on the
next seed**, which the entry now says.

### 2b. My remedy was prose-as-control, argued down, and the fix got built the same evening

I wrote the remedy as *recorded practice*, reasoning that a clean census does not justify new
platform surface. The `loancalculator_couk` lane argued that down and was right: practice-as-remedy
is what the owner ruled on in 2026-08-02 §2, and it is `RFC_006`'s exact shape, where a CI-time
check **structurally cannot** gate live config so the ruled fix was a **daily runtime** one. My own
entry had just established that the only guard is a CI test.

The premise also never applied — I declined "new surface" without checking whether any was needed.
`evidence-freshness` already runs daily over **every** site with a current `evidence_base`. The
`claims-verification` lane built the check into that existing loop the same evening (`e5b1a0f01`),
**council-APPROVED** (`bc3697a5`, 3 advisory objections, none high-severity).

Both wrong calls are in `WRONG_CALLS.md` (`01276a88e`, `da9d68848`). The second is the sharper one:
**I verified the claim that code WORKED and repeated the claim that code was BROKEN**, in the same
hour, about the same commit — relaying a council objection into `LANDMINES` without opening the
function. The objection was against the submission's *sketch*; the code was right throughout.

### 2c. RFC_060 — FULLY RULED, all seven questions

- **Q1–Q4 (2026-09-02):** registers required on finance sites; build order (ii)→(iii)→(i); the tier
  is a **record**, not a flag; and the axis is **semantic, not sector** — the three-rung posture
  ladder `standard` / `sourced` / `relied_upon`, approved as proposed.
- **Q5, Q6, Q7 (2026-09-03), §3f:** Q6 and Q7 **build as proposed**. **Q5's answer INVERTED** — I
  had advised holding the sector PRESETS back on this RFC's own instinct not to design ahead of a
  second consumer; the owner supplied the fact that instinct turned on: *"I will be extending to vet
  and legal quite soon."* So presets are IN, and **`legal` was not in the original sketch**
  (veterinary and medical were).

---

## 3. THE ONE OUTSTANDING VERIFICATION — and it is NOT this lane's to run

**Verified live today, 2026-09-03**, at the artefact and not at the commit:

- New replicaset `75b987cbd7`, pods ~17 min old at 09:21.
- **Binary probe of `/proc/1/exe` with BOTH controls**: target `invalid_banned_claim_pattern` → **6
  matches**; a must-be-absent control → **0, exit 1**; a must-be-present old symbol
  (`stale_evidence`) → **6**. So the detector **is in the running binary**.
- `evidence-freshness` **ran at 09:10:23 and completed**, under those new pods (started ~09:04).
- `site_work_items` where `item_type='invalid_banned_claim_pattern'` → **0**.

⚠ **AND THAT ZERO CANNOT BE READ AS SUCCESS.** Both result fields are declared `omitempty`
(`refresh_evidence_base_action.go:216` and `:221`), so **a clean result serialises to NOTHING**.
Confirmed: of 23 evidence runs since 09:00, **0 mention the field** — a figure that comes out
identical whether the code ran and found nothing or never executed at all. This is the
*post-fix-zero-with-no-demand-control* shape, and here the blindness is **mechanical and citable**,
not a matter of judgement.

**So the outstanding item is a DEMAND CONTROL:** plant a deliberately broken pattern on a scratch
site, confirm the next pass **files** an item, then remove it. Until that is done, "the detector is
live" is proven and "the detector works" is not. **Owner: the `claims-verification` lane** — they
own the code, the council round and the seam; they have been told, including the `omitempty`
reasoning. **Do not do this from this lane.**

**UPDATE, same day — the two obstacles to running it are both gone.**

1. *"There is no scratch site"* is **false**, and both that lane and I got it wrong the same way
   (searching by DOMAIN NAME). Non-production is marked by **`sites.status`**: `deployed` 39,
   `pool` 17, **`test` 3**, `system` 1. The three test sites — `buytoletcalculator.uk`,
   `copyonline.co.uk`, `indoorplanters.co.uk` — carry ordinary client-looking domains and have
   **ZERO pages**, so nothing is served and there is no register to disturb. Now landmined
   (`71b85fcc2`).
2. *"It needs the daily tick"* is also false. `resolveEvidenceSites` (`:281`) takes an optional
   single `site_id`, and its fleet-wide query (`:290`) has **no status predicate** — so a `test`
   site with a register is swept like any other, and can be targeted directly.

**So the reversible rehearsal is:** write an `evidence_base` on one test site containing a single
deliberately non-compiling pattern → dispatch `refresh_evidence_base` at that one `site_id` →
assert one `invalid_banned_claim_pattern` row appears → supersede the spec → cancel the row.

⚠ **This exercises the arm the lane's own fix cannot reach.** Their follow-up (`996b40542`) adds an
always-fired Info log with `patterns_checked`, which proves `:423` executes and — because a
non-zero count means it read real data — is stronger than a bare "I ran" line. But it is **Go, so
inert until the next roll**, and the **write** path (`:700` →
`createInvalidBannedClaimPatternItems`) only fires on a non-empty finding, so it stays unproven
outside mocks either way. A detector that detects and does not file looks identical to a clean
fleet.

**UPDATE 2026-09-03 09:49 UTC — the probe is PLANTED by its owner, and the "inert until the roll"
clause above is now MEASURED rather than inferred.**

The `claims-verification` lane planted it at **09:34:48 UTC**, four minutes after this lane's
landmine (`71b85fcc2`) named the test sites: `buytoletcalculator.uk`
(`dc7a8ebf-9c23-45e7-970e-32147615bb12`), spec row `623c1de8`, `created_by='claims_verification_probe'`,
one `banned_claims` entry with pattern `guaranteed(`. **The assertion has not yet succeeded** — at
09:49 UTC `site_work_items` for `invalid_banned_claim_pattern` is still **0** fleet-wide and no
`orchestration_states` row since 09:25 names that site. So: planted, not yet dispatched. **Do not
plant a second one.**

`[MEASURED 2026-09-03 09:47 UTC]` The `patterns_checked` Info line is **absent from BOTH replicas**
of replicaset `75b987cbd7` — probed at `/proc/1/exe`, `0` and exit 1 on each, against
`invalid_banned_claim_pattern` **6/6** present, control `stale_evidence` **6/6** present, control
`zzz_not_a_real_symbol_qx7` **0/0** absent. The pods started **08:57:46 / 08:58:07 UTC** and
`996b40542` was committed **09:29:46 UTC**, so the absence is arithmetic as well as measurement.
**The consequence is a live trap on one branch:** if the dispatched pass files nothing, grepping for
`patterns_checked` returns silence that is the *un-deployed line*, not a non-executing check.
Relayed to the owning lane in
`docs/agent_docs/docs024_key_docs_latest/claims_verification/CONTRIB_2026-09-03_from_414_your_demand_control_is_planted_and_the_log_line_you_will_reach_for_is_not_deployed.md`.

One further observation, since the probe **created** that site's register rather than editing one:
`resolveEvidenceSites`' fleet query (`:290`) carries **no `sites.status` predicate**, so the probe
site is now in the daily tick as well as directly dispatchable. Reverting the spec is what takes it
back out.

---

## 4. WHAT IS LEFT ON THIS LANE — nothing, and here is the accounting

| item | state | owner |
|---|---|---|
| `bugs_closed/414` — the bug | **CLOSED**, fix live, verified at the artefact | — |
| Framework fix (3 changes + citation exemption) | **LIVE**, council-approved | — |
| RFC_060 — all 7 questions | **FULLY RULED** 2026-09-02/03 | owner: done |
| RFC_060 Q5/Q6/Q7 builds | not written; none rode today's roll | **claims-verification** |
| Q7 `banned_claims` half | **BUILT, APPROVED, LIVE** — demand control **PLANTED 09:34 UTC, not yet dispatched** (§3) | **claims-verification** |
| `loancash.co.uk` register | **the one genuine gap, re-confirmed 09:49 UTC 2026-09-03**: no `evidence_base` row at all (not an empty one) beside 14 other current specs, on a **deployed** finance site serving **30 pages**. RFC_060 Q1 makes a register required here; owner informed; **still unowned**. **It is the LAST of RFC_060 §1d's five — the other four are done** (`lendzy` landed, `farmerinsurance` 09-02 18:34), so §3c track 1 is now one site. Wider frame recorded in the RFC: **13 of 39 `deployed` sites hold no register**, and **nothing can ever raise the absence** — the sweep's target list is built from the sites that HAVE one | unowned |
| ~~`farmerinsurance.uk` — 0 `banned_claims`~~ | **DISCHARGED 2026-09-02 18:34 UTC** by migration 713 (lendzy relay), *before* this table was written; today's 09:11 refresher carried 7 facts + 5 patterns forward unchanged. Spec history, not the current row, is what shows this — the current row alone reads as "the daily sweep fixed it", which is false | — |
| `lendzy.co.uk` migration 695 | written + council-reviewed, blocked on rolls | lendzy lane |

**No owner decisions outstanding.** RFC_060 is fully ruled; nothing here waits on a person.

---

## 5. THE FIVE THINGS WORTH CARRYING OUT OF THIS LANE

1. **Stripping the origin of a planted instruction does not retract it.** An agent may already have
   copied it into another aspect, in its own words. Sweep **every** `is_current` aspect, not the one
   you edited. That single query is in the RUNBOOK and in `016b` §9.
2. **A CI test cannot guard data.** Wherever a Go-authored set and a data-authored set share a code
   path, the guard on the first is not a guard on the second — and the difference is invisible.
3. **`omitempty` erases the clean case**, which is exactly the case a demand control needs to see.
   A field that only appears when something is wrong cannot prove the check ran.
4. **The failure mode inverts between a ban list and an exemption list.** A broken banned-claim
   pattern leaves a guard inert; a wrong `citation_codes` entry makes the scan **blind**. Both fail
   open and both are silent, but the second disarms *detection* on the sites the design exists to
   protect. **Probe an exemption list in BOTH directions.**
5. **Verification gets applied asymmetrically.** Claims that something WORKS get checked; claims
   that something is BROKEN get repeated, because passing on a defect feels like diligence. A
   reviewer's objection wears the form of a checked finding while being an unverified reading.

---

## 6. HOW TO RE-DERIVE EVERY FIGURE ABOVE

```bash
# the pattern census (needs a stdlib-only Go file; controls are mandatory)
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -At -F $'\t' -c "
  SELECT s.domain, bc->>'pattern' FROM sites s
  JOIN site_specs eb ON eb.site_id=s.id AND eb.aspect='evidence_base' AND eb.is_current
  CROSS JOIN LATERAL jsonb_array_elements(eb.data->'banned_claims') bc;"

# is the detector in the RUNNING binary? (always run both controls)
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers | awk 'NR==1{print $1}')
kubectl -n ai-persona-system exec "$POD" -- grep -ac "invalid_banned_claim_pattern" /proc/1/exe  # target
kubectl -n ai-persona-system exec "$POD" -- grep -ac "zzz_not_a_real_symbol_qx7"      /proc/1/exe  # must be 0
kubectl -n ai-persona-system exec "$POD" -- grep -ac "stale_evidence"                 /proc/1/exe  # must be >0

# did the sweep run, and did it file anything?
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT now()::timestamp(0), last_completed_at::timestamp(0) FROM scheduled_tasks WHERE name='evidence-freshness';
  SELECT count(*) FROM site_work_items WHERE item_type='invalid_banned_claim_pattern';"
```

⚠ **Never `strings`** (absent from the debian-slim images) and never a *discovery* grep for "some
40-hex string" — it matches Go's internal digit table and returns the same wrong answer on every
service.
