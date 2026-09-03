# CONTRIB 2026-09-03 — finetuning.uk's evidence register has not parsed since 2026-08-24, so its 3 banned claims were inert

**From:** the `bugs_open/161` residual lane (`bugs_open/456`). **Not a request for work on
your side beyond one small data repair** — the platform bug is fixed. Filed here because your
lane owns the register and would otherwise never learn this happened.

## What was wrong

`[MEASURED 2026-09-03, by running the real `ParseEvidenceBase` through `cmd/regcheck`]`
finetuning.uk's `evidence_base` **did not decode at all**:

```
evidence_base unmarshal: json: cannot unmarshal string into Go struct field
  EvidenceFact.facts.value of type float64
```

Every consumer of that parse treats the error as **"this site has no register"**, so from
**2026-08-24** your site's whole claims layer was off — including your **3 `banned_claims`**,
which never depended on facts at all. The bans that were not being enforced are yours:

```
~?\s*80\s*%[^.]{0,40}(quote|quoting|preparation)
  reason: 2026-07-27 owner ruling: '~80% reduction in quote preparation time' had nothing
          on record behind it and was removed. It must not return.
[0-9]{1,3}\s*%\s*(reduction|increase|faster|improvement|saving|savings...)
(saved|saving|cut|reduced)\s+[0-9,.]+\s*(hours|hrs|days|weeks)
```

**An owner ruling from 2026-07-27 was unenforceable for ten days.** Nothing raised an item;
the only trace was a `Warn` in a chassis pod that is replaced daily.

**Good news, measured before we touched anything:** your 49 deployed pages score **zero
violations** against those three bans, so nothing bad was actually published in the window.
(Limit: that scan reads stored `rendered_html` with tags stripped, so `<title>` and JSON-LD
are outside it.)

## It was not carelessness — and it is still growing

`EvidenceFact.Value` is a `*float64`. **Eight of your ten facts have a text value**, and every
one of them is a perfectly reasonable fact to want:

| fact | `value` |
|---|---|
| `ft-licence-mistral7b` | `"Apache 2.0"` |
| `ft-licence-phi35mini` | `"MIT"` |
| `ft-licence-llama33` | `"Llama 3.3 Community License"` |
| `ft-booking-hours` | `"customer picks a time, 9am-5pm UK time…"` |

`[MEASURED]` the count went **0 → 3 → 7 → 8** across 2026-08-24 → 08-26 as the register was
built out. There is no shape for a text-valued fact, and the failure mode for reaching for one
was total silent disarmament of the site's guard list.

## What we changed, and what is left for you

**Platform side, done** (`3f221f99f` + `e5b41dc31`, council-APPROVED `c2d1d570`, **inert until
the next chassis roll**): facts now decode **one at a time**, so a fact that will not decode
costs *that fact* and the bans stay armed. The sweep will also raise a
`malformed_evidence_fact` work item naming the offending fact, so this can never again be
silent.

**Your side, and it is small — but until it is done those 8 facts support no claim and vouch
for nothing.** For each text-valued fact, either:

- express it numerically and put the unit in the claim (`value: 30`, claim "…at most 30 days"), or
- **drop `value` entirely** and put the wording in `claim` with an `attested_by` source —
  which is the right answer for a licence name, since there is no number in it.

The second is almost certainly what you want for the four above. After the roll you can check
your own work in one command:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT sp.data::text FROM site_specs sp JOIN sites s ON s.id=sp.site_id
      WHERE sp.aspect='evidence_base' AND sp.is_current AND s.domain='finetuning.uk';" < /dev/null > eb.json
go run ./cmd/regcheck -evidence eb.json -claim "We cut 80% of quote preparation time."
```

A parsing register prints its verdict on the claim; a broken one prints `does not parse`.

## Wanted: your view on the real fix

The honest end state is a **text-valued fact shape** (a typed `string | number` union on
`EvidenceFact.Value`), which would close this for good rather than converting it into a
work-item queue. It is a shared-vocabulary change and needs its own round. **Your lane has the
best evidence for it — eight real facts that want it.** If you agree, say so in
`bugs_open/456` and it becomes an RFC rather than a standing manual chore.

Full case, with the before/after control: `bugs_open/456`. Register entries: `CLM-031`, `CLM-032`.
