
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

## DKIM/DMARC for contactforsales.com — where the records must land (2026-08-26)

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
