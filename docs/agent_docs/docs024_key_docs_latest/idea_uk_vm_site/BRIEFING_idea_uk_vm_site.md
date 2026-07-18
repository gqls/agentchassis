# idea.uk → the VM: a full briefing

**Written to be read aloud. As of 2026-07-18.**
Companions in this folder: `SUMMARY` (one-page status), `PLAN`, `RUNBOOK` (the how),
`RUNNING_NOTES` (the chronological record), the `box/` folder (ready-to-run box scripts),
and `sql/` (every database change, in order).

---

## 1. In plain terms

idea.uk exists today as two halves that have never been able to see each other. On one side there
is a marketing website — nine pages, built automatically by our platform — which is published to
cloud storage where, as it happens, **nobody can actually see it**, because the domain name points
somewhere else. On the other side there is the thing that earns money: a small paid tool that sells
a £29 report, running on a rented server in Germany. The domain points at that server, and the
server currently shows only the tool. So a visitor to idea.uk sees the tool's single built-in page,
and the nine-page marketing site sits in storage as an elaborate rehearsal that never goes on stage.

The goal of this work is to make idea.uk **one complete website**: the marketing pages and the paid
tool, together, behind one front door, on that same server — without ever taking the money-making
tool offline.

Over the course of this work we have done the groundwork for that, fixed a serious security problem
we stumbled into along the way, hardened the tool against spam it was quietly collecting, and
prepared everything so that the final switch-over is a short, well-rehearsed job. The remaining steps
run on the live server itself and need hands with access to it. Nothing we have done so far is
visible to the public yet — which is deliberate. It means every change has been staged and checked
before anyone relies on it.

---

## 2. What we set out to do

Two goals were set at the outset, and a third was added during the work.

The **first** was the one above: publish idea.uk's automatically-built marketing pages onto the
German server instead of into invisible cloud storage, so the site is whole and still includes the
paid tool.

The **second** was a list of loose ends on the site itself: three catalogued pages that had never
actually been built, so their navigation links led to "page not found"; and the spam that the tool's
public enquiry form was collecting.

The **third**, added partway through, was simply to record the work and the reasoning properly — in
this folder — so that anyone picking it up later inherits the decisions and the traps, not just the
code. Two facts shaped everything: there will eventually be **thousands** of these domains, so
nothing we build for idea.uk is allowed to become a per-site special case that could not scale; and
the existing publishing machinery already treats each domain as a folder in a shared repository.

---

## 3. The decisions that shaped the work

Before describing what was built, it is worth stating the handful of decisions that everything else
follows from, because they were deliberate and they are the reason the design looks the way it does.

**We kept the site static; the server is a second destination, not a second factory.** Our platform
is, at heart, a machine that turns data into finished web pages. That rendering happens in our
cluster, where the data lives, and the result is plain HTML. We decided the server should simply
*serve* those finished pages, not try to build them itself. The server has no route to our database
and no ability to render, so building on the server would mean either giving every server a dangerous
line into our central database, or rebuilding the whole rendering machine on each box — which is the
finished page again, with more moving parts and more to go wrong.

**We put the whole site behind one front door — the server — rather than stitching two together.**
One certificate, one DNS record, one origin. Forms, links and cookies simply work when everything
lives at one address. This also delivers a quiet reliability win: because the pages are just files
with no application behind them, a crash of the paid tool can no longer take the marketing site down
with it. Today it can — because the server currently sends *everything* to the tool.

**The server pulls its files; we never push into it.** Each server reaches out and fetches its own
pages on a schedule, rather than something in our cluster reaching in to deposit them. This is a
security decision, and section 7 explains it in full — it is the difference between one stolen key
exposing the whole fleet and a compromised box being able to damage only itself.

**This stays the exception, not the rule.** For the thousands of ordinary marketing sites, publishing
to cloud storage remains the default. The server model is reserved for the handful of sites that
actually sell something and therefore need a backend beside their pages. We were careful never to
build something that only works if every site gets its own server.

---

## 4. What we've done

### The security incident we walked into

While mapping how the tool worked, we found something that was not on anyone's list: **real
credentials sitting in a public code repository.** A configuration file — misleadingly named as an
"example" file, which is why nobody had looked inside it — contained genuine credentials for sending
email as idea.uk, and the tool's own internal master key. They had been there, publicly readable, for
around six weeks.

We judged the file by the *length* of each value rather than its name, because that is the only
reliable tell, and confirmed that the email credentials and the internal key were real, while the
payment and AI-service keys in the same file were only truncated placeholders — so, importantly, **the
payment path was never exposed.** What was exposed was the ability to send email as idea.uk, which
risks the domain's sending reputation and the associated bill, and the internal key that guards the
tool's order-approval actions.

We cleaned the file immediately and installed an automatic guard that now blocks any full-length
credential from ever being committed again, and we made sure that guard is tracked so every copy of
the code inherits it. But cleaning a file does not undo history — the old values remained readable in
the repository's past. **Rotation is what actually closes it, and that is now done.** On the 17th, a
fresh email-sending user was created and confirmed working, the old one was deleted outright, the
internal key was regenerated, and the tool was restarted healthy. The values still visible in history
are now dead keys. We deliberately did not rewrite the repository's history, because once the
credentials are rotated those old values are worthless, and rewriting history would break every copy
of the repository for no security gain.

### Completing the site

The site had three catalogued-but-never-built pages, whose links led nowhere. Building them exposed a
genuine trap in our own site-planning machinery: the documented "safe" way to fill in missing pages —
re-running the planner — turns out to **silently discard the design of pages that were already built**,
and still cannot fill the empty ones. We caught this after it had regressed four finished pages, and
because nothing was public yet, we were able to recover cleanly: we restored the four pages exactly,
removed ten pages the planner had invented unprompted, and kept the two genuine fixes. The site is now
a coherent nine pages, all published, with the two previously-broken navigation pages built.

The ninth page deserves a note, because it became a small reusable pattern. The "Free Audience Check"
was catalogued as a page but is really a doorway to the live tool. Rather than build a redundant
in-between page, we turned it into a **pointer**: the link on the site now goes straight to the live
tool, and the page itself is pinned so the system leaves it alone and never tries to publish an empty
file for it. One click to the real thing, no dead end.

This whole episode is written up as its own fix-handoff and recorded as a fleet-wide trap, because
left unfixed it will quietly degrade any site that ever gets re-planned.

### Moving the deploy target — and discovering the pipe was never connected

The mechanism that lets a specific site publish to a server instead of to cloud storage had been
*designed and written months ago, but never actually wired in* — the right function existed and
nothing called it. We connected it, and it shipped in the platform image. Turning it on for idea.uk
was then gated behind one safety step, and that step turned out to matter more than expected.

The shared publishing repository already had an automated job that copied **every changed domain to a
single server** — and that server was the wrong one for idea.uk (it belongs to another site). The
moment idea.uk's files landed in that repository, that job would have pushed them onto the wrong
machine. So before switching anything on, we replaced the single-destination job with an **explicit
map** of which domain goes to which server, from which idea.uk is simply absent because it pulls its
own files instead. We proved this live three times: idea.uk is skipped, and the other site still
deploys correctly.

In the middle of proving it, we found something more fundamental: **that publishing job had never
successfully run in its life.** It had no build machine registered to run it, and even when we
provided one, the machine's image was missing the very tools the job needs to copy files over a
secure connection — so it had been failing invisibly, with its error output sent nowhere. "Publishing
is automatic" had only ever been true for the cloud-storage path; the other site's files must have
been copied across by hand at launch. We rebuilt the machine image with the missing tools, stood up a
dedicated build machine for this repository, and confirmed the whole chain works end to end.

With the safety map in place and proven, we switched idea.uk over to the server destination and seeded
the repository with the site's finished pages. From here on, every future rebuild of the site
publishes to the right place automatically.

### The tool's spam problem, and an email that was being blocked

The tool's public enquiry form had no defences and was collecting fake submissions. We added the
standard protections — a hidden honeypot field, a too-fast-submission check, per-sender rate limiting,
proper email validation, and capture of the sender's address so a block-list can be built later — all
covered by automated tests. This is ready to go out with the tool's next update.

Separately, the owner noticed that the *first* email of an order arrived but the important one — the
draft report for review — had been flagged as spam. We read the blocked message's own scoring report
and found the cause precisely: the tool was **putting the customer's entire multi-hundred-character
business description into the email's subject line**, which reads exactly like keyword-stuffing to a
spam filter. The first email survived only because its subject was short. We fixed the tool to keep
the subject brief and leave the full text in the body, with tests. The scoring report also confirmed
that the sender's own reputation was fine and the message was otherwise clean — so the subject was the
whole problem.

That investigation led into email deliverability more broadly. We confirmed the domain's
email-signing is correctly set up and passing, and identified the one genuinely missing piece — a
"custom bounce address" that makes the tool's mail fully aligned with the idea.uk domain — which is
two small DNS records for the owner to add. This matters not just for the internal review email but
for the paid reports going out to customers, which otherwise carry the same handicap into other
people's inboxes.

### The contact form, and a stale address it revealed

The site's contact page had a form whose "send" button did nothing useful — it pointed at a dead
anchor. The owner chose to convert it into a direct email link. Tracing where that form's behaviour
comes from, we found the correct per-site lever and changed only that, leaving the shared template
that every other site uses untouched. While there, we caught that the form's description still quoted
the **old** contact email — the address had been updated elsewhere on the page but this copy had been
missed — and aligned it too. The fix is staged at the source and will publish with the next rebuild of
that page.

### The server-side work, prepared in advance

Finally, so the owner's part is as short and safe as possible, we wrote out every server-side step as
ready-to-run scripts in the `box/` folder: the pull-sync installer, the sync script and its schedule,
and the complete web-server configuration — including a correction we found late, that the site ships
its own copies of the legal pages, so the configuration now redirects those to the tool's authoritative
versions to keep the published terms from drifting from the ones a buyer actually agreed to.

---

## 5. Where we are now

The status, plainly:

- **The security leak** — repository cleaned, automatic guard installed, and **both credentials
  rotated with the old ones destroyed.** Closed.
- **The site** — a coherent nine pages, all published, navigation fixed. Done.
- **The planner trap** that caused a scare — recovered from, and handed off as its own fix. Done for
  idea.uk; the underlying fix is a separate job.
- **Publishing to the server** — the mechanism is wired, shipped, guarded, switched on for idea.uk,
  and the repository is seeded with the site. The build machine that was silently missing has been
  built and proven. Done.
- **The tool's spam defences and the subject-line fix** — written and tested, waiting to go out with
  the tool's next update.
- **Email deliverability** — signing confirmed good; one two-record DNS addition left for the owner.
- **The contact form** — fixed at the source, publishes on the next rebuild.
- **The final switch-over on the server** — fully prepared as ready-to-run scripts; needs hands on
  the server.

In short: everything that could be done away from the live server is done. What remains is the
switch-over itself, which is intentionally a server-side job.

---

## 6. Where we're going

Three things remain, all on the live server and all prepared:

**First, set up the pull-sync** (section 7 explains it). This is safe to do at any time — it puts the
site's files onto the server without changing anything the public sees. It is one script; it pauses
once to let the owner register a read-only key, then does the rest and checks that all eight pages
have arrived.

**Second, the web-server switch-over.** This is the one moment that changes what visitors see. The
prepared configuration serves the marketing pages as the site while sending the tool's own paths
straight to the tool — all sixteen of them, a number we checked carefully against the running tool,
because an earlier draft listed only seven and the gap would have silently broken the free taster and
the operator's approval flow. Before the switch, every one of those paths is tested, and a real
payment is put through the money path, to prove nothing broke. The switch itself is a single change,
the domain does not move, and rolling back is one line — the tool and its data are never touched by
any of it.

**Third, deploy the tool's pending update** — the spam defences and the subject-line fix — by building
the small binary, copying it across, and restarting. The tool has no automated release, so this is a
deliberate manual step.

Alongside those, two small owner tasks close the remaining edges: adding the two DNS records for the
custom bounce address, and clearing the historical spam out of the tool's order file.

---

## 7. How the pull-sync works, and why it is built this way

This is the heart of the server side, so it is worth understanding in full.

**What it does.** A small scheduled job on the server, every five minutes, fetches idea.uk's finished
pages from the shared repository and mirrors them into the folder the web server reads from. The
server reaches out and fetches; nothing in our cluster ever reaches into the server.

**What it syncs.** Exactly one folder — idea.uk's — out of a repository that will eventually hold many
domains. That folder is the finished artefact: the eight HTML pages, the stylesheet, the images, and
the small pieces of page JavaScript. It is only files — no code, no database, no templates. The ninth
page, the Free Audience Check, produces no file on purpose, because it is a doorway to the live tool,
which the web server hands straight to the tool.

**How it is built.** There are three tiny pieces: a script that does the sync, a service that runs
that script as the web user, and a timer that fires it a minute after boot and then every five
minutes. The sync script does three things — it fetches the latest from the repository; it forces its
local copy to match exactly, throwing away any local drift rather than trying to merge, because the
repository is the single source of truth and the server is only a mirror; and it copies the result
into the web folder, deleting anything that was removed upstream so the mirror is faithful. A one-off
installer sets all this up: it creates the folders, generates a **read-only** key for the server to
identify itself to the repository, pauses so the owner can register that key, then fetches only
idea.uk's folder — even from a repository of thousands, it downloads just its own — installs the three
pieces, runs one sync, and confirms the eight pages are present. It pointedly does not touch the web
server, so setting up the sync can never accidentally flip the site over.

**Why the server pulls, instead of us pushing to it.** The tempting alternative is to have a job in our
cluster copy files *onto* the servers. But then a single key, held in our cluster, can write to every
server in the fleet — and compromising that one place reaches everything. Pulling inverts this: each
server holds only its *own* read-only key that can do nothing but read from the repository. A
compromised server cannot write anywhere and cannot reach its neighbours, so the damage is contained to
the one machine already lost. It also self-heals — a server that was offline during a change simply
catches up on its next five-minute tick, with nothing to retry. The honest cost is that a change takes
up to five minutes to appear, rather than being instant, which for a marketing site is a fine trade.

**Why go through the repository at all, rather than deploying pages straight onto the server.** Three
reasons. First, the platform already publishes every site this way — the server is just a *second
destination for the same finished pages*, chosen by a single setting, not a new and bespoke pipeline
that would drift from how every other site works. Second, it keeps our cluster from holding keys to
internet-facing servers at all — the cluster only writes to a repository, and the server reaches out
to read. Third, every publish is a recorded change you can see, compare, and roll back with ordinary
version control; a direct copy onto disk leaves no history and no undo. And underneath all of it, the
server stays deliberately simple — a web server and a scheduled fetch, nothing more — which is exactly
what lets the marketing site stay up even when the paid tool is having a bad day.

---

## 8. Things we are carrying, and decisions we don't want re-litigated

**Noticed but not chased** (logged for their own threads): a spare copy of the fleet's build machine
has been crash-looping for weeks, so that redundancy is currently gone though publishing still works on
its healthy twin; the generated contact forms across the whole fleet point at a dead endpoint (idea.uk
is now fixed, but the fleet-wide version is a separate job); and the build system occasionally re-does
work it has already finished. None of these block idea.uk.

**Decisions taken, so they are not re-opened:** the site stays static and the server only serves it;
the whole site lives behind one front door on the server; the server pulls its own files rather than
being pushed to; the Free Audience Check links straight to the live tool with no in-between page; and
the tool keeps its own legal pages as the authoritative ones. This whole model is kept as the
*exception* for sites that sell something — the thousands of ordinary sites stay on the default
cloud-storage path.

**Two small open questions, neither blocking:** whether, after the switch-over, the privacy page is
served by the tool or the static site (the default is the tool, since it is tied to the purchase
terms); and confirming whether the domain sits behind the content-delivery network in its protective
mode, which decides whether the visitor's true location is visible to the tool for any future
abuse-blocking.

---

*End of briefing.*
