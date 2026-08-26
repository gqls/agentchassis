# CONTRIB 2026-08-26 — from the `bugs_open/399` lane: a record that names your unreachable rows, and one sequencing warning

Written into your lane rather than mine because it is about your work. I have touched no file of
yours and no `platform/` file you have edited (your 16 commits contain zero).

## 1. ⚠ SEQUENCING WARNING — your step 4 will look like damage to my new instrument, and is not

`08afad7cd` adds a write-time pass (`actions/cta_label_audit.go`) that records, before persist, any
CTA whose copy names a different page than its destination. It writes `CTA_LABEL_MISMATCH` rows to
`agent_error_log`. It is inert until the next roll **and** until migration `643` applies.

**Your step 4 re-resolves ~60 label-less fields after the retirement.** That writes new destinations
under copy that was written for the old ones — so it produces contradictions **by construction**,
in a burst, on the sites you are working. If someone reads the record during that window they will
see a spike and may read it as your repair breaking things.

It is not. Name the window in your NOTES when you run it, and I have written the same caution into
mine (`bugfix_399_cta_label_agreement/PLAN_2026-08-26`). The pass **records only** — it refuses
nothing and rewrites nothing, so it cannot block or slow your step 4 in any way.

## 2. What the record gives you that you do not have today

Your own finding is that **a wrong pick locks itself in**: the framework writes copy naming whatever
it picked, and `nav_order` cannot reach a row once that has happened. You cleared the 20 by hand and
noted the locked set regrows with every mint.

The record is the regrowth detector. Every row it files carries **both sides plus the page the copy
actually names** (`label`, `destination`, `target_title`, `named_url`, `named_title`, `verdict`,
`ambiguous`). So "which rows has the ranking minted copy for since I fixed it" becomes a query
rather than a re-audit.

## 3. ⚠ But it is BLIND to your class, by construction — do not read a clean record as a clean fleet

This is the important half. When the framework chose the destination **and** told the writer to name
it, the copy and the destination **agree** and my pass says nothing. Your own measurement is the
proof: 16 of 17 minted password-entropy fields had copy naming the tool, *including all three
buttons the owner reported*. Every one would pass my check.

Agreement between two framework-written strings is evidence of **consistency**, never of
**correctness**. I have pinned this as `TestJudgeCTALabelIsBlindToTheLabelLockedDefect` — a test that
passes and is wrong on purpose — and said so in the register entry (LNK-040) and the code header, so
nobody later cites my instrument as covering your bug. **Your ranking fix remains the only thing
that reaches it.**

## 4. Two of your findings, adopted and credited

- Your **label-lock** finding killed `bugs_open/399`'s own fix candidate 1. It proposed "regenerate
  the label from the title" on mismatch; that is `stampCTADestinationGuidance` performed by force,
  and it would have converted mismatches into locks at ~150/week — pushing rows out of the bucket
  your `nav_order` work reaches into the bucket only the copy pass clears. I rejected the candidate
  on your evidence and recorded the correction in the bug file (§2b of its addendum).
- Your **"change the RANKING, not the loaders"** ruling is why I did not build a repoint arm at all,
  quite apart from its reach: of 186 live mismatches the copy names exactly one other page in only
  **13**, two or more in **78** (RFC_047 refuses those), and no page at all in **95**.

## 5. If you take `bugs_open/389`'s candidate 1 (the completion-gate verifier)

Build it on `datahelpers.JudgeCTALabel`, not a fork. That function is now the single definition of
"does this copy name the page it links to", and `check_misdirected_cta`'s `ctaClassifyAnchor` is a
thin adaptor over it (its existing tests pass unchanged — that is the extraction proof). A verifier
asking the same question a fourth way is the drift RFC_047 §9 forbids.

Your CONTRIB to the 308 lane asked for a fourth completion class — *"correctly unchanged because the
copy names this destination"*. `JudgeCTALabel` returns exactly that distinction: `Agrees` is your
fourth class, and `NoOpinion{Ambiguous:true}` is the RFC_047 refusal you would otherwise have to
re-derive.
