# Liability framework — idea.uk and SFI26 assessment tool

**Important up front:** I'm not a lawyer. This is a working framework to think
clearly about the exposure and the practical mitigations, plus draft starter
terms. **Before you take real money, get a UK solicitor to review the T&Cs**
— it costs £200–500 for a fixed-fee review and it's worth it, especially for
the SFI tool. Standard caveat: nothing in this document is legal advice.

This covers two products with quite different risk profiles:
- **idea.uk** — recommends AI product ideas for someone's business. Lower
  stakes, indirect causation.
- **SFI26 single-farm assessment** — tells a farmer or advisor what scheme
  actions to apply for, with real money attached. Higher stakes, direct
  financial consequence.

We treat them differently below.

---

## 1. What could actually go wrong

### idea.uk

| Risk | Likelihood | Cost to customer | Cost to us if litigated |
|---|---|---|---|
| Recommended idea turns out to be bad business | Likely (most ideas fail) | They wasted time | Very low — they made the build decision |
| Verified claim turns out wrong ("no competitor exists" but one does) | Possible | Wasted dev effort | Moderate — could be misrepresentation |
| Cited source outdated, customer relied on it | Possible | Decisions on stale info | Low if cited; rises if not |
| Report leaks confidential business info | Low | Business harm | High if it happens |

**Net: low-to-moderate.** Customer is making business decisions on their own
judgement; we're providing analysis. The relevant exposure is misrepresentation
of facts (we said X was true and it wasn't), not the strategic decision to
build something.

### SFI26 assessment — much sharper

| Risk | Likelihood | Cost to customer | Cost to us if litigated |
|---|---|---|---|
| Wrong eligibility call → application rejected | Possible | Lost time + lost £k of grants | High |
| Wrong action stacking → some actions disallowed in combination | Possible | Reduced payment, possibly forced agreement amendment | High |
| Miscalculated combined-area cap or £100k limit | Possible | Application capped lower than possible | Moderate–high |
| Wrong window guidance, miss Window 1 | Possible | Pushed to Window 2 (lower priority, possible funding exhaustion) | High |
| Recommended action turns out incompatible with existing 2024 agreement | Possible | Penalty / forced exit | Very high |
| Tenancy/landlord-consent miss | Possible | Application invalid mid-stream | Very high |
| Scheme rule changes after our report, customer acts on stale info | Likely over time | Variable | Moderate if disclaimed; high if not |

**Net: meaningfully higher.** This is one step removed from financial advice in
substance, even though SFI navigation isn't formally regulated. A wrong number
in a £49 report could cost a farmer £5k–£50k. That's the asymmetry to design
around.

---

## 2. Mitigations (technical, operational, legal)

### Technical

- **Every claim cited inline** to a specific gov.uk page, handbook version, or
  Defra blog post. No unsourced assertions. The verify step already does this;
  enforce it in output formatting.
- **Date stamps everywhere.** Every cited source includes the date we fetched
  it; every report includes a "valid as of YYYY-MM-DD" line at the top.
- **"Verify before acting" callout** at the top of every SFI report —
  prominent, not buried at the bottom.
- **Versioned corpus** so we can show what guidance we used (audit trail if a
  dispute arises). Stamp the report with the corpus version hash.
- **No advice on bespoke legal/tenancy matters.** Tool flags "this depends on
  your tenancy agreement — get your land agent to confirm" rather than asserting.
- **Uncertainty signalling.** Where rules are ambiguous or recently changed,
  the report says so, doesn't paper over it.

### Operational

- **AUTO_DELIVER off for the first batch** on any new tool. Operator reviews
  every report before it goes out. Catches confident-wrong errors before they
  reach customers.
- **Refund policy that's generous and visible.** Stripe loses the fee, we lose
  the engine cost (~£1) — small price to avoid a £5k dispute over a £29 product.
- **No giving the customer a faster process if they request it.** Take the time
  the method needs. Rushing is where mistakes happen.
- **Keep every input and every output for at least 6 years** (matches UK
  contract limitation period). If a dispute arises in year 4, you can show
  exactly what they got and what guidance was current at the time.

### Legal

- **Clear T&Cs** every customer accepts before paying. Draft below. Key clauses:
  - Information service, not professional advice.
  - No solicitor / accountant / land agent relationship created.
  - Liability capped at the fee paid.
  - Customer responsible for verifying before acting.
  - Sources cited; customer can check.
- **Professional indemnity insurance** (PII) — for the SFI tool especially.
  £100k–£500k cover from Hiscox / AXA / similar is typically £200–500/year
  for low-revenue consultancy. Disclose to the insurer that the analysis is
  AI-driven and that you operator-review every output — this is the standard
  way to be honest about the risk.
- **Limited company** to take payment, not a sole trader. Limits personal
  liability if it ever escalates beyond what insurance covers.

### Reality check on UK regulation

- **SFI scheme navigation isn't a regulated profession.** Anyone can give it.
  BASIS / FACTS qualifications exist for crop/livestock advice; CAAV for
  valuation; neither is required for SFI advice.
- **Financial Conduct Authority — doesn't apply.** Scheme payments aren't
  securities or financial products in the FCA sense.
- **Solicitors Regulation Authority — doesn't apply** unless we drift into
  giving legal advice on tenancy or contracts (which the tool will explicitly
  not do — it'll flag and refer).
- **Negligent misstatement (Hedley Byrne v Heller, common law)** is the real
  liability route. If we hold ourselves out as having skill and give negligent
  advice that someone relies on and loses money, we can be liable for that
  loss — unless we have disclaimed the duty of care effectively.
- **Effective disclaimers**: must be conspicuous (not buried), proximate to
  the advice (in the report itself, not just on the website footer), and the
  customer must reasonably have notice. T&Cs at checkout meets the bar; a
  disclaimer in the report itself reinforces it.

---

## 3. Draft starter T&Cs — idea.uk

> **Draft only. Get a UK solicitor to review before going live.** Plain English
> deliberately; you can have a solicitor tighten the wording without losing it.

```
idea.uk — Terms for Reports

1. What you're buying.
You're buying a written analysis report. We give you ideas we've generated
using our structured method, tested against free alternatives, verified with
public web sources, and ranked.

2. This is information, not professional advice.
We are not your solicitor, accountant, business advisor, or consultant. We
don't know your circumstances in detail. Our report is research and analysis
that helps you think — it isn't a recommendation that you should act on
without your own judgement and, where appropriate, professional advice.

3. We try to be accurate. We also make mistakes.
Every factual claim in the report is cited so you can check it. The internet
changes; sources go stale. Before you spend money or make a commitment based
on anything in the report, verify the current state of whatever it depends on.

4. Refunds.
If, on receiving the report, you don't think it surfaces anything worth acting
on, email us within 14 days and we'll refund you. No quibble. We'd rather give
you your money back than have you feel ripped off.

5. What we promise.
We promise to run the method as described on the site, with care, and to
deliver within the time stated. We don't promise that any idea in the report
will succeed if you build it, or that any factual claim will remain true after
the date of the report.

6. What we cap our liability at.
If something goes wrong and we're responsible, our maximum liability is the
fee you paid for the report (currently £29). This doesn't apply to the limited
things UK law says can't be capped (e.g. death or personal injury from
negligence, fraud), which we don't expect to come up here.

7. Your data.
You tell us about your business so we can analyse it. We use that to produce
your report and to improve our method. We don't share it with anyone outside
the people producing your report. We don't sell it.

8. Things you can't do.
You can read the report, use it however you like in your own business, and
share it within your own organisation. You can't resell it or republish it as
if it were your own analysis.

9. Jurisdiction.
These terms are governed by the law of England and Wales. UK courts have
jurisdiction.

10. How to contact us.
[your contact email]
```

---

## 4. Draft starter T&Cs — SFI26 assessment tool (sharper)

Same shape, more explicit:

```
[product name] — Terms for SFI26 Reports

1. What you're buying.
You're buying an informational report on SFI26 eligibility and action options
for a farm you describe to us. It is produced by an AI-driven analysis system,
reviewed by an operator, and references current Defra and RPA published
guidance at the date of the report.

2. This report is informational. It is not advice on which to apply.
We are not your land agent, agronomist, accountant, or solicitor. We do not
know the details of your tenancy agreements, your existing scheme commitments,
your soil tests, your business plans, or your specific RPA correspondence
beyond what you tell us. The report helps you think about your options. It is
not a recommendation that you should apply for any specific action or in any
specific window without:

   (a) checking everything in the report against current Defra/RPA guidance at
       the time you intend to apply;
   (b) reviewing the recommendations with a qualified land agent, agronomist,
       or scheme advisor who has full sight of your specific situation.

3. Scheme rules change. Often.
SFI 2024 was reopened, restricted, and closed within a single year. SFI26 was
launched with reduced action options, changed payment rates, and removed
features compared to earlier rounds. Our report reflects guidance as published
at the date stamped on the report. We are not liable for changes that happen
after that date.

4. We cite our sources.
Every claim in the report links to a specific Defra page, RPA page, or
handbook section, with the date we accessed it. Before you act on any specific
recommendation, click through and confirm it's still current.

5. Refunds.
If the report contains a material factual error and you've not yet acted on
it, email us within 30 days and we'll refund you and produce a corrected
report. If you've already acted on it and lost money, see clauses 7 and 8.

6. What we don't cover.
The report does not advise on:
   - tenancy law, AHA vs FBT specifics, landlord consent obligations;
   - tax treatment of scheme payments;
   - whether the holding qualifies as a single agricultural business unit;
   - environmental permit interactions outside the scheme itself;
   - capital grants or non-SFI scheme matters except where flagged.

Where any of these matter, we say so in the report and refer you to a
qualified professional.

7. What we cap our liability at.
If we make a negligent factual error in the report and you reasonably rely on
it before checking it against current Defra/RPA guidance, our maximum
liability is the fee you paid for the report. This applies whether your loss
is direct financial loss, lost grant payments, lost time, or any other type
of loss. This cap doesn't apply to the limited things UK law says can't be
capped (e.g. death, personal injury, fraud).

8. What you accept by buying the report.
By placing the order you accept that:
   (a) the report is informational, not professional advice;
   (b) responsibility for the decision to apply (and how) sits with you;
   (c) you will verify the report's specific recommendations against current
       Defra/RPA guidance before acting;
   (d) where the report flags an issue requiring professional input, you will
       get that input before acting on the relevant section.

9. Your data.
You tell us about your farm so we can produce the report. We use that data to
produce the report, audit our method, and improve it. We don't share it with
anyone outside the people producing your report. We don't sell it. We retain
inputs and outputs for 6 years (UK contract limitation period).

10. Jurisdiction.
England and Wales. UK courts.

11. Contact.
[your contact email]
```

This is materially stricter than the idea.uk T&Cs because the stakes are
genuinely higher.

---

## 5. What needs to appear in the report itself, not just the T&Cs

T&Cs at checkout are a baseline. To make the disclaimers effective for the SFI
tool especially, the report itself must carry them prominently. Suggested
in-report elements:

**At the top, in a box, not at the bottom in small print:**

> **Read this first.** This is an informational report, not professional
> advice. It reflects Defra/RPA scheme guidance as of [date]. Scheme rules
> change. Before you apply, click through every cited source to confirm it's
> still current, and review your specific application with a qualified land
> agent or agronomist. We make mistakes; you make the decisions.

**Every cited source includes a date** ("Defra SFI26 handbook v12.0, accessed
2026-05-28") so the reader knows the freshness of each individual claim.

**Wherever the report calls a close eligibility question:** an explicit
"this depends on facts we don't have full sight of — confirm before applying"
flag.

**At the end:** "If you spot anything in this report that contradicts current
Defra/RPA guidance, email us and we'll correct and refund."

---

## 6. Operational practices that reduce real risk

These matter more than the legal wording:

1. **Operator-review every SFI report** for the first 50–100 deliveries.
   Catches the confident-wrong errors before they cause harm.
2. **Refund readily.** A £29 refund is dramatically cheaper than a complaint
   that escalates. The page already promises this; honour it without quibble.
3. **Don't push timing.** If a farmer says "I'm submitting in 3 days,"
   produce the report carefully — don't rush. Tell them to verify with their
   agent before submitting.
4. **Keep the audit trail.** Inputs, outputs, sources, corpus version. If
   anyone ever complains, you can show exactly what they got and what
   guidance was current.
5. **Track corrections.** When you spot something wrong in a delivered report
   (operator review caught it, or customer flagged it), record it and update
   the method. Pattern detection is how the tool gets safer over time.
6. **Have one named, qualified UK agricultural advisor on call** as a sense-
   check resource even if you're operating the tool. £tens per consultation
   for sanity-checking edge cases. Cheap insurance against a confident-wrong
   recommendation getting out.

---

## 7. Insurance — practical notes

UK professional indemnity, low-revenue consultancy / information services:

- **Hiscox, AXA, Markel, Direct Line for Business** all offer PII for
  consultants. Online quote, usually 24–48h turnaround.
- **£100k–£500k cover** is typical for a small consultancy; £200–500/year
  premium ballpark for low turnover.
- **Disclose AI-driven analysis.** Insurers will ask; lying invalidates the
  policy. Frame honestly: "AI-assisted research output, operator-reviewed
  before delivery." Some insurers exclude pure-AI output but cover
  human-reviewed AI work.
- **For SFI specifically**, an insurer may want to see your draft T&Cs and
  process. Have them ready.

Don't take real money for the SFI tool without PII in place. £29 reports for
idea.uk are arguably OK without, given the lower stakes, but a £200/year
policy that covers both for £100k is so cheap it's not worth being uninsured.

---

## 8. If a customer complains

Practical sequence, ordered to keep small complaints from becoming big ones:

1. **Reply within 24 hours.** Speed matters more than wording.
2. **Acknowledge the specific issue.** Not "thank you for your feedback" —
   "you said the report claimed X about action GRH6 and that's wrong because Y."
3. **Refund immediately** if there's any plausible case. The £29 is not worth
   the relationship cost.
4. **Offer a corrected report** if they want one and you can produce it.
5. **Don't admit liability for downstream losses** unless you've talked to your
   insurer. "I'm refunding and correcting" doesn't admit liability for any
   loss they say they incurred from acting on the report.
6. **Log the complaint and the fix.** Update the method/corpus if there's a
   pattern.
7. **If they're claiming downstream loss** (lost grant money, etc.), don't
   negotiate alone — call the insurer's claims line immediately. Most policies
   require notification of potential claims; failing to notify invalidates cover.

---

## 9. Summary — what to do, in order, before launch

For idea.uk launch:

1. T&Cs reviewed by a solicitor (~£200, ~1 week). Use the draft above as
   starting point.
2. PII quote from Hiscox or AXA. Decide on cover level. ~1–2 days.
3. T&Cs page live on idea.uk, with checkbox-accept at checkout.
4. Disclaimer paragraph integrated into the report template.
5. Operator-review on for the first 20+ reports.

For the SFI26 tool launch (after idea.uk is running):

1. The above, plus:
2. PII confirmed in force, with SFI scope.
3. Operator-review on indefinitely until 100+ reports without material errors.
4. Top-of-report disclaimer designed as part of the report template.
5. Audit trail and corpus version stamping live before first paid report.
6. One named UK agricultural advisor on call.
EOF
