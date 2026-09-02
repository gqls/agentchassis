
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
