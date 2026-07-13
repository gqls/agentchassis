# wayfaringlondoner.com — domain notes

Second probe domain. Sister docs: project plan/runbook/running-notes; the
relojistas notes are the template.

## Provenance
- **What it was:** a personal **travel blog**, "Wayfaring Londoner — doesn't
  bite", author **Csilla**, active ~2015–2016 (WordPress). Content: London life
  (Richmond Park deer, the London Cat Village/café, riverboats) and trips
  abroad (Bangkok temples/Grand Palace, Transylvania fortified churches, Jersey,
  a Downton-Abbey day). Tags seen: Travel, Castles, photography, Bangkok,
  Transylvania.
- **Evidence:** two Wayback snapshots in the project
  (`wayfaring_londoner___doesn_t_bite.html`, `…2.html`).
- **Status:** parked → probe candidate. Not yet provisioned.

## Decisions
- 2026-06-13 — **Framing:** English, `kind=search`, single text input — but a
  BLOG posture, not a marketplace one: the question asks for a destination /
  London spot / particular story ("What were you looking for?"), grounded in
  the real post topics. No category buttons (operator chose referer/query
  capture over the categories variant for this round).
- 2026-06-13 — **Thanks page = `/thanks.html`** (English), NOT gracias. This
  matters because `THANKS_PATH` is a single engine-wide env var: on the SHARED
  multi-vhost box every domain must use the SAME thanks filename. Standard for
  the shared box = `THANKS_PATH=/thanks.html`, and every domain ships its own
  `thanks.html`. (relojistas keeps `/gracias.html` because it's on its OWN box.)
- 2026-06-13 — **Deploy target:** the SHARED multi-vhost box (small domains
  share one box per the plan), not a dedicated VM — wayfaringlondoner is a
  low-traffic blog, no reason for its own box. Add it via setup.sh
  `DOMAINS="… wayfaringlondoner.com"` extend-and-re-run.

## Open choices
- Provision the shared box now, or wait until 2–3 small domains are ready to
  batch onto it? (One setup.sh run can take all their domains at once.)
- www handling: same CNAME approach as relojistas if inbound links used www.

## Passive signals (shared design, see project notes)
Inbound referer / landing query / 404 intent-paths / UA are captured by nginx
(combined log) for every domain on the box; harvested by the P4 collector via
log parse. The engine additionally attaches `landing_query` to submission
events. Nothing per-domain to configure for this.

## Log
- 2026-06-13 — File created; page built + validated
  (`wayfaringlondoner-site/index.html` + `thanks.html`), grounded in the
  snapshot. Awaiting a shared box + DNS.
