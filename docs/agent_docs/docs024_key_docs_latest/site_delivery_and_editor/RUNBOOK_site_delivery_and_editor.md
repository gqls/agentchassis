
## The delivery email's SMTP settings (owner-supplied 2026-08-26; password NEVER in a session or a file)

`platform/mailer` (`FromEnv`) — set as env/secret on whatever sends, when the send step is built:

| key | value |
|---|---|
| host | `mail.contactforsales.com` |
| port | `465` — implicit TLS; `mailer.UsesImplicitTLS("465")` already returns true, no code change |
| username / from | `webdesign@contactforsales.com` |
| password | secret ref ONLY (`secretKeyRef`), never a value; the owner holds it |

Gotchas measured 2026-08-26: the cert's CN is `cpanel.contactforsales.com` but the SAN
covers `mail.contactforsales.com`, so strict TLS verification passes — do not "fix" the
hostname to the CN. 465 = SSL/TLS from the first byte; 587 = STARTTLS — crossing the modes
is the classic silent client failure. Port 25 externally unanswered is NORMAL here (the
email-cluster PMG fronts inbound). DKIM + DMARC are absent as of 2026-08-26 and should be
enabled at the HOST (cPanel → Email Deliverability) before real customer mail, or
recipients will junk it — the domain's DNS is at uk-noc.com, NOT Cloudflare.

## DKIM/DMARC for contactforsales.com — ✅ RESOLVED 2026-08-26 night: BOTH RECORDS LIVE at the authoritative NS

> **VERIFIED 2026-08-26 (late)**, cache-bypassed — asked `dns1.uk-noc.com` directly:
> `default._domainkey` answers with the exact key cPanel generated (two-string chunking
> is normal for >255-char TXT), and `_dmarc` answers `v=DMARC1; p=none;`. The host's
> cluster evidently takes the cPanel zone, so route 1 below was never needed — the
> records were already there. **BOTH RESIDUALS CLOSED the same evening (owner, 18:52):**
> (a) the DMARC record now carries full reporting —
> `v=DMARC1;p=none;sp=none;adkim=r;aspf=r;pct=100;fo=0;rf=afrf;ri=86400;rua=mailto:webdesign@contactforsales.com;ruf=mailto:webdesign@contactforsales.com`
> — daily aggregate + forensic reports to the webdesign inbox; (b) **the artefact check
> PASSED**: a real test email verified at Gmail's own Authentication-Results —
> `dkim=pass` (selector `default`, our key), `spf=pass`, `dmarc=pass`. Delivered in 4s.
>
> One operational fact learned from the headers, worth keeping: **outbound mail relays
> through MailChannels** (rs17 → relay.mailchannels.net, 23.83.222.30) and SPF passes for
> that IP via the `include:relay.email-cluster.com` chain — so the sending IP customers
> see is MailChannels', not the box's, and the existing SPF already covers it. Nothing to
> change in DELIVERY_SMTP_*: we submit to mail.contactforsales.com:465 and the relay is
> the host's own plumbing. **Email infrastructure for the delivery email: FULLY CLOSED
> 2026-08-26.** Everything below kept as the recipe for the next domain.


**Do NOT change email host for this.** The server (rs17.uk-noc.com) is healthy end-to-end
(measured 2026-08-26: MX accepts, 465 AUTH over implicit TLS, cert valid, PTR valid) and
switching hosts would not help anyway: DKIM/DMARC are DNS records that must live at the
domain's AUTHORITATIVE nameservers — `dns1/dns2.uk-noc.com` — whoever the mailbox is with.
cPanel's warning means only "this box is not where DNS answers from".

The two records (values from cPanel's own generator, 2026-08-26):

| name | type | value |
|---|---|---|
| `default._domainkey.contactforsales.com` | TXT | `v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtG1wgGs4VquDQhTlNiyHpsfwWf4KkWqQKJyULVj9WIy2buLK18pqp2Hkr9LyfEGo/rUkztbH9BUfXIM309Y5eR/QmcPixEhBfYsm0ZcEx6UcsDLMbR45o3eAJxYGEXm+57mkAA4TyZc2AqJq4YkuBNnKLnnAVVEhujxZqt3OdBRcnVQ93YjxCuUEBrp+DylcWcszFHpgYmzwR5jXOhZHBR+AJ5fx3TCK9zu1X0HcQWpo6jFMMCy7EhRwWuGZMnvMQYaBLRQ4Sc9O6yJDEEukpN2bY7y8mweVDdaH/sqbiFzPjb7xCmraKsk3UTgqkECAWabOHyNJ5iwhNbyi2ncgfwIDAQAB;` |
| `_dmarc.contactforsales.com` | TXT | `v=DMARC1; p=none; rua=mailto:webdesign@contactforsales.com` (start at p=none = monitor only; tighten later) |

**Route, in order of cheapness:**
1. cPanel → Zone Editor → add both TXT records → wait ~30 min → verify from outside:
   `dig +short TXT default._domainkey.contactforsales.com` and `dig +short TXT _dmarc.contactforsales.com`.
   SPF and MX already sit correctly in the authoritative DNS, so the host's cluster has
   taken records for this zone before — the Zone Editor may well sync.
2. If nothing appears: ONE support ticket to the host (uk-noc) pasting the table above.
3. Permanent-control alternative: move the DOMAIN'S DNS (not the mailbox) to the estate's
   Cloudflare account — add the zone in the CF dashboard, CHECK the imported records
   include the MX trio + SPF exactly as-is, then change nameservers at the registrar.
   Email keeps flowing throughout (same MX values); we then own every future record.

## Delivery email: where customer_email comes from (corrected 2026-08-31, bugs_open/420)

```sql
SELECT direction->>'customer_email' FROM build_queue WHERE domain = '<domain>';
```
NEVER `sites.email` — since the 420 contract split that column is the PUBLISHED
contact only (empty unless the customer explicitly asked the site to show one).
The delivery-email-sender action never read the column in code (input_data only);
the sites.email wording in older recipes was convention, and it is now wrong.

## ⚠ BLOCKED until the 420 fix ROLLS: re-seeding boxingonline (d2aa5206)

`build_queue.direction.customer_email` still holds the payer address (durable
order record, correct) and `sites.email` is deliberately EMPTY — so the CURRENT
binary's fill-only-if-empty seed would REFILL the address onto the site on any
re-seed / canonical build retry (bugs_open/326 path). Every pre-check reads clean
because they all describe the state after the LAST seed. Verified live 2026-08-31
by two sessions independently. If a retry is unavoidable before the roll, run the
full 19-page served-sweep AFTERWARDS — the post-action check is the only one that
can catch it. Fix is committed (162877051), inert until the chassis roll.

### RE-SEED BLOCK LIFTED 2026-09-02 — the 420 fix is verified LIVE in the chassis

Verified at the binary (pods restarted 2026-09-01 21:00Z): `published_contact`
PRESENT, `email_was_intake_value` ABSENT (a REMOVED-string control — commit
162877051 deletes that log literal, so absence confirms the new code), invented
control absent. A re-seed of boxingonline no longer refills sites.email — the
intake email lands as delivery contact only, and the published contact requires
the explicit `direction.published_contact` key. The block above is historical.

### Hand-filing a `needs_page` rebuild (the shape that actually gets CLAIMED) — 2026-09-02

The claim gate is `claim_work_item_action.go:135`: `status IN ('triaged','approved')`
— an item filed at `pending` sits unclaimed FOR EVER with no error (LANDMINE
2026-09-02, "hand-filed work item … UNCLAIMABLE"). And `handler_agent` must be set
as a COLUMN (the `swi_no_handlerless_promotable` CHECK), not just as a spec key.

```sql
INSERT INTO site_work_items
  (site_id, item_type, priority, summary, source, created_by, spec, status, handler_agent)
VALUES (
  '<site_id>', 'needs_page', 10, '<summary>', 'operator', '<session>-session',
  '{"reason": "<reason>", "handler": "page-build-handler",
    "page_name": "<page>", "page_role": "<role>"}',
  'triaged', 'page-build-handler')
RETURNING id;
```

Gotchas attached: check the queue for open items on the page FIRST (CLAUDE.md
dispatch rule); `needs_page` resolves by `page_name` (unlike `page_rerender`,
which needs `page_id` in spec AND column — its own LANDMINE); claim latency when
correctly filed is ~2 min (measured twice 2026-09-02); ~300s no-dispatch after a
chassis pod restart. Worked example: item `7f1f4993` (guides-index, this dir's
NOTES sibling in webdesign lane, 2026-09-02 ~17:1xZ).


## boxingonline /index.html — BUILD-path rebuild for the card decks (prepared 21:28Z on 2026-09-02; NOT FIRED — owner's go)

**Why this and not a rerender:** on both `v1.0.1354` and `v1.0.1355` the RERENDER path does not
execute the fixed list-item producer (bugs_open/425 §2; measured again 21:22Z: rerender
`b238bed9` wrote, `articles[0]` still has no `excerpt` key). The BUILD path does (guides-index
`7f1f4993`, 17:23:02Z, same site). The components lane has asked for exactly this rebuild as the
experiment that breaks their ambiguity (their handoff §2 ⭐) and says it cannot damage the repro.

**⚠ What it changes beyond the cards:** the build path DELETE/re-INSERTs the page's components and
**regenerates the LLM-written fields** (webdesign NOTES 17:3xZ: "build regenerates; the stored-carry
expectation was the RE-RENDER path's behaviour") — so hero / featured / info-card-grid /
call-to-action copy on the home page is rewritten by the current planner, not carried. The owner
reviewed this page point by point (OWNER_REVIEW 08-31); a regeneration can re-introduce the classes
he flagged (meta-copy, AI-tell) on the page that passed. Previous render is archived in
`page_component_history` keyed on `page_id` (revert = manual restore, not a button). The 4 open
`empty_section:…:featured-content` items on this page (`66aab479`, `ea4de903`) may resolve or
re-file. Chrome will be current (GTM id now in `site_config`; 423 footer gate live), so the GTM
count on index moving after THIS is the rebuild, not the stale_chrome pass.

**Pre-flight (all three, every time):**
```sql
-- 1. queue: nothing in flight on the page
SELECT id,item_type,status,spec->>'reason' FROM site_work_items
 WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND status NOT IN ('complete','cancelled','rejected')
   AND item_type IN ('needs_page','page_rerender') AND (spec->>'page_name'='index' OR page_id='0ff07948-8e6f-477a-9069-452d1a2aecca');
-- 2. baseline you will compare against (record it)
SELECT id, (content_data->'articles'->0) ? 'excerpt' AS has_excerpt, updated_at FROM page_components
 WHERE page_id='0ff07948-8e6f-477a-9069-452d1a2aecca' AND content_data ? 'articles';
```
3. chassis pods older than 300 s (`kubectl -n ai-persona-system get pods -l app=agent-chassis`).

**The dispatch** — the exact shape of `7f1f4993` (which is the exact shape of `bccedf9c`), page
fields swapped. The two fields a hand-filed row silently needs are `pipeline='build'` and
`approval_mode='auto'` (build-dispatch-loop sets no pipeline default; a row without them sits
`triaged` for ever — the previous session's seven-minute trap). `handler_agent` must be set.
```sql
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, page_id, priority, handler_agent, status,
   created_by, approval_mode, pipeline, spec)
VALUES
  ('d2aa5206-73bc-4707-a69c-2702c1eb9152', 'operator', 'needs_page', 'medium',
   'Rebuild index (BUILD path) so content-listing re-resolves through the fixed producer: excerpt key + suffix-free titles on the home page cards (owner item 14; bugs_open/425 §2 experiment step 1 for the components lane)',
   '0ff07948-8e6f-477a-9069-452d1a2aecca', 10, 'page-build-handler', 'triaged',
   'site_delivery_and_editor-session', 'auto', 'build',
   '{"reason":"rebuild_cleared_component","handler":"page-build-handler","page_name":"index","page_role":"landing"}'::jsonb)
RETURNING id, created_at;
```
Expect claim within ~2–7 min (`claimed_by='build-dispatch-loop'`), complete ~2–8 min after
(`7f1f4993`: filed 17:14:58 → claimed 17:21:10 → complete 17:23:02). A missing claim at 10 min
= read the row, not re-file.

**Verify at the row, then the artefact:**
```sql
SELECT pc.id, pc.updated_at, jsonb_array_length(pc.content_data->'articles') AS n,
       (pc.content_data->'articles'->0) ? 'excerpt' AS has_excerpt,
       left(pc.content_data->'articles'->0->>'title',80) AS title0,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'article-card__excerpt','g')) AS decks_in_html
FROM page_components pc WHERE pc.page_id='0ff07948-8e6f-477a-9069-452d1a2aecca' AND pc.content_data ? 'articles';
-- the write: history keyed on page_id (component_id is NULL on 44,555/45,285 rows — never key on it)
SELECT count(*), min(created_at), max(created_at) FROM page_component_history
 WHERE page_id='0ff07948-8e6f-477a-9069-452d1a2aecca' AND created_at > now() - interval '15 minutes';
```
Success = `has_excerpt` true, `title0` suffix-free, `decks_in_html` > 0 (the 682 fingerprint
INVERSION — decks present is the success state on data-carrying items). Then the served page after
the next mirror tick (~:52): `curl -s https://boxingonline.ugg2.com/index.html | grep -c
article-card__excerpt` non-zero with `article-card` as the control; date the object against
`pages.deployed_at` before accepting lag. **Then tell the components lane the row id** — their step 2
is a `template_changed` rerender on top of the new baseline, and which way the key goes is their
discriminator. Also read the regenerated copy against OWNER_REVIEW's classes before calling item 14
done — the cards are not the only thing that changed.


## Chrome refresh by hand when the detector's `stale_chrome` item was born `unresolved` (2026-09-03)

The detector re-files `needs_rerender`/`stale_chrome` on every chrome-input drift, and the
loader parks the third one in 7 days as `unresolved` (terminal) — `bugs_open/451`. Check first:
```sql
SELECT id, status, left(summary,60), created_at FROM site_work_items
 WHERE site_id='<site>' AND item_key='stale_chrome' ORDER BY created_at DESC LIMIT 3;
```
Two terminals inside 7 days above an `unresolved` one = the ladder. Re-file exactly as
`bugs_open/451` §"Operator interim" (copy the last COMPLETED row's `spec`; `source='operator'`,
`status='triaged'`, `handler_agent='rerender-pages'`, `pipeline='build'`,
`approval_mode='auto'`, `priority 8`). boxingonline 2026-09-03: `ec92320f`. Verify at
`site_components` (`updated_at`, `rendered_html LIKE '%<gtm id>%'` and `'%cc_v1%'`), then at
the served pages after the mirror tick. ⚠ The wave re-assembles every page (`_assemble`
items, reason-less ⇒ `rerender_single_page`, stored arrays re-shipped) — it does NOT re-resolve
list items; still-suffixed cards after it are not a new failure.

## boxingonline /index.html call-to-action — "the calendar below" copy fix, framework route (prepared 2026-09-03 ~12:2xZ; NOT FIRED — gated on the components lane's batch 691 reading)

**The defect** (boxingonline session's read of the 12:09:49Z rebuild, credited): the repaired CTA's
subheadline reads *"New results, announcements and fight news land regularly, and the calendar below
tells you what's coming up next."* The CTA is the LAST section on the page; there is no calendar on
the home page (it is `/tools/fight-calendar/index.html`, which the secondary CTA already links). A
falsifiable spatial claim of the very class item 1 was about. Do NOT fix it by hand — the framework
writes the content (owner 08-06).

**The route, read at the live definitions:** `needs_copy_edit` → handler **`copy-editor`** (active,
registered) → `run_copy_edit` (LLM) → **`request_review` (`checkpoint_for_review`)** — the proposal
lands in the OWNER's review queue on admin.apis.uk (approve / request_changes, verified end-to-end
09-02) → on approval the edit is applied as a `section_edit` (`apply_section_edit`, `content_edit`,
`field_updates`, re-renders ONLY that slot from its template and re-assembles the page from stored
components; commits and deploys the page directly — no `page_rerender` item, no re-resolve of the
listing). ⚠ The existing deferred row `6776b944` on this page is an RFC_056 VERDICT row about the
OLD CTA (third-person tool voice) with a stale `current_value`/`acceptance_test` — its
`release_recipe` would dispatch a rewrite against copy that no longer exists. Leave it; file fresh.

**Pre-flight:** 691's reading in hand (key SURVIVES ⇒ go; key GONE ⇒ still go, because this route
does not re-resolve the listing — but say so explicitly and re-read the key after the apply);
no open `needs_copy_edit`/`section_edit` on the page; then:
```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, page_id, priority,
  handler_agent, pipeline, approval_mode, status, created_by, spec)
VALUES ('d2aa5206-73bc-4707-a69c-2702c1eb9152', 'operator', 'needs_copy_edit', 'medium',
  'index CTA subheadline claims "the calendar below" — the CTA is the last section and the calendar is a separate page (secondary CTA links it); rewrite so nothing on the page is described that is not on the page (owner item 1 class; boxingonline session read 2026-09-03)',
  '0ff07948-8e6f-477a-9069-452d1a2aecca', 20, 'copy-editor', 'build', 'auto', 'triaged',
  'site_delivery_and_editor-session',
  jsonb_build_object(
    'page_name','index', 'slot_name','call-to-action', 'component','call-to-action',
    'category','accuracy', 'origin','human_review', 'audit_source','site_delivery_and_editor',
    'description','The subheadline says "the calendar below tells you what''s coming up next". The call-to-action is the LAST section of the home page; nothing is below it, and the fight calendar is a separate page (/tools/fight-calendar/index.html) that the secondary button already links. The sentence describes the page''s own layout and gets it wrong.',
    'current_value','New results, announcements and fight news land regularly, and the calendar below tells you what''s coming up next.',
    'suggestion','Keep the headline and both buttons. Rewrite the subheadline (one sentence, second person, no more than ~120 characters) so it talks to the reader about what they get — results, announcements, fight news — and lets the "See the full fight calendar" button carry the calendar; do not describe where anything sits on the page.',
    'acceptance_test','The subheadline contains no spatial reference (below/above/here/this page/the list) to content that is not on the home page, keeps second-person address, and both CTA labels and URLs are unchanged.',
    'max_fix_attempts', 2)
) RETURNING id, created_at;
```
Then: watch the item → `checkpoint_for_review` → the owner's queue; after approval, re-read
`page_components` on index — content-listing still 6 excerpt keys, CTA subheadline changed — and the
served page after the mirror tick.

## The delivery recipient is REFUSED, not inherited — measured at the code 2026-09-03

`RFC_058` §5.5 records the delivery chain's source as *"convention, not code"*. Half of that is
now measurable and stronger than the RFC claims, and the distinction matters to the identity
decision the owner is being asked for:

- **What value is sent IS convention** — it comes from the step's `input_mapping`, and this lane's
  recipe points it at `build_queue.direction->>'customer_email'`.
- **But the refusal to guess is CODE.** `send_delivery_email_action.go:55` declares
  `Required: []string{"site_id","customer_email","live_site_url"}`, and `:99-102` does
  `customerEmail := strings.TrimSpace(inputs.Get("customer_email"))` then errors
  `customer_email resolved empty`. So the action cannot fall back to any other identity: an
  unnamed or empty recipient FAILS the delivery step rather than silently emailing whoever is
  nearest.
- **Nothing in the delivery path reads `sites.email`** `[MEASURED 2026-09-03, by column name and
  not by SQL verb — the literal-verb census misses runtime-assembled writers, see LANDMINES]`.
  The live `sites.email` readers are all PUBLISHED-contact uses: `rerender_page_sections_action.go:1096`
  (page content), `maintenance_actions.go:171` (a reviewed brief field), and
  `check_contact_form_undeliverable.go:254` (the form-action check).

**Consequence for RFC_058:** whichever identity wins, it must be NAMED in the 651 recipe — and
that is cheap to honour precisely because nothing currently inherits it and the action refuses an
empty value. The choice is the owner's; this measurement only says the plumbing will not quietly
make it for him.

## Verifying the links host WITHOUT a real token (added 2026-09-03)

The delivery email's two links are `https://links.webdesign.uk/c/<token>` (confirm the transfer) and
`/d/<token>` (download the zip). A previous reading of this lane's own handoff called the host
"unverified, and untestable until a token exists". **Both halves were wrong**, and the reason is
worth more than the recipe: the probe used `/` and an invented top-level path, **neither of which is
one of the routes.**

```bash
# Probe the ROUTES, with a deliberately invented token, plus a control that must 404.
for p in "/c/definitely-not-a-real-token-xyz123" \
         "/d/definitely-not-a-real-token-xyz123" \
         "/zzz-invented-control-path"; do
  printf "%-42s " "$p"
  curl -s -o /dev/null -w "http=%{http_code} size=%{size_download}\n" --max-time 15 \
       "https://links.webdesign.uk$p"
done
```

Expected, and what it means:

| path | expect | why |
|---|---|---|
| `/d/<invented>` | **200**, "That download link is no longer active" | the correct refusal for an unknown token — one message for unknown/expired/revoked/spent, deliberately |
| `/c/<invented>` | **200**, the confirm button page | **by design, owner-ruled 2026-08-25.** `HandleConfirmPage` does NO database access (`internal/core-manager/handlers/delivery.go:145`); a read-only lookup would be a free validity oracle for a guessed token. The customer learns on pressing |
| `/zzz-invented-control-path` | **404** (nginx's own) | the control. Without it, two 200s prove nothing — a catch-all would give the same |

⚠ **`/c/` serving a live-looking button for a dead token is NOT a bug.** It reads like one from the
probe alone, and the handler's own comment block says why it is not. Read the handler before filing.

⚠ **Use GET, not `curl -I`.** HEAD on `/c/` is refused (`renderSpeculativeRefusal`) and answers 404 —
so a HEAD probe reports the route as missing when it is live.

## The delivery rehearsal on a NON-customer site (started 2026-09-03, owner-authorised)

Authorised as a **full** rehearsal — the whole chain including the email — on a site we own, in the
knowledge that it burns that site's once-only handover stamp and sends a real message through the
live SMTP account.

**Site chosen: `idea.uk`** (`1244516d-014d-421c-88c6-090bb1e9552a`). Ours, 38 pages, never handed
over, not the `webdesign.uk` shopfront and not a lane another session is actively working.

**What burning the stamp actually costs, checked before firing rather than assumed.** `handed_over_at`
is read in exactly one file — `platform/delivery/handover.go` — and its own comment states the scope:
*"IsHandedOver is the gate Phase 5's editor session exchange uses. It is the ONLY thing handover
gates: not deploys, not rewrites, not locks, not reconciliation."* So the cost is: this site can
never be delivered through the chain again, and its editor-session gate flips open. Nothing about how
it is built or served changes. `[MEASURED 2026-09-03]` **0 of 60** sites are stamped, so nothing on
the fleet depends on the column today either.

### ⚠ AN OPERATOR CANNOT RUN THIS CHAIN ALONE, AND THAT IS THE DESIGN

Step 2 is the owner's APPROVE button and there is no operator path around it. `POST
/api/v1/admin/work-items/:item_id/approve` sits behind `AuthMiddleware` + `AdminOnly()`
(`internal/core-manager/api/server.go:338,364`), and `HandleApproveWorkItem` is the **only** writer of
`result.approved_by`, which is the entire predicate `platform/delivery.Reviewed()` gates on. There is
no seed, no dispatch and no SQL shortcut that is not forgery. **Do not invent one** — plan the
rehearsal around a real click, and expect to wait for it.

### The four steps, in order

```bash
. "$PWD/scripts/kafka-publish-lib.sh"     # OPP-009: publishes with an ASSERTED receipt
SITE=1244516d-014d-421c-88c6-090bb1e9552a
DOMAIN=idea.uk
```

**1 — file the review.** ✅ **DONE 2026-09-03 18:22:03Z** — item `e370e0bb`, `status =
needs_human_review`, `spec.checkpoint = true`, `item_key = delivery_review_idea.uk`. First time
`delivery-review-filer` has ever run on any site.

> ⚠ **THE FIRST ATTEMPT WAS REFUSED, AND THE REFUSAL LOOKED EXACTLY LIKE LATENCY.** I sent the three
> headers `651`'s header then listed (`message_type`, `action`, `from_agent_type`). `client_id` and
> `orchestration_id` are ALSO REQUIRED, and without them the message is consumed and rejected —
> leaving no orchestration row, which is the exact symptom CLAUDE.md tells you not to retry on. I
> waited 37 minutes on a message that had died in seconds. **Run `kafka_verify_landing` at the FIRST
> check, not as an escalation** — it names the missing headers. `651`'s recipe is corrected, and the
> trap is in `LANDMINES.md`. A green `kafka_publish_checked` receipt asserts PUBLICATION, not
> acceptance; only the second predicts a run.

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(jq -c -n --arg s "$SITE" --arg d "$DOMAIN" --arg b "<one paragraph>" '{action:"orchestrate",
  config:{agent_type:"delivery-review-filer"},
  input_data:{site_id:$s,domain:$d,site_url:("https://"+$d),brief:$b,
    # review_data is what the OWNER is shown on the approve screen, and it is
    # composed HERE on purpose (migration 752, bugs_open/474). Omit it and
    # create_work_item REFUSES the item outright — a loud failure in front of
    # you now, instead of a silent one in front of him later.
    review_data:{domain:$d, site_url:("https://"+$d), brief:$b}}}')
kafka_publish_checked --topic system.agent.generic.requests --payload "$PAYLOAD" \
  --correlation "$CORR" \
  --header "orchestration_id=$ORCH" \
  --header "orchestration_name=delivery-review-filer-$(date +%H%M%S)" \
  --header "client_id=demo_client" \
  --header "request_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "message_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "step_name=start" \
  --header "message_type=request" --header "action=orchestrate" \
  --header "from_agent_type=user" --header "from_agent_id=cli" \
  --header "responses_topic=system.agent.generic.responses"
# 0 = published + receipt seen · 10 = never published, retry now · 11 = indeterminate, verify

kafka_verify_landing "$CORR" 30     # <- ALWAYS, and FIRST
# 0 = landed · 13 = published not landed (this really is latency, wait)
# 12 = CONSUMED AND REFUSED — it will never run; the error names what is missing
```

**Do not budget the 29-minute latency figure for this dispatch.** That measurement (2026-07-20) was a
council dispatch under fleet load. This one landed and COMPLETED in **under 25 seconds**. A slow
dispatch is a reason to verify, not a reason to assume you are inside a known-good window.

**2 — THE OWNER APPROVES** the `needs_delivery_review` item on admin.apis.uk. Not resolve — **approve**;
resolve writes `resolved_by`, which the gate deliberately ignores. Confirm it landed before step 3:

```sql
SELECT status, result ? 'approved_by' AS gate_open, result->>'approved_by'
  FROM site_work_items WHERE item_type='needs_delivery_review' AND site_id='<SITE>';
-- gate_open must be TRUE. That key, not the status, is what delivery.Reviewed() reads.
```

**3 — cut the zip:**

```bash
CORR=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(jq -c -n --arg d "$DOMAIN" '{action:"orchestrate",
  config:{agent_type:"zip-deliverable-dispatch"},input_data:{domain:$d}}')
kafka_publish_checked --topic system.agent.generic.requests --payload "$PAYLOAD" \
  --correlation "$CORR" --header message_type=request --header action=orchestrate \
  --header from_agent_type=user
```

Note `presigned_url` and `expiry_minutes` from its output — step 4 needs both.

**4 — send it.** ⚠ `customer_email` comes from `input_data` and **nowhere else** — the action never
reads `sites.email`, and since the 420 contract split that column is the PUBLISHED contact, legitimately
NULL on a post-420 site. `idea.uk` has no `build_queue.direction->>'customer_email'` either
(`[MEASURED 2026-09-03]`), so for the rehearsal it must be an address we control and it must be typed
in deliberately:

```bash
CORR=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(jq -c -n --arg s "$SITE" --arg d "$DOMAIN" --arg e "<an address we control>" \
  --arg z "<presigned url from step 3>" --arg m "<expiry minutes from step 3>" '{action:"orchestrate",
  config:{agent_type:"delivery-email-sender"},
  input_data:{site_id:$s,customer_email:$e,live_site_url:("https://"+$d),
              zip_presigned_url:$z,zip_presign_minutes:$m}}')
```

**This step is irreversible.** It stamps the handover once and only once, mints the customer tokens,
and sends. A second dispatch for the same site is REFUSED by the stamp — that refusal is the
double-send guard working, not a fault. Recovery for stamped-but-unemailed is in
`sql_for_agents/651_delivery_review_and_email_agents_HOLD.sql`'s header.

### Prerequisites, all re-checked 2026-09-03

| thing | state |
|---|---|
| migration 650 (`customer_access_tokens.stored_url`) | applied — both columns present |
| `DELIVERY_SMTP_*` on `agent-chassis` | present; `DELIVERY_SMTP_PASS` → secret `delivery-smtp-secrets`, which **exists** |
| ⚠ that secretKeyRef | `optional: true` — a missing secret is silently EMPTY, not a start failure. The action builds the sender before the claim, so it fails loudly with nothing stamped, but do not read "the pod started" as "the password is there" |
| the four agents | all `is_active`; `zip-deliverable-dispatch` and `zip-deliverer` are `status: experimental` |
| `links.webdesign.uk` | verified — see the section above |
| DKIM/DMARC at the mail host | **still unchecked from here** (absent as of 2026-08-26). This is the one that decides whether the message lands in an inbox or a junk folder |

### ✅ THE REHEARSAL RAN END TO END — 2026-09-03, idea.uk. All four agents, first time ever.

| step | agent | result |
|---|---|---|
| 1 | `delivery-review-filer` | item `e370e0bb` filed 18:22:03Z (after a refused first attempt — §the envelope) |
| 2 | **the owner's APPROVE** | 19:20:58Z. `status=complete`, `result.approved_by='admin'`, `resolved_by` absent — the gate's own predicate, read back independently |
| 3 | `zip-deliverable-dispatch` → `zip-deliverer` | 19:23:14Z. **45 files**, 2,394,857 B zipped from 2,846,881 B, `deliverables/idea.uk/idea.uk-af9039a61dbd.zip`, presign **10080 min (7 days)** |
| 4 | `delivery-email-sender` | 19:30:31Z. `send_email => {"to":"aaa@…","sent":true,"zip_link":true,"advertised_days":30}` |

**The first handover stamp in the estate's history**: `sites.handed_over_at = 2026-09-03 19:30:31Z`,
`live_link_expires_at = 2026-10-15` (six weeks). Was 0 of 60 sites all day. **Two tokens minted**
(`confirm_transfer`, `zip_download`), both to 2026-10-15, the zip one carrying a `stored_url` — the
first `customer_access_tokens` rows ever.

**The zip link was proved live BEFORE the email went, with two negative controls** — because a 200
on its own cannot tell a working signature from an open bucket:

```
ranged GET (Range: bytes=0-0)          -> 206, application/zip, 1 byte
same URL, last signature char altered  -> 403     <- the signature is doing work
the object path with no signature      -> 401     <- the bucket is not open
```

`https://idea.uk` also probed 200 with an invented path 404ing, so the live-site link in the email is
not a parked catch-all.

### ⚠ THE STANDING FINDING: three different lifetimes, and the longest one rests on a mechanism that has never run

| what | lifetime |
|---|---|
| the presigned S3 URL inside the zip link | **7 days** |
| what the email tells the customer | **30 days** |
| the tokens themselves | **42 days** (to 2026-10-15) |

This is **not** a broken link at day 8, and the code is careful about exactly that: `HandleZipDownload`
refuses to redirect to a stale presign — *"an expired presign answers 403 SignatureDoesNotMatch, which
reads as broken credentials, not an old link"* — and renders a refresh page instead
(`ErrZipURLStale` → `RecordStaleZipLink`). The 30-day promise is therefore kept by
`zip-link-refresher` re-signing `stored_url` before the presign lapses.

**That refresher is scheduled (`scheduled_tasks.zip-link-refresh`, enabled, every 21600s, last
triggered 2026-09-03 15:04:32Z) and has never refreshed anything, because until 19:30 today there
were ZERO `zip_download` tokens fleet-wide for it to act on.** `enabled` plus a fresh tick is not
evidence of work — the scheduler firing an agent that finds nothing to do looks identical to one
doing its job.

> ## ✅ TESTED AND IT WORKS — 2026-09-03 21:07Z. The owner designed the test; it took four minutes.
>
> Rather than wait a week for the presign to approach its lapse, he proposed shortening a link and
> watching. **The non-destructive form: move OUR RECORD of the expiry, not the signature.** The
> refresher's `pre_query` selects on `stored_url_expires_at < now() + interval '48 hours'`, so setting
> that column to `now() + 1 hour` makes the row selectable while the real S3 signature keeps its full
> seven days — the customer's link never stops working during the test.
>
> | | before | after |
> |---|---|---|
> | `stored_url` fingerprint | `2496cd49cf1f` | **`a38face0a118`** |
> | `stored_url_expires_at` | 21:56:06Z (the 1-hour test value) | **2026-09-10 21:08:16Z** |
>
> Orchestration `4d543a40` ran 21:07:40 → 21:08:18Z, COMPLETED, no error. The new URL was verified at
> the artefact, not at the fingerprint: ranged GET **206 `application/zip`**, tampered signature
> **403**, and it is a genuinely different signature from the original rather than a re-stored copy.
>
> **So the email's 30-day promise is kept by something that actually runs.** The three lifetimes still
> differ, and that is now by design rather than by neglect: the presign is short, the refresher keeps
> it ahead of use, and the token is the real limit.
>
> ⚠ **MY DETECTOR REPORTED THE OPPOSITE, and it was a race.** It polled tick-advanced against
> value-unchanged at one instant. The refresher takes 38 seconds; the poll landed inside that window
> and declared "FIRED AND DID NOTHING", which is the exact failure it was built to catch. **Do not
> gate this check on the trigger's timestamp.** Wait for the orchestration to reach a terminal state,
> then compare. Full write-up in `WRONG_CALLS.md` 2026-09-03(c).
>
> **The re-check, if you need it again** (the shape above is reusable; snapshot first, the row is a
> live customer token):
> ```sql
> SELECT purpose, stored_url_expires_at, expires_at
>   FROM customer_access_tokens WHERE purpose='zip_download';
> ```
> `stored_url_expires_at` advancing past 2026-09-10 is the proof. If it does not move, the email's
> 30-day promise is good for 7, and the customer meets a refresh page — which is the honest failure,
> but still a failure of the thing we told them.
