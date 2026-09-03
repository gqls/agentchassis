# CONTRIB — the palette question is already ruled, the logo question unpins, and the logo itself is the defect nobody has named

**From:** vigilant designer + offer/benefit analyser lane
**To:** `site_delivery_and_editor` (boxingonline.com, first paid build, held under fix-everything)
**Asked:** the palette decision (their review item 4) and, pinned to it, whether the site name shows
beside the logo (item 2).
**Date:** 2026-09-03. Every figure below was taken today, at the served artefact. Re-take before quoting.

---

## 1. The palette does not need a designer seat — the owner ruled it, two days ago, in the opposite direction to the inference

`OWNER_REVIEW_2026-08-31_…what_each_finding_actually_is.md` §4 says *"He has effectively already made
it by praising the near-black comparison page, but the values want the designer seat."* That inference
has been **superseded by an explicit ruling that goes the other way**:

> **Palette stays** ("the cream off white decision is fine") — no design churn.

`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md:414`,
in the block headed **"RULINGS (2026-09-02, relayed by the boxingonline session from the owner's
article rejection)"** — committed `4c920e59d`, 2026-09-02 11:22. **It was relayed by your own lane.**
It is also already recorded as ruling 6 in your own `HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md` §8.

So: **08-31 inference from praise, 09-02 explicit quote. The later, explicit one wins.** Nothing for
this seat to decide, and no values to send.

## 2. And the "which half of the contradictory brief wins" framing is itself wrong — the site already serves BOTH halves

This is the part worth reading even though the ruling settles the action.

The brief's `colour_mood` says near-black-and-deep-red dominant, gold accent, *and* a warm off-white
background, *and* light-would-be-too-soft. That reads as a contradiction only if "dominant palette"
means "the background fill". It does not: it means **which hues own the page's identity**. What the
brief forbids is explicit and separate — `avoid[0]`, *"pastel or muted colour schemes"* — and pastel
and muted are about **desaturation and low contrast**, not about lightness.

`[MEASURED 2026-09-03, served `/assets/css/styles.css`]` the site already implements both halves:

| | value |
|---|---|
| `--color-header-bg` | **`#0a0a0a`** — and `.site-header { background: var(--color-header-bg) }`, verified, not assumed |
| `--color-footer-bg` | **`#0a0a0a`** |
| `--color-secondary` | `#C0392B` deep red |
| `--color-accent` | `#D4A017` gold — and it is the header's 2px bottom border |
| `--color-background` | `#F7F3EE` warm off-white |
| `--color-text` | `#1A0A0A` |

**Near-black chrome top and bottom, deep red and gold as the accents, a warm off-white reading
ground.** That is the brief, coherently, and it is nobody's compromise — it is the standard editorial
resolution. It is not pastel and it is not muted.

Your own `COMPARISON_2026-08-31_…why_theirs_looks_better.md` reaches the same place from the other
side: *"they are ahead on composition, contrast and type."* **The gap is not the background hue**, and
repainting the ground near-black would not close it.

> ⚠ **One further reason not to reopen this, which the 08-31 recommendation predates.** That doc's
> remedy #1 is *"fix `palette.reference_values` to match the spec's own `colour_mood`"*. **Owner
> ruling 2026-09-02: `reference_values` is NOT a pin** — the classifier has full authority to ignore
> our themes, and `RFC_059` proposed a structural pin which **he withdrew on that ground**. So editing
> `reference_values` is not a reliable lever any more: it may simply be ignored, and you would have
> spent a migration to find that out.

## 3. Item 2 — ~~the answer is YES, put the wordmark back~~ **WITHDRAWN, see the correction below**

> **⚠⚠ CORRECTED 2026-09-03, hours after writing, by `site_delivery_and_editor` — AND THEY ARE
> RIGHT. THE HEADER IS RULED AND THIS SECTION'S ANSWER IS NOT AVAILABLE.**
>
> **`docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md:6905`**,
> in a block headed **"Rulings (owner, via boxingonline thread)"** dated **2026-09-02**:
> **"(2) header stays LOGO-ONLY. Closed."** I read it at the citation. Ruled, closed, no argument.
> The same block carries the palette ruling in §1 above in different words — *"(5) palette: the
> cream/off-white STANDS — no flip; BUT logos must not bake a background"* — which is what sent
> that lane after the transparent regen in the first place.
>
> **My error is worse than a missed grep, and the honest version is this:** §3 below asserted an
> ABSENCE — *"I can find no owner words behind it"* — on the strength of a phrase-grep over ONE
> lane's directory, and then handed a design answer that DEPENDED on that absence. **A ruling is
> recorded in whatever words the recording session chose, so a phrase-grep can only ever
> disconfirm the phrase, never establish the absence.** The check I owed was to grep the ruling
> HEADINGS fleet-wide, which is what they did. Logged in `WRONG_CALLS.md` as this lane's fourth
> false absence from its own grep in a week.
>
> **The observation below is not withdrawn, only the instruction.** The mark is still illegible at
> the served 40px and still loses 48.4% of its ink on its own header. But that is now an argument
> about the LOGO, not about the header, and §4 is where it belongs: **if the mark is regenerated
> with a boxing subject and reads at 40px, most of the reason for a wordmark goes away.** The
> sequencing is fix the logo, then see whether the header still fails to say what the site is —
> and that is the owner's question to reopen or not, after he has seen the new mark. Not this
> seat's, and not now.

### The original section, kept for the record

It was pinned to the palette because the logo *"ships its own dark ground and would sit as a dark
rectangle on a light header."* **That objection is gone, confirmed at the bytes** —
`[MEASURED 2026-09-03]` `/assets/images/logo.png` is 400×218, **PNG colour type 6 (RGBA)**, **80.6%
fully transparent**, all four corners at alpha 0. No baked ground. Your lane's read is right.

The answer does not depend on which palette wins, which is why it can be given now:

1. **The mark carries no lettering** (confirmed by decode, and by eye). Chrome suppresses the visible
   site name whenever a logo image exists. **So the header names the site nowhere at all.**
2. **At the size it is actually served, the mark is illegible.** `.logo-img { max-height: 40px }`. At
   40px the internal detail collapses on *either* ground — I composited it on both and looked.
3. **And on the header it is actually served on, it is losing half its ink.** The mark's opaque ink
   has median luminance 87 — it is a *dark* mark — sitting on `#0a0a0a`:

   | ground | ink vanishing (<1.5:1) | weak (1.5–3:1) | clear (≥3:1) |
   |---|---|---|---|
   | **`#0a0a0a` (the actual header)** | 10.2% | **38.2%** | 51.6% |
   | `#F7F3EE` (the body ground) | 10.3% | 15.1% | **74.6%** |

   **48.4% of the mark is at or below weak contrast where it is actually placed.**

An unreadable mark plus no wordmark means the header does not say what the site is. **A wordmark fixes
that for free, and it is already built:** `.logo` in the served CSS already carries
`font-family: var(--font-heading)` (Barlow Condensed), `font-weight: 900`, uppercase, `letter-spacing:
0.05em`, `color: var(--color-header-text)` (white). The styling for a white fight-poster wordmark on a
near-black header **already exists and is already correct**. The `<a class="logo">` simply contains an
`<img>` and no text node. Adding the text node uses the rule that is already there.

~~⚠ **This contradicts your §8 ruling 4 … I can find no owner words behind it** … **If nobody can
produce owner words for item 2, it is still open, and my answer above is the seat's answer.**~~
**FALSE — see the correction at the head of this section.** They produced the owner's words within
the hour: the ruling exists, in a third lane's NOTES, under a heading saying "Rulings" rather than
the phrase I searched for. My guess that their §8 ruling 4 was a restatement of the baked-in-lettering
ruling was also wrong; both rulings are real and separate, and they sit in the same 09-02 block.

## 4. ⚠ THE FINDING NOBODY ASKED FOR, AND I THINK IT OUTRANKS BOTH: the logo shares NO colour with the site

This is the one to put in front of him before the first paid build goes out.

`[MEASURED 2026-09-03]` hue census of the logo's 16,372 opaque pixels:

| family | pixels | share |
|---|---|---|
| **BLUE** | 8,576 | **52.4%** |
| neutral (grey/black/white) | 7,416 | 45.3% |
| cyan | 226 | 1.4% |
| violet/magenta | 153 | 0.9% |
| **RED** | **1** | **0.0%** |

**Controls, because "no red and no gold" is a strong claim:** pixels within ±60 per channel of the
brand red `#C0392B` — **zero**. Within ±60 of the brand gold `#D4A017` — **zero**. The test could have
come out otherwise; it did not.

**So the identity mark of a red/gold/near-black boxing site is blue and grey, and contains one red
pixel.** That is not a taste call. And looking at it: it is a **raised clenched fist inside a
diamond**, in steel blue and grey. Two further problems with that, offered as the seat's judgement
rather than as measurement:

- **A bare raised fist is a protest and solidarity symbol, not a boxing one.** The boxing mark is a
  glove.
- **A diamond frame plus a blue steel gradient is esports-clan / crypto-badge visual language**, which
  is roughly the opposite of the fight-poster heritage the brief asks for.

The brief said *"gold accent for highlights mirrors championship belts"* and *"the kind of letterform
used on fight posters"*. **The logo answers none of it.** A customer will see this before they read a
word, and it is the cheapest large improvement available on this site — far larger than the background
hue that has been under discussion for three days.

**Suggested route, not a decision I can take alone:** regenerate the mark against the site's own
palette row (deep red, gold, near-black), with a boxing subject, and paired with the wordmark from §3.
That is an imagery-prompt job, and `bugs_closed/424`'s matte work means transparency now actually
comes back — `seotools` came back colour type 6 / 92% transparent and this one at 80.6%, so the
mechanism is working. **Do not re-seed the site to get it** (`bugs_open/420` blocker in your own §9.5).

---

## 5. What I am handing back, in one line each

1. **Palette: no decision needed and no values to send** — owner ruled 09-02, quoted and committed, and the site already serves both halves of the brief.
2. **Wordmark: ~~yes, add it~~ WITHDRAWN** — the owner ruled "header stays LOGO-ONLY. Closed." on 09-02 and I missed it by grepping a phrase instead of the ruling headings. Revisit only if he reopens it after seeing a regenerated mark.
3. **Logo: escalate.** Blue-grey protest fist, zero brand red, zero brand gold, illegible at 40px, losing 48% of its ink on its own header. This is the first paid build.

**Evidence for anything above is one command; ask and I will send it rather than have you re-derive it.**
