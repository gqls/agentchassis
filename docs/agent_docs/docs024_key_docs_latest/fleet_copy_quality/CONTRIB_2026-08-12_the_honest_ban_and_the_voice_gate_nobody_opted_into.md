# CONTRIB 2026-08-12 — the 'honest' ban went fleet-wide, and it found a mechanism at 2/23 adoption

**From the `idea_uk_vm_site` lane, at the owner's instruction. Written for the eleven
other lanes whose sites I edited today** — per the 2026-07-29 ruling that a shared
mechanism's other consumers must be **told**, not merely measured.

Full working: `idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.53–§X.54.

---

## 1. What the owner asked for, and the sentence that matters

> *"let's ban the word 'honest' across all the sites … it works in the hero copy … but
> we've overused it everywhere else. I would very seldom use honest in my speech if
> ever as it's such a strong word, yet the copy uses it spattered all over the place."*

Also in the same brief, and worth carrying into any voice work: **fewer "riddles"** —
the owner's word for *"slightly obscure follow-on text that you have to think hard to
understand"*.

## 2. THE FINDING — the owner already ruled this on 2026-07-18, and it never propagated

`leopardessconsulting.co.uk`'s `voice` spec has carried this for 25 days:

```
voice_gate.banned_phrases[11].pattern   \bhonest(ly)?\b
voice_gate.banned_phrases[11].reason    owner 2026-07-18: overused; show the honesty, do not label it
banned_language[8]                      honest / honestly — demonstrate it, never label it
```

**There is a live, code-enforced mechanism for exactly this class of rule:**
`check_voice_tells` (`platform/orchestration/actions/discovery_checks/check_voice_tells.go`),
driven by `quality-discovery-agent`, compiling each site's `voice_gate.banned_phrases`
as case-insensitive regexes and filing `voice_tells` items at `needs_human_review`,
deduped on `voice:<page_id>`.

**It is opt-in with the unsafe default OFF — and only 2 of 23 sites have opted in**
(leopardess 14 patterns, oufe 10). `[MEASURED 2026-08-12]`. The mechanism works; the
adoption is the defect. If your lane has been hand-policing voice rules in prose, this
is the place to put them instead.

## 3. What I changed on your site

**Voice specs (13 sites).** 39 prescriptive strings across 15 specs, superseded not
edited in place, so your history is intact (`is_current=false, superseded_at` set).
`formatted` regenerated for every `content_direction` per
`datahelpers.FormatContentDirection` — **the writer reads only `formatted`**, so a
change to the arrays alone would have steered nothing.

Applied the owner's own principle — **show it, do not label it** — so each replacement
states the concrete behaviour the label stood for. The spec gets sharper, not shorter:

| site | example |
|---|---|
| noted.co.uk | *"Earns trust by being honest about what the tool cannot do"* → *"…by saying what the tool cannot do"* |
| mortgagecalculator | *"Trust comes from clear sums, plain words, and honesty about what an estimate cannot know"* → *"…and saying what an estimate cannot know"* |
| webdesign.co.uk | *"State limitations honestly where they exist"* → *"State limitations plainly where they exist"* |
| fundamentallyai | *"Frame it as honest positioning, not a weakness"* → *"Frame it as accurate positioning"* (also kills an antithesis) |

**Page copy (3 sites, 78 `section_edit` items at `triaged`, `item_key
honest20260812:<page_component_id>`):** finetuning.uk 48, fundamentallyai.com 17,
idea.uk 13. Targeted by `page_component_id`, **not** slot name — slot names repeat
within a page and a name-keyed edit is ambiguous.

## 4. What I deliberately did NOT touch, and why

A mechanical ban-list is this directory's own recorded mistake (2026-08-11: *"the filler
list is a smell, not a crime"* — four sound card descriptions were reported as defects
purely for matching a list written two hours earlier, and the owner overruled it). So:

- **the ban regex itself** and the owner's reason line — removing them disarms the rule;
- **`submission` / `mission_brief`** — the record of what was originally *asked for*.
  webdesign.co.uk's submission literally reads "TONE AND HONESTY"; rewriting it would
  falsify history rather than fix copy;
- **`vertical_landscape`** — research findings about *other people's* sites
  (*"Gousto's… honesty about costs before cart stage reduces abandonment anxiety"*);
- **`briefing.honesty_rails`** (dartsonline) — a named truth-constraint mechanism
  ("never claim to stock, carry, hold or ship products"). Compliance, not voice.
  `[VERIFIED]` no Go reader (`grep -rn honesty_rails --include=*.go` → nothing), so
  renaming it was *possible* and still wrong;
- **"dishonest"** on finetuning.uk/pricing — *"pricing either job the same way would be
  dishonest to whoever pays less"* means **unfair**. Different word, not self-labelling.

Post-sweep: **4 rows of 17 left, all four the exclusions above.**
**Control: 26 specs still match `'%plain%'`**, so that zero is not a blind query.

## 5. Landmines from doing it — read these before you run a similar pass

- **A blunt find-and-delete produces ungrammatical copy, and it will look fine in a
  truncated diff.** Six distinct defect classes surfaced only because the review printed
  the **changed region** rather than the first N characters: `"or an 'not yet'"` (the
  a/an guard cannot see past a quote mark); `"Sometimes,, it's"` (deleted parenthetical);
  `"a short, way to find out"` (deleted from an adjective list); `"nobody was about what
  the project required"` (copula + "honest about"); `"give you an A Clear Assessment"` (an
  unanchored heading rule firing mid-sentence); and lowercase sentence starts, because
  lowercase replacement literals turn *"An honest steer"* into *"a steer"*.
- **Two of those reached the queue before I caught them.** All 48 items were still
  `triaged`, so the specs were corrected in place — but the lesson is that the
  post-edit check has to assert on **shape** (double commas, dangling articles, residual
  matches), not just on whether the word is gone. A find-and-replace that removes the
  target word 100% of the time can still be wrong 100% of the time.
- **Do NOT arm `voice_gate` fleet-wide before remediating.** ~108 pages across 14 sites
  carry the word today; arming first files ~108 items straight into
  `needs_human_review`, where **leopardess already has 33 sitting unactioned**. Clean,
  then arm, so the gate starts from a clean baseline and only ever reports regressions.
- **Arm it NARROW.** The gate also runs em-dash density, triads, long sentences,
  contractions and flourish endings, and those thresholds are `zero → default`. A
  minimal banned-phrases-only gate needs them set high **deliberately**, or 679 pages
  get a full voice audit nobody asked for.

## 6. Still owed

- **~30 pages on the remaining 11 sites** (leopardess 12, ai-agent-orchestration 8,
  loanandmortgagecalculator 7, vonc 5, dartsonline 4, cookly 3, loancalculator 2,
  mortgagecalculator 2, webdesign.co.uk 1, robot-hands 1, gaswholesalers 1). Your site,
  your voice — the measurement is above, the method is §3, and the traps are §5.
- **Then arm the narrow gate**, fleet-wide, and the rule stops depending on anyone
  remembering it.
