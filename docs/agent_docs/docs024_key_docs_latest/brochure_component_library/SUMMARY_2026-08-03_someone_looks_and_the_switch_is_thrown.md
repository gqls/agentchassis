# Someone looks, and the switch is thrown — 2026-08-03

**What we're trying to do.** Build fundamentallyai.com's brochure site out of reusable,
provable components, and — because three separate defects reached the owner on pages
where every automated check passed — close the gap between "the machine says the page is
fine" and "a person has actually looked at it".

**Where we've come from.** Yesterday's summary was called "the camera works and nobody
looks yet": a passing acceptance run photographs its page and files the pictures as
storage links inside a technical note, and that was where it ended. Separately, another
lane had built us a duplicate-section checker — detection, a deterministic repair, a
guard that refuses to delete anything the site's own plan asks for — council-approved
twice and deliberately left switched off, with the switch named as this lane's to throw.

**What we've done.** Both halves moved today. The checker is enabled: every safety figure
was re-measured this morning rather than requoted (the guard proven in the running
binary on both replicas, the would-delete census re-run over all 1,189 live sections —
zero — and the one plan-specified repetition confirmed guarded), the switch went in as a
numbered, guard-checked seed, and the first watched run deleted nothing and filed one
honest advisory: our own pages repeat facts at each other, which is precisely the
structural problem (151 candidate 1) still on the books. And somebody now looks at the
photographs: a one-command contact sheet pulls the renders out of the private bucket and
puts them on one page, the first sheet is published to the owner, and the first look
found a real defect in the camera itself — it fires after the checks have driven the
page, so a healthy page can be photographed in a state no visitor ever sees. Every image
now carries that warning. The look also settled a design question with a 22,491-pixel-tall
mobile photograph: renders now record their viewport (implemented, reviewed, inert until
the next release). The content loose ends are cleared too: /tools.html exists and the
"Explore All Tools" buttons point at it, the calculator's dead buttons work, both
companion guides — one promised by live copy since 25 July — were written and published
through the real pipeline, and a nine-day-old 404 stub is retired.

**Where we are now.** The looking loop is closed at its cheapest useful form: machinery
photographs passing pages, one command puts the photographs in front of a person, and
the first person to look found something worth fixing. The duplication checker is live
fleet-wide and has bitten nothing. The site's three tools have a shopfront, working
buttons, and complete companion guides.

**Where we're going.** Three open calls, none technical. Whether the contact sheet comes
to the owner on a cadence or stays on-demand. Whether the camera should photograph the
landing state before checks drive the page — that changes what a render means, so it
needs the owner, not a session. And 151 candidate 1 — assigning facts to sections at
plan time so sibling pages stop restating each other — which today's first sweep measured
on our own site (nine overlapping fact-pairs) and which remains this lane's largest
unbuilt piece.
