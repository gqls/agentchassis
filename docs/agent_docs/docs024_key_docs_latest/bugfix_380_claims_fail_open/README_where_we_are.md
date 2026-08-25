# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-24 — picked up bug 380 and shipped the config half

What the bug is, in one breath: when a site has no verified facts on record, the three
things that were supposed to keep the writer honest all quietly switch themselves off — the
planner stops assigning facts, the writer is shown nothing, and the claims auditor exits
without reading a page and reports success. So the sites that know the least get checked the
least. That is how garden-tools.uk got a review methodology for a business that has never
tested a product, and how loanzy.uk got an invented credit broker.

What I found that the bug file did not say: the auditor was not merely leaky, it was never
being run at all — no seed file, no schedule, one call in its whole life. And the "make an
empty register" fix candidate would not have worked, because everything keys on the FACTS,
not on the register row. Someone had already diagnosed the real gap on 20 July and designed
the fix ("cold audit") and it was never built. So this is that design, built.

What is live now (all applied to the database this afternoon, no image needed):
- The auditor always runs. With no facts it runs "cold" — every claim about the business is
  treated as unsupported, and first-person practice claims ("we test", "we buy", "we garden")
  are the class it reports first. A database error now fails the run instead of quietly
  finishing it. Every run leaves a receipt, so "was this site audited?" is a query.
- The planner now tells the writer, on a factless site, to mark every section factless and
  never to brief a page as a description of practice.
- The auditor is on a clock: one site an hour, every site every seven days, new sites first.
- The auditor's page-reading was broken in a way nobody had seen (it had only ever run
  once): a PostgreSQL regex quirk meant it was throwing away most of every page. Fixed; the
  proof run now reads the whole of how-we-assess.

The proof: I pointed the cold audit at garden-tools.uk without touching the site. Its first
two findings were the two sentences you quoted — "We garden ourselves, and we test what we
can get our hands on" and "Where we can, we buy the tool at the same price a reader would
pay" — both at severity high, with the suggested fix being your own framing: reframe as an
aim and say when we have not. The control run on leopardess (which has a register) took the
normal path and found one real stale number.

What you decided today, and what I did with it: no empty registers get minted (absence is
the cold posture now); hourly tick with a seven-day window; the Go practice-claims detector
ships at "record, never refuse"; and the writer-prompt change waits for you to read it. That
last one is sitting as a held migration with the full new prompt text in
`brochure_component_library/sql/page_content_writer_prompt_v5_2026-08-24.txt` — say
"approved" and it goes live; say what to change and I will change it.

What is still to do: the Go half (the deterministic practice-claims family, warning only,
plus the fix to a latent hazard the regulated-attestation work left behind); the council
verdict on the config half (submitted, correlation e684fc8d); the register/landmine/notes
housekeeping; and telling the neighbouring lanes. Nothing on garden-tools.uk has been
changed — it remains the untouched measurement the loanzy lane asked for.

One thing I got wrong today and caught: I generated your read-copy of the writer prompt from
the committed migration file, and the live prompt had moved on by 1,700 characters since. The
copy you will read is regenerated from the live database, and the migration counts its
anchors on the live text at apply time, so the file-vs-live gap cannot bite the apply.

## 2026-08-24 evening — you approved the writer prompt; it is live

You said "approved". Before switching it on I checked that the text you read still matched the live
prompt: the only difference was another lane's change to the formatting rules (rules 9 and 10),
made an hour after I took my copy and reviewed separately — nothing in the part you approved had
moved. So the writer now carries the no-register arm: on a site with no verified facts and no
recorded operating history it is told, in plain words, not to describe testing, buying, measuring or
receiving samples as things we do, and that "say what we do" means only how the content is sourced.
That closes the third of the three fail-open paths at the source. Everything in the config half is
now live; the deterministic Go detector still waits for the next image roll.

## 2026-08-25 — the new chassis is out; everything checked at the running system; closing the bug

The build that went out this morning carries the Go half (I checked the running binary itself, with a
positive and a negative control, not the tag). Overnight the auditor did exactly what it was built to
do: one site an hour, fifteen sites, a receipt for every run, findings filed for ten of them and a clean
receipt for four. The writer has carried the "no operating history" instruction on every one of the
150 content calls since you approved it. The only thing not yet seen live is the deterministic
practice-claims warning at the build gate — the three pages built since the roll simply had no such
sentences to flag, so its silence there means nothing; it will show on the next rebuild of a page like
garden-tools' about page, and the handoff says how to check.

So I am closing bug 380: all three fail-open paths are shut, live, and exercised. What remains is other
people's work that this uncovered — the stale-render bug (386), the six other agents that complete on
a missing target, the four other queries with the same regex trap, wiring real research into new
builds (your call), and the human-review queue nobody works. The handoff lists each with its owner.
