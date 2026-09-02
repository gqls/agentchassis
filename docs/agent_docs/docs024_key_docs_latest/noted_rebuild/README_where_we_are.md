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

---

## 2026-08-17 — the old app is retired (written by the assistant)

You asked to retire the old app and remove the notice, and it's done. The facts
made it an easy call: in the whole grace period, the only visits to the old app
were my own tests — not one real person went there — and nothing on the new site
pointed at it.

Anyone who does turn up at the old address now lands on the rescue page instead,
which is the right place: it finds the notes still sitting in their browser and
hands them back as a file. I re-proved that end to end after the retirement —
wrote a note the old way, and the rescue page found it, text and recording intact.

Nothing was deleted. The old app still exists in three places — on the server's
disk, in the source repository, and in the bucket — it just isn't served any more.
The "being refreshed" notice is gone from everywhere a visitor can reach, which is
right, because the refresh has happened.

---

## 2026-08-24 — photos, video and audio in a note: built and tested, one step from live (written by the assistant)

You asked (on the 22nd) for images, videos, GIFs and audio all in one note, with
a sort of pasteboard to paste them onto, move them around, even edit them — in
stages. The stages are written down now: first get every kind of file into a
note and safely back out; then the board you arrange them on; then editing in
place. The first stage is what got built today.

What it does: on the Write page you can paste a picture straight from the
clipboard, drag files in, or use a button — photos, GIFs, video clips, voice
recordings. They appear on the note (pictures shown, video and audio playable),
each with a remove button that asks first. The same honesty rules as the text
apply: nothing ever claims to be stored until the server has actually said so; a
failed upload stays visible with a Try again; closing the tab warns you if
something hasn't finished. A little line shows how much of your storage you've
used. And a nice side effect: anyone who imported from the old app finally SEES
their old photos and recordings — until now the service kept them but showed
them nowhere.

All of it is tested the strict way this page has always been: the checks were
deliberately broken to prove they can fail, then restored. The service tests all
pass too, including one that proves one person cannot delete another's files.

Where it stands: NOT live yet, deliberately. The service on the box has to be
updated before the page, and this session wasn't permitted to copy files onto
the box — so the update needs your hand (the exact commands are in the runbook,
two blocks, five minutes). Once the service is updated I fire the page update
through the framework and re-run the live checks end to end.

Two numbers are yours to set when you like: each account gets 50 MB for files
and each single file can be at most 25 MB. Fine for photos and voice notes,
small for video. Both are one-line settings on the box — but the 50 MB is also
what protects the shared disk under the webdesign shopfront, so raising it is
your call, not ours. And the missing "delete my account" button matters a bit
more now that accounts can hold real files.

One piece of tidy news: the formal review of the Get Started button change came
back approved, so that thread is closed.

---

## 2026-08-24, evening — media notes are LIVE (written by the assistant)

You ran the two server commands this afternoon, and everything after them went
through cleanly: the service updated, the page updated through the framework,
and the rebuild queue delivered the new editor page about two and a half hours
later.

Then the full live rehearsal ran against the real site, and all fifteen checks
passed. The one worth saying out loud: a picture was uploaded through the real
page, a completely separate browser signed in as the same person and got the
actual picture back — decoded, not just a broken image tag — and then removed
it, with the removal confirmed by the server. Save-honesty still holds
everywhere: pull the network cable mid-save and the page says so loudly, keeps
your words, and recovers on retry.

So: noted.co.uk now takes photos, GIFs, video clips and audio in a note, today,
on the live site. Next stage when you want it: the pasteboard itself — arranging
what you've pasted, moving things around — and after that, editing in place.

---

## 2026-08-25 — the pasteboard is live, and files now live in Backblaze (written by the assistant)

Everything you asked for yesterday morning is on the live site this afternoon.

Storage first: photos, clips and recordings now go to a private Backblaze
bucket instead of the server's own disk, which is what makes "pay a little for
more space" a settings change later rather than a rebuild. Your storage
allowance stays at 50 MB for now, as you said. One find on the way: Backblaze
had quietly retired the API version everything documents, which we only caught
by testing the real service before trusting our own test copies — the code now
speaks the current version and was proven with a real upload, download and
delete against the real bucket.

The board: every note now has an "Arrange on board" button. Your text and each
photo, clip or recording become tiles you can drag around and resize — by
finger on a phone (built for touch first, as you asked) or by mouse. The
arrangement saves with the note under the same honesty rule as everything else:
nothing says Saved until the server has actually saved it, and an arrangement
can never be wiped out by an older page that doesn't know about boards.

The whole thing was rehearsed live after deployment: fifteen checks passed on
the real site, including a picture uploaded through the page landing in the
real Backblaze bucket, read back and displayed by a second browser, then
removed — with the bucket confirmed empty afterwards, so nothing leaks.

Also done today: the review-board's advisory points on the Get Started change
are all answered (the one real gap — a config change that had no ledger entry —
now has one), and the account-deletion plan is written and waiting on your two
choices: delete immediately or after a grace period (I recommend immediately —
it's what the privacy page already promises), and the one honest sentence about
backups for the privacy page.

---

## 2026-08-25, evening — you can close an account now, caption things, and edit photos (written by the assistant)

The three things you approved this morning are all live tonight.

Closing an account works, properly: you type your password again, confirm, and
everything goes — files first, then the records, and the page only says so once
it is actually true. If even one file cannot be removed, nothing at all is
deleted and it tells you plainly. The live rehearsal proved the whole journey,
and as its final act it deleted its own test account and showed the sign-in no
longer works. The four old test accounts from earlier rehearsals were then
removed the same way — through the real button, not a database shortcut — so
the only account left on the service is yours.

Captions: any photo, clip or recording can carry a line of text, shown in both
views. And photos can be edited — rotate and crop, on your phone or desktop.
An edit never touches the original until the edited copy is safely stored; if
anything goes wrong you keep what you had. GIFs politely refuse editing for
now, because a naive edit would silently lose the animation — that belongs to
the fuller editing surface we sketched from your note about an "editor" type
of note, which is written into the plan for when you want it.

Two pieces of wording wait on you in the chat: the honest backup sentence for
the privacy page, and the paid-storage copy (held back until there is actually
a way to pay, so the site never offers a door that isn't there).

---

## 2026-08-25, late — the honest backup sentence is on the privacy page (written by the assistant)

Your approved wording is live, word for word — checked sentence by sentence
against the whole page: all twenty-six of them, twenty-five on the public page
and the one containing the email address verified on the server itself, since
Cloudflare disguises addresses at the edge. The 30-day figure is also now a
registered fact in the site's evidence base, so the machinery that polices
claims knows where that number comes from. The paid-storage copy is written,
approved and waiting in the plan for the day there's a way to pay.

---

2026-08-25, late night. Picked up the one piece of homework we owed ourselves
from tonight's handoff. Back on the 18th, when we gave you the ability to point
the site's header button wherever you like, the reviewers accepted it but one
of them made a fair point: if you ever set that button to a page that doesn't
exist, the system quietly falls back to its own choice and the only trace is a
line in a server log nobody reads. Your instruction would be ignored and
nothing would ever tell you.

That's fixed now. If your choice is ever refused, the system files a note in
the review queue addressed to a human, saying what you asked for, what it
served instead, and the three ways to resolve it. The note always describes
the CURRENT situation — if you change the button again and that's refused too,
the note updates rather than going stale. The site itself never breaks over
this; the button keeps working on the fallback in the meantime.

The change went through the review council and was approved first time. The
reviewers suggested two genuine improvements — one we adopted (a future-
proofing detail in how the notes are filed), one we answered with a check
(we confirmed this was the only place where an instruction of yours could be
silently ignored this way). Nothing here is live on the site yet; it rides
the next routine platform release.

Small bonus while a review was running: your open question about whether mail
to noted@contactforsales.com reaches anyone. The domain does have a working
mail service behind it, so mail won't just bounce off a dead domain — but
whether that particular address lands in a mailbox someone reads, only your
one test email can prove. That question is still yours.

---

2026-08-26, mid-morning. The platform's automatic housekeeping woke back up
today after a fortnight off, and noted was one of the first sites it visited.
Mostly this is fine and working as designed: it re-checked the site top to
bottom, and I verified afterwards that everything you care about is untouched
— the editor passes all its live checks, the privacy wording is in place, and
the queued "re-render" jobs would simply reproduce the current site.

Two things you should know. First, the site briefly lost its Google
Analytics tag early this morning — a re-render rebuilt the page furniture
from source and the tag had only ever been patched into the finished page,
not the source. The analytics team spotted it within hours and is putting it
back the durable way. Nothing else was lost.

Second, a decision that is genuinely yours: noted has never had a favicon
(the little browser-tab icon), a logo image, or a link-preview picture — the
pages point at image addresses that serve nothing, and they always have. The
housekeeping noticed and is queuing up MACHINE-GENERATED artwork to fill
those slots. If you'd rather choose your own icon and preview image for your
own product — or want nothing there at all — say so and we'll put your
choice in before the machine invents one. This is cosmetic either way; the
site works regardless.

One more piece of automation wants to rewrite the note editor itself to fit
a platform-wide code convention. The rewrite instructions are careful — they
insist on preserving every behaviour we built — but I've asked the team that
owns that machinery whether a tool like ours can be exempted, since the
convention solves a problem our editor can't have. Nothing will change
without the editor's full test suite being run against it.

---

2026-08-26, later. The editor-rewrite story turned out to be better and worse
than it looked. Better: nothing ever failed — the platform's convention
converter actually succeeded twice. Worse: each time it did, our own editor
updates overwrote its work minutes or days later, so the platform concluded
its conversion "keeps failing" and queued up a full rewrite of the editor in
response. Two machines politely undoing each other, each reporting success.

I've paused the three queued jobs that would touch the editor (legitimately —
parked with the reasoning attached, not deleted), written the whole story up
for the team that owns the convention, and offered them the clean way out
either direction: exempt our editor, or we'll apply their convention
ourselves, in our own source file, with the editor's full test suite run on
the result. One of those is a question only you can settle if the teams
can't: does a platform-wide code convention apply to the one tool that is
itself the product? My view, for what it's worth: exempting it is the honest
answer — the convention solves a problem this editor cannot have.

---

2026-08-26, evening — owner rulings on the seven open items (chat, recorded
here so they don't live only in scrollback):

1. Brand assets: LET THE AI GENERATE them. The queued jobs run; we watch
   that the results actually land (three icon deploys have already claimed
   success without the files appearing at the referenced paths).
2. Launch: real users CAN be invited. Timing of actual invitations is the
   owner's act.
3. Paid tier: GO AHEAD — £9.99 per month per terabyte, subscription — with
   the details to be discussed with the owner before building (discussion
   opened in chat; the approved copy stays held until the mechanism exists).
4. Contact email: owner will test it himself later.
5. Backup key: owner will copy it to a second place; the drill re-run from
   that copy, then the workstation deletion, follow after.
6. Editor-note vision: DO NOT PROGRESS. The owner first wants to determine
   what "noted" actually means — possibly a move towards recording anything
   in realtime and searching it easily (embeddings, visual embeddings,
   CLIP-style), possibly a different end game entirely. Stage-3 continuation
   stays parked until that direction discussion happens.
7. Editor convention question: fuller explanation requested and given in
   chat; no ruling yet.

---

2026-08-26, evening (second ruling round). Two more owner calls from chat:

- **The paid tier is HELD until the 25 MB file-size blocker is fixed, and
  that fix is the next build.** Investigation already done: the limit is not
  just our setting — uploads travel through Cloudflare, which caps any single
  request at ~100 MB regardless of what our server allows. So big files mean
  a chunked upload: the editor slices the file, sends pieces, and the engine
  reassembles them into B2 storage piece by piece, never holding the whole
  file. Plan being written; the tier copy stays held meanwhile.

- **The sweep-vs-tools problem gets a FLEET-WIDE solution, not a one-off
  exemption for our editor.** The owner wants a general mechanism so the
  platform's automated sweeps stop recreating tools that a team ships and
  maintains from its own source. Direction recorded with the convention
  team's lane; the paused editor jobs stay paused until that mechanism
  exists.

---

2026-08-26, afternoon. Your GitHub error: it was GitHub itself having a
wobble, not anything we pushed — their own error message says so, the
workflow definition hadn't changed in three weeks, and runs were succeeding
again nine minutes later. Better still, that workflow doesn't deliver
noted.co.uk anyway: your box pulls the site from the repository directly
every five minutes, and the proof it worked end-to-end is that the analytics
tag lost this morning is already back on the live page. The push you saw was
the platform's own re-render wave putting that tag back.

While checking I found one leftover typo in the page's styling from the
August css-patch saga — measured carefully: it currently breaks nothing
(the piece it damages is an exact duplicate of a piece that already loaded),
so I've recorded it with the evidence rather than poking at a surface other
repair work owns. Nothing for you to do.

---

2026-08-26, evening. The big-files work is done and live on the editor side:
the page now knows how to send a large file in slices, proved by its full
test suite (75 checks) and the live smoke (all 17, still green). Nothing
changes for anyone until the server half is installed — that's the two
commands waiting for you in the runbook, and after them one test upload
proves the whole chain. The paid tier stays parked until you've seen that
proof, exactly as you ruled.

One small operational note: your site's queued rebuild sat two hours behind
other sites' backlogs today — a scheduling quirk another team diagnosed this
very morning — so I used the documented direct route instead, and left the
measurement with that team.

---

2026-09-02. You ran the two server commands and the big-files machinery is
now fully in place: the server speaks the new protocol (proved by a
disposable test account that saw the new capability, then deleted itself),
the site's other tenant was untouched (byte-identical before and after), and
all seventeen live checks pass. Nothing changes for any existing account
until we deliberately raise its limit — that's the paid tier's switch, still
parked for your pricing decisions. The one proof left, whenever you feel
like it: the real ~35 MB upload described at the end of the runbook.
