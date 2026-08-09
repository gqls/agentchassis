# RFC 020 — Third-party harm in the Gauntlet: what the publish gate must refuse, and what is already exposed before anything is published

**Status: OPEN — filed 2026-08-09 by the `provocation_pipeline` lane, for the
`gauntlet_dead_cta` lane, which owns `internal/tools-api` and the publish path.**

**Nothing here is a criticism of that build.** The publish path was shipped
deliberately, with an owner ruling behind it, a plainly-worded button, and a
demonstrated unpublish. This RFC exists because the owner raised a risk question
this week that the original design round did not have in front of it, and because
answering it turned up **three measured gaps** that are cheap to close now and
expensive to close after the first real incident.

> ⚠ **NOT LEGAL ADVICE, and this RFC must not be read as any.** The author is not
> a lawyer. Everything below is engineering: what the system does, what it stores,
> who can see it. Where a legal question is load-bearing it is **named as a
> question for a solicitor**, not answered. The owner has stated his risk appetite
> is low and that he still wants the viral effect; this RFC is written to that
> brief.

---

## 1. Problem + evidence

### 1.1 What the owner asked

> *"I am worried that someone can say something insulting about someone and then
> it gets validated by the site and posted over the internet. I may be liable …
> I still want the viral effect somehow but my risk appetite is low."*

and then, sharpening it:

> *"even if we don't publish is it still a liability — if e.g. the user
> complains?"*

### 1.2 The finding that reframes the whole question

**Publishing is an amplifier, not a threshold.** For defamation, communicating a
statement to **one person other than its subject** is already publication. The
Gauntlet generates argument and a verdict and shows them to a user; if that text
concerns a real, identifiable third party, the public URL is not what brings the
exposure into existence. It multiplies reach, discoverability and therefore
damages.

That inverts the intuitive priority order. **The publish gate is not the first
control — it is the third.** Ahead of it sit: what the AI is willing to generate
about a named person, and what we store.

*(Whether any given statement is actionable, and under which of defamation, data
protection, harassment or the Online Safety Act, is exactly the question for a
solicitor. The engineering point stands regardless of how that resolves: the
system's exposure does not begin at `POST /publish`.)*

### 1.3 What is measurably true today [MEASURED 2026-08-09]

| # | claim | how checked |
|---|---|---|
| 1 | **Nothing sets `noindex`.** Published rounds are search-engine indexable. | `grep -rniE "noindex\|robots"` over `internal/tools-api/` and the round-record sources → **zero hits** |
| 2 | **There is no moderation or filtering of user text.** Not before the AI sees it, not before storage, not before publish. | `grep -rniE "moderat\|profan\|filter\|banned\|abuse\|toxic"` over `internal/tools-api/handlers/*.go` and `store/*.go` → **zero hits** |
| 3 | **There is no retention policy.** Rounds are stored indefinitely. | `198_tools_api_gauntlet_rounds.sql` has `created_at` and an index on it; no expiry column, no reaper |
| 4 | The user's prose and the AI's output are both persisted and both reachable by slug once published. | `publish.go` `PublicRoundHandler`; `store.GetPublishedRound` |
| 5 | Posters are **anonymous**, and more completely than intended: `client_ip_hash` was `sha256("172.18.0.1")` in 83 of 83 rows. | `bugs_open/139` |
| 6 | Last recorded published-round count: **3, all the lane's own harness rounds** (2026-07-31). | lane records — **STALE, 9 days**, and unverifiable from the cluster: `gauntlet_rounds` is on the island VM |

**(6) is the first thing to check and it is not checkable from the main cluster.**
If it is still 3 and they are all ours, there is **no live third-party exposure at
all** and this can be done calmly. If strangers' rounds are public, the ordering in
§5 changes.

### 1.4 Why "we are just a platform" is a weaker position here than for a forum

The usual operator protections concern content **someone else** posted. Two design
facts cut against relying on that shape:

- **The verdict is the site's own text.** When the system pronounces that an
  argument succeeds, that sentence was authored by the service, not by a user. If
  it reads as endorsing a factual assertion about a named person, that is not
  user-generated content.
- **The poster cannot be identified** (1.3 #5), which is the condition under which
  an operator generally has to rely on a proper notice-and-takedown process to keep
  the defence — **and there is no published route to complain.**

Both are questions for a solicitor. Both are also cheap to engineer around.

---

## 2. The questions this RFC asks

**2.1 Should the publish path refuse a round whose text names an identifiable
third party?** (Recommended: yes.) The round is still played, scored and shown to
its author; it simply gets no public URL.

**2.2 Should published rounds carry `noindex`?** (Recommended: yes,
unconditionally, today, independent of every other answer.)

**2.3 Should there be a retention policy on `gauntlet_rounds`,** and does it differ
for published vs unpublished rounds?

**2.4 Should the AI's own generation be constrained** — i.e. should the counter-argument
and verdict be steered away from asserting facts about named real people at all?

**2.5 Should there be a published report/takedown route,** and who receives it?

**2.6 Is a timed/expiring public URL wanted?** (Recommended: **not as a primary
control** — see §5.6. Recorded because the owner raised it specifically.)

---

## 3. Blast radius

All of `internal/tools-api` — `publish.go`, `position.go`, `defend.go`,
`store/rounds.go` — plus the round-record page component and the share-card JS.
**All of it is the `gauntlet_dead_cta` lane's.** This RFC proposes no edits by the
filing lane.

One cross-lane note: the provocation the round argues is supplied by
`provocation_feed_action.go` (ours). **Provocation selection is itself a lever** —
see §5.5 — so the two lanes' choices interact even though the code does not.

---

## 4. What is already done, so this asks only for the judgement

- **Publishing is opt-in and says so.** Owner-ruled publish-on-share, and the
  button label and consent note are set in the JS beside the handler deliberately,
  so markup can never say "save a card" on a control that publishes.
- **Unpublish works and was proven in the negative direction** — the same slug 404s
  once unpublished (2026-07-31).
- **`client_ip_hash` is not returned** by `GET /round/:slug`.
- **Rate limiting exists** (1 rps / burst 5 per caller).
- **Only the lane's own rounds have ever been published**, as of the last record.

So the gap is not carelessness — it is that the threat model was "does the control
do what it says", and this RFC's is "what if the user writes something about a real
person".

---

## 5. Recommendation, ordered by risk removed per unit of virality lost

The owner's brief is *low risk appetite, keep the viral effect*. That ordering is
the whole point of this section: **the cheapest controls cost no virality at all,
and the owner's own suggestion costs the most for the least.**

### 5.1 `noindex` on the round record page — do this regardless

Zero viral cost. Sharing a link still works identically. What it removes is the
worst multiplier: something findable by searching a person's name is a
categorically different problem from something at a URL you have to be handed.
**This is independently true today and should not wait for the rest of this RFC.**

### 5.2 Refuse to publish a round naming an identifiable third party

Targets the actual harm rather than its duration. Most rounds will name nobody, so
the viral path is barely touched. The machinery already exists in the estate: the
`provocation-gate` shipped 2026-08-05 judges text for exactly this class, and the
fleet-wide banned-claims rail is a second precedent.

**Fail closed**, and note the specific asymmetry: a false positive costs one person
a share button; a false negative is the incident this RFC exists to prevent.

### 5.3 A published report route, and a written takedown process

This is what preserves the operator's position given anonymity (§1.4). Cheap, and
it is the one item whose absence is hard to explain after the fact.

### 5.4 Make the verdict's scope explicit where it is read

The AI scores **argument quality**, not whether claims are true. That distinction
is already in this estate's DNA — the provocation gate deliberately exempts the
thesis from the truth rail while holding its factual assertions to it. Saying so on
the card and the record page is both honest and useful. **A disclaimer alone is
weak** (and `bugs_open/126` is this estate's own worked example of a consent fence
that could not be satisfied), so this is a supporting control, never a primary one.

### 5.5 Choose provocations that do not invite naming a target

A provocation about a *category* invites category argument; one that implies a
villain invites naming one. This is a real lever and it is free.

Worked example from the drafts written this week for the ruled audience: five of
seven are about the arguer rather than any target and are low-risk. Two are not —
*"A film that needs explaining has failed"* invites naming directors, and
**"Restaurant food has got worse"** invites naming a specific business and
asserting something factual about it, which is the actionable shape. **The
`provocation_pipeline` lane is pulling the second one before it goes live**,
regardless of how this RFC is decided. Recorded here so the interaction is visible
to both lanes.

### 5.6 A timed/expiring URL — recommended AGAINST as a primary control

The owner raised this specifically, so the reasoning is recorded rather than the
conclusion alone:

- **It does not remove publication**, only shorten it. It may reduce damages; it
  does not address whether liability arises.
- **The viral artefact is a screenshot-able card.** The screenshot outlives the
  URL. An expiring link therefore destroys *our* copy — and with it the ability to
  take it down, to evidence what was actually said, and to demonstrate a working
  process — while the harmful thing keeps circulating.
- **It taxes exactly the thing the owner wants to keep.** A link that dies is a
  link nobody shares.

**As a last layer it is fine** — a retention policy on published rounds (§2.3)
bounds the tail. As the first or only control it buys the least and costs the most.

---

## 6. What this RFC does NOT ask

It does not re-open publish-on-share, which the owner ruled on 2026-07-31 after the
options were costed by measurement. It does not propose the `provocation_pipeline`
lane edit `tools-api`. It takes no position on which legal regime applies — that is
for a solicitor, and §1.2/§1.4 are written to be handed to one.

---

## 7. OWNER RULING

*Not yet given — filed 2026-08-09.*
