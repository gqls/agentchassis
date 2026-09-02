# REVIEW 2026-09-02 — the form-endpoint pre-plan's D1, against the publish seam

**Reviewing:** `../static_site_form_endpoint/PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md`
(boxingonline session's, at the owner's instruction). **Reviewer:** `site_delivery_and_editor`,
as the owner of the publish seam D1(c) proposes to put application logic in. This is the review
that pre-plan's §5 records as offered-but-not-happened. Scope: **D1 only**, plus the two
interactions (D3, D5) that touch the seam. All claims measured/read 2026-09-02 ~20:0xZ; sources
cited inline. This review decides nothing — it arms the owner's D1 choice with what the seam
actually is.

## 0. Verdict in three sentences

**D1(c) as written conflates two components, and the one that could actually receive a POST is
the worst place in the estate to put application logic** — four grounds below, each cited at
code or at tonight's own commits. The seam has no objection to (b), and can contribute to it.
Whatever wins, the seam's build side will stamp `form_action` values and publish thank-you
pages — that half is cheap, in-scope, and decided by D3/D5, not D1.

## 1. First, a naming correction: "the publish worker" is two components

- `platform/publish/b2worker.go` — the **Go publish backend** (mirror). Batch: copies the built
  tree to `<hostname>/<key>` prefixes in one bucket, ETag-verifies, and (since `b60d66e3c`,
  tonight) CONVERGES by deleting destination keys whose source is gone. It is **not in the
  request path** — it cannot receive a form POST at all.
- `scripts/cloudflare/worker.js` — the **serving edge**: a single ~207-line Cloudflare worker
  (`[MEASURED 2026-09-02]` wc -l) that resolves `object key = hostname + path` for **every**
  b2-hosted site through **one deployment**, holding only B2 read credentials.

D1(c)'s receiver can only be the second. The pre-plan should name it, because every cost below
attaches to the serving worker specifically. (Small table correction while here: the
`publish_target = b2worker` row's "can it execute server code? **no**" is about *site-authored*
code — the edge itself already executes logic: www→apex 301, directory-index rewrite,
`/worker-health`, a `?debug=1` endpoint, and the robots.txt rewrite that 429's probe had to
exclude. That distinction is exactly why (c) is temptingly close and structurally wrong.)

## 2. The four grounds against D1(c)

**2a. No safe storage exists at the edge — and tonight's sweep makes the obvious one
destructive.** The worker's only credentialed store is the serving bucket. Submissions written
there under a site's prefix are now **deleted by design**: `b60d66e3c` makes Publish converge
(destination keys with no source are swept — its own LANDMINES entry: anything hand-placed
under a `*.ugg2.com` prefix is swept on next drift). So (c) needs new infrastructure anyway
(KV/queue/D1 + notification secrets in the worker) — which erases "zero new infrastructure and
zero new credentials", the exact argument `b2worker.go`'s own header gives for this seam
existing. (c)'s headline advantage does not survive contact with its storage question.

**2b. Blast radius inverts the failure the pre-plan is guarding against.** The pre-plan
(rightly) marks (a) down because "box down → customer enquiries silently lost". (c) is
strictly worse: one worker serves every static site, so a form-handling bug, a crash on a
malformed POST, or a D4-class abuse event degrades **serving fleet-wide** — sites down, not
messages lost. Under (b), the receiver failing loses submissions and nothing else. The
request path's most valuable property today is that it holds no application state and ~nothing
that can fail per-site.

**2c. It makes the seam's acceptance evidence theory-laden.** As of tonight, publish acceptance
is a served **404/200 pair before `published_hash` writes** (`b60d66e3c`), and this lane's whole
verification practice (handoff §3: read the artefact first) rests on a served status meaning
exactly one thing: object present or object absent. The edge's existing behaviours already tax
this — the 429 probe had to carve out robots.txt because the edge rewrites it to 200, and a
parked domain 200s every path (both in LANDMINES). Every piece of application logic added to
the worker adds carve-outs to every future probe. A POST path that can answer 200 turns "the
page serves" into "the page serves, unless that 200 was the form app" — a per-probe tax on the
seam's verifiability, paid for ever, by every lane.

**2d. It moves application logic outside the estate's own review gate.**
`scripts/cloudflare/worker.js` is **not in council scope** (`COUNCIL_SCOPE_CODE_RE`,
council-scope.sh:129 — `platform|internal|pkg`, two cmd/ trees, pattern-check.py), deploys by
hand outside the chassis roll, and is invisible to the 098 coverage report. A public,
unauthenticated, data-accepting endpoint — the thing D4 says is "the part most likely to be
under-scoped" — would live precisely where no council seat reviews it and no roll carries it.
(b) puts the same logic where the gate, the roll, the provenance stamp and the acceptance
machinery already are.

**The honest form of (c):** a *separate* worker on its own route (`forms.<domain>/…` or an
`/api/forms` route bound to a different script) answers 2b and most of 2c. But that is a new
service at the edge with its own storage and secrets — i.e. option (b) relocated to Cloudflare,
keeping 2a and 2d. If per-request cost ever matters enough to pay those, it should be chosen
as "(b) at the edge", not as "the publish worker already sits there".

## 3. What the seam offers whichever receiver wins

- **D3, build side:** stamping a real per-site `form_action` at build is this seam's normal
  work (the 22 `#contact` components are `content_data` values the pipeline already writes).
  One caution from the identity practice (proxy-chain lesson): at any shared receiver, **derive
  the site from something the visitor cannot choose** — a per-site token or the receiving
  route/hostname — never from the POST's Origin header alone. That choice belongs to D3 and
  survives any D1 outcome.
- **D5:** a static thank-you page is just a page — the framework builds it, the mirror publishes
  it, any receiver 303s to it. No D1 coupling; D5 should not be allowed to drag D1 toward (c).
- **D6:** if D3 lands on the convention form, retrofit is a no-op at the seam (nothing per-site
  to rewrite); if on per-site URLs, it is a fleet content migration this lane would want staged
  behind the same acceptance probes as any rerender.

## 4. For the pre-plan's §4.1 (does the box path generalise?)

Not this seam's code to judge, but one property is worth carrying whatever D1 decides: the
shopfront's **poll-collector** shape (receiver stores; `order-intake-collector` pulls) means the
cluster exposes no inbound surface for submissions and receipt survives cluster downtime. A (b)
receiver that stores-and-is-polled keeps that property while shedding (a)'s single-box coupling.

## 5. Standing notes

Review recorded here rather than edited into the pre-plan (their doc; §5's "not yet reviewed"
line is theirs to update). D2 untouched — correctly deferred to 420's identity replumb.
Findings owed back to the boxingonline session and the handoff §1.6 entry; this file is the
citation.
