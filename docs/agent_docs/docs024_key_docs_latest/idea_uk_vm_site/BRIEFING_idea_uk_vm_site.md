# idea.uk → the VM: a full briefing

**Written to be read aloud. Updated 2026-07-19 — the migration is live.**
Companions: `HANDOFF_RESUME` (start a fresh chat), `RUNNING_NOTES` (chronological record + missteps),
`RUNBOOK` (how), `PLAN` (design record), `box/` (server scripts), `sql/` (every change, in order),
and `/bugs_open/016, 017, 018` for the open faults.

---

## 1. In plain terms

idea.uk used to exist as two halves that had never met. A marketing website — nine pages, built
automatically by our platform — was published to cloud storage where **nobody could see it**, because
the domain pointed somewhere else. Meanwhile the thing that earns money, a small tool selling a £29
report, ran on a rented server in Germany, and that server showed only the tool. A visitor to idea.uk
saw the tool's single built-in page; the nine-page site sat in storage like a rehearsal that never
went on stage.

**That is now fixed.** As of the 18th of July, idea.uk is one complete site: the marketing pages and
the paid tool, behind one front door, on one server. The money path was never taken offline.

Getting there also turned up things nobody had been looking at — including one serious problem that
had been hiding in plain sight for as long as the site has existed. The honest headline is: **the
migration succeeded, and it revealed that the site underneath it is in worse shape than anyone
knew.** That is a good outcome. It was always true; now it is visible and fixable.

---

## 2. What we set out to do

Two goals at the outset. First, publish idea.uk's automatically-built pages onto the German server
instead of into invisible storage, so the site is whole and still includes the paid tool. Second, a
list of loose ends: three catalogued pages that had never been built, and spam arriving through the
tool's enquiry form. A third was added during the work — record the reasoning properly, so whoever
picks this up inherits the decisions and the traps rather than just the code.

Two facts shaped everything. There will eventually be **thousands** of these domains, so nothing
built for idea.uk was allowed to become a special case that could not scale. And the existing
publishing machinery already treats each domain as a folder in a shared repository.

---

## 3. The decisions that shaped it

**We kept the site static; the server is a second destination, not a second factory.** Our platform
turns data into finished pages inside our cluster, where the data lives. The server simply *serves*
those files. Building on the server would mean either giving every server a dangerous line into the
central database, or rebuilding the whole rendering machine on each box.

**We put the whole site behind one front door.** One certificate, one DNS record, one origin — so
forms, links and cookies simply work. It also delivers a quiet reliability win: the pages are just
files with no application behind them, so a crash of the paid tool can no longer take the marketing
site down. Before this, it could.

**The server pulls its own files; we never push into it.** Each server fetches what it needs on a
schedule, rather than something in our cluster reaching in. If a server is ever compromised, it holds
only a read-only key that can fetch from a repository — it cannot write anywhere or reach its
neighbours. The alternative would have put one key, able to write to every server, on a machine in
our cluster.

**This stays the exception.** For the thousands of ordinary sites, publishing to cloud storage
remains the default. The server model is reserved for the handful that actually sell something.

---

## 4. What we've done

**Closed a live security hole.** While mapping the tool we found real credentials — for sending email
as idea.uk, and the tool's internal master key — sitting in a **public** code repository, and they
had been there about six weeks. We judged them by the length of each value rather than its name,
which is the only reliable tell, and confirmed the payment keys were only placeholders, so **the
payment path was never exposed**. We cleaned the file and installed an automatic guard that blocks
any future credential from being committed. But cleaning a file does not undo history — the old
values stayed readable in the repository's past. **Rotation is what actually closes it, and that is
done:** a fresh sending identity, the old one deleted outright, the internal key regenerated. The
values still visible in history are now dead.

**Completed the site.** Three catalogued pages had never been built, so their navigation links led
nowhere. Building them exposed a genuine trap in our own planning machinery: the documented "safe"
way to fill in missing pages silently discards the design of pages already built. We caught it after
it had damaged four finished pages, recovered them exactly, removed ten pages the planner had
invented unprompted, and wrote the trap up so it cannot bite the next person.

**Moved the site onto the server.** The mechanism that lets one site publish to a server instead of
storage had been designed months ago but never actually connected. We connected it — and then found
something more fundamental: **the publishing job for that repository had never successfully run in
its life.** It had no machine registered to run it, and even once we provided one, the machine's
image was missing the tools the job needs, so it had been failing invisibly with its error output
sent nowhere. We rebuilt the image, stood up a dedicated machine, and proved the whole chain. Before
switching idea.uk on, we also replaced a job that would have pushed its files onto **the wrong
server** with an explicit map of which domain belongs to which machine.

**Made the server serve the site.** A scheduled job on the box now fetches idea.uk's pages every five
minutes. Then the front door was switched over: the marketing pages became the site, while all
sixteen of the tool's own paths continue to reach the tool. We checked that number carefully against
the running tool, because an earlier draft listed only seven, and the gap would have silently broken
the free taster and the operator's approval flow.

**Fixed the tool's spam problem and an email that was being blocked.** The enquiry form had no
defences, so we added the standard set — a hidden honeypot, a too-fast-submission check, rate
limiting, proper validation. Separately, the owner noticed the important email of an order was being
flagged as spam. Reading the blocked message's own scoring report showed the cause exactly: the tool
was putting the customer's entire multi-hundred-character description **into the subject line**,
which reads as keyword-stuffing. Fixed, with tests. Both changes are built and waiting for one deploy.

**Restored the funnel after the cutover broke it — see below.**

---

## 5. Where we are now, honestly

The migration is **done and live**. The site and the tool are one thing at one address, the server
keeps itself up to date, and rolling back is a one-line change that never touches the tool or its
order data.

Two problems surfaced once real eyes were on the live site.

**The first we caused, and have fixed.** The tool used to serve its *own* landing page, and that page
carried the actual forms — the free audience check and the report request. When the marketing site
took over the front page, those forms went with it. The tool was left running, reachable, and
completely unusable: there was no way to buy anything. Every automated check passed, because nothing
was erroring — the funnel was simply absent. We have now authored both forms properly as parts of the
site, with the field names read from the tool's own source rather than guessed, and verified them
working on the live site.

**The second we merely revealed, and it is worse.** The site's shared furniture — the header, the
navigation, the footer — renders with **almost every link empty**. Of thirty-three links on the home
page, thirty-one go nowhere: the entire navigation, every call-to-action button, all the social
links. The logo image has no address either, which is why it shows the words "idea.uk logo" instead
of a picture. **The site is, in practice, unnavigable.** This is not something the migration broke —
it has almost certainly been true since the pages were first built. It stayed invisible because the
pages were published to storage nobody ever looked at. The migration's real service here was to make
it visible. It is written up, not yet fixed, and it is the top job.

We also learned that our automated quality checks — which do exist, and are good — **had never once
been run against this site**. When finally pointed at it, they found the header and footer problems
immediately, and correctly identified them as shared-furniture faults rather than page-by-page ones.

---

## 6. Missteps worth admitting

These cost real time, and two nearly shipped defects. They are recorded in full in the running notes.

- **We diagnosed from an error message about a file path as though it described the file's contents,**
  and "fixed" a file the system was never going to read. The real cause was a subtlety of how secure
  connections find their keys. One diagnostic command would have ended it in seconds.
- **We copied a fixed script to the server, ran the old one by mistake, and read the identical failure
  as "the fix didn't work."** A directory-copy command nests when the destination already exists.
- **We nearly installed a server configuration that was a silent downgrade** — it dropped a setting
  that allows the report engine several minutes to run. Every test would still have passed, and
  reports would have died after sixty seconds. Caught by reading the script that originally built
  the server rather than assuming.
- **We cut over without the dry run we had written for exactly that purpose.** It came out clean —
  but that was luck, not method.
- **We treated a set of green checks as proof the migration worked,** while the funnel was gone.
- **We told the owner the automated checks had missed the faults. They had not — they had never been
  run.** And we wrote a duplicate tool before searching for the one that already existed.
- **We nearly shipped a security defence that would have looked present and done nothing** — a
  timing check whose supporting code would never have been published. Caught by checking our own work
  against our own written guidance before deploying.

The thread running through most of these: **a check that cannot see the thing it is asked about
returns a clean answer.** A status code cannot see a missing form; a search for the wrong tag cannot
see a footer; a symbol grep cannot see that the symbol is the wrong one. Our own guide says "zero
results is not decisive" about database queries — it turns out to apply just as much to file paths,
tag names and HTTP responses.

---

## 7. Where we're going

**First, fix the chrome.** Restore the navigation, the buttons and the logo across all nine pages.
Before fixing it site-by-site, check whether other sites are affected — the code that fills in those
links is shared, so this may be a fleet-wide fault, and the fix belongs in the shared code.

**Then finish closing out the migration:** send a test payment event through the new front door to
prove the money path end-to-end; confirm one performance setting really landed on the server; clear
the content-delivery cache; deploy the tool's pending update, which carries both the spam defences
and the email fix in a single build; and add two DNS records so the tool's outgoing mail is fully
aligned with the domain — which matters for the paid reports reaching customers' inboxes, not just
our own.

**Then the wider question.** Every automated check assumes a site is served the way it was built. A
change of address or serving model happens *underneath* those checks, and nothing re-examines the
site afterwards. We have proposed a new check for the specific gap that bit us — a site whose links
point at a backend in a way the backend cannot answer — but the general lesson is larger, and applies
to every site that will follow this one onto its own server.

---

*End of briefing.*
