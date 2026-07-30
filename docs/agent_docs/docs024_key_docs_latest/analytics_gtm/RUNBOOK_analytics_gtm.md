# RUNBOOK — GTM on a chassis-built site

Commands that were hard to get right, with the gotcha attached. Worked example:
idea.uk, 2026-07-30 (`idea_uk_vm_site/sql/p4_34_gtm_container.sql`).

---

## 0. Before you touch anything — find out what you'd break

Chrome components are SHARED. Always resolve the blast radius first.

```sql
-- Which sites share this site's head/header component?
SELECT sc.slot_name, cc.name,
       (SELECT count(*) FROM site_components x WHERE x.component_id = sc.component_id) AS sites_using
FROM site_components sc JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = '<SITE_ID>' AND sc.slot_name IN ('head','header');
```

`sites_using > 1` ⇒ **never hardcode the container id in the template.** Gate it.

Fleet-wide map (14 deployed sites → 3 head, 6 header components as of 2026-07-30):

```sql
SELECT sc.slot_name, cc.name, count(*) AS sites, string_agg(s.domain, ', ' ORDER BY s.domain)
FROM sites s JOIN site_components sc ON sc.site_id=s.id AND sc.slot_name IN ('head','header')
JOIN content_components cc ON cc.id=sc.component_id
WHERE s.status='deployed' GROUP BY 1,2 ORDER BY 1, 3 DESC;
```

## 1. The per-site container id

`idx_site_specs_current` is UNIQUE on `(site_id, aspect) WHERE is_current`, so an
existing current row must be **superseded**, not duplicated:

```sql
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='<SITE_ID>' AND aspect='site_config' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by)
VALUES ('<SITE_ID>','site_config',
        '{"analytics":{"gtm_container_id":"GTM-XXXXXXX"}}'::jsonb,
        'operator','<who>');
```

`resolveConfigPath` searches aspects **`site_config`, `identity`, `design_intent`** in
that order — any of the three works; `site_config` is the honest home.

## 2. Template + schema (durable) — gated so other sites are untouched

Add to `html_template`, immediately after `<meta charset="UTF-8">` for the head and at
the very TOP for the header. Add to `input_schema` a **map-valued** key:

```sql
input_schema = COALESCE(input_schema,'{}'::jsonb) || jsonb_build_object(
  'gtm_container_id', jsonb_build_object(
     'type','string','source','config.analytics.gtm_container_id','required',false))
```

> ⚠ **A SCALAR value is silently ignored.** `render_site_components_action.go:612-615`
> skips any `input_schema` entry that is not a map — which is why `Document Head`'s
> existing `title`/`description` scalars never resolved. If your field does nothing,
> this is why.

> ⚠ **Gate with `{{if}}`.** An ungated `{{.gtm_container_id}}` renders empty on every
> site that has no id, and an empty `src=`/`href=` gets **dropped and filed as a dead
> control** (`DropDeadURLControls`, bugs_open/054).

## 3. The stored artefact (immediate) — or nothing changes

Pages assemble from `site_components.rendered_html`, NOT the template
(`bugs_open/117`). Write the resolved snippet there too:

```sql
UPDATE site_components SET rendered_html = replace(rendered_html,
  '<meta charset="UTF-8">', '<meta charset="UTF-8">' || E'\n' || '<GTM SCRIPT>')
WHERE site_id='<SITE_ID>' AND slot_name='head';

UPDATE site_components SET rendered_html = '<GTM NOSCRIPT>' || E'\n' || rendered_html
WHERE site_id='<SITE_ID>' AND slot_name='header';
```

Guard first — assert **exactly one** `<meta charset="UTF-8">` anchor and **zero**
existing `googletagmanager` matches, and `RAISE EXCEPTION` otherwise. `replace()`
replaces ALL occurrences; a second anchor would inject the tag twice.

## 4. Re-assemble and deploy the pages

Chrome is baked into deployed pages, so they must all be re-rendered.

```bash
./docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/scripts/fire_reassemble_idea_uk.sh
```

- **ASSEMBLE mode** (no `input_data.spec.reason`) is required. `section_data_resolved`
  bails at page level for pages with no stored components and **neither bail-out
  deploys**, so a chrome change silently misses them (`bugs_closed/031`).
- `page_id` is **required** — assemble mode fails "page_id not found" with only a name.
- ⚠ **Never publish via `kubectl run -i … kcat -P`** — it loses ~4 of 5 messages at
  exit 0 (kubectl attaches stdin asynchronously; kcat sees EOF first). Payload goes in
  the container **COMMAND**, `--command` is required (the image ENTRYPOINT *is* kcat),
  and every publish must print `PUBLISH_OK`.

## 5. Verify — at the artefact, then live. Never at the status.

`COMPLETED` is not proof (`016b`). Check what was actually rendered:

```sql
SELECT count(*) FILTER (WHERE (collected_data->'render_page'->>'html') LIKE '%gtm.js%')  AS script,
       count(*) FILTER (WHERE (collected_data->'render_page'->>'html') LIKE '%ns.html%') AS noscript,
       count(*) FILTER (WHERE (collected_data->'render_page'->>'html')
              LIKE '<!DOCTYPE html>%<body>' || chr(10) || '<!-- Google Tag Manager (noscript) -->%') AS adjacent
FROM orchestration_states WHERE correlation_id IN (…);
```

Then live (VM sitesync lag is ~5 min, but measured <1 min on 2026-07-30):

```bash
curl -s https://<domain>/<path> | grep -c googletagmanager   # expect 2
```

> ⚠ **Follow redirects, and check what you land on.** `/privacy.html` returns **301**
> to `/privacy`, which is served by a *different application* and was untagged. A
> naive `curl` without `-L` scores it as a pass-with-0-hits and a naive one with `-L`
> scores the redirect target as if it were your page. Both hide the same gap.

## 6. If the domain has a non-chassis application behind it

Grep nginx / the runbook for a **reserved-path set** before claiming "every page".
idea.uk proxies 16 routes to its own Go binary; 11 render HTML through one
`App.page()` wrapper, including the two conversion pages. Those need the tag in
**that** codebase — see `idea.uk/golang_files/gtm_test.go`.

Deploy of that binary (⚠ live payment service — take the backup):

```bash
GOOS=linux GOARCH=amd64 go build -C <golang_files> -o /tmp/idea ./...
ssh root@116.203.204.115 'cp /opt/idea/idea /opt/idea/idea.bak.$(date +%Y%m%d-%H%M%S)'
scp /tmp/idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'mv /opt/idea/idea.new /opt/idea/idea && chmod +x /opt/idea/idea && systemctl restart idea'
```

> ⚠ **`/etc/idea/idea.env`: systemd EnvironmentFile does NOT strip inline comments.**
> Every comment must be on its own line. `PORT=8080 # the port` makes the port
> `"8080 # the port"` → exit 1 → restart loop → nginx 502. This crashed a real deploy.
