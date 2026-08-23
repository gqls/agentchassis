# RFC 047 — a shared matcher may REFUSE to answer: `BestLabelMatch` and the alphabetical tie

**Status: DRAFT — raised 2026-08-23 by the `bugfix_308_cta_destination_provenance` lane.
RETROSPECTIVE, in the shape RFC_002 established: the change is committed (`7f85aa814`) and
inert until the next fleet roll. It is here because the commit hook's architecture signal fired
on an exported-symbol change and because the 2026-07-29 owner ruling §1 makes the trigger exact —
this changes what a shared mechanism GUARANTEES, which is the one case that needs an RFC even
when the addition is small.**

Companion evidence, not restated here:
`docs024_key_docs_latest/bugfix_308_cta_destination_provenance/CALIBRATION_2026-08-23_phase_b_widening_report.md`.
The bug is `bugs_open/308`; the register entries are LNK-036 and LNK-037.

---

## 1. What the thing IS, before any rule about it

`datahelpers.BestLabelMatch(label, candidates)` answers one question: **given a button's own
words, which real page does it name?** No LLM, no embeddings — distinctive-token overlap against
each candidate page's name, title and nav_label, ranked by four keys in order: identity overlap,
total overlap, interactive (tool/game) pages first, and finally **the candidate's name in
alphabetical order**.

It was extracted from the misdirected-CTA check in `bugs_open/203` precisely so that the half
that DETECTS a wrong button and the half that REPAIRS it could not drift apart.

**The problem is key 4.** When two different pages tie on the first three, the winner is decided
by which page's name sorts first. That is not a fact about which page the label names; it is an
alphabet.

## 2. Problem + evidence

Key 4 is not a rare tail. Measured 2026-08-23 against a frozen fleet dump (829 page rows, 667
CTA component rows, 1,266 label-bearing url fields; the real `datahelpers` package imported, with
a control run of all 1,266 pairs against the shipping function returning **0 disagreements**):

| pool | matches | decided by alphabetical order alone |
|---|---|---|
| today's narrow writer pool | 781 | **177 (23%)** |
| the widened universe (`bugs_open/308` Phase B) | 1,146 | **263 (23%)** |

On the widened universe, **137 of the 428 writes** the widening would perform were among them —
i.e. a third of a fleet-wide content change decided by spelling. Two families in that set are
demonstrably wrong and were about to be executed:

- **finetuning.uk, "how we work"**, stored `/how-we-work.html` — the page its copy names — losing
  to `/about.html`, because the About page's *title* is "About Finetune | Who We Are and **How We
  Work**". 13 live findings.
- **dartsonline.com, "Read the guides"**, stored `/guides/index.html`, losing to an About page
  whose title ends "…Spec-First Darts **Guides**". 6 live findings.

Both are single-token ties (`[work]`, `[guides]`) where one candidate matched on its own NAME and
the other on marketing copy in its TITLE.

## 3. Design

`BestLabelMatch` returns a third value, `ambiguous bool`. When the winner is separated from a
**different page** (compared on `NormalizePagePath`, so two rows for one page are not a
disagreement) by nothing but key 4, it returns `ok=false, ambiguous=true`.

Two mechanics carry it:

- **The safe reading is the default.** A caller that ignores the new value sees `ok=false` —
  "this label names nothing" — which for both CTA writers means their keep branches hold the
  stored value, and for the detector means no finding is filed. **The ambiguous winner is
  unreachable by accident.**
- **The signature change is the enforcement**, not a comment. All three call sites had to be
  edited to compile, so each one decided. Same reasoning as `storedCTADestinationIsAuthored`'s
  signature (LNK-035) and the owner's 2026-08-02 ruling that a comment is not a control on a tree
  this many sessions share.

## 4. Alternatives considered — three tie-break keys, all measured, all rejected

**This is the load-bearing section, because the obvious answer is "break the tie better".**

| alternative | measured | why rejected |
|---|---|---|
| **do nothing** | 137 of 428 writes are coin flips, including 19 findings that are known-wrong | The widening executes them fleet-wide. Doing nothing was the option that shipped the damage. |
| **token-set-size key** (smaller candidate wins) | 2026-08-11 calibration | Dropped before shipping: a stray hyphen ("Break-Even" vs "breakeven") flipped **9** already-correct live gaswholesalers.com CTAs. Decided by tokenisation artefacts. |
| **name-tier key** (a match in the candidate's own `name` outranks a title-only match) | 2026-08-23, 43 picks changed on the narrow pool | Repairs both §2 families and gamesdesign's five "Browse … Tools" labels — and breaks others by the same mechanism: this estate names every tool page `tool-…`, so the token **`tool` sits in every tool page's name** and hands them all a point. |
| **path-depth key** (an area's own index outranks its children) | 2026-08-23, 61 picks changed | Repairs the section-index cases; drags unrelated labels onto `/guides/index.html` by demoting the interactive preference wholesale. |
| **stopword recalibration** (add `about`) — `bugs_open/308`'s own standing suggestion | 2026-08-23 | Suppresses four `Talk to us about …` → `/about.html` false matches **and** the correct `Learn More About Us` → `/about.html`. The lever does not separate them. |

**Three independent keys, two calibrations, the same failure shape: a tie at overlap 1 carries no
signal, so any key that breaks it is deciding by an artefact.** That is the finding this RFC rests
on, and it is why the proposal is to stop answering rather than to answer differently.

## 5. Blast radius, named

Derived with `go list -deps` per `cmd/` target, plus a symbol grep for callers outside
`platform/`:

- **Call sites of `BestLabelMatch`: three, all non-test, all inside the CTA seam** —
  `resolve_internal_links_action.go` (build writer), `rerender_page_sections_action.go` (repair
  writer), `check_misdirected_cta.go:ctaClassifyAnchor` (detector, itself reused by
  `check_cta_nonpage.go`). **No caller anywhere in `cmd/`, `internal/` or `pkg/`.**
- **Binaries whose BEHAVIOUR changes: `cmd/agent-chassis` and `cmd/core-manager`** — the two that
  execute the actions. `cmd/component-render-check`, `cmd/config-key-audit`, `cmd/instanceaudit`
  and `cmd/test-spawning` link the actions package but do not run these paths; six further
  binaries link `datahelpers` only and merely relink.
- **What changes for the detector's owner** (2026-07-29 ruling §3 — consumers must be *told*, not
  merely measured): **263 anchors stop being classified as naming a page**, so no finding is filed
  for them. 19 of those (the two §2 families) are known false and stop correctly. **The rest are
  unexamined**, and that recall cost is accepted rather than measured.

## 6. Staged rollout

One stage — the change is Go-only, has no config surface and no migration, and is inert until the
next fleet roll. There is nothing to arm and nothing to sequence.

Watched at the roll, in this order:

1. **Capability at the pod, with a control** — `LoadCTALabelUniverse` present, a string the change
   did not add absent, in the same exec, on every replica (the LNK-034 method; the `build
   provenance` startup line has usually scrolled by the time anyone looks).
2. **The artefact, not the status**: a `misdirected_cta` finding naming `/contact.html` must move
   the button on the SERVED page after an induced `cta_links_stale` rerender.
3. **The refusal must be visible as well as the repair**: `misdirected_cta` items filed per day
   should fall (the two known-false families stop), while `page_components` rows whose CTA url is
   a utility destination should rise from ~0.

**Induced-fault tests, already run** (mutation, against a clean `git archive HEAD` tree, 7 of 7
killed their test): removing the ambiguity refusal; re-pointing either writer at the old narrow
supply; restoring either keep's utility exclusion; dropping the universe's build-state predicate;
dropping the mint stamp from the label-match branch.

## 7. Rollback

Image-first and total: revert the commit and roll. No migration, no config row, no data written
that a previous binary cannot read — the only persisted artefact touched is
`content_data.__cta_minted`, which predates this change (LNK-035) and whose absence is already the
safe reading. **A rolled-back binary re-widens nothing and re-freezes nothing**; CTA destinations
already rewritten stay as they are and are re-derived on the next pass under the old rules.

## 8. Acceptance evidence — what would retire this RFC's risk

1. The three roll-time observations in §6, dated, at the artefact.
2. **A human pass over a sample of the 291 surviving writes.** §4 rejects the alternatives on
   measurement, but precision itself was judged by hand on samples — there is no ground truth for
   "this label names that page", and this RFC should not be marked IMPLEMENTED on the strength of
   an unaudited 291.
3. The residual stated rather than assumed closed: **a CONFIDENT false match is not caught.**
   dartsonline's *"See how each brand differs, spec by spec"* → `/about.html` wins on identity
   overlap (`spec`, from the About page's title), not on a tie. No tie rule sees it and no
   stopword list can.

## 9. The decision asked for

**Is "a shared matcher may decline to answer" the right shape, given that three attempts to answer
better have now failed on measurement?** The alternative the owner may prefer is that the CTA
writers should refuse ambiguity while the DETECTOR goes on reporting its best guess (a finding is
cheaper than a write). That would be a deliberate re-drift of the two halves — the exact thing
`bugs_open/203`'s extraction and `bugs_open/308` both exist to stop — so it is not proposed here,
but it is the live choice and it belongs to the owner rather than to this lane.
