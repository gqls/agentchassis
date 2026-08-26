
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
