# RESUME HANDOFF — idea.uk VM site (start a fresh chat here)

> ## ▶ START HERE — state as of 2026-07-28 ~20:00 UTC (supersedes everything below)
>
> # The product is finished enough. The open question is demand, and it is not an engineering one.
>
> **Owner's direction, 2026-07-28: "maybe we start thinking of how to get a buyer on the site" — IN
> ANOTHER THREAD.** See *The next thread* at the foot of this block. Do not start it here.
>
> ### What idea.uk is now
>
> Nine guides in journey order, four tools, a £29 paid report, a **£8 example place**, a **published
> specimen report**, and a funnel where every link goes where its text says. Box `116.203.204.115`,
> systemd `idea`, orders in `/var/lib/idea/orders.json` (a FILE, no DB), capacity 1/5.
>
> **75 orders. Genuine external buyers: still ZERO.** One order on 28 July looked external and was a
> test — I asserted "our first customer" from an unfamiliar IP and user-agent and was wrong (see
> `WRONG_CALLS.md`, and §X.27's struck-through heading). Treat any claim of demand with suspicion
> until money has moved from someone who is not us.
>
> ### Shipped 2026-07-28 — ten binary deploys, eleven DB changes, all verified live
>
> | what | how it was proven |
> |---|---|
> | Slot-leak recovery (`running` orders survive a restart) | stranded a spare order, watched it recover |
> | **Unconditional cost logging** | a call with `created=0 read=0` — previously invisible |
> | Three-paragraph report intro | present in a real report |
> | Score bars in the HTML report | nested tables; SVG and remote images do not render in mail |
> | Styled error pages + `maxlength` | induced: 52 bytes of `text/plain` → 3,081 of styled HTML |
> | Subheadings through the assessment | both renderers; the text one had the same wall |
> | Hero + CTA → the form (was `/contact.html`) | negative control: the misdirect is gone |
> | Six cards → the specimen (were self-links) | 0 still self-linking |
> | "Couldn't I just ask an AI myself?" | live at 66% down |
> | **The £8 example place** | **Stripe asked directly: `amount_total` 800 pence, not 2900** |
>
> **COST IS NOW KNOWN AND SELF-MEASURING: £1.23 per report** ($1.23; rises to ~$1.32 after
> **2026-08-31** when Sonnet 5's introductory rate ends). ~3% of £29. Every future order records its
> own cost — `journalctl -u idea | grep '\[usage\]'`.
>
> ### Verify the box in one go
>
> ```bash
> grep -ac "refused fake=1"       /opt/idea/idea   # 089            → 1
> grep -ac "X-Real-IP"            /opt/idea/idea   # 090            → 1
> grep -ac "claude-opus-5"        /opt/idea/idea   # 5-family       → 1
> grep -ac "\[usage\] %s"         /opt/idea/idea   # cost logging   → 1
> grep -ac "EXAMPLE PLACE"        /opt/idea/idea   # the £8 tier    → 1
> grep -ac "margin:18px 0 5px"    /opt/idea/idea   # subheadings    → 1
> ```
> Rollbacks kept, newest last: `idea.prev-2026-07-28-pre-usage-log` · `-pre-intro` ·
> `-pre-scorebars` · `-pre-subheads` · `-pre-tier`. Orders backed up before every deploy.
>
> ### The £8 example place — LIVE, and how it is wired
>
> A full report in exchange for permission to publish it anonymously. `EXAMPLE_PRICE_GBP=8`,
> `EXAMPLE_MAX_PLACES=10` in `/etc/idea/idea.env`. **Places used: 0 of 10.**
>
> - **Two checkboxes, both unticked, wording beside them.** They are two decisions and the Go reads
>   them as two: asking without agreeing gets £29; agreeing without asking records no consent.
> - **Price moved from the PROCESS to the ORDER.** It used to live on `StripeProvider` and be read
>   from config at checkout time, so a price change mid-order would have billed a figure never
>   quoted. `sendPayLink` now reads it **once** for both the charge and the email.
> - **The cap counts declined and expired orders**: it bounds promises made to people, and declining
>   to run a report does not un-ask.
> - **Switching it on has an ORDER**: binary (tier off) → env → form. Any other sequence shows an
>   offer the backend ignores and charges £29 to someone who ticked £8.
>
> ### Open, with my read on priority
>
> 1. **Demand.** The only question that building cannot answer. → the next thread.
> 2. **The doubled-full-stop fix is UNEXERCISED after four runs.** It is in the binary and covered by
>    tests; it needs a submission whose text *ends in a full stop*. Do not report it as proven.
> 3. **Refresh the specimen** once current formatting has run on a real order (owner: "we can refresh
>    it at a later date").
> 4. **`row()` in the idea cards** still glues label to value the way `arow()` did before 28 July.
> 5. Two SES DNS records are **DONE** — `bounce.leopardess.uk` MX + TXT both resolve. The earlier
>    SERVFAIL was a broken sub-zone, not propagation.
>
> ### Landmines earned on 2026-07-28 — read before touching the chassis side
>
> - **RUNBOOK TRAP 1b: a missing REQUIRED field escalates the page to the LLM writer, and the job
>   still reports `COMPLETED`.** The tell is an ABSENCE: `rerendered`/`carried` are missing entirely
>   (versus the `slot_name` trap, which gives both as zero) plus `escalated:true`. Checking
>   "rendered == section_count" reads the NULL as a zero and sends you after the wrong cause.
>   `generic-text-block` requires **`heading`** as well as `content`.
> - **That escalation queues an LLM writer over your authored copy.** On the specimen page — which
>   publishes the claim that nothing was reworded — it would have made a live provenance statement
>   false, silently. Cancel the `needs_page` item before it is claimed.
> - **A dead CTA and a WRONG CTA look identical to a link checker.** `/contact.html` returns 200. The
>   hero button sent buyers away from the purchase for weeks and nothing ever complained.
> - **`pages.sections` is not written by the rerender path.** A page with `sections='[]'` serves
>   perfectly and is invisible to `ListedPageEligibilitySQL` and the imagery sweep.
> - **A bare-word grep is not an attribute check.** `checked` matched twice in prose and briefly
>   looked like two pre-ticked consent boxes.
> - **Verify each page change again after the NEXT one ships** — a later rerender is exactly when an
>   earlier `content_data` fix quietly disappears.
> - A fresh chassis roll (28 July) changed nothing here: the tool is a **standalone Go module** with
>   its own build→scp→systemctl path and no chassis coupling. All seven live checks re-passed after it.
>
> ### The next thread — getting a buyer
>
> **Brief:** idea.uk is complete, correct, and has never had a customer who is not its owner. Find
> out whether anyone wants it.
>
> What that thread should know before it starts:
>
> - **The measured baseline.** 18–28 July, bots filtered: **26 views of `/report.html` from 20 unique
>   IPs, 8 form submissions.** Reverse-DNS: most were `googleusercontent.com`, Tencent ranges and a
>   Tor exit; **4 of the 26 views were the owner and 4 of the 8 submissions were our own tests.**
>   Genuine prospects: between nought and a small handful.
> - **Do not compute a conversion rate from that.** 8/26 = 31% is meaningless and must never be
>   quoted — both numerator and denominator are dominated by us.
> - **The funnel data is free and already there.** No Cloudflare (refuted — Hetzner NS, A record
>   straight to the box). `/var/log/nginx/access.log` spans 5 June → now, unrotated.
> - **Two price points now exist** — £29 (private) and £8 (published example). The £8 is a demand
>   experiment as much as a revenue one.
> - **Email is provably ours now** (SPF + DKIM aligned), so mailing strangers is safe.
> - The site's whole credibility rests on not overstating. Any acquisition copy has to survive the
>   same bar — see `bugs_open/043` and the `no figure in any brief` rail.
>
> ---
>
> ## ▶ PREVIOUS STATE — 2026-07-27 11:13 UTC (superseded by the block above)
>
> # 🎉 idea.uk HAS SOLD AND DELIVERED ITS FIRST REPORT.
>
> **2026-07-27 11:13:13 UTC — `ord_1785090638951163875` is `delivered`.** The owner paid the £29,
> Stripe's signed webhook landed, the report was emailed, and the capacity slot released itself.
> **Every link in the chain has now run in production**: request → operator confirm → engine →
> draft → approve → pay link → **real card payment → webhook → delivery → slot freed**. Until this
> moment the business end of the product had never executed; that sentence is now retired.
>
> Verified from the box and the access log, not from the fact that someone said so:
>
> ```
> order   : status=delivered  updated=2026-07-27T11:13:13Z  session=cs_live_a1j7o8uz…
> capacity: {"active":0,"max":5,"open":true}      ← slot released automatically
> journal : email to aaa@designconsultancy.co.uk sent: "Your idea.uk report"
> nginx   : POST /stripe/webhook  200  from 3.130.192.231  "Stripe/1.0 (+…/docs/webhooks)"
>           GET  /order/success?o=ord_1785090638951163875  ← referred from checkout.stripe.com,
>                                                            and note: NO `fake=1`
> ```
>
> **That last line is worth keeping.** The genuine Stripe redirect does not use the `fake=1`
> shortcut, so removing it (`bugs_closed/089`) demonstrably did not break the real payment path —
> the fear you would reasonably have when deleting something the checkout appears to touch. The
> same access log shows my three `&fake=1` attack probes returning 200 without ever progressing the
> order. Attack refused and real payment accepted, in one file.
>
> **Caveat on what was actually delivered:** the report was generated at 18:40 on 26 July, so it
> predates both the three copy fixes (deployed 21:10) and the move to the Claude 5 family
> (deployed ~21:55). The customer's copy therefore still contains the doubled full stop and the
> "out of 5 —" line — verified, not assumed. **The next order is the first report on the new models
> and the first with the copy fixes**, and it is also the first that will measure per-report cost.
>
> **The site is complete, the box is healthy, the report format is proven, and every queued deploy
> has shipped.** Four deploys went out on 26 July: the automatic order expiry that this file
> used to call the top job, plus two security fixes found while preparing it, plus three copy
> fixes found by reading a real report, plus the move of the report engine onto the Claude 5
> family (`claude-opus-5` / `claude-sonnet-5` — see Open item 2, which is now done).
>
> ~~**THE ONE THING OUTSTANDING IS A HUMAN ACTION**~~ — **DONE 2026-07-27 11:13, see the top of this
> block.** Kept here rather than deleted because the shape of the gap is the useful part: the 12:40
> run on 26 July proved the report *format* and was then declined, which left
> `approve → pay → webhook → delivery` unexecuted even though everything looked finished. A product
> can be complete, verified and demonstrably working and still have never once done the thing it
> exists to do. **The next site to be declared done gets asked this question first.**
>
> ### The two security defects — both were live, both are fixed, deployed and proven
>
> Neither had been exploited. Both were found by reading the code to plan the expiry deploy, not by
> any alarm — nothing in the platform would ever have reported either.
>
> | | |
> |---|---|
> | **`bugs_closed/089`** | **The £29 report could be taken without paying.** `/order/success` honoured `?fake=1` on a query parameter alone, marking an order paid and delivering. The shortcut exists for `FakeProvider` and the type is commented *"NEVER in production"* — the handler simply never checked which provider was configured, and the box runs Stripe. Reachable by an ordinary buyer: `CreateCheckout` gives Stripe a `cancel_url` carrying the order id, so cancelling a real checkout discloses it. |
> | **`bugs_closed/090`** | **A visitor could choose the IP it was rate-limited as.** `clientIP` took the FIRST `X-Forwarded-For` entry; our nginx uses `$proxy_add_x_forwarded_for`, which *appends* the real peer — so entry 1 is the part the caller writes. That key drives the free taster's 3/hour limit, which is the only bound on LLM spend at an unauthenticated endpoint that costs ~£0.02 a call. |
>
> Both were **induced live, not argued**: 089 against a genuine `awaiting_payment` order with a
> real `cs_live_` session (held; pre-fix it reached `delivered`), and 090 by sending
> `X-Forwarded-For: 203.0.113.77` — logged verbatim as that address before the fix, and as the real
> IPv6 peer after it.
>
> ### Verify what is on the box — per-fix markers, never "a deploy happened"
>
> ```bash
> grep -ac "refused fake=1"   /opt/idea/idea   # 089  → 1
> grep -ac "X-Real-IP"        /opt/idea/idea   # 090  → 1
> grep -ac "(each out of 5)"  /opt/idea/idea   # copy → 2
> grep -ac "YOUR IDEA, ASSESSED" /opt/idea/idea # 07-25 engine work → 1
> grep -ac "claude-opus-5"    /opt/idea/idea   # 5-family models → 1
> grep -ac "claude-opus-4-8"  /opt/idea/idea   # the models it replaced → 0
> grep -ac "recovered after a restart" /opt/idea/idea  # 07-27 slot-leak recovery → 1
> ```
>
> **Why each fix carries its own marker, and this is the lesson of the evening:** the 18:29 deploy
> could not have contained 090 — that defect was *found at 18:32, while verifying 089 on the box* —
> so a second deploy was needed at 18:44 and a third at 21:10 for the copy fixes. Only re-running
> the probe after the first deploy revealed that 090 was still live. **The deploy that closes a bug
> is not the deploy that happened to precede it.** All four markers verified present 21:12 UTC, and
> both attacks re-run against the final binary and refused.
>
> Rollbacks kept, newest last: `idea.prev-2026-07-25` · `idea.prev-2026-07-26-089only` ·
> `idea.prev-2026-07-26-089-090` · `idea.prev-2026-07-26-opus48` (the last pre-5-family binary).
> Orders backed up before each deploy (`orders.json.bak-2026-07-26-predeploy`, `…2`, `…3`, `…4`).
> 73 orders intact throughout; the pending order kept its status, its 10,109-char report and its
> Stripe session across every restart.
>
> ### What the run told us that the format proof could not
>
> - **9.5 minutes**, not the 20–30 this file predicted (18:31:13 → 18:40:46).
> - **The report is good, and honest where it costs it**: submitted a vet price-comparison idea, and
>   it told the submitter a free government comparison service is coming, that a dozen rivals exist,
>   and to spend £50–£100 testing one town before writing any software. Real, checkable sources.
> - **Three copy defects, visible only by reading the artefact** — a doubled full stop, a score line
>   reading "out of 5 —" with no number (literal text in *both* renderers), and a sentence-cased
>   field spliced mid-sentence ("using A form the receptionist fills in"). Fixed, deployed 21:10.
>   Checking that the job succeeded would never have found any of them.
>
> ### Open, and deliberately not done
>
> 1. ~~**`ExpireStale` skips `running`.**~~ **DONE 2026-07-27 12:48 — fixed, deployed (5th deploy) and
>    INDUCED LIVE** (`5c3081e3f`, NOTES §X.24). Built once the box was idle, which was the
>    precondition. `Store.RecoverInterrupted` runs at startup — the only place the question is
>    decidable, since a process cannot inherit another's goroutines, so every `running` order there
>    is by definition abandoned. `ExpireStale` still refuses to touch `running`, deliberately: on an
>    hourly ticker it cannot tell a live 20-minute run from a dead one.
>    **Two corrections to what this item used to propose** — it said "reset any `running` order",
>    and that would have cost money. (a) A **paid** buyer must NOT go back to `requested`: under
>    charge-first, payment precedes the engine, so `/confirm` would issue a second pay link and
>    charge them twice. (b) Resetting to just any status fixes nothing — `ExpireStale` skips
>    `requested` and `paid` as well, so the target must either free the slot or be genuinely re-run.
>    `ProviderSessionID` (written only by `sendPayLink`) discriminates exactly: unbilled →
>    `requested`, slot freed, re-startable from the operator's existing `/confirm` link; paid →
>    `paid` and re-run, slot correctly retained.
>    Proof is the induction, not the deploy: a spam row was stranded `running` on the live box and
>    came back `requested` with the operator emailed, leaving the order distribution identical
>    (73; 60/5/4/4). The **paid branch is [UNPROVEN LIVE]** — under `review_before_pay=true` it is
>    unreachable in production, so it is tested only.
> 2. ~~**Margin lever, owner's call:** the engine still runs `claude-opus-4-8` / `claude-sonnet-4-6`~~
>    **DONE 2026-07-26 — the engine is on `claude-opus-5` / `claude-sonnet-5`, deployed and
>    verified** (NOTES §X.22). **And a correction to what this line used to say:** it called the four
>    model env vars "a lever needing no rebuild". They were not. The Go picked the thinking wire
>    format from an allow-list of adaptive-thinking models, so any model it had not heard of got the
>    legacy `budget_tokens` format — which the 5 family rejects. Setting `GEN_MODEL=claude-opus-5` on
>    the box alone would have returned **400 on every call** and taken the paid product down. Proven
>    against the live API before changing anything:
>
>    ```
>    claude-opus-5   OLD format (budget_tokens)   : 400 "thinking.type.enabled" is not supported
>    claude-opus-5   NEW format (adaptive+effort) : 200 stop=end_turn
>    ```
>
>    The selector is now a deny-list, so an unrecognised (newer) model gets the modern format and a
>    future upgrade cannot fail this way. **Cost is [UNMEASURED]** — Opus 5 is the same price as
>    Opus 4.8 and Sonnet 5 is at or below Sonnet 4.6 (intro rate to 2026-08-31), but Sonnet 5's
>    tokenizer counts ~30% more tokens for the same text and two steps now do a little thinking that
>    did none before. The next real order measures it: the `[cache]` log lines carry token counts.
> 3. **Refuted, do not re-litigate:** the standing "real-client-IP in nginx — idea.uk is behind
>    Cloudflare" item. It is **not**: Hetzner nameservers, A record straight to `116.203.204.115`,
>    no `cf-ray`. No `set_real_ip_from` is needed; the defect was in the Go, and is fixed.
>
> ### A coordination note worth keeping
>
> This session worked for two hours from the **15:34** version of this file after another session
> rewrote its "▶ START HERE" at **15:57** to record that the format run had already happened. The
> result: a question to the owner premised on a dead state, and a duplicate production run at ~2×
> the old per-report spend. It landed on the one untested leg by luck, not judgement.
> `ls -la` on this directory, or `git log --oneline -5`, costs two seconds and would have caught it.
> **Do that immediately before any expensive or outward-facing action, not once at session start** —
> and note the perverse incentive: the more valuable a "next action" looks in a handoff, the likelier
> another session is already doing it. Written up in `WRONG_CALLS.md` and NOTES §X.20.

### What idea.uk is, today (all curl-verified 2026-07-26)

| | |
|---|---|
| **Guides** | 9, live, in journey order on a self-populating hub: creating-ideas → building-it → testing-it → user-acceptance → feedback-loops → patents → copyright → funding-ways → funding-sources |
| **Tools** | 4 cards on `/tools.html`: the £29 Verified Idea Report · "Should you patent it?" (free) · "Which funding route fits?" (free) · Free Audience Check |
| **Paid tool** | Extended and DEPLOYED: the report leads with an assessment of the submitted idea, carries "Check it yourself" source links, discloses AI use in the report itself, and renders an honest "too early to assess" outcome. **Binary superseded 07-26 21:10** — now also carries auto-expiry, `bugs_closed/089`, `bugs_closed/090` and three copy fixes (markers in ▶ START HERE) |
| **Box** | `116.203.204.115`, service `idea` restarted **07-27 12:48** (5th deploy — the slot-leak recovery), queue **open, 0/5** (the first sale delivered and released its slot), 73 orders, `OPERATOR_EMAIL=idea-uk@leopardess.uk`, no `CONTACT_EMAIL` line (correct — see §Email) |
| **Locks** | 27 authored sections locked; both hub listings deliberately unlocked so they keep deriving |

Site id `1244516d-014d-421c-88c6-090bb1e9552a`. SQL applied this arc: `sql/p4_01`…`p4_19`.

### ✅ DONE 2026-07-26 — the new report format is PROVEN in production

Owner submitted a real idea, confirmed it, received the draft, judged it good, and declined it
(declining is the right way to close a test without self-charging; it releases the slot and emails
the requester politely at no cost).

Verified from the stored order, not from impression — `ord_1785069609860726188`, 13,227 chars of
text / 20,207 of HTML, every new-format marker present and the old-format giveaway absent:

| Marker | |
|---|---|
| `YOUR IDEA, ASSESSED` / `Your idea, assessed` (text + HTML) | ✅ |
| `Check it yourself` source lists | ✅ — 16 links in the HTML |
| `A considered next step` | ✅ |
| `We use AI to research…` (disclosure in the report, not just the T&Cs) | ✅ |
| `FURTHER IDEAS WORTH PURSUING` (ideation half, retitled) | ✅ |
| Old intro `"You asked us to find AI product ideas for…"` | ✅ **absent** |

`too early to assess` did not appear — correct, the submitted idea was assessable. That branch is
covered by tests (`service_test.go`), not by this run; a deliberately vague submission would
exercise it if you ever want the live proof.

**So the whole chain works**: request → operator confirm → step-0 assessment with live web search
→ cross-vendor cut → verify → score → draft to operator → decision. Two things worth watching on
the next few real runs: wall-clock (two long search passes now) and spend (~2× per report).

### TOP JOB NOW — a deploy is written, tested and waiting

Committed, `go build`/`go vet` clean, full suite green — **inert until the owner builds and
deploys** (the tool has no CI). Contents:

- **Automatic order expiry** — `Store.ExpireStale` + `App.sweepStale()` at startup, hourly, and
  before `/capacity` answers. `STALE_REVIEW_DAYS` (14) / `STALE_PAYMENT_DAYS` (7); `0` disables.
  New terminal status **`expired`**, distinct from `declined`, retaining the row.
- **`OPERATOR_EMAIL` as single source of truth** — `reportContact()` no longer reads an env var
  directly with a hardcoded fallback; it is wired from config at `NewApp`.
- Tests: `expire_stale_test.go` (3 cases: releases the right ones only, CreatedAt fallback for
  legacy rows, disabled thresholds are a no-op).

```bash
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
GOOS=linux GOARCH=amd64 go build -o idea .
scp idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'systemctl stop idea; mv /opt/idea/idea.new /opt/idea/idea; systemctl start idea; sleep 2; systemctl is-active idea; curl -s http://127.0.0.1:8080/capacity'
```

### THE TRAPS — read these before touching anything

1. **NEVER edit `/var/lib/idea/orders.json` under a running service.** It is read ONCE at startup
   and rewritten wholesale from memory on every order change. An edit while running is invisible
   to the process AND gets clobbered by the next request. Cost two failed attempts on 07-26.
   Always: `systemctl stop idea` → edit → `systemctl start idea` → `curl /capacity`.
   Corollary: **`systemctl start` on an already-active unit is a no-op** — use `restart`, or
   stop-then-start. This is what made the second attempt look like the first had failed.
2. **Never lock a section whose component schema has a `query.*` source.** Locks are row-granular;
   `SavePageSectionsAction` re-attaches locked rows verbatim, so locking a derived listing freezes
   it while every render still reports success. Nearly killed the self-populating guides hub on
   07-25. Every lock script since carries a guard that refuses. **This now bites for real**:
   `bugs_open/058` (the lock gate) went CLOSED & LIVE on **v1.0.1165** today — locks are enforced,
   so editing a locked page means unlocking first.
3. **`page_components.slot_name` must equal `content_components.function`** — the renderer keys
   its component lookup on `slot_name`, not `component_id`. NULL ⇒ every section is "carried"
   ⇒ nothing renders ⇒ **the job still reports COMPLETED**. See RUNBOOK Phase 5.
4. **`pages.sections` is not written by the rerender path** — backfill it, or the page is invisible
   to `ListedPageEligibilitySQL` and the imagery sweep.
5. **"No orchestration row" means QUEUED, not dropped.** Check whether *anyone's* orchestrations
   are starting before re-firing. Latency ranged from <1 min to ~12 min this week.
6. **Verify against the live page by curl, never the job status or the DB.** VM sitesync is a
   5-minute timer on top of the render. Every "verified" claim in these docs names the curl.
7. **A schema field with `source: static` AND a `fallback` is UNOVERRIDABLE** — the fallback is
   written into resolved_data unconditionally and resolved_data merges last. Hit twice
   (`guide-list`, `tool-list`); both fixed by dropping the fallback after verifying no-op.

### Email — settled, do not re-litigate

`OPERATOR_EMAIL=idea-uk@leopardess.uk` is **correct for the tool**. The site and the tool
deliberately use different addresses; an earlier claim that leopardess was "stale" was a site fact
wrongly widened to the tool (logged in `WRONG_CALLS.md`). The `CONTACT_EMAIL` line has been
removed and that is right: today's binary falls back to the correct hardcoded address, and the
queued deploy makes it resolve from `OPERATOR_EMAIL` properly.

### Open decisions for the owner

- **Margin**: each report now makes 6 model calls including two long web-search passes — roughly
  double the previous spend, at the same £29. Worth revisiting after a few real runs.
- **Ceiling**: `MAX_ACTIVE_ORDERS=5` is a deliberate throttle on spend and operator attention.
  Once the funnel produces real volume, is 5 still right?
- **Report copy** now *undersells* slightly (it mentions both halves but the further-ideas half
  briefly). Safe direction; revisit only if conversion suggests it.

### Optional housekeeping, no urgency

- ~60 spam `requested` rows from 2026-06-11 (the "elBd" injection strings) still sit in
  orders.json. They hold no slots. The /request hardening (honeypot, timing, limiter) went live
  with the 07-25 deploy, so the flood should not recur.
- Two SES custom MAIL-FROM DNS records outstanding since 07-18 (deliverability, not blocking).

### Where the record lives

- `RUNNING_NOTES_idea_uk_vm_site.md` §X.12–§X.18 — the technical log for this arc, including every
  misstep and correction.
- `README_where_we_are.md` — the plain-prose owner log.
- `SUMMARY_2026-07-25` → `07-26` → `07-26b` → `07-27` → **`SUMMARY_2026-07-27b` (current)** — how
  the understanding moved. Read the newest for state; read the series to see how wrong we were and
  when. 26b is the instructive one: it closed on "the format is unproven", which was true for about
  fourteen hours. **27b marks the inflection**: the engineering queue emptied and the binding
  constraint moved from correctness to demand.
- `EVIDENCE_2026-07-27_ai_unit_economics.md` — the canonical cost figures, what is and is not
  assertable, and the rail for the two copy threads that were given a pointer to it. **$0.641 is a
  FLOOR across two of five model calls; the complete figure arrives on the next real order.**
- `AUDIT_2026-07-25_paid_tool_vs_copy.md` — what the paid tool does vs what the page claimed, the
  owner's ruling, and what was built.
- `RUNBOOK_idea_uk_vm_site.md` **Phase 5** — the repeatable recipe for adding a guide or tool.
- `features_open/014` — the pipeline vision and its build log; `015` — the fleet-wide ladder.

### If you are adding the next guide or tool

Follow RUNBOOK Phase 5 verbatim; it is now proven ten times. The hubs list new pages
automatically (`query.pages_where_type:guide|tool`) — no hub edit needed, just a re-render after
the page ships. Content policy from `014`: stages 6–9 (patents/copyright/funding) stay
hand-authored until claims-verification V5 is live; stages 1–5 may take generated copy. Any new
tool where one answer can be decisive must **gate before it scores** — see both existing finders.


> ## ▶ PREVIOUS STATE (2026-07-22) — superseded by the block above, kept for history
>
> **The migration is DONE and LIVE, and `/bugs_open/018` (the broken chrome) is FIXED AND VERIFIED
> LIVE.** idea.uk now serves the chassis static site with full navigable chrome and a working free
> tool, on one origin. Nothing on the live site is currently broken by a defect this workstream owns.
>
> **What was fixed since the last handoff (all live, all verified against the deployed page — not the
> work-item status):**
> 1. **`018` — chrome links all `href=""` → 0.** Root cause was NOT the theory 018 guessed: the chrome
>    renderer (`render_site_components_action.go`) fills templates from a HARDCODED value map and never
>    reads `input_schema`; idea.uk's two per-site components declared other field names, so every one
>    resolved empty, and the templates had no `{{if}}` gates so each empty became a visible dead link.
>    Fixed by `sql/p3_01` (rewrote both templates against the real vocabulary, gated every anchor, set
>    `sites.logo_url`) + `sql/p3_02` (promoted the stuck rerender item). **NOT fleet-wide — idea.uk was
>    the only affected site.** Verified: `curl https://idea.uk/ | grep -oE '<a href="[^"]*"'` → 0 empty.
> 2. **The free taster ("no chrome, just text" + "POST only").** `/audience-check` is an AJAX fragment
>    endpoint by design; the form was seeded as a native POST with no JS, so the browser navigated to
>    the bare fragment. Same defect `p2_02` fixed for the report form, never applied here. Fixed by
>    `sql/p3_03` (JS interceptor + result div; corrected the pointer-page URL that fed the POST-only
>    cards) + `sql/p3_04` (forced a SECTION rerender — a plain `rerender-pages` cannot apply a template
>    edit; see the landmine below). Verified: real POST returns the 2537B fragment; taster runs in place.
>
> **THE PLATFORM FIX is SHIPPED AND LIVE — the council thread is CLOSED (2026-07-22). Not the live
> thread any more.** The chrome renderer's schema-blindness (a CLASS, not just idea.uk) went to the
> council gate as `SUBMISSION_CORR=7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5`. **All 3 rounds = REVISE**
> (round-1 void was `bugs_open/019`, since fixed & live). The base observability fix (schema resolver +
> named dead-control Errors + `bugs_open/041`'s dead-JS UNION) was committed & rolled during rounds 2–3
> by concurrent sessions and is **live in v1.0.1146+**. Round 3 (verdict 2026-07-21 11:17, REVISE,
> 11/13 approve, non-veto) surfaced three REAL objections the handoff had wrongly predicted were
> "wording": (a) the missing-field detector used a control-flow-blind **regex** that false-flags
> `{{range}}`/`{{if}}`-nested fields — ~30 active components would log false Errors; (b) a second
> silent-drop **sibling** (`RenderTemplateWithMap`) the fix hadn't touched; (c) `RenderTemplate`'s
> caller set is large/diverse (8 sites, 5 pipelines). **Owner ruled 2026-07-22: ship the fix, no round
> 4** (council is advisory, stays at REVISE, **no `Council-Reviewed:` trailer** — same posture as
> `bugs_open/053`).
>
> **Shipped `78482c86b` (2026-07-22), VERIFIED LIVE on v1.0.1149:** rewrote `missingBareFields` as a
> scope-aware `text/template/parse` walk (only ungated root-scope fields reported; regex kept as the
> unparseable-template fallback) + routed the sibling `RenderTemplateWithMap` through the same detector.
> Pod-grep (symbols created by ONLY my commit): `missingBareFieldsRegex`/`bareFieldName`/
> `scanTemplateFuncs` all present in `agent-chassis-7d4ff8b54-cm786`. **Caveat:** the sibling half is
> **dead code today** (`RenderTemplateWithMap`=0 in the binary — its only caller `rerenderContactInfo`
> has no callers; linker-eliminated), so it is *correct-if-revived*, not doing runtime work now.
>
> **`bugs_open/054` (block/escalate) is the remaining platform follow-on — OWNED BY ANOTHER SESSION,
> not started as of this handoff.** Number collision: `054_…_chrome_unresolved_field_escalation_and_
> consumer` is OURS; `054_…_unguarded_range_items…` is the relojistas thread's. My parse-tree fix
> SHARPENS the exact `inURLAttr` signal 054 will escalate on (no more false positives) — coordinate,
> don't duplicate. Do NOT re-submit `7152c7cf`; the council thread is done.
>
> **NEW BUGS THIS THREAD FILED (all real, none started):**
> - `bugs_open/030` — the dispatch queue: ONE partition, ONE consumer, so every session's trigger
>   serialises. Measured latency 16–36 min, and under load it DIVERGES (lag 21→161 in 2h). A council
>   review costs an hour+ before it starts. **Cheapest fix: print the lag at publish time** (snippet in
>   030). **Check lag before submitting anything** (`kafka-consumer-groups.sh --describe --group
>   generic-requests-group`), and NEVER re-fire a queued dispatch — it double-spends and lands further back.
> - `bugs_open/041` — chrome component JS is never published (`collectJSAssets` reads `page_components`
>   only); idea.uk's mobile menu is dead on every page (`/tools/assets/site-header.js` 404s). Fixed IN
>   the council submission (edit 5), so it lands with the platform fix.
> - `bugs_open/054` — the block/escalate follow-on (above).
> - `bugs_open/006` C addendum — a claim that dies BEFORE doing work stalls indefinitely (no
>   `claimed_at` requeue predicate exists in `platform/`); operator reset is in the addendum.
> - `bugs_open/024` — got a 2nd reproduction: `rerender-pages` sets no `spec.reason`, so it can NEVER
>   apply a template edit, for any component, on any site, while reporting success.
>
> **Still owed on the tool (owner box-side, unchanged from before, NOT this thread's work):** deploy the
> tool binary (hardened `/request` + email-subject fix + it should emit `/report.html#request-a-report`
> so `p3_03`'s client-side `#request` retarget stopgap can be deleted); prove Stripe through the new
> nginx; confirm `proxy_read_timeout`; purge Cloudflare; two SES bounce DNS records.
>
> **Four rules this workstream learned the hard way:**
> 1. **Verify against the deployed artefact, never the work-item status.** `complete` is not proof:
>    a rerender reported 9/9 complete, deployed real files, published a JS asset — and changed nothing
>    on the page (`§X.4`). Read the work item's `result` JSON for what actually deployed.
> 2. **A rerender has TWO modes and the default cannot see template edits.** `rerender-pages` sets no
>    `spec.reason` → assemble-from-stored-HTML. To apply a `content_components.html_template` edit you
>    need `reason='section_data_resolved'` (or `image_landed`/`cta_links_stale`) — insert the
>    `page_rerender` item by hand (`sql/p3_04` is the template). Guard it: that path escalates to the
>    LLM content writer (rewrites live copy) if any section has NULL `content_data`.
> 3. **A schema `fallback` is not a safe default — on a URL field it is a fabrication licence.** Applying
>    `header-bold-gradient.cta_url`'s `/contact.html` fallback on a miss re-creates the phantom-CTA bug
>    LNK-007 killed. Correct-or-absent (LNK-005): leave it unset, let the gated template render nothing.
> 4. **A rendered value that looks hand-authored may be a resolved query.** The tool-card URLs came from
>    `source: query.pages_where_type:tool` (the pointer page), not the stored `content_data`. Fix the
>    source, not the copy.

**Updated 2026-07-22.** This is the single entry point to continue the idea.uk → VM workstream.
Read `SUMMARY_idea_uk_vm_site.md` for the plain-English state, then this for the operational detail.
Companions in this directory: `PLAN`, `RUNBOOK`, `RUNNING_NOTES` (execution log — newest at the
bottom, §X.1–§X.8 cover this thread; §X.8 is the round-3 close), `README_where_we_are.md` (owner's
plain-prose log), `council_submission_chrome_schema_driven.json` (the council submission — CLOSED, do
not resubmit), and `sql/` (every DB change applied, in order — `p3_01`…`p3_04` are this thread's). The
`HANDOFF_replan_clobbers_built_pages_FIX.md` here is a SEPARATE chassis-fix task.

**Where to pick up (new thread, in order):**
1. **The council thread (`7152c7cf`) is CLOSED — shipped on the owner's ruling, live in v1.0.1149. Do
   NOT read its verdict expecting a live decision, and do NOT resubmit.** (See START HERE + NOTES §X.8.)
2. `bugs_open/054` (block/escalate — make an unresolvable render field ESCALATE, not just log) is the
   remaining platform follow-on. **OWNED BY ANOTHER SESSION — check `scripts/who-owns.py 054` and read
   their docs before touching it.** It consumes the `inURLAttr` signal my parse-tree fix just made
   accurate; it also overlaps `bugs_open/023` fix #3 (same consumer). Coordinate, don't build a
   parallel handler. Number collision: resolve `054` by slug (chrome-escalation is ours).
4. Side-finding to close: `sites.content_data` for idea.uk still holds the stale
   `idea-uk@leopardess.uk` (a reviewer's check surfaced it; the p1_05/p1_06 sweep missed this column).
   Not rendering today, but a live wrong address one code path away.

## Goal
Make idea.uk one complete site behind the VM's nginx: the chassis-built static pages **and** the live
£29 tool, on one origin. Today they're disconnected — static → B2 (invisible; DNS → the VM), and the
VM's nginx serves only the tool.

## Working rules
Go not Python. British English. **Schema first** (`\d <table>`, read the function before changing it).
Structural fixes over patches. Reuse existing functions. `logger.Info` not `.Debug`. A 0-row result
isn't decisive until the query is cleared. Go changes are **inert until a chassis image rebuild**;
DB/workflow config is live immediately. The idea **tool** is a separate stdlib-only Go module
(`docs024_key_docs_latest/idea.uk/golang_files/`, `module idea`) with **no CI** — ship by building a
linux/amd64 binary, scp to the box, `systemctl restart idea`.

## Key facts
- idea.uk site_id `1244516d-014d-421c-88c6-090bb1e9552a`.
- Box: Hetzner (Nuremberg) `116.203.204.115`, `ssh root@116.203.204.115`. Tool: systemd `idea`,
  `127.0.0.1:8080`, orders in `/var/lib/idea/orders.json` (a FILE, **no DB**), env `/etc/idea/idea.env`.
- **Not this box:** `167.233.33.159` is relojistas' box. `setup.sh` has takeover semantics — never
  point it at the live idea.uk box.
- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
- Deployed chassis image at last check: **v1.0.1123** (built from local filesystem, not git — verify
  behaviour against the pod, not commits).

## DONE (this workstream)
1. **Credential scrub** — `idea.env.example` real AWS SES creds + `INTERNAL_API_KEY` replaced with
   placeholders; `scripts/check-secrets.sh` + `.githooks/pre-commit` guard installed
   (`git config core.hooksPath .githooks`). Stripe/Anthropic there were placeholders — never exposed.
   ⚠️ **Guard + hook are UNTRACKED** — commit them so other clones get the protection.
2. **Site completed** — 9 coherent pages, all `deployed`. `guides-index` + `news-index` composed
   (were 404). `tool-audience-check` is a **pointer page** (`url=/audience-check`, `build_status=
   deployed` pinned to the current plan so reconcile skips it, 0 sections); the tool-list cards on
   `index`+`tools` link straight to the live `/audience-check`.
3. **Per-site deploy target wired** — `resolveGitRepoName` (`helpers.go:206`) now called by
   `GitCommitAction` + `deploy_image_asset_action`; `upsertSite`/`EnsureSiteRecordAction` surface
   `github_repo`. Committed and **shipped in v1.0.1123**. **ACTIVATED 2026-07-16** — see #7–#9.
4. **`/request` hardened** — honeypot (`company_url`) + timing gate (`_elapsed` < 2500ms) + intake
   rate limit (`newIntakeLimiter`, 5/hr+15/day) + `mail.ParseAddress` validation + length caps + IP/UA
   capture on the `Order`; `request_hardening_test.go` (6 subtests PASS); `go build`/`vet` clean.
   NOT yet deployed to the box.
5. **Contact email** set to `idea.uk@contactforsales.com` across all sources the validator COALESCEs
   (`sites.email` is the canonical; `site_specs.identity.email` is what renders — both aligned; see
   `sql/p1_05` + `sql/p1_06`).
6. **Docs corrected** — `../idea_uk_section_data_missing/HANDOFF_spam_and_ip_blocklist.md` rewritten
   (it named the wrong process + datastore); `spam_read.sql` neutered as void.
7. **§2b DONE (2026-07-16)** — `gqls/vm-sites` Action guarded: `deploy-targets.json` allowlist
   (relojistas.com → its box only; unmapped domains skipped), `VM_HOST` secret retired. Verified
   LIVE: idea.uk skip proven 3×, relojistas deploy green through the map. Secrets guard + hook now
   tracked (were swept into `.gitignore` by bulk commits).
8. **vm-sites runner EXISTS now** — the Action had NEVER run (no runner on the repo; runner image
   had no ssh/rsync — silent exit-127). Fixed: image `aqls/github-actions-runner:v1.0.1126`
   (+openssh-client +rsync) + new `github-actions-runner-vmsites` deployment
   (`deployments/kustomize/services/github-actions-runner-vmsites/`). RUNNING_NOTES §N.
9. **§2c DONE + repo seeded** — `sites.github_repo='vm-sites'` for idea.uk (rollback: set NULL);
   `gqls/vm-sites` seeded with the built artefact from `gqls/sites` (8 pages + assets, 4cbaf2a).
   ⚠️ RUNBOOK §3b corrected: static `terms.html`/`refund-policy.html` DO exist and are footer-linked
   with `.html` — cutover config needs 301s to the tool's canonical legal pages.

## PENDING — next actions

### Owner (need box SSH / external access; can't be done from the chat sandbox)
- ~~ROTATE the exposed creds~~ **DONE 2026-07-17** — old SES user deleted, new SMTP user verified,
  `INTERNAL_API_KEY` rotated, service restarted healthy. Leaked history values are dead. `/op` links:
  issue fresh on next use (old ones no longer verify).
- **Deploy the hardened tool**: `cd …/idea.uk/golang_files && GOOS=linux GOARCH=amd64 go build -o idea .`,
  scp to `/opt/idea/idea.new`, mv, `systemctl restart idea`. RUNBOOK Phase 4 shipping note.

### Next (all remaining steps run on the box — owner's hands; the chat prepares/verifies)
1. ~~Provision pull-sync on the idea.uk box~~ **DONE 2026-07-18** — `/var/www/idea.uk` holds all 8
   pages, `sitesync.timer` syncs every 5 min, read-only deploy key accepted, nginx untouched.
   Traps fixed en route: `ssh` ignores `$HOME` (`/bugs_open/016`); `scp -r` nests on an existing
   destination (RUNBOOK §3a). RUNNING_NOTES §S.
2. **nginx cutover** (RUNBOOK §3b–3e): static root + proxy the **16** reserved tool paths
   (`service.go:527-543` — the full list is in the RUNBOOK; the old runbook's 7-path list would break
   the taster + operator flow) **+ the three `.html→` 301s for the legal pages (§3b correction)**.
   Prove `/stripe/webhook` through the new config BEFORE cutting over.
3. **Real-client-IP in nginx** (RUNBOOK §4a): idea.uk is behind Cloudflare but `setup.sh` never sets
   `set_real_ip_from`/`real_ip_header CF-Connecting-IP`, so nginx (and the new `/request` IP capture)
   would see Cloudflare's IP. Confirm the record is proxied (orange) first. Needed before any IP block
   list is meaningful.
4. **Remove existing spam** from `/var/lib/idea/orders.json` (owner-side; RUNBOOK §4c): back up, filter
   the all-`test` rows, restart. No DB, no `Delete` method — edit the file.

## Landmines (do not relearn the hard way)
- **Never re-run `build-site-planner` to compose missing pages** on a non-adoption-locked site — it
  silently regresses built pages and can't fill empty ones. Full write-up:
  `HANDOFF_replan_clobbers_built_pages_FIX.md` + memory `replan-clobbers-built-pages`. To compose one
  page, drive its build; don't re-plan.
- **git-adapter `createOrGetRepo` makes repos PUBLIC** — create any target repo by hand.
- **`page-build-handler` claim-timeout churn**: a build can succeed (page deployed) yet its work item
  reverts and re-runs. Verify via `page_components`/deployed HTML, and mark a verifiably-complete item
  `complete` to stop the churn. Logged in `aaa_fails_to_mend/006`.

## Open decisions (none blocking)
- ~~`/privacy` after cutover~~ **RESOLVED 2026-07-18: the tool keeps all three legal pages**; the
  static `.html` copies 301 onto them (already staged in `box/idea.uk.nginx`).
- ~~`/contact.html` form~~ **RESOLVED 2026-07-17** — mailto (owner's choice); staged at source
  (`sql/p1_07_contact_form_mailto.sql`), publishes on the next contact-page build. Also fixed a stale
  `idea-uk@leopardess.uk` in the form description. RUNNING_NOTES §Q.

## Errors parked for other chats
`aaa_fails_to_mend/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md` — (A) crash-looping runner replica,
(B) fleet-wide dead contact-form, (C) claim-timeout churn. `001_…replan_clobbers_built_pages_FIX.md`
— the planner bug.
