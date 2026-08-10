# PLAN — webdesign.uk built and hosted on a VM, through the framework

**Written 2026-08-04, at the owner's direction: "plan the creation and hosting of
webdesign.uk on a vm, I'd like to do it all through the framework".** Grounded in
the same day's measurement (`SUMMARY_2026-08-04b_dynamic_site_capability.md`): the
VM path is live in production, so this plan is mostly *reuse*, not construction.

Read with: `HANDOFF_2026-08-04_continue_here.md` (state + blockers) ·
`vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md` (the estate ruling
this must not diverge from) · `traffic_probe/HANDOFF_vm_sites_permanent_thread.md`
(Thread B is this plan's ancestor) · `idea_uk_vm_site/box/` (the scripts to reuse).

---

## 1. What "all through the framework" means here, honestly

The framework today does three of the four things needed, in production:

| piece | who does it | evidence |
|---|---|---|
| **Build** the site (plan → content → design → render) | **framework** | the running webdesign.uk build (blocked on `bugs_open/192`, but the mechanism is proven fleet-wide) |
| **Deploy** the pages to a VM | **framework** | `git_deployer_actions.go:95-101` — per-site hop: `sites.github_repo='vm-sites'` routes every deploy to the box's repo instead of B2. relojistas.com: 20 pages, deployed 2026-08-04 |
| **Monitor** the box | **framework** (built, unseated) | `check_backend_unreachable` probes `/health` on `deploy_config.target='vm'` sites — enablement tracked in `bugs_open/149` B1/B1a |
| **Generate the backend** (the chat service) | **NOBODY — hand-written, once** | the measured boundary: `site-engine` is one hand-written Go binary; register DYN-001 tier 2 is aspirational; CTS-049's component gate has no column to hang on |

And one thing the framework does not do yet by *ruling* rather than gap:
**provisioning the box is a human running a script from this repo.** The estate
plan's owner ruling is to merge the *generator* (one profile schema, one renderer,
one drift check) but **not** the control path ("merge the generator, not the trust
boundary"). None of that is built (estate P1–P5 all pending), so this box is
provisioned the way idea.uk's was — by hand, from versioned scripts — while
**conforming to the estate profile shape** so it becomes a row in that system the
day it exists, not a fourth divergence.

**The honest sentence: the framework builds, deploys and monitors the site; a
human provisions the box once and writes the chat service once; and both of those
artefacts live in this repo where the estate plan will absorb them.**

## 2. Target architecture

```
                     Cloudflare (zone: webdesign.uk, proxied)
                          │ cloudflared tunnel (outbound-only from the box)
                          ▼
   Mythic Beasts VDS ─ nginx
   ├── location /            → /var/www/webdesign.uk/   (framework-built pages)
   ├── location /api/chat    → 127.0.0.1:8081           (hand-written intake service)
   ├── location /stripe/     → 127.0.0.1:8081           (no rate limit — Stripe retries in bursts)
   └── location /health      → 127.0.0.1:8081           (what check_backend_unreachable probes)

   pages arrive:   chassis → git-adapter → vm-sites repo (webdesign.uk/ folder)
                   → box pulls: sitesync.timer, every 5 min, reset --hard + rsync
   nothing inbound: no public IP needed, no origin firewall to maintain,
                    CF-Connecting-IP unforgeable through the tunnel
```

Decisions inherited rather than taken (each has a worked precedent):

- **One host, no `api.` subdomain.** Same origin kills CORS and the second cert.
  idea.uk's layout verbatim (`box/proxy_tool.conf`, `box/proxy_stripe.conf`).
- **Pull deploy, not push.** idea.uk's `sitesync` (5-min timer, `reset --hard`,
  rsync per-domain folder), not relojistas' self-hosted GitHub runner. Pull is
  the owner's island ruling, it needs no runner maintenance, and it makes
  ordering safe: **the repo can fill before the box exists** — the box catches up
  on first sync. Known trap already solved in the script: ssh resolves `~` from
  the passwd entry, not `$HOME` (`bugs_open/016`), hence the explicit
  `GIT_SSH_COMMAND` identity.
- **Tunnel ingress, not public TLS.** The island profile
  (`{tunnel, …, pull_agent}` in estate §3.2), not relojistas' certbot profile.
  No inbound ports, no certbot, no origin firewall step, IPv4 not required
  (Mythic Beasts charge for it).

## 2a. One box or several? (owner question, 2026-08-04 evening)

**The question:** rent one Mythic Beasts box and host all the dynamic sites on it
— idea.uk, relojistas.com, webdesign.uk, and more coming, including sites
customers order through webdesign.uk?

**Short answer: many sites per box is fine and is already the estate's design.
The trouble is not density, it is mixing three different trust classes.** Sort
the sites by class and the answer falls out per class.

### Multi-site per box is the built pattern, not an aspiration

The site-engine deploy artefacts are already class-level and multi-domain:
webroots are `/var/www/vm-sites/<domain>` (CTS-050), `sitesync` rsyncs one
folder per domain, nginx adds a vhost per site, one `cloudflared` tunnel routes
any number of hostnames, and the vm-sites thread already ruled this way for its
own sites ("wayfaringlondoner … goes on an EXISTING box — no new boxes for
now"). Resource-wise these are thin Go services plus static files; a 2-core/4 GB
box hosts a handful without noticing. **Density is not the risk.**

### The three trust classes, and what each gets

| class | sites | where it lives | why |
|---|---|---|---|
| **Money-live** — real Stripe keys and orders on disk | idea.uk today | **stays on its own box** | §6c stands, in both directions: `setup.sh` has box-takeover semantics (`ufw --force reset`), and a live earning box must not share a disk with an anonymous-input spend faucet. It is also *freshly* secured (CF + origin firewall, 08-02) — migrating a working money machine has real risk and zero upside |
| **Our product sites** — marketing/intake, spend faucets, Stripe **test** | webdesign.uk now; the "few more coming" | **the new Mythic box, together** | same trust class, same profile (tunnel, pull-sync, thin services). Each new site = one repo folder + one vhost + one tunnel hostname — marginal cost ≈ zero, exactly what the pull model is for |
| **Customer deliverables** — sites ordered through webdesign.uk | future | **B2, not a box at all** | §6a.i: chassis-built sites are **static by construction**. A customer site needs no VM; it deploys down the same Worker→B2 path as everything else. The box is never in a customer's serving path |

**The third row corrects the question's premise, and it is the important one:
customer sites do not consume box capacity.** They are static, they go to the
bucket, and scale in customers is scale in *objects*, not in servers. The only
customer site that would ever need a box is one sold with a backend feature —
which the framework cannot generate today (DYN-001 tier 2), so it would be a
deliberate, priced, per-case decision, and *that* is the moment to think about
customer isolation (it is already flagged in §6c's second caution and the P3
isolation decision).

### So, concretely

- **Rent ONE new box now** — for webdesign.uk and every future site of ours in
  the same class. If several chat-bearing sites are genuinely coming, take 8 GB
  instead of 4 at order time (still cheap, saves a resize).
- **Do not migrate idea.uk onto it.** Revisit only under the estate plan's
  profile work (P1–P2), when a rendered profile and a drift check make a
  migration verifiable rather than hopeful — and even then the money-live class
  arguably keeps its own box for the §6c disk-sharing reason alone.
- **relojistas: leave it, migrate later if the estate work makes it free.** It
  works, it deployed today, and it is the odd one out only in ingress profile
  (public TLS + push runner). Fold it in when profiles exist, not before.
- **One failure domain per class, not one per estate.** The consolidation that
  saves ~£10/month by putting live Stripe keys, a hostile-input chat and every
  future product site behind one kernel is the version that is asking for
  trouble: one compromise or one bad provisioning run takes all revenue at once
  — and we have already *had* the bad provisioning run (`ufw --force reset`).

## 2b. The order sheet (owner question, 2026-08-04 evening: "full spec — ssd? backup?")

Checked against mythic-beasts.com on 2026-08-04 (virtual servers page + IPv6
support pages), not from memory. Their configurator prices the combination;
the page states "from £4.90/month", annual = 12 months for the price of 10.

| knob | order | why |
|---|---|---|
| Product | Virtual Server (VPS) | dedicated cores buy nothing for an I/O-bound chat that spends its life waiting on the Anthropic API |
| Cores | **2** | concurrency is capped by the §5.1 spend ceiling long before CPU |
| RAM | **8 GB** | was 4; the owner has since said more product sites are coming to this box (§2a) — 8 saves the resize |
| Storage | **50 GB SSD** — their page offers "choice of SSD or HDD"; **take SSD** | transcripts/orders are text; container-less Go binaries are small; 50 is comfortable headroom |
| OS | Ubuntu 24.04 LTS | estate standard; keeps the `setup.sh`/provision lineage usable |
| Zone | **UK (Cambridge or London)** | chat transcripts are customer data from UK businesses; keep them in the UK. Latency to Cloudflare LHR is a bonus |
| IPv4 | **take one** — CORRECTED, see below | ~pennies against the product's £1,200 price point |
| Bandwidth | smallest tier | static pages ship from the box via CF; chat traffic is tiny |
| Add-ons | **"Backup space"** add-on: optional second copy; our own backup is the load-bearing one (below) | |

### The IPv4 correction (this section supersedes "IPv4: not required" in §2, Phase 2 and HANDOFF §4)

Measured 2026-08-04: **`github.com` and `api.stripe.com` have no AAAA records at
all.** An IPv6-only box cannot reach either natively — and those are the
**deploy path** (`sitesync` pulls from GitHub every 5 minutes) and the **money
path** (Stripe API calls at P3). `api.anthropic.com` and the Cloudflare tunnel
edge are both natively v6, so the chat and ingress are unaffected.

Mythic Beasts provide **free NAT64/DNS64** for IPv6-only servers (confirmed on
their IPv6 support page; resolvers `2a00:1098:0:80:1000::12`,
`2a00:1098:0:82:1000::10`), so IPv6-only **does work** — git pulls and Stripe
calls transit their shared NAT64. That is fine for P1.

**But take the IPv4 anyway:** it removes a shared-infrastructure dependency from
the two paths that matter most (deploys self-heal on the 5-minute timer, but
Stripe calls at P3 would fail closed during any NAT64 wobble), and it costs a
rounding error. The frugal variant — IPv6-only now, add IPv4 when Stripe goes
live — is legitimate; say which at order time so nginx/`cloudflared` are
configured once. **Inbound is unaffected either way**: everything arrives
through the tunnel, so the IPv4 is never listened on and needs no firewall
thought.

### Backup, designed rather than bought

The box is **rebuildable by design** — that is what the pull model buys:

| data | uniqueness | protection |
|---|---|---|
| site pages | none — `vm-sites` repo is the source of truth | re-pull; nothing to back up |
| nginx conf, units, provision scripts | none — versioned in this repo (`box/`) | re-run provisioning |
| chat service binary | none — built from this repo | redeploy |
| **transcripts, orders, request log** | **THE unique data** — and per §5.1 it is the demand signal P1 exists to collect | **nightly dump → encrypt → push to `personae-prod-uk001-backups` (B2)**, the island's proven pattern (`pg_dump` + off-box copy). Off-box copy is push-from-box, outbound-only — consistent with the tunnel posture |

So recovery-time is: provision script + first `sitesync` pull + restore last
night's dump. The Mythic "Backup space" add-on is worth taking as a *second*
copy of the dumps (different provider trust domain than B2), but it is not the
primary — a provider-side backup of a box whose disk is 95% rebuildable
artefacts protects the wrong bytes.

**Two rules that make it a backup rather than a hope:** encrypt before upload
(transcripts are customer conversations; B2 is a shared bucket), and **restore
one dump, once, before go-live** — an untested backup is a `[UNMEASURED]` claim
with your revenue attached. Put the restore drill in the provisioning runbook.

## 3. Phases

Ordering is chosen so **nothing is ever publicly broken**: the holding 302 stays
up until the very last step, and every phase before it is invisible.

### Phase 1 — flip the site to the VM class (framework side; do NOW, it is one statement)

```sql
UPDATE sites SET
  github_repo  = 'vm-sites',
  deploy_config = deploy_config
    || '{"target":"vm","capabilities":["backend"]}'::jsonb
WHERE domain = 'webdesign.uk';
```

Do this **before** `bugs_open/192` clears, precisely because the build is blocked:
when the landing page finally builds, its **first** deploy then goes straight to
`vm-sites/webdesign.uk/` instead of leaving a stale copy in the B2 repo. DB config
is live immediately (CLAUDE.md); no image roll involved. Setting
`deploy_config.target` now also means the box is monitored from its first breath
once 149 seats the check — and note **idea.uk's missing flag is the B1a trap; do
not recreate it here.**

Verify: `SELECT github_repo, deploy_config FROM sites WHERE domain='webdesign.uk';`

### Phase 2 — the box exists (owner action + one script run)

1. **Owner orders the VDS** — spec per `HANDOFF_2026-08-04` §4: 2 cores, 4 GB,
   40–60 GB SSD, Ubuntu 24.04, **IPv6-only is fine** (tunnel = no inbound).
2. **Owner supplies access** (an SSH key for the initial login; nothing else).
3. Adapt `idea_uk_vm_site/box/provision-pullsync.sh` + `sitesync` for
   `webdesign.uk` (parameterise `WEBROOT=/var/www/webdesign.uk`, repo folder,
   box name in the key comment). **Keep the adapted copies in this lane's
   `box/` directory** — versioned, so estate P1 ("extract the truth") reads them
   as data later. The script pauses for the deploy key to be added to the
   `vm-sites` repo on GitHub (owner or a session with `gh` access).
4. nginx vhost from idea.uk's pattern: static root + the three proxy locations
   (§2). Include `cloudflare-realip.conf` **but see Phase 4's warning — with a
   tunnel the real IP arrives differently.**
5. `cloudflared` tunnel: create, route `webdesign.uk` + `www` to it. **The tunnel
   writes its own DNS records — do not pre-stage them** (this lane's own
   RUNBOOK gotcha from 07-31: staged records get overwritten).

Exit test: `curl -H 'Host: webdesign.uk' http://127.0.0.1/` on the box serves the
synced pages; the tunnel is up but the public DNS still 302s (Phase 6 flips it).

### Phase 3 — the framework fills the box (no new work; watch it happen)

When 192 clears, re-drive the failed `needs_page` (plan_id
`4ecaa120-1fa6-4de1-9cd0-39d60c64b729`). The cascade completes → git-adapter
commits to `vm-sites/webdesign.uk/` → `sitesync` pulls within 5 minutes.

Read the built copy **against the seed's rules** before cutover: no em dash
survived, no "a person checks it" phrasing, price stated as £1,200 total. If an
em dash IS present, check the `banned_claims` sweep coverage first, not the
writer_block (the seed's own `[UNVERIFIED]` note).

### Phase 4 — the chat service (the one hand-written piece)

A small Go service, stdlib-first like `site-engine`, listening on `127.0.0.1:8081`.
**Written once, versioned in this lane's `box/` directory, deployed like the
site-engine binary is.** Endpoints: `POST /api/chat` (the input box), `GET /health`
(for the discovery check), later `POST /stripe/webhook`.

**Sibling of `site-engine`, not an extension of it** — recommendation, owner may
override: site-engine is the *traffic-capture* class artefact shared by other
boxes; the intake chat is product-specific, takes secrets (Anthropic key, later
Stripe), and must be replaceable without touching the estate's capture binary.

The §5.1 controls are **in the service from its first commit**, not added after:

1. per-IP limit — **through a tunnel the client IP is in `CF-Connecting-IP` and
   nginx's `$remote_addr` is the tunnel's localhost**, so key on the header, and
   prove it with the `bugs_open/139` two-network check
   (`count(DISTINCT ip) > 1`); one machine cannot tell a constant from a key;
2. hard turn cap per conversation;
3. per-day spend ceiling that **fails closed to the contact details**, not to an
   error page;
4. request log with tokens + cost per call;
5. transcripts stored as structured rows — they are the demand signal P1 exists
   to collect.

Model: `claude-haiku-4-5` (intake is not the product; §7b). The Fable pre-flight
(retention ≥30 days, no `temperature`/`top_p`/`thinking` in the call layer)
applies only when a paid *build* is later wired, not to the chat.

Stripe: **test mode**, orders stored on the box. Live money is P3 of the main
plan and is gated on written terms (§7a).

### Phase 5 — the chat reaches the page through the framework

The input box on the page should be a **pinned section** in the site plan
(roadmap `section_types`, which the planner honours — the vm-sites thread's own
mechanism), whose markup posts to `/api/chat` on the same origin. Client-side
behaviour follows the estate's own generation-time guards (CTS-044): external
loader file, no inline script, `data-runtime-fill` shell.

**Do not** try to express "this component needs a backend" in the component
library — CTS-049's gate has no column to hang on (measured 2026-08-04). The
pairing "this site has `capabilities:["backend"]`, this section is pinned on this
site" is the whole of the safety story today; write that in the site's
`site_config` notes so the next planner run has it in context.

### Phase 6 — cutover (about a minute, all Cloudflare)

Preconditions: Phases 2–5 verified on the box via `curl -H 'Host: …'`.

1. **Resolve the Worker binding question first** (HANDOFF §3 `[UNVERIFIED]`):
   `webdesign.uk` reaches `portfolio-sites-router` with **no zone route**, so it
   is probably a **Workers Custom Domain**, and that binding must be *removed* in
   the dashboard (Workers & Pages → the script → Domains & Routes) or it will
   keep claiming the hostname regardless of DNS.
2. Delete the holding Page Rule (`b8e08b35028315a274b2f5c7fea9154d`).
3. Let `cloudflared` write the tunnel DNS records for apex + `www` (delete the
   `192.0.2.1` placeholders it replaces).
4. Verify from outside: `HTTP/2 200` + `text/html` on `/`, the chat POST works,
   `/health` returns 200 through the tunnel, and `preview.ugg2.com` still serves
   (the ugg2 path must be untouched by all of this).
5. Delete the stale hand-built objects: `portfolio-sites/webdesign.uk/index.html`
   (+ `preview.ugg2.com/index.html` if desired).

### Phase 7 — the estate absorbs it

- Seat `backend_unreachable` **with** the idea.uk `deploy_config` fix — that is
  149's call (B1/B1a); contribute, don't duplicate.
- When estate P1–P2 land (profile schema + renderer), this box's nginx conf and
  units become a rendered profile: `{tunnel, systemd_binary+chat, none→decide,
  pull_agent}`. Nothing here should make that harder — which is why every
  artefact stays in the repo and the profile vocabulary above is used verbatim.
- Register the chat service in the concept register when it exists (a new
  callable thing another workstream could reach for).

## 4. What I need from the owner, in order

| # | what | blocks |
|---|---|---|
| 1 | Order the Mythic Beasts VDS (spec §Phase 2; IPv6-only OK) + initial SSH access | Phases 2–6 |
| 2 | Add the box's deploy key to the `vm-sites` GitHub repo when the script pauses | Phase 2 |
| 3 | Cloudflare dashboard: the Worker-binding look (+removal at cutover), tunnel creation if not done via `cloudflared login` on the box | Phases 2, 6 |
| 4 | Anthropic API key for the chat service (scoped, not the fleet's) | Phase 4 |
| 5 | Decision: chat as sibling service (recommended) or folded into site-engine | Phase 4 |
| 6 | Still open from before, not blocking: correction fee number; written terms before live Stripe | Phase 6+ |

**Not needed from the owner:** anything about the site build itself — that is the
framework's job and it is already running; and no CF token IP fix is required for
this plan (everything Cloudflare-side here is dashboard or on-box `cloudflared`).

## 5. Risks and traps, each with its landmine

- **`bugs_open/192`** blocks Phase 3 only. Phases 1–2 proceed regardless; the
  pull model makes the ordering safe by construction.
- **Dispatch ≤300s after a chassis roll is silently dropped** — re-drives of the
  failed page must respect the standing rule.
- **cloudflared writes its own DNS** — never pre-stage tunnel records (RUNBOOK).
- **The Worker Custom Domain** may claim the hostname independent of DNS — it
  *must* be found and removed at cutover, and the token cannot see it (no
  account scope); dashboard only.
- **`sitesync` ssh HOME trap** (`bugs_open/016`) — already encoded in the script;
  do not "simplify" the explicit `GIT_SSH_COMMAND` away.
- **Real-IP through a tunnel ≠ real-IP behind proxied DNS** — `set_real_ip_from`
  Cloudflare ranges is the *proxied* pattern; through a tunnel the peer is
  localhost. Trust `CF-Connecting-IP` at the tunnel boundary and verify with the
  two-network check (`139`).
- **A green first run of `backend_unreachable` proves nothing** until the
  denominator includes this box (B1a's exact lesson).

## 6. What this plan deliberately does not do

- No new platform code, no council submission needed: every mechanism used is
  already live (`vm-sites` hop, pull-sync, tunnel profile, pinned sections). The
  chat service is box-side, not platform code.
- No second box for previews — `*.ugg2.com` → B2 stays exactly as proven.
- No attempt to make the framework generate the backend. That is DYN-001 tier 2;
  if this product sells, that tier is the natural next investment, and this box
  becomes its reference case. Filed as direction, not scope.

## 7. Pricing decision (owner, 2026-08-10) — deposit, traffic, and a future tier

**£75 non-refundable deposit, live.** Researched comparable AI website-builder
pricing (Lovable: subscription, $25-50/mo; Durable: $12-25/mo — neither is a
direct match since both are self-serve, not done-for-you) and measured this
site's own actual build cost from `llm_call_log` (roughly $1.50 in text
generation against confirmed Anthropic rates, likely $5-10 total with
imagery) before recommending a number. Both anchors pointed well below the
owner's hoped-for £80-150 range; landed on £75 as a meaningful commitment
(6% of £1,200) that still reads as "competitive and a bit cheaper" than a
comparable tool, not a cost-recovery calculation. Implemented 2026-08-10:
`evidence_base` (facts + writer_block), the three live pages that stated
"full refund", and the chat bot's `systemPromptFacts` all updated in
lockstep and verified live.

**Owner's stated goal for this phase: enough traffic to find customer-handling
bugs, not to get flooded.** Raised whether a technical traffic limiter is
needed. **It isn't, yet** — there is no live self-serve payment flow (Stripe
is still test-mode per §4's ledger), so the actual bottleneck today is the
owner personally building each site by hand. The natural limiter already
exists. Revisit once Stripe goes live and orders can land unattended without
a human in the loop.

**Owner also raised the "pay £75, copy the site, refund the rest" risk**
explicitly, and asked to be talked out of the worst case rather than have it
engineered around pre-emptively. Discussed rather than built: the worst case
(customer pays £75, gets a real 5-page build, declines) is not actually bad
revenue — it's in the range of what a comparable AI tool charges for a month
of the customer's own effort, and it costs the customer real work to re-host
static HTML they don't own the domain for. No anti-abuse mechanism was added;
this low-volume phase is how the owner finds out whether the fear is real
before building anything against it (see CLAUDE.md's "survey the premise
before building" practice — the same logic applies to a fear as to a feature).

**Future direction, not current scope**: a **£19 all-in tier** — a full
website including static hosting, presumably closer to genuinely self-serve
— was raised by the owner as a later product, once the £1,200 done-for-you
tier has real customer-handling data behind it. Not designed, not scheduled;
recorded here so it isn't lost.
