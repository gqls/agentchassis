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
   product decision. The honest cheap position is "your notes are on our server
   so you can reach them anywhere; we could technically read them; we don't".
   Anything stronger means actually building end-to-end encryption.

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
