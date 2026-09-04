# README — where we are, bugfix 463 (plain prose, append-only, newest at the bottom)

2026-09-03.

You asked me to look at bug 463, check nobody else was already on it, and fix it properly —
fixing the framework rather than the one site.

Here is what the bug was. When the system plans a website, it has a step that cleans up the
plan before saving it. One of that step's jobs is to throw away a page that would clash with an
existing section page — if the site already has an "Articles" section at `/articles/`, and the
planner proposes a *second* page also called Articles, the second one should go. Sensible.

The trouble is how it decided what "clashes". It only looked at the first part of the web
address. So `/articles.html` (a genuine clash) and `/articles/how-to-hire-a-designer.html` (an
actual article, living inside the section) both came out as just "articles", and it threw away
both. Every new article, every time.

Nobody noticed for three months, and the reason is worth knowing, because it is why the bug
could sit in plain sight. A later step in the same routine puts back any page the site already
has. So on an established site the article gets deleted and immediately restored, and
everything looks fine. The damage only shows up where there is nothing to restore — a section
that is empty *right now*. Those can never be filled, ever, no matter how many times you
rebuild. That is gamedesign.uk's empty articles page, and it is not just that site: 53 of the
78 section pages across the estate currently have nothing in them.

Then I found a second problem, and this is the part I want to flag, because without it I would
have shipped a fix that looked right and did nothing.

When the plan is finally saved, the system does not use the web address the planner chose. It
rebuilds the address from scratch from the page's type. For an article, the rule is "put it in
`/blog/`" unless something tells it otherwise. Nothing ever tells it otherwise — the planner is
never even asked which section a page belongs to. I checked the live instructions we give it:
32,000 characters, and the field simply is not mentioned. And of the 109 article pages in the
system, all 109 have it blank.

So even with the first fix, the article would have survived the cull and then been filed under
`/blog/` instead of `/articles/`. The Articles page would still have found nothing in it, and
the safety check that holds back empty section pages would still have held it back. Same empty
page, different reason — and the bug report specifically says this part is *not* the problem.
It was, for the second half. I have corrected the bug report where it says otherwise.

Both halves are now fixed and committed. It is being reviewed by the automated review council.
It will not take effect on the live system until the next time the software is rebuilt and
deployed, which is not something this thread does on its own.

Three other things came out of it.

I found a separate fault while I was in there, and filed it rather than fixing it, because it
needs a decision rather than a repair. There is a cap of 20 pages on a plan. Once a site has 20
pages already built, the cap throws away *every* newly proposed page — not the excess, all of
them. 26 of our 42 sites are already past that line, the biggest at 151 pages. So on most of
the estate a rebuild currently cannot add a single new page, silently. That is bug 467.

I also got something wrong myself, and it is recorded. I ran a check to size a risk in my own
fix and it told me five sites were affected. The number was believable and I nearly wrote it
down as fact. It was my query that was wrong, not the sites — the real answer was zero. What
caught it was looking at the five rows it returned rather than the number.

Finally, five other threads were working nearby and I spoke to all of them. One was building
something inside the very same file at the same time — half of the reporting side of this same
bug — so I deliberately did not build that half, and we agreed how the two pieces fit. Another
corrected me on which thread owns the related safety check. A third checked my claims
independently and then found a fourth place with the same underlying problem, which I have
recorded but not fixed, because it is a dormant part of the system and needs a different repair.

What is still owed: proving it works on the real site once the software rolls. Until then the
honest status is "fixed in the code, not yet proven in the world", and I have left the bug open
on exactly that basis.

---

**2026-09-04, later — the owner ruled on three things, and turned the fourth into a bigger
question.**

He decided the page cap first. Today, when a site is re-planned, the system keeps every page
that is already built and then refuses to add anything new once the site has twenty pages. That
is why two-thirds of our sites cannot grow by re-planning at all. His ruling is that a re-plan
may add up to ten new pages, and the important word is *add*: the limit is on what one re-plan
brings in, not on how big a site is allowed to be. A site of a hundred and fifty pages can still
grow by ten. Worth writing down because there is an easy way to get this wrong — there is already
a setting called max_pages that looks like the right knob, and turning that up would raise the
ceiling on site size, which is precisely what he did not ask for.

He agreed to leave this bug open until we have proved it on a real site, rather than closing it
because the code is live. That is the whole lesson of this bug, so it would have been a strange
one to skip.

He also said the verification re-plan can wait its turn in the normal queue. It costs a build
cycle to jump it forward and there is no reason to.

The fourth one he did not decide, and instead opened up. My question had been narrow — how do we
fix one producer that always writes articles into a folder called "blog". His answer was that the
folder names themselves are a constant source of bugs, and he wants it looked at properly.

So I went and counted. Across the whole estate there are fifty-eight different folder names. Only
eleven of them appear on more than one site. Thirty-nine contain exactly one page. Three names do
almost all the work — tools, guides and blog — and those three are the only ones written into the
code; everything else is whatever the language model happened to type that day. That is how we
have ended up with guides, guias and buying-guides meaning the same thing, and with articles,
insights, news and noticias all being the same idea.

The reason is simple once you look. A page's folder is either a fixed word in the source code,
chosen by what kind of page it is, or it is free text from the model with nothing checking it.
There is no middle option — no list of folders a site actually uses that a new page has to pick
from. So every new page is either forced into one of three folders or free to invent its own.

I also found the duplicate-name problem he mentioned, and it is two separate things. One is small
and clearly wrong: seven pages across three sites exist at two addresses at once, like /news.html
and /news/index.html. That is the same confusion that caused this very bug, because the code
treats both as the same section. The other is bigger and might not be wrong at all: thirty-eight
tools have a companion guide page named after them. The trouble is the companion lands in a
folder called guides on eight sites and blog on three, and on one site it does both. That is not
a decision anyone made; it is drift.

I have written to the four threads that hold pieces of this — the one that owns the older
flat-address bug, the one that filed the producer bug, the one closest to the planner, and the
site we are using to prove the fix. I have not proposed a solution, because changing a shared
vocabulary like this is the kind of change our own rules say needs a proper architecture review
rather than a patch, and because there is one thing I could not answer from the code alone:
whether we already have a place where a site declares its own folders. If we do, this is a much
smaller job than it looks.
