# Evidence — who actually requests the legacy board feeds (25 July 2026)

Gathered to decide whether to build **per-forumid category feeds** (the last
unbuilt item on the relojistas rebuild list). The answer changed the decision:
see `relojistas_rebuild_plan.md` §P8. This file preserves the raw evidence so
nobody has to re-derive it — the board-name recovery in particular took a
Wayback fetch that is slow and may not stay available.

## 1. Method (reproducible)

Board names — the old vBulletin index, fetched raw from the Wayback Machine:

```bash
curl -s "http://web.archive.org/cdx/search/cdx?url=relojistas.com/forum.php\
&output=text&fl=timestamp,statuscode&filter=statuscode:200&collapse=timestamp:6&limit=20"
curl -s -A "Mozilla/5.0" \
  "http://web.archive.org/web/20150925002629id_/http://www.relojistas.com/forum.php" -o forum_2015.html
# TRAP: the page is ISO-8859-1. grep treats it as binary in a UTF-8 locale and
# silently matches NOTHING. Use LC_ALL=C grep -a, and iconv before reading names.
iconv -f ISO-8859-1 -t UTF-8 forum_2015.html > forum_2015.utf8.html
grep -oE 'forums/[0-9]+-[^"?]+' forum_2015.utf8.html | sort -u
# ids missing from the index page (44, 288) resolve individually via CDX:
curl -s "http://web.archive.org/cdx/search/cdx?url=relojistas.com/forums/44-*&output=text&fl=original&limit=3"
```

Demand — the live access log, scoped to the feed endpoint (see §5 trap):

```bash
ssh root@167.233.33.159 'zcat -f /var/log/nginx/access.log* \
  | grep -oE "GET /[^ ]*external\.php\?[^ ]*" \
  | grep -oE "forumids=[0-9]+" | cut -d= -f2 | sort | uniq -c | sort -rn'
```

## 2. The decisive finding

**Board-param feed requests are crawler traffic. Real feed clients ask for the
bare feed.** Post-fix window (18–25 Jul), 793 board-param requests:

| client | requests | class |
|---|---|---|
| meta-webindexer (Facebook) | 349 | crawler |
| Googlebot | 153 | crawler |
| Chrome/30.0.1599.66 (2013 UA) | 160 | scraper spoofing an ancient browser |
| SERankingBacklinksBot | 22 | crawler |
| DotBot (Moz) | 17 | crawler |
| modern desktop/mobile browser UAs | ~90 across ~10 UAs | **unattributable** — see §4 |

~700 of 793 (88%) are self-identified crawlers. **Zero conditional GETs (304) on
board params.** By contrast the bare feed shows 42 × 304 from a single
Chrome/108 client — textbook feed-reader polling with `If-Modified-Since`, and
the strongest available evidence of a genuine surviving subscription.

Bare-feed clients (all history) include `FeedFetcher-Google` (138),
`Apache-HttpClient/4.5.5` (160), empty UA (306) and a spread of current
Safari/Firefox — i.e. real subscription plumbing, none of it board-scoped.

## 3. The corpus cannot serve most of those boards anyway

Matching the 75-item curated corpus against each board's subject:

| board | matching items |
|---|---|
| 4 — Marcas de relojes (brand news) | 32 |
| 44 — Seiko | 1 |
| 288 — Louis Erard | 0 |
| 78 — Auténtico/Falso | 0 |
| 145 — Relojes perdidos y robados | 0 |
| 13 — Sorteos y concursos | 0 |

Most requested boards are forum-social, not news-shaped: legal advice, group
buys, member introductions, photography, off-topic, voting. A news portal has
nothing to put in them — a per-board feed would be an empty feed, which reads
to a subscriber exactly like the dead site we are undoing.

## 4. What this evidence CANNOT settle [MARKED LIMITATION]

Every client IP in the log is a Cloudflare edge address (104.22.x, 172.71.x),
because `real_ip_header CF-Connecting-IP` is not yet configured — the P0 item in
the pending owner box run. So distinct *people* cannot be counted, and the ~90
modern-browser board requests cannot be attributed to humans or to spoofing
scrapers. This is the measured justification for the CF real-ip change, not a
preference.

## 5. Trap that produced a wrong intermediate reading

Grepping `forumids=` across the whole access log (rather than scoping to
`external.php` first) also sweeps crawlers walking *old forum thread URLs* that
carry the same parameter. It inflates the picture and reads like subscriber
demand. Scope to the endpoint first. Both numbers happened to be close here
(123 boards either way), which is exactly why it would have gone unnoticed.

## 6. Full demand table — 123 boards, feed endpoint only

Requests are the whole retained log window (June–25 July 2026), which spans the
pre-fix period when every one of these returned 404.

| forumid | feed requests | board name (Wayback 2015 index) |
|---|---|---|
| 41 | 148 | Panerai |
| 72 | 90 | Consultorio-jurídico-denuncias-y-quejas |
| 13 | 80 | Sorteos-y-concursos |
| 337 | 75 | Revisiones-y-comparativas |
| 28 | 73 | Vintages-y-otros |
| 78 | 71 | Auténtico-Falso |
| 4 | 70 | Marcas-de-relojes |
| 12 | 68 | Compras-conjuntas |
| 353 | 65 | Relojes-de-piloto-Pilot-Flieger |
| 75 | 63 | Técnica-y-reparaciones |
| 142 | 62 | Presentacion-de-nuevos-foreros |
| 216 | 62 | Relojes-de-buzo-Divers |
| 352 | 61 | English-català-euskera-galego |
| 315 | 60 | Seiko-y-otros-japoneses |
| 16 | 59 | Fotografía |
| 193 | 59 | Colaboradores |
| 328 | 59 | Consultorio-pre-venta-y-post-venta |
| 2 | 58 | Foro-general |
| 319 | 58 | A-todo-gas |
| 48 | 58 | Presentación-de-relojes |
| 15 | 55 | Beber-y-vivir |
| 309 | 55 | Prometheus |
| 21 | 54 | Cajón-de-sastre |
| 58 | 52 | Área-de-votaciones |
| 348 | 44 | Nuevos-Articulos-en-Venta |
| 51 | 44 | Tarifas-de-precios |
| 29 | 42 | Breitling |
| 103 | 41 | Ferias-de-relojería |
| 145 | 36 | Relojes-perdidos-y-robados |
| 43 | 34 | Rolex |
| 241 | 31 | Proyecto-S80-RLJ01 |
| 253 | 30 | Hublot |
| 332 | 29 | Año-2013 |
| 159 | 28 | Jaeger-LeCoultre |
| 336 | 27 | Fútbol-con-reservas |
| 238 | 26 | Año-2010 |
| 330 | 25 | Proyecto-Bauhaus-RLJ02 |
| 36 | 24 | Hamilton |
| 39 | 24 | Omega |
| 47 | 24 | Otras-marcas |
| 239 | 23 | Año-2009 |
| 244 | 23 | Temario-de-Richard-Samper |
| 237 | 22 | Año-2011 |
| 214 | 21 | (no name in 2015 index) |
| 269 | 21 | (no name in 2015 index) |
| 240 | 19 | Año-2008-y-anteriores |
| 284 | 19 | (no name in 2015 index) |
| 248 | 18 | Año-2012 |
| 8 | 18 | Off-topic |
| 311 | 17 | Mondani-Editore |
| 335 | 17 | Certina |
| 351 | 17 | Año-2014 |
| 354 | 17 | Año-2015 |
| 212 | 15 | Baselworld-2011 |
| 250 | 15 | Baselworld-2012 |
| 299 | 15 | (no name in 2015 index) |
| 1 | 14 | General |
| 20 | 14 | (no name in 2015 index) |
| 292 | 14 | (no name in 2015 index) |
| 293 | 14 | (no name in 2015 index) |
| 298 | 14 | (no name in 2015 index) |
| 77 | 14 | English-corner-(Foro-en-inglés) |
| 272 | 13 | (no name in 2015 index) |
| 283 | 13 | (no name in 2015 index) |
| 3 | 13 | Relojes |
| 257 | 12 | (no name in 2015 index) |
| 287 | 12 | (no name in 2015 index) |
| 325 | 12 | Baselworld-2013 |
| 19 | 11 | (no name in 2015 index) |
| 251 | 11 | SIHH-2012 |
| 255 | 11 | Outlet |
| 276 | 11 | (no name in 2015 index) |
| 288 | 11 | (no name in 2015 index) |
| 291 | 11 | (no name in 2015 index) |
| 346 | 11 | SIHH-2014 |
| 347 | 11 | Baselworld-2014 |
| 98 | 11 | (no name in 2015 index) |
| 268 | 10 | (no name in 2015 index) |
| 270 | 10 | (no name in 2015 index) |
| 282 | 10 | (no name in 2015 index) |
| 5 | 10 | Club-social |
| 104 | 9 | Baselworld-2010 |
| 266 | 9 | (no name in 2015 index) |
| 294 | 9 | (no name in 2015 index) |
| 355 | 9 | (no name in 2015 index) |
| 308 | 8 | Foros-oficiales |
| 320 | 8 | Racó-català-(Foro-en-catalán) |
| 339 | 8 | (no name in 2015 index) |
| 17 | 7 | Centro-comercial |
| 277 | 7 | (no name in 2015 index) |
| 295 | 7 | (no name in 2015 index) |
| 321 | 7 | Euskal-txokoa-(Foro-en-vascuence) |
| 50 | 7 | (no name in 2015 index) |
| 201 | 6 | (no name in 2015 index) |
| 338 | 6 | (no name in 2015 index) |
| 197 | 5 | Máquinas-del-Tiempo |
| 18 | 4 | (no name in 2015 index) |
| 256 | 4 | Profesionales |
| 263 | 4 | (no name in 2015 index) |
| 264 | 4 | (no name in 2015 index) |
| 267 | 4 | (no name in 2015 index) |
| 274 | 4 | (no name in 2015 index) |
| 278 | 4 | (no name in 2015 index) |
| 280 | 4 | (no name in 2015 index) |
| 285 | 4 | (no name in 2015 index) |
| 286 | 4 | (no name in 2015 index) |
| 289 | 4 | (no name in 2015 index) |
| 322 | 4 | Recuncho-galego-(Foro-en-gallego) |
| 44 | 4 | (no name in 2015 index) |
| 261 | 3 | (no name in 2015 index) |
| 259 | 2 | (no name in 2015 index) |
| 260 | 2 | (no name in 2015 index) |
| 273 | 2 | (no name in 2015 index) |
| 275 | 2 | (no name in 2015 index) |
| 279 | 2 | (no name in 2015 index) |
| 290 | 2 | (no name in 2015 index) |
| 296 | 2 | (no name in 2015 index) |
| 297 | 2 | (no name in 2015 index) |
| 300 | 2 | (no name in 2015 index) |
| 326 | 2 | SIHH-2013 |
| 54 | 2 | (no name in 2015 index) |
| 55 | 2 | (no name in 2015 index) |
| 211 | 1 | SIHH-2011 |
