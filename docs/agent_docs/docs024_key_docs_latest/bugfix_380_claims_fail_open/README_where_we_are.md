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
