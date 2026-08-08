# POINTER — a live defect in the `empty_section` verifier you built

**2026-07-19, reasoning-dataset thread. This is a pointer, not the substance.**

`VerifyEmptySectionResolved` (`check_empty_sections.go:205`) — the verifier this
workstream built, and still the only one registered on the platform — reports
**success when its target row is absent**:

```go
	if err == sql.ErrNoRows {
		// Component removed — nothing left to be empty.
		return VerifyResult{Resolved: true, Detail: "component no longer exists"}, nil
	}
```

A missing `page_components` row is equally the signature of a rebuild silently
deleting the component. So a content-loss incident is recorded as a *verified
fix* — by the mechanism built to stop `complete` being taken on trust. Found by
the council gate's `bug_historian` seat while reviewing a plan to copy this
branch to two more item types.

**Where the substance lives:**

- **`bugs_open/032`** — the case: evidence, conservative fix (return an error so
  the gate fails open and records "could not verify" instead of asserting
  success), the stronger option, and verification queries.
- **`work_item_completion_integrity/HANDOFF_2026-07-19_verifier_absent_row_defect_and_coverage.md`**
  — the full handoff. That thread owns `CompleteWorkItemAction` and had already
  named this gap in its own PLAN, so it holds the primary.
- **`bugs_open/021` §INSTANCE 2** — the coverage half: `RegisterVerifier` has been
  called once, for ~50 item types.
- **`016b` §9** — the transferable pattern (*a verifier that treats a missing
  target as success cannot distinguish repair from deletion*).

Flagged here because you built the verifier and `empty_section` is the item type
carrying the defect — you may hold context on whether a removed component *should*
read as resolved that neither the council nor we have. If so, say so on `032`;
the conservative fix is deliberately reversible.

---

> ## UPDATE 2026-08-08 — the conservative fix shipped, has now fired twice, and both times the answer was NOT ambiguous
>
> From the `bugfix_201_page_content_writer_dispatch` lane, while measuring `RFC_017`.
>
> `VerifyEmptySectionResolved`'s absent-row branch is live as the error-return described above
> (`check_empty_sections.go:412`), so a missing row no longer reads as success. It has been
> consulted **twice**, and errored **both** times — the only two verifier errors on the platform,
> ever. On both, the page **still declares the slot** (`pages.sections` lists `featured_article` on
> pages `ai-guides` and `insights`, site `1368e337…`), which is exactly the *"page still expects
> this component"* condition this file's §"stronger option" says makes absence **deletion, not
> ambiguity**. The gate fails open, so both items are `complete` with `attempt_count` 0, and both
> pages now serve a deployed 334-byte empty shell in slot `featured-content`.
>
> **So the stronger option you were left to weigh is no longer hypothetical — it is correct on 2 of
> 2 observed cases, and it is the cheapest fix for them** (per-verifier, no shared-struct change).
> No re-detection has followed in five days, although the detector's predicate matches both
> components right now — so the two-strike backstop is not covering this.
>
> Substance and caveats (`n=11` consultations, `result` overwritten so 2 is a floor):
> `architecture_review/RFC_017_verifier_registry_fails_open_on_error.md` § "The missing number —
> MEASURED 2026-08-08", and the primary is still
> `work_item_completion_integrity/HANDOFF_2026-08-08_fail_open_measured_and_it_landed_on_the_deletion_horn.md`.
> Nothing has been changed in your code; this is a report.

> ### FOLLOW-UP, same day — the owner flipped the policy, so your verifier's error branch now BLOCKS
>
> `RFC_017` was decided hours after the update above: **verifier errors fail CLOSED by default**
> (built 2026-08-08, inert until the next chassis roll; council corr
> `a104d454-a4ff-4c95-a578-9a7e48c95100`, register entry `WII-011`).
>
> **What that does to `VerifyEmptySectionResolved` specifically.** Its absent-row branch —
> `check_empty_sections.go:412`, written deliberately as an error *because* the gate failed open —
> now stops the completion instead of annotating it. The item returns to the queue, the handler
> **rebuilds the page again, up to `max_attempts` (3)**, and then lands in `failed`. For a case the
> verifier structurally cannot answer, that is three wasted rebuilds before a human sees it. Nothing
> about your code changed; what changed is what an error MEANS.
>
> **So the "stronger option" is now the cheap fix as well as the honest one.** Ask whether the page
> still declares the slot (`pages.sections`) and return `Resolved:false` when it does: correct on the
> 2 observed cases, and it converts three futile rebuilds into one true verdict. If instead you think
> `empty_section` genuinely wants the old behaviour, the opt-in exists and is one line —
> `RegisterVerifierWithPolicy("empty_section", VerifyEmptySectionResolved,
> VerifierPolicy{FailOpenOnError: true})` — but it now has to be argued at the registration site,
> where a reviewer can see it, which is the whole point of the ruling.
