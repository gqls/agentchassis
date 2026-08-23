# SUMMARY 2026-08-23 — apis.uk, the bees page

## What we're trying to do

Put a page about bees on apis.uk, and do it without disturbing the tools API that runs on
the same domain. The domain name is the reason the page exists: Apis is the genus the
honey bee belongs to, so a domain called apis.uk ought to be about bees. The owner wanted
the home page only, written as a personal enthusiast page rather than a beekeeping guide
or a conservation campaign. It sells nothing and collects nothing.

## Where we've come from

apis.uk was bought and, until now, used for one thing: a small API on a subdomain,
`tools.apis.uk`, which powers the debate game on vonc.com. Everything else on the domain
answered "not found" while quietly writing down who had asked, as a passive traffic probe
set up in July. Three of our own documents recorded the owner's intention to put a bees
page on the bare domain one day, and all three said that when it happened we would need to
repoint one DNS record.

## What we've done

We measured the DNS constraint before designing anything, and it turned out we did not
need to touch DNS at all. The bare domain was already being served by the same Cloudflare
worker that serves our other sites, because a routing rule intercepts the request before
Cloudflare ever looks up where the domain points. So the page could go live with no change
to the zone whatsoever, which is the safest possible outcome for the API.

We did find a real way to break that API, and it is not the one anybody had written down.
It is not the DNS records, which are per-name and independent. It is the routing rule: a
"everything under this domain" wildcard rule would swallow the API's hostname, hand it to
the wrong server, and return "not found" — with no DNS record changed, nothing looking
wrong, and our own tidy-up script reporting success. Two dozen other domains carry exactly
that wildcard rule quite correctly, because they do not have an API living on a subdomain.
This one does. That is now written into the standing trap list.

Then we built the page through the framework, seeded so it would produce one page and no
others, and it went live.

And then we made a serious mistake, which the owner caught by reading his own website.

The page carried four sentences telling the world that the domain also runs an unrelated
technical service on another hostname. The framework did not invent that. The brief we
wrote asked for it in as many words, on the reasoning that a developer arriving by mistake
should not be confused. That reasoning was wrong. The owner had asked us to protect the
API, and we turned that into describing it, which is close to the opposite. There is no
population of confused developers worth putting infrastructure on a public page for.

Three things made it worse than one bad sentence. We said "somewhere unobtrusive, once",
which names no particular place, so the writer put one in every section. The instruction
then spread into seven separate planning documents the system keeps for the site, including
one where it had become a formal acceptance criterion — so a check could in principle have
failed the page for leaving it out. And we verified the wrong things afterwards: we
confirmed the page count and confirmed the API still answered, both true and both
irrelevant, and never read the page.

It is fixed. The sentences are gone, confirmed by fetching the page rather than trusting a
status, and the API still answers. We also made the sentence refused rather than merely
un-asked-for, because deleting an instruction leaves no trace for the next person to see.

Two things caught us during the repair and are worth knowing. We cleaned the page's stored
content, confirmed it was clean, re-rendered, and got back a file identical to the byte —
the renderer had used a cached copy in a different column. And our clean-up left an orphan:
it removed the sentence naming the service but left the next one, which said "state it
once, style it quietly", and that fragment contained none of the words we were searching
for, so every check said clean. We found it by luck, and have said so.

Separately, the owner said the copy reads as machine-written, citing "worth sitting with"
and "not just" phrasing. He is right, and it traces to the same origin. Another team here
had already established that negatively-framed copy comes from the site's recorded
identity rather than from the model, and that reproduced here exactly: four of the five
recorded distinguishing features were written as "X, not Y" — a faithful summary of a
brief that was mostly a list of prohibitions. But the sharper cause was worse and simpler.
The system keeps a handful of example sentences for the writer to imitate, and four of the
five were themselves written in the style we forbid. The writer copied its examples: one of
them turned up three times on the finished page in barely altered form.

## Where we are now

The page is live, honest, and asserts no invented facts. Its voice is still wrong.

The style controls have been fixed properly, through the framework's own machinery rather
than by hand: the example sentences now demonstrate the style we want, the house
"de-AI-ify" rules are in the writing rules, and the specific phrases the owner objected to
are on the banned list. The site's recorded identity now says what the page is instead of
what it is not.

The rewrite itself has not run. It was dispatched and failed, because the account has hit
its API usage limit and access does not return until the first of September. The owner
already knows about that. Everything the rewrite needs is in place and the page is still
flagged for rebuilding, so it is one command when access returns. We have deliberately not
written the copy by hand: the owner asked for the framework to do the copywriting, and the
current text, while in the wrong voice, is safe to leave standing.

## Where we're going

Three things, in order.

When the API limit lifts, re-fire the rebuild and then read the resulting page as prose —
not as a row count, which is the check this lane has already got wrong once. Confirm both
the stored content and the cached rendered copy are clean, and re-check the API's liveness
as a separate fact.

Second, the traffic probe on the rest of the domain has never been read, and was due a
review on the eighth of August. The bare domain stopped being visible to it some time ago,
so the bees page takes nothing away, but every other hostname is still being logged and
nobody has looked. That is a small, separate job and the owner may or may not want it.

Third, our findings about where machine-sounding copy comes from have been handed to the
team working on exactly that problem, including a cheap experiment they could run to
settle a question they have open. That is theirs to take or leave.
