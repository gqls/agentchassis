# noted.co.uk — where we are

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-08-10 — first day

**The short version: the live site is safe and now warns people, the backup
button was quietly broken and now isn't, and the framework knows the domain
exists. The rebuild itself hasn't started, on purpose.**

### What I found when I looked

noted.co.uk is your hand-built notes app — text, voice recordings, photos,
a bit of version history, a share button. It works. I checked it in a real
browser rather than just reading the code, and it loads clean with no errors.

The surprise was where it lives. It's already sitting in `portfolio-sites`,
which is the framework's own deploy bucket — the same one every framework site
goes to. It just got there by hand rather than through the pipeline, and it
wasn't in any git repository at all. So there was no way to review a change to
it, no way to revert one, and no record of what had been deployed. It's the same
shape of thing as the webdesign.uk shopfront, except this one predates that
ruling by six months and is an actual application rather than a page.

I also found three generations of the app on your disk, and the live one is
none of them exactly — it sits between the second and third. The third
generation, the one with the improved sharing, was never deployed. It also has
two broken links in it (a social preview image and a favicon that don't exist
anywhere), so it's just as well.

### The thing that would have gone wrong

You asked for a notice telling people to back up. I nearly shipped one, and then
checked what the backup button actually does.

**It only saves your text.** Voice recordings, photos and version history aren't
in the file. I proved it rather than assuming: I put a note with a real 4 KB
audio clip and an 8 KB photo into the app, clicked Backup, and got a 334-byte
file with nothing in it but the words.

So the notice on its own would have been actively harmful — it would have sent
people to a button that loses exactly the things they can't retype. A typed note
you can write again. A recording of someone's voice you cannot.

I fixed that first. There's now a **"Save everything"** button that includes the
recordings and the photos, and the Restore button understands the new file. I
tested the whole round trip: save everything, wipe the browser's storage
completely, restore from the file, and the audio and photos come back
byte-for-byte identical. That test is the point — an export nobody can restore
isn't a backup, and it would have looked fine without it.

### What's live now

The notice is up, in a soft amber bar across the top. It says the app is being
rebuilt, that the new one will let you sign in so your notes follow you between
browsers, that right now everything lives only in this browser, and please save
a copy. There's a "Save everything" button right there in the bar, and another
in the sidebar. It's dismissible, and it works on a phone.

I also ran the framework's contrast checker over it, which found four
accessibility failures — three of them already there before I touched anything
(the delete icon, the small print in the sidebar, the footer on the about page).
All fixed, and the checker now reports zero across both page sizes.

All of this went in through the proper deploy path — committed to the `sites`
repo, deployed by the GitHub Action, Cloudflare cache purged automatically. So
the app is now under version control for the first time, which matters more than
the notice does.

### One mistake worth telling you about

Partway through I convinced myself the live site was badly broken — that a
missing button meant the JavaScript was crashing on startup and the backup
button was dead. I'd read a `diff` backwards. It was wrong, and I only caught it
because I ran the actual page in a browser instead of trusting my reading.

I've logged it properly in the fleet's wrong-calls file. The part I want to flag
to you is that the false alarm *felt* more certain than the truth did, because it
was the more interesting story. The real problem turned out to be one function
below the invented one — and only actually running the thing found either.

### Your clarification about decomposition

You corrected me mid-way, and it changed the plan substantially:

> the keyword I was looking for was decomposition rather than locked

That's a much sharper requirement than what I'd been planning against, and it
rules out the easy route. The framework has an adoption mode that swallows an
existing site whole — each page stored as one lump, deployed and monitored but
never re-planned or rewritten. That would have been fast and it would have
looked like success. It's exactly what you've ruled out: the framework would be
*hosting* it, not *controlling* it. Upgrades would never reach it and the
section-level checks would have one giant section to look at.

So the plan is now the slower, real one: every prose section its own object,
the editor as a proper tool component, its JavaScript deployed as real assets,
and — the important bit — the app's *behaviours* written down as contracts the
framework can check. That last mechanism already exists and is used by the
darts gauntlet: a behaviour like "the note you wrote on your phone is there when
you sign in on a laptop" becomes a named thing with a test attached, rather than
something that either works or doesn't.

### The one place "everything" isn't currently possible

I need to be straight with you about this rather than discover it in three
weeks.

**The framework doesn't write server code, and nothing in it does user
accounts.** Every backend on the estate — the relojistas engine, the idea.uk
payment service, the webdesign chat box — is hand-written Go. The one thing that
does have accounts, `auth-service`, is for platform operators, is only reachable
inside the cluster, and isn't reachable from a public site at all.

So the notes server will be written by hand, the same way the other three were.
That's the normal pattern here, not a shortcut. What the framework *can* own is
the contract it must honour, whether it's up, whether the behaviours actually
work in a real browser, and whether it fails honestly when it's down. Everything
except the binary itself.

If you want the framework generating backends too, that's a real platform
project and should be decided as one, not smuggled in under a site rebuild.

### What I need from you

1. **Which VM.** My recommendation is the webdesign.uk box — 8 GB and 50 GB of
   disk versus the island's 2 GB, and its Cloudflare tunnel means the notes
   service never has to accept a connection from the internet directly. The
   island is tempting because it already has a database, but I don't think one
   core and 2 GB shared with the tools API is right for the only copy of
   strangers' notes. Nothing I've done today commits us either way.

2. **What the privacy promise becomes.** This is the one I'd most like you to
   think about. The current site's whole pitch is that there is no server —
   "we can't see your notes, read your text, or listen to your recordings". The
   moment you can sign in from another browser, that sentence is false. I've
   already blocked the old wording at the framework level so it can't get copied
   forward by an agent that thinks it's brand voice. But what *replaces* it is a
   product decision, and it's yours to write. Anything strong means actually
   building end-to-end encryption.

3. **Migration.** There's no server-side copy of anyone's notes and no way to
   reach into their browsers. The only path is a person exporting and importing.
   I've designed today's backup format to be exactly what the new server will
   accept, so the path exists — but it needs the person to act, which is the
   other reason the notice matters.

---

## 2026-08-10, later — the machine is chosen

You ruled: Mythic Beasts over Hetzner where it's an easy choice, and nothing on
the apis.uk machine except the API.

That made it easy. Four machines: two Hetzner (idea.uk payments, relojistas), and
two Mythic Beasts (apis.uk, webdesign.uk). Your ruling deprioritises the Hetzner
pair and takes apis.uk off the table, which leaves exactly one — **the
webdesign.uk box in Cambridge**. It was already what I'd recommended, so there
was nothing left to weigh.

I logged into it rather than trusting its paperwork, which was worth doing: its
own planning docs specify 4 GB of memory and the machine actually has **7.8 GB,
with 6.2 free**, two cores, and **47 of its 50 GB disk unused**. It's nearly
idle. Postgres isn't on it yet — that's the main thing to add.

**One thing I found while I was in there, which isn't mine to fix.** The
webdesign chat service is listening on every network interface rather than just
the machine's own loopback. Nothing is exposed: the firewall is on, denies
everything inbound except SSH, and I checked from outside that the port genuinely
can't be reached. But that box's own nginx config says, in as many words, that
things bind to loopback so that *even a firewall mistake exposes nothing* — and
the chat service doesn't. So the firewall isn't a backstop for it, it's the only
thing there, and one wrong firewall rule would put a service that spends money on
the Anthropic API on the open internet. Worth whoever owns that lane knowing.

**The one real objection to using that box**, and I want to be straight about it:
it already serves the webdesign.uk shopfront, which is a live commercial front
door, and putting a consumer app with user data next to it couples the two. Our
own track record argues against exactly this — we've twice bought a separate
machine to avoid product coupling, once over a saving of about three euros a
month.

Your ruling points the other way and it stands. But I've answered the coupling
rather than shrugging at it: **the voice recordings and photos won't live on that
machine's disk.** They go to the object storage we already run, and the database
on the box holds text and bookkeeping only. That's better design anyway, and it
removes the only coupling that could actually hurt — noted's storage grows
without limit and the shopfront's doesn't, so a shared 50 GB disk was the real
risk. With media held elsewhere the box runs a small text database that'll stay
small for years.

Written up as a summary in `SUMMARY_2026-08-10b_noted_rebuild.md`.

The two things still waiting on you are unchanged: what the privacy wording
becomes once people can sign in, and the fact that migrating anyone's existing
notes needs them to press export and then import — we hold no copy and can't
reach their browsers.

---

## 2026-08-10, evening — backup plan ready, and the shopfront is on the list

### The backup plan

Written up in `PLAN_2026-08-10_box_backup.md`. Two things in it are worth your
attention before anything else.

**The dumps get encrypted before they leave the box, and the box can't decrypt
them.** This isn't belt-and-braces. Whatever we end up saying about privacy, a
nightly unencrypted copy of everyone's private writing sitting in a third-party
bucket would undercut it, whether or not anyone ever looked. So the box gets only
a public key — it can encrypt and cannot read anything back.

**The cost of that is one thing you must not lose.** If the private key goes
missing, every off-box backup is permanently unreadable, and no support call
fixes it. It needs to live in two places you control, and we should do one real
restore *from your stored copy* before we rely on any of this.

**What I need from you: generate that key.** That's the blocking step; everything
else is built around it.

### I tested the B2 key rather than trusting the documentation, and it found two things

I created a throwaway key, tried every operation against it, then deleted it and
the test file.

**The b2 tool refuses to start without `listBuckets`** — even though the job only
ever uploads. Reasoning it out gives the wrong answer, and the failure would have
turned up at 3:20am with nobody watching.

**And a write-only key can still *hide* files.** After hiding my test file, a
normal listing showed nothing at all — the backup looked deleted. The data was
still there and one command brought it back, but it means I can't honestly tell
you "a stolen key can't touch the backups". What I can tell you is: it can't read
them, it can't delete them, and it can hide them in a way that's recoverable but
looks exactly like deletion. That changes how we monitor — a check that asks "is
today's backup there?" can't tell hidden from missing, so it has to ask
differently.

The good news: the object-lock protection is real. When I tried to delete a
protected file with my *full admin* key, B2 refused. It only worked when I added
an explicit override flag.

### The shopfront — you're right, and here's exactly what's wrong

I've measured it so it's a concrete item rather than a vague one:

- `webdesign.uk` and `www.webdesign.uk` both **redirect to webdesign.co.uk** — a
  different site entirely, served from the bucket rather than the box.
- `preview.webdesign.uk` **works fine** and reaches the box.
- The chat's API endpoint redirects away too, so **the chat is unreachable at its
  own domain.**

So the box itself is healthy — nginx, the chat service and the tunnel are all
serving correctly. The redirect is happening at Cloudflare, in front of the
tunnel, on the zone. That's why `preview` gets through and the other two don't.

One thing worth passing to whoever picks it up: there's an existing note
attributing your "I tested the chat and nothing happened" to a stale cached
JavaScript file. That may be right about the caching, but it wouldn't explain
this — the API endpoint is being redirected before it ever reaches the box, and no
amount of cache clearing fixes that. I haven't worked out which problem came
first, so I'm not saying the earlier diagnosis was wrong, only that this one is
also true today.

**Agreed on the approach**: rather than hand-fix it, put it through the
framework's own checks. That's the better outcome, because this is precisely what
an availability check exists to catch — a site whose public address sends
visitors to someone else's domain — and right now no such check is switched on
anywhere in the fleet. The code exists; its configuration is deliberately held.
Fixing it that way turns one broken shopfront into a check that watches all of
them. I'll follow it through in this thread.

---

## 2026-08-13, evening — the new site exists, end to end (written by the assistant)

Three days compressed: the framework built the whole site, and everything it
built is now on our own machine, though nobody can see it yet — noted.co.uk still
shows the old app with its notice, exactly as before.

What happened, in order. The build you dispatched on the 11th sat in a queue for
seven hours behind other sites' work — nothing was wrong, we were simply last in
line — and then ran by itself overnight: five pages, images, the lot. Then we
found the pages weren't reaching our server even though every check said success;
the sync job had been told to fetch only the shopfront's folder and skipped ours
without saying so. Fixed so it can't happen to the next site either.

The buttons were missing from every page — the framework wrote "Sign in" but
didn't know where signing in happens, since our app lives on its own subdomain.
Pointed them at it. Built the rescue page: it reads the notes still sitting in
people's browsers from the old app and hands them back as a file, without an
account. That's the migration. It's tested against real browser data, including
proof it can't create, change or delete anything.

You wrote the privacy copy (via the draft you approved, with your one edit) and
it's now on a privacy page word for word — after a false start where the writing
agent quietly wrote its own version because my instruction pointed at text it
couldn't see. And your "we don't want to say no server" ruling exposed something
odd: the instruction telling writers not to use the old wording was itself phrased
"the old site had no server", so the writer copied its instructions and got
blocked by your own ban. Reworded to say what the old app *did* — kept everything
in one browser — and the guide rebuilt cleanly first try.

Waiting on you: whether the privacy page should mention that deleted notes
survive in encrypted backups for up to thirty days (I left it unsaid — your
call); the by-hand test that a failed save shouts instead of losing text; and,
when you're ready, the cutover itself.

---

## 2026-08-16 — noted.co.uk is the new site (written by the assistant)

It's done. Going to noted.co.uk now gets the rebuilt site, not the old app.

How it was done safely is worth a paragraph, because it wasn't obvious. The old
app was never really served by "the domain" — a small piece of Cloudflare code sat
in front and fetched the files straight out of the bucket, using its own keys. That
meant everything else could be put in place first, with nobody noticing: the domain
pointed at your machine, the tunnel taught to accept it, all of it inert while that
piece of code still answered. The cutover itself was switching that one thing off,
and switching it back on is a single command — which I wrote down and committed
before touching anything.

The old app is still there, at noted.co.uk/legacy-app/. That matters more than it
sounds: notes live in the browser under the exact web address they were written at,
so the old app had to stay on the same address to still see them. On a different
one it would open and show every visitor an empty screen, which looks precisely
like their notes being gone.

Then I proved the thing everything depends on. On the live site, I wrote a note the
way the old app does, went to the rescue page, and it found it — text intact,
voice recording intact, in a file the new service accepts. Signing in, writing,
saving and reopening from a second browser all work on the new address too, and a
save still fails loudly with your text untouched when the connection dies.

The shopfront was checked either side and never moved. Nothing was lost.

What's left is watching rather than building: give it a few days, see whether
anyone needs the old app, and decide when to retire it. The one loose end I'd
flag is that the smoke tests leave throwaway accounts behind, because the service
has no way to delete an account — worth adding before real people sign up.
