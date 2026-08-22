# Where we are — live-object declaration drift

Plain prose, append-only, newest at the bottom. The owner's document.

---

## 2026-08-22 — asked to fix bug 317; it was already fixed, and something else turned up

I was asked to fix bug 317. It isn't in the open pile any more — another session fixed it three
days ago and filed it away yesterday.

I checked that rather than believing it. There are three things that had to be true, and all
three are: the live database setting that the bug was about now has the right fourteen entries
in it, those fourteen are exactly the fourteen the code says they should be, and the automated
guard that keeps the two in step passes when run against the committed code. So 317 is
genuinely done, and I have not touched it.

While checking, I found something else, and it is worth explaining because it is a shape rather
than a single fault.

We keep some of the platform's behaviour in the database rather than in code — a scheduled job's
query, a database trigger, a rule about which values a column will accept. Because that
behaviour is not in the code, an ordinary test cannot see it. What we have done instead, in
several places, is write a test that reads **the migration file** — the file that was run once,
long ago, to put that behaviour into the database — and check the code agrees with the file.

The problem is that a migration file is a receipt, not a description. Our own rule, and a good
one, is that you must never edit a migration after it has been applied, because the system
records a fingerprint of the file and editing it would make that record a lie. So the file is
frozen at the moment it was written, while the live setting carries on being changed by every
later migration. The test therefore checks the code against a photograph, and passes whether or
not the photograph still resembles the thing.

The example I hit makes it vivid. The guard for 317 reads a migration file and pulls the list of
fourteen values out of it — but in that file the list only appears **inside a comment**. It has
to, because the migration edits the live setting by find-and-replace rather than writing the
whole thing out. So the specification our test checks against is a sentence somebody typed in a
comment. And the live setting has since been changed again by a further migration whose filename
the guard doesn't even look at.

Nothing is broken today. I checked: the file and the live setting still agree, in this case and
by luck rather than by machinery. What is missing is anything that would tell us if they stopped
agreeing. I looked for such a thing specifically and there is none — the migration system records
a fingerprint of the file and never looks at the live object, and nothing anywhere in our code
asks the database what its triggers actually contain.

It is not one guard, either. I counted seven places doing this, covering scheduled jobs, two sets
of database triggers, a column rule, and a workflow definition.

So this is the lane: work out the general fix rather than patching the one guard. I have sent it
to the diagnosis loop for an independent read, and asked our planning model for a design. My own
view before either comes back is that the fix cannot be a test, because tests must not need a
live database, and it cannot be a commit hook, because at commit time the migration hasn't been
applied yet — we already ruled on exactly that point in a different case a few weeks ago, and the
answer there was a daily check running against the real system. I expect we land somewhere near
that, but I would rather see the plan than assume it.

One honest note: I got something wrong on the way and it is written up. I counted the values in
the code two short of the live list and briefly thought the guard was failing. I had grepped for
one spelling of a function call when there are two. The lesson is small and general — I measured
my guess about the interface rather than the interface — and I have logged it.
