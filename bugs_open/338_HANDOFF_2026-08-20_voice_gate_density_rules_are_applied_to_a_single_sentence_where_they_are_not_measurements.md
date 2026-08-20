# 338 — the voice gate's DENSITY rules are applied to a single sentence, where they are not measurements

**Filed** 2026-08-20 by the `meta_description_never_backfilled` lane. **Status: OPEN.**
Needs a Go change and a fleet roll.

> **Resolve by SLUG** (`voice_gate_density_rules_on_a_single_sentence`) — bug numbers
> collide on this tree, and `git log` the FILE PATH, not the number.

---

## 1. What a reader needs first

A **voice gate** is a per-site set of copy rules held in `site_specs` (`aspect='voice'`,
key `voice_gate`). It carries **two different kinds of rule**, and the distinction is the
whole of this bug:

- **CONTENT rules** — `banned_phrases`, a list of regexes with reasons
  (*"harness — hype verb"*, *"trust — overused; say what is checked/verified instead"*).
  These are true of **any** string, however short.
- **DENSITY / DISTRIBUTION rules** — `mean_sentence_words`, `long_sentence_words`,
  `long_sentence_share`, `em_dash_per_1000_words`, `triads_per_page`,
  `expect_contractions`. These are **statistics computed across a corpus**. `ScanVoiceTells`
  computes them over a page's rendered blocks, which is what they were designed for.

`save_page_meta_description` (register **SEO-004**, `bugs_open/320`) applies the **whole**
gate — both kinds — to a single meta description, which is one sentence of ~15-25 words.

**Over one sentence the second kind degenerates.** "Mean sentence length" of one sentence
is just its length. "Long-sentence share" is 0 or 1. "Expects contractions" over 20 words
is noise. None of them is a measurement at that sample size.

## 2. The failure, observed in production

`[MEASURED 2026-08-20]` first scheduled run of the backfiller, on
`leopardessconsulting.co.uk`. The writer produced this, which is good ordinary copy well
inside a search result's display budget:

> "Read research-backed insights on AI adoption across healthcare, finance, hiring and
> data security, with findings on what builds and breaks confidence in these systems."

and the action refused it:

```
reason: voice_tell
detail: long_sentences: average sentence too long for the register
        ("mean sentence length 24.0 words")
```

24 words, against a default trip of **22** (`voicetells.go:179-180`,
`meanTrip := defaultF(g.cfg.MeanSentenceWords, 22)`).

The page stayed blank. **On an hourly schedule that is a permanent hourly LLM bill for a
page that never fills**, and nothing about it reads as a failure: the orchestration
COMPLETEs, the scheduled task stamps a clean run, and only `save_result.reason` says
otherwise.

## 3. The evidence that this is not just one lane's opinion

**`[MEASURED 2026-08-20]` of the 9 sites with an enabled voice gate, SEVEN have already
switched the length checks off by hand:**

| domain | `mean_sentence_words` | `long_sentence_words` | `long_sentence_share` |
|---|---|---|---|
| gamesdesign.co.uk | **100000** | **10000** | 1.1 |
| lendzy.co.uk | **100000** | **10000** | 1.1 |
| loancash.co.uk | **100000** | **10000** | 1.1 |
| noted.co.uk | **100000** | **10000** | 1.1 |
| relojistas.com | **100000** | **10000** | 1.1 |
| vetcomparison.uk | **100000** | **10000** | 1.1 |
| webdesign.uk | **100000** | **10000** | 1.1 |
| leopardessconsulting.co.uk | *(unset → default 22)* | *(unset → 25)* | *(unset)* |
| oufe.com | *(unset → default 22)* | *(unset → 25)* | *(unset)* |

Those are not thresholds. `100000` words and a share of `1.1` (i.e. >100%) are values
chosen so the check can never fire, while keeping the banned-phrase list.

**The estate has already voted on this, one site at a time, and nobody wrote down why.**
That is the part worth acting on: seven sites carry an undocumented workaround for a rule
that does not fit, and the two that did not carry it are the two this bug bites.

## 4. Where to fix it

`metaDescriptionFailsCopyGates` in
`platform/orchestration/actions/save_page_meta_description_action.go`.

It currently does:

```go
if findings := gate.ScanVoice(blocks, false); len(findings) > 0 {
    f := findings[0]
    return "voice_tell", fmt.Sprintf("%s: %s (%q)", f.Check, f.Reason, f.Matched)
}
```

`VoiceFinding.Check` already names which rule fired — `banned_phrase`,
`em_dash_density`, `triad_density`, `long_sentences`, `no_contractions`,
`flourish_ending`, `strawman` (`voicetells.go:85-93`). So the filter is available and
needs no new plumbing.

**The fix is to keep the content findings and drop the distribution ones** for a
single-value field. `banned_phrase` stays. `long_sentences`, `no_contractions`,
`em_dash_density` and `triad_density` are corpus statistics and should not gate one
sentence. (`flourish_ending` / `strawman` are judgement calls — decide deliberately
rather than by omission, and say which way in the code comment.)

⚠ **Do NOT fix this by raising a site's thresholds.** That is relaxing a checker to fit
the content, and worse, it disables the rule for the site's **pages** too — where it does
work and is the thing the gate was built for.

⚠ **Do NOT drop the em-dash rule as "a density rule".** The house style bans em dashes
outright; a *rate per 1000 words* is the wrong instrument over 20 words, but **any** em
dash in a meta description should still be refused. If the density check is dropped, put
a flat "contains an em dash" test in its place rather than losing the rule.

## 5. Blast radius

Small and bounded today: `save_page_meta_description` is the only caller applying the gate
to a single-value field, and only **2 of 27** sites are affected.

**It grows the moment anyone reuses the gate on another short field** — a page title, a
nav label, an alt text — which is a reasonable thing to want, because the banned-phrase
half genuinely belongs there. That is why this is worth fixing at the seam rather than
per-caller.

## 6. How to verify a fix

```sql
-- the two affected sites, before and after
SELECT s.domain, p.name, length(COALESCE(p.meta_description,'')) AS len
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain IN ('leopardessconsulting.co.uk','oufe.com')
  AND p.status='active' AND COALESCE(p.meta_description,'')='';
```

**Induce both arms, or the test proves nothing:**
1. A candidate containing a banned phrase must still be **REFUSED** with
   `reason=voice_tell`, `detail` naming `banned_phrase`.
2. A clean 24-word candidate on `leopardessconsulting.co.uk` must now be **WRITTEN**.

A test that only exercises (2) is indistinguishable from one that removed the gate.

## 7. Interim mitigation already in place

Migration `501` asks the writer for **≤20 words**, which clears the default trip of 22.
That is a workaround and the migration says so in its own header. It is also better copy
on its own merits, which is why it was acceptable as an interim — but it does not make the
gate correct, and a site that sets `mean_sentence_words: 15` would break it again.

## 8. Provenance

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and stating the substitution rather than omitting it:** the failure was
observed in a real production run (the refusal text is quoted verbatim from
`orchestration_states.collected_data.save_result`), the deciding code path was read
(`voicetells.go:179-180`, `save_page_meta_description_action.go`), the threshold arithmetic
is arithmetic, and the 9-site config census is a direct query over `site_specs`. The claim
is narrow — one function applies a corpus statistic to a one-element sample — and does not
assert a cause outside the symptom.

Landmine written (`LANDMINES.md`, *"The voice gate's DENSITY rules are statistics over a
CORPUS…"*). Full trail: `bugs_open/320` §13 and
`docs024_key_docs_latest/meta_description_never_backfilled/`.
