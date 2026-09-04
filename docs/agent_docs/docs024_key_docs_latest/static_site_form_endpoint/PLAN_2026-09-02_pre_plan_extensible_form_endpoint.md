# PRE-PLAN 2026-09-02 — an extensible real form endpoint for static sites

> **STATUS 2026-09-04 — EXTENDED AND DECIDED. The decision space below is answered in
> `PLAN_2026-09-04_form_endpoint_build.md`; read that for what is being built.** This file stays
> as history and as the record of what we believed on 2026-09-02. **Five of its premises are
> corrected inline** (§1, §2 twice, D1's hosting table, D2) — struck through and dated, never
> edited away, because what it was wrong about is *what already exists*, which is the expensive
> mistake on this estate. Evidence for every correction: `NOTES_static_site_form_endpoint.md` §1.

**Status: PRE-PLAN. Nothing is built, nothing is decided, no code or config was touched.**
Written at the owner's instruction (2026-09-02): *"please create a pre-plan that I can extend in
another thread to build an extensible real form endpoint for static sites. We have dynamic sites
like idea.uk that submit to real backends."*

Its job is to **map the decision space and name the trade-offs** so the owner can extend it
elsewhere — not to choose. Every choice below is left open on purpose; where I have a view I say
so and mark it as a view. All figures `[MEASURED 2026-09-02]` against the live DB and tree.

---

## 1. Why this exists — the case that produced it

boxingonline.com (first paid customer build) shipped a contact page whose form had
`form_action = "#contact"` — it submits nowhere. `[MEASURED]` **22 components fleet-wide carry
that same value, and it is the ONLY value any of them has**:

```sql
SELECT content_data->>'form_action', count(*) FROM page_components
 WHERE content_data ? 'form_action' GROUP BY 1;
--  #contact | 22
```

So this is not a boxingonline defect. ~~**Every form the estate has ever built is a decoration.**~~
A visitor filling one in gets nothing, and no operator is told. The owner has separately ruled
that boxingonline's contact page is deleted and that contact pages become opt-in on explicit
request — which removes the immediate exposure but not the class.

> **CORRECTED 2026-09-04 — the census was taken one layer above the defect, and could not have
> come out any other way.** `content_data` is the value the content LLM *wrote*; the render seam
> rewrites it on the way out. `deliverableFormAction`
> (`platform/orchestration/actions/component_library.go:1495`) replaces a non-delivering action
> with `mailto:<sites.email>` — the pattern the owner chose on 2026-07-17 — and
> `check_contact_form_undeliverable` files a human-review item for the sites where no address
> resolves. Both predate this pre-plan by six weeks.
>
> Re-measured at **both** layers on 2026-09-04 (RUNBOOK §1): `content_data` now holds `#contact`
> ×**27** (it grew by 5 overnight — a form census goes stale by *addition*), but of those,
> **21 serve a real `mailto:`** and only **6 components on 6 sites** still serve `#contact` —
> exactly the address-less sites (`boxingonline.com`, `cv1.co.uk`, `farmerinsurance.uk`,
> `garden-tools.uk`, `relojistas.com`, plus the `pool-ai-agents.internal` pool row).
>
> So the class is real but far smaller, and the honest statement is not "every form is a
> decoration" but **"the estate's forms deliver by `mailto:`, and `mailto:` cannot carry a
> structured payload, cannot route to a recipient that changes without a rebuild, and cannot be
> measured."** That is still worth building for; it is a different argument.
>
> **What caught it:** re-running the pre-plan's own query against `rendered_html` as well as
> `content_data`, before quoting the figure. Logged in `WRONG_CALLS.md` — the marker rule was
> followed in full (dated, `[MEASURED]`, query written out) and the measurement still could not
> have been disconfirmed.

## 2. What exists today `[MEASURED 2026-09-02]`

- ~~**No form-handling machinery in the platform.**~~ No submissions / enquiries / leads / messages
  table (`information_schema` search returned only `processed_messages`, an unrelated
  orchestration table, and `bak_rrform_tpl_20260726`, a dated backup I have not opened).
- ~~**No form endpoint in the Go tree.** A search for `form_submission`, `FormSubmission`,
  `contact_submission`, `/api/contact`, `form-submit` across `platform/ internal/ cmd/ services/`
  returns one unrelated file.~~

> **CORRECTED 2026-09-04 — the search vocabulary matched nothing that exists, and three of the
> four things this pre-plan says must be built are already in the build.** The table claim stands;
> the two Go claims do not. The named search terms are the only reason: none of the real
> mechanisms is called anything like them.
>
> | this pre-plan's question | what is actually in the tree |
> |---|---|
> | **D4**, "the part most likely to be under-scoped… the reason a form endpoint is not a small job" | **`platform/httpguard`** (`3632874d4`, 2026-07-28, `features_open/024` A3) — `CheckIntake(honeypotValue, elapsedMillis, minFill)`, i.e. *a form intake gate*; `NewLimiter(bands…)/Allow(key)`; `ClientIP(r, front)` with `Nginx()`/`CloudflareTunnel()`/`Direct()` |
> | **D2**, notification | **`platform/mailer`** (`1d747f5e8`, 2026-07-28, A2) — "the platform's ONE way to send email", whose header names its third intended caller: *"idea.uk's paid report today, the gripper dossier next, **contact forms after that**"* |
> | **D1(b)**, "a new public surface to defend" | **not new.** `https://tools.apis.uk/api/v1/tools/*` is live: Cloudflare tunnel → island Caddy (that prefix only, 1 MB body cap, `gauntlet_dead_cta/infra/island/Caddyfile`) → `tools-api`, which **already imports both packages above**. Probed with controls 2026-09-04: `OPTIONS …/gauntlet/round` **204**, `OPTIONS …/gripper/session` **204**, unregistered path **404**, `/health` (outside the prefix) **404** |
> | the nearest existing thing is the box order path | the nearest existing thing is **`internal/tools-api/handlers/gripper.go:227` `GripperSubmitHandler`** — a live, browser-facing form intake that runs the bot gates first, returns an indistinguishable rejection, hashes the IP and logs structurally only |
>
> **What caught it:** grepping `OPEN_THREADS_RESTART_LIST.md` for the workstream rather than the
> symbol, which surfaced the consolidation programme's line *"build `platform/mailer` (item A2),
> then `platform/httpguard` (A3)"* — and both had landed. A symbol search cannot find a package
> whose name you have not guessed; the index of what the estate is building can.
- ~~**Two hosting modes, and this is the crux:**~~ **CORRECTED 2026-09-04 — it is not the crux,
  and the table covers 5 of 39 live sites.** `[MEASURED 2026-09-04]` 34 live sites have
  `github_repo` **and** `publish_target` NULL, because `publish_target` is itself an opt-in seam
  with the default OFF (`publish_site_action.go:120`: *"publish_target not set for &lt;domain&gt;
  (seam is opt-in, default OFF)"*). Only `idea.uk`, `webdesign.uk`, `relojistas.com` and
  `noted.co.uk` are `vm-sites`; only `boxingonline.com` and `noted.co.uk` are `b2worker`.
  More to the point, **the distinction does not bear on the receiver at all**: a static page can
  POST cross-origin to a real backend regardless of how it is hosted, and two estate sites already
  do — `robot-hands.com` and `vonc.com` both carry published markup that posts to
  `tools.apis.uk`. Hosting mode decides whether a backend can live *beside* the site; it says
  nothing about whether the site can reach one.
  | mode | example | can it execute server code? |
  |---|---|---|
  | `github_repo = vm-sites` | idea.uk, webdesign.uk | **yes** — VM-hosted, a backend can live beside the site |
  | `publish_target = b2worker` | boxingonline.com | **no** — static objects behind a worker |
- **One working form→backend path already exists and is NOT general:** the webdesign.uk shopfront
  takes real paid orders, collected from a box endpoint (`/internal/orders`, bearer token
  `WEBDESIGN_BOX_ORDERS_TOKEN`) and polled by `order-intake-collector`. It works, it is in
  production, and it was built for one purpose. **Whether it generalises is the first question
  worth answering** — reuse before building is the platform convention, and this is the nearest
  existing thing.

## 3. The decision space

### D1 — where does the receiver live? (the load-bearing choice)
- **(a) The box.** Reuses the shopfront's proven shape and its token pattern; one more endpoint
  beside `/internal/orders`. Couples every customer site's forms to one host we operate.
- **(b) The cluster.** A service behind the existing ingress, reachable from anywhere.
  Consistent with the rest of the estate; a new public surface to defend.
- **(c) The publish worker.** ~~The b2worker already sits in front of every static site. A form
  POST never leaves the edge. Cheapest per request; puts application logic in the publishing seam,
  which is a boundary this estate has otherwise kept clean.~~
  > **CORRECTED 2026-09-02 by the publish-seam owner's review (see §5). My (c) named a component
  > that cannot receive a POST.** "The publish worker" is TWO things: the Go **mirror**
  > (`platform/publish/b2worker.go`) is a BATCH job, not in the request path, and can never be a
  > receiver; the **serving edge** (`scripts/cloudflare/worker.js`) is the only candidate, and it
  > is ONE ~207-line deployment serving every static site. Any future version of (c) must name
  > the edge worker specifically.
- **(d) Per-site, VM-hosted only.** Do nothing for static sites; make a real backend the reason a
  site is VM-hosted. Honest and cheap, but it means static sites can never take a message, which
  contradicts "extensible".

~~*My view, offered as a view:* (b) or (c).~~ **(c) IS OPPOSED BY THE SEAM OWNER, on four cited
grounds, and I accept the verdict — the argument I made for it does not survive its own premise:**

1. **It has no safe store.** The edge's only credentialed storage is the serving bucket — and
   since `b60d66e3c` the mirror **converges**, sweeping anything under a site prefix with no
   source. Edge-stored submissions would delete themselves. So (c) needs new infrastructure and
   secrets anyway, **which erases the "cheapest, nothing new" argument that was its whole appeal.**
2. **It would poison every served-status probe, fleet-wide, for ever.** Acceptance is now a served
   404/200 pair checked before `published_hash` is written. A POST path that can return 200 makes
   every such probe theory-laden and adds carve-outs to all of them (`robots.txt` is the
   precedent).
3. **It would put D4's riskiest surface where no gate can see it.** `worker.js` is outside council
   scope (`council-scope.sh:129`) and deploys by hand, off-roll — so the abuse surface I named as
   "most likely to be under-scoped" would live in the one place nothing reviews.
4. **(b) is supported.** And the honest form of (c) — a separate worker/route with its own
   storage — is really **"(b) at the edge"** and should be costed as such rather than as a free
   ride on existing infrastructure.

**(a) remains weak for the reason I gave** — it makes a customer-facing capability depend on a
host built for our own shopfront, and "the box is down, enquiries silently lost" is a failure this
estate has been bitten by. **So the live choice is (b), (d), or (b)-at-the-edge honestly costed.**

### D2 — what happens to a submission?
Storage (a table, per-site scoped), notification (email to an operator or the site owner), or
both. **Note the identity ruling of the same day**: *"The identity of the person creating the
site can be independent of the operation of the site (they might be the design agency) so the
contact details need to be independent of the site build."* **So "email the site owner" is not
yet answerable** — the estate does not currently model who that is. This pre-plan therefore
**depends on the identity replumb** (`bugs_open/420`'s class fix), and should not choose a
notification target before that lands.

> **DISCHARGED 2026-09-04 — the dependency is real but already satisfied, and by a route this
> paragraph did not anticipate.** `bugs_open/420`'s class fix (committed `162877051`) separated
> the two contracts that had shared one column: the payer's address is **delivery-only** and stays
> in `build_queue.direction.customer_email`, while `sites.email` is now **only** the published
> contact, written solely from an explicit opt-in `direction.published_contact` key. The owner's
> class ruling behind it: *"When a customer pays but is never explicitly asked what contact details
> the site should show, the site publishes NONE. The payer's address is used only to deliver."*
>
> Neither of those is a form recipient. So this lane does not wait on 420 — it adds the **third**
> identity that 420's split makes room for, in its own table (`site_form_routes.recipient_email`,
> per site **and per intent**), where changing it is a config update rather than a rebuild. That
> is also what makes the copyonline CONTRIB's requirement satisfiable at all.

### D3 — how does a site declare it has a form?
A `form_action` that means something (per-site endpoint URL written at build), versus a
convention (every site posts to one endpoint and is identified by origin/site id). The second is
simpler and makes the 22 dead components fixable in one pass; the first is more explicit.

### D4 — abuse.
A public unauthenticated POST endpoint attached to every site we host. Rate limiting, a spam
control, size caps, and a decision about whether submissions are ever shown in the admin console
(if so, they are untrusted user input rendered in our own UI). **This is the part most likely to
be under-scoped**, and it is the reason a form endpoint is not a small job.

### D5 — what does the visitor see?
A static page cannot easily show a server-rendered result. Options: a thank-you page redirect,
an inline JS success state, or the endpoint returning a redirect. Interacts with D1(c).

### D6 — retrofit or forward-only?
22 existing components carry `#contact`. Fixing them all at once is a fleet migration; leaving
them is 22 live decorations. Either is defensible; the choice should be explicit rather than
implicit.

## 4. What I would settle first, in order

1. **Does the shopfront's order-intake path generalise?** One read of that code answers D1(a)
   and may collapse half of this.
2. **D1**, because everything else follows from where the receiver lives.
3. **The identity replumb**, because D2 cannot be answered before it.
4. **D4**, before any endpoint is exposed, not after.

## 5. What this pre-plan deliberately does NOT do

Choose an option, estimate effort, or propose a schema. The owner asked for something he can
extend in another thread; committing to an implementation here would foreclose exactly the
choices he wants to make. Nor does it touch the 22 existing components — that is D6, unsettled.

**REVIEWED 2026-09-02 by the publish-seam owner** — `site_delivery_and_editor`, who own the seam
D1(c) proposed to use:
`docs024_key_docs_latest/site_delivery_and_editor/REVIEW_2026-09-02_form_endpoint_preplan_D1_vs_publish_seam.md`
(commit `df56c1cf0`). **Verdict: (c) opposed on four cited grounds, (b) supported, D2's deferral
to 420 confirmed correct.** Folded into D1 above with my original view struck through rather than
edited away — the reasoning I got wrong is the part worth keeping, since it was wrong about *what
the component is*, not about what would be desirable.

**Two further review points, carried into the decisions they affect:**

- **D3 — at any shared receiver, derive the site's identity from a TOKEN or the receiving ROUTE,
  never from the `Origin` header.** An attacker sets Origin. This is the estate's proxy-chain
  lesson applied to a surface that does not exist yet, which is the cheapest possible moment to
  apply it.
- **D5 must not drag D1 toward (c).** A thank-you page is just a published page, so the
  visitor-feedback decision creates no pressure to put the receiver at the edge. Worth stating
  because "the response has to come from somewhere" is exactly the argument that would otherwise
  smuggle (c) back in.

**Still unreviewed:** D4 (abuse) by anyone with a security seat, and the whole plan by the owner,
who asked for it as something he would extend in another thread rather than approve here.
