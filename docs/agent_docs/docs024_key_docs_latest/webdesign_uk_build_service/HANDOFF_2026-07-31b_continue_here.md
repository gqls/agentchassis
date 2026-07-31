# HANDOFF — webdesign.uk, continue here: BUILD P1 (2026-07-31, later)

**COLD-START ENTRY POINT. Supersedes `HANDOFF_2026-07-31_continue_here.md`**,
which was written when P4 looked like the next task. It no longer is: P4 is
planned and its premise is verified, and **P1 is now the next thing to build.**
That earlier handoff is still correct about the council/register obligations
(its §4) and the lane's landmines (its §5); everything else here supersedes it.

Standing docs: `PLAN_2026-07-28_webdesign_uk_build_service.md` (the design and
every owner ruling) · `PLAN_2026-07-31_p4_order_intake.md` (P4, planned and
verified — **read its §7.2, it constrains P1**) · `NOTES_…` (evidence, every
misstep) · `RUNBOOK_…` (commands) · `README_where_we_are.md` (owner's log).

---

## 1. State

| phase | state |
|---|---|
| P0 | **DONE.** Fable-5 retention verified live (HTTP 200), unit cost floor $0.01479, SSRF guard built |
| **P1 — the shopfront** | **NEXT. NOT STARTED. This handoff's subject.** |
| P2 — free teaser | not started |
| P3 — manual fulfilment, real money | not started |
| P4 — automate the trigger | **PLANNED + PREMISE VERIFIED LIVE**, not built (needs P1's box to poll) |
| P5 — automate the seeding | not started; `direction`'s 5 shapes are most of it |

Shipped from this lane already, and reusable: **`platform/fetchguard`**
(register `DBI-025`) — the outbound/SSRF guard, live on `v1.0.1215`, which also
closed `bugs_closed/159`.

**Ownership: this lane is unowned.** Re-run the three checks before starting —
`git log` on the workstream dir, `scripts/who-owns.py`, and open
`site_work_items` — this file goes stale within hours on a shared tree.

## 2. What P1 is

**The shopfront with nothing behind it.** webdesign.uk on a VM: the minimal
page, the LLM chat that takes a domain, the optional questionnaire, Stripe in
**test mode**, orders stored on the box. **Nothing builds.** It answers the
only question that matters first — does anyone type a domain in and go through
with it?

This is deliberately the fake-door idea.uk already ran
(`idea.uk/idea_uk_fakedoor.html`), and the reason it comes first is written in
`PLAN_2026-07-28` §2a: **idea.uk is complete, verified, working, live — and has
sold nothing to a stranger.** Building the engine before the demand test is the
mistake this phase exists to avoid repeating.

## 3. The precedent — copy idea.uk's engine, do not invent one

`docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/` is a **working,
production, money-taking** version of almost exactly this shape. Measured:

| file | lines | what it gives P1 |
|---|---|---|
| `billing.go` | 162 | **Stripe, complete.** `Provider` interface, `CreateCheckout(orderID, email, priceGBP)`, HMAC-verified `ParseWebhook`, and a `FakeProvider` for end-to-end tests. Price passed **per call**, not held on the provider — they hit that bug and fixed it |
| `store.go` | 302 | Order model + persistence. Single JSON file, mutex-guarded, atomic rename on write. Also `ExpireStale`, `RecoverInterrupted`, and `Events map[string]bool` for **Stripe webhook idempotency** |
| `main.go` | 85 | Wiring, env config, `INTERNAL_API_KEY` |
| `page.html` | 46KB | One embedded page compiled into the binary. No framework, no build step |
| `setup.sh` | 15KB | The box: nginx + certbot + systemd + hardening |

It serves 17 routes including `/stripe/webhook`, `/order/success`,
`/order/cancel`, `/confirm`, `/health`, `/terms`, `/privacy`, `/refund-policy`.
**The whole payment half of P1 is a copying job.**

Its `Order` struct already carries `Domain` — which is webdesign.uk's entire
intake — plus `Email`, `Status`, `ProviderSessionID`, `PriceGBP`, `IP`,
`UserAgent`, and `PublishConsent`.

## 4. What P1 must ADD that idea.uk does not have

Three things. The first two are the interesting work; **the third is easy to
miss and will strand P4 if it is missed.**

**(a) A real LLM chat, not a form** (owner ruling, `PLAN_2026-07-28` §5.1).
idea.uk uses fixed forms. webdesign.uk's intake is a conversation that collects
the domain and, ideally, conducts the briefing. **This resizes P1**: a fake
door with a form costs nothing to run; one with a chat **spends money on every
visitor, including hostile ones**, so §5's controls ship *with* it, not after.

**(b) The questionnaire, optional** (owner ruling §5.3) — and note §5.3a's
corrected position: the platform already fail-*closes* on email
(`bugs_closed/063`), so the control P1 still owes is on **telephone, address
and accreditations**, which have no equivalent check.

**(c) A COLLECTION CONTRACT for P4 — this is the one to not forget.**
P4 polls this box for paid orders and must mark them collected. **idea.uk's
Order has no such concept** — grepped: no `collected`, `exported`, `delivered`
or `dispatched` field; its lifecycle is `awaiting_review | delivered | failed`,
which is about *fulfilment*, not about *the cluster having picked it up*. So P1
must add:

- a field on the order (e.g. `CollectedAt *time.Time`);
- an **authenticated** endpoint listing paid, not-yet-collected orders — follow
  idea.uk's existing internal-auth pattern, `INTERNAL_API_KEY` +
  `Authorization: Bearer` (`main.go:36`, used by `/internal/run` and `/op`);
- an acknowledge path so the cluster can mark them collected.

Build this in P1 even though nothing consumes it until P4. It is ~30 lines, and
retrofitting it later means touching a box that is by then taking real money.

## 5. The controls that ship WITH P1, not after

From `PLAN_2026-07-28` §5.1's table. **`platform/httpguard` already provides
most of this** — it is the platform's inbound-abuse package, and P1 is exactly
its intended consumer:

| need | reuse |
|---|---|
| per-visitor rate limit | `httpguard.NewLimiter(bands...)` — banded, returns retry-after |
| trustworthy client key | `httpguard.ClientIP(r, front)` |
| bot gate on the form | `httpguard.CheckIntake(honeypot, elapsedMillis, minFill)` |
| turn cap per chat session | new, trivial |
| **per-day global spend ceiling** | new — the only control that bounds total loss |
| request log from deploy #1 | `bugs_open/083`: the island had no denominator |

> ⚠️ **The landmine that makes or breaks the rate limiter.** `httpguard.ClientIP`
> takes a `FrontEnd` argument **and it is required for a reason**: behind
> Cloudflare + Caddy the real address is in **`CF-Connecting-IP` only**. Get this
> wrong and every visitor keys to one constant — measured on the island as
> `sha256("172.18.0.1")` in **83 of 83 rows**, a "per-IP" limiter that was one
> global bucket. Use `httpguard.CloudflareTunnel()` or `Nginx()` to match what
> actually fronts the box, and **verify with `count(DISTINCT key) > 1` from two
> different networks** — one test machine cannot tell a constant from a working
> key. Full evidence: `bugs_open/139`.

**Transcript-as-data, not prose-in-a-prompt.** The chat transcript flows into
the brief, which later flows into a build. A visitor typing *"ignore your
instructions and…"* is writing into a document agents will read. It must enter
as **quoted customer statements in a named field**, never spliced into a prompt.

## 6. Landmines for this phase

Grep `LANDMINES.md` by footprint before touching anything unfamiliar. The ones
that bite here:

- **`CF-Connecting-IP`** — §5 above. The single most likely silent failure.
- **`platform/httpguard` is inbound-only**; `platform/fetchguard` is the
  outbound/SSRF one. P1 is inbound (httpguard); P2's teaser fetching a
  customer's domain is outbound (**`fetchguard.NewClient`, never a bare
  `&http.Client{}`**).
- **`DEPLOY_USER=deploy` is not optional** — omitting it kills every `vm-sites`
  deploy (relojistas lane).
- **A chassis roll makes a scheduled task look broken for ~5 minutes** — only
  relevant once P4 exists, but check pod AGE before diagnosing any dispatch
  silence.

## 7. Owner rulings — settled, do not re-ask

Real LLM chat (not a stepped form) · questionnaire **optional**, re-open only
at P5 · trust boundary **pull-only from day one**, isolation method decided at
P3 · full sites at a high quality-based price · guarantee and fee model both
hinge on **acceptance** (before = refund only, preview comes down; after =
customer's changes paid, **our defects free**) · builds on `claude-fable-5` ·
**P4 poll interval 15 min** · **repeat domains = reject and alert a human,
never drop** · the "about a thousand sites" figure is accepted as
forward-looking.

## 8. Still needed from the owner

1. **The preview domain** — a different, shorter domain, still to be supplied.
   Not blocking P1 (it is P3 that needs it), but P1's copy may want to name it.
2. **The price.** Unblocked — P0 measured the cost floor. Needed before Stripe
   goes anywhere near live mode (P3), not for P1's test mode.
3. **DNS for webdesign.uk** is still not pointed.

## 9. First moves

1. Re-run the three ownership checks (§1).
2. Read `PLAN_2026-07-28` §5.1 (chat + controls), §7a (the offer), §9 (P1's
   scope) and `PLAN_2026-07-31_p4_order_intake.md` §7.2 (the collection
   contract P1 owes P4).
3. Copy `billing.go` + `store.go` first — the proven, boring half — before
   writing any chat code. It is the part that must not be got wrong, and it is
   already written.
