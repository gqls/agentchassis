# 105 — `EvidenceFact.Kind` is declared, documented, and read nowhere

**Filed** 2026-07-27 from the oufe.com workstream.
**Severity** low as a defect, higher as a trap — a field that looks like a
contract and silently governs nothing will eventually be written by someone who
expects behaviour from it.
**Status** OPEN.

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
