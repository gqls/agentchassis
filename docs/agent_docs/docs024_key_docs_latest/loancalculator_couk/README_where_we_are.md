# Where we are — loancalculator.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-30, afternoon.** Started by just looking at the two folders — the source
one in `domains/` and the deployed one in `sites/`. They're the same site: the first
builds it, the second is the copy that gets published. Nothing has touched either
since the 20th of March.

Then I checked whether the site was actually up, and it wasn't. Worth explaining the
symptom because it's a misleading one: the domain resolved fine, the padlock was
valid, the request went out — and then nothing came back at all. It just hung until
it gave up. No error page, no 404, no 500. That looks like a network fault, but it
wasn't.

Every working site of ours has a tiny programme running at Cloudflare's edge that
fetches pages out of our storage bucket. I checked it directly (there's a
`/worker-health` address that answers "Worker is running!" if it's there) —
gamesdesign.co.uk answered it, loancalculator.co.uk didn't. So the edge programme was
never connected to this domain, and requests were still being sent to the old Amazon
storage location from the original March setup, which is long gone. That's why it
hung: the proxy was alive, the thing behind it wasn't.

You fixed that this afternoon by adding the missing route, and the site is back —
confirmed at 16:11, all the pages I tried return properly.

**Then the more interesting bit.** You asked for this to go in through the framework
so it's managed properly, and then for anything the framework's own adoption process
would have got wrong to be reported or fixed. So I read that process carefully, and it
would have done three things we very much don't want.

It would have changed every web address. All our addresses look like
`/tools/standard-calc.html`; the adoption code rewrites them into
`/tools/standard-calc/index.html`. It doesn't matter what the real address was — it
throws it away and builds a new one. That's all 28 addresses broken.

It would have rewritten the calculators. Every page gets marked "recreate", and pages
with interactive bits get handed to a language model to rebuild. Twelve calculators
that work, that do correct arithmetic, regenerated from scratch. Even if it went
perfectly it's a pointless risk; realistically some of them come back subtly wrong.

And the setting that's supposed to prevent exactly this doesn't do anything. There's a
"fidelity" option — how faithfully to keep what you found — and you can pass it, and
it gets written into the record, and then nothing on earth reads it. I checked the
whole codebase rather than trusting the comment that says so: ten mentions of the
word, every one of them about something unrelated. So the promise is in the interface
and the implementation was never written.

**What I want to do about it.** Not work around it for this site. If I hand-port this
one, the site is fine and the framework is exactly as unable to adopt the next one.
So the plan is to build the missing faithful-preservation mode properly — make
"fidelity locked" mean "keep the addresses, keep the bytes, don't regenerate
anything" — and then use it on this site as the proof. It's two focused changes, both
opt-in, so no existing site can be affected either way. They'll go through the review
council before I call them done.

**One thing I found that you may not know about.** The site has real problems of its
own, and they've been live since March. Four pages aren't really pages — they're
loose fragments with no page structure, so they arrive as unstyled text with no
navigation. Ten of the 28 pages have no menu at all, and that includes the main loan
calculator and the overpayment tool, so anyone arriving on those from Google has no
way to get to anything else. The menu that does exist has a link to a calculator page
that doesn't exist, and it's on every page that shows the menu. And the main
calculator has been loading its stylesheet from the wrong place, so it's been
unstyled this whole time.

I'm going to fix all of that first, before adopting, for a slightly non-obvious
reason: the framework learns the site by reading the live pages. If I adopt first, it
faithfully preserves the broken version and we've frozen the bugs in.

**Order of work from here:** fix the site where it sits and publish it; build the two
framework changes and get them reviewed; run the real adoption process with the new
mode; check every page and every calculator still works, including that a later
rebuild can't quietly change them back.

**Something for you to decide, but not now.** The written content quotes interest
rates and a "last updated" date from March. I'm leaving those alone — changing the
copy isn't part of adopting the site, and I'd rather not mix the two. But once it's
managed, keeping figures like that current is exactly what the platform could be
doing for us, and that's worth a conversation after this lands.
