# README — where we are (bugfix_400 news goto URLs)

*Append-only, newest at the bottom. Plain prose for the owner.*

## 2026-09-03 — what this is, and why it turned out to be a different bug

Some of our sites carry a news feed. Each item links out to the article. On several sites, a
chunk of those links do not point at the publisher — they point at `google.com/goto?url=…`
followed by a long meaningless token. The link still works: Google bounces the reader on to the
real article. But the reader hovering over it sees Google, not the newspaper, and we lose the
one thing the news feature is for, which is naming a real source.

**It looked like an ongoing problem and it is not — quite.** 1,378 of these are stored across
eleven sites. But they stopped arriving on **28 August**, and nothing has arrived since: six days,
about 1,300 new news items, not one bad link. The feed is busy and healthy throughout.

**We did not fix it, and that matters.** The same news sources are still running and still
producing items — they have simply started returning proper links. So something changed at
Google's end, not ours. Which means it can start again tomorrow, and **nothing in our system
would notice**. Nobody would find out until someone happened to look at a page.

**Meanwhile the old ones are still on the sites today.** I checked two: idea.uk is serving two bad
links out of six, mortgagecalculator one out of six. So the thing that is actually live is not the
intake — it is the backlog nobody has cleaned up. The original bug report proposed three fixes and
all three address the intake; every one of them could ship tomorrow and those pages would look
exactly the same.

**Two useful things I established.** The token cannot be decoded — it is meaningless without
Google. But asking Google once, politely, still gives back the real address: I tested three and got
hpcwire, Fortune and Nature. So the old links can be repaired rather than deleted.

**I have stopped at the design.** The work splits into three pieces — a watchdog so a silent
restart cannot go unnoticed, a fix so the bad form never gets stored again, and a clean-up of the
1,378. I have written the order and the traps into the handoff rather than half-building it,
because the code sits in a part of the system that should go through review, and because the
clean-up has a sharp edge: repairing a link can collide with a copy of the same story that was
stored correctly, and that needs measuring before anything is rewritten rather than discovered
mid-way.

One small thing worth saying because it nearly cost me the whole diagnosis: my first test of
whether the old links still work used a stronger form of the request, which follows the redirect
all the way to the newspaper — and the newspaper blocks us, so it came back as a failure. The
weaker request, which stops at Google, works perfectly. I had briefly concluded the backlog was
unrecoverable on the strength of the more thorough-looking test.
