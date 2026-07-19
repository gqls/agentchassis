-- ============================================================================
-- relojistas.com — evidence_base site_spec (P7 D1: cite-or-omit)
-- Written 2026-07-19. Site id ecf15e75-a966-4900-bcb0-1c85f689dbfd.
--
-- WHAT THIS DOES
--   Opts relojistas.com into the claims-verification layer. The layer is gated
--   purely by the PRESENCE of a site_specs row with aspect='evidence_base'
--   (loadEvidenceBase returns nil → every check silently skips —
--   validate_page_content.go:683-686). Inserting this row turns on, with no
--   code change and no image roll:
--     1. V2  — the writer prompt gains "## Verified Facts (the ONLY numbers and
--              named entities you may assert…)" rendered from data->>'writer_block'.
--              Live prompt path: {{if .site_specs.specs.evidence_base}}
--              {{if .site_specs.specs.evidence_base.writer_block}} …
--              It explicitly OVERRIDES the writer's unbounded STRICT RULE 14
--              ("never invent") with a bounded "state only these".
--     2. V1a — validate_page_content check 8 runs between writer and save:
--              ScanBannedClaims (severity BLOCKER) + ScanUnregisteredNumbers
--              (severity error, never a blocker).
--     3. V1b — the post-deploy check_unverified_claims sweep.
--
-- SCHEMA (datahelpers/claims.go:56-104, read before writing this file)
--   facts[]:        {id, claim, kind: metric|capability|entity|attestation,
--                    source: EXACTLY ONE of {sql|artifact|attested_by},
--                    verified_at, value?, tolerance?, context_terms?, writer_line?}
--   banned_claims[]: {pattern (case-insensitive regex; invalid regex degrades to
--                    literal substring — a typo never silently drops a ban), reason}
--   allowed_entities[]: names that are NOT claims (carried for the V3 auditor)
--   writer_block:   the string the prompt renders verbatim.
--
--   NOTE: a fact WITHOUT `writer_line` is omitted from a composed writer_block
--   entirely (composeWriterBlock, refresh_evidence_base_action.go:485-489) —
--   "never auto-phrased". We hand-author writer_block here and ALSO carry
--   writer_line on each fact so a future V4 refresh regenerates it faithfully.
--   `writer_block_managed` is deliberately NOT set, so V4 will not overwrite
--   the hand-authored block.
--
-- ============================================================================
-- THE LIMITATION THAT SHAPED THIS FILE — read before trusting the gate
--
--   ScanUnregisteredNumbers is INERT on this site. It only inspects a number
--   whose surrounding window matches businessClaimContextRe
--   (claims.go:338), which is an ENGLISH, BUSINESS-ORIENTED word list:
--   clients|customers|records|businesses|companies|agents|sites|users|
--   subscribers|departments|awards|employees|staff|…|years of experience.
--
--   No Spanish watch sentence contains any of those. "Sumergible hasta 300
--   metros", "reserva de marcha de tres días", "caja de 44 mm" match nothing,
--   so no number in our content is ever scanned. Measurements are excluded
--   again by unitSuffixRe (claims.go:348), and currency amounts by
--   isExcludedNumber.
--
--   CONSEQUENCE: on this site the numeric half of the gate does not fire, and
--   the enforcement that DOES work is (a) banned_claims — patterns we author
--   ourselves, which are language-agnostic because we write them — and (b) the
--   writer_block prompt fence. banned_claims is therefore doing the load-bearing
--   work here, which is why it targets SHAPES OF FABRICATION rather than
--   specific numbers.
--
--   Do not read a clean claims report on this site as "no invented numbers".
--   It means "no banned pattern matched".
-- ============================================================================

-- Pre-flight: confirm no current evidence_base already exists (another thread
-- may have added one). Expect 0 rows.
--   SELECT id, created_at, created_by FROM site_specs
--    WHERE site_id='ecf15e75-a966-4900-bcb0-1c85f689dbfd'
--      AND aspect='evidence_base' AND is_current;

BEGIN;

-- Versioned write: supersede any current row, then insert the new current one.
UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id = 'ecf15e75-a966-4900-bcb0-1c85f689dbfd'
   AND aspect  = 'evidence_base'
   AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by)
VALUES (
'ecf15e75-a966-4900-bcb0-1c85f689dbfd',
'evidence_base',
$json${
  "audit_doc": "docs/agent_docs/docs024_key_docs_latest/traffic_probe/relojistas_rebuild_plan.md (P7 decisions D1-D4, 2026-07-19); corpus = content_feed_items status='relevant', 34 rows at credibility='high' as of 2026-07-19",

  "governing_rule": "relojistas.com is a NEWS PORTAL. It does not sell watches, service them, or represent any brand, and it has no membership or community. No product specification, price, service interval, water-resistance figure, production number, date, or claim about this site's own audience may ship unless it traces to a fact listed below (each carries its source URL) or to a manufacturer's own published specification cited in the text. Horological vocabulary listed in allowed_entities is DEFINITIONAL and is not a claim. When a figure is wanted and not listed: omit it and describe the thing without the number, or tell the reader to consult the manufacturer's published figure. Never approximate, never extrapolate, and never generalise one model's specification into a rule about watches in general.",

  "schema_notes": "facts[]: {id, kind, claim, source:{artifact=source URL}, verified_at, writer_line}. Sources here are ARTIFACT (a published URL), never sql — nothing on this site is live-queryable, so V4 freshness cannot re-prove these; verified_at is the date the corpus row was read. banned_claims[] carries the enforcement on this site because ScanUnregisteredNumbers is inert for Spanish product copy (see the header). allowed_entities[] = definitional vocabulary + brands we REPORT ON — this site asserts no relationship with any of them.",

  "writer_block": "CONTEXT: relojistas.com is an independent Spanish-language watch NEWS portal. It sells nothing, services nothing, and represents no brand. Write in Spanish (es-ES).\n\nSOURCED FACTS — these are the only specific watch claims you may assert. Each traces to a published article. State them as reported facts about a named model, never as general rules:\n- Panerai Submersible PAM01756: dive watch, 44 mm case, steel bracelet, three days of power reserve (reported by TR Magazine)\n- Certina DS Action Diver 38 mm Titanium: titanium dive watch rated to 300 metres (reported by Relojes y Estilo)\n- Longines Conquest Heritage Central Power Reserve: revives a historic central power-reserve complication (reported by TR Magazine)\n- Richard Mille RM 64-01 Tourbillon Colnago: a limited-edition tourbillon made with the cycle maker Colnago (reported by TR Magazine)\n- Audemars Piguet Neo Frame Horas Saltantes: uses a jumping-hours complication, displaying the time without hands (reported by Máquinas del Tiempo)\n- Zenith DEFY Extreme Ultraviolet: a high-frequency chronograph that measures to the hundredth of a second (reported by Debajo del Reloj)\n- Roger Dubuis Excalibur The Kabuto Legacy: a limited edition inspired by 17th-century Japan (reported by TR Magazine)\n- MoonSwatch Mission to the Moon 1969: an 18-carat gold tribute to Apollo 11 (reported by Debajo del Reloj)\n- F.P. Journe Chronomètre à Résonance 022/99R: sold at Phillips, reported as the first commercial example of the model (reported by TR Magazine)\n- Hublot Big Bang Sapphire Sky Blue: a sapphire-cased watch with a substantial power reserve (reported by Tiempo de Relojes)\n- Grand Seiko 62GS: a version in 18-carat yellow gold inspired by the Sakura-Wakaba (reported by TR Magazine)\n- Watch cases in current reporting use stainless steel, titanium, gold and sapphire; some dials are cut from natural stone, which makes each piece visually unique (reported by Máquinas del Tiempo)\n\nVOCABULARY you may define freely — these are definitions, not claims: tourbillon, cronógrafo, reserva de marcha, movimiento automático, movimiento manual, calibre, complicación, edición limitada, hermeticidad, esfera, bisel, corona, horas saltantes, repetidor de minutos, acero inoxidable, titanio, zafiro, lumen.\n\nBRANDS you may REPORT ON (we have no relationship with any of them, and must never imply one): Rolex, Omega, Patek Philippe, Audemars Piguet, Panerai, Longines, Zenith, Grand Seiko, IWC, TUDOR, Hublot, Richard Mille, Vacheron Constantin, Roger Dubuis, Bell & Ross, Certina, Czapek, H. Moser & Cie., F.P. Journe, Louis Vuitton, Hermès, Swatch, Bulova.\n\nHARD PROHIBITIONS — no framing makes these acceptable:\n- Service or maintenance intervals. We hold no manufacturer service schedule. Say 'consulta el intervalo que publica tu fabricante', never a number of years.\n- Prices, valuations, or any suggestion that a watch is an investment or will hold or gain value.\n- Any claim about relojistas.com's own readers, members, community, traffic or history as a forum.\n- Water-resistance or size figures for any model not listed above.\n- Superlatives that cannot be sourced ('el mejor', 'el más preciso del mundo', 'el único').\n- Turning one model's specification into a general rule ('todos los relojes de buceo…').\n\nIf a fact you want is not above: write the sentence without the number, or point the reader to the manufacturer's own published figure. Omitting is always correct; guessing never is.",

  "facts": [
    {"id": "R1-panerai-pam01756", "kind": "entity",
     "claim": "Panerai Submersible PAM01756 is a dive watch with a 44 mm case, a steel bracelet and three days of power reserve",
     "source": {"artifact": "https://trmagazine.es/panerai-submersible-pam01756-el-nuevo-reloj-de-buceo-de-44-mm-con-brazalete-metalico-y-tres-dias-de-reserva-de-marcha"},
     "verified_at": "2026-07-19",
     "writer_line": "Panerai Submersible PAM01756: dive watch, 44 mm case, steel bracelet, three days of power reserve (reported by TR Magazine)"},

    {"id": "R2-certina-ds-action-diver", "kind": "entity",
     "claim": "Certina DS Action Diver 38 mm Titanium is a titanium dive watch rated to 300 metres",
     "source": {"artifact": "https://relojesyestilo.es/certina-ds-action-diver-38mm-titanium-bucear-con-estilo/"},
     "verified_at": "2026-07-19",
     "writer_line": "Certina DS Action Diver 38 mm Titanium: titanium dive watch rated to 300 metres (reported by Relojes y Estilo)"},

    {"id": "R3-longines-central-power-reserve", "kind": "entity",
     "claim": "Longines Conquest Heritage Central Power Reserve revives a historic central power-reserve complication",
     "source": {"artifact": "https://trmagazine.es/longines-conquest-heritage-central-power-reserve-la-elegancia-atemporal-recupera-una-complicacion-historica"},
     "verified_at": "2026-07-19",
     "writer_line": "Longines Conquest Heritage Central Power Reserve: revives a historic central power-reserve complication (reported by TR Magazine)"},

    {"id": "R4-richard-mille-rm6401-tourbillon", "kind": "entity",
     "claim": "Richard Mille RM 64-01 Tourbillon Colnago is a limited-edition tourbillon produced with the cycle maker Colnago",
     "source": {"artifact": "https://trmagazine.es/todo-sobre-el-nuevo-richard-mille-rm-64-01-tourbillon-colnago"},
     "verified_at": "2026-07-19",
     "writer_line": "Richard Mille RM 64-01 Tourbillon Colnago: a limited-edition tourbillon made with the cycle maker Colnago (reported by TR Magazine)"},

    {"id": "R5-ap-neo-frame-horas-saltantes", "kind": "entity",
     "claim": "Audemars Piguet Neo Frame Horas Saltantes uses a jumping-hours complication and displays the time without hands",
     "source": {"artifact": "https://www.maquinasdeltiempo.com/audemars-piguet-neo-frame-horas-saltantes-el-tiempo-sin-agujas/"},
     "verified_at": "2026-07-19",
     "writer_line": "Audemars Piguet Neo Frame Horas Saltantes: uses a jumping-hours complication, displaying the time without hands (reported by Máquinas del Tiempo)"},

    {"id": "R6-zenith-defy-extreme-uv", "kind": "entity",
     "claim": "Zenith DEFY Extreme Ultraviolet is a high-frequency chronograph that measures to the hundredth of a second",
     "source": {"artifact": "https://www.debajodelreloj.com/zenith-defy-extreme-ultraviolet-analisis-tecnico/"},
     "verified_at": "2026-07-19",
     "writer_line": "Zenith DEFY Extreme Ultraviolet: a high-frequency chronograph that measures to the hundredth of a second (reported by Debajo del Reloj)"},

    {"id": "R7-roger-dubuis-kabuto", "kind": "entity",
     "claim": "Roger Dubuis Excalibur The Kabuto Legacy is a limited edition inspired by 17th-century Japan",
     "source": {"artifact": "https://trmagazine.es/roger-dubuis-rinde-homenaje-al-legado-samurai-con-el-excalibur-the-kabuto-legacy-una-edicion-limitada-inspirada-en-el-japon-del-siglo-xvii"},
     "verified_at": "2026-07-19",
     "writer_line": "Roger Dubuis Excalibur The Kabuto Legacy: a limited edition inspired by 17th-century Japan (reported by TR Magazine)"},

    {"id": "R8-moonswatch-gold-1969", "kind": "entity",
     "claim": "MoonSwatch Mission to the Moon 1969 is an 18-carat gold tribute to Apollo 11",
     "source": {"artifact": "https://www.debajodelreloj.com/moonswatch-mission-to-the-moon-1969-oro-moonshine/"},
     "verified_at": "2026-07-19",
     "writer_line": "MoonSwatch Mission to the Moon 1969: an 18-carat gold tribute to Apollo 11 (reported by Debajo del Reloj)"},

    {"id": "R9-fpjourne-resonance-phillips", "kind": "entity",
     "claim": "The F.P. Journe Chronomètre à Résonance 022/99R was auctioned by Phillips and reported as the first commercial example of the model",
     "source": {"artifact": "https://trmagazine.es/phillips-sacara-a-subasta-el-f-p-journe-chronometre-a-resonance-022-99r-el-primer-ejemplar-comercial-de-uno-de-los-relojes-mas-importantes-de-la-relojeria-independiente"},
     "verified_at": "2026-07-19",
     "writer_line": "F.P. Journe Chronomètre à Résonance 022/99R: sold at Phillips, reported as the first commercial example of the model (reported by TR Magazine)"},

    {"id": "R10-hublot-big-bang-sapphire", "kind": "entity",
     "claim": "Hublot Big Bang Sapphire Sky Blue is a sapphire-cased watch with a substantial power reserve",
     "source": {"artifact": "https://tiempoderelojes.com/hublot-big-bang-sapphire-sky-blue/"},
     "verified_at": "2026-07-19",
     "writer_line": "Hublot Big Bang Sapphire Sky Blue: a sapphire-cased watch with a substantial power reserve (reported by Tiempo de Relojes)"},

    {"id": "R11-grand-seiko-62gs-gold", "kind": "entity",
     "claim": "Grand Seiko presented a 62GS in 18-carat yellow gold inspired by the Sakura-Wakaba",
     "source": {"artifact": "https://trmagazine.es/grand-seiko-presenta-un-nuevo-62gs-en-oro-amarillo-de-18-quilates-inspirado-en-el-sakura-wakaba"},
     "verified_at": "2026-07-19",
     "writer_line": "Grand Seiko 62GS: a version in 18-carat yellow gold inspired by the Sakura-Wakaba (reported by TR Magazine)"},

    {"id": "R12-case-and-dial-materials", "kind": "capability",
     "claim": "Watch cases in current reporting use stainless steel, titanium, gold and sapphire; some dials are cut from natural stone, making each piece visually unique",
     "source": {"artifact": "https://www.maquinasdeltiempo.com/cuando-la-piedra-natural-convierte-cada-reloj-en-una-pieza-irrepetible/ (also Certina titanium, Grand Seiko 18ct gold, Hublot sapphire — see R2/R10/R11)"},
     "verified_at": "2026-07-19",
     "writer_line": "Watch cases in current reporting use stainless steel, titanium, gold and sapphire; some dials are cut from natural stone, which makes each piece visually unique (reported by Máquinas del Tiempo)"},

    {"id": "R13-site-is-a-news-portal", "kind": "attestation",
     "claim": "relojistas.com is an independent Spanish-language watch news portal. It sells nothing, services nothing, represents no brand, and has no membership or community. Its news is aggregated from named third-party magazines and links out to them.",
     "source": {"attested_by": "owner via P7 decisions, 2026-07-19; site rebuilt from a dead forum, see relojistas_rebuild_plan.md"},
     "verified_at": "2026-07-19",
     "writer_line": "relojistas.com is an independent Spanish-language watch news portal that links out to the magazines it cites. It sells nothing, services nothing and represents no brand."}
  ],

  "banned_claims": [
    {"pattern": "cada\\s+(dos|tres|cuatro|cinco|seis|siete|ocho|nueve|diez|\\d+)\\s+a[nñ]os",
     "reason": "service-interval class: we hold no manufacturer service schedule, so any interval stated as general advice is invented. Point at the manufacturer's published figure instead."},

    {"pattern": "(revisi[oó]n|mantenimiento|servicio)\\s+(completo\\s+)?cada\\s+(dos|tres|cuatro|cinco|seis|siete|ocho|nueve|diez|\\d+)",
     "reason": "service-interval class, phrasing variant. Deliberately requires a following quantity: an earlier draft matched bare '<noun> cada', which would have fired on the ordinary Spanish 'cada vez' and blocked innocent builds — banned_claims are BLOCKER severity, so breadth here is expensive"},

    {"pattern": "(nuestra|nuestro)\\s+(comunidad|foro|club|taller)",
     "reason": "relojistas.com is a rebuilt news portal with no membership, forum or workshop; any first-person community claim is a fabrication about ourselves"},

    {"pattern": "(miles|cientos|millones)\\s+de\\s+(lectores|usuarios|aficionados|suscriptores|miembros|seguidores)",
     "reason": "audience-size class: no analytics support any audience claim, and CF edge IPs make one uncountable anyway (see the reactivation measurement)"},

    {"pattern": "(se\\s+)?revaloriza|buena\\s+inversi[oó]n|inversi[oó]n\\s+segura|garantiza\\s+(su\\s+)?valor|mantendr[aá]\\s+su\\s+valor",
     "reason": "investment-advice class: unsourceable financial claim on a site with no price data, and a regulatory risk on top"},

    {"pattern": "m[aá]s\\s+(preciso|precisa|exacto|exacta|fiable|resistente)\\s+del\\s+mundo",
     "reason": "unsourceable absolute superlative. The leading article was dropped from an earlier draft: '(el|la)\\s+más…' missed 'el RELOJ más preciso del mundo', because a noun sits between the article and the superlative. Caught by a compile-and-match test, not by reading it"},

    {"pattern": "el\\s+(mejor|[uú]nico)\\s+reloj",
     "reason": "unsourceable superlative presented as fact"},

    {"pattern": "todos\\s+los\\s+relojes\\s+(de\\s+buceo\\s+)?(son|tienen|resisten|ofrecen)",
     "reason": "over-generalisation class: a specification reported for one model must never become a rule about all watches — the single likeliest way a sourced fact turns into a false one here"},

    {"pattern": "(precio|cuesta|valorado\\s+en)\\s*[:€$]?\\s*\\d|\\d+\\s*(euros?|d[oó]lares|francos\\s+suizos)|[€$]\\s*\\d",
     "reason": "we hold no price data for any watch; every price on this site would be invented. 'desde' was dropped from an earlier draft of this pattern: it is ordinary Spanish and 'desde 1954' — a legitimate founding date — would have tripped a BLOCKER"}
  ],

  "allowed_entities": [
    "tourbillon", "cronógrafo", "reserva de marcha", "movimiento automático",
    "movimiento manual", "calibre", "complicación", "edición limitada",
    "hermeticidad", "esfera", "bisel", "corona", "horas saltantes",
    "repetidor de minutos", "acero inoxidable", "titanio", "zafiro", "lumen",
    "Rolex", "Omega", "Patek Philippe", "Audemars Piguet", "Panerai", "Longines",
    "Zenith", "Grand Seiko", "IWC", "TUDOR", "Hublot", "Richard Mille",
    "Vacheron Constantin", "Roger Dubuis", "Bell & Ross", "Certina", "Czapek",
    "H. Moser & Cie.", "F.P. Journe", "Louis Vuitton", "Hermès", "Swatch", "Bulova",
    "TR Magazine", "Tiempo de Relojes", "Máquinas del Tiempo", "Debajo del Reloj",
    "Relojes y Estilo"
  ]
}$json$::jsonb,
'manual',
'relojistas-p7-thread',
true,
'claude-session-relojistas-2'
);

COMMIT;

-- ----------------------------------------------------------------------------
-- Verify (run after COMMIT):
--   SELECT jsonb_array_length(data->'facts')         AS facts,
--          jsonb_array_length(data->'banned_claims') AS bans,
--          length(data->>'writer_block')             AS block_chars
--     FROM site_specs
--    WHERE site_id='ecf15e75-a966-4900-bcb0-1c85f689dbfd'
--      AND aspect='evidence_base' AND is_current;
--   -- expect 13 | 9 | ~2600
--
-- Smoke-test the ban patterns actually compile as intended (Postgres regex is
-- not Go's, so this is indicative, not authoritative — Go compiles them at
-- ParseEvidenceBase):
--   SELECT 'una revisión completa cada cinco años' ~* 'cada\s+(dos|tres|cuatro|cinco|seis|siete|ocho|nueve|diez|\d+)\s+a[nñ]os';  -- expect t
--   SELECT 'consulta el intervalo que publica tu fabricante' ~* 'cada\s+(dos|tres|cuatro|cinco|seis|siete|ocho|nueve|diez|\d+)\s+a[nñ]os';  -- expect f
-- ----------------------------------------------------------------------------
