# PROCESS — reports and takedowns for The Gauntlet

**Status: LIVE from 2026-08-10.** This is RFC_020 §5.3, the half that is not code.
Owner rulings it rests on: *"go ahead as you suggest"* (2026-08-10, covering §5.2,
§5.3 and §5.4) and, on the two open details, *"no stated time, and address can be
the email"* (2026-08-10, evening).

This document is the process. It is short on purpose: a procedure nobody can hold in
their head is one that gets improvised under pressure, which is the exact situation
it exists for.

---

## 1. What this covers, and why it exists at all

The Gauntlet publishes a **round record page** at a permanent URL when somebody
shares their argument. RFC_020 §1.4 sets out why the usual "we are just a platform"
position is weaker here than for a forum, and both halves of that are design facts
rather than accidents:

- **The verdict is the site's own text.** When the system pronounces that an argument
  succeeded, that sentence was written by the service, not by a user.
- **The poster cannot be identified.** There are no accounts. Rounds are anonymous by
  design (`bugs_open/139`), which is the condition under which an operator generally
  has to rely on a proper notice-and-takedown process to keep its position.

Until 2026-08-10 there was **no published way to complain**. That was the gap. A
process that exists but cannot be found is the same as no process, which is why §5.3
is a page change and a document, never a document alone.

## 2. The published route

**Address:** `vonc@contactforsales.com` — read from `sites.email` for vonc.com, which
is where the owner's contact data actually lives (the two `site_specs` identity stores
are both empty for this site; see LANDMINES, *"A site's contact email lives in THREE
stores"*).

**Where it appears:** the round record page footer, immediately after the sentence
that asserts nobody is named — because that sentence is precisely what a complainant
is disputing. `round_record/round_record_component.html`, `.gr-report`.

**What it says:**

> Does this page say something about a real person or business? **Report this page**
> — please include its web address in your message. Every report is read, and
> anything we cannot stand behind comes down.

**No response time is published, by owner ruling.** "Every report is read" is a
commitment one person can keep without a rota. A number would be a promise the
staffing cannot back, and a *missed published deadline* is worse evidence about an
operator than no published deadline ever was. Do not add one later without deciding
who is on the hook for it.

## 3. What to do when a report arrives

**The default is to take it down first and work it out afterwards.** A round is one
anonymous person's argument on a debate toy. It is worth approximately nothing to us
and potentially a great deal to the person named in it, so the asymmetry only points
one way. Nothing below should be read as a reason to leave a contested page up while
somebody deliberates.

1. **Find the round.** The reporter is asked to include the URL; the slug is its last
   path element.

   ```sql
   -- gauntlet_rounds lives on the ISLAND VM, not clients_db. See RUNBOOK
   -- gauntlet_dead_cta §5 for how to reach it.
   SELECT slug, created_at, published FROM gauntlet_rounds WHERE slug = '<slug>';
   ```

2. **Unpublish it.** This is the reversible action and it is the first one:

   ```sql
   UPDATE gauntlet_rounds SET published = false WHERE slug = '<slug>';
   ```

   ⚠ **Verify at the artefact, not at the row.** `tools-api` caches rounds in
   `provocStore` (`round.go:25-29`) with a five-minute TTL, so the page can serve for
   up to five minutes after the database says it is gone. Re-fetch until it 404s:

   ```
   curl -sI 'https://vonc.com/tools/gauntlet/round.html?slug=<slug>' -H 'Origin: https://vonc.com'
   ```

3. **Reply to the reporter.** Say what was done. Do not admit or deny anything about
   the substance of the claim — that is a question for a solicitor, and §1.4 says so
   twice.

4. **Record it.** Append to `NOTES_gauntlet_dead_cta.md`: the date, the slug, what the
   report said in one line, and what was done. **The tally is the point** — a takedown
   log with several entries about the same failure mode is the evidence that a control
   upstream needs to change, and no single incident can tell you that.

5. **Only then** decide whether anything is restored, and get a human decision on
   record before it is.

## 4. What should have stopped it earlier, and what to check afterwards

A report that reaches us is a control that did not fire. Check which one, and say so
in the log:

| control | state | if this is what failed |
|---|---|---|
| `namecheck` publish gate (RFC_020 §5.2) | **built, NOT live** — ships from the island VM | Nothing was ever going to stop this. Chase the island deploy; that is the fix |
| `namecheck` allow-set tuning | live only once the above ships | The term was too narrow. Widen it and record the term |
| `pages.noindex` (§5.1) | live | The page was findable by searching a name — the worst multiplier. Check the page went through the head producer that honours the field (`bugs_open/232`: there are two, and only one does) |
| verdict scope line (§5.4) | live on both surfaces | The reader took the verdict as a finding of fact. That is a wording problem, and this document's §2 is the wrong place to fix it |

**The most likely answer today is the first row**, and it is worth being blunt about
that: §5.2 is written, council-approved and **not running**, so the platform currently
has *no* automated control that stops a round naming somebody. This process is
carrying that gap on its own until the island deploy happens.

## 5. What this process deliberately does not do

- **It does not promise a timescale** (§2).
- **It does not adjudicate truth.** We remove what we cannot stand behind; we do not
  rule on whether an allegation is accurate.
- **It does not identify the poster**, because it cannot. That is the design, and it
  is the reason this document exists rather than a "contact the author" link.

---

**Related:** `architecture_review/RFC_020_third_party_harm_in_the_gauntlet_before_and_after_publish.md`
(§1.4 the position, §5 the controls, §7 build status) · `bugs_open/139` (anonymity) ·
`bugs_open/232` (noindex, two head producers) · `RUNBOOK_gauntlet_dead_cta.md` §5 (how
to reach the island VM).
