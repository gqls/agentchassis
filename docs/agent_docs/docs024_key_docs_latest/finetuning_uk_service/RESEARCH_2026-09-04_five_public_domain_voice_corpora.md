# RESEARCH 2026-09-04 — five public-domain corpora with a defined voice, and the filter each one needs

Commissioned on the owner's *"please pick say 5 that are safe from copyright and do them"*. UK copyright
is life of the author + 70 years, so a work is public domain in 2026 only if the author died in **1955 or
earlier**. All five clear that by decades. Every figure below is directly measured against the actual
Project Gutenberg files except where marked.

| # | Author | Dates | PG ebooks (verified) | Pieces after filter | Median length | The voice |
|---|---|---|---|---|---|---|
| 1 | **Samuel Pepys** | 1633–1703 | 4200 | **3,149** diary entries | **276 words** | A busy naval official's private shorthand: money, food, quarrels and vanity in one breath, ending "and so to bed." |
| 2 | **Horace Walpole** | 1717–1797 | 4609, 4610, 4773, 4919 | 1,154 letters | 771 words | Aristocratic gossip written to be read aloud, giving politics and dresses equal weight. |
| 3 | **Joseph Addison** | 1672–1719 | 12030; Tatler 13645, 45769, 31645, 49009 | 273 papers | 1,385 words | "Mr Spectator": urbane, ironic, arriving at a moral without raising his voice. |
| 4 | **Charles Lamb** | 1775–1834 | 9365, 10851, 10343 | 612 letters + ~74 essays | 470 words (letters) | A punning London clerk who turns a trivial domestic fact into real feeling. |
| 5 | **Katherine Mansfield** | 1888–1923 | 1429, 44385, 66871, 1472 | ~65 stories `[APPROX]` | ~2,000 words `[APPROX]` | Sharp modernist miniatures: a mood caught in one scene, the feeling left just under the surface. |

`[APPROX]` — Mansfield's counts are approximate: two of her four files defeated the automatic heading
detection. Everything else is measured.

## The filter each corpus needs — this is the deliverable, not an appendix

1. **Pepys** — drop entries containing the macaronic sex-code (`hazer`, `con elle`, `avec la`, `tocar`)
   or the six named women (Bagwell, Burroughs, Michell, Willet, Martin, Lane): **203 of 3,352 removed
   (6.1%)**. Also strip Wheatley's bracketed footnotes (4.4% of the file).
   ⚠ **The Victorian edition does NOT do this for you.** `Bagwell` appears **58** times and `hazer` **21**
   in PG 4200; the coercive material survives Wheatley's censorship. The first version of this research
   claimed otherwise and was corrected by measurement.
2. **Walpole** — no content filter; strip the Victorian editorial footnotes, **15.4%** of the four files.
3. **Addison** — keep only papers signed "Addison" (discards Steele and Budgell); strip Morley's notes.
4. **Lamb** — drop the single essay *Imperfect Sympathies*; separate E. V. Lucas's notes, which are
   **interleaved between the letters**, not in an appendix.
5. **Mansfield** — none.

## Rejected, and why it matters that they were

- **Saki** — a peer-reviewed article (Gibson, *Jewish Culture and History* 9:1, 2007) defends one story
  while conceding *"the unironic, overt anti-semitism of earlier stories"*. A corpus-level
  characterisation by a cited scholar means an exclusion list built from our own lexical sweep cannot be
  defended in public. Mansfield replaced him: half the volume, nothing to explain.
- **Chesterton, Leacock** — their problems sit in **plain English** and cannot be filtered with
  precision. That, not severity, is the discriminator: Pepys survives because his are written in a
  distinctive private code.
- **A. A. Milne** — died **31 January 1956**, so UK public domain on **1 January 2027**. An obvious
  candidate for exactly this demo, and unusable until then.

## Two traps for whoever ingests

- **PG #28919 is a STUB** — a ~6 KB duplicate listing of *Frenzied Fiction* that returns ~484 words, not
  the text. A pipeline resolving by title silently produces an empty corpus. Use **#8457**.
- **Editions carry their own copyright.** A modern editor's introduction, notes or translation is a new
  work with its own term. Every corpus above is stripped of editorial apparatus for that reason as well
  as for cleanliness.

## The sentence to build the copyright page on

**"We only used Project Gutenberg" is not a defence.** Every damaging passage discussed here is on
Project Gutenberg, free, and greppable in seconds by anyone who reads our launch. **The corpus filter is
the deliverable, not the source.**

## Recommendation

**Build Pepys first.** 3,149 entries at a median of 276 words is an order of magnitude better in SHAPE
than anything else on the list — short, first-person, dated, and about ordinary daily business, which is
what the demo's task shapes ("write this in my voice") actually need. The voice is recognisable in one
line, which is what makes a before-and-after legible to a visitor.
