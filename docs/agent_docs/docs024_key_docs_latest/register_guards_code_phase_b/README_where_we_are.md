# Where we are — the register that guards our words, and now our sums

Plain prose, append-only, newest at the bottom. Written for the owner.

## 2026-08-16

**The short version.** You asked me to take the next unworked bug in `bugs_open`.
That was 225 — the stamp duty calculator that charged first-time buyers using a rule
that stopped being law in March 2025. The good news is it was already fixed, and had
been since the 9th; I re-checked the live pages today to be sure rather than taking
the file's word. So the ticket itself was just tidying, and it has now moved to the
closed pile.

**The bad news is what the file said about itself.** Buried in it was a section
titled "Why no existing check could ever have caught this", and it was right. Every
check we own was blind to that mistake, and would be blind to the next one.

**Here is the shape of it, and it is worth a minute.** We keep, for each site, a
register of facts — every tax band, every rate, every threshold, each with a link to
the GOV.UK page it came from, and a robot re-checks those links every morning. It is
good machinery. But that register only ever governed what a page could **say**. It
never governed what a calculator **works out**. So the site could have the correct
figure sitting in its register, freshly verified that very morning, while the
calculator three feet away quietly used a figure that expired sixteen months ago. And
nothing anywhere would notice, because nothing was ever asked to compare the two.

There were three separate reasons nothing noticed, and the awkward part is that all
three are sensible decisions on their own. Our text checks ignore anything inside a
program — quite right, code is not prose. Our number checks skip calculator pages —
also right, a calculator's own help text is full of numbers that aren't claims. And
our number checks ignore money amounts entirely, because otherwise every price on
every site would trip them. Each is defensible. Together they leave a hole shaped
exactly like this bug, and nobody could see it because nothing was broken.

**What I've done about it.** I found that this had already been thought through —
there is a plan in the mortgagecalculator work, written on the 9th, that you have
seen, which sets out four pieces of a fix. The first piece is already live: when we
rebuild a calculator, the builder is now shown the register's facts. Pieces two and
three had been designed and then left, because that team moved on to logos and hero
images. So rather than invent something new alongside their plan, I built their
pieces two and three.

In plain terms: **a calculator can now declare which registered facts it relies on,
and when one of those facts changes, the morning check tells that calculator's owner
the same day.** If the Chancellor moves a stamp duty threshold, the register notices
overnight — as it always has — and now the calculators that encode that threshold get
named, instead of the change stopping at a note nobody reads.

**One decision I want to flag, because it is a judgement rather than a technicality.**
When a figure moves, I have made it *hard* for the system to fix a calculator by
itself. It will only hand the job to the automatic fixer when the calculator owns its
own code and its own settings say auto-fixing is allowed. In every other case — and
that includes every case where it's the *evidence* that changed rather than the
number, such as a GOV.UK page disappearing — it stops and asks a person. There are two
scars behind that. Our automatic fixer has twice rewritten a shared template and
changed a hundred-odd pages when it was asked to fix one. And a false alarm once
pointed the fixer at a page's legal disclaimer. Neither of those is something I want
happening to arithmetic on a page that quotes tax law.

The honest consequence: on today's sites, *every* route ends with a person, because
both stamp duty calculators are set to "don't auto-fix" and neither owns its own code.
The automatic path exists and is tested, but it has never run in anger. I'd rather
tell you that than let it read as more automatic than it is.

**Two more things I should say plainly.**

The first is a limit. This tells us when a figure has **moved**. It cannot tell us
whether a figure is **right**. If the register and the calculator are both wrong in
the same direction, they agree, and this says nothing. Answering that properly is the
fourth piece of that plan — a thing that works out the correct answer independently
from the published rules — and it needs its own review before anyone builds it. The
tool that actually caught this bug does exactly that, but it lives in a folder and
only runs when a person runs it.

The second is that none of this is switched on yet. The code goes live with the next
release, and after that somebody has to tell the stamp duty calculator which facts it
relies on — one line in its settings. I've written that up and handed it to the team
that owns the site, since it's theirs, not mine. Until that line exists, this machinery
is real but idle, and I've said so in the tracking file rather than letting a green
tick imply otherwise.

**Also worth knowing:** I got something wrong today and logged it. I quoted two fleet
statistics in a commit message from a research assistant's summary rather than
checking them myself; they were about ten percent out. The conclusion they supported
was still true, which is precisely what makes that kind of error easy to leave in
place. It's written up in our wrong-calls log.
