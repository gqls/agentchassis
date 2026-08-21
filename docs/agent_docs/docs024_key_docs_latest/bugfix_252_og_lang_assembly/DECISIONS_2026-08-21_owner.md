# DECISIONS — what needs your call, bug 252 (og/lang slug), 2026-08-21

Written at your request: *"explain the decisions I need to make here."*

**Decision 1 closes this lane. Decisions 2–4 are things this lane uncovered that will bite something
else if left. Decisions 5–10 are lower stakes and can wait.** Each says what it is in plain terms, why
it needs you rather than me, the options, and what I'd do.

Already decided by you and recorded — not reopened here: locale lives in the head component (option 3,
08-11); opt the estate in now rather than shipping the switch off (08-20); non-English sites must not
be `en-GB`, and that generalises to future language sites (08-20); **do not force the page rerenders,
let rebuilds carry it (08-21)**.

---

## 1. Can bug 252 close? — the only decision needed to finish this lane

**What happened.** The fix is live on `v1.0.1321` (proven in the binary on both replicas, with
controls) and proven on real pages: a page now states its own share-preview identity, agreeing with its
canonical, and declares `en-GB` (or `es-ES`). Every site's stored page-header is repaired — 22 of 24
now carry a language, none still bakes the homepage address, and the duplicated tags are gone.

**What is not finished.** A repaired page-header is not a repaired *page*. A page only picks it up when
it next rebuilds. Measured this morning: **252 of 722 pages (35%) carry the fix, 470 do not — and 13 of
26 sites are at zero percent.** Your ruling was to let rebuilds carry it, which I've followed. But the
natural rate is about **one page an hour fleet-wide and bursty**, so for the quiet sites that means
*effectively never*: finetuning.uk (49 pages), loancalculator.co.uk (43), leopardessconsulting.co.uk
(40), mortgagecalculator.co.uk (30) and nine more have had nothing rebuild since.

**So the two questions are genuinely separate**, and I don't want "fixed and live" to imply the fleet
is clean:

- *Is the defect fixed?* **Yes** — it cannot recur. Any rebuild from now produces a correct page.
- *Is the damage gone?* **No** — 470 pages, with no scheduled end.

**Options.**
- **(a) Close 252 now and let the residual decay.** Honest if the bug file says the residual out loud.
  The defect is what a bug file tracks, and this one is dead.
- **(b) Close 252 and open one small tracking item** for the 13 zero-percent sites, so "these sites
  still serve a wrong share URL" has somewhere to live and can be swept whenever a lane next touches
  each site. Costs a bug number.
- **(c) Keep 252 open until the residual clears.** I'd advise against: it may never clear, and an open
  bug that cannot be closed by anyone stops meaning anything.

**My recommendation: (b).** Close the defect, and let a one-line tracking item carry the 13 sites — the
natural fix is "when a lane next rebuilds its site, this heals for free", and that only happens if
someone can see the list. (a) is defensible if you'd rather not spend the number.

---

## 2. webdesign.co.uk has no page-header element at all, and it is our largest site

**What it is.** Its page-header component is a fragment — the opening and closing `<head>` tags are
simply missing. Browsers cope, so the site looks fine. But every tool we have that adds something to a
page header looks for that closing tag, and they don't all behave the same way when it isn't there.
One of them gives up silently, which is why **that site alone gets no share-preview image and no
favicon tags at all** — and now, no language either.

**Why it needs you.** It is a hand-authored component on a live 117-page site — our biggest. Fixing it
is a small edit plus a canary, but it will change that site's served bytes, and it is not part of bug
252.

**Options:** fix it now (one component edit, one canary page, then let rebuilds carry it as above); or
accept that this site opts itself out of every current and future page-header feature.

**My recommendation: fix it.** It is cheap, and the cost of leaving it compounds — it has already
silently excluded the site from three separate features, and nobody noticed for weeks because every
tool involved reports success.

---

## 3. New sites will keep defaulting to English, and nothing will notice

**What it is.** I set each site's language explicitly, by name, so no site got a guess. That worked —
and it caught a real case: the first attempt **aborted** because `indoorplanters.co.uk` had been
created that same day and wasn't on my list. But that was a one-off script. **The next site created
gets no language at all**, falls back to English, and no check anywhere would report it.

**Why it needs you.** "What language is this site?" is a product question, and a silent default is
exactly the fault this bug was about.

**Options:** add a small daily check that lists real sites with no language set (cheap, and catches
sites created by paths nobody remembers); or set a default at site-creation time (faster but
re-introduces a silent default); or accept and handle it per site.

**My recommendation: the daily check.** It surfaces the question instead of answering it wrongly.

---

## 4. The share-preview block still has the flaw that caused all this — I only fixed the symptom

**What it is.** The block that writes those tags into each site's shared header skips itself entirely
if it finds either of two specific tags already present. That guard is why webdesign.co.uk gets
nothing (decision 2), and why four sites had *duplicate* tags — it couldn't see a blank one.

I removed the one page-specific tag from that block. **But the guard is untouched, so the next
page-specific tag anyone adds there reproduces bug 252 exactly.** The review council's bug-historian
seat raised this independently and framed it better than I had: I fixed a symptom; that guard is the
mechanism.

**Why it needs you.** It's filed as item 4 of `bugs_open/322`, currently reading like a tidy-up among
five items. It is the highest-value item in that file.

**Options:** re-prioritise 322 around item 4; or leave it in the queue.

**My recommendation: re-prioritise.** This is the difference between fixing an instance and closing a
door.

---

## 5. Two of our own checks now state something false, and neither will complain

`verify_site.py` exempts the page-address tag from validation as an "accepted loss", and a fleet
checker excludes it with a written rationale. **Both rest on "the shared page-header cannot carry a
per-page value", which is no longer true.** Neither fails loudly — they just keep passing, which is
how a blind check outlives its blindness.

**Why I haven't just fixed them:** the first belongs to another lane, and the second sits in a file
covered by a test that scans that source, so editing even a comment there has a documented way of
going wrong.

**Options:** route both to their owning lanes with a note (safe, slower); or have me do them with a
canary.

**My recommendation: route them.** Small, and not worth me reaching into two other lanes' files.

---

## 6. An unroutable instruction to the platform reports success having done nothing

**What it is.** I sent a valid-looking instruction with an incomplete envelope. The platform accepted
it, recorded it, and marked it **completed** — having run nothing. It cost me a wrong conclusion: I
wrote up two working agents as broken before I found my own error.

**Options:** make the platform reject an instruction that resolves to no workflow; or document the
envelope in one place instead of every lane copy-pasting it from whichever script it found (mine was a
partial copy).

**My recommendation: both, cheaply — the rejection is the real fix.** A success that did nothing is
the worst possible answer.

---

## 7–10. Lower stakes, listed so they aren't lost

- **The diagnosis loop can't see what a site actually serves.** It looks for page headers in three
  database columns that have been empty fleet-wide for months, and truncates the one place the
  evidence lives. It returns "unverifiable", which reads as *your claim is doubtful*. Worth fixing;
  documented as a trap meanwhile.
- **Migration numbering has no allocator.** Two numbers were already used twice before I started;
  five consecutive numbers were taken by three sessions while I wrote two files. A duplicate-number
  check at commit time would catch it.
- **A page-header template can reference a setting that was never defined, silently and for ever.**
  The template behind four sites' blank tags did exactly that, in four places. An audit comparing
  template references against declared settings would find the rest.
- **A description-filling helper can still write into the wrong tag.** I narrowed it so it can only
  reach tags this fix rewrites, but any *other* blank tag in any page header is still exposed. A
  census would probably show the compatibility behaviour it exists for is needed by nothing.

---

## What I'd do next, given your rulings

1. Your call on **decision 1** — then I either close 252 or close it plus a tracking item.
2. Your call on **decision 2** (webdesign.co.uk) — I'd take it as a small follow-on.
3. **Decisions 3, 4, 5** become notes routed to the right places; no code from me unless you say so.
4. Everything else sits in `FINDINGS_2026-08-21_errors_caught.md` until someone picks it up.
