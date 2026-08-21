# HANDOFF 2026-08-20 — webdesign.uk build service + site delivery (Phase 4 is now genuinely next)

**SUPERSEDES** `HANDOFF_2026-08-19_continue_here.md` (bannered). That file merged this
lane with `../site_delivery_and_editor/`; **the merge stands** — one session drives
both, and each lane still keeps its own NOTES / PLAN / RUNBOOK / README / SUMMARY and
its own register. Write findings into the lane where the work happened.

**Read order, cold:** this file → this lane's NOTES tail (the 2026-08-19 and 08-20
entries) → `../site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md`
(owner decisions + build order) → `PLAN_2026-08-14_site_delivery_and_editor.md` §Phase 4
→ `README_where_we_are.md` (owner-facing) → register `DGH-011` for the ZIP mechanism
→ `DECISION_2026-08-21_domain_transfer_out_from_nominet.md` before touching anything
about domains, transfers or the £200.

---

## 0. State in one paragraph

The commercial position is settled, applied at the register and verified at the served
pages: £149 paid in full before the build, one-shot, no approval stage, no changes
included, no refunds, delivered as a ZIP to keep plus the site live at a link we host
for about a month, domain rentable at £10/mo or buyable for £200 one-off, any sort of
site, no pre-sales service, delivery "two or three days". **Every page now carries the
attested turnaround and every active component scans clean against the live register
(0 findings across 20).** The retired next-day figure is now BANNED, so the class
cannot return. Phase 3 (the ZIP deliverable) is complete and live-proven.
**Phase 4 — handover state + the delivery email — the STATE half is built (see §2); the
HTTP surface, the email, the chase and the retraction job are the next build.**

> **UPDATED 2026-08-21:** the TLD scope is now owned and attested (`.co.uk` and `.uk`
> only, §3 and §5 item 6), and `writer_block` has been brought back into line with the
> 2026-08-19 rulings it had been left behind by — one of which was instructing copy the
> deploy gate refuses. Nothing else in this file moved. One owner ruling from 2026-08-19 evening is still unactioned: lengthen the
ZIP download link from 7 days to 30.

---

## 1. What CLOSED since the 08-19 handoff (all verified, none inferred)

### The five turnaround rebuilds are DONE

The `tool-website-brief-starter-guide` was the last one wrong. It took **four**
attempts and was stopped by **three different defects**, only one of which was the
writer's doing. It completed 2026-08-19 16:21:39Z.

| check | result |
|---|---|
| served page, `next day` | **0** |
| served page, `two or three days` | 1 |
| stored `article-body`, `next day` | **false** |
| component `updated_at` | 2026-08-19 16:21:39Z — so the components really were rewritten, not rerendered |

Repeat it (and note it checks BOTH directions — a zero on "next day" alone is also
satisfied by copy that dropped the turnaround entirely):
```bash
for p in index faq how-it-works what-you-get guides/tool-website-brief-starter-guide; do
  echo -n "$p: "; curl -s "https://preview.webdesign.uk/$p.html?cb=$(date +%s%N)" \
    | grep -o -i -e "next day" -e "two or three days" | sort | uniq -c | tr '\n' ' '; echo
done
```

### The next-day ban is ARMED — `SQL_2026-08-19e`

A promise shape, not a bare token (see §6). **Its census guard ran for real twice:**
it REFUSED on 2026-08-19 naming `tool-website-brief-starter-guide/article-body`, and
passed on 2026-08-20 on a census of 0. That is the `bugs_open/161` order — repair
first, then arm — made mechanical instead of remembered.

Verified after arming, at the live register: **34 bans (was 33), 22 facts unchanged**;
the three offending shapes fire, five must-pass sentences clean; **all 20 active
components scan clean, 0 findings.**

### Three rule defects fixed, and one of them had never worked at all

- **`SQL_2026-08-19f`** — the `round of changes` ban narrowed to OFFER shapes, on the
  owner's ruling. Denials went from **3 of 6 blocked** to **0 of 6**; offers still
  5 of 5. Over the whole corpus it loses nothing and gains nothing.
- **`SQL_2026-08-19h`** — the testimonial ban **could never have done what its reason
  said.** `claims.go:296` compiles every pattern as `regexp.Compile("(?i)" + p)`, so
  the `[A-Z][a-z]+ [A-Z]` tail that identifies an attributed NAME degraded to "a word,
  then a letter" — i.e. any quotation followed by prose. It was blocking a quoted
  question and a quoted anti-example as readily as a real testimonial. Fix is four
  characters: `(?-i)`. **Filed in `LANDMINES.md`** (verifier dispatched, correlation
  `cf717466-1d10-47c6-9a78-883f38b74050`).
- **`SQL_2026-08-19d`** — seven stale ban REASONS corrected. Five stated a retired
  commercial term as the CURRENT position (one told a reader "a refund is available
  until the customer accepts the site"); two cited facts that no longer exist.

> **CORRECTION carried forward, because the lane believed the opposite:** the site
> writer **never reads a ban reason.** `page-content-writer`'s template pulls
> `{{.site_specs.specs.evidence_base.writer_block}}` and nothing else, and matches
> neither `%banned%` nor `%banned_claims%`; every Go consumer of `BannedClaims` is a
> VALIDATOR, and no `ValidationIssue` is fed back into a writer prompt. **A ban cannot
> teach; it can only stop.** The steering levers are `writer_block` and the work-item
> brief. The reasons still matter — `checkBannedClaims` copies them verbatim into the
> blocker message a triaging session reads — but do not try to fix a copy problem by
> editing a ban's reason.

### `domain_buy_once` re-attested — `SQL_2026-08-19g`

Old wording (*"the customer is then free to transfer it"*) was wrong twice: "free to"
is an option where the owner means an obligation, and it **generated banned copy** —
the 15:56Z blocker was the writer elaborating it into *"…whenever you like"*.
New writer_line: *"Buying the domain is a one-off £200. It is then yours, and you move
it to your own registrar; we give you what you need to do that."*

---

## 2. THEN: Phase 4 — handover state (the next build), and the one ruling still owed

> **UPDATE 2026-08-20, later the same day: the STATE half is BUILT and LIVE.** Register
> entry **DGH-014**, migration **511** applied and recorded, `platform/delivery` in the
> build, council submitted (`905d9078-86c2-47a7-af0a-781723a46c08`). Schema:
> `sites.handed_over_at`, `sites.live_link_expires_at`, `sites.transfer_confirmed_at`,
> and `customer_access_tokens` — ONE hashed/expiring token table for every
> customer-facing link, `purpose` a CLOSED CHECK (`zip_download`, `confirm_transfer`;
> Phase 5's `editor_session` next, and it costs a migration on purpose).
>
> **What remains, in order:** the HTTP surface (`/d/<token>` mints a clamped presign and
> redirects; `/c/<token>` records the confirmation), then the delivery email through
> `platform/mailer`, then the weekly chase, then the retraction job that gives
> `live_link_expires_at` teeth. **Nothing mints or redeems a token in production yet and
> the helper has no live caller** — treat its deployment contract as unverified until
> something real calls it.
>
> **Migration practice, learned here:** `--apply` takes EVERY pending file and the pending
> list is other lanes' work, several of whose probes come back "inconclusive, the live
> config has drifted". Apply yours BY HAND, then register it with `--record-only <file>
> --note "<why>"`. 511 was done that way.
>
> Full account, including the two mutations that proved the harness: this lane's sibling,
> `../site_delivery_and_editor/NOTES_site_delivery_and_editor.md`, 2026-08-20 entry.

Source of truth: `PLAN_2026-08-14` §Phase 4 mechanics + `PLAN_2026-08-17` decision 3.
Owner confirmed 2026-08-19 evening that this is what to pick up next, and that the
delivery EMAIL's wording waited on the £200 question — **which is now answered**, so
the email is no longer blocked on it.

**What Phase 4 has to carry**, folding in the 2026-08-19 decisions:

- `sites.handed_over_at timestamptz` — the stamp. Single Go reader (the Phase 5 editor
  gate). The register entry must state explicitly what handover does NOT gate
  (deploys, rewrites, locks).
- **A 6-week expiry on the live link.** *Nothing expires today* — serving is unbounded
  (git repo synced to B2; no scheduled retraction, no retention job, no TTL). So this
  is a mechanism to build from nothing, not a value to change. The copy says "about a
  month", which under-promises against a 6-week reality; that is deliberate.
- **A confirmed-transferred state, scoped by the owner to the simplest possible
  thing:** a tokenised link in the email, and recording the click IS the state. No
  form, no reply parsing.
- **A weekly chase email until that click arrives.** Note it has TWO subjects: the site
  off our hosting, and a bought domain off our registrar account.
- **The delivery email itself**, via `platform/mailer` (the sanctioned mailer),
  carrying: the ZIP link (dispatch `zip-deliverable-dispatch` with `{domain}`; recipes
  in `sql_for_agents/459_zip_deliverer_agent_HOLD.sql`'s header, APPLIED), the
  live-site link, a Netlify-connect invite, both domain links, and the Stripe hosted
  portal.

> ### ⚠ MIGRATION NUMBER: use **494 or higher**. 493 IS TAKEN.
> `493_loop_nested_item_key_suffixes.sql` already exists. Confirmed by the dry run of
> 2026-08-19. **Re-run the dry run yourself before applying anything** — it is per
> session and after every roll, and there has been a roll since:
> `./scripts/migration/run-migrations.sh` (takes >2 min; background it). Nothing in
> the current pending list belongs to this lane.

> ### ⚠ THE ZIP-LINK RULING CANNOT BE DELIVERED AS A NUMBER — 7 DAYS IS THE CEILING
>
> The owner ruled 30 days (2026-08-19) and then "the longest time we have, which I
> think is 6 weeks" (2026-08-20). **Both are above the maximum, and `expiry_minutes:
> 10080` is already that maximum.** A presigned URL is bounded by the SigV4 signing
> protocol at 604,800 seconds, and **nothing in this stack enforces it** — the SDK
> signs any duration and returns a well-formed URL, so a longer link mints cleanly,
> the action reports success, and it fails only in the customer's browser, as
> **`SignatureDoesNotMatch`** — which reads as broken credentials, not a long expiry.
>
> `[MEASURED 2026-08-20]` against the live bucket, key deliberately absent so the
> status is the whole answer: **`604800` → HTTP 404 `NoSuchKey`** (the control:
> signature accepted), `604801` → 403, `3628800` (6 weeks) → 403. Exact to the second.
>
> **The presign has NOT been changed and no substitute number has been slipped in.**
> Every live caller already sits exactly on the ceiling (five sites), so nothing is
> broken and nobody has hit this yet. Full entry in `LANDMINES.md`.
>
> **What delivers the intent — and it is Phase 4 work anyway:** stop shipping the
> presigned URL to the customer. The delivery email carries a link of OURS with a
> token; the click mints a fresh ≤7-day presign server-side and redirects. The window
> then belongs to our token, so "lasts as long as we host it" becomes true by
> construction rather than by a number, and the 6-week figure lives in ONE place
> instead of two. **It is the same token mechanism the confirmed-transferred click
> needs**, so build them together. This is now the recommended shape, not a workaround.

**Blocked on the owner, and this gates first revenue:** Stripe keys; the Stripe webhook
edge exception; second Nominet TAG (domain programme only).

---

## 3. Owner rulings in force (do not re-litigate)

**2026-08-21 — TLD SCOPE: *"we only sell .co.uk and .uk tlds for now."*** Attested as
fact `domain_tlds_offered` and live at the bot. Closes §5 item 6. The "for now" is
recorded in the fact's `source.attested_by`, NOT in the claim: on a page it invites
"so when will you do .com?", which nobody may answer because no pre-sales service is
included. Do not restate the £10/£200 figures inside this fact — they are attested once
each already. **This ruling answers a scope question and does NOT settle the .uk
transfer-out mechanism; if anything it makes that sharper. See §5 item 6.**


**2026-08-19 evening, four rulings, all put with their evidence and all answered:**

1. **Narrow the `round of changes` ban to offer shapes.** APPLIED (`19f`). The test
   that decided it: *if the register attests the thing, the copy must be able to deny
   it in normal English.* Three bans in this family have now been judged and they did
   NOT all go the same way — `\brefunds?\b` (attested → narrowed), `whenever you like`
   (nothing attested → ban right, copy wrong), `round of changes` (attested →
   narrowed). **Apply the test, never the precedent.**
2. **The £200 buys the DOMAIN ONLY, and the customer transfers it.** *"Keep
   nobody's-time-included absolute. We document the steps and hand over what they need;
   the transfer itself is theirs to do."* APPLIED (`19g`). **This resolves the
   collision in `no_presales_service`'s favour and supersedes the hand-holding half of
   2026-08-19's decision 6.** That fact is UNCHANGED and stays absolute; there is no
   carve-out for the £200. Do not reopen it.
3. **The ZIP download link should last "the longest time we have"** — ruled 30 days
   on 08-19, restated as 6 weeks on 08-20. **NOT DONE, and not doable as a number:**
   7 days is a protocol ceiling and the code is already on it. See the boxed note in
   §2 for the measurement and the design that does deliver it.
4. **Phase 4 handover state next.**

**Still in force from before:**

- **2026-08-19: delivery is "two or three days."** `build_duration` re-attested,
  `value: 3` (the upper bound — a stat field publishes a bare figure and a range cannot
  be one number, so the stat takes the end that cannot over-promise). The hedge lives
  in `claim`/`writer_line`.
- **2026-08-19: the caps ban `whenever you like` is RIGHT — do not narrow it.**
  Nothing we operate is open-ended. Ownership may be written as permanent; timing may
  not.
- **2026-08-19: OWNERSHIP is permanent, anything WE OPERATE is temporary.** Permanent:
  the ZIP, and a bought domain. Temporary: the hosting and the registrar account.
- **2026-08-19: the live link's MINIMUM is fine — do NOT build a month-long serving
  mechanism.** Serving is already unbounded. (The 6-week EXPIRY in §2 is the opposite
  end and IS to be built.)
- **2026-08-19: leave the apex 302.** `webdesign.uk` → `webdesign.co.uk` stays.
- **2026-08-18: a better product beats a faster promise.** Do NOT re-plumb the builder
  on this ruling — it was about the trade-off, not the mechanism.
- **2026-08-18: examples deferred.** No example-site links until this route has
  produced sites.
- **2026-08-04: every site goes through the framework.** Never hand-build a page.

---

## 4. What is LIVE, and how it was proven (not inferred)

| thing | state | proof |
|---|---|---|
| All five pages carry "two or three days" | **LIVE** 08-19 | served pages, both directions; plus the stored components and their `updated_at` |
| Next-day turnaround BANNED | **LIVE** 08-20 | 34 bans; 3 offending shapes fire, 5 must-pass clean; 20/20 active components clean |
| Changes ban narrowed to offer shapes | **LIVE** 08-19 | claimscan, 5 offers blocked / 0 denials blocked |
| Testimonial ban made case-sensitive | **LIVE** 08-19 | claimscan, 3 testimonials blocked / 3 innocent quotations clean |
| 7 stale ban reasons corrected | **LIVE** 08-19 | 7 of 33 carry `REASON CORRECTED 2026-08-19` |
| `domain_buy_once` re-attested | **LIVE** 08-19 | live fact's writer_line; new wording scans clean |
| Delivery "two or three days" at the bot | **LIVE** 08-19 | bot answer after a cache-beating restart |
| Chat prompt-maker | **LIVE** 08-18 | `434d2b64b` at the running service; smoke-tested |
| £1,200 swept from all 9 specs | **LIVE** 08-18 | seven phrases asserted nowhere |
| Phase 3 ZIP deliverable | **LIVE** | register DGH-011, canary 8/8 byte-verified |
| Chassis | **`v1.0.1320`** | pod image, 2026-08-20T16:09Z (was 1317, was 1314). The `build provenance` line is out of range in `--tail=600`, which means **"not in range", not "unstamped"**. It rolls several times a day |
| TLD scope attested (`.co.uk`, `.uk` only) | **LIVE** 08-21 | facts 22 → 23; bot answers *"We only register and rent .co.uk and .uk domains"*, with a price/turnaround control in the same session proving the other facts intact |
| `writer_block` caught up with the 08-19 rulings | **LIVE** 08-21 | 0 occurrences of "next day"/"free to transfer"/"transferred freely"/"never a range of days"; contains build_duration's own writer_line. 6 outcome guards each proven to fire against the real pre-fix row |
| Presign ceiling is 7 days | **MEASURED** 08-20 | 604800 → 404 `NoSuchKey`; 604801 and 6 weeks → 403 `SignatureDoesNotMatch`. See §2 |
| Phase 4 handover STATE | **schema LIVE** 08-20, helper UNCALLED | migration 511 recorded; 10 SQL semantics checks against real Postgres in a rolled-back transaction, harness mutation-proved twice; `go build`/`go test` green from a clean `git archive HEAD` |

---

## 5. STILL OPEN

1. **Phase 4's REMAINING half** (§2): the HTTP surface, the delivery email, the weekly
   chase, and the retraction job. The state half is built (DGH-014, migration 511).
2. **The prompt-maker pointer is now DUE.** The chat conduct deliberately does not name
   the Website Brief Starter tool, because that tool's guide page was still selling the
   retired model. **It no longer is** — the guide landed 2026-08-19. Read the live
   conduct text before editing; this item is inherited, not re-verified.
3. **`bugs_open/299`** — the home-page CTA names the Website Brief Starter and its href
   **dials the phone**. Filed, deliberately not patched. The producer question survives
   every rewrite: the section was written after the 268 fleet fix, so something still
   generates it.
4. **`what-you-get` shrink gate** — `SECTION SHRINK REFUSED, call-to-action 594→264
   visible chars (44% kept, floor 50%)`. Raising `section_shrink_floor` would silence a
   copy decision rather than make one, and it is the same CTA as 299.
5. **Contact email** `webdesign@contactforsales.com` (domain mismatch, item
   `a8d6f440`); Stripe webhook hostname; Stripe keys via terraform.
6. ~~**Which TLDs do we actually sell?**~~ **ANSWERED 2026-08-21 — owner ruling:
   *"we only sell .co.uk and .uk tlds for now."*** Attested as fact
   `domain_tlds_offered` (`SQL_2026-08-21_domain_tlds_are_couk_and_uk.sql`, applied;
   facts 22 → 23, bans unchanged) and verified at the bot, which now answers *"We only
   register and rent .co.uk and .uk domains"* where before it improvised *"right now
   those are on .uk domains"* off no fact at all.

   **The transfer-out half is now VERIFIED and RULED — see
   `DECISION_2026-08-21_domain_transfer_out_from_nominet.md`**, which supersedes the
   `[UNVERIFIED]` note this item used to carry. Owner, 2026-08-21: *"It is likely to be
   a manual step for now for each domain and that is ok for now."* In short: selling a
   domain is **two** Nominet operations, not one — a **Registrant Transfer** (owner
   change, registry-only, £10–35+VAT) and then a **tag release** (free, and the customer
   can self-serve it once they are the registrant). Procedure in the RUNBOOK, marked
   UNTESTED because zero domains have been sold. **Do not automate it** (owner ruling);
   the EPP client that would be reused is noted in §7 of the decision doc.

   > **⚠ 9 FEBRUARY 2027 — Nominet retires the IPS TAG transfer process** and replaces it
   > with a Transfer Authorisation Code, the mechanism every other TLD uses: we generate
   > a code, the customer hands it to their new registrar, transfer completes
   > immediately. Formal notice 4 June 2026. `[VERIFIED 2026-08-21 at
   > registrars.nominet.uk]` **Check the detail by 2026-12-01.** This is a PRODUCT
   > opportunity, not just an operational change: a code can be pre-issued at handover
   > and put in the delivery email, which makes "nobody's time is included" true by
   > construction — the same shape as Phase 4's ZIP token. It does NOT replace the
   > Registrant Transfer.

   **STILL OWED BY THE OWNER, and it changes what we do on every sale:** whose name the
   domain is registered in *during the rental*. Ours protects the rental but makes a sale
   two operations; the customer's makes a sale nearly free but lets a renter walk off
   with the domain having never paid the £200. Recommendation and the full table:
   decision doc §4. Nothing encoded either way.
7. **`writer_block` breaks its own first rule — MEASURED 2026-08-21, deliberately NOT
   fixed.** Its opening line is *"Never use an em dash. Not anywhere, not once."* and
   the block contains **six**, in paragraphs 12, 14, 20 and 21. Counted before and after
   the 08-21 edits: 6 and 6, so none were introduced by them. Left alone because it was
   a third step away from that session's ask and, unlike the next-day instruction, it
   blocks nothing. It is not nothing either: prompt text is read by the model as an
   example, and the lane already treats this class as test-worthy for the sibling prompt
   (`box/chat-service/facts_test.go` tests the chat conduct for exactly this, on the
   grounds that *"a half-followed rule is worse than no rule"*). Five-minute fix; the
   replacements the rule itself prescribes are a full stop, a comma, a colon or brackets.

8. **HITL as a briefing step** — owner accepted the ordering: questions first (DONE),
   then HITL, routed through the **work-item** queue, which has a working screen. The
   orchestration HITL path has never fired: `collect_via_hitl` 0, `brief_answers` 0,
   `hitl_mode` 0 across 369 briefing orchestrations, against `briefing_answers` = 3 as
   the control.
9. **Reseller market supposition — PARKED by the owner, do not develop.** Discussion
   only; written to NO fact, writer_block, mission or spec. Do not encode it, do not
   re-open it unasked. (Separately and NOT parked: *"our websites aren't necessarily
   for business owners"* is already encoded and verified — `any_site_type`.)

---

## 6. Traps this joint work has paid for

- **The REGISTER is the wire.** Never steer via item-spec prose, never hand-edit HTML.
  `writer_block` edits are BY ANCHOR with exactly-once guards.
- **Never assert a fixed fact count on this register** — two lanes write it and the
  other edits IN PLACE. Compare against the row your transaction supersedes.
- **A bare-token ban blocks the DENIAL too.** The negation guard scans **backwards
  only** and excludes bare "no" and "non-" *by design* (claims.go), so "we do not
  include X" is safe and "there is no X" is not. Some bans block denials **on
  purpose** — the `template` ban says so in its own reason — so read the reason before
  calling one a defect, then apply the attestation test in §3.1.
- **Every pattern is compiled `(?i)`.** A ban that leans on capitalisation silently
  matches everything, and it fails by BLOCKING, which reads as a strict gate rather
  than a broken one. `(?-i)` inside the pattern is the fix. Full entry in
  `LANDMINES.md`.
- **Prove a ban change with a probe set carrying BOTH halves.** The shapes that must
  still block AND the innocent shapes that must pass. A set of only must-pass
  sentences cannot tell a narrowed ban from a disarmed one — an early candidate for
  `19f` let *"We give you a round of changes"* through and only the BLOCK half caught
  it. The instrument is `go run ./cmd/claimscan -evidence <eb.json> -components <tsv>`;
  it runs the same engine as the deploy gate. Recipes in this lane's NOTES, 2026-08-19.
- **A guard that has only seen the state it was written for proves nothing.** Every SQL
  applied on 08-19 was mutation-tested first, and two of those mutations found real
  problems: a probe run caught a contradiction between "assert the stale strings are
  gone" and a convention of quoting what you replace.
- **The damage from an over-broad ban is an ABSENCE.** A page blocked at save never
  becomes a stored component, so a corpus census returns the same findings before and
  after the fix. Only a self-authored probe set can see it.
- **A fact edit that asserts `writer_block` UNCHANGED is CORRECT and is how the retired
  value keeps steering the writer.** `facts[]` is what may be stated; `writer_block` is
  the only text the page writer reads. `SQL_2026-08-19` and `SQL_2026-08-19g` both carry
  that guard, and both left the wire two days stale — one of them instructing a sentence
  the deploy gate now REFUSES (claimscan: `BANNED "ready the next day"` on the
  writer_block's own instruction). Fixed by `SQL_2026-08-21b`. **Add the agreement check,
  not just the unchanged check:** grep `writer_block` for the string your edit RETIRES.
  Full entry in `LANDMINES.md`. ⚠ And do NOT use "is this `writer_line` inside
  `writer_block`?" as a fleet drift sweep — measured, it convicts every healthy register
  on the estate (19/21 here, 35/35 on oufe.com); `writer_block` is prose and does not
  quote writer_lines. That mistake is logged in `WRONG_CALLS.md`.
- **When you write a replacement INSTRUCTION, do not put a live instance of a banned
  phrase in it.** Prompt text is read as an example. A first draft of the new turnaround
  rule was itself BANNED on "in one day" — the negation guard scans backwards a short
  way, so `Never` reached the first list item and not the third. Name the shape in prose
  the gate has no opinion about; the writer never needs the literal, because a refusal
  prints the ban's own reason.
- **A rewrite brief is scoped to what it NAMES.** A page can be simultaneously repaired
  and still wrong, and the item reads `complete` either way.
- **A `failed` work item is not work in flight**, but a `NOT EXISTS … status NOT IN
  (terminal)` guard counts it as one.
- **A `complete` work item is not a repaired artefact** — verify at the served page.
- **Stat/figure fields publish only attested numbers; hedges stay prose.**
- **`submission` embeds its own differing copies** of mission_brief and roadmap_brief;
  **`content_direction` carries a rendered `formatted` duplicate.** Fix one copy, the
  other stays stale and authoritative.

---

## 7. Falsifiers — check these before trusting anything above

- A newer handoff in either lane dir.
- §1's page checks, at the **served** pages, not the statuses.
- `sites.handed_over_at` existing (someone started Phase 4).
- The register's ban count (**34**) and fact count (**23** since 2026-08-21) — two lanes write this row, so compare against the row your transaction supersedes rather than trusting either number.
- Whether Stripe keys / the webhook exception / the second Nominet TAG have landed.
- The chassis tag (**v1.0.1320**); it rolls several times a day.
- **The council verdict on `905d9078-86c2-47a7-af0a-781723a46c08`** (DGH-014). Budget ~30
  minutes and find it by payload, not by the printed id:
  `SELECT current_step, status FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id' = '905d9078-86c2-47a7-af0a-781723a46c08';`
  A missing row is latency, not a dropped dispatch — do not resubmit on that evidence.
- The landmine verifier's verdict on the 2026-08-21 entry (`8a437a0c-41db-4b7b-ab14-1e96706482aa`), and on the two entries filed by this lane — the `(?i)`
  compile trap (`cf717466`) and the 7-day presign ceiling (`5c958a5f`):
  `SELECT subject_key, left(body,200), created_at FROM doc_notes WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT 3;`
