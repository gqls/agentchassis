Good — the site is coherent again: all 9 pages deployed, guides-index/news-index composed, tool-audience-check repointed to /audience-check. The index/tools card re-render is mid-flight (1 claimed, 1
triaged) and the monitor will tell me when the card link flips. Let me give you a proper checkpoint.

Where things stand

Phase 0 — leaked credentials: done my side, one action is yours.
The scrubbed idea.env.example is committed (a concurrent session swept it into v1.0.1121) — verified 10 placeholders, zero real secrets. A length-based pre-commit guard (scripts/check-secrets.sh +
.githooks/pre-commit) is installed and active locally; it's still untracked, so it needs committing to protect other contributors. The rotation (task #10) is still yours and is the only thing that
actually closes it — the real SES creds + INTERNAL_API_KEY are in pushed public history from June 4th.

Phase 1 — complete the site: essentially done, and it exposed a fleet-wide landmine.
The handoff's "just re-run build-site-planner, the union is safe" is wrong in both directions — it clobbered 4 built pages (the LLM re-proposes them and its sections win; the union only protects pages it
doesn't re-propose, and only adoption-locked pages are shielded) and it can never fill an empty page (the union preserves sections: [] faithfully). It also invented 10 pages. I caught it, rolled back with
zero section diffs, and repaired the two that had deployed regressed. Nothing was ever live-visible — DNS points at the VM, so the whole episode played out on the invisible B2 copy. All three original
targets are now resolved:
- guides-index, news-index — composed and deployed.
- tool-audience-check — resolved as your chosen option: a pointer page (url=/audience-check, marked deployed-and-pinned so reconcile never re-queues it). The tool-list card will point straight at the live
  tool once the index/tools re-render lands.

I recorded the landmine in memory and the docs so it can't bite the next session.

Phase 2 — per-site deploy target: code done, committed, compiles. All four wires landed (resolveGitRepoName is now called, was dead); inert until a chassis image rebuild.

What needs you next

The remaining work (tasks #4, #5, #7, #8, #9) all touches the live Hetzner box over SSH — the pull-sync provisioning, the nginx cutover, the Cloudflare real-IP fix, the /request hardening, and the spam
removal. I can't reach 116.203.204.115 from here, and these touch the earning service, so they're owner-executed with me preparing the exact scripts and code. Two smaller things I can do autonomously
right now: write the tool's /request hardening as Go changes into the in-repo source (golang_files/, ready for you to build+scp), and produce the ready-to-apply vm-sites Action guard + pull-sync files.

Three decisions are still open (none blocking): the /privacy collision (tool vs static, default tool), what business email /contact.html should show (its form also posts to a dead /contact), and whether
to commit the pre-commit guard now.

Want me to prepare the /request hardening code and the VM deploy files while you handle rotation and SSH-side steps — or would you rather steer somewhere else first?


---

All tasks complete. Here's where the idea.uk VM workstream now stands:

Phase 2 is done — idea.uk's deploy path to the VM is guarded, activated, and seeded

The RUNBOOK §2b→§2c sequence executed cleanly, but it took a detour through a significant discovery: the vm-sites deploy Action had never actually run. Not once. No self-hosted runner was ever registered on
gqls/vm-sites (the cluster runner serves only gqls/sites), and the shared runner image had no ssh, ssh-keyscan, or rsync — the workflow's SSH step died with a silent exit-127 because its stderr was piped to
/dev/null. "Commit is deploy" was only ever true for B2; relojistas' box content must have been hand-rsynced at go-live.

What's now in place:

- Allowlist guard (§2b) — deploy-targets.json maps domain→host (relojistas.com → its box only); unmapped domains are skipped, the VM_HOST secret is retired. Before pushing I defused a live booby trap: a stale
  queued run from another session's 11:44 push would have rsync --delete'd the repo's then-one-file relojistas.com/ folder over the live webroot the moment any runner appeared. I cancelled it; the relojistas session
  independently mirrored the box's webroot into the repo shortly after, which settled it properly.
- A real runner — image rebuilt as aqls/github-actions-runner:v1.0.1126 with openssh-client + rsync (I held the push until your chassis v1.0.1126 landed on the Hub — no clash, different image repo), plus a
  one-replica github-actions-runner-vmsites deployment on the healthy node. The sites runner was untouched.
- Verified live, both directions — three green runs show Skipping idea.uk (no mapped host), and the relojistas news-feed push deployed to 167.233.33.159 successfully, proving the mapped path end-to-end.
- Activation (§2c) — sites.github_repo = 'vm-sites' for idea.uk (branch main was already correct; rollback is one UPDATE to NULL).
- Repo seeded — rather than churning nine rebuilds through the pipeline, I copied idea.uk's built artefact (8 pages + all assets, 45 files) verbatim from gqls/sites into gqls/vm-sites. Future chassis builds commit
  there automatically.

One cutover-relevant correction to the RUNBOOK: static terms.html and refund-policy.html do exist (the runbook said they didn't) and the built footers link all three legal pages with the .html extension — so the
nginx cutover needs three 301s (/terms.html → /terms etc.) to keep the tool's versions canonical. Recorded in RUNBOOK §3b and the handoff.

Also fixed: the secrets guard (check-secrets.sh + pre-commit hook) was untracked because bulk doc commits had swept it into .gitignore — those lines are removed and the guard is now committed.

What remains is all owner-side on the box: credential rotation (still the urgent one), §3a pull-sync provisioning, the nginx cutover (16 reserved paths + the three 301s, Stripe webhook proven first), and the
hardened-tool deploy. All four docs and the memory are updated; commits are local on 085_debug_and_feature_loops (guard 30b6aa30d, runner c9eafa3c8, docs 88d2475b7) — not pushed. One housekeeping note: c9eafa3c8
accidentally swept in a pre-staged empty doc file from another session (SUMMARY_of_the_json_leak.md) — harmless, but it's in that commit.


----

You're right that the file is clean now — it was scrubbed on 2026-07-14. The issue is that git history is immutable and this repo is PUBLIC: I just verified (checking lengths and prefixes only, never printing
values) that from 2026-06-10/11 to the 2026-07-14 scrub, the pushed history contains:

- SMTP_USER — 20 chars starting AKIA → a real AWS access-key ID
- SMTP_PASS — 44 chars → the exact length of a real SES SMTP password
- INTERNAL_API_KEY — 64 hex chars → real openssl rand -hex 32 output

Anyone can run git log -p on a clone of gqls/agentchassis and read them today. That's why the runbook says the scrub is necessary but not sufficient — rotation is what actually closes this. What those values allow:
sending email as idea.uk through your SES account (your AWS bill, your sender reputation), and hitting /confirm /decline /approve /internal/run — approving orders and triggering Claude-billing runs on the live
tool. Stripe and Anthropic keys were verified to be truncated placeholders — no action needed for those.

One contradiction to resolve first: the scrubbed file's header says the creds were "caught and rotated", but the 2026-07-16 handoff still lists rotation as the urgent outstanding owner action. Only you can tell
which is true, so start with step 1.

Step 1 — Check whether rotation already happened

In the AWS console: IAM → Users → ses-smtp-user.20260611-195505 → Security credentials. Look at the access key's creation date:

- Created on/after 2026-07-14 → you already rotated; skip to step 4.
- Created 2026-06-11 (or anytime before 14 July) → the leaked key is still live; do steps 2–3.

While you're there, check whether AWS attached an AWSCompromisedKeyQuarantine policy to that user (GitHub's secret scanning often triggers this for public AWS keys), and glance at SES → Sending statistics and your
billing for anything anomalous since 11 June.

Step 2 — Rotate the SES SMTP credential (new first, then kill old — no email downtime)

1. SES console → SMTP settings → Create SMTP credentials. This creates a new IAM user with a fresh SMTP username/password pair. Save both.
2. On the box, update the real env file (not the repo one):
   ssh root@116.203.204.115   
   nano /etc/idea/idea.env    # replace SMTP_USER and SMTP_PASS
2. ⚠️  Known trap in this exact file: systemd's EnvironmentFile does not strip inline comments — never put # ... after a value on the same line (this once caused a restart loop).
3. Back in IAM: deactivate, then delete, the old access key on ses-smtp-user.20260611-195505 (or delete that user entirely once the new credentials are confirmed working).

Step 3 — Rotate INTERNAL_API_KEY

openssl rand -hex 32        # run anywhere; copy the output
Put it in /etc/idea/idea.env as the new INTERNAL_API_KEY, then:

systemctl restart idea && systemctl is-active idea

Then re-issue any operator /op links — they're signed with this key, so every old link (including any a stranger lifted from the repo) dies the moment the service restarts with the new key. That's the point.

Step 4 — Verify

- Send a test email through the tool (or run whatever flow triggers a notification) — proves the new SES pair works.
- Confirm systemctl is-active idea says active and the site still responds.
- If you like, hit a gated endpoint with the old internal key and confirm it's rejected.

What I'd not bother with

Rewriting the public repo's history (git filter-repo + force-push) would be disruptive — every clone breaks, every SHA referenced in your docs goes stale — and GitHub caches old commits anyway. Once the credentials
are rotated, the values in history are worthless strings. Rotation is the fix; history rewrite is optional cosmetics.

If it turns out in step 1 that you already rotated on the 14th, tell me and I'll correct the runbook, handoff and summary so "rotation" stops appearing as an urgent outstanding action.


--

All three remaining chat-doable items are done. Here's where this leg landed:

Contact form → mailto (your choice), plus a stale email caught

I traced the dead form to its real seam: the fleet contact-form template renders <form action="{{.form_action}}">, and idea.uk's per-site value was #contact. I changed only idea.uk's section data, not the fleet
template (that's the separate fleet-wide dead-form thread). While in there I found the form's description still named the old idea-uk@leopardess.uk — the block above it had been moved to idea.uk@contactforsales.com
back in p1_05/p1_06 but this copy was missed. Both fixed in sql/p1_07 (applied and verified):
- form_action: #contact → mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry
- description email → the new address (phone left untouched)

One honest caveat: this is fixed at the source (the authoritative section data), not yet in the deployed artifact. It publishes on the next contact-page build. I deliberately didn't force a single-page rebuild to
push it out now, because (a) I watched a rebuild of an already-deployed page bounce straight to needs_human_review at attempt 0 — that's the pipeline's review gate, appropriate for a real site — and (b) the site
isn't live yet (nothing pulls vm-sites until you set up the box sync), so there's no visible gap to close in a hurry. The pre-cutover build carries it. If you'd rather I force it through the review queue now so you
can approve and publish it immediately, say the word.

Deploy-path proof — reframed to verification

I set out to prove the chassis→vm-sites deploy by rebuilding a page, but an unchanged already-deployed page just bounces to review (no build). So instead of a synthetic build, I verified against the running pod per
CLAUDE.md: resolveGitRepoName is present in v1.0.1134 (3×) and sites.github_repo='vm-sites' is set — the routing is live in the image. The actual commit-to-vm-sites will be observed on the first natural build. I
cancelled the synthetic proof item so it doesn't sit in your review queue mislabelled.

Box artifacts ready to paste

Committed docs024_key_docs_latest/idea_uk_vm_site/box/ — the complete §3a/§3b payloads: provision-pullsync.sh (idempotent — deploy key, sparse clone, timer, first sync + 8-page check), the sitesync script and
systemd units, proxy_tool.conf, idea.uk.nginx (all 16 tool paths + the three legal-page 301s + loud-404 static root), and a README with the §3d proof loop and §3e swap/rollback. When you're ready for the box, it's
scp -r box … then run the script.

Four commits this leg (guard-hook, box artifacts, contact fix, and the earlier Phase-0/subject work), all local on 085_debug_and_feature_loops with explicit pathspecs per CLAUDE.md — nothing pushed. The workstream
is now fully prepared up to the box-side cutover, which needs your SSH.

---



talk me through implementing the pull-sync script and what we are syncing and explain to me why we are not deploying the pages directly to the box

● Good question to slow down on — the "why" here is a real architectural decision, not just a mechanical step. Let me pull up the exact files I committed and the decision record so I explain what's actually there
rather than from memory.

Read 2 files

● Here's the whole picture, from what's on disk to why it's shaped this way.

The shape in one line

A tiny systemd timer on the Hetzner box pulls idea.uk's built HTML out of the gqls/vm-sites git repo every 5 minutes and mirrors it into the folder nginx serves. The box fetches; nothing in the cluster ever reaches
into the box.

What we're syncing

The source is one folder — idea.uk/ at the root of gqls/vm-sites. That folder is the built static artefact: the eight rendered HTML pages (index, tools, report, about, contact, privacy, guides/index, news/index),
plus assets/ (the CSS, the hero and icon images, the JS) and tools/assets/ (the news-feed JS). It's exactly what the chassis renders in Kubernetes and the git-adapter commits — HTML and static files, no code, no
database, no templates. The ninth page, the "Free Audience Check", deliberately produces no file: it's a pointer to the live tool at /audience-check, which nginx will proxy, so there's nothing static to sync for
it.

The destination is /var/www/idea.uk on the box, which is what nginx's static root will point at.

Crucially, we sync only that one folder. gqls/vm-sites is designed to eventually hold many domains (relojistas.com is already there), and this box has no business holding anyone else's. That's what the sparse
checkout buys us — more on that below.

The three parts, and implementing them

There are three files, and one script that installs all of them:

1. sitesync — the sync itself (installed to /usr/local/bin/sitesync). Four lines that matter:

cd /var/lib/sitesync/repo
git fetch --quiet origin main          # download new commits; working tree untouched
git reset --hard --quiet origin/main   # force the tree to EXACTLY match origin
rsync -a --delete idea.uk/ /var/www/idea.uk/

The reset --hard (rather than git pull/merge) is deliberate: the box is a read-only mirror, not a place work happens. If anything on the box ever diverges from the repo — a hand-edited file, a half-applied change —
we want it obliterated on the next tick, not merged. The repo is the single source of truth, full stop.

The rsync --delete step then copies that clean tree into the webroot, and --delete means a file removed from the repo also disappears from the webroot — a true mirror with no orphaned pages. You might ask why there
are two copies (a git checkout and a webroot) rather than pointing nginx straight at the checkout. Two reasons: the checkout contains a .git/ directory, and you never want that under a web root
(https://idea.uk/.git/config would be downloadable); and rsync lets the served copy be a clean, plumbing-free tree. The git checkout is the staging area, the webroot is what the world sees.

2. sitesync.service — a systemd oneshot that runs the script as www-data, so everything rsync writes is already owned by the web user (no chown, no root in the sync path).

3. sitesync.timer — fires the service 1 minute after boot, then every 5 minutes. This 5 minutes is your deploy latency (a commit to the repo shows up on the box within 5 minutes, not instantly) — the accepted cost
   of this model.

The installer, provision-pullsync.sh, runs once as root on the box and does the whole §3a in order: creates the directories owned by www-data; generates a read-only deploy key as www-data; then pauses and prints
the public key for you to paste into gqls/vm-sites → Settings → Deploy keys with write access unticked; then does the sparse clone; installs the script and units; runs one sync; and verifies all eight pages are on
disk. It explicitly does not touch nginx — that's a separate, deliberate step so provisioning can't accidentally cut the site over.

The sparse clone is this bit:

git clone --filter=blob:none --no-checkout git@github.com:gqls/vm-sites.git repo
git sparse-checkout set idea.uk
git checkout main

--filter=blob:none means "don't download file contents until asked" and sparse-checkout set idea.uk means "only ever materialise the idea.uk/ folder." So even when the repo holds a thousand domains, this box
downloads only its own — the clone stays cheap forever.

Why pull, and not push

The obvious alternative is push: a GitHub Action (there's already one in the repo for relojistas) holds an SSH key on the cluster's self-hosted runner and rsyncs files onto the boxes. It works, but look at where
the credential lives. One SSH key, on the runner, authorised to write to every box. Compromise the runner — which sits in your cluster and runs whatever a repo push tells it to — and you can write to the entire
fleet. That blast radius grows with every box you add.

Pull inverts the trust. Each box holds its own key, and that key is read-only and only talks to GitHub. Now think about what a compromised box can actually do: it can't write to the repo (read-only, so it can't
poison the content other sites serve), and it has no SSH path to its sibling boxes (it only knows how to read from GitHub). The damage is contained to the box that's already compromised. One honest caveat so you're
not surprised later: a GitHub deploy key can read the whole repo — the sparse checkout is about download efficiency, not an access boundary — so the containment is specifically "can't write anywhere, can't reach
other machines," not "can't see other domains' HTML." For public marketing pages that's a non-issue.

There's an operational bonus too: a box that's offline during a deploy simply catches up on its next 5-minute tick, with no failed-push to retry and no coupling to whether the runner happens to be up.

Why go through git at all — why not deploy pages directly to the box

This is the deeper question, and it's worth separating from pull-vs-push. Why does the chassis render pages, commit them to a git repo, and let the box pull — instead of the renderer just scp/rsync-ing the finished
HTML straight onto the box?

- One artefact, two sinks — not a new pipeline. The chassis already renders in Kubernetes and commits HTML to a git repo for the whole fleet (that's how the thousands of B2-hosted sites work — "commit is deploy").
  The VM is a second destination for the same artefact, selected by one column (sites.github_repo). It is not a second renderer and not a second transport. If we bolted a direct box-deploy onto idea.uk, it would be a
  bespoke path that drifts from how every other site ships. Reusing the git seam keeps idea.uk a normal site that merely points somewhere else.
    - The cluster shouldn't hold keys to internet-facing boxes. Same blast-radius argument as above, one level up: a direct-deploy renderer would need SSH credentials to the box baked into the cluster. Git-as-transport
      means the cluster only ever writes to a git repo, and the box reaches out to fetch. Neither side holds a credential that can attack the other's machine.
    - Versioning and rollback for free. Every deploy is a commit. You can see exactly what changed, diff two states, and roll back by reverting — all with ordinary git. A renderer that rsyncs straight to disk leaves no
      history; a bad page is just gone, overwritten.
    - The box stays dumb, and that's the point. Its entire job is nginx plus a git pull. No application sits in the request path, so a crash of the paid tool can't take the marketing pages down with it — which is the
      whole reliability win of this migration. The moment the renderer deploys directly, you've coupled the box's content to the cluster's availability and lost that separation.

  And the furthest version of this question — why not have the box render the pages itself, dynamically — was rejected for a concrete reason: the chassis renders out of clients_db, the box has no route to that
  database, and the tool binary has no DB driver at all. Rendering on the box would mean either giving every box a path into the cluster database (coupling and a security problem across a fleet) or re-implementing
  rendering against a copied dataset — which is just the static artefact again, with extra moving parts. So: render once, in the cluster, where the data lives; ship the finished files; let the box be a file server.

  The cost of all this, stated plainly, is the 5-minute deploy latency and one more mechanism to run. For a marketing site that changes rarely and must stay up when the tool wobbles, that's a good trade — and it's
  exactly why this VM model is kept as the exception for the handful of sites that sell something, with B2-static remaining the default for the rest.

✻ Churned for 1m 46s

※ recap: Goal: put idea.uk's static site and paid tool on one box behind nginx. Phase 2 is done and the box-cutover files are prepared; I just explained the pull-sync design. Next action: you run
provision-pullsync.sh on the box to set up §3a.

---

● Fixed and committed (fdbfac3f6). Here's what to run on the box now.

Unblock it in one step

Re-copy the fixed script and re-run it — it's idempotent, so it will skip the directories and reuse the key you've already registered, populate known_hosts properly, and carry on to the clone:

scp -r box root@116.203.204.115:/root/idea-uk-box
ssh root@116.203.204.115 'cd /root/idea-uk-box && bash provision-pullsync.sh'

When it reaches the pause it will print the same public key you already added — just press Enter, no need to add it again.

What to look for as it runs

It now prints GitHub's host-key fingerprints before trusting them. The ED25519 one should read:

SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU

If it matches, you're good. (Worth a glance against GitHub's published fingerprints — this is the one moment we decide to trust the server, and blindly accepting a scanned key is how you'd miss an interception.)

Then a new pre-flight test runs before the clone. Success looks like:

Hi gqls/vm-sites! You've successfully authenticated, but GitHub does not provide shell access.
-- OK: deploy key accepted.

That greeting naming the repo confirms the read-only deploy key is working, and GitHub's "Never used" will flip to a timestamp.

If it stops again

The script now tells you which problem you have rather than failing cryptically:

- "could not fetch GitHub's host keys" — outbound port 22 is blocked by the box's firewall. The script prints the fallback: use GitHub's SSH-over-443 endpoint (ssh.github.com:443) via a ~/.ssh/config entry. Hetzner
  doesn't block this by default, but setup.sh configures ufw, so it's worth knowing.
- "GitHub refused the key" — the key isn't registered against gqls/vm-sites (wrong repo, or added to a personal account instead).

What went wrong, for the record

My original line was ssh-keyscan github.com >> known_hosts 2>/dev/null. Two flaws compounded: ssh-keyscan exits 0 even when it reaches nothing, so the set -euo pipefail guard never fired, and 2>/dev/null hid the
reason. The result was an empty known_hosts and a script that marched confidently into a clone that couldn't possibly work. Your deploy key was never the problem — "Never used" was the tell that SSH aborted at
host-key verification, before authentication. The fix asserts the file is non-empty and shows you what it's trusting, which is what it should have done from the start.

---
The fix was never tested — you ran a stale copy. Run this:

ssh root@116.203.204.115 'rm -rf /root/idea-uk-box'
scp -r box root@116.203.204.115:/root/idea-uk-box
ssh root@116.203.204.115 'grep -c pre-flight /root/idea-uk-box/provision-pullsync.sh'
ssh root@116.203.204.115 'cd /root/idea-uk-box && bash provision-pullsync.sh'

That third line must print 1 or more. If it prints 0, you're about to run the old script again and there's no point continuing.

How I knew

Two things were missing from your output that the new script always prints: == pre-flight: can this box authenticate to GitHub? ==, and the -- fetching GitHub host keys line. It jumped straight from the deploy key
to == sparse clone ==, which is the old script's shape exactly. The transferred file is 5572 bytes and does contain the pre-flight stage, so the copy was fine — it just landed somewhere you didn't execute.

The cause is a standard scp -r behaviour: when the destination directory already exists, scp -r box …:/root/idea-uk-box copies the source into it, giving you /root/idea-uk-box/box/provision-pullsync.sh. Your cd
/root/idea-uk-box then ran the original file still sitting there from the first copy. This is nasty precisely because the symptom is "the fix didn't work" rather than "wrong file" — I've documented it in the
RUNBOOK with the rm -rf and the grep check, because it would have caught us again.

A second real bug your output exposed

The Press Enter once the deploy key is added... prompt never appeared either. That's because ssh host 'bash script' gives no TTY, so read gets EOF immediately — the pause silently did nothing. My design was wrong:
it prompted unconditionally even though your key has been registered since yesterday.

I've reordered it so the script authenticates first and only prompts if that fails. Since your key is already on gqls/vm-sites, the pre-flight will now pass silently and go straight to cloning — no interaction
needed. If a key ever is missing and there's no TTY, it prints the key and exits with instructions rather than pretending to wait.

What to expect

== pre-flight: can this box authenticate to GitHub? ==
-- OK: deploy key accepted.
== sparse clone (this box fetches ONLY idea.uk/) ==

If instead you see Permission denied (publickey) at the pre-flight, that's now a genuine result rather than an artefact of the wrong script — it would mean the key is registered somewhere other than gqls/vm-sites.
And if you see anything mentioning /var/www/.ssh, the GIT_SSH_COMMAND fix isn't reaching ssh and I'd want to see it, since that's the bug 016 mechanism resurfacing.

I haven't added the scp nesting trap to /bugs_open/016 — it's an operator-workflow trap rather than the ssh mechanism that bug documents, and it's now captured in the RUNBOOK where someone doing this work will
actually meet it. Say if you'd rather it were filed as its own case.

✻ Baked for 2m 12s

※ recap: We're putting idea.uk's static site behind the VM's nginx alongside the paid tool; right now we're provisioning the box's pull-sync, which kept failing on a stale copied script. Next: re-copy after deleting
/root/idea-uk-box, then re-run provision-pullsync.sh.


---

That's Phase 3a done — idea.uk's site is now physically on the server, syncing itself every five minutes, and nothing the public sees has changed yet.

Reading your output back, everything landed as intended: the pre-flight passed (deploy key accepted), proving the GIT_SSH_COMMAND fix reached ssh; the clone pulled only idea.uk/ as designed; the timer is registered
with the next run five minutes out; and all eight pages verified OK.

One incidental confirmation worth noting: the webroot listing includes terms.html, refund-policy.html and privacy.html. That's the collision I flagged as a late correction — it's now proven on the box rather than
inferred, so the three .html → redirects in the staged nginx config are genuinely required. Without them you'd end up with two competing copies of your purchase terms.

Where this leaves you

The next step, §3b–3e, is the only one that changes what visitors see — everything so far has been reversible staging. The sequence is deliberately cautious: stage the config, validate it, test all sixteen tool
paths and push a real Stripe event through, and only then do the one-line swap. Rollback is restoring one file and reloading. The tool's binary and its order data are never touched at any point.

Worth doing before the cutover, while it's cheap: leave the sync running for a bit and confirm a second tick is clean (systemctl status sitesync.service after a few minutes, or journalctl -u sitesync.service). The
first sync was really part of the install; seeing an unattended one succeed is what proves the timer path works with the deploy key under systemd rather than under your shell.

Recorded

Docs updated and committed (7249d52a4) — RUNNING_NOTES §S with the full run, RUNBOOK "where we are", the resume handoff, and the summary table. The two traps that cost us cycles are captured where the next person
will hit them: /bugs_open/016 and 016b §9 for the ssh/$HOME mechanism, and the RUNBOOK for the scp -r nesting. Memory updated too.

Still outstanding on your side, unchanged: the nginx cutover, the tool binary deploy (spam defences plus the email subject fix), and the two DNS records for the SES bounce address.


---

Everything needed for the cutover is now verified and settled.

The sync is genuinely self-sustaining. That status=0/SUCCESS from a timer-triggered run is the one I wanted to see — the earlier run was part of the install, driven by your root shell. This one proves the deploy
key and GIT_SSH_COMMAND work under systemd as www-data, which is the path that will run unattended forever.

The route list is confirmed complete. Your box check correctly reported the source isn't deployed there (binary only), so I verified against service.go:596-612 instead: 16 tool routes plus /. The staged config
covers all 16 in 15 blocks — 12 exact matches plus three prefix matches, where ^~ /order/ deliberately covers both /order/success and /order/cancel. No gaps. Worth flagging that I briefly mis-reported /stripe/,
/internal/ and /order/ as missing: that was my own grep treating ^ as a regex anchor, not a real hole. I re-checked directly rather than let a shell-quoting artefact stand near the money path.

Your decision is recorded — the tool keeps all three legal pages, with the static .html copies redirecting onto them. That closes the last open decision, and the staged config already implemented it, so nothing
needs changing.

One thing to schedule deliberately

The box is reporting *** System restart required *** with 19 pending updates, one of them security. Not urgent and not blocking, but I'd choose the timing rather than have it happen mid-diagnosis: rebooting after
the cutover would make any wobble ambiguous between "the reboot" and "the new nginx config". Doing it before is cleaner, and it doubles as a free test of sitesync.timer's OnBootSec — you'd confirm the sync comes
back by itself after a restart, which is worth knowing before you depend on it.

What's left

The cutover itself (§3b–3e), which is the only remaining step that changes what visitors see: copy the two config files in, paste your real certificate lines and the port-80 block from the live config, nginx -t,
run the sixteen-path test loop, push a Stripe test event through, then the one-line swap. Rollback stays one file restore and a reload.

Then, separately and at your convenience: the tool binary deploy (spam defences plus the email subject fix, one build carries both), and the two DNS records for the SES bounce address.


-----
Build verified — clean linux/amd64 ELF, and the source is fully committed at 08e767b7c, so what you ship will match the repo.

Here's the full sequence. I've put the DNS first deliberately (it needs propagation time and carries zero risk, so it verifies in the background while you do everything else), and the tool deploy before the cutover —
right now nginx sends everything to the tool, which is the simplest possible topology to debug a new binary in. After the cutover, a tool bug and an nginx bug look alike.

  ---
Step 1 — SES bounce address (do first, then forget)

In the SES console, region Europe (London) eu-west-2 → Identities → leopardess.uk → Custom MAIL FROM domain → Edit:
- Tick "Use a custom MAIL FROM domain"
- Subdomain: bounce.leopardess.uk — not mail., which is typically what your mail client's hostname uses; claiming it could break your mailbox access
- On MX failure: Use default MAIL FROM domain (safe fallback — mail keeps flowing if the record ever breaks)

Then in Clook's DNS zone for leopardess.uk, add exactly two records:

┌──────┬────────┬──────────────────────────────────────────┐
│ Type │  Name  │                  Value                   │
├──────┼────────┼──────────────────────────────────────────┤
│ MX   │ bounce │ 10 feedback-smtp.eu-west-2.amazonses.com │
├──────┼────────┼──────────────────────────────────────────┤
│ TXT  │ bounce │ v=spf1 include:amazonses.com ~all        │
└──────┴────────┴──────────────────────────────────────────┘

Do not touch the root SPF, root MX, or _dmarc — incoming mail stays with Clook, and the SES SPF lives only on this subdomain. That separation is the whole point of a custom MAIL FROM.

Check progress any time:
dig +short MX bounce.leopardess.uk
dig +short TXT bounce.leopardess.uk
Both empty right now; when they answer, SES flips "Custom MAIL FROM" from Pending to Successful.

  ---
Step 2 — Reboot the box

The restart-required flag comes from updates already applied, so a plain reboot clears it. I'd deliberately not bundle the 19 pending updates into cutover week — one change at a time; apply them (or just the security
one) as a separate exercise afterwards.

Baseline first, so you know what "back to normal" looks like:
ssh root@116.203.204.115 'systemctl is-active idea; systemctl is-enabled sitesync.timer; uptime'
Reboot:
ssh root@116.203.204.115 '

---

Reboot passed cleanly, and that missing file saved us from a bad cutover.

The reboot is verified

All gates green: idea active, 8 pages on disk, /health 200 — and the timer's last run was 14:21:33, about a minute after boot, which is OnBootSec firing. The sync re-arms itself after a restart, so that's now proven
rather than assumed.

sites-enabled/idea.uk doesn't exist — and chasing that found a real problem

Rather than guess the path, I read setup.sh (the script that provisioned the box). The live config is /etc/nginx/sites-available/idea.conf. But comparing it against what I'd staged for you turned up something worse than
a filename: my config would have silently downgraded the box. The live one carries settings mine dropped:

- proxy_read_timeout 930s — setup.sh's own comment says "the engine can take minutes". My snippet had nothing, so nginx's 60s default would have cut report generation off mid-run. This is the one that would have
  actually hurt: reports failing after the cutover, looking like a tool bug.
- limit_req zone=idea_rl and its limit_req_zone preamble — rate limiting would have vanished entirely.
- The port-80 ACME + redirect block, IPv6 listeners, ssl_protocols, four security headers including HSTS, client_max_body_size 1m.
- The ACME webroot is /var/www/letsencrypt, not /var/www/html as I'd written — certificate renewal could have broken later.

There's also a subtlety I'd have got wrong: setup.sh deliberately gives /stripe/webhook its own location with no rate limit. That's correct — Stripe retries webhooks in bursts, and a 503 from limit_req would read as an
outage and delay payment confirmation. I've split that into a separate proxy_stripe.conf so the money path can't inherit a limit from the shared snippet.

All of it is now rebuilt as a faithful superset and committed (c5357595b), with all 16 routes verified present.

One standing landmine, now documented

idea.conf says "managed by setup.sh — do not edit by hand". After the cutover, re-running setup.sh would overwrite it and revert the site to tool-only (and it does ufw --force reset). If the box ever needs
re-provisioning, these server blocks have to go into setup.sh's stage-2 template first.

Revised cutover commands

# 1. see which preamble the live conf uses — copy it verbatim into idea.uk.nginx's top
ssh root@116.203.204.115 "sed -n '1,/^server {/p' /etc/nginx/sites-available/ide

# 2. re-copy the corrected box files, then stage
ssh root@116.203.204.115 'rm -rf /root/idea-uk-box'
scp -r box root@116.203.204.115:/root/idea-uk-box
ssh root@116.203.204.115 '
cp /root/idea-uk-box/proxy_tool.conf   /etc/nginx/snippets/proxy_tool.conf
cp /root/idea-uk-box/proxy_stripe.conf /etc/nginx/snippets/proxy_stripe.conf
cp /root/idea-uk-box/idea.uk.nginx     /etc/nginx/sites-available/idea.conf.new'

Then check the cert paths in the live file match the staged one, run the port-8443 rehearsal from my previous message (pointing at idea.conf.new), and swap with:

ssh root@116.203.204.115 '
cp /etc/nginx/sites-available/idea.conf /root/idea.conf.bak.$(date +%Y%m%d-%H%M%S)
cp /etc/nginx/sites-available/idea.conf.new /etc/nginx/sites-available/idea.conf
nginx -t && systemctl reload nginx && echo CUTOVER_RELOADED'

The sites-enabled/idea.conf symlink already exists and stays untouched — only the file it points at changes.

Steps 1 (SES DNS), 3 (tool binary) and 5 (SES confirm) from my previous message are unaffected. Given the tool deploy is independent of all this, that's still worth doing first — it's low-risk and gets your spam and
subject-line fixes live while the nginx work waits for a moment you can watch it.

Wait ~60 seconds, then verify — this doubles as a free test of sitesync.timer's OnBootSec:
ssh root@116.203.204.115 '
uptime
systemctl is-active idea
systemctl list-timers sitesync.timer --no-pager
systemctl status sitesync.service --no-pager | tail -3
ls /var/www/idea.uk/*.html | wc -l'
curl -s -o /dev/null -w "site: %{http_code}\n" https://idea.uk/health
Gate: idea active, the timer listed with a next run, the webroot still populated, /health returning 200. If the sync didn't re-arm after boot, stop and tell me — better to know now than after the cutover.

  ---
Step 3 — Deploy the tool binary

This ships both pending fixes in one go: the /request spam defences and the email subject fix. Both are backward-compatible (the honeypot only rejects when filled; the timing gate fails open when _elapsed is absent), so
the tool's existing embedded form keeps working.

Confirm the real paths — don't assume:
ssh root@116.203.204.115 'systemctl cat idea | grep -E "ExecStart|EnvironmentFile|WorkingDirectory"'
Expect ExecStart=/opt/idea/idea. If it differs, use what it actually says.

Build, from the repo folder:
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
GOOS=linux GOARCH=amd64 go build -o idea . && ls -la idea

Ship with a backup in the same breath:
scp idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 '
cp -a /opt/idea/idea /opt/idea/idea.bak.$(date +%Y%m%d-%H%M%S)
mv /opt/idea/idea.new /opt/idea/idea
chmod +x /opt/idea/idea
systemctl restart idea
sleep 2
systemctl is-active idea
journalctl -u idea -n 20 --no-pager'

Verify:
curl -s -o /dev/null -w "health: %{http_code}\n" https://idea.uk/health
curl -s https://idea.uk/capacity
Then the real proof: submit a genuine request through the site and confirm the REVIEW email arrives (not quarantined) with a short subject — [idea.uk] REVIEW ord_… (first 60 chars…) rather than the whole pitch. That
single email tests the subject fix and the hardened /request path end to end.



Rollback if needed — instant, no rebuild:
ssh root@116.203.204.115 'mv /opt/idea/idea.bak.<timestamp> /opt/idea/idea && systemctl restart idea'

  ---
Step 4 — The cutover

4a. Read the live config and copy out your real certificate lines and the whole port-80 block:
ssh root@116.203.204.115 'cat /etc/nginx/sites-enabled/idea.uk'

4b. Stage the new files:
ssh root@116.203.204.115 '
cp /root/idea-uk-box/proxy_tool.conf /etc/nginx/snippets/proxy_tool.conf
cp /root/idea-uk-box/idea.uk.nginx  /etc/nginx/sites-available/idea.uk.new'
ssh -t root@116.203.204.115 'nano /etc/nginx/sites-available/idea.uk.new'
In that edit: replace the two ssl_certificate* lines with the real ones from 4a, and paste in the port-80 server block verbatim.

4c. Rehearse on a spare port before touching the live one. This is the belt-and-braces bit and it's worth it on a money path:
ssh root@116.203.204.115 '
sed "s/listen 443 ssl http2;/listen 8443 ssl http2;/" /etc/nginx/sites-available/idea.uk.new \
> /etc/nginx/sites-available/idea.uk.test
ln -sf /etc/nginx/sites-available/idea.uk.test /etc/nginx/sites-enabled/idea.uk.test
nginx -t && systemctl reload nginx && echo TEST_CONFIG_LIVE_ON_8443'
Now test every reserved path against the new config without any public exposure — expect the tool's codes (200/400/401/405), never a static 404:
ssh root@116.203.204.115 '
for p in /health /capacity /audience-check /subscribe /request /confirm /approve /decline \
/op /stripe/webhook /internal/run /order/success /order/cancel /terms /refund-policy /privacy; do
printf "%-16s -> " "$p"
curl -sk -o /dev/null -w "%{http_code}\n" -H "Host: idea.uk" https://127.0.0.1:8443$p
done
echo "--- static + redirects ---"
for p in / /about.html /tools.html /terms.html /privacy.html /nonexistent-xyz; do
printf "%-16s -> " "$p"
curl -sk -o /dev/null -w "%{http_code}\n" -H "Host: idea.uk" https://127.0.0.1:8443$p
done'
Gate: all 16 tool paths reach the tool; / and the .html pages return 200; /terms.html and /privacy.html return 301; /nonexistent-xyz returns 404 (loud, not the homepage). Anything wrong here is free to fix — nothing
public has changed.


Tear the rehearsal down:
ssh root@116.203.204.115 '
rm -f /etc/nginx/sites-enabled/idea.uk.test /etc/nginx/sites-available/idea.uk.test
nginx -t && systemctl reload nginx && echo TEST_REMOVED'

4d. The swap. Note the safety property: replacing the file changes nothing until reload, and a failed nginx -t cannot take the site down — nginx keeps serving the old config.
ssh root@116.203.204.115 '
cp /etc/nginx/sites-enabled/idea.uk /root/idea.uk.nginx.bak.$(date +%Y%m%d-%H%M%S)
cp /etc/nginx/sites-available/idea.uk.new /etc/nginx/sites-available/idea.uk
ln -sf /etc/nginx/sites-available/idea.uk /etc/nginx/sites-enabled/idea.uk
nginx -t && systemctl reload nginx && echo CUTOVER_RELOADED'
If nginx -t fails, it stops there and the live site is untouched — restore the backup and investigate.

4e. Verify publicly, then purge the Cloudflare cache for idea.uk:
for p in /health /capacity /audience-check /subscribe /request /confirm /approve /decline \
/op /stripe/webhook /internal/run /order/success /order/cancel /terms /refund-policy /privacy; do
printf '%-16s -> ' "$p"; curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk$p
done
curl -s -o /dev/null -w 'home: %{http_code}\n' https://idea.uk/

4f. The money path — the one that actually matters. From the Stripe dashboard, send a test event to the https://idea.uk/stripe/webhook endpoint and confirm the tool receives and verifies it:
ssh root@116.203.204.115 'journalctl -u idea -n 30 --no-pager | grep -i stripe'
Then, when you're ready, a real end-to-end purchase: /request → /confirm → /approve → pay → order reaches paid in /var/lib/idea/orders.json.

Rollback, any time, seconds:
ssh root@116.203.204.115 '
cp /root/idea.uk.nginx.bak.<timestamp> /etc/nginx/sites-enabled/idea.uk
nginx -t && systemctl reload nginx && echo ROLLED_BACK'
The tool's binary, service and order file are never touched by any of this — rollback reverts only the front door.

  ---
Step 5 — Close out SES

Back in the SES identity page, confirm Custom MAIL FROM: Successful. Then send one more test email through the tool and check the headers show spf=pass … smtp.mailfrom=bounce.leopardess.uk alongside the dkim=pass
header.d=leopardess.uk you already have. That's full alignment, and it's what stops your paid report emails carrying a deliverability handicap into customers' inboxes.

  ---
Two things I'd deliberately leave for afterwards rather than bundle in: the 19 pending apt updates, and the real-client-IP work (§4a — the Cloudflare set_real_ip_from / CF-Connecting-IP config, which needs you to
confirm whether the DNS record is proxied first). Neither blocks anything above.

Want me to fold the port-8443 rehearsal into RUNBOOK §3d? It's a genuine improvement on what's written there — the current text says "prove it before cutting over" without specifying how, and this is the how.

----
You've found something important, and the headline is worse than the symptoms: the paid funnel is currently unreachable on a live earning site.

What's actually wrong

The tool served its own landing page at /, and that page carried the real entry UI — the audience-check form and the report-request form. Those POST to /audience-check and /request, which are handlers, not pages. The
cutover gave / to the static site. nginx proxies all 16 routes correctly and the tool is perfectly healthy — but the forms that drove it are gone, and the static site never had any. I verified: the entire site contains
one form, a newsletter box with no action, and zero things posting to /request. The two /audience-check references are plain href links, so clicking one is a GET to a POST-only handler — your "POST only" page.

So the tool is running, reachable, and unusable. Every smoke test I ran passed, because nothing is erroring; the funnel is simply absent.

Painfully, this was half-seen. RUNBOOK §3b carried the note "…and / (the landing page it loses)" — recorded as a routing fact. Nobody asked what was on that page. The earlier pointer-page decision then aimed the "Free
Audience Check" card at /audience-check on the assumption it was a page. It never was.

Your other two symptoms are real and separate:

┌─────────────────────┬─────────────────────────────────┬────────────────────────────────────────────────────────────┐
│       Symptom       │             Markup              │                            Note                            │
├─────────────────────┼─────────────────────────────────┼────────────────────────────────────────────────────────────┤
│ "idea.uk logo" text │ <img src="" alt="idea.uk logo"> │ asset exists (logo.jpg → 200) — the URL is never filled in │
├─────────────────────┼─────────────────────────────────┼────────────────────────────────────────────────────────────┤
│ Buttons do nothing  │ <a href="">, <a href="#">       │ several per page                                           │
└─────────────────────┴─────────────────────────────────┴────────────────────────────────────────────────────────────┘

On the auditors

They exist and are configured — dead_controls catches href="#", phantom_internal_links catches href="", plus misdirected_cta and the image checks — spread across three discovery agents. But no discovery check has ever
run against idea.uk. The site has 112 work items across 13 sources; discovery isn't one of them. I cleared that query against the full source list rather than trust a zero.

I wrote an on-demand trigger (TRIGGER_discovery_agent.sh) and hit two snags worth knowing: the workflow resolves by domain and dies "domain not found in input_data" if given only a site_id (cost one run — now fixed in
the script), and after that fix later dispatches produced no orchestration row at all. The pod was six hours old, so the documented 300s-post-restart drop doesn't explain it. On-demand discovery dispatch needs its own
diagnosis; I stopped rather than sink more time while you have a broken funnel.

And there's a genuine auditor gap. No check models the backend. They ask "does this href resolve?" — and /audience-check does resolve. Nothing asks whether a link's target accepts the method the link will use, or
whether a site with a backend still contains any form that posts to it. That gap is exactly what let this ship.

I've filed /bugs_open/017 with the full diagnosis, a proposed backend_entry_orphaned check (two findings: method_mismatch_link and no_backend_entry, both cheap static analysis that would have fired the moment the
cutover landed), and the site-fix options.

What I'd do about the funnel now

The fastest restoration is to give the tool a GET page path — add a /start handler serving the landing markup it already has embedded, proxy it, and point the CTAs there. That brings both forms back in one small change.
The better end state is authoring the forms as chassis sections in the static build (same origin, so no CORS, and the /request honeypot and timing fields can be authored properly) — but that's a bigger piece of work.

Given it's live and selling, I'd suggest the quick path first. Want me to write that /start handler against the tool source? It would ship with the binary deploy that's already pending, so it costs no extra deployment.

-----
## 2026-07-19 — the broken chrome (bugs_open/018): found it, and it is not what 018 guessed

Short version: **the nav data is fine, the templates just ask for it by names nobody supplies.**

I started with the question 018 said to answer first — is this fleet-wide? It is not. Across all
11 sites, idea.uk is the only one with any empty links in its header or footer. Everyone else is
clean. So this is not a shared filler that broke; it is idea.uk sitting on two components nobody
else uses.

Here is the mechanism, and it is simple once you see it.

When the platform renders a site's header and footer, it builds a fixed list of values and hands
them to the template — things like `company_name`, `logo_url`, `nav_items_html` (the whole
navigation, pre-rendered as one blob), `cta_url`. That list is **hardcoded in Go**. The renderer
never looks at what the component says it needs; there is no `input_schema` anywhere in
`render_site_components_action.go`.

Nine sites use older header/footer components written against exactly that vocabulary. idea.uk
uses two newer ones — literally named `site-header` and `site-footer`, created 2026-05-06 — which
ask for a completely different set: `nav_link_1_url`, `nav_link_1_label`, … `cta_primary_url`,
`nav_aria_label`, and in the footer `col1_link1_url` through `col3_link4_label`. **None of those
names exist in the renderer's list.** Every one comes back as an empty string.

The confirming detail is the one field that does work. `company_name` **is** in the renderer's
list — and sure enough it is the one thing on the page that rendered: the logo says "idea.uk" and
the link's aria-label says "idea.uk home". Everything absent from that list is blank. That is not
a coincidence; that is the whole bug in one line.

Two things then turn a data gap into a visibly broken page:

1. **The old components gate their links, the new ones do not.** vonc's header wraps every link in
   `{{if .cta_url}}…{{end}}` — no value, no anchor, nothing shows. idea.uk's two templates contain
   **zero** conditionals, so an empty value becomes `href=""` — a visible, clickable, dead control.
2. **The schemas point at data that has never existed.** `site-header` declares its URLs come from
   `site_specs.navigation.link_1_url`. There is no `navigation` aspect in `site_specs` for any site
   on the platform — I checked the full aspect list across all 11. The logo comes from
   `site_assets.logo`, similarly unread. Those `source:` lines are decorative: nothing resolves
   them. That is also why `nav_aria_label` rendered empty despite having a static fallback of
   "Main navigation" — the fallback machinery never runs for chrome either.

**A correction to 018, which I got to test.** 018 theorised the empties came from a URL-shape
mismatch — the resolver thinking in `/about` while the site is built as `/about.html`. That is
wrong. idea.uk's navigation table is in good order: 6 primary items, 1 utility, 1 legal, all
`active`, all correctly `.html`-suffixed (Home, Tools, About, Guides, News, Report, Contact,
Privacy and Terms). The data the header needs is sitting right there. Nothing consumes it.

So the good news is real: **there is nothing to rebuild and nothing to re-plan.** The pages are
fine, the nav is fine, the assets are fine. Two templates are asking the wrong questions.

The fix is a rewrite of those two templates to use the vocabulary the renderer actually provides,
with every anchor gated. It is a database change, so it goes live on the next render — **no chassis
image, no deploy**. The CSS stays exactly as it is, so the site keeps the sticky header and mobile
menu it was designed with; the only visible change is that the nav will show the six real pages
instead of four empty slots.

One separate, smaller thing while I am in here: `sites.logo_url` is empty for idea.uk even though
`/assets/images/logo.jpg` serves 200. vonc has the same empty value but its gated template falls
back to a glyph, so nobody noticed. Setting that one field makes the logo appear.

Where this connects: this is the header/footer half of `/bugs_open/023` (a button's label and its
destination are never checked against each other), which measured `page_components` only and said
in as many words that header and footer were "not yet counted". Now they are. 023's fix #2 —
"gate every CTA anchor" — is exactly what this needs.


## 2026-07-19 (later) — put to the council, and a wrong turn of mine on the way

Per your steer, I put the **Go fix** to the council rather than doing it, and I went looking for the
prior discussions first. You were right that they exist, and they changed the submission rather than
just decorating it.

The important one: **this exact surface was fixed once before, the opposite way.** A decision on
record as LNK-007, deployed in June, fixed header/footer phantom links by *hardcoding more* into the
renderer — and the register says plainly why: the template-level fixer "excludes ContentData values
and literal anchors — which is why the header/footer phantoms had to be fixed at source in Go
instead." So I am arguing with a deployed decision, and the submission says so in its first
paragraph rather than pretending to be a fresh idea. My argument is narrow: LNK-007 repaired the
*values the map already supplies*; it could never help a component asking for names the map does not
contain, because the map is the ceiling. I am proposing to stop the ceiling being fixed, not to undo
LNK-007.

The second useful find: the schema-driven principle is **already approved by you** — the CTA plan
names "one derivation function, three call sites: plan_sections, resolve_internal_links,
applyCTARecompute." The chrome renderer simply isn't on that list. So the submission is framed as
adding the missing fourth call site. And a runbook note explains why nobody spotted the gap: the CTA
census covers page components only and explicitly "does not cover site_components (header/footer)".

**The prior-art search caught a regression I was about to ship.** My first draft applied each
field's declared fallback whenever a value couldn't be resolved. That sounds harmless. But the
header component nine sites use declares its button URL as `pages.contact` with a fallback of
`/contact.html` — which the notes describe as "the literal fossil of the 143-of-144-buttons-point-at-
contact bug". My rule would have re-created that phantom across nine live sites, undoing the very
thing LNK-007 was deployed to fix. Changed before submitting: fallbacks now apply only to fields
declared as static literals; anything that fails to resolve is left empty for the template to hide.

**Now the wrong turn, which cost a little money.** After submitting I couldn't find the run in the
database and found nothing in the logs, while other people's council runs were visibly completing.
I concluded the message had been dropped and re-sent it. That was wrong. The requests all go through
a **single queue with a single consumer**, strictly in order, and it was **26 minutes behind** — my
message was sitting near the front of a backlog, not missing. "Mine hasn't run but later ones have"
is what a queue looks like from the outside, and I read it as a hole. The cost: the same plan will
be reviewed twice and bill twice. I can't recall it — it's already sent.

**Worth knowing beyond this task**, because it isn't really an idea.uk matter: while I watched, the
backlog grew from 41 to 62 messages and the consumer advanced by one. Production is outrunning
consumption by roughly ten to one. With several sessions firing work at the cluster at once, every
thread's dispatch is queuing behind every other thread's. It also retro-explains an earlier puzzle
in this workstream — a session recorded discovery dispatches "producing no orchestration row at all"
and couldn't explain it; that was almost certainly this, and it cost that session a run too. I've
noted the one command that settles it in future.

**Where the fix itself stands:** submitted, trail id `7152c7cf`, waiting on the queue. It is
advisory — a verdict, not a gate — so it records a judgement either way. Separately, and still true:
this Go change would not by itself make idea.uk navigable, because its two templates have no
conditional guards at all. That part is a database-only change and needs no deploy.

## 2026-07-20 (late) — the council said REVISE twice; the second time it's asking a real question

Quick honest update on the chrome-renderer fix I put to the review council.

Round 1 (yesterday's plan) came back REVISE and caught two things that were genuinely wrong in my
code — worth it on its own. I widened the fix as you asked: instead of just the chrome renderer, it
now also fixes the shared blanking mechanism underneath it and the second, related bug (the dead
mobile menu). Resubmitted.

Round 2 came back REVISE **again** — 6 of 11 reviewers approve, 5 object. Most of the objections are
"you asserted this, prove it", and I've now measured every one of them (the numbers are in the notes;
they all held up, including the blast radius). Two were real small bugs in my code sketch, which I'll
fix.

But there's one objection that keeps coming back, and it's not a mistake I can quietly fix — it's a
genuine disagreement about how far this fix should go, and I think it's your call.

**The disagreement in plain terms:** when the site builder can't fill in a link, my fix makes it
**shout about it in the logs** — names the exact dead link, flags it as an error. The council says
that's not enough: a log nobody reads isn't really "failing loudly", so the fix should either **stop
the build** or **file a job for someone to fix it**, not just write a log line.

They have a point. My reason for *not* doing that: we already have a pile of 34 correctly-detected
broken-link findings sitting unread because nothing consumes them (that's a separate known problem).
Adding a 35th kind of unread finding doesn't help anyone. And stopping the build outright is
dangerous right now — 30-odd live components across the fleet have this exact issue today, so a hard
stop would fail the next rebuild of most of our sites until each one is fixed.

So it's a real fork, and both sides are defensible:
- **My version (ship it as a loud log):** safe, immediate, improves the situation, doesn't pretend
  the downstream consumer exists. But the council is right that it's still "documented silence".
- **The council's version (make it block or file a job):** genuinely fixes the "fail loudly" promise,
  but it's a bigger piece of work, needs the staged rollout so it doesn't break the fleet, and needs
  something built to actually consume the jobs it files.

I don't want to just keep rewording the plan to get a green light — the council has said the same
thing twice and it deserves a straight answer, not a third revision that talks around it. **Which way
do you want it:** the safe loud-log version now, or the bigger block/escalate version as a follow-on
piece? Happy to do either; I just don't think I should pick for you.

(Everything on the actual live site is already fixed and verified — nav, logo, the free tool. This is
purely about the underlying platform fix and how thorough to make it before it ships in the next
image.)

## 2026-07-21 — the funnel bug (017): built the auditor that would have caught it

Picking up the older funnel bug — the one where, right after the cutover, clicking "Free Audience
Check" gave you a bare "POST only" page. The site itself was already fixed a couple of days ago (the
two forms are back on the tools and report pages, and I checked again today: they're still there and
working, nothing on the site points a plain click at that POST-only address any more). So the live
site is fine.

What was still missing is the *watchdog*. The whole reason that bug was scary is that nothing noticed
— every automatic check we run said the page was healthy, because they only ask "does this link go
somewhere?" and it did (it went to the tool). None of them asked the one question that mattered: "if
someone *clicks* this, will the tool actually accept a click?" A form sends its data one way (a
"POST"); a plain link is a different way (a "GET"); the tool only accepts the first, so a plain link
to it is a dead end that looks alive.

Two things I found that the old write-up got wrong, now that a few days have passed. First, it said no
automatic check had ever run on idea.uk — that's no longer true, they have run (30 findings), and
tellingly not one of them caught this funnel problem, which proves the gap is real. Second, the site
isn't even *labelled* in our records as having a tool behind it (that field is just empty), which is
why none of our backend checks ever looked at it. So I deliberately built the new check to not rely on
that label at all — it just tries the link the way a visitor would and watches for the "wrong method"
rejection.

I've written that check and proved it works against the real site: clicking-style requests to the two
tool addresses come back with exactly the "POST only" rejection it's meant to catch, while ordinary
pages, the health check, and missing pages all come back clean — so it flags the real problem and
doesn't cry wolf. It's gone into the review council (advisory) and I've committed it. It won't
actually start running until the next image ships and I switch it on for the checker — I'll do that
then. The bug stays officially open until it's live.

One piece I left for later on purpose: a second check for "this site has a tool but no form anywhere
to feed it" needs us to first fix that empty label problem (how does the platform know a site has a
backend at all?), which is really its own decision rather than something to bolt on here.


## 2026-07-21 — you decided, and the platform fix is now built

You picked the safe route: ship the loud-log version now, and do the bigger block-or-escalate version
as a separate follow-on. That's the fleet-safe one, and it doesn't wave the council's point away — it
schedules it.

So I've built the actual platform change. Until today it was only a plan — a set of sketches in the
council submission; now it's real code, committed. What it does, in plain terms: when the site builder
renders a page's header or footer, it now reads each component's OWN list of the fields it needs,
instead of only knowing a fixed shopping-list baked into the code. Anything it still can't fill it
leaves blank AND says so in the logs — and if the blank is a link or an image it says so louder (an
error, naming the exact dead control), because a dead link on a live page is worse than a missing word.
It also fixes a second, related thing while it was in there: the header and footer's little bit of
JavaScript — the mobile "hamburger" menu — was never being published at all, so that menu was dead on
every page of every site. That now publishes.

Two honest caveats. First, this is a code change, so it does nothing until the next image is built and
rolled out. The live idea.uk site is already fine — that was the database fix days ago; this is the
fleet-wide plumbing underneath. Second, I did NOT put a "reviewed by council" stamp on it, because the
council never actually approved it — they said REVISE on the one point you've now overruled. That's the
honest record: you ruled, I shipped, and the stamp stays reserved for a real green light.

The follow-on — making a dead link actually block the build or raise a job for someone, not just log —
is filed as its own bug (054) so it isn't forgotten. It needs a staged rollout and something that
actually actions the jobs it would raise, which is exactly why it's a separate piece and not bolted on
here.

One thing left for you: it's committed, so it'll ride the next fleet build whenever that happens. If
you want it live sooner I can build and roll the chassis myself — just say the word.


## 2026-07-21 (later) — it's live, and the dead mobile menu is fixed for real

You deployed a new chassis (v1.0.1146) and my change rode it. I checked it the careful way — not by
trusting the version label, but by reading the actual program running in the pod and confirming the
new code is in it and the old code is gone. It is.

Then I proved the visible half actually works. The header and footer's little menu script — the
mobile "hamburger" that had been dead on every page — now loads and runs. I nudged one idea.uk page
to rebuild (the safe kind of rebuild: no copy touched, no AI involved) so it would publish that
script, waited for the box to pick it up, and checked: both scripts now load (they were "not found"
before), the file really is the menu code, the page points at it, and — importantly — nothing else
on the page broke: every link still goes where it should. So the hamburger menu on phones now works.

There was one honest wobble worth recording: right after the rebuild finished, the script still
showed as "not found" for a few minutes. That's not a failure — idea.uk is served from the box, which
only pulls fresh files every five minutes, so there's a short lag between "chassis published it" and
"the public site serves it." A minute later it was live. (I nearly called it broken; the lag is the
answer, and it's now noted so the next person doesn't.)

So both parts of this are now done and live: the site reads its own components properly, and the menu
works. I've marked the two bugs (018 and 041) as closed. The bigger follow-on — making a dead link
actually *stop* the build rather than just shout in the logs — is still its own tracked job (054),
exactly as you decided.

---

**2026-07-22 — the "make a dead link stop the build" follow-on (054) is now built.**

This is the bigger job I flagged last time — the one you decided to split off from the chrome fix.
Two quick decisions from you today set the shape of it. First, *how* a dead link should be caught:
rather than editing every template by hand or hard-failing the whole build, I made the site's chrome
renderer itself quietly **drop** a broken link/button before the page is saved, so a dead control can
never reach a live page. Second, you confirmed the review queue is a real queue that someone works —
so alongside dropping the control, the system now files a proper job that actually gets picked up and
retried (a re-render fixes the common case where the link's data just arrived late), instead of
dumping it on a pile nobody reads.

Worth saying plainly: when I checked the live sites first, the original fire was already out — no live
page is currently shipping one of these dead links, and the only components that still could are old
stock that nothing uses. So this change is a **safety net against it happening again**, not an
emergency repair. It's written, tested (16 cases), and committed, but like all our Go changes it does
nothing until the next image roll. It's gone to the advisory review council as usual.

One small thing I'll mention because it's the sort of thing that bites: another session happened to be
editing the same file as me at the same time. I noticed before committing and moved my bit into its
own file, so I didn't accidentally sweep up their half-finished work under my change. No harm done.

## 2026-07-22 — the funnel watchdog is live; 017 closed

You rolled a new chassis image (v1.0.1149) and it carried the funnel watchdog I'd written. I checked
it's actually in the running system (the code is present in the live binary), then switched it on for
the completeness checker (a one-line database change, seed 188 — it took a safety snapshot first, so
it's a one-command undo if ever needed). So the check that would have caught the original "POST only"
funnel break is now running in production.

To be straight about what "verified" means here: I proved the check *detects the fault* by pointing it
at the real tool earlier — a click-style request to the audience-check address really does come back
"405 POST only", which is exactly what it flags — and its unit test drives that same path. What I
didn't do is sit and watch it run inside a full site-scan, mostly because there's nothing for it to
find right now (no site currently has that broken shape), so a scan would just come back clean. The
detection itself is proven; the scan plumbing is the same machinery twenty other checks already use.

So I've closed the funnel bug (017) as fixed-and-live. The *prevention* half — a pre-flight check that
stops a future cutover silently dropping a form in the first place — I've filed separately as a feature
(011) and dropped a ready-to-run version into the cutover runbook (§3d(ii)), per your call to ship the
detector now and keep the bigger prevention piece as its own thing.

## 2026-07-22 — the chrome "dead link" detector is now sharper, and it's live

Quick background: back when the idea.uk site had every link broken, the fix taught the page-builder to
notice when a link comes out empty and shout about it in the logs (an empty link is a dead button on a
live page). That fix has been live for a while. But the advisory review council kept flagging one flaw
in *how* it noticed: it was doing a crude text search of the page template, and that crude search
mistook some perfectly-fine links for broken ones — it would shout "dead link!" about buttons that
were only ever meant to appear under certain conditions, or that were part of a repeating list. On
about thirty of our components fleet-wide it would have cried wolf. That matters because the *next*
piece of work (the one that will actually *block* a broken page from going out) is supposed to act on
those shouts — and a channel full of false alarms is one everyone learns to ignore.

You ruled: ship the smarter version now rather than spend another review round arguing wording. So I
rewrote the detector to actually understand the template's structure instead of pattern-matching text,
so it now only flags a link that is genuinely dead and unconditional. I also found a second, older
copy of the same "render a template" machinery that had the same blind spot and fixed that too — though
I should be straight with you: that second copy turns out to be code nothing currently calls, so fixing
it is insurance for the future, not something doing any work today.

It's committed and — because another session rolled a new image (v1.0.1149) that happened to pick up my
change — it is already live in production. I checked the running system directly to confirm the new
code is really there (not just merged). So the false-alarm problem is gone fleet-wide.

Two honest loose ends, neither urgent: the council never formally "approved" this — it stayed at
"revise" because the council can't actually block anything and you'd told me to ship — so there's no
"reviewed" stamp on it, same as the footer fix (053) we handled the same way last week. And there's a
stale old email address (idea-uk@leopardess.uk) still sitting in one database field for idea.uk; it
isn't shown anywhere today, but it's a wrong address one wrong turn away, so it's on the list.

---

**2026-07-22 (later) — the 054 follow-on passed review, with one genuinely useful correction.**

I put the change through the advisory review council. It came back "revise" first — and it was right
to. My first version filed the "I dropped a broken link" note onto an automatic fixer that re-renders
the page and then marks the job done *without checking the link was actually fixed*. So a link that
genuinely has nowhere to go would have been quietly re-dropped and marked "done" forever, never
reaching a person. That's the exact silent-loss trap this whole piece of work exists to close, so I'm
glad the review caught it. You confirmed the fix: send those notes straight to the human review queue
instead (the same place the equivalent page-level check already sends them). The common case — a link
whose data just arrived late — fixes itself on the next normal rebuild, so nothing is lost.

Second pass came back approved. I made one more small safety tweak it suggested (leave alone the
special "fill this in later on the visitor's screen" placeholders). It's committed and reviewed, and —
as always — does nothing until the next image roll. It also quietly depends on another change that's
already live (v1.0.1149), which happens to be the same file another session was editing while I
worked; that ordering is fine.

## 2026-07-23 — the home page now sends people to the paid tool

You flagged that the home-page buttons still didn't go to the paid report tool. They didn't: the four
main call-to-action blocks on the page (the big banner, the "how it works" strip, the tools grid, and
the closing call-to-action) had button *text* stored but no *destination*, so every one of them fell
through to a default contact page — or, in one case, a dead button that went nowhere. And the
automatic link-fixer we have can't help here, because by design it only ever points buttons at
"section" hub pages (guides, news), never at the report page (which is a product/landing page). So
sending the home page at the paid tool is a deliberate choice that had to be set by hand.

I set every one of those buttons to the destination you confirmed: the report/"get started"/"learn
more"/"see how it works" ones go to the paid report page; "run the free idea check" goes to the free
taster; "browse all tools" goes to the tools page. One of them ("Get Started" in a lower block) was a
dead button hardcoded to go nowhere — I fixed the underlying template for that, in a careful way that
leaves the two other sites using the same template completely unchanged.

Our page-rebuild queue happened to be stuck (nothing had run through it for about 15 hours — a known
recurring problem elsewhere), so I pushed the rebuild through directly instead of waiting. Then I
checked the actual live page, not just the "done" status, and confirmed all the buttons now go where
they should. I've locked those blocks so a future automatic pass can't quietly undo them.

One thing left, and it's a decision for you rather than a bug: the small "Get Started" button up in the
site *header* (the one that sits on every page, not just the home page) still goes to the contact page.
Pointing it at the report tool as well is easy, but because it's in the shared header it would change
on all nine pages at once — so I've left it and asked you which you'd prefer.

## 2026-07-24 — header button done too; the whole home page now points at the tool

You said yes to pointing the header "Get Started" button at the report tool as well. Done and checked
on the live site: that button now goes to the report page on every page of the site, not just the home
page. The only thing still going to the contact page is the actual "Contact" menu link, which is right.
So the whole funnel — the big page buttons and the header button — now sends people to the paid tool.

## 2026-07-24 — owner: turn idea.uk into a full "idea lifecycle" pipeline (captured for later)

You've asked for something much bigger than the migration: you want idea.uk to become a whole guided
journey for someone with an idea — a section on **patents and what to do with an idea**, and then a
full set of **guides and tools** running from helping people create ideas, through building, testing,
user acceptance, feedback loops, patents, copyright, funding (both the *ways* to fund and the actual
*sources*), and "a whole load more in that pipeline". Along the way, tools — some free, some paid —
built using the same three-tier AI-tooling strategy that so far has only used idea.uk as an *example*
(the point being: use it to enhance idea.uk itself too).

I've written this down properly so it isn't lost — as `features_open/014_FEATURE_idea_uk_ideas_pipeline.md`
— with the pipeline stages, how the free/paid tools map onto the AI-tooling funnel, and what idea.uk
already has (the live paid report funnel) so nobody rebuilds it. I've also left a cross-note on the
AI-tooling-strategy feature (`013`) so that thread knows you now want idea.uk *enhanced*, not just used
as a reference. I've NOT started designing or building any of it — it's captured as a request, waiting
on your go and on which stage (patents first, I'd assume) you want to start with.

## 2026-07-25 — the first guide is built and live: patents

You said to carry on with idea.uk in this thread, so I've started building the pipeline itself,
beginning where you led: patents.

**First, what was actually there.** I looked before building. The site had nine pages, and the
Guides section was an empty shell — the page loaded, showed a heading saying "Guides", and then
nothing. Same for News. There was not a single guide page on the site. So this isn't an addition to
a library; it's the first one.

**Worse, the Guides page could never have filled itself in.** The bit of the page that was supposed
to list the guides was a component that reads from a fixed, hand-written list — not from the site's
actual pages. So we could have written twenty guides and that page would still have shown nothing.
I've swapped it for the component the rest of the estate uses, which asks the database "what guide
pages exist?" and lists whatever it finds. That means every future guide appears on the hub
automatically, with no extra work.

**The guide itself: "Patents: how to protect an idea in the UK".** It's live at
https://idea.uk/guides/patents/index.html — about 2,000 words, eight sections, written plainly. It
opens with the mistake that actually ruins people, which is telling anyone before you file: in the
UK there's no grace period, so a pitch without an NDA, a crowdfunding page or a conference talk can
destroy your own novelty. Then: what a patent really is and the long list of things that can't be
patented (business methods and software "as such" catch a lot of good ideas); an honest section on
whether it's worth it at all, given you'd have to fund the lawsuit yourself; how to do a free
first-look prior-art search; the timeline, and why the twelve-month priority year is the genuinely
valuable part; what it costs, split into the small official fees and the much larger attorney fees;
the alternatives, which are frequently the better answer (copyright, registered designs, trade
marks, plain confidentiality); and who to actually talk to. It ends with a clear statement that it's
general information, not legal advice, and to see a registered patent attorney.

**I wrote it myself rather than letting the site's AI write it, deliberately.** The platform has a
half-built system for checking factual claims against evidence, and it isn't switched on yet — and
we have an open case elsewhere in the estate about AI inventing statistics. Legal guidance, on a
commercial site, that invites people to rely on it, is the worst possible place to find out that
something was made up. So it's hand-written, and I'll lock the page so a later automated content
pass can't quietly rewrite it.

One thing worth telling you because it nearly shipped: my own first draft said the cheap small-claims
route makes patent enforcement affordable for small businesses. That's wrong — that court's small
claims track specifically doesn't hear patent cases. I caught it re-reading before it went in, and
fixed it. I mention it because "I caught it myself this time" is not a system I'd want to rely on,
and it's the strongest argument for getting the evidence-checking feature finished.

**Two things went wrong on the way, both worth you knowing about.** The first: the page rendered
absolutely nothing, and the job reported success. A field I'd left blank meant the renderer couldn't
find the components, so it quietly reused the (empty) previous version and reported "complete". That
is the recurring shape of trouble here — green status, nothing behind it — and it's why I check the
actual live page every time rather than the job. The second: when the next job didn't appear, I
assumed the request had been lost and re-sent it. It hadn't been lost; the whole system's queue was
stalled and *everybody's* jobs were waiting. Harmless this time (the same render twice), but I'd
written myself a note about exactly that mistake weeks ago and it didn't stop me, so I've written it
down again somewhere it'll be found.

**Where that leaves us.** The patents guide is live and correct, and every button on it points at
the £29 report. The Guides hub is switched over in the database and will show the patents card as
soon as the stalled queue drains — I'll confirm on the live page rather than assume. After that I'll
lock both pages.

**Next, unless you'd rather something else:** the natural companions to patents are **copyright**
(what's automatic, what it does and doesn't cover) and then the **funding** pair (ways, then
sources). There's also the free "should you patent this?" checker we discussed as an option — a
short questionnaire that gives a steer and then points at the report. Say which you'd like first.

## 2026-07-25 (later) — copyright guide, and the first free tool

You said yes to both, so both are built.

**First, the loose end from this morning closed.** The Guides page button that was pointing at
itself now correctly says "Get a verified idea report" and goes to the £29 report. Worth noting
because it also settled the argument with myself: the job I'd assumed was lost had simply been
queued, and it ran. The system proved me wrong rather than me spotting it.

**The copyright guide** — https://idea.uk/guides/copyright/index.html. It leads on the thing that
quietly costs small businesses the most money, and which the patents guide gives you no way to
discover: **if a freelancer made it, they own it.** Paying an invoice does not transfer copyright.
The designer owns your logo, the developer owns your code, the photographer owns the photos of your
own product — unless you have a written, signed assignment. Employees are different; their work is
yours automatically. That one section is worth the page.

The rest covers what copyright actually is (automatic, free, nothing to register in the UK), the
crucial limit that it protects *what you wrote* and not *what you thought of*, how long it lasts,
proving what you had and when, licensing and the open-source terms you're already relying on, and
what to do if someone copies you.

Two things I want to flag about how it's written. On **AI**, it says plainly that the law is
unsettled — whether AI output is owned by anyone, and whether training on other people's work is
allowed, are genuinely open questions in the UK right now. It would have been the easiest paragraph
in the guide to write confidently and wrongly, so it doesn't. And it says copyright **can** use the
cheap small-claims court route — which is the exact mirror of the mistake I caught in the patents
guide this morning, where I'd wrongly said patents could. Getting that pair right in both
directions is the useful result of that near-miss.

**The free checker** — https://idea.uk/tools/patent-check/index.html. Six questions, about two
minutes, no sign-up, nothing stored, nothing sent anywhere; it all runs in the visitor's browser,
so it costs us nothing per use. It ends on a plain verdict and then points at the £29 report.

**One design decision I'd like you to know about, because it's the difference between a useful tool
and a harmful one.** There's an existing quiz component on the estate I could have reused. It works
by scoring: add up the answers, pick a band. That is the wrong machine for this question. If
someone has already shown their idea publicly, they have very likely destroyed their own patent
rights — in the UK there's no grace period — and *nothing else they answer changes that*. A scoring
quiz would have added up their five good answers and cheerfully told them they looked ready to
patent. So this one is built differently: it checks the fatal questions first and stops there if it
finds one, and only scores the remaining commercial judgement if the idea is genuinely still in
play. Someone who has already published gets a "speak to a patent attorney this week" answer, not a
score.

It's also honest about the common case. A lot of people will get "a patent is the wrong tool for
this" — for a business model, an app with no technical angle, or a creative work — and the tool
tells them what *does* protect it instead, rather than pushing them toward an expensive answer.

**Where that leaves the site.** idea.uk now has two real guides and a free tool, both hubs list
their contents automatically as we add more, and everything funnels to the paid report. Three
months ago the Guides page was an empty heading.

**Still finishing:** cross-links between the new pages (so the patents guide offers the checker and
the copyright guide, rather than each page being a dead end), a re-render of the Tools page so it
lists the new tool, and locking the written content so no automated pass rewrites it. All three are
in progress — the box that serves the site pulls updates on a five-minute timer, so there's a wait
between each step and being able to check it properly.

**Next**, when you want it: the funding pair — the *ways* to fund something, then the actual
*sources*. That's the other half of what you asked for, and it's the stage where a free tool would
earn its place again (a "which funding route fits you" finder).

## 2026-07-25 (afternoon) — your tools-page report, the paid-tool audit, and the funding guides

You reported four things about the Tools page, and all four were real. The two buttons went to
the contact form because that page's buttons had never been given destinations — same disease as
the home page had, now cured the same way: "Try the free patent check" goes to the checker,
everything else goes to the £29 report. The diagram wasn't showing because the page pointed at an
image file that simply doesn't exist — while a perfectly good illustration made for that page sat
unused; it's now wired in. And the paid report wasn't in the tool listing because the listing
only shows pages classed as "tools", and the report page was classed as a "landing page". It's
now classed as what it is, so it appears in the listing — first, with a proper description.

While in there I found something you hadn't reported and wouldn't have liked: the page claimed
"**8** tools available free" — there are two — and "data stays on your device — always", which
isn't true of the audience check (it uses our server) or the paid report. On a site whose whole
pitch is honesty, an invented statistic is the worst kind of bug. Both now say true things.

**Your bigger question: has the paid tool changed, and does it do what we say?** I read the
tool's actual source code against the report page's copy. The tool has *not* changed — it is
what it has always been, and it's genuinely good at it: you tell it about your business, and it
generates a couple of dozen possible AI product ideas, kills the weak ones using a second
opinion from a different AI vendor, checks the survivors against the real web (do competitors
exist, would people pay), scores what's left, and sends a ranked report that includes what
didn't make the cut and why — with a human reviewing every order first, and an honest "nothing
cleared the bar" outcome when that's the truth.

But the sales page describes a different product. It says the report is "produced for a single
idea you submit" — in fact the tool doesn't assess your idea, it finds ideas *for* you. And it
says "where we cite a figure or a claim, we explain its source so you can check it yourself" —
the report contains no sources at all; its instructions actually tell the AI *not* to list
names. The parts about honesty, refusal, human review and the £29 are all true and genuinely
well built. So the product is real; the promise is misdescribed. The choice is yours: rewrite
the page to sell what the tool actually is (quick, I can do it today), or extend the tool to
match the promise — assess the submitted idea and include checkable sources — which is more
work and needs you to deploy the new version, but is arguably the better product. Or the first
now and the second later. The full claim-by-claim table is in the audit document.

One small thing from the same reading: if the box's settings file doesn't set a contact
address, the report emails fall back to the old dead leopardess address — worth a one-line
check on the box (`grep CONTACT_EMAIL /etc/idea/idea.env`).

**And the funding guides you asked for are live** — the *ways* (the eight shapes money comes
in, what each really costs, and why most early ideas should take none of them yet) and the
*sources* (the actual UK map: Innovate UK, the British Business Bank and its start-up loans,
the Scottish/Welsh/Northern Irish agencies, growth hubs, angels, the King's Trust, and how to
work the map without losing six months). Neither guide states a single amount, rate or
deadline — those change every fiscal year, and a stale figure looks exactly like an invented
one — so everything points at the institutions' own sites. Guides section now: patents,
copyright, funding ways, funding sources. Four real guides where there was an empty heading
yesterday morning.

## 2026-07-25 (evening) — you chose to extend the tool, and it's built

You decided the tool should grow to match the promise rather than the promise shrink to match
the tool. That's done, in code, tested — it now needs you to build and deploy it (the short
checklist is at the bottom of the audit document).

What the report will do once deployed: it **leads with an assessment of the idea the customer
actually submitted** — the problem and the evidence people really have it, the signs of demand
found by live web search, who else is out there, what people use instead today, where it's
defensible and where it's exposed, and one specific, affordable next step. Under it, and under
every further idea the tool proposes, sits a **"Check it yourself" list of the actual web pages
the research relied on** — so the "we explain its source so you can check it yourself" sentence
on the sales page becomes literally true. The report itself now also says plainly that AI did
the research and drafting and a person reviewed it — previously only the small print said so.

Two honesty details worth knowing. If someone submits something too vague to assess, the report
says exactly that — "this is too early to assess honestly, here's what's missing" — rather than
inflating it into a verdict; the sales page always promised that, and now the code enforces it.
And the research is told it may never invent a source: if a finding rests on general knowledge
rather than a page it read, the list stays shorter rather than getting padded.

The cost of all this is that each report will use roughly twice the AI spend and take somewhat
longer to produce — two long research passes instead of one. At £29 with human review per order,
that's a margin question you may want to look at once you've seen a few real runs.

One bonus from the same session: a test in the tool's own suite had been failing for a while —
not because the tool was broken, but because the test was checking for an old wording of the
payment email. Fixed, so the whole suite is green again and a real failure won't hide behind it.

## 2026-07-25 (late) — your deploy checked out, the blockers explained, and the journey is complete

Three findings from your deploy attempt. First: **your deploy worked** — the mangled output was
just the pasted command splitting in two; the service restarted cleanly at 15:11 running the new
binary. Second: the "at capacity" error is **five of your own old test orders** (June and mid-July,
all to your address) sitting in "awaiting review" and "awaiting payment" — the tool counts those
against its limit of five, so every slot is taken by a ghost. Third: your grep found a real bug —
the settings file sets CONTACT_EMAIL **twice**, and the second, stale leopardess address is the one
that wins. So report emails have been carrying a dead reply address. I wasn't permitted to change
the live box myself (reasonable), so the exact fix commands are in the chat above: back up, release
the five slots, delete the stale line, restart. One thing to confirm yourself: the surviving line
says idea@contactforsales.com while the website uses idea.uk@contactforsales.com — check which
mailbox actually exists.

**And the guide journey is now complete.** Five new guides tonight: creating ideas, building it,
testing it, user acceptance, and feedback loops. The Guides page now reads as the whole journey in
order — create → build → test → acceptance → feedback → patents → copyright → funding ways →
funding sources. Nine guides, every one hand-written, every one locked against automated rewriting,
and every one funnelling to the £29 report. Two days ago that page was an empty heading.

## 2026-07-25 (last entry of the day) — the funding finder and the report page, both done

**The funding-fit finder is live** — https://idea.uk/tools/funding-fit/index.html — and listed on
the Tools page alongside the others. Seven questions, two minutes, nothing stored. It's built on
the same principle as the patent checker: the killer questions come first. If the money is really
for living costs while you work things out, it says the honest thing nobody advertises — almost
nothing funds that, and knowing it now saves you months of applications. If there's no evidence
anyone wants the idea yet, it says: buy evidence first, cheaply, here's how. Only past those two
gates does it map your answers onto actual routes — grants if the work is genuinely novel, angels
or crowdfunding if you'd sell a share, bank routes if you can carry repayments, customer money if
none of those fit — with the free front doors (growth hubs, the devolved agencies, the King's
Trust for under-30s) signposted and a warning about anyone who charges to open free doors.

**And the report page now tells the whole truth in both directions.** Since your decision to
extend the tool, the page's promises (assessing your idea, sources you can check) are delivered —
and I've added the part the page never mentioned: that the report also generates further ideas
around your business, checks them against the live web, ranks the survivors with their own
sources, and shows you what didn't make the cut and why. Two edits, verified live.

**The Tools page now reads:** Verified Idea Report (£29) · Should you patent it? (free) · Which
funding route fits? (free) · Free Audience Check. The Guides page carries all nine guides in
journey order. Everything hand-written is locked; everything that lists is derived and will pick
up whatever we add next; and every path ends at the report.

Still with you: the three box commands from earlier (release the five stuck test slots, delete the
stale email line, restart) — then one real end-to-end report to see the new format arrive.

## 2026-07-26 — where things stand, and the last chores on the box

The site itself is done and verified — nine guides, the two free finders, the audience check and
the paid report, all cross-linked, all funnelling to the £29 report. A fresh summary of the whole
journey is in SUMMARY_2026-07-26.

On the box, I checked this morning: none of yesterday's three fixes has gone through yet — the
capacity endpoint still says five of five slots taken (your old test orders), and the settings
file still has the contact address twice with the dead one winning. The exact steps are in the
chat: back up, release the five slots, fix the email line (deciding which contactforsales
address is the real mailbox), restart, then order one real report to see the new format arrive.
One nice side effect already banked: yesterday's deploy also took the anti-spam hardening live —
it had been sitting finished in the source since mid-July waiting for exactly this rebuild.

## 2026-07-26 (correction) — the leopardess address is the right one for the tool

You've told me idea-uk@leopardess.uk is the correct address for the tool — I had it backwards. I'd
carried "leopardess is old" over from the website work, where you chose a contactforsales address,
and assumed the tool should match. They're deliberately different. The comfortable consequence:
because the last line in the settings file wins, the tool has been using the right address the
whole time — nothing was ever mis-sent. The only real tidy-up is deleting the *other* line (the
contactforsales one), so nobody later "fixes the duplicate" by deleting the wrong half. The
corrected command is in the chat; everything else in the walkthrough stands.

## 2026-07-26 (evening) — you cleared it, and here's the handover

Confirmed from here: the queue is open again (nothing active, five slots free), the five old test
orders are marked "expired" rather than declined so your records stay honest, and the duplicate
email setting is gone. All good.

Two things I've written up so you can pick this up in a fresh chat whenever you like. There's a
new summary (SUMMARY_2026-07-26b) explaining why this morning's "just four commands left" turned
out to be a real design gap — the tool had no way to release an order that went cold, which is why
clearing by hand kept failing. And the resume handoff is rewritten so it opens with today's state
instead of last Tuesday's; the old text is kept underneath.

The handoff leads with the one thing genuinely outstanding: **nobody has yet received a report in
the new format.** Everything is built and deployed for it, but until you run one real order all
the way through — submit, confirm, read the draft, approve, pay — the biggest change of the week
is unproven. That's the first thing I'd do.

Behind it sits a second deploy that's written and tested but not on the box: automatic expiry of
cold orders, so the queue can never silently close itself again. Build-and-copy commands are in
the handoff.

One useful thing landed today from elsewhere in the estate: the platform's page-locking now
actually enforces. All the guide and tool text we wrote is genuinely protected from being
rewritten by an automated pass — which it wasn't, quite, until this afternoon.

## 2026-07-26 (late) — it works, and it's proven

You ran a real one through and said the report was good. I checked the stored copy rather than
take that at face value, because "looks right" has fooled us more than once this week — and it
holds up properly: about 13,000 words' worth of report, opening with the assessment of your actual
idea, sixteen source links you can click and check, the AI disclosure in the report itself, and no
trace of the old opening line. Declining it was the right way to close the test — it releases the
slot, costs nothing, and sends a polite note.

So the biggest change of the week is no longer a claim, it's a delivered artefact. The two
promises on the sales page that were false a day ago — assessing the idea you actually sent, and
sources you can check yourself — are now true of what a paying customer receives.

Two things worth keeping an eye on across the next few real orders: how long they take (two
research passes now, not one) and what they cost us (roughly double, at the same £29). Neither is
a problem yet; both are worth a look once there's real volume.

The one job left on the box is the second deploy — automatic expiry of cold orders, so the queue
can never quietly close itself the way it did this morning. Commands are in the handoff.

## 2026-07-26 (evening) — the deploy is done, and on the way I found two ways the tool was losing money

Short version: the second deploy is on the box. While preparing it I read through the tool's code
properly and found two live faults, both now fixed and both proven fixed against the real running
service. Neither had bitten us yet. One of them would have, sooner or later, and it would have
looked like nothing at all.

**The first one: someone could have taken the £29 report without paying.**

There is a shortcut in the code so we can test a purchase end-to-end locally without Stripe. You
add `fake=1` to a web address and the order is marked paid. The intention was always that this
only works in local testing — the code even says so in a comment, in capitals. But nobody had
made the code *check*. It looked at the web address, saw `fake=1`, and marked the order paid. On
the live site. With real Stripe running.

The reason this matters more than it first sounds: the order number isn't a secret from the
buyer. When someone starts a real payment and then clicks "cancel", Stripe sends them back to a
page on our site with their own order number in the address. So a customer who paid attention for
ten seconds could have gone back, added `fake=1`, and been sent the report for nothing. No
hacking, no guessing — one visible address and one word.

It's fixed: the shortcut now only works when the fake payment system is actually the one running,
which it never is in production. And an attempt is written to the log, so if anyone does try it,
we'll see it. I proved the fix on the live service against a genuine order that was sitting
mid-purchase — I ran the attack myself, from the outside, and the order stayed unpaid.

**The second one: anyone could pretend to be a different visitor.**

The free "audience check" on the site does a real AI call every time someone uses it, and it
costs us a couple of pence each time. To stop that being abused, we limit each visitor to three
goes an hour. The limit works by looking at the visitor's internet address.

Except we were reading that address from a piece of information the visitor is allowed to write
themselves. I sent a request claiming to come from an address that cannot possibly exist, and the
service dutifully wrote it down as mine. Which means the three-an-hour limit was never really
three an hour: change the claimed address each time and you get as many free AI calls as you
like, at our expense. The same wrong reading was being stored against every order — and we'd been
storing it specifically so we could block troublemakers later, which would not have worked.

Also fixed, also proven live: the same request that fooled it an hour ago now correctly records
my real address.

**About that end-to-end test you asked for.**

You'd already done it. Another session ran a full report at lunchtime and updated the handover
notes at 15:57 to say so — but I'd read that document at 15:34 and never looked at it again, so I
asked you a question about proving something that was already proven, and then started a second
report run. That's my mistake, and I've written it up properly where we keep such things. The
check that would have caught it takes two seconds and I didn't do it.

It wasn't wasted, but that's luck rather than judgement: your lunchtime order was *declined* at
the end, so the payment half of the chain — approve, send the pay link, take the money, deliver —
had still never actually run. This one is being taken through exactly that.

**So there is one thing waiting for you.** There's an email at your usual address with a Stripe
link for £29. Paying it is the last untested step in the whole product. It's real money on a real
card (you'd be paying yourself, less Stripe's fee of about a pound), so it's your call — but
until someone does it, "customer pays and receives the report" remains the one part of idea.uk
that has never happened.

**On the report itself.** It's good. The vet price-comparison idea I submitted came back with a
genuinely honest assessment: it told the submitter the market is about to be flooded by a free
government comparison service, that a dozen private rivals already exist, and that the sensible
move is a £50–£100 test in one town before writing any software. That is exactly the tone the
sales page promises and it would have been very easy for it to flatter instead. Sources are real
and checkable. It took nine and a half minutes, not the twenty to thirty we'd been expecting.

Three small copy faults did show up in it — a doubled full stop, a scoring line that reads "out
of 5" with the number missing, and a sentence that reads "using A form the receptionist fills
in". All three are now fixed in the code, and they'll take effect at the next deploy. They only
turned up because someone read the actual report rather than checking that the job succeeded.

**One thing I'd flag for the margin question you raised.** The report engine is still pointed at
last generation's AI models. They're set in a config file, so changing them needs no rebuild at
all — but which model writes your paid product is your decision, not mine, so I've left it alone
and noted it.

**And one gap I found but deliberately didn't fix.** The new automatic expiry releases orders
that are stuck waiting for us or waiting for a customer's payment. It doesn't release one that
was mid-report when the service restarted — and since the report runs in memory, a restart always
strands one. That's the same "slot lost forever" problem we just fixed, by a different door. The
fix is small but it's only safe to write when nothing is running, and something was running all
evening.

## 2026-07-26 (late) — the report engine was a generation behind, and updating it was nearly a disaster

You asked me to move everything onto the newest AI models. The engine was running last
generation's — fine, but a year behind what we could be using. It's now on the current ones, live
on the box, and I've checked it end to end.

What's worth telling you is what nearly happened.

I'd said earlier in the evening that the models were "a lever needing no rebuild" — that you could
change which AI writes the reports by editing one line of configuration on the server, no
programming needed. **That was wrong, and wrong in the way that costs money.** If you or I had
done that, every single report would have failed instantly with an error, and the tool would have
been dead until someone worked out why.

The reason is a small piece of logic buried in the engine. Different generations of these models
want to be asked to "think" in different ways, and the code had a list of the models that use the
newer style. Anything not on that list got the older style. That sounds sensible and is exactly
backwards: a model the code has never heard of is almost certainly a *newer* one, so the code was
guaranteed to get every future upgrade wrong. I caught it because I tested the old and new
approaches against the live AI service before changing any code — the old style came back with a
flat refusal, in writing.

So I turned the list inside out. Now the code only uses the old style for models it specifically
knows are old, and anything unfamiliar gets the modern treatment. The next upgrade won't hit this.

Two smaller things fell out of the same job. There were two places where the engine relied on a
default that has quietly changed meaning between model generations — it used to mean "don't think
about this", and now means "think as much as you like", which would have eaten into the space
reserved for the actual answer and could have cut reports off mid-sentence. Both are now explicit.
And I found that when a reply *did* get cut off, the engine didn't notice: it took the half-answer
and carried on as though it were complete. That now stops and says so, rather than putting a
fragment in front of a customer.

**On cost, I'm deliberately not giving you a number.** The new top-tier model costs exactly what
the old one did, and the cheaper model is currently on an introductory rate that's *below* what we
were paying. Pulling the other way, the new models count text slightly differently and use a few
more tokens per job. I could guess the net, but I'd rather not — the next real report will tell us
precisely, because the engine logs its own token counts on every call.

**Still waiting on you: the £29.** The pay link from yesterday is in your inbox. Everything else in
the chain has now run in production at least once; taking a payment and delivering the report off
the back of it is the one thing that never has.
