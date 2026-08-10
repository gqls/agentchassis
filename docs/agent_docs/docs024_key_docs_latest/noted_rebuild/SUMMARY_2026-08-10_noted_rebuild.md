# noted.co.uk — first read-out

2026-08-10. Written to be read aloud.

---

## What we're trying to do

noted.co.uk is a note-taking app: you type a note, you can record your voice or
add a photo, and it keeps a few old versions in case you delete something by
accident. It was built by hand and it works.

Two problems with it. The first is that everything a person writes lives only in
the browser they wrote it in — clear your browsing data and it's gone, and
there's no way to pick up on your phone what you started on your laptop. The
second is that it isn't part of the framework at all, so none of the machinery
we've built for making sites good — the checks, the upgrades, the maintenance —
touches it.

So: rebuild it on the framework, move the notes onto a server on one of our own
machines so people can sign in and find their notes anywhere, and have the
framework genuinely own every part of it.

That last phrase got sharpened partway through the day. The owner's word is
**decomposition**: the site broken down into the framework's own parts, not
swallowed whole as a few opaque lumps. That distinction turns out to be the
whole design, and it rules out the quick route.

## Where we've come from

The app has been live since January, in three generations, gaining voice notes
and then sharing. It was uploaded by hand straight into the storage bucket.

The thing nobody had noticed is that this is the framework's *own* bucket — the
same one every framework site deploys to. So noted has been sitting inside our
delivery system all year while being invisible to it: no database record, not in
any repository, not mentioned anywhere in the codebase.

## What we've done

**We made the existing app safe before touching anything else.**

The owner asked for a notice telling people to back up before the rebuild. Before
shipping it, we checked what the backup button actually saved. It saved the text
and nothing else — no voice recordings, no photos. We proved it rather than
assuming it: put a note with a real audio clip and a real photo into the app,
pressed Backup, and got a file with the words in it and nothing else.

That mattered enormously. The notice on its own would have sent people to a
button that quietly loses the one thing they can't recreate. You can retype a
note. You cannot re-record someone's voice.

So the notice went up **with** a fix: a "Save everything" button that includes
the recordings and photos, and a Restore that understands the new file. The whole
round trip is tested — save, wipe the browser completely, restore, and the audio
and images come back identical, down to the file type.

Two other things came out of the same pass. The app is now in version control for
the first time, so a change to it can be reviewed and undone. And the framework's
own contrast checker found four accessibility failures, three of which had been
there all along; all four are fixed and it now reports a clean sheet.

**We brought the domain into the framework.** It now has a database record and
its foundational specs, including a claims policy that blocks the old privacy
wording — more on that below.

**And one mistake, recorded.** Partway through, I convinced myself the live site
was crashing on startup. I'd read a comparison backwards. Running the actual
page in a browser refuted it in seconds. It's logged in the fleet's wrong-calls
file, because the false alarm felt *more* certain than the truth did — it was the
more interesting story. The real defect turned out to be one layer below the
imaginary one, and only running the thing found either.

## Where we are now

The live app is unchanged in how it works, now warns people, can genuinely save
everything, and is under version control. All of that is deployed and verified on
the real site, not just locally.

The framework knows the domain exists and has its specs. Nothing has been
rebuilt yet, and that's deliberate — building it now would have deployed straight
over the running app, because the default deploy path syncs with deletion onto
exactly the place noted lives. That trap is disarmed in the seed and written up
as a landmine, because it's invisible: the deploy would go green and the listing
afterwards would look exactly right.

Three things are worth saying plainly about what comes next.

**The framework cannot write the server.** Every backend we run — the watches
site, the payments box, the web design chat — is hand-written Go, and nothing in
the framework generates one. The one component we have with user accounts is for
platform operators and isn't reachable from a public site. So the notes server
will be written by hand, the way the others were. Everything *else* — the pages,
the editor, its behaviour, its checks — can be framework-owned. Being straight
about that boundary now is better than discovering it in three weeks.

**The product's central promise is about to become false.** Today the site says
"we can't see your notes, read your text, or listen to your recordings." That's
true because there is no server. The moment you can sign in from another browser,
it isn't. We've blocked that wording at the framework level so it can't be
carried forward by an agent that mistakes it for house style — but what replaces
it is a decision about the product, not about the copy.

**We don't know how many people use it.** Access logging stopped in May, and even
the older records can't tell a person from a bot. So we've designed as though
there are users, because nothing available can show there aren't.

## Where we're going

Pick the machine — the recommendation is the web design box, which has the room
and the safest network setup. Put a database on it. Write the notes server,
including an import that accepts exactly the backup file we shipped today, so
existing users have a path across.

Then, before rebuilding the app, write down its behaviours as contracts the
framework can test: signing in and finding your notes, capturing a thought while
offline, attaching a recording, getting your data back out. That mechanism
already exists and is what makes "the framework checks it" a real claim rather
than a hopeful one.

Then build the site through the pipeline properly decomposed, and only then move
the domain across — keeping the old app reachable for a while, because the notes
in people's browsers are still the only copy that exists.
