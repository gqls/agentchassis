# idea.uk — open discussion (2026-05-28 checkpoint)

We've paused to think before pushing further. Captured here so the discussion is
recoverable in a new chat. Five questions on the table. Each has my honest
working answer; final decisions are still yours.

---

## 1. What does one idea run cost me?

Pricing verified (May 2026) on the Anthropic site:

- Opus 4.7: **$5 / $25** per million tokens (input/output)
- Sonnet 4.6: **$3 / $15**
- Haiku 4.5: **$1 / $5**

The pipeline makes 5 LLM calls per run. Rough token estimates per step:

| Step | Model | Input | Output | Step cost (USD) |
|---|---|---|---|---|
| 1. Audience | Opus 4.7 | ~700 | ~500 | $0.016 |
| 2. Generate (12–24 candidates) | Opus 4.7 | ~2,000 | ~2,500 | $0.072 |
| 3. Cut | Sonnet 4.6 or GPT-4o | ~3,000 | ~1,500 | $0.025–0.032 |
| 4. Verify (with web search, up to 6 searches) | Opus 4.7 | ~20,000–25,000 | ~2,000–2,500 | ~$0.18–0.25 + ~$0.06 search calls |
| 5. Score | Sonnet 4.6 | ~3,500 | ~1,000 | $0.026 |

**Total per run: roughly $0.40, call it £0.30–£0.60 (working estimate £0.50).**

The verify step dominates because the web_search tool's results count as input
tokens fed back into Opus. The other steps are small change in comparison.

Easy ways to cut this:
- Move scoring to Haiku 4.5 (5× cheaper than Sonnet) — quality risk is low for
  numerical scoring against a clear rubric.
- Use prompt caching (90% off cached input) on the prompts that stay the same
  across runs — the capability menu and system prompt are identical every time.
  Could cut ~£0.05–0.10 per run.
- Cap web_search at fewer uses if verify is well-grounded after 3–4 calls.

A realistic *optimised* cost is probably £0.20–0.30 per run with these levers
applied. **Your real cost will appear in the Anthropic console after each run** —
the dashboard breaks down tokens spent per request, so you can confirm against
your actual usage rather than my estimates.

---

## 2. Stripe break-even — how much to charge

UK Stripe fees: **1.5% + £0.20** for UK/EEA cards, **3.25% + £0.20** international.
**Stripe does not return the processing fee on refunds** — you lose both the
percentage and the fixed fee. (Confirmed in current Stripe docs and several 2026
fee guides.)

For a UK-card charge of £Y you receive Y × 0.985 − £0.20:

| Charge | Stripe takes | You net | Profit after £0.50 engine cost |
|---|---|---|---|
| £5  | £0.275 | £4.725  | £4.22 |
| £9  | £0.335 | £8.67   | £8.17 |
| £29 | £0.635 | £28.37  | £27.87 |
| £49 | £0.935 | £48.07  | £47.57 |
| £99 | £1.685 | £97.32  | £96.82 |
| £199 | £3.185 | £195.82 | £195.32 |

The break-even price (just covering Stripe + engine) is about **£0.72**. Anything
above that has positive contribution margin. But the meaningful number is
**refund cost**: if you refund a £49 order, you lose the £0.93 Stripe fee +
£0.50 engine cost = **~£1.43 per refund**. That's the worst case to budget for.

**My recommendation for the early-access phase:**
- Start at **£29–£49 per report** (not £199 — too steep to gather feedback fast).
- Keep the strong refund guarantee already on the page.
- At a 20–30% refund rate (worst case for new offers), per-order economics still
  work: £49 charge − £1.43 worst-case refund cost = healthy net.
- Move price up once you've seen 5–10 unrefunded orders and have a feel for
  reliability.

---

## 3. The "real door" — does the client get a result at the moment of paying?

**Not at the moment of paying, currently.** The flow today is:

```
pay → webhook → engine runs (2–10 min) → AUTO_DELIVER off → operator reviews
→ operator sends → customer gets the report by email
```

So they pay, see "your report is being prepared, within 72 hours," and the
report arrives by email later. The 72h promise is honest and matches the manual
review phase.

To make it feel like a "real door" with results at the point of payment, three
honest options:

**Option (a) — Streaming progress page after payment (best UX).** After
Stripe redirects to `/order/success`, that page polls a `/status/{order_id}`
endpoint and shows progress in real time: "generating candidates… 12 ideas
found… cutting against free alternatives… 8 remain… verifying claim 1 of 8…
done." The report appears in the browser when ready, with an email copy. Takes
5–10 minutes on the page, but the customer sees real work happening. **This is
the version I'd build if you want it to feel live.** Modest extension: add a
status field and a polling endpoint to the service, render progress on a
post-payment page.

**Option (b) — Run before pay, release after pay.** Engine runs during the
"confirm" step (before charging) so the report is ready when they pay; payment
just unlocks delivery. Pros: instant delivery on payment. Cons: 20–30% of pay
links go unpaid; you'd spend £0.50/run on orders that never convert.

**Option (c) — Keep the 72h email model.** Pros: simplest, what's built, fits
the operator-review safety gate. Cons: less "real door" feeling.

My lean: **(a)** is the right answer if you want the real-door feel. Build it
when you've got real demand to justify it. **(c)** is honest for now; the page
already promises 72h, not instant.

---

## 4. How to actually charge them and refund them

### Charging (one-off setup, maybe 30 minutes)

1. **Sign up at stripe.com.** UK businesses need: company / sole trader details,
   UK bank account, ID, address. Usually verified within hours.
2. **Get test keys** (`sk_test_…`, `whsec_test_…`) — use these while developing.
3. **In Stripe dashboard → Developers → Webhooks:** add an endpoint at
   `https://idea.uk/stripe/webhook`, subscribed to
   `checkout.session.completed`. Stripe gives you a signing secret.
4. **Set env vars on the service:** `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`.
5. **Test using Stripe's test card numbers** (e.g. `4242 4242 4242 4242`).
   Walk the full request → confirm → pay → fulfil flow once.
6. **Switch to live keys** (`sk_live_…`). Update the webhook endpoint's signing
   secret to the live version.
7. **Payouts** land in your bank account daily or weekly (configurable).

### Refunding (two ways)

**Manual (zero code needed):** in the Stripe dashboard, find the payment, click
Refund. One click. Full or partial. Stripe keeps the fee.

**Programmatic (small addition to the service):** add an `/refund` endpoint
(operator-gated, like `/confirm`) that calls `POST /v1/refunds` on Stripe's API
with the payment_intent ID we already store. About 30 lines of Go. Useful if you
expect any volume — saves dashboard time. **Want me to add this when we resume?**

### Cost of a refund

Stripe doesn't return the fee. On a £49 refund:
- Customer gets £49 back.
- Stripe keeps the £0.93 fee from the original charge.
- Plus your engine cost on the report (£0.30–0.60).
- **Total cost of a refund: ~£1.23–1.53.**

This is small enough that a generous refund policy is affordable, but big enough
that you shouldn't run a £5 product with frequent refunds.

---

## 5. Voluntary pay / "two free goes" — honest analysis

My read: **probably not a good idea in this form.** Here's why:

**Voluntary pay ("pay if satisfied").** Industry data: voluntary pay converts at
roughly 1–10% for B2B-style products, even when the value delivered is real. At
£0.50/run cost and 5% pay-£29, your average revenue per run is £1.45, leaving
~£0.95 margin. Workable. But it has bad failure modes: it doesn't filter
serious users, attracts abuse (one IP runs 100 reports = £50 of engine cost),
and gives you no demand signal to set price by — you can't tell if "no one
paid" means "no one valued it" or just "no one bothered."

**Two free goes per customer.** Easy to circumvent (new email each time). Per-IP
blocking hits mobile users and corporate networks. You'd be giving away £1
of engine cost per email address, and the people who'd use multiple emails are
exactly the ones who weren't going to pay anyway.

**The better pattern that gets you the same benefits:**

**Free "audience challenge" taster + paid full report.** Step 1 of the method
(challenge the stated audience, surface 2–3 alternative audiences) is **cheap
(~£0.02 per run)** and often the single most valuable line in the whole report
— the v2 test runs showed it changed the verdict on multiple domains. Offer
this as a free, instant sample on idea.uk. Visitors enter their domain and
audience, click submit, and see (in ~10 seconds) "here's how a structured
challenge looks at your audience" with 2–3 alternative framings and why.

Then: *"Pay £29 for the full report — 3–6 ranked candidates, web-verified, with
the cheapest test for each. Refund if nothing's worth acting on."*

This gives you what voluntary pay was supposed to give (proof of value before
purchase) without the abuse risk and with a clear demand signal. Engine cost
for free tasters: about £20 per 1,000 tasters. Easily covered by even a 1%
conversion rate to the full £29 report.

**My recommendation: free audience-challenge taster + £29 paid full report with
refund guarantee.** Drop voluntary pay and the multi-free-go idea. The taster is
the better hook.

---

## 6. Self-hosted LLMs — the realities

Short version: **don't self-host for idea.uk right now.** At your likely volume
(early access, maybe single-digit reports per week, possibly dozens per week if
it works), commercial models are dramatically cheaper than self-hosting. Here's
the actual maths.

### What the method actually needs

Multi-step reasoning, reliable structured JSON output, web search (built-in or
self-orchestrated), and — critically for the cut step — *ruthless judgement*
that doesn't agree with itself.

### Commercial frontier vs open-weight reality

| | Commercial frontier (Opus 4.7, GPT-4o) | Llama 3.3 70B / Qwen 2.5 72B |
|---|---|---|
| Reasoning quality | Strong | Roughly mid-commercial-tier (think Sonnet, not Opus) |
| JSON reliability | Strong | More format errors, more retries needed |
| Web search | Built into API (Anthropic web_search) | You build the agent loop + buy search API ($5–50/1k queries) |
| Cut step ruthlessness | Sharper | Tendency to agree with the generator (the very failure mode the cut exists to prevent) |
| Per-run cost (low volume) | ~£0.50 | ~£0.30–0.50 on rented GPU + your ops time |
| Per-run cost (high volume) | ~£0.50 | Can fall to £0.05–0.15 if you own GPUs and amortise |

### Hardware reality if you wanted to own it

- Llama 3.3 70B at FP16 needs ~140GB VRAM. That's 2× H100 80GB (~£40k+) or 4× A100.
- Q4 quantised (still decent): ~40GB VRAM. 1× H100 80GB (~£25k) or 1× used A6000 48GB (~£5k).
- You then need to operate it: vLLM or similar, monitoring, restarts, model upgrades.

### Cloud GPU rental

- Lambda Labs H100 ondemand: ~$1.99/hr. At 10 min/report: ~$0.33 per run on GPU.
- Plus the search API: ~$0.05–0.10 per run for verification.
- Plus your ops time setting it up and keeping it working.
- **Roughly the same cost as commercial APIs at this volume, with worse quality
  on the cut step.**

### When self-hosting starts to make sense

- Volume of hundreds of runs per day.
- You have GPU operations capability already in the team.
- Privacy requirement that needs on-prem inference.
- Or: there's a meaningful "frontier-class" open model that closes the gap (the
  field is moving — worth watching, but Llama 3.3 70B isn't it for this method's
  cut step).

### What I'd actually do for cost optimisation

1. Move scoring to Haiku 4.5 (5× cheaper than Sonnet, same task quality).
2. Turn on Anthropic prompt caching for the system prompt + capability menu.
3. Keep generation, cut, and verify on the frontier models — they're doing
   the work that matters.

That probably gets you to **£0.20–£0.30 per run** without giving up quality.
Self-hosting is a 2027 decision, not a 2026 one.

---

## 7. Linking from other sites — white-label, branding, and pricing

Three architectural options:

### Option A — Iframe embed
`<iframe src="https://idea.uk/embed?tenant=othersite">` on othersite.com.
- Pros: simplest; URL stays on othersite domain.
- Cons: visible iframe boundary; cookie/auth complications; awkward on mobile;
  any links inside lead back to idea.uk.

### Option B — Subdomain reverse-proxy / CNAME
`ideas.othersite.com` → CNAME or HTTP-proxies to the central service.
- Pros: cleanly branded URL throughout; Stripe Checkout's merchant name can
  still say "Othersite Reports."
- Cons: per-tenant DNS + TLS cert handling (manageable with Cloudflare or
  Let's Encrypt SAN certs); service must detect tenant from `Host` header.

### Option C — Branded request page per tenant + central engine (RECOMMENDED)
Each tenant site has its own static "request a report" page (built through your
normal pipeline: own copy, own branding, own price). The form POSTs to
`https://idea.uk/request` with a `tenant_id`. After payment, redirect back to
the tenant's "thanks" page on their domain.

- Pros:
  - Branded experience on the tenant's domain.
  - One central engine = no per-site infrastructure to maintain.
  - Per-tenant pricing (different audiences can pay different amounts).
  - Per-tenant tracking (which tenant brought this customer).
  - Cheap to add a new tenant (just deploy a new static page through the
    normal pipeline).
  - Stripe Checkout supports merchant branding (logo, brand colour, name) so
    the payment page can look on-brand to the customer.
- Cons:
  - Stripe Checkout URL is `checkout.stripe.com` (customer leaves the
    tenant's domain to pay) — unavoidable unless you build embedded checkout.

**Option C is the right one for what you described:** "offer them something,
charge not too much, get feedback, don't lose money."

### What's needed in the service to support this

Small additions:
- A `tenant_id` field on Order (or use `domain` as the key — it's already there).
- A small `tenants` config: per-tenant name, success/cancel URLs, price, allowed
  origins.
- Per-tenant Stripe branding (set once in the Stripe dashboard, or use Stripe
  Connect if each tenant should have their own account).
- A tenant-aware `/request` endpoint that records who sent the order.

This is maybe 100–200 lines of Go, plus a tiny config file. **Want me to do this
in the next round?**

### Suggested early-access offer per tenant

- Free audience-challenge taster (section 5).
- £29 full report, refund guarantee.
- Same engine, branded landing per tenant.
- You operate manually with `AUTO_DELIVER=false` for the first batch on any
  tenant until you trust the output for that audience.

That gives you the test of "do strangers pay for this" across multiple audiences
at low risk, with each tenant's landing page costing essentially nothing to spin
up through your existing build pipeline.

---

## Open decisions to make next session

In rough priority order:

1. **Price** for the early-access phase: £29? £49? Or free-taster + £29 full?
2. **Real-door** option (current 72h email, or build the streaming progress page).
3. **Refund endpoint** added to the service (programmatic), or just use the Stripe
   dashboard (manual)?
4. **Multi-tenant** (Option C above): build the small additions, or keep idea.uk
   as the only domain for now?
5. **Cost-optimisation pass:** move scoring to Haiku 4.5, turn on prompt caching?
   Cheap wins, probably worth doing before any real volume.
6. **Self-hosted LLMs:** parked until volume justifies it (suggest revisiting if
   you're seeing >100 reports/week).
7. **Stripe account setup** — when you're ready to take real (test-mode first)
   money.

Tell me which of these you want to dig into when we resume.
