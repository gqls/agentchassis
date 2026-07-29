# RFC 003 — The build gate may now refuse a page on a site that never opted in

**Status: OPEN — raised 2026-07-29** by session "bugsearch 6" (`bugs_closed/104`), under
**owner ruling 2026-07-29 §1**: an addition to a shared mechanism needs an RFC when it
changes what that mechanism **GUARANTEES**. This one does, and I should have filed it when
the council's `architecture` seat said so instead of calling it "an owner call on venue".

> **RETROSPECTIVE, like RFC 002, and for the same structural reason.** The change is live
> (chassis v1.0.1196, pod-verified on both replicas) and council-APPROVED on round 3
> (`Council-Reviewed: 899ed92e-1bf7-4707-96d8-24f102aa14fa`). Per ruling §2, review here is
> **after the fact by design** — HEAD is shared, `make build-*` builds from committed HEAD,
> and another session's roll shipped it. I claim no ordering constraint and do not pretend I
> could have waited. This is the `bugs_closed/124` shape: the code stays, the precedent
> gets fixed, and the useful product is the rule for the next one.

---

## 1. Problem + evidence

### 1.1 What changed about the guarantee

`validate_page_content` check 8 previously read, in its own file header: *"Sites without an
evidence_base skip both silently."* A site that had never opted in could not fail a build on
a claims finding — the mechanism's guarantee to such a site was **"I will not touch you"**.

After `bugs_closed/104` it can. Nine fleet-wide banned-claim patterns
(`datahelpers/claims_global.go`) are joined at scan time by `ScanAllBannedClaims`, which is
**nil-safe**: a site with no `evidence_base` row is scanned, and a match is severity
**blocker**, which fails the page build.

That is the same shape as RFC 002's trigger — *"made the Tier 2 evaluator able to **refute**
where its stated rule had been confirm-never-refute"*. Here: **a gate that skipped silently
can now refuse.** Not additive-and-inert; additive-and-guarantee-changing.

### 1.2 Why the change was made (the defect it closes)

`banned_claims` was per-site only, so a fabrication pattern learned on one site could not
reach another and every new site was born unarmed. 8 of 15 sites carried no pattern,
including **vetcomparison.uk** (which had published fabricated prices for 3,124 named real
vet practices) and **idea.uk** (the only site taking money). The deferral behind it was
filed with a numeric trigger — *"per-site only until two sites have evidence bases"* — which
fired months earlier at 9 and had no watcher. Owner ruled it as oufe decision **O11** on
2026-07-28, option "narrowed set, all 15 sites", with the dry-run table in front of them.

### 1.3 The process finding, which is the uncomfortable part

The council's `architecture` seat raised exactly this in round 3: *"New exported symbols
(`ScanAllBannedClaims`, `GlobalBannedClaimCount`) create a fleet-wide, blocker-severity
shared mechanism across 3 production call sites without a filed RFC. Substance of an RFC
(blast radius, rollback, scope decision) exists…"* — and `guardian` said the same at medium.

**I answered that it was a decision about venue for the owner, and did not file.** Under the
ruling published the same day the test is not "is the vocabulary shared" but "does the
guarantee change", and by that test the seats were right and my answer was wrong. The
substance did exist; what was missing was the artefact and, per ruling §3, **telling the
consumers**.

## 2. Design — what is actually in place

- `globalBannedClaims()` — nine patterns, held **outside** any parsed `EvidenceBase`.
- `ScanAllBannedClaims(blocks, eb)` — fleet-wide set ∪ per-site set, deduped by pattern,
  nil-safe receiver. Wired at both enforcement surfaces (CLM-004) so they cannot drift.
- `check_claims_fleet_wide` — reversal lever, **default TRUE**, `validate_page_content` step
  config. DB config is live immediately, so the behaviour can be withdrawn fleet-wide in
  seconds without a build. Off restores the pre-104 scan exactly. Pulling it logs at **Warn**
  naming the site, because the compliance seat pointed out a silent disarm path would unarm
  vetcomparison.uk or idea.uk with nothing saying so.
- **Default ON is deliberate and now has explicit backing**: ruling §2 states a default-OFF
  switch will **not** be required, its cost being "a mechanism rotting unexercised".

Two constraints shaped it and are worth carrying forward, because both look like the obvious
implementation and both are wrong:

1. **NOT unioned into `ParseEvidenceBase`**, though `voicetells.go` does exactly that for the
   voice layer. `EvidenceBase` is marshalled **back** to `site_specs` by
   `refresh_evidence_base_action.go` and `evidence_citations.go`, so seeding
   `eb.BannedClaims` at parse time would persist the fleet-wide set into every site's stored
   register through write paths that never intended to touch it.
2. **`ParseEvidenceBase`'s nil contract is unchanged.** Only the banned half went fleet-wide;
   the numeric scan stays strictly opt-in, because its false-positive rate is why it is never
   a blocker.

## 3. Alternatives considered

| option | why not |
|---|---|
| **Per-site only, arm the 8 by hand** (status quo ante) | It is the state that produced the bug; two sites had already been armed one at a time by somebody remembering, and site sixteen is born unarmed. |
| **Fleet set, armed sites only** (gated on `eb != nil`) | Measured: reaches 9 of 15 sites and **misses vetcomparison.uk and idea.uk**, which have no row. Same measured cost as the full option, so the containment bought nothing. |
| **Ship all ten candidate patterns** | Dry-run: 4 of 7 findings were FALSE POSITIVES on *negated* honest sentences ("has **not** been independently verified"), at blocker severity. One pattern caused all four; it is excluded pending a code-level negation guard (RE2 has no lookbehind; no prior art in the estate). |
| **Default-OFF switch** | Explicitly not required by ruling §2, and it would leave 104 live while the mechanism rotted unexercised. |

## 4. Blast radius, named — derived mechanically

**Binaries linking `orchestration/actions`** (`go list -deps` per `cmd/` target):
`agent-chassis`, `core-manager`, `config-key-audit`, `test-spawning`, `workflow-monitor` —
five link it; **one, `agent-chassis`, executes the gate.**

**Consumers — every active agent definition with a `validate_page_content` step, enumerated
rather than grepped:**

| consumer | step | `check_claims` | affected? |
|---|---|---|---|
| `page-build-handler` | `validate_content` | unset (ON) | **YES** |
| `content-reviewer` | `validate_content` | unset (ON) | **YES** |
| `tool-recreation-handler` | `validate_tool` | unset (ON) | **YES** |
| `report-builder` | `validate_page` | **false** | no — already opted out entirely |

**Three of four.** No consumer sets `check_claims_fleet_wide`, so all three run with the
fleet-wide set ON by default.

**The newly-refutable population — the 6 live sites with no `evidence_base` row**, i.e. the
sites whose builds could not previously fail on a claims finding and now can:

```
dartsonline.com · gaswholesalers.com · idea.uk
system.internal · vetcomparison.uk · webdesign.co.uk
```

Two of those six — **idea.uk** and **vetcomparison.uk** — are the estate's most exposed
sites, which is the argument *for* the change and also why the guarantee change deserves
naming rather than a footnote.

**What the measurement does NOT cover** (the guardian's standing point, and ruling §3's):
that the three affected consumers' owners would have agreed. Zero findings is a fact about
today's copy, not consent. Discharged below.

## 5. Acceptance evidence

- Shipped: chassis **v1.0.1196**, pod-verified on **both** replicas with markers this change
  created (`completeness-of-exclusion` 3, `verification-of-everything` 1) plus a positive and
  a **negative** control.
- The shipped set over the stored `rendered_html` of the **entire** enforcement surface —
  measured with `sites.status` dropped as a scoping variable, after the council's
  `debug_historian` correctly objected that I had filtered by a column the gate never reads:
  **908 components / 14 sites, all `deployed`; 0 findings.** Pool and system sites hold zero
  stored components.
- Positive control same run, with **no register supplied**: 6 of 6 overclaim shapes blocked.
- 13 tests: 8 in `datahelpers/claims_global_test.go` (including the four real negated
  sentences as regression fixtures, and two guarding the silent regex fallback), 5 in
  `actions/validate_page_content_fleetwide_claims_test.go` driving the real action against a
  DB returning no `evidence_base` row.
- Council APPROVED round 3, 12 reviewers, 5 abstained, **0 unreadable**, no high-severity.

## 6. Rollback plan

`check_claims_fleet_wide: false` on the step config of the three affected consumers. DB
config, live immediately, no image and no roll. Restores the pre-104 scan exactly. Pulling it
emits a Warn per build naming the site, so a withdrawn gate is visible rather than silent.

Reverting the code is the second-line option and is not required for behaviour withdrawal.

## 7. Telling the consumers (ruling §3)

Named above; told by a dated cross-reference added in the same commit as this RFC to each
affected consumer's own register category, because that is where a session working that
pipeline looks:

- `register/page-build-pipeline.md` → `page-build-handler`
- `register/content-quality.md` → `content-reviewer`
- `register/tool-lifecycle.md` → `tool-recreation-handler`

Each says what changed about **their** guarantee — *your validate step can now fail a build
on a site with no evidence_base* — and points at CLM-015 and the lever, rather than listing
my new symbols.

## 8. The questions for the owner, stated plainly

1. **Is the guarantee change acceptable as it stands?** A site that never opted in can now
   have a page build refused. You ruled O11 knowing it applied to all 15 sites; this asks the
   narrower question the seats raised — whether "unarmed sites are now refusable" is the
   intended reading, or whether it should have been confined to armed sites (which would have
   left idea.uk and vetcomparison.uk uncovered, the outcome O11 rejected).
2. **Should the 6 register-less sites get their own `evidence_base` rows anyway?** The
   fleet-wide set is a floor, not a substitute for a per-site audit, and the two exposed
   sites currently have only the floor.
3. **Does `image_url_404`-style renaming apply here too?** `ScanAllBannedClaims` is honest,
   but the *item* the gate emits is still `banned_claim` with a reason string that now
   sometimes means "no site may say this" rather than "this site was audited". Cheap to
   leave; cheaper to name correctly now than after a third consumer arrives.

I have deliberately not acted on any of the three. §1.3 is the reason: the last time I
judged a scope question myself, the answer published the same day was that I had got it
wrong.

---

## 9. UPDATE 2026-07-29 (later the same day) — two figures in §4/§5 are superseded, and question 3 answered itself

Filed by the same session. Neither change alters the RFC's question — whether the
guarantee change is acceptable — but both make its evidence honest.

1. **The enforcement surface is 919 components / 14 sites, not 908** (§5). 908 was
   right on 2026-07-28; the surface grew by 11. Re-derived with `sites.status` grouped,
   not filtered: `deployed|919|14`.
2. **"0 findings" (§5) is spent, and it was an ARTEFACT.** The tenth fleet-wide
   pattern, excluded on 07-28 for false-positiving on negated sentences, is now armed
   behind a clause-local negation guard (CLM-017, commit `116fdffd8`, council
   `8a41e1a5-e670-4e50-a875-f8418ee15738`). Armed, the same corpus yields **2 findings
   + 4 suppressed**. Both findings are robot-hands.com asserting its spec data is
   "independently verified" (`bugs_open/147`); **those two components will not rebuild
   until the copy changes.** Deployed pages keep serving.

   **This strengthens §8 question 1 rather than changing it.** The guarantee change is
   no longer hypothetical: a real page build will now be refused. It is refused for
   making a false claim, which is the design — but the owner should answer question 1
   knowing the gate has teeth in practice and not only in principle.
3. **§8 question 3 (renaming the emitted item/reason) is answered: no rename needed.**
   The new pattern ships with a reason string that says what it means without
   pretending the site was audited — "external-verification claim: asserts our content
   was checked by someone outside this system. Nothing does that." That is what a
   blocked page author reads. The generic `banned_claim` item type stays; the reason
   carries the meaning. Recording it as closed so it does not sit as an open question
   nobody will act on.

**Not RFC-scope in itself, and here is the test applied rather than asserted.** Under
owner ruling 2026-07-29 §1, an addition needs an RFC when it changes what the shared
mechanism GUARANTEES. Adding a pattern within an already-refutable gate does not — the
guarantee change is the one this RFC already documents — and the guard **reduces** false
refusals. It went through the normal council gate. The one part a reviewer should weigh
as shared-mechanism: the guard applies to **per-site registers too**, so nine sites'
human-audited patterns are now negation-aware without their owners asking. Measured
blast radius: **zero suppressions on any per-site pattern** across all 919 components.
That is a fact about today's copy, not consent, and it is named as the main risk in the
submission for exactly that reason.
