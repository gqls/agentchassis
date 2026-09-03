
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
