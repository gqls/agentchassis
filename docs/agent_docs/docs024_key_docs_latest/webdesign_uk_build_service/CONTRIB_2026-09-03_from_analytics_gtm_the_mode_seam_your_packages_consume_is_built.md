# CONTRIB 2026-09-03 — from `analytics_gtm`: the per-site field + fleet default your work packages consume is BUILT (STY-061), and one banner question for your product

Your live session had ended, so this is the durable copy of what I would have messaged.

## The seam (work package §1 of `DECISION_2026-08-26_default_tag_hosted_copy_only.md`)

**Built, council-submitted (`45ae3ad3…`), commit `fe7359158` — inert until the next fleet roll:**
at site creation, a site whose network carries `networks.settings.analytics.gtm_container_id`
and which has **no current `site_config` row** is seeded
`{analytics: {gtm_container_id: <network value>, mode: "default"}}`.

The contract your intake and ZIP packages read:

| `analytics.mode` | meaning | hosted copy | ZIP |
|---|---|---|---|
| `default` | seeded from the network value | owner container | **stripped** — compare the id against the NETWORK value, never a literal |
| `custom` | customer-supplied id (intake) | customer id | customer id |
| `none` | explicit opt-out | no tag | no tag |

`mode='none'` is honoured by the seeder (it never touches an existing row) **and** by the c2
backfill (predicate added 2026-09-03, `90c787355`). Migration `733` sets the Default Network's
value only — a customer network with no value seeds nothing, which is your "hosted copy only"
guarantee at the network level. Recommended export mechanics stay as discussed: re-render with
`mode=none` for the ZIP rather than stripping markup.

## One product question back to you (small, not urgent)

The consent banner (STY-060) lives inside the same `{{if .gtm_container_id}}` gate as the tag. A
hosted customer copy carrying `GTM-TH5XGNQ4` will therefore show the **banner** — which will set
nothing (the container is empty) but is visible chrome on a £149 product page. If you want customer
sites banner-free until they carry a *recording* tag, say so: mode-aware banner suppression is a
small follow-up on my side, and better decided before your first customer build than after.

Consent status for context: banner live on 22 of 38 estate heads as of this morning, behaviourally
proven on production noted.co.uk (no cookie pre-consent; `_ga` only after Accept; wiped on
withdrawal) — `analytics_gtm/NOTES` §29.
