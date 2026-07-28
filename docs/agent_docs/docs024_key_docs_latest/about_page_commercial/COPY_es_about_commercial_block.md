# Spanish copy for `about-commercial-block` — the first non-English site

**Written 2026-07-28 for relojistas.com (tier 2), by the traffic_probe thread.**

> **STATUS: SHIPPED 2026-07-28.** Owner authorised the build. The `language` seam is in the
> component's schema, the Spanish branches are in `html_template`, and this copy is **live on
> relojistas.com/sobre-nosotros.html** — the first non-English render of the block. Absent
> `language` ⇒ English, so the rest of the fleet is unchanged. See NOTES 2026-07-28.

Grounded in relojistas' own `site_specs.content_direction`, not in a translator's instinct —
that spec is the reason several obvious renderings are wrong (see *Rejected* below).

---

## The copy

`{{.domain}}` interpolates as today. Heading, then built-by, then the gated commercial line —
same order as the English.

### Heading

```
Sobre este sitio
```

### Built-by (class ∈ {portfolio, storefront})

```
Diseñado y construido por fundamentallyai.com — ver cómo se hace.
```

link on `ver cómo se hace` → `built_by_url`, `rel="nofollow"`.

### For-sale line — the tier ladder (portfolio ∧ for_sale_requested ∧ ¬advertising_active)

**Tier 1 (flagship)** — real monitored brokerage route:
```
El dominio {{.domain}} está disponible para su adquisición. Las consultas se atienden
a través de nuestro equipo de dominios.
```
button / link text: **`Hablar con nuestro equipo de dominios`**

**Tier 2 (mid) — relojistas.com:**
```
El dominio {{.domain}} está disponible para su adquisición — puedes manifestar tu interés.
```
link on `manifestar tu interés` → `marketplace_url`. Button form: **`Manifestar interés`**

**Tier 3 (rest):**
```
{{.domain}} forma parte de nuestra cartera de dominios y podría estar disponible para
su adquisición.
```
button / link text: **`Hacer una consulta`**

### Advertise line (portfolio ∧ inventory_open ∧ ¬for_sale_requested)

```
Publicidad en {{.domain}}. Este sitio ofrece un número reducido de espacios patrocinados,
con una tarifa mensual fija, que se contratan en advertise.co.uk.
```
button: **`Anunciarse aquí`**

*Deliberately `un número reducido`, never a count.* The English note applies verbatim in
Spanish: the moment it says "tres espacios" on a site where nothing counts placements, the
claims gate is right to treat it as unsupported.

---

## Rejected, and why — these are the decisions, not the translation

| rejected | why |
|---|---|
| **`en venta` / `se vende`** | The classified-ad register, and the exact domainer cliché the English design rejects. Worse here than in English: relojistas' spec bans *"lenguaje de e-commerce o de tienda"* outright. |
| **`oferta`, `precio`, `comprar`** | Named in that spec's *things to avoid* list. `oferta` is doubly wrong — it means both "offer" and "special offer", and the second reading is the banned one. |
| **`¡Adquiere este dominio!`** | Imperative + exclamation. The spec's CTA rule is *"de servicio, no de presión… sin urgencia artificial ni escasez fabricada."* |
| **`dominio premium`** | Adjectival value claim. The English design rejects "premium" outright; the site's `authority_level` also demands *"autoridad ganada por precisión factual… no por declaraciones de grandeza."* |
| **`agente`** | Overloaded — this is an AI-agents company. English uses "domain team"; Spanish `equipo de dominios` keeps that. |
| **`registrar tu interés`** | Calque of "register your interest"; `manifestar` is the idiomatic verb. |
| first-person `Diseñamos y construimos…` for built-by | The spec reserves **`nosotros` for Relojistas itself**. A first-person built-by line puts fundamentallyai's "we" in Relojistas' mouth on Relojistas' own page. Third person (`Diseñado y construido por…`) removes the collision — a problem the English doesn't have because "We design and build sites like this one" reads as the badge speaking. **Worth fixing in the English too.** |

**`está disponible para su adquisición`** is the load-bearing phrase. It is formal without
being legalese, carries the *representation* framing the design asks for, and — unlike every
natural-sounding alternative — does not collide with the site's e-commerce ban.

## Register checks against the spec

- **Formality**: *"semiformal, culto pero accesible… español estándar inteligible en España y
  América Latina"* — no peninsular-only or LatAm-only lexis. `adquisición`, `cartera`,
  `manifestar interés` are neutral across both.
- **CTA style**: the spec's own verbs are infinitives — *Leer la noticia completa, Consultar la
  guía*. `Manifestar interés`, `Hacer una consulta`, `Anunciarse aquí` match that shape.
- **Tuteo**: the spec says *"sin tuteo forzado"*. Tier 2's `puedes manifestar tu interés` uses
  tú, consistent with the site's existing intent-probe privacy text (*"lo que escribes"*).
  Tiers 1 and 3 avoid person entirely, which is right for a colder notice.

## The tension the owner should see

**relojistas' editorial rail and this block point opposite ways.** The spec says, in the site's
own characteristic phrases, *"En Relojistas no vendemos relojes ni cobramos por recomendarlos"*.
A commercial block is the first thing on the site that sells anything.

It is defensible — and only because of a distinction the copy must keep visible: **Relojistas
does not sell watches; the DOMAIN is what is available.** Every tier above therefore leads with
`El dominio {{.domain}}` / `{{.domain}} forma parte de nuestra cartera de dominios`, never with
a bare "está disponible". Drop that and the block reads as the editorial contradicting itself.

The site already has a voice for this, from its own spec's example phrases:

> *"Este dominio fue durante años un foro activo. El foro ya no existe, pero la audiencia sigue
> aquí — y nosotros también."*

If the owner wants the block warmer than the ladder allows, that sentence is the register to
aim at — but it is a **design change to the locked wording**, not a translation, so it is his
call and this workstream's, not a site thread's.

## What is still needed to ship it

1. **A language seam.** There is none: no `language`/`locale` column on `sites`; the only
   per-site language anywhere is inside `deploy_config.rss_feed.language` (`"es"` for
   relojistas). Verified 2026-07-27.
2. **Wire this copy** behind it in `about-commercial-block`, whose strings are currently Go
   template literals.
3. Then relojistas is **one section-insert** away — its `site_specs.commercial` is already
   written and true (`portfolio / tier 2 / for_sale_requested / marketplace_url`), and inert
   until a section exists.

*(An earlier draft of this file said the Afternic minimum offer was 0 and the anti-lowball
floor absent. **Retracted 2026-07-28 — the floor is $12,000.** My misreading of a dashboard
paste; the design's protection was never missing.)*
