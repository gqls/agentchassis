# RUNBOOK — static site form endpoint

Every command this lane had to get right, with its gotcha attached. Change a command **here**, not
in your scrollback.

---

## 1. Census: what the fleet's forms actually do

**⚠ The gotcha that cost this lane its opening premise: `content_data` is not what the visitor
gets.** The render seam rewrites `form_action` on the way out (`deliverableFormAction`,
`component_library.go:1495`). A census on `content_data` alone reports every form as dead and
**cannot come out any other way**, because the repair happens below the layer you queried. Always
run both halves, and quote the second.

```sql
-- (a) the authored value — what the content LLM wrote
SELECT content_data->>'form_action' AS action, count(*)
FROM page_components WHERE content_data ? 'form_action' GROUP BY 1 ORDER BY 2 DESC;

-- (b) the SERVED value — what the visitor's browser receives.  QUOTE THIS ONE.
SELECT COALESCE(substring(rendered_html from 'action="([^"]*)"'), '(no action attr)') AS served,
       count(*) AS components, count(DISTINCT p.site_id) AS sites
FROM page_components pc JOIN pages p ON pc.page_id = p.id
WHERE pc.rendered_html ~* 'data-component="contact-form"'
GROUP BY 1 ORDER BY 2 DESC;
```

Wider shape census — **not** restricted to `data-component="contact-form"`, which is how you see
the class the discovery check deliberately excludes:

```sql
SELECT CASE
         WHEN rendered_html ~* 'action="mailto:' THEN 'mailto:'
         WHEN rendered_html ~* 'action="#'       THEN 'self-anchor #'
         WHEN rendered_html ~* 'action="https?:' THEN 'absolute http(s)'
         WHEN rendered_html ~* '<form[^>]*action="' THEN 'other action value'
         ELSE 'form with NO action attr' END AS shape,
       count(*) AS components, count(DISTINCT p.site_id) AS sites
FROM page_components pc JOIN pages p ON pc.page_id = p.id
WHERE pc.rendered_html ~* '<form' GROUP BY 1 ORDER BY 2 DESC;
```

**Date every number you take from these.** A form census goes stale by *addition* and reads as
current for ever: the authored-value count went 22 → 27 in the twenty-four hours between the
pre-plan and this lane picking it up.

## 2. Is a form's action target actually answered?

**The gotcha: a plausible-looking action is not a live one, and no list can tell you which.** The
detector matches literals (`#`, `#contact`, `/contact`, empty, absent); anything else passes
unprobed. Ask the network instead — and never without both controls, because a parked or
catch-all domain 200s every path and would make the answer meaningless.

```bash
probe(){ printf '  %-40s -> %s\n' "$2" \
  "$(curl -sS -o /dev/null -w '%{http_code}' --max-time 12 "https://$1$2" 2>&1 | tail -1)"; }
TS=$(date +%s)
probe <domain> /<a-known-good-deployed-page>   # control B: MUST be 200
probe <domain> /invented-$TS                   # control A: MUST be non-200
probe <domain> /<the-form-action>              # the target
```

**Reading it:** `405` = a real POST-only handler (the method is wrong, the route exists).
`404` **with both controls holding** = nothing answers; the form is dead. Verified discriminating
on 2026-09-04: `idea.uk/request` 405, `relojistas.com/intent` 405, `gamesdesign.co.uk/request`
404. Without control B a 404 could just mean the site is down; without control A it could mean
nothing at all.

For a *page* rather than a form action, use the estate's script instead of hand-composing a URL —
four sessions have filed a 404 they created themselves by guessing the URL form:

```bash
./scripts/probe-page-url.sh <domain> <page-name>...   # reads pages.url, never composes it
./scripts/probe-page-url.sh <domain> --all
```

## 3. Probing the live receiver

**The gotcha: `/health` returns 404 and the service is fine.** The island Caddy forwards
`/api/v1/tools/*` and 404s everything else, so a health check outside that prefix tells you
nothing. Use a registered route's `OPTIONS` as the positive control — the CORS preflight is
answered by the app, so a 204 proves the whole chain (tunnel → Caddy → gin) is up.

```bash
curl -sS -o /dev/null -w '%{http_code}\n' -X OPTIONS \
  -H 'Origin: https://vonc.com' https://tools.apis.uk/api/v1/tools/gauntlet/round   # expect 204
curl -sS -o /dev/null -w '%{http_code}\n' -X OPTIONS \
  https://tools.apis.uk/api/v1/tools/gauntlet/nope_$(date +%s)                      # expect 404
```

**Do not POST to `/round`, `/position`, `/defend` or `/gripper/submit` as a liveness check** —
they create rows and, for the gripper, send email. `OPTIONS` creates nothing.

## 4. Where the pieces live

| what | where |
|---|---|
| the render seam that decides a form's destination | `platform/orchestration/actions/component_library.go` — `sanitiseFormAction` :1466, `sanitiseFormActionStrings` :1483, **`deliverableFormAction` :1495**, `nonDeliveringFormActions` :1448 |
| its tests | `platform/orchestration/actions/component_library_form_action_test.go` |
| the detector | `platform/orchestration/actions/discovery_checks/check_contact_form_undeliverable.go` |
| the prior-art intake handler | `internal/tools-api/handlers/gripper.go:227` `GripperSubmitHandler` |
| site identity from the request | `internal/tools-api/middleware/cors.go:17` → `internal/tools-api/store/sites.go:23` `ActiveSiteByOrigin` |
| abuse gates | `platform/httpguard/{intake,limiter,clientip}.go` |
| email | `platform/mailer/mailer.go`; config via `mailer.FromEnv("<PREFIX>")`, see `internal/tools-api/config/config.go:66` |
| route registration | `internal/tools-api/api/server.go` (`NewRouter`, `RouterOption`) and `cmd/tools-api/main.go` |
| the public edge | `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/infra/island/Caddyfile` |

## 5. Which sites can be reached, and how they publish

```sql
-- publish mode. NOTE: 34 of 39 live sites have BOTH columns NULL — publish_target is
-- itself an opt-in seam (default OFF), so "NULL" is the common case, not an anomaly.
SELECT COALESCE(publish_target,'-') tgt, COALESCE(github_repo,'-') repo, count(*)
FROM sites WHERE status IN ('active','deployed') GROUP BY 1,2 ORDER BY 3 DESC;

-- liveness predicate: use IN ('active','deployed'), NOT status='deployed' alone.
-- (744 / CLM-033, 2026-09-03.) Today 39 deployed / 0 active, so the narrow form is
-- latently wrong rather than visibly wrong — which is why it survives unnoticed.
SELECT status, count(*) FROM sites GROUP BY 1 ORDER BY 2 DESC;
```

## 6. DB access

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

A fleet-wide `Unauthorized` means the kubeconfig token expired (every 3 days) — the owner
refreshes it; it is not a cluster fault.
