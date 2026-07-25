# Where we are — the machines we run things on

Plain-prose running log for the VM estate, append-only, newest at the bottom.
This is the owner's document: add to it, never rewrite it.

---

**25 July 2026 — opened this on your instruction to bring setup.sh under the framework
and to stop running the tools-api box as a separate project.**

I read the whole setup script rather than skimming it, and I want to say first that it's
good work. It asks no questions, so a machine could run it; running it twice is safe, and
that's how you rebuild a box; nothing about a particular host is baked in. There are
touches in there that only get written by someone who's been burned — it refuses to turn
off SSH passwords unless it can first find a key to let you back in, and it deliberately
leaves out one HTTP/2 setting because the correct spelling changed between nginx versions
and this way the file works on both.

The cleverest part is how it handles certificates. You can't get a certificate until the
site answers on plain HTTP, but you can't serve HTTPS until you have the certificate. So
it writes the web config once without HTTPS, fetches the certificates, then writes the
whole config again — and this time each domain that succeeded gets its HTTPS block. If one
domain's certificate fails, that domain stays on HTTP and the run carries on rather than
dying.

**I found one real bug and fixed it.** In the certificate loop there's a line that uses a
bash keyword only allowed inside a function. Outside one it's an error, and because the
script is set to stop on any error, it would kill the run. The good news: it can only
happen for a domain that doesn't have a certificate yet, so **your pending relojistas
re-run is unaffected** — that domain has had a certificate since July. The bad news is
that it breaks precisely the thing the file's own instructions tell you to do: add a new
domain to the list and re-run. I've fixed it and left a note explaining why the line is
written the way it is.

Two details worth knowing. The bug is *not* in the older idea.uk copy of this script — it
crept in when the script was copied and changed. And the two copies have now drifted so far
apart that they share 61 lines and differ on 614. They aren't one script kept in two places
any more; they're two scripts with a common ancestor, each with its own bugs. The tools-api
box is quietly starting a third lineage of the same kind.

**The second problem I found I deliberately did not fix.** The script gives *every* domain
on the box the relojistas-specific bits — the old forum's feed addresses and the search
page. Today that's harmless because only relojistas lives there. The moment a second site
lands, it inherits a dead watch forum's URL scheme. I could add an if-statement, but that
would make the current design a bit more entrenched, and the design is the thing you've
asked me to replace.

**Here's the argument for framework control, in one sentence.** The web config on that box
is the same kind of thing as a web page's HTML — and we already decided, in writing, that
editing the HTML directly is not allowed, because the next legitimate rebuild wipes it. In
July someone hand-edited the live config on the box to fix the feed; the generator didn't
know, so the next run would have deleted the fix, and last week I had to reconcile it by
hand. That's the exact manual repair we refuse to accept for pages, happening unchallenged
on machines. If the config is instead produced by the framework from what the database
knows, the per-site problem disappears on its own, drift becomes something we can detect
rather than notice, and "one box session" stops being a step in a plan that only you can
perform.

**The good news on cost:** almost none of this needs inventing. The platform already
provisions machines and runs commands on them — that's how GPU training boxes work today.
What's missing is the same thing pointed at ordinary servers, plus something to generate
the config, which has a close cousin already (the thing that writes the RSS feed reads the
database and emits a file — this reads the database and emits a web-server config). The
script's own header has promised for months that a "service deployer" would take this over;
I checked, and no such thing exists in the code. This would be that.

**On merging with the tools-api island — one thing I want your decision on.** I can merge
almost all of it: one description of what a box is, one generator, one drift check. But
that island was built so the production cluster appears nowhere in its path and it holds no
production credential, and the obvious version of "framework controlled" — our cluster
holds an SSH key and pushes config to it — undoes exactly that. My recommendation is to
merge the generator but not the trust boundary: the public boxes get pushed to as normal,
and the island *pulls* its own config outward, the same direction it already dials to
Cloudflare. It stops being hand-maintained without giving up the isolation it was built
for. If you'd rather have everything under one control path, that's a legitimate call —
it just spends something we paid for.

Nothing is built. What exists so far is the plan, the walkthrough, and one fixed bug.
