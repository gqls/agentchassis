The spam handoff is built on a false premise — I'll show you below. First, the task list.

● Task list is up (9 items). Here's what the search turned up — including a problem with the spam handoff itself.

The spam handoff is built on a false premise

HANDOFF_spam_and_ip_blocklist.md is wrong on three counts, and following it would waste the next session:

1. It says the /request handler is "the chassis Go process." It isn't. It's the standalone idea binary — module idea, stdlib-only, no dependencies, parked at docs/.../idea.uk/golang_files/. Nothing to do
   with the chassis.
2. It says the submissions are "almost certainly into a store we can query" and instructs the next session to hunt harder in clients_db, dismissing the earlier 0-row result as "a search miss, not a 'no
   data' result." The 0 rows were correct. idea.uk has no database at all — orders are a JSON file, /var/lib/idea/orders.json (store.go:3-5, setup.sh:150). So spam_read.sql is void; it should be discarded,
   not extended with more ILIKE needles.
3. It describes nginx as already serving static pages and proxying reserved paths. That's the future state — the cutover in thread 1 hasn't happened. nginx proxies everything to the tool today.

The one part it gets right is the conclusion: honeypot + timing beats an IP list, because spammers rotate IPs.

Two facts that constrain the fix. The Order struct has no IP field, so the existing spam rows cannot seed a blocklist retroactively — the only historical IP source is the nginx access log. And more
awkward: idea.uk is behind Cloudflare, but setup.sh never sets set_real_ip_from / real_ip_header CF-Connecting-IP. So nginx sees a Cloudflare edge IP, meaning the existing limit_req zone buckets all
Cloudflare traffic as one, and any nginx/fail2ban IP ban would ban Cloudflare, not the spammer. That's task 7, and it gates every IP-blocking approach. (Your own traffic_probe runbook documents this exact
trap; it was never back-ported.)

Good news: the ingredients exist and are just unwired. A per-IP sliding-window rateLimiter (3/hr, 20/day) and a clientIP() extractor both live in audience_check.go and are used only by the free taster.
handleRequest calls neither.

Transport, now that I know the scale

Several thousand domains kills the per-repo option outright — so the shape stays what you have: domain subdirectories in a small number of shared repos, one repo per hosting class, not per site. sites →
B2, vm-sites → VM. sites.github_repo picks which. That's exactly the design already in your concept register; the code was written and never wired.

The one thing that doesn't scale as-is: the vm-sites Action rsyncs to a single VM_HOST secret (relojistas' box). idea.uk is on a different box. With a growing VM fleet you can't keep adding host secrets —
so the hosts want to be data, not secrets: a deploy-targets.json (domain → host) committed at the vm-sites repo root, with one ed25519 deploy key authorised across the boxes.

One decision left, and it's about blast radius rather than scale:
One decision left, and it's about blast radius rather than scale:
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
←  ☐ Push or pull  ☐ Page collisions  ✔ Submit  →

How should the built files get from the git repo onto the VM boxes?

1. Push: Action rsyncs (Recommended)
   Reuse the working vm-sites Action, add a domain->host map so one repo serves the whole VM fleet. Instant on commit, matches what you already run for relojistas. Cost: one SSH deploy key held by the
   self-hosted runner and authorised on every box — compromise the runner, reach the fleet.
2. Pull: each box syncs itself
   Each VM runs a systemd timer that sparse-checkouts its own domain folder from the repo. No SSH key fan-out (a compromised box gets read-only repo access, not SSH to its siblings), and it scales to
   thousands of domains cleanly. Cost: sync latency = timer interval, and it's a new mechanism rather than the one you've proven.

