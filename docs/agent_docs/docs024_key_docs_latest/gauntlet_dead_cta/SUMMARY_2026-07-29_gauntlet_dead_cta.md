# SUMMARY — gauntlet_dead_cta, 2026-07-29

**What we're trying to do.** Make vonc.com's Gauntlet — a timed, judged debate
against a live AI opponent — an honest product someone would return to. The
standing rail: nothing on the page claims what is not true by construction,
and no control changes state except as the consequence of a real API response.

**Where we've come from.** The owner's usability audit (131 A–G) was fully
engineered by the evening of the 28th: sealed reveal, question card, step
ranking, and a shareable verdict card, all live and verified. That left one
open product question — H: why argue here rather than in a free chat window.
On the morning of the 29th the owner ruled: the distribution experiment first
(he does the posting), feeding the arena thesis — plus a named feature in his
own words, "a (dated) personal history of your opinions might be a goldmine".

**What we've done.** The opinion ledger — that dated personal history — is
built, delivered and live. Every round a visitor finishes on a device becomes
a dated diary entry at the foot of the page: the day's provocation, the
position they actually filed, the judge's verdict. It is device-local
(localStorage, no accounts, nothing transmitted), entries can only be born
from a real /defend response — never on restore, never backfilled — and the
visitor can erase the record with two presses. It was proven twice over: a
25-check local harness through two real live rounds, then 13 checks on the
served production page including a full real round whose verdict became
exactly one entry. Building it also caught a real engine defect: the newer
Claude judge model now thinks by default and was silently spending the
engine's 2,048-token answer budget on thinking, truncating verdicts — the
armed log from bug 083 named it on first read, the fix (four times the room)
was approved unanimously by the council on round one, deployed to the island
and verified with six clean verdicts. 083 is closed.

**Where we are now.** The page a visitor meets today: a sealed door to
today's provocation; a real round with a real opponent and a real judge; a
verdict they can keep as a card; and — new — their own dated ledger of every
round they've finished, giving a returning visitor something that is theirs.
The engine judges reliably again. One watch item left the workstream: the
platform-wide sibling of the engine bug (a closed bug's "thinking is off
fleet-wide" premise is no longer true on Claude 5 models) is filed for
diagnosis rather than asserted.

**Where we're going.** The distribution experiment is the owner's leg: the
share card and daily provocation travel to where people already argue, and
real behaviour chooses between the examination and arena theses. On the build
side, the arena thesis's ingredients (categories, one provocation per day
each, communal views once participants exist) wait on that evidence. The
og:image fix (the other bug numbered 131) would give shared links a face;
131-B still wants a witness on the acceptance lane.
