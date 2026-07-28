# 105 — `EvidenceFact.Kind` is declared, documented, and read nowhere (CLOSED — readers shipped and LIVE on v1.0.1180)

**Filed** 2026-07-27 from the oufe.com workstream.
**Severity** low as a defect, higher as a trap — a field that looks like a
contract and silently governs nothing will eventually be written by someone who
expects behaviour from it.
**Status** **CLOSED 2026-07-28 — candidate 1 done, council-APPROVED, both rounds LIVE on
v1.0.1180.** Candidate 2 untouched by design and belongs elsewhere.

> ## ✅ CLOSED — verified in the running pod with a removed-symbol control
>
> Chassis **v1.0.1180** (pod up 2026-07-27T22:06:22Z). The verification is unusually
> strong because round 2 both ADDED and REMOVED a symbol, so it has a built-in negative
> control rather than a borrowed one:
>
> ```
> ADDED    "resolved through an alias"        -> 1   (AliasedKinds, round 2)
> ADDED    "not in the documented vocabulary" -> 1   (UnrecognisedKinds, round 1)
> REMOVED  "IsLiveVerifiable"                 -> 0   (deleted in round 2 — must be absent)
> KEPT     "CanonicalKind"                    -> 5   (positive control)
> ```
>
> A grep that finds what was added AND fails to find what was deleted cannot be vacuous.
>
> **The defect this file names is fixed.** `Kind` had no readers; it now has three
> (`CanonicalKind`, `KindIsRecognised`, `AliasedKinds`), all consumed by
> `validate_page_content`'s register load, which is on the deploy path.
>
> ### Deliberately NOT done, and not a residual of this bug
>
> **Candidate 2** (distinct V4 treatment for `capability`/`attestation`) is a new
> capability, not this defect. This file says do not bundle it, and V4 is the only half of
> the layer with a live daily cadence. It also needs a map-side sibling, because V4
> iterates raw `map[string]interface{}` facts rather than typed `EvidenceFact` — that cost
> belongs with candidate 2.
>
> ### One open question, for whoever owns the four registers
>
> **Is `count` meant to be a distinct kind?** It is used by 18 facts across 4 sites
> (ai-agent-orchestration.com, gamesdesign.co.uk, robot-hands.com, vonc.com) and appears in
> no documentation. This change treats it as an alias of `metric`, which is an
> interpretive judgement the council's guardian seat flagged. It no longer resolves
> silently — `AliasedKinds` announces it at every register load — so if the answer is "no,
> they are different", the signal is there to contradict.

> ## STATUS 2026-07-27 (bugs thread)
>
> **Council APPROVED** (correlation `47bde1b6-7abf-4a82-b44d-bb1e071a9948`,
> `complete_approved` 20:34:44, 2 medium advisory objections, no veto). **7 seats
> abstained**, so this was judged by a small panel — both objections came from the two
> seats that always run.
>
> **Live:** `606f485f7` shipped in chassis v1.0.1179. Verified in the running pod by a
> marker that read **0** on v1.0.1177 and **1** on v1.0.1179 — a real before/after rather
> than an inference from build timing:
> `strings /app/agent-chassis | grep -c "not in the documented vocabulary"`.
>
> ### The two measurements that changed the fix from what this file proposed
>
> 1. **The live vocabulary is not the documented one.** metric 46, **count 18**, entity
>    11, capability 9, attestation 4. `count` is used by **four sites** and appears in no
>    spec, so candidate 1 as written — *reject unknown kinds* — would have failed four
>    registers closed. It is handled as an alias.
> 2. **`EvidenceBase` is marshalled BACK to `site_specs`** (`refresh_evidence_base_action.go:677`,
>    `evidence_citations.go:350`). Normalising `Kind` at parse time would have silently
>    rewritten 18 stored facts from `count` to `metric` through a write path that never
>    intended to touch them. So **nothing is normalised in place** — the accessors are
>    read-side, and a test pins that parse plus a marshal round-trip leaves the stored
>    value byte-identical.
>
> ### Round 2 (`b18dd564d`) — both council objections acted on, NOT yet live
>
> - **`IsLiveVerifiable` REMOVED.** It had no production caller, only tests. The
>   editquality seat pointed out that this reproduces, for a second symbol, the exact
>   defect being fixed. I had flagged it in my own risk section and shipped it anyway,
>   which is worse than not noticing. It belongs with candidate 2's consumer.
> - **`AliasedKinds` ADDED.** The guardian seat objected that `count`→`metric` is an
>   interpretive judgement over 18 facts on 4 sites and that, resolved silently, no signal
>   would surface if the guess is wrong. It now announces itself where the register loads.
>   **If whoever wrote those four registers meant `count` as a distinct kind, say so —
>   this is the open question the council left.**
>
> ### Owed
>
> - A roll to make round 2 live, then a pod-grep for `"resolved through an alias"`.
> - **The trailer gap:** `606f485f7` carries the fix with NO `Council-Reviewed:` trailer,
>   because it was committed ahead of the verdict to make an imminent chassis build. The
>   `098` coverage report will list it as un-reviewed. It is not — the verdict is
>   `47bde1b6` and the trailer is on the round-2 commit.

## The finding

`platform/orchestration/datahelpers/claims.go:73`:

```go
Kind       string         `json:"kind"`            // metric | capability | entity | attestation
```

The four values are documented again in the spec
(`claims_verification/SPEC_claims_verification.md:104`) and appear in worked
examples and seeds. **Nothing in the platform ever reads it.** Every `.Kind`
reference under `platform/`, `internal/`, `cmd/` belongs to a different struct —
`imageryplan.Row.Kind`, the diagnosis loop's `CodeRequest.Kind`, the imagery style
guide's `Kinds` map. The one place `EvidenceFact` is iterated
(`claims.go:522`) matches on value and context terms and never consults `Kind`.

So a fact declared `kind: "capability"` is treated identically to one declared
`kind: "metric"`, `kind: "banana"`, or with the field absent.

## Why it matters more than an unused field usually would

**It is the slot a whole class of claim needs, and its emptiness is why that class
has no home.**

`EvidenceSource` (`claims.go:60-64`) already models exactly one of
`sql | artifact | attested_by`, and V4 (`refresh_evidence_base_action.go`) already
re-runs sql-sourced facts daily across every site with a register and raises
`stale_evidence` on drift. Put those together and the register already describes:

> a claim, its kind, the mechanism that backs it, when it was last checked, and a
> sweep that re-checks it.

That is precisely what is wanted for a **promise** — "we correct errors when told",
"14-day refund", "every figure is dated" — which is a `capability` claim whose
source is the mechanism that keeps it. The design is present; only the consumer is
missing. On 2026-07-26 this workstream came close to proposing a *new* promise
register because the existing slot was invisible (it does nothing, so nothing in
the code points at it).

`kind: attestation` has the same problem from the other side: the spec
(`SPEC:124-126`) draws a real distinction — sql facts are *live-verifiable and go
stale*, artifact/attested facts are *checked for presence in the register, not
re-proved* — and no code enforces it, because the discriminator is never read.

## Fix candidates, ordered by what closes the door

1. **Validate on parse and reject unknown kinds** (`ParseEvidenceBase`,
   `claims.go:110-130`), defaulting an absent kind to `metric`. Makes the
   documented vocabulary real, and makes a typo fail loudly instead of silently
   meaning nothing. Smallest change that stops the field lying.
2. **Give `capability`/`attestation` distinct treatment in V4**: sql-sourced facts
   re-run as today; attested facts are checked only for presence and for
   `verified_at` age, raising `stale_evidence` when an attestation passes an
   agreed shelf-life. This is where a promise-with-a-mechanism would actually be
   kept honest, and it reuses the existing sweep rather than adding one.
3. **Delete the field.** Legitimate if we decide the distinction is not wanted —
   better an absent contract than a decorative one. Records the decision instead
   of leaving the ambiguity.

Recommend 1 immediately (it is a few lines and removes the trap), then 2 as the
route to promise-keeping, which is a live open decision in
`docs024_key_docs_latest/oufe/DECISIONS_2026-07-26_oufe.md` §O4.

## How to verify a fix

Write a fact with `kind: "nonsense"` into a site's register and load it. Under
candidate 1 the parse must fail loudly rather than accepting it. Then write a
legitimate `kind: "capability"` fact and confirm it is accepted and, under
candidate 2, that V4 treats it differently from a `metric` — **induce the bad
input**, because a register that parses cleanly today proves only that nobody has
typed the wrong thing yet.

## Related

- `bugs_open/104` — the other half of why this register under-delivers: the
  patterns reach 5 of 15 sites.
- `DECISIONS_2026-07-26_oufe.md` §O4 — the promise-keeping decision this unblocks.
- `docs024_key_docs_latest/experience_register/harvest/HARVEST_01_2026-07-26_vonc_provocations.md:69`
  — *"A promise ledger the platform cannot mechanically check is prose."* The
  EXPERIENCE_PLAN has had a promise ledger since 167 and it has never been
  mechanically checked, for the same reason.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — re-checked from the READER side, still zero

Verification sweep, not a fix. This file asserts an absence, and the standing rule is that
"writes the field ≠ reads the field" cuts both ways — an absence claim needs the reader
search, not the writer's declaration. Re-run across every consumer of the register:

```
grep -n "\.Kind" platform/orchestration/datahelpers/claims.go \
                 platform/orchestration/actions/discovery_checks/check_unverified_claims.go \
                 platform/orchestration/actions/validate_page_content*.go \
                 platform/orchestration/actions/refresh_evidence_base_action.go
-- (no matches)
```

**Zero readers in the entire claims path**, including the two consumers that shipped since
this was filed (`check_unverified_claims_stats.go`, the `v1.0.1172` stat audit). Confirmed
against the running image: no Go commits exist after `e96d42226`, which is in `v1.0.1174`,
so the tree and the binary agree.

**Sizing.** Candidate 1 (validate on parse, default absent → `metric`) is genuinely small —
a few lines in `ParseEvidenceBase` — but it is a `platform/` change, so it is council gate +
chassis roll, and it will **reject registers that parse today**. Nine sites now hold a
current `evidence_base` row; audit their `kind` values before making unknown ones fatal, or
a strictness fix becomes a fleet-wide parse failure. That check is a single query and
belongs in the plan, not after it.

**Do not bundle it with candidate 2.** Candidate 2 (distinct V4 treatment for
`capability`/`attestation`) is the one with the design question in it, and V4 —
`evidence-freshness` — is the *only* half of this layer with a live cadence
(`bugs_open/083`: its sibling post-deploy sweep has not run since 2026-05), so it is the one
place a mistake here reaches production daily.
