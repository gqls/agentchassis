# CONTRIB from the `bugs_open/463` lane — triage of owner-flagged decisions in `bugs_open/`

**Written 2026-09-04 at the owner's direction:** *"please send the triage of the owner flagged
bugs to the parked findings thread."* Not routing work at you and taking nothing — this is a
list, and the lane that owns each bug still owns it. Contributed here because your lane is the
one working through things that are parked awaiting a decision, and this is the same shape in a
different table.

## What this is, and what it is not

I swept all **185** files in `bugs_open/` for text indicating a decision sits with the owner.
**~50 files matched. The large majority record decisions he has ALREADY made** — the phrase
"owner decision" appears far more often as provenance for a past ruling than as an open ask.
Separating those two is the whole value here, because a grep alone reads them identically.

**Confidence is marked per item and it is not uniform.** For groups A–D I opened the file and
read the actual question. For group E I have only the matching line, so they are candidates, not
findings. **Do not present group E to the owner as established.**

**One caveat that applies to everything below.** A bug file is a record of what a lane believed
when it wrote that line. Several of these are weeks old, and this estate's own rule is that a
claim decays faster than its damage. **Re-read the file and check for a later correction before
putting any of these in front of him** — the 2026-08-31 case where a closed blocker kept being
obeyed for 20 days is exactly this failure.

---

## A. Genuinely blocked, and the stakes are high — verified by reading the file

**1. `bugs_open/233` — rotate the leaked credentials. The highest-stakes item on this list.**
The B2 application key pair and the `clients_db` password were logged in plaintext at INFO,
and the file says they *"should be presumed disclosed to anyone who has had log access since
2025-10-28"*. **The logging fix is LIVE and verified at the pod; the rotation is not, and is
explicitly an owner decision, deliberately not taken.** The file's own summary of what is left:
*"Nothing technical. One owner decision."* Rotating B2 touches `personae-default-secrets` plus
the GitHub-secrets copy; rotating the DB password touches every service. **Nothing degrades
while this waits, which is precisely why it can sit unnoticed for a month.**

**2. `bugs_open/259` — one provision request builds several billable GPUs.** Status is OPEN and
CONTAINED, and the containment is that **provisioning is PAUSED fleet-wide** by owner decision.
So a capability is switched off estate-wide pending a call. Worth asking whether the pause is
still wanted or whether the fix should be prioritised to lift it.

**3. `bugs_open/243` (the Anthropic usage-limit file — note there are TWO 243s, resolve by slug)
— whether to add a second LLM provider.** Recorded as *"still an open owner decision"*, with
`127 of 127` configured LLM steps on a single provider. This is a resilience decision, not a
bug fix.

## B. Blocked, and blocking other work

**4. `bugs_open/407` — header membership.** Fix is written, council-approved round 1, committed,
inert until its migration applies. **One decision is owed and the file deliberately does not
assume it:** the `guardian` seat objected that the fix lets a site's declaration override
*three* independent membership guards at once (`pages.in_header`, `neverPrimaryTypes`, and the
child-URL bar), which is a semantic widening that deserves explicit sign-off rather than being
cited by analogy to the 2026-08-02 ruling. The file agrees the seat is right.

**5. `bugs_open/384` — does it close on "blog listings recover by rotation"?** The seam files a
rerender whose workflow has no listing-rebuild step, so on a blog listing the item completes
without rebuilding; the one site that recovered did so on ordinary rotation, not via the seam.
**Two defensible readings, laid out in the file.** The bug cannot close either way until it is
picked.

**6. `bugs_open/361` — the 18 dispositioned findings: gate or bank?** Three of four closing
items are done. This one is untouched and is an owner decision, and item four
(`lastSuccessfulTime` moving) **cannot happen until it is answered** — so a daily job stays red,
correctly, on 18 real findings in 5 components, for as long as this waits.

**7. `bugs_open/033` — human review queue.** Council seats reached opposite conclusions in the
same round. The file's reasoning is worth quoting: *"Seats disagreeing with each other is the
signal it needs a human, not a resubmission."* The ask is whether the contract gets a mechanism.

**8. `bugs_open/131` — vonc gauntlet usability.** Item **H is the owner's call and the file says
it blocks the rest** — items C–F should not be done before it. So this is one decision gating a
batch of design work.

## C. Spend and scope — cheap to answer, and nobody can proceed without it

**9. `bugs_open/114` — widen `check_content_image_missing` to `page_type='content'`?** Flagged in
the file as **"unasked as yet"**, i.e. it has never been put to him. Implies fleet-wide image
generation spend, which is why the lane would not assume it. Also holds a second, smaller ask:
the disposition of five parked rows whose pages resolve no sections.

**10. `bugs_open/320` — meta descriptions, 56% of live pages have none.** The file states plainly
that **none of its remedies should be actioned without an owner decision**, and raises a sharper
question underneath: descriptions written before a certain migration came from a possibly
truncated view of their own page, so the model wrote fluent copy from a fragment and nothing
downstream could tell.

**11. `bugs_open/395` — routing findings at a writer that can overwrite.** Marked an owner
decision that **"neither lane may take"**, so two lanes are both correctly stalled on it.

**12. `bugs_open/436` — CTA destinations.** Approved in principle on 2026-08-25, **not built**,
and no page is opted out yet.

## D. Judgement calls that are genuinely his, not engineering

**13. `bugs_open/446` — a games site with no identity.** The evidence rules forbid inventing an
author, so the page either **gets an owner-supplied identity or stays anonymous by design**.
There is no technical answer available here; that is the point.

**14. `bugs_open/462` — logo regeneration.** *"Regeneration is the owner's call, not the
check's."*

## E. Candidates — matched the sweep, NOT verified. Do not present as findings.

`181` (where a file should live) · `178` (restoring clobbered content) · `388` (apply a key or
not) · `398` (retire a sibling scheduled row) · `236` (blocked on the RFC_012 decision) · `153`
(needs an owner cycle; scope veto) · `257` (token-budget configuration end) · `296` (225 parked
contrast findings — plausibly overlaps YOUR lane directly) · `309` · `417` (architecture-scope,
unbuilt) · `427` (build the calendar mechanism) · `428` (carve out a narrow exception) · `442`
(needs a decision plus a council round) · `203` · `234` · `315`.

**`296` is the one I would check first** — 225 parked contrast findings whose retraction is
already live is your subject matter, not mine.

## F. Already decided — listed so nobody re-raises them

`467` (**ruled today**, below), `480`, `316`, `205`, `116`, `136`, `440`, `432`, `113`, `248`,
`071`, `392`, `420`, `481`, `445`, `457`, `447`, `241`, `406`, `151`, `040`.

## Rulings the owner gave today, for your records

- **`bugs_open/467`** — a re-plan may add **up to 10** new pages. The cap binds what a re-plan
  **ADDS**, not what a site may **CONTAIN**. ⚠ Do not implement by raising `max_pages`: that
  raises the site ceiling, which is not what was asked.
- **First plans keep ~20** — *"a brand new site should have as many pages as is necessary but we
  can cap it at approximately 20 for now."* **20 is an interim pragmatic cap, not a principle.**
- **`bugs_open/463`** stays OPEN despite meeting the "fixed AND live" bar — and the verification
  run has since justified that: half the fix was inert in production.
- The page-directory vocabulary question (`468`/`460`) was **widened** rather than answered, and
  is heading for an RFC.

---

*Contributed by the `bugs_open/463` lane. Nothing here is owned by me; correct anything I have
mis-stated in place rather than forking a second account of it.*
