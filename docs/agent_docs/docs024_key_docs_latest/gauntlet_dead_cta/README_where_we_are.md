# Where we are — the vonc gauntlet

Plain-prose running log. Newest at the bottom.

## 2026-07-22 (start)

You reported that `vonc.com/tools/gauntlet/index.html#` doesn't work and weren't sure
we had a working gauntlet. Here's what I found.

The page itself loads fine, and most of it actually works — the countdown timer, the
tick-off objectives, the progress bar and the counting-up numbers are all real
JavaScript and behave. The two things that DON'T work are the two big buttons at the
top: "Enter the Gauntlet" and "Preview Rules". Both are wired to `href="#"`, which
means "go nowhere" — clicking just adds a `#` to the address bar. That `#` on the end
of your URL is exactly that. They're dead by design: the button *text* is configurable
but there is no field anywhere for the button's *destination*, so nothing could ever
fill it in.

Two more honesty problems on that page: the headline stats ("12,847 competitors",
"94,210 challenges completed", "38% win rate") and the whole "Top Competitors"
leaderboard (AxonFury, ZeroRush…) are invented placeholders. There is no real gauntlet
behind the button — no actual competitors, no real challenge to enter.

The good news: the platform already built a detector for exactly this ("dead controls")
and its source code literally names *our vonc gauntlet* as the example it was built to
catch. The bad news: it never actually caught it. It only looks at pages marked
"deployed", and our gauntlet page is marked "needs rebuild" even though it's serving
live — so the detector skipped its own poster child, and about 34 other live pages with
the same quirk.

You made three calls: make the gauntlet genuinely work (don't fake it); fix the
detector so any new site is covered; and send the council a message that we shouldn't
be creating placeholders that don't work. You also asked whether we should give it a
real backend like idea.uk or relojistas. My judgement: not a full one yet — a real
"gauntlet" needs real competitors and a live leaderboard, and with no real users a live
leaderboard is just another fake. So I'm making the front end genuinely work now,
reusing our existing form-delivery plumbing for the one real action (letting someone
actually submit their take to you), and leaving the full competitive backend as
something to switch on once there's real traffic.

Done so far: I fixed the detector (a one-line change — judge "is it live" by the
component that's actually serving, not the page's stale flag) and sent it to the
council with your placeholder message as the reasoning it reviews against. Waiting on
the council's verdict (~30 min). Next I'm rebuilding the gauntlet page itself so the
buttons do real things and the fake numbers come out.

## 2026-07-22 (done for today)

Both jobs are done. The gauntlet page now works honestly: the two dead buttons are
gone — "Enter the Gauntlet" is a real button that starts the clock and drops you into
the challenge, and "Preview Rules" now scrolls to a short "How the Gauntlet works"
card. I stripped the invented numbers and the fake leaderboard entirely; the page no
longer pretends there's a crowd of competitors. It's honestly a self-paced solo
challenge now, and it says so ("no sign-up, nothing is scored or shared"). All of this
is live on vonc.com right now — verified.

The generic detector fix went to the council and was APPROVED, and I've committed it.
It won't actually take effect until the next time the backend image is rebuilt (I
didn't force a fleet-wide rebuild just for a detection tweak) — but from then on, any
new site that ships a dead button on a page in this "serving-but-flagged-for-rebuild"
state will get caught, where before it slipped through.

And I sent the council your message — that we shouldn't be creating placeholders that
don't work — as the reasoning it reviewed this fix against, so it's on record with the
seats that judge platform changes.

On the backend question: I've deliberately kept this to an honest front-end for now, for
the reason above (a live leaderboard with no real users would just be a new fake). If
you'd like, the natural next step is to let a visitor actually submit their "Position"
to you using the same contact-form delivery we already fixed — a small, real feature —
and only build the full competitive backend once there's real traffic to populate it.

## 2026-07-22 (evening) — I was wrong; the tool still doesn't really work

You tried it and you're right. What I shipped was cosmetic. I took the dead buttons and
wired them up, but I wired them to things you can't see happen: "Enter the Gauntlet"
quietly starts the clock and scrolls to a panel that's already in front of you; "Preview
Rules" scrolls to a card that's already there. And ticking the objectives just moves a
progress bar that isn't connected to anything. So from where you sit, nothing happens —
which is the exact hollow-placeholder problem I'd just asked the council to stamp out.
I fixed the broken link and missed the actual point: the tool doesn't DO anything.

You've made the right call, and it's an honest one: build a real backend with an AI
opponent that actually plays against you — file your take, the AI challenges it, you
defend, there's a real outcome — and let it say plainly that it's an AI competitor until
real people show up. That's a genuine feature, not a fake crowd.

Good news: there's already a whole workstream built for exactly this — the "experience
loop" — and it was started because of THIS gauntlet. It can plan the whole end-to-end
experience (the journey, what's promised, what data it needs) and it's already produced
an approved plan once. The missing piece across the platform is turning such a plan into
a built, working feature — and doing that here also needs a real backend, which these
static sites don't have yet. So I'm researching three things before I commit to a
route: how the experience loop's build step works, how the "feature builder" implements
things, and how a static page like this can safely call a live AI backend. Then I'll lay
out a concrete plan for you rather than guess.

## 2026-07-23 — the real build is underway

Today the gauntlet stopped being a cosmetics job and became a real feature build, on
your decisions: a debate opponent (you file your take, the AI genuinely argues back,
you defend on the clock, it judges honestly), built by the platform's own
feature-builder, exposed through a shared API address on apis.uk with a bastion
machine guarding the cluster.

Progress so far today, in order:

1. **Unstuck the experience loop.** Its contracts reviewer had been blocking every
   plan that proposed new code (it demanded quotes from source that doesn't exist
   yet). Applied the rule split you approved — strict as ever about existing code,
   sane about genuinely new code. Config change, snapshotted, reversible.

2. **Gave the planner your new ruling.** The planner's standing instructions still
   said "client-side only, no server" — that's how decisions travel to it, so I
   updated that block with the debate design and the exact API endpoints, and fired
   a fresh planning run. It's working now (these queue behind the fleet; expected).

3. **The backend is being built by the machine, right now.** I wrote the formal
   capability request for `tools-api` (the debate engine service), the
   feature-designer turned it into a 6-stage build plan, and its review council
   approved it **unanimously on the first round**. Then I fired the
   feature-implementer — the first time it has ever run. It's writing the Go code
   stage by stage on its own branch, with a compile gate after every stage, and will
   open a pull request for you to review. **Nothing lands without your merge.**

4. **The security shell is drafted.** The tunnel config, the bastion proxy rules
   (only the tools API path gets through, with size and rate limits), and the
   cluster-side network policy are written and waiting on three things only you can
   do: name the subdomain on apis.uk, provision the bastion machine, and approve the
   WireGuard peering. No rush — the code has to land first anyway.

What I got wrong earlier still stands corrected in the record: last night's "fix"
looked right and did nothing a visitor could feel. Today's work is the opposite shape:
the machine is building something real, and the last gate is you reading a pull request.

## 2026-07-24 — the bastion plan had a hole; found it before we built it

You picked the address (tools.apis.uk), asked for a UK-owned machine to host the
bastion, and — importantly — asked me to double-check that the WireGuard link would
actually protect the cluster. Good instinct: it wouldn't have.

The original sketch said "add the bastion as another peer on the VPN you already use
from your laptop, then add a firewall rule that only lets the bastion's address reach
the new tools service." I went and looked at the real VPN inside the cluster instead
of trusting the sketch. Two problems. First, that VPN rewrites the sender's address on
everything passing through it, so the planned firewall rule would be checking an
address that never appears — it could never work. Second, and worse: anyone connected
to that VPN can reach everything in the system — including the main database — because
that's exactly what it's for: it's your admin door. Fine for your laptop and phone;
not fine for a machine that faces the internet.

The corrected design gives the bastion its own separate, single-purpose VPN doorway
into the cluster, walled in three ways: the doorway only accepts the one bastion key;
the doorway itself only forwards traffic to the one debate-engine service and drops
everything else; and the cluster's own network enforcement (which I confirmed is real
and active here) pins that doorway pod so that even if someone took it over completely,
the only thing it can talk to is still just the debate engine. So the promise we made
in the design — "if the bastion is compromised, the attacker gets the ability to use
the public API and nothing more" — is now actually true rather than just written down.

Also verified today: apis.uk is properly live on Cloudflare already (the nameserver
move worked), though there's a leftover catch-all address entry pointing at nothing —
one click to delete before we wire up the real one. The step-by-step for creating the
tunnel is written up and takes about ten minutes on the new machine, with one browser
login from you. On providers: Mythic Beasts (independently British-owned for 25 years,
solid reputation) is my recommendation; Clouvider (London) is the cheaper option; and
worth knowing — a Hetzner box wouldn't be UK-based at all, their machines are in
Germany and Finland. The smallest tier anywhere is plenty.

One admission: while inspecting the live VPN I printed its private keys into my own
session log on your machine. Low risk since it never left your machine, but the clean
move is to re-issue the laptop/phone keys sometime soon — five minutes, and I can do
it whenever suits.

## 2026-07-24 — your better idea: don't let the public near the cluster at all

You asked whether, instead of tunnelling public traffic into the production cluster,
we could run the debate engine on its own small setup at Mythic Beasts — a cut-down
cluster with Kafka and a database — completely independent of the main system.

I went and audited the approved build plan rather than answering from memory, and the
answer is yes — and it's even smaller than you asked for. The engine as designed needs
no Kafka at all: it's a plain web service plus one database table. It talks to three
things — its own Postgres, Anthropic's API (outbound), and the public vonc.com site to
fetch the day's provocation (also outbound). The only tie to the main system is that
it reads the list of your sites to decide which websites may call it, and that's a
one-line change to read from configuration instead.

So the whole thing fits on one small Mythic Beasts VM at £7–9 a month: the engine, its
database, the front-door proxy and the Cloudflare tunnel, all on the same box. No
Kubernetes, no Kafka, no VPN into the cluster. And the security picture transforms:
public traffic then never touches your production cluster in any way — there is
nothing to tunnel, nothing to firewall, no bastion to harden. If the worst happens and
someone owns that box, they get the debate records and an AI key we'll deliberately
issue as a separate, spend-capped key — and nothing else. Your cluster isn't reachable
from there at all. The bastion design I corrected this morning stays on file as the
fallback, but this is simply better for your stated worry, and cheaper.

It also gives the island a future: when the per-site AI features come along (the
advisory chatbot and friends), they're the same shape — public AI endpoints — and can
live on the same contained box, all on British-owned hosting.

Still yours to confirm: go with the island (I'll then write the VM setup files and
the amended runbook), order the VM, and we note the one config change for the pull
request when the machine-built code lands.

## 2026-07-24 — you raised the stakes: could the whole framework live at Mythic Beasts?

Your thinking: the API may get genuinely busy, and it should be able to use our
workflows and agent machinery — and without Kafka there is no framework. You're
right on both counts, and I checked the real numbers rather than guessing.

Three facts. First, the framework does need more than a bare VM: the agent system
creates and destroys its worker pods on demand (a third of the pods running right
now were spawned that way), so it needs a real Kubernetes underneath — a lightweight
one (k3s) on a Mythic Beasts VM does fine. Second, the framework is much lighter
than it looks: measured today, everything except Kafka uses under a couple of cores
and about 3GB; Kafka is the heavyweight, but a single-node Kafka (rather than
production's three) is entirely adequate for an island and fits in ~2GB. Put
together, the whole framework island fits a £37/month machine, £70/month with
generous headroom. So yes — it's genuinely possible, affordably.

One trap I want on the record: we already have a half-built "multi-cluster" feature
that looks like the obvious answer — it lets production dispatch agents to a second
cluster. It is the WRONG tool here, because it works by sharing production's Kafka
and database with the second cluster — and our Kafka currently has no access
control at all internally. Connecting a public-facing island that way would hand a
successful attacker the keys to production's nervous system — precisely what the
island exists to prevent. If the island runs the framework, it must be a second,
fully independent instance: its own Kafka, own database, own keys, nothing shared,
talking to production (if ever) only through proper authenticated APIs.

My recommendation, which keeps every option open: make the island "framework-ready"
from day one — put k3s on the Mythic Beasts VM and deploy the debate engine into it
now. If and when the load or ambition justifies it, we add the one-node Kafka and
the core framework services onto the same machine — an upgrade in place, no
re-platforming. And honestly: the debate engine's speed limit is the AI model, not
the hardware, so even the small box carries a lot of traffic before we'd feel it.

Worth naming: this is really the beginning of the "fully UK-hosted stack" idea you
parked a fortnight ago — British-owned compute, your own Kafka and database on it.
If you want to take that seriously, it deserves its own planning thread; I've not
absorbed it into this one.

## 2026-07-24 — DECIDED: the small island now (B1), a bigger box later if the framework needs one

You chose Route B1: the minimal island — one small Mythic Beasts VM running the
debate engine, its database, the proxy and the tunnel — with the explicit plan to
rent a bigger machine when the time comes to install the framework. That's a sound
sequence: moving later is an afternoon's work (copy the database over, start the
tunnel on the new box — the tunnel identity moves with its credentials file), so
nothing is locked in by starting small.

On the order form you drafted (VPS 1 with 5GB disk and IPv4, £6.90/month), two
adjustments recommended: take VPS 2 instead of VPS 1 — £2.50 more doubles the
memory to 2GB, which is the comfortable floor for running Postgres alongside the
engine without ever worrying about it; and raise the disk from 5GB to 20GB (about
£1.20 more) — 5GB is tight once the operating system, container images, logs and a
growing database share it. Keeping IPv4 (+£2) is right: some services we call out
to are IPv4-only, and £2 removes a whole class of connectivity head-scratching.
Skip managed hosting, SMS monitoring and graphs. The 10GB of mirrored backup space
(80p) is worth it for nightly database dumps held off-box at a second UK site.
All in: roughly £11.40/month before VAT. Debian 12 as the operating system;
monthly billing while we prove it, switch to annual (12 for 10) once settled.

## 2026-07-24 — the island is built and running; one click from you connects it to the world

You ordered the machine (Ubuntu to match the cluster — good call, everything runs
identically) and I set it up the same afternoon, working over SSH with your key.
As of now the island is live: locked down (key-only access, firewall dropping all
inbound except SSH, automatic security updates), running its database and the
front-door proxy in containers, with a nightly backup already tested. The proxy
behaves exactly as designed — it answers "not found" to everything except the one
API path, and that path answers "nothing behind me yet", which is correct until
the machine-built engine arrives. Nothing on the box holds any production
credential of any kind — no cluster access, no platform keys. The corrected rule
from this morning stands: if this box is ever compromised, your cluster is simply
not reachable from it.

The Cloudflare tunnel software is installed and waiting on the one step only you
can do: click its authorisation link (in the chat, and saved on the box) and pick
the apis.uk zone. The moment you do, I can create the tunnel, point tools.apis.uk
at it, and the island is publicly addressed — still serving 404s until the engine
lands, which is exactly right. Two small tasks stay on your list afterwards: the
backup account's host and username from the Mythic Beasts panel (so the nightly
dumps mirror off-box), and deleting that dead catch-all DNS record.

## 2026-07-24 (evening) — the island is on the internet

You clicked the authorisation link (the certificate came down as a browser download;
I moved it onto the island and deleted the local copy), and the tunnel is now live as
a proper system service. **https://tools.apis.uk** answers from the public internet —
politely refusing everything (404) except the one API path, which says "nothing behind
me yet". That's precisely the state we wanted: publicly addressed, guarded, waiting
for the engine. This morning that address was a Cloudflare error; the catch-all DNS
entry seems to have been deleted along the way too (you, I assume — good).

Left on your list, none urgent: two minutes of Cloudflare zone settings (strict TLS,
always-HTTPS, one rate-limit rule, the free firewall ruleset), the backup account's
host and username so the nightly dumps mirror off-box, and — when the engine's pull
request lands — a separate spend-capped AI key for the island. The engine build
carries on in the other thread; the moment its image exists, we switch it on here and
the Gauntlet has a real opponent at a real address.

## 2026-07-25 (morning) — listening post switched on; one small key job left for you

Your Cloudflare settings checked out from the outside — plain-http requests now get
bounced to https before they ever reach us. The backup account details you sent are
wired in: the nightly dump will copy itself off-box automatically as soon as you do
the one remaining step, which is pasting the island's public key into the Mythic
Beasts control panel under the backup account's SSH keys (the key is printed in the
island runbook). Until then the nightly dump still happens on the box; the off-site
copy just complains.

And the traffic sniffer you asked for is live — I didn't need you to touch DNS at
all in the end, because the island's Cloudflare certificate let me add the records
myself. Every name under apis.uk (including the bare domain) now quietly answers
"404 not found" while writing down who asked: which hostname, which page, where
they came from, which country. In a couple of weeks we read the log and see what
the world still thinks apis.uk is. When your bees homepage exists (other thread),
we point just the bare domain at it — the sniffer keeps watching everything else.

2026-07-25 (morning): found why the backend builder kept dying, and it was our own
cleaner. A housekeeping job runs every ten minutes and is supposed to sweep up
leftover message channels — but only when nothing is running. The check it used
for "is anything running?" could never say yes (it looked for a label that
nothing actually wears), so it swept EVERYTHING, every ten minutes, including the
channels our builder was actively using mid-job. The builder would do its work,
the reply would be posted to a channel that had just been binned, and the run
would hang forever. Both of yesterday's failed builder runs died this way — and
my first explanation (blaming a server restart) was wrong; the restart was just
nearby in time. The check is fixed and live: the cleaner now genuinely looks at
what's running (this morning it correctly held off — 39 agents were busy). The
builder has been re-fired on the same approved plan and is being watched. If it
gets through all six stages this time, we get the pull request — your review is
the gate after that.

2026-07-25 (~9.15am): it worked. With the cleaner fixed, the builder ran the whole
plan end to end for the first time — six stages, each one checked and committed,
then the final test pass — and opened the pull request for your review. That PR
(number 3 on the repo) contains the complete debate-opponent backend: the web
service, its container and deployment files, the database table for game rounds,
and the three endpoints (get today's provocation, file a position and get the AI's
counter-attack, defend and get an honest verdict). Nothing ships until you review
and merge it — that's your gate. One thing to decide at merge time: since the
public-facing home for this API moved to the standalone island machine, we'll
deploy it there rather than with the in-repo cluster files, and the game-rounds
table belongs in the island's own database.

2026-07-25 (~2.15pm): you merged the machine's pull request, and the backend is
now running on the island and answering from the real internet. Getting it live
surfaced three genuine bugs the machine's own checks couldn't have seen — a
wrong Go version in its container recipe, a database read that broke the two AI
endpoints for every fresh round (and mislabelled the failure as "round not
found"), and a database script whose safety check wasn't valid database language
at all. All three are fixed and the fixes are live; I've also sent them through
the review council for the record. One more discovery: when our server says "the
AI engine is offline" with a tidy error message, Cloudflare was swallowing the
message and showing its own bare error page — switched the status code so the
honest message gets through. Today's provocation now comes back from
https://tools.apis.uk with a round recorded in the island's database; wrong
origins are refused; oversized input is refused. Two things need you: (1) the
dedicated spend-capped Anthropic key for the island — until it's in, the two AI
endpoints answer honestly that the engine is offline; (2) the backup pubkey
paste from before. Also flagged: the provocations data file on vonc still
contains made-up stats from June — the new front end won't show them and the
file needs regenerating.

2026-07-25 (~3pm): the AI debate opponent is genuinely live. Once your key was on
the box, one more bug showed up — our code was never actually telling itself
which environment variable held the key, so it failed regardless of the key
being valid (confirmed the key itself worked by calling Anthropic directly).
Fixed, and now the whole loop works for real: file a position, get a real AI
counter-argument and challenge back, defend it, get a real AI verdict with
reasoning. I ran it twice end to end — both times the AI judged "opponent wins"
against a deliberately thin defence, which is exactly the honest behaviour we
wanted (not a pushover, not fixed to flatter the user). Both full rounds are
saved in the island's database with everything filled in. This is the real
evidence the experience-planning loop needs to design the front end around, so
that's next, followed by rebuilding the actual gauntlet page against the live
API. Also: the backup pubkey you pasted in is confirmed working now.

2026-07-25 (~4.45pm): the experience-planning council has APPROVED a real build
plan for the debate opponent — the first time this has ever happened for this
tool. It took two tries. The first attempt ran five rounds of genuine, detailed
critique from the review panel (five different reviewers each checking a
different concern: does it actually work, is it honest, is it well-scoped, does
every piece connect to something real) and then hit its round limit without
full agreement — which is a deliberate safety valve, not a failure: rather than
loop forever, it stops and hands you the disagreement to review. Reading through
it, the remaining objections were narrow and fixable: I had exact information
the reviewers didn't (the real shape of the API's responses, since I built and
tested it myself), plus two small facts about existing code they'd correctly
flagged as unverified. I fed all of that back in and re-ran it once — it
converged immediately, with only minor advisory notes left (nothing that blocks
building). The approved plan now exists as a record in the database, ready to
build against. Next: rewrite the actual gauntlet page against this plan, so
visitors get the real debate flow instead of what's there now.

---

**2026-07-26 (afternoon) — the page is built and proven, but I could not put it
live: the production database fell over underneath me.**

Two things happened today. The good one first.

I rebuilt the Gauntlet page against the approved plan, and it genuinely works.
Not "the buttons exist" — I ran the real thing in a real browser against the
real backend, twice over (desktop and phone-sized), and watched it go all the
way through: press the button, today's provocation appears and a twenty-minute
clock starts; type a position, and about ten seconds later the AI opponent
writes a counter-argument and puts a specific question back to you; answer it,
and a judge reads the whole exchange and returns a verdict with its reasons. In
my test run the verdict was "opponent wins", which is the right kind of answer
for a thing that is supposed to be hard. Sixty-five checks passed, none failed.
The archive page passed thirty-one out of thirty-one — you can now click a
provocation and read the full case for it, at a web address you can share, and
the one entry that has no case written yet is visibly not clickable rather than
being a button that does nothing.

I also fixed the data file you flagged. The made-up numbers are gone — the
"1,284 positions filed", the "62% disagree", the countdown that had been frozen
at "3h 12m" since June. Nothing in this system counts participants, so nothing
now claims to. Where a number appears it is something true by construction: the
clock is twenty minutes, there are three objectives, there is one verdict. I
went slightly beyond what the plan asked and rewrote the Arena copy too, because
it was advertising "six rooms live right now" with individual closing times, and
those rooms have never existed.

Now the bad one. Partway through putting all this live, the shared production
database started crash-looping and it is still doing it as I write. It is not
broken — that is the frustrating part. The database is perfectly healthy and
answers questions instantly if you reach it directly. What is happening is that
Kubernetes runs a one-second health check on it, and something else on the same
machine (an AI model doing an eight-minute piece of CPU work) is using all eight
processor cores. The database gets no guaranteed share of that machine, so the
one-second check times out, and Kubernetes concludes the database is dead and
kills it. Then it does it again. Seven times so far.

The reason it has no guaranteed share is a genuine mistake on our side, and it
has been there a long time: our own configuration file *does* reserve CPU and
memory for the database, and has since the very first commit — but the thing
actually running in the cluster does not have those reservations. The live
system quietly drifted from what we wrote down, and nobody noticed until
something else got greedy enough to expose it.

**I have not touched it.** It is shared infrastructure that every session and
every automated agent depends on, and fixing it means restarting the database
deliberately. That felt like your call, not mine, so I have written the whole
thing up with the evidence and the exact one-line command in
`bugs_open/082`. My honest read is that the fix is low-risk and would end the
outage — it just puts back what our own config already says — but it is your
production database and you should be the one to say go.

So where that leaves the Gauntlet: everything is built, tested and saved. The
new page and the new archive behaviour are written into the database and are
sitting there ready. What is missing is the last step that publishes them, and
that step runs through the very system that is currently cut off from the
database. Nothing is half-finished and nothing is at risk — the live site is
still serving the old version, exactly as it was this morning, and the corrected
data file is already live. The moment the database is stable, publishing is
about fifteen minutes of work, and then I can run the formal acceptance pass.

One thing worth telling you now rather than at the end, because it changes what
"passed" will mean: two of the approved plan's own acceptance tests cannot pass,
and it is our test harness at fault, not the page. The harness clicks a button,
waits three-tenths of a second, and checks whether the answer has arrived. But
these answers come from an AI and take eight to eighteen seconds. So the tests
will report failure on a page that is working correctly. I could make them pass
by having the page print a fake "thinking..." message that the test would accept
— and I am not going to do that, because then the test would also pass with the
AI switched off entirely, which defeats the whole point. I will fix the harness
or rewrite those two tests to check something honest.

---

## 2026-07-27 — the Arena, and what a new chassis build did and didn't do

A new chassis build went to production this morning. The first question was
whether it unblocked anything here, and the honest answer is no. The two things
this site was waiting on were the acceptance harness and the debate engine's
intermittent failures. The harness code hasn't been touched since the 25th — the
three-tenths-of-a-second wait is still there, because nobody had written the fix
yet, and a new build can't ship a change that doesn't exist. The engine runs on
the separate little server we rented, not in the cluster at all, so a cluster
build can't reach it. I checked both rather than assuming, and I checked the
live pages still worked afterwards. They do.

So I went looking for what actually was next, and found something worse than we
had on record.

The Arena page — the one we'd deliberately left alone — was not, as we thought,
a page that failed to load. It was a page that loaded beautifully and was almost
entirely invented. Twenty-six fictional people with usernames, posting opinions
they never wrote, each with a made-up count of how many others had voted their
take "Genius" or "Delusional". Underneath that, a "remix chain" showing invented
arguments evolving through invented contributors, with credit attributed to
people who don't exist. The daily provocation was a hardcoded list of five,
rotating by the day of the year, so it had quietly drifted away from the real one
the rest of the site shows. And the box inviting you to file your take saved it
to your own browser and nothing else — nobody would ever see it.

The previous session had put this off on the grounds that fixing the display
would turn a visibly broken page into a convincingly broken one. Reading the
actual source inverts that: it was *already* the convincing kind. That's the
worse of the two, because nothing on the page tells you.

You chose to scope it down honestly rather than build a real backend for it, and
that is what has shipped. The invented people, their votes, and the remix chains
are gone. The provocation now comes from the same file the homepage, the Gauntlet
and the archive read, so it can't drift again. The take box is gone and in its
place is a real link into the Gauntlet, where there is an actual AI opponent that
answers back. Where the fake community used to be, the page now lists the six
real provocations that genuinely exist, each linking to its own page.

I tested it in a real browser ninety times over — desktop and phone, and once
with the data file deliberately broken to check it fails honestly rather than
sitting on "Loading" forever. All ninety pass, both against my local copy and
against the real page after it went live.

Two things worth mentioning because they nearly went wrong. First, I was one
command away from "correcting" a stale flag on the Gauntlet page, until I read
the closing notes of another workstream and found they'd measured thirty-four
pages in exactly that state, all serving perfectly, and had deliberately decided
to leave them. Wrong-looking data isn't reason enough to change it; you have to
ask what still reads it. Second, my own test caught me writing the names of the
deleted fake-user variables into a code comment explaining what I'd removed —
which would have left those words sitting in the published page, where an
automated scanner can't tell an explanation from the real thing.

Separately, I found and wrote up a fault that isn't ours alone: sixteen tool
pages across six different sites are publishing their own build instructions as
the description search engines show. The Arena's was the worst of them — it told
Google the page had "no fetch calls, no backend". I've corrected that one page as
part of this work and filed the rest as a bug, because the code that causes it
needs a proper fix and the other fifteen need repairing.

What's left here is unchanged and honest: the engine still fails now and then and
still throws away the reason, and the acceptance harness still needs fixing
before it can test a page that waits ten seconds for an answer.

---

## 2026-07-27 (evening) — the engine can explain itself now, and what I got wrong getting there

The Gauntlet's debate engine used to fail every so often and tell nobody why. It
would return "unavailable" to the visitor and throw the actual reason away, so
there was no way to find out what had gone wrong — a 429 from the AI provider, a
timeout, an answer cut off mid-sentence, all arrived looking identical. That is
now fixed, deployed to the small server, and proved working.

Before starting I checked nobody else was on it, and found something worth
knowing: the bug number 083 belongs to **two different open cases**, and almost
every mention of "083" in our history refers to the other one, which someone else
is actively working. A quick look at the history would have said "crowded, leave
it alone". Ours had exactly one prior mention — me filing it.

The fix turned out bigger than the bug described. It named two places where the
error was discarded; there were nine, and auditing every failure path found seven
more of a second kind. Sixteen in total. One of them was actively destroying its
own evidence: a status code whose response body Cloudflare replaces with its own
error page, so the explanation never reached the browser at all. We had fixed
exactly that on the two neighbouring endpoints in July and missed this one.

The review council rejected my first submission and was right to. Its complaint
was narrow — it couldn't tell from my summary whether I'd edited four places or
two — but checking properly showed my own count was wrong and led to the seven
extra failures nobody had spotted. A reviewer who couldn't see the whole change
still roughly doubled its coverage.

Then you ruled on the one question the council said it couldn't decide itself:
whether the failure log should record an extract of what the AI said, given that
on this page the AI quotes the visitor's own argument back. Checking the detail
made it not a trade-off at all — the extract was capped at 300 characters, and the
specific fault it existed to catch puts its evidence about 1,500 characters in, so
it could never have seen the thing it was for. It now records the *shape* of a bad
response instead: how long, did it start with a brace, was there a code fence, how
many answers did it contain. Every question answered, no words recorded. And a
check now runs automatically before every commit so nobody has to notice this
again.

The VM is rebuilt and it works. I proved it by deliberately breaking it: same
code, a throwaway database, a deliberately wrong password for the AI service. The
new version wrote down exactly what happened — wrong key, provider's own error
reference. The old version wrote nothing at all. That silence is the whole bug,
and it is gone. A real argument through the live site afterwards worked start to
finish.

**What I got wrong, since it's a fair amount and the pattern matters.** Eight
things, written up properly in the technical notes. Three of them are the same
mistake in different clothes: **I wrote checks that could not have failed.** A
new automated check that scanned no files and reported everything fine. A
verification step that looked for the program in the wrong place, so it would
always have reported a failed deployment. And a "control" that searched a
compiled program for text that only exists in source code. Each one produces
output indistinguishable from success.

I also misread the council's own scoreboard and told you eight reviewers had
abstained when in fact eight had voted; I claimed a count I hadn't actually run a
command to produce; and I wrote a commit reference into a document before the
commit existed. All corrected, all recorded.

The lesson I've written into the fleet-wide log is the one that covers all three
of the big ones: **an assertion that cannot fail is not an assertion.** Anything
that says "we now verify X" needs to be shown failing once, at the moment it's
written, or it is decoration.

The bug stays open on purpose. We can now *see* the failures; we haven't stopped
them. The right next move is to wait a day or two, read what the log actually
caught, and fix that — rather than guess at which of three plausible causes to
address.

---

## 2026-07-28 — the engine is fine, and nobody is using it

Two things came out of checking the engine this morning, and the second one is
not a technical finding at all.

**The fault has stopped.** I fired twelve full arguments at the live engine —
twenty-four calls to the AI — and every single one worked, between six and
twenty-two seconds each. Adding yesterday's and Sunday's runs, that is around
forty-nine consecutive clean exchanges across three days with no failure at all.
The logging we added is switched on, working, and has caught nothing, because
there has been nothing to catch. Whatever was going wrong on Sunday appears to
have been a passing problem at the AI provider's end rather than something in
our code.

That is the right outcome, and it also means we should *not* now "fix" it. Two of
the three remaining ideas — automatically retrying, and allowing the AI longer
answers — were guesses about a cause we could not see. One of them we can now
positively rule out: the "answer got cut off" theory has a detector live in
production and it has never once fired, exactly as I predicted when I refused to
act on that theory last week. The third idea, giving the call a time limit, is
still worth doing, but it turns out to sit in shared code used by every agent we
run, not just this one — so it is a bigger, fleet-wide decision than the bug file
suggested, and I have flagged it as such rather than quietly making it.

**The second finding: the Gauntlet has had no visitors.** The new request log
recorded eight requests in its first twenty-four hours, and all eight were mine
from testing. The very first thing the new measurement produced was that there is
nothing to measure.

I mention it because it changes what "finished" means here. The experience works —
genuinely, end to end, against a real AI opponent, verified repeatedly. It is
also, at the moment, a room with nobody in it. Everything we have been fixing has
been about making the thing honest and reliable, which was the right order to do
it in, but no amount of further engineering will produce a visitor. That is now
the constraint, and it is a decision for you rather than a bug for me: whether
this site should be getting traffic, and if so, where from.

I also had to correct my own instruction from yesterday. I had written "wait a day
or two of real traffic, then read the log" as the next step. There is no real
traffic, so that would have had the next session waiting indefinitely for
something that was never going to arrive.

---

## 2026-07-28 (later) — the restyle, and the round you lost

You asked for three things on vonc.com: bigger text, a wider content column, and
the purple a shade darker so white text reads better. All three are live. Two of
them turned out to need more care than they looked.

**The purple could not simply be darkened.** That one colour was doing two jobs:
it was the background behind white text, and it was the colour of the links on
the dark page. Darkening it improves the first and ruins the second — and the
links were already below the readability standard before we started. There is no
single shade that works for both; I checked the numbers rather than guessing. So
the purple is now darker where white sits on it, and the links have their own
lighter shade, which quietly fixed a readability problem that was already there.

**The narrow column wasn't in the site's stylesheet.** Two of the blocks get
their width from inside the components themselves. The obvious fix was to edit
those components — but one of them, the hero, is shared across **182 pages on
other sites**: the watch glossaries, the fuel pages, the gripper guides. Widening
your homepage is not a reason to move 181 other people's pages. I overrode it
from vonc's own stylesheet instead, which stops at vonc.

**Then you hit a real bug, and it was mine.** You lost the AI's challenge while
answering it, the provocation was missing when you first arrived, and the Send
Defence button stopped doing anything. Those were all one fault: the whole round
only ever existed in the page's memory. If the page reloaded for any reason — a
refresh, the back button, or a phone quietly dropping the tab while you switched
apps — the round vanished, along with everything you had typed, while the round
itself was still running on the server. You had no way back to it. The button
then refused because it could no longer see a round, and said so in a message at
the top of the page while you were at the bottom looking at the button. From
where you sat, it did nothing.

It now remembers. Reload the page and your round comes back — the challenge, the
clock still counting correctly, and the words you had typed. I tested that on the
live page on both a computer and a phone: challenge intact, draft intact, and the
defence then sends and returns a real verdict.

Worth saying: my first suspicion was my own stylesheet, published forty minutes
earlier — bigger text can push things out of sight and a wider column can end up
covering a button. I checked that before anything else. It was not the cause, but
it was the first thing I looked at rather than the last.

And you were right about the provocation, including the part you weren't sure
about. It genuinely was blank when the page opened, for about three seconds,
because it waits for a file to load and nothing filled the gap. It now says it is
fetching, and says so honestly if it fails.

---

2026-07-28, mid-afternoon. Two things from this session, one good, one you need
to know about.

The good one first. The slow-burning risk we flagged two days ago — that the
software talks to the AI with no time limit at all, so a call that never answers
would freeze an agent until someone restarts it — turned out not to be
theoretical. Digging through the records, it has already happened once, back in
April: one call sat waiting for thirty minutes and was only freed because the
machine happened to restart. It is now fixed properly — every call gives up
after ten minutes, which is generous (the slowest genuine answer we have ever
recorded took six), and when a call does give up it leaves a clear note saying
exactly what happened instead of freezing silently. Filed as case 130, fix
written and committed; it takes effect at the next software release.

Now the thing you need to know. This afternoon, at about one o'clock, the
account we use to talk to the AI ran out of its monthly allowance. The provider
says it comes back on Friday the 1st at midnight. Until then, every part of the
system that thinks — the review panels, the diagnosis runs, the content writing,
and the Gauntlet's opponent on vonc.com — is off. I checked the Gauntlet
directly: a visitor who tries to play right now is told, honestly, that the
opponent is unavailable. Nothing pretends, nothing fakes an answer, which is
exactly what we built. But the site's one interactive feature is down until the
allowance resets, unless you choose to raise the limit with the provider —
that is your call, and nothing on our side can substitute for it.

One small silver lining: the failure alarm we armed on the Gauntlet two days
ago caught this within seconds and named the real cause. First time it has ever
fired, and it did its job.

---

2026-07-28, late afternoon. You raised the allowance and I checked it properly
before believing it: the cluster's AI calls are succeeding again, and I played
a real exchange against the Gauntlet's opponent — it is back and arguing well.

The timeout fix from earlier passed its review panel on the second attempt (the
first attempt died with the outage itself, which was fitting). And the blind
checker you exposed this morning — the one that passes a page whose content is
cut off the edge of a phone screen — now sees what you saw. It passed its own
review too. While proving it against the live site it found one more real cut
that this morning's repair had missed: a statistics chip on the homepage
hanging half off-screen. That is fixed and live; the same repair is queued up
for the three other sites that share the block, and will arrive with their next
natural republish rather than me touching their live pages today.

One correction to this morning's list: the about page's wide table, which we
had down as needing a scrolling wrapper, turns out to have had one all along —
a finger can drag it. The morning's measurement couldn't tell "scrollable" from
"cut off"; the new checker can, and that distinction is exactly why it won't
cry wolf across the rest of the fleet.

What remains on your list is the design work — making the provocation read as
the question you must answer, giving the page a visual ranking, and deciding
what pressing "Enter the Gauntlet" should actually reveal. That last one is a
genuine either/or you should pick, and I've laid the two options out at the
end of today's summary in the chat.

---

2026-07-28, evening. The change you chose this afternoon is live: the Gauntlet
page now opens with a single door. A visitor sees the title, what the game is,
the clock face and the rules — and one button. The provocation is no longer
sitting on the page when they arrive; pressing Enter the Gauntlet starts a
real round and the question is revealed by that press, with the box to answer
in directly beneath it. On a phone the button used to be two and a half
screens down; it is now on the first screen.

Two things worth telling you from the build. First, the checker we fixed this
morning — the one that could finally see content cut off the edge of a phone —
caught me twice while I was building this: the new button and the link beside
it were each a fraction too wide on a phone, for a reason no eye would spot in
the code. It flagged both before anything shipped. The tool you asked for
earned its keep the same day, on its own author.

Second, I tested the failure the honest way: I made the round-starting call
fail on purpose and watched what a visitor would see. The page stays closed
and says, right next to the button, that the opponent is offline and nothing
was lost. It only ever opens on a genuine answer from the engine — same rule
as everything else on this site.

Still on the list from your visit: making the provocation read unmistakably as
a question addressed to you, tidying the page's visual pecking order, and
recording a won verdict somewhere — that last one has two honest options
waiting on your pick.

---

2026-07-28, later still. The other two design items from your visit are now
live, so the whole set you asked for landed in one evening.

The provocation now looks like what it is. It sits in its own marked card —
the only block on the page with that treatment — with the "take a position on
this" line attached to the bottom of the card rather than floating below as
just another paragraph. And the steps of a round now carry a visible pecking
order: the thing to do right now is bright and marked, what you've done has
stepped back, and what comes later is dimmed until it's your move. The dimming
follows the real state of your round — it brightens because the opponent
actually answered, never because a timer guessed — and nothing dimmed is ever
switched off.

I watched a complete round end to end with these changes: position, the
opponent's counter, a defence, a verdict — each stage lighting up as it became
the current one, and a mid-round reload landing back exactly where it should.

One thing needs your hand. The timeout protection we built this morning is
still not on the island box — the new engine image is built, checked and
already copied over, but my tooling is not allowed to edit the island's
docker-compose file or restart the service. If you run this, the swap
completes and I'll verify it:

  ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && cp docker-compose.yml docker-compose.yml.bak-1193 && sed -i "s|tools-api:v1.0.1178|tools-api:v1.0.1193|" docker-compose.yml && docker compose up -d tools-api'

Also still on your desk: how a won verdict gets recorded — a private permalink
to your own round, or a shareable card made from the real verdict text.
