# PRE-PLAN 2026-09-02 — an extensible real form endpoint for static sites

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

So this is not a boxingonline defect. **Every form the estate has ever built is a decoration.**
A visitor filling one in gets nothing, and no operator is told. The owner has separately ruled
that boxingonline's contact page is deleted and that contact pages become opt-in on explicit
request — which removes the immediate exposure but not the class.

## 2. What exists today `[MEASURED 2026-09-02]`

- **No form-handling machinery in the platform.** No submissions / enquiries / leads / messages
  table (`information_schema` search returned only `processed_messages`, an unrelated
  orchestration table, and `bak_rrform_tpl_20260726`, a dated backup I have not opened).
- **No form endpoint in the Go tree.** A search for `form_submission`, `FormSubmission`,
  `contact_submission`, `/api/contact`, `form-submit` across `platform/ internal/ cmd/ services/`
  returns one unrelated file.
- **Two hosting modes, and this is the crux:**
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
