# HANDOFF 2026-09-03 — `bugfix_414_planted_marker_as_claim`

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

---

## 4. WHAT IS LEFT ON THIS LANE — nothing, and here is the accounting

| item | state | owner |
|---|---|---|
| `bugs_closed/414` — the bug | **CLOSED**, fix live, verified at the artefact | — |
| Framework fix (3 changes + citation exemption) | **LIVE**, council-approved | — |
| RFC_060 — all 7 questions | **FULLY RULED** 2026-09-02/03 | owner: done |
| RFC_060 Q5/Q6/Q7 builds | not written; none rode today's roll | **claims-verification** |
| Q7 `banned_claims` half | **BUILT, APPROVED, LIVE** — demand control outstanding (§3) | **claims-verification** |
| `loancash.co.uk` register | the one genuine gap; owner informed; **unowned** | unowned |
| `farmerinsurance.uk` | has a register (7 facts) but **0 `banned_claims`** — the 707 residue | flagged to lendzy relay |
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
