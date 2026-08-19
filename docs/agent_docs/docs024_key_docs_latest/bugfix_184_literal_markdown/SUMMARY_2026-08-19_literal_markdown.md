# Summary — bug 184, literal markdown on live pages — 2026-08-19 (closed; follow-up approved)

**What we're trying to do.** Our AI content writers sometimes type markdown formatting
symbols — `**bold**`, `# headings`, backticked code, `[link text](web address)` — into
page fields that the site shows as plain text, so visitors see the raw symbols. We want
those symbols off the live pages and the mistake to stop recurring, on every site the
framework builds.

**Where we've come from.** Filed on the 3rd of August. The detector worked from the
start; the first repair (ask an AI to rebuild the page) did not, because the rebuilding
AI has the same habit. On the 18th the repair went mechanical: a small shared cleaner that
can only delete symbols, run through the existing reliable page re-render for exactly
the repair jobs, plus the same cleaner where new content is born. The council approved
that at round four.

**What we've done.** Two fleet releases landed on the 19th. The first proved the design
works — and found the next layer: news-listing pages pull their items fresh from the news
feed every time they render, and the feed itself carries raw markdown (about 700 of
10,900 stored articles), so cleaning the stored copy was undone in the same run. We
cleaned that layer too — at the feed projection, where the value is display text — and
the second release carried it. Then the proof: the repair was run on seven pages; six
completed first time, each certified by the honesty checker, and five were confirmed
clean by fetching the live page with their text intact (fundamentallyai.com/news went
from 13 visible symbols to none). The bug's own detection query returned nothing; the
three pages it was filed on are clean. **The bug closed on that evidence.** The council's
fifth round asked for polish on the feed-side cleaner — mainly an off switch — and round
six delivered it tonight (an operator can disarm the feed cleaner without a rebuild; it
ships switched on; it logs what it cleans; two over-claims in my earlier write-up
corrected) and was **approved** with advisory notes only.

**Where we are now.** Closed, live, verified at the served page. The round-six code is
committed and rides the next release; nothing depends on it urgently — it is an off
switch and a log line. Residuals are routed, not hidden: 41 items on webdesign.co.uk's
"ported" tool pages are structurally out of this route's reach (the tool-rebuild
programme owns them); one robot-hands page has clean stored content but the old file still
serves (a publish gap, the re-render lane's); one robot-hands URL is a plain 404; markdown
tables in a handful of feed articles are outside the pattern set (feed quality); and the
RSS feed — the third reader of the same feed column — is now tracked as bug 332, latent:
only one site publishes RSS and its feed serves no markdown today.

**Where we're going.** After the next release: prove the round-six code landed on every
chassis pod (the runbook says how). The detector's ordinary sweeps will close the
duplicate items on their own now the pages scan clean. Two things to own from tonight: a
comment edit of mine was carried into another session's commit a few minutes before I
committed (nothing lost), and I sent the round-six review twice by re-running the send
script to read its output — one duplicate round of reviewer credits, logged with the check
that prevents it. Nothing further is owed on this lane unless bug 332's trigger fires.
