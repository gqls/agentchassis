# HANDOFF 2026-08-21 — webdesign.uk build service + site delivery

**SUPERSEDES** `HANDOFF_2026-08-20_continue_here.md` (bannered). The merge with
`../site_delivery_and_editor/` stands: one session drives both lanes, each keeps its own
NOTES / PLAN / RUNBOOK / README / SUMMARY, and findings go in the lane where the work
happened.

**Read order, cold:**
1. this file
2. `PLAN_2026-08-21_todo_from_here.md` — the ordered work list, kept current
3. this lane's NOTES tail (the four 2026-08-21 entries)
4. `DECISION_2026-08-21_domain_transfer_out_from_nominet.md` — before touching domains,
   transfers or the £200
5. `DECISION_2026-08-21b_zip_download_link_needs_a_credential_home.md` — before touching
   the download link, storage credentials, or core-manager's exposure
6. `DECISION_2026-08-21e_owner_reviews_before_the_delivery_email.md` — before building the
   delivery email
7. `README_where_we_are.md` (owner-facing, plain prose)
8. `../site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md` for the
   original build order; register `DGH-011` (ZIP) and `DGH-014` (handover state)

---

## 0. State in one paragraph

The commercial position is settled, attested and verified at the served pages and the chat
bot: £149 in full before the build, one-shot, no approval stage, no changes, no refunds,
delivered as a ZIP to keep plus the site live at our address for about a month, domain
rentable at £10/mo or buyable for £200, **.co.uk and .uk only**, **and a content policy
that amends an unacceptable brief rather than cancelling it**. **Nothing can be sold yet**:
there is no way to take money (Stripe deferred by the owner) and no way to deliver what was
bought (the delivery email is not built). Phase 4's confirm-transfer link is built and
**live in the cluster but not reachable from the internet**; the download link is designed
but blocked on one owner decision. Two owner decisions are outstanding and both are named
in §2.

## 1. What the owner ruled on 2026-08-21 (four rulings, all applied or recorded)

1. **TLD scope: *"we only sell .co.uk and .uk tlds for now."*** Attested as fact
   `domain_tlds_offered` (`SQL_2026-08-21`). Closes the old item 6, open since 08-19.
2. **Registrant: *"my name until we agree a sale."*** So every £200 sale includes a
   **Registrant Transfer performed by us**, with a Nominet fee. That is the process now,
   not a branch of it.
3. **Content policy: no porn, violence, politics or otherwise distasteful sites — and the
   remedy is an AMENDED BRIEF, not a refund.** Attested as `content_we_will_not_build`
   (`SQL_2026-08-21c`, corrected by `21d`). *"we reserve the right to change the brief and
   deliver a site that is within the bounds of respectability within their genre of
   request."*
4. **The owner sees the client brief and the rendered site before the delivery email
   goes out.** Designed, not built — decision doc `21e`. Routes through the **work-item
   queue**, which already has a screen. Do not build a new UI.

Plus: **Stripe deferred** (*"I will do Stripe later"*), so it no longer gates items 2–5 of
the plan; it gates only the moment money can actually be taken.

## 2. ⚠ THE TWO DECISIONS OWED BY THE OWNER — both block real work

### D-A. May core-manager be reachable from the internet at all?

**This is the council's gating objection and it is correct.** The `/c/<token>` submission
bundled two decisions: *how a confirmation should work* (the owner ruled on that
2026-08-19: click-is-state, no form) and *whether the service holding every site's data
becomes publicly reachable* (never asked). `sitefacts.go` in that same service documents
"NO PUBLIC EXPOSURE ... ClusterIP only, no Ingress". Options: **(a)** approve public reach
for named paths only, **(b)** move `/c/` and `/d/` to a service allowed to be public,
**(c)** keep both behind the box and let the box hold the state. Full text of all six
seats' objections: `DECISION_2026-08-21b` §5–6.

### D-B. Does the "two or three days" promise absorb the owner's review, or get re-cut?

`build_duration` attests *"usually ready in two or three days"* and was already tightened
once (from next-day) because measurement refuted the faster figure. A human review step
spends that budget: if the owner is away two days, the **gate** breaks the promise rather
than the build. Three options costed in `21e` §3; none chosen.

## 3. What is LIVE, and how it was proven — all re-checked 2026-08-21 ~19:30Z

| thing | state | proof |
|---|---|---|
| TLD scope `.co.uk` / `.uk` | **LIVE** | bot: *"We only register and rent .co.uk and .uk domains"*; before the change it improvised *"right now those are on .uk domains"* off no fact |
| Content policy | **LIVE** | bot separates adult-adjacent business from pornographic content and offers to build within the customer's own line of work. Polled until the 5-min facts cache turned over, not assumed |
| `writer_block` caught up with the 08-19 rulings | **LIVE** | 0 occurrences of "next day" / "free to transfer" / "transferred freely" / "never a range of days"; contains `build_duration`'s own writer_line |
| Register | **24 facts, 34 bans** | last row written 2026-08-21 19:26:15 by `SQL_2026-08-21d`. Two lanes write this row — compare against the row your transaction supersedes, never against these numbers |
| Served pages | **clean** | `index` 6× / `faq` 3× "two or three days", 0× "next day" |
| `/c/<token>` confirm-transfer | **LIVE IN CLUSTER, NOT PUBLIC** | see the box below |
| `/d/<token>` download | **NOT BUILT** | blocked on D-A; design settled otherwise (§4) |
| Delivery email, chase, retraction | **NOT BUILT** | — |
| core-manager / agent-chassis | **v1.0.1323** | rolls several times a day; re-read before trusting |
| Nominet transfer mechanism | **VERIFIED** | primary sources, 2026-08-21. See `DECISION_2026-08-21` |

> ### ⚠ `/c/` IS ALREADY RUNNING IN PRODUCTION, AND THE COUNCIL SAID REVISE
>
> Not a mistake and not an emergency, but know it before you touch anything. On this tree a
> commit ships on whoever rolls next: core-manager `v1.0.1323` was built from
> `70e7b4f9` which contains the route commit `47185f0d5`, and
> `GET http://core-manager:8088/c/x` answers **200** from inside the cluster right now. The
> guardian seat predicted exactly this ("even if nginx is not yet deployed, in-cluster
> callers ... hit it as coded").
>
> **What makes it safe today, measured rather than assumed:** the nginx block is written
> but **NOT deployed** to the box, so nothing outside the cluster can reach it; and
> `customer_access_tokens` holds **0 rows**, with **0 sites handed over and 0 confirmed**.
> There is no token to redeem and no state to corrupt.
>
> **What that means for you:** do not deploy the nginx block until D-A is answered. The
> block lives in `box/webdesign.uk.nginx` and applying it is `nginx -t && systemctl reload
> nginx` on the box — one command away from making D-A moot by accident.

## 4. Storage credentials — RESOLVED, and no new container is needed

The owner directed: *"The storage credentials are in the framework, please search how to
use them. Create an agent container that uses the s3 client and you can add that container
type to the spawn action to contain those credentials."* He is right about the mechanism:

- **`isStorageEnabledAgent`** (`platform/orchestration/actions/spawn_actions.go:3049`) is
  the sanctioned per-type grant. A spawned pod of a listed type gets the four AWS/B2 keys
  by **`secretKeyRef` against `personae-storage-secrets`**, plus `S3_ENDPOINT` /
  `S3_REGION` / `IMAGE_BUCKET` from the `storage-config` ConfigMap. That is
  `bugs_open/245`'s fix: credentials never pass through the spawner's own env, and a
  missing key fails LOUD (`CreateContainerConfigError`) rather than silently at first use.
- **`zip-deliverer` is already on that list**, is live and active (`category: executor`,
  steps `zip` / `complete`), and its `zip` step **already mints a presigned URL**. So
  **nothing new holds credentials, the spawn action is not edited, and
  `isStorageEnabledAgent` is not touched.**

**The open question is WHEN, not where, and it is a latency fact.** A spawn is a pod start;
a browser following a download link cannot wait for one, and dispatch is dropped within
~300s of a chassis restart. So the shape is **pre-mint and refresh**: store the presigned
URL against the `zip_download` token, make `/d/` a pure DB lookup and 302, and refresh
before the 7-day signing ceiling (six weeks in ~7 hops).

> **AND THE ONE THING THAT MUST BE BUILT INTO IT, not written in prose:** if the refresher
> stops, every live link dies a week later and nobody learns until a customer says so.
> `/d/` must compare the stored expiry against now and, when stale, render an honest page
> **and file a work item**. That turns a silent death into a queue row somebody sees.

## 5. The council REVISE, and what each seat is owed

`99b5af22-7150-4e91-a5e3-809fd06504c0` — **REVISE**, gated by `guardian` (high).
`complete_revise | FAILED` is the terminal state. **Resubmit on the SAME correlation**
(`RESUBMIT_CORR=99b5af22-…`) so the trail accumulates.

| seat | sev | what is owed |
|---|---|---|
| `guardian` | **high** | D-A above. Unbundle the exposure decision from the confirmation-UX ruling |
| `editquality` | med | Rate limit keys on `$http_cf_connecting_ip`, a client-suppliable header, with no `real_ip_from` shown. **Verify the trust boundary, don't argue it**: nginx binds `127.0.0.1` only and ingress is a cloudflared tunnel dialling out, so a forged header must originate on the box — but that is currently only a comment |
| `editquality` + `guardian` | med | The prefetch hazard ships unmitigated. **I was wrong that it needed the minting site**: refuse to mutate on `Sec-Purpose: prefetch` / `Purpose: prefetch` / `X-Purpose: preview` and on HEAD. Independent of how tokens are minted |
| `tooling_provenance` | med | The `/d/` deferral lives only in markdown, not `doc_notes`. Write the DB-resident entry |
| `reuse_agent` | low | Check for an existing handler-deps testing convention in the package, and for another token-confirm-by-GET handler. Say what the searches found |
| `guardian` | low | Nothing stops someone applying the nginx block later without re-review. State the gate in the file |
| `architecture` | low (approve) | Watch accumulation: a second and third publicly-proxied prefix should trigger a boundary review |

## 6. STILL OPEN (beyond §2 and §5)

1. **The delivery email** (plan item 3) — unblocked, and the next build. It can ship
   **without** the ZIP link and gain it later. ⚠ **Mint the confirm token NOT single-use**
   (`21e`/`21b` §4): a mail scanner's prefetch would otherwise spend it and lock the
   customer out. Then the weekly chase, then the retraction job.
2. **Nothing enforces the no-approval-stage rule.** `[MEASURED 2026-08-21]` *"You will be
   able to approve the site once you have seen it"* scans **CLEAN** against the live
   register. `one_shot_no_approval` is attested and `writer_block` forbids approval copy,
   and there is no ban. **This matters more now the owner has asked for an internal review
   step** — internal steps leak into copy. The fix is an **offer-shape** ban (the 08-19
   `round of changes` narrowing is the worked precedent); a bare-token ban blocks the
   denial too. Prove it with a probe set carrying BOTH halves.
3. **`writer_block` breaks its own first rule** — "never use an em dash" and it contains
   **six** (paragraphs 12, 14, 20, 21). Counted 6 before and 6 after the 08-21 edits, so
   none were introduced then. Five-minute fix.
4. **The chat bot broke its own conduct rule** in a verified answer on 2026-08-21: it used
   an em dash. `promptConduct` forbids it and `facts_test.go` tests the conduct string for
   exactly that. Different failure from the register's (there the ban is enforced at the
   gate; here the rule is in the prompt and the model ignored it).
5. **There is no terms page.** 8 pages, none of them terms, while `writer_block` already
   tells the writer to point at "the full terms". The content policy and the commercial
   terms are attested but have nowhere to live as a page. Framework job.
6. **`bugs_open/299`** — the home-page CTA names the Website Brief Starter and its href
   **dials the phone**. Filed, deliberately unpatched: the producer question survives every
   rewrite.
7. **`what-you-get` shrink gate** — `call-to-action 594→264 visible chars (44% kept, floor
   50%)`. Same CTA as 299. Raising the floor silences a copy decision rather than making one.
8. **The prompt-maker pointer is DUE** — the chat conduct deliberately does not name the
   Website Brief Starter because its guide page was selling the retired model. It no longer
   is. Read the live conduct before editing.
9. **Contact email** `webdesign@contactforsales.com` (domain mismatch, item `a8d6f440`).
10. **Stripe** keys, webhook edge exception, webhook hostname — owner-deferred.
11. **Second Nominet TAG** — domain programme only, not this lane's critical path.
12. **Reseller market supposition — PARKED, do not develop.** Written to no fact,
    writer_block, mission or spec. Do not encode it, do not reopen it unasked.
13. **HITL as an orchestration step — not next.** The work-item queue is the sanctioned
    route (`21e` §2). The orchestration path is measured dead: `collect_via_hitl` 0,
    `brief_answers` 0, `hitl_mode` 0 across 369 briefing orchestrations, against
    `briefing_answers` = 3 as the control.

## 7. Dated checkpoints — a date with no owner is not a plan

| when | what | where |
|---|---|---|
| before the FIRST domain sale | prove Nominet EPP access still works. **A greeting is not a login** — only a completed login tests the IP allowlist, and the greeting is served to any IP. Pin to IPv4 | RUNBOOK, "Transferring a sold domain out" |
| **by 2026-12-01** | re-read Nominet's transition detail: has the TAC flow, locking or timing moved since the 4 June 2026 notice? | https://registrars.nominet.uk/registry/dot-uk/faq/ |
| **2027-02-09** | transition day. Nominet retires IPS TAG transfers for a **Transfer Authorisation Code**. Rewrite RUNBOOK step 2; `domain_buy_once`'s *"we give them what they need to move it"* becomes literally true; **fold a pre-issued code into the delivery email** — that is the product win | same |
| after the first real transfer | correct the RUNBOOK from what actually happened. It is written UNTESTED on purpose | RUNBOOK |

## 8. Traps this work has paid for (read before editing the register)

- **A fact has TWO consumer-facing fields with DIFFERENT READERS.** `writer_line` steers
  the page writer via `writer_block`; **`claim` is what the chat bot reads out verbatim**
  (`renderSystemPrompt` writes `"- " + f.Claim` and never sees `writer_line`). Narrowing
  one does not narrow the other. This cost a live defect on 2026-08-21: the bot answered
  *"Yes, we can build a site for that"* to an explicit adult brief minutes after the
  content policy went live. **Narrow claim, writer_line and context_terms together, and
  verify at the BOT, not at the row.**
- **A fact edit that asserts `writer_block` UNCHANGED is correct AND is how a retired value
  keeps steering the writer.** Add the agreement check: grep `writer_block` for the string
  your edit RETIRES. Full entry in `LANDMINES.md`.
- **Two contradictory claims in one prompt is not a coin toss.** The permissive one answers
  the customer's question and the restrictive one reads as being about something else.
  Collisions produce a confident wrong answer, not hedging.
- **Prove a ban or copy change with a probe set carrying BOTH halves** — the shapes that
  must block AND the innocent shapes that must pass. `go run ./cmd/claimscan -evidence
  <eb.json> -components <tsv>`. A must-pass-only set cannot tell a clean scan from a dead one.
- **Do NOT put a live instance of a banned phrase into an instruction.** Prompt text is read
  as an example. A draft turnaround rule was itself BANNED on "in one day".
- **Never assert a fixed fact/ban count** — two lanes write this row. Compare against the
  row your transaction supersedes.
- **Verify SQL guards by reconstruction** (apply the same `replace()` chain to the
  superseded text, demand equality) and **mutation-test them in a rolled-back transaction
  before applying**. A guard that has only seen the state it was written for proves nothing.
  Read the clean run's output too: it must print `INSERT 0 1`, not pass on a no-op.
- **A `complete` work item is not a repaired artefact** — verify at the served page.
- **`submission` embeds its own differing copies** of mission_brief and roadmap_brief;
  `content_direction` carries a rendered `formatted` duplicate. Fix one, the other stays
  stale and authoritative.

## 9. Falsifiers — check these before trusting anything above

- A newer handoff in either lane dir.
- The register's counts (**24 / 34**) and the row's `notes` — two lanes write it.
- §3's bot checks, at the bot, polling until the 5-minute facts cache turns over.
- The served pages, not the statuses.
- The image tags (**v1.0.1323**) — they roll several times a day, and **whether `/c/` is
  still only in-cluster**: `kubectl -n ai-persona-system exec <core-manager pod> -- sh -c
  'wget -q -S -O /dev/null http://127.0.0.1:8088/c/x 2>&1 | head -1'` and, separately,
  whether the nginx block has been applied on the box.
- `customer_access_tokens` row count (**0**) and `sites.handed_over_at` /
  `transfer_confirmed_at` (**0 / 0**) — the moment any of these is non-zero, the "nothing
  to corrupt" reasoning in §3 expires.
- Whether Stripe keys, the webhook exception or the second Nominet TAG have landed.
- The council correlation `99b5af22-…` — has it been resubmitted, and what did it say?
