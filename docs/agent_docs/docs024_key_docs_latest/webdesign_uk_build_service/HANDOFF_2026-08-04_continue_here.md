# HANDOFF 2026-08-04 — webdesign.uk: from here to a live VM with the chat box

**Start here if you are picking this lane up cold.** Read alongside
`PLAN_2026-07-28` (§5.1 chat controls, §6a delivery, §6c the box, §7a offer,
§7d price), `RUNBOOK` (Cloudflare + B2 + build queries) and
`SUMMARY_2026-08-04`. The two earlier handoffs
(`2026-07-31b`, `2026-08-03`) are superseded by this one; keep them for the trail.

Chassis at time of writing: **v1.0.1247**.

---

## 0. The one-paragraph version

webdesign.uk is now a **properly seeded, framework-built site** and its build is
running, but it **cannot finish**, because the landing page build is hitting
`bugs_open/192`, a live fleet-wide defect filed this morning by another lane. The
domain still serves a holding redirect, so nothing broken is public. Getting to
"live VM with the chat box" is four pieces of work, and only the first is blocked:
finish the static site (blocked on 192), provision the box, build the chat with its
spend controls, then wire the two together at the Cloudflare edge.

---

## 1. Exactly where things stand

| thing | state | how to re-check |
|---|---|---|
| `sites` row + specs | **done**, 12 current aspects | `SELECT aspect, source_agent FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND is_current;` |
| research → strategy → briefing → site plan → composition | **all complete** | work-item query in §7 |
| imagery | several complete, some in flight | same |
| **landing page** | **FAILED — `bugs_open/192`** | same; `status='failed'` |
| `pages` rows | 1 | `SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='webdesign.uk';` |
| `sites.build_status` | `pending` | `SELECT build_status, last_built_at FROM sites WHERE domain='webdesign.uk';` |
| webdesign.uk DNS | apex + `www` A `192.0.2.1`, **proxied** | `dig +short A webdesign.uk @1.1.1.1` |
| webdesign.uk serving | **302 → webdesign.co.uk** (Page Rule `b8e08b35028315a274b2f5c7fea9154d`) | `curl -4 -sSI https://webdesign.uk/` |
| `*.ugg2.com` previews | **live and proven** with a real 200 | `curl -4 -sSI https://preview.ugg2.com/` |
| CF API token | **IP-blocked from this machine** | §6 |

**Nothing is publicly broken.** The holding redirect means the half-built site is
not visible. Do not remove the redirect until the site actually builds.

### The stale hand-built page still sitting in the bucket

`portfolio-sites/webdesign.uk/index.html` (13,254 B) is the **hand-written page that
should never have existed** (see §8). It is currently unreachable because the apex
points at `192.0.2.1` and the Page Rule intercepts. **The framework build will
overwrite it** at deploy. If you want it gone sooner:
`b2 rm b2://portfolio-sites/webdesign.uk/index.html`. Leave
`preview.ugg2.com/index.html` alone or delete it too; it is only a render witness.

---

## 2. THE BLOCKER: `bugs_open/192`

```
step process_sections_loop failed: failed to execute action loop: failed to get
collection at 'sections_for_render.sections_ready': key 'sections_ready' not found
at position 1 in path 'sections_for_render.sections_ready'
```

`page-content-writer`'s `select_sections` step, failing broadly since ~2026-08-03
21:00. **Not caused by this lane's seed** — it is hitting other sites and other item
types. I have added this lane's instance to the bug file as a third data point,
which removed two possible scopings (it is `needs_page` not `content_rewrite`, and a
brand-new site not an adopted one).

**Same error string as `bugs_open/087` but a different cause.** 087 is the
`page-rebuild` path, which supplies no section plan at all; this is the
`page-build-handler` path, which 087's own text names as the known-good control.
Do not treat 087 as the fix.

**Owned by another thread** (`75fceb501`, `ece9f8e66`). A `090` diagnosis run is
owed and had not been run at time of writing. **Contribute to the file, do not
compete.** Check before acting: `python3 scripts/who-owns.py 192`.

**Do not work around it.** Hand-building the page is exactly what the owner ruling
of 2026-08-04 forbids (`CLAUDE.md` → Platform conventions). This lane is genuinely
blocked, not merely inconvenienced, and that is the correct state to be in.

**When it is fixed**, re-drive the page rather than resubmitting the whole domain:
the upstream specs are all present and re-running `082` would redo work that
succeeded. Re-queue the failed `needs_page` item (its spec carries
`plan_id 4ecaa120-1fa6-4de1-9cd0-39d60c64b729`, `page_name index`,
`page_role landing`).

---

## 3. The architecture question you must settle FIRST

**The framework builds a STATIC site into B2. The chat box needs a live backend.
Those are two different delivery paths and they have to coexist.** Decide this
before provisioning anything, because it determines what the box actually serves.

> ### ⚠ CORRECTED later on 2026-08-04, after searching the code and the live DB
>
> **The recommendation below was written believing the framework only deploys
> static sites. That is wrong, and the correction changes the shape of the work.**
>
> **The framework already builds AND deploys VM-hosted sites, today, in
> production.** The mechanism is a per-site hop in the git deployer: a site whose
> `sites.github_repo` is **`vm-sites`** deploys to the box instead of the
> B2-backed default repo (`git_deployer_actions.go:95-101`, which names idea.uk as
> the example). Measured live: **36 sites → default → B2; 2 sites → `vm-sites` →
> a box** (`idea.uk`, `relojistas.com`). `relojistas.com` has 20 framework-built
> pages and `last_deployed_at` of **today**.
>
> There is also a real backend-site *class*, not just a deploy target:
> `deploy_config = {"target":"vm","capabilities":["backend"],"engine":{…}}`, and a
> discovery check (`check_backend_unreachable`) that probes `/health` on exactly
> those sites and NOOPs for static ones.
>
> **So `api.webdesign.uk` as a separate hostname is probably the WRONG shape.**
> The cheaper and better-precedented option is to make webdesign.uk a `vm-sites`
> site outright, exactly like relojistas.com: the framework builds the pages and
> deploys them to the box, nginx serves them, and the chat API is another location
> on the same host proxied to a local port. That is precisely idea.uk's existing
> layout (`box/proxy_tool.conf`, `box/proxy_stripe.conf`, both
> `proxy_pass http://127.0.0.1:8080`, with the Stripe webhook deliberately given
> its own no-rate-limit location because Stripe retries in bursts).
> **Same origin, so no CORS at all**, and no second hostname to certificate.
>
> **What is genuinely NOT built**, and is the real boundary: **the framework does
> not generate backend code.** The `site-engine` is a single hand-written Go
> service, the same binary on every box, exposing `/health`, `/stats`, `/events`,
> `/intent`, `/api/hit`. Register `DYN-001` tier 2 ("agent-powered per-site
> backends") is still `aspirational`. **The chat backend must therefore be written
> once, by hand, as a service** — either added to `site-engine` or run alongside
> it. Nothing generates it, and no agent will.
>
> Also not built, and relevant if you want the chat to arrive as a *component*:
> the `requires-backend` capability gate (`CTS-049`). Verified today: there is
> **no `semantic_tags` column on `site_components`**, and **0 active agent
> definitions** reference `requires-backend`. So the planner cannot be told a
> component needs a backend, and cannot be stopped from placing one on a static
> site. Treat the chat as page markup plus a hand-written service, not as a
> capability-gated component.
>
> **Read `SUMMARY_2026-08-04b_dynamic_site_capability.md` before designing this.**
> The register's `dynamic-applications.md` is frozen at **2026-07-13** and its
> DYN-001 line ("none built beyond tier 1 basics") is stale in exactly the way its
> own banner warns about.

**Superseded recommendation, kept for the trail: keep the site static and put ONLY the API on the box.**

```
webdesign.uk          → Cloudflare → Worker (portfolio-sites-router) → B2   [static, framework-built]
api.webdesign.uk      → Cloudflare → cloudflared tunnel → the VM            [chat + orders, dynamic]
```

Why this way round:

- **It keeps the site framework-built**, which is now a rule and is also the product
  demonstration. Serving the site off the VM would quietly recreate the hand-built
  problem in a different shape.
- **The static half stays free and un-attackable.** The thing facing strangers is
  the API only, so the spend controls have one door to guard rather than two.
- **A separate hostname avoids path-routing entirely.** Routing `/api/*` to a tunnel
  while `/*` goes to a Worker is fiddly and easy to get subtly wrong; two hostnames
  is boring and obvious.
- CORS is manageable: same registrable domain, so set
  `Access-Control-Allow-Origin: https://webdesign.uk` explicitly rather than `*`.

> **⚠ UNVERIFIED and load-bearing: I do not know how `webdesign.uk` reaches the
> Worker.** `GET /zones/{webdesign.uk}/workers/routes` returned **an empty list**,
> yet on 2026-07-31 the domain served the Worker's JSON 404 with
> `objectKey: "webdesign.uk/index.html"`. So the binding is *not* a zone Worker
> route. Most likely a **Workers Custom Domain** or an account-level route, neither
> of which that endpoint lists. **Settle this before designing the API routing** —
> if the binding is a Custom Domain on the apex it may also claim subdomains, which
> would change the `api.` plan. Check in the dashboard under Workers & Pages →
> `portfolio-sites-router` → Settings → Domains & Routes, or
> `GET /accounts/{acc}/workers/scripts/portfolio-sites-router/domains` once the
> token can reach account scope (it currently cannot; see §6).

---

## 4. The box

**Copy the tools-api island, not the idea.uk box.** The island is already a Mythic
Beasts VDS running the exact profile §6c wants: containers + Postgres, Caddy,
**`cloudflared` tunnel with no inbound**, nightly `pg_dump` + off-box rsync
(`vm_estate/PLAN_2026-07-25...:184`).

| | spec | why |
|---|---|---|
| CPU | **2 cores** | The chat is I/O-bound, waiting on the Anthropic API. Concurrency is capped by the spend ceiling long before CPU is. |
| RAM | **4 GB** | The one number not to trim. It all fits in 2 GB until you also want to build or deploy on the box. |
| Disk | **40–60 GB SSD** | Transcripts and orders are text and stay small. Container images are what consume it. |
| OS | **Ubuntu 24.04 LTS** | Matches the estate; keeps the `setup.sh` lineage usable. |
| IPv4 | **not required** | With a tunnel there is no inbound at all, and Mythic Beasts charge for IPv4. |
| Backups | nightly `pg_dump` + off-box copy | Copy the island's, verbatim. |

**Two properties matter more than the numbers.** No inbound ports at all, which
makes `CF-Connecting-IP` unforgeable rather than merely conventional and removes the
origin-firewall step idea.uk needed. And this box **faces strangers and spends money
on every visitor**, so §5.1's controls ship with it or it does not ship.

> **⚠ `bugs_open/139` fires the moment the tunnel is up.** Behind Cloudflare the true
> client address is in **`CF-Connecting-IP` only**. Get it wrong and the per-IP
> limiter silently becomes **one global bucket that still looks like it is working**.
> The discriminating check is `count(DISTINCT ip) > 1` **from two different
> networks**; one test machine cannot tell a constant from a working key (139:
> 83/83 identical rows). Reuse the estate's `cloudflare_realip` module rather than
> hand-rolling it, and prove it with the two-network check before going live.

---

## 5. The chat box, and the controls that are not optional

§5.1's ruling is a **real LLM chat**, not a stepped form. That resized P1: a fake
door with a form costs nothing to run, a fake door with a chat **spends money on
every visitor**. So the control table is part of P1, not P2 polish.

Build these **before** the chat is reachable, not after:

1. **Per-IP rate limit**, keyed on `CF-Connecting-IP` (see the 139 warning above).
2. **Turn cap per conversation** — a hard ceiling on messages, not a soft nudge.
3. **Per-day global spend ceiling** that fails closed. When it trips the chat must
   degrade to the contact details, not error.
4. **Request log** — every call, with tokens and cost, so the ceiling is auditable.
5. **Transcript-as-data** — the transcripts are the demand signal P1 exists to
   collect. Store them as structured rows, not as log lines.

**Model:** the intake chat is **not** the product, so per §7b it runs on the cheap
fast tier (`claude-haiku-4-5`), while paid *builds* run on Fable 5. Do not put Fable
on the anonymous chat.

**Before pointing anything at Fable**, P0's two pre-flight checks still apply:
verify org data retention is ≥30 days (a ZDR org gets a 400 on every Fable request),
and grep the chassis LLM call layer for `temperature`/`top_p`/`top_k`/
`budget_tokens`/`thinking`, which Fable rejects. A model swap is **not**
config-only if the call layer passes those.

**Stripe stays in TEST mode for P1.** Nobody pays. The price is displayed and the
flow is real, but the demand test is whether people go through with it, not whether
the money arrives. Live Stripe is P3, and §7a's terms must be written before it.

---

## 6. The Cloudflare token is IP-blocked from this machine

```
code 9109  Cannot use the access token from location: 5.65.164.9
```

Token `f0089a62ce6ea218b8c8137956d28297` has **Client IP Address Filtering** and this
machine's address is outside it. **Two traps recorded as landmines:**

- **`/user/tokens/verify` is exempt from the filter.** Measured the same minute:
  verify returned `active` over both families while a real zone call from that same
  IPv4 was refused. The health check answers a different question, so a green result
  proves nothing.
- **The same cause reports two ways.** One call returned `9109` naming the address,
  the next returned the generic `10000 Authentication error`. If you only see 10000
  you will go re-issue the token and fix nothing.

**Fix, in order of preference:** allow the ISP **CIDR** rather than one address (it
rotated within a day); pin the client to `curl -4` so only one address needs
listing; and long-term run Cloudflare changes **from the new box** and lock the
token to that fixed address. Removing the filter is the weakest option: this token
reaches **36 zones** with DNS write, so the filter is the blast-radius control.

The token also **cannot reach account scope**, which is why the Worker-binding
question in §3 is still open, and **cannot write Rulesets** (Page Rules work).

---

## 7. Order of work

1. **Watch `bugs_open/192`.** Nothing else on the static path can finish first.
   ```sql
   SELECT wi.item_type, wi.status, wi.handler_agent, LEFT(COALESCE(wi.error,''),80)
     FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
    WHERE s.domain='webdesign.uk' ORDER BY wi.created_at;
   ```
2. **Settle the Worker-binding question** (§3). It is one dashboard look and it
   unblocks the API design. Do this while waiting on 192; it costs nothing.
3. **When 192 clears:** re-drive the failed `needs_page`, let the cascade finish,
   then **read the built page against the copy rules** — specifically that no em
   dash survived and no "a person checks it" phrasing returned. If an em dash is
   present, the writer_block is not the thing to fix first; check that the
   `banned_claims` sweep actually runs over every slot (noted in the seed as
   `[UNVERIFIED]`).
4. **Go live:** delete Page Rule `b8e08b35028315a274b2f5c7fea9154d`, repoint the
   apex A from `192.0.2.1` to `199.59.243.228` keeping it **proxied**. Verify
   `HTTP/2 200 + text/html`, not a 302 and not a JSON 404.
5. **Provision the box** (§4), tunnel first, real-IP proven with the two-network
   check, then the app.
6. **Build the chat with its controls** (§5), controls first.
7. **Wire `api.webdesign.uk`** and replace the page's contact route with the real
   input box.

Steps 5–7 are owner-blocked on hosting credentials: there is **no `hcloud`, no
Mythic Beasts API access and no `cloudflared`** on this machine.

---

## 8. What went wrong on 2026-08-03, so it is not repeated

I hand-wrote the shopfront as a single HTML file and uploaded it straight to the
bucket. It rendered, I verified it thoroughly, and every check passed. It was still
wrong, for two reasons either of which is sufficient: **a hand-rolled shopfront
demonstrates nothing on a page whose product is framework-built websites**, and it
**silently opted out of every control** — no `evidence_base`, so the claims layer was
not lenient but *absent* (`loadEvidenceBase` returns nil and every lane no-ops), no
banned-claim sweep, no hallucinated-email check, no imagery guide.

**What caught it was the owner asking, not any check of mine.** A thorough
verification of the wrong artefact returns green. The question I did not ask was not
"is this correct" but "should this exist in this form".

Now an owner ruling in `CLAUDE.md` → Platform conventions, plus `WRONG_CALLS.md` and
a landmine footprinted on the bucket and the `b2` upload command.

---

## 9. Still owned by the owner

- **The correction fee.** §7d suggested £150/change or £600/day. The page states
  that post-acceptance changes are charged and our own mistakes are not, but quotes
  **no number**. Not blocking until someone asks.
- **The contact email.** `webdesign@contactforsales.com` is shared with
  webdesign.co.uk, a different product. It is real (which is why it was used;
  inventing one is the fabricated-contact defect class). **Change it in the seed
  before the build writes it onto the contact page** if enquiries need separating.
- **Hosting credentials** for the box (§7 steps 5–7).
- **Terms of business**, before Stripe goes live in P3, not before P1.

Settled and needing nothing further: the **price (£1,200)** and **VAT (none, owner
not registered)** — both now registered as attested facts in `evidence_base`, which
is what permits the writer to state them at all.
