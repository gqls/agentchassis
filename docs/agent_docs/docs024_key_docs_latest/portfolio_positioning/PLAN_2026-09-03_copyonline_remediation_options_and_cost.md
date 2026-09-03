# copyonline.co.uk — what a clean remediation costs, and the three ways out

**Status: NOTHING HERE HAS BEEN DONE.** This is a costed decision paper for the owner, written so the
choice is one command whenever he makes it. The site is still building as I write.
Lane: `portfolio_positioning`. Underlying defect: `bugs_open/453`.

## 1. The position in one paragraph

copyonline.co.uk has an owner-approved mission brief, current since 13:42:37Z on 2026-09-03, planning
30 pages around a copywriting authority site with a single converting lead-route page, an AI opening
angle and a UK copywriter directory. The brief is stored in a shape neither of its two consumers can
read, so every agent downstream of it has run blind. The site has been classified as a two-sided
marketplace, has had five SEO tools selected for it by domain-guess, and has ten live pages, none of
which is among the 30. The page planner has not run yet.

## 2. Every derived spec was produced without the brief [MEASURED 2026-09-03 17:50Z]

The brief's two most distinctive elements are its stance (*"Copy is a craft with learnable rules, not
a talent"*) and its lead-route page. Testing each current spec for either string, with the brief
itself as the positive control:

| spec | written (UTC) | echoes the stance | echoes the lead route |
|---|---|---|---|
| `mission_brief` **(control)** | 13:42:37 | **yes** | **yes** |
| `tools` | 16:01:52 | no | no |
| `identity` | 16:57:07 | no | no |
| `classification` | 16:57:10 | no | no |
| `content_direction` | 16:57:13 | no | no |
| `design_intent` | 16:57:15 | no | no |
| `vertical_landscape` | ~17:33 | no | no |

Six derived specs, six misses, on a control that hits. (A third marker, the word "directory", matches
three of the six — but that is the *marketplace* misclassification talking, not the brief's copywriter
directory, so it is a false positive and is not counted above. Recording it because a reader running
this test themselves will see it and should not read it as partial success.)

Two of the six say so outright in their own `reasoning` field: the classifier wrote *"no mission brief
was supplied"*, the tool-suggester wrote *"Without existing pages loaded, I'm inferring from the
domain"*.

## 3. How far it has spread, and how long the window is

Measured from `site_work_items`, all times UTC 2026-09-03:

```
15:56:45  needs_composition      completed NOT READY — identity+classification missing; queued a
                                 backfill classifier and closed itself
16:58:07  needs_domain_research  the classifier ran here, BLIND → category=hub,
                                 tags marketplace / community-platform / tool-portal, inferred
                                 (it says so) from the old Drupal 7 rules page
~17:33    vertical_landscape     researched PeoplePerHour and copywriters.co.uk — marketplace
                                 exemplars, correct for the wrong classification
17:33:22  needs_strategy         TRIAGED, priority 8 — has not run yet
   —      the plan               does not exist; no resolved_composition has ever been written
```

**The window is the gap before `needs_strategy` runs**, and it was already queued when this was
written. After strategy comes the plan, and the plan is what commits the page set.

## 4. What is already live and would be in scope for a redo

Ten pages, all `active` and deployed between 16:15:24Z and 17:20:00Z: five tools and one companion
guide each.

| tool page | note |
|---|---|
| Website Brief Starter | plausibly on-brief-adjacent; the owner may want to keep it |
| Insight Injector | plausibly on-brief-adjacent |
| **SERP Snippet Previewer** | **duplicates seotools; already awaiting the owner's retire-or-keep ruling** |
| **Keyword Intent Classifier** | **duplicates seotools; same open ruling** |
| **Title Tag Scorer** | **duplicates seotools; same open ruling** |

A sixth, **Copy Brief Builder**, was suggested and its content generated, but no page exists — the
known silent-loss shape of `bugs_open/218`, recorded earlier by this lane and not re-filed.

None of the five is among the brief's four tools (Headline Scorer, Readability and Clarity Checker,
Call-to-Action Tester, Length and Character Counter). **Three of the five were the subject of an open
question to the owner that they outran.**

## 5. The three options

### Option A — let it finish, then rebuild properly
Do nothing now. The site completes as an SEO tools portal. Fix the brief afterwards, re-run
classification, and rebuild to the 30-page plan.
- **Cost:** the largest, and it grows for as long as the build runs. Every page built between now and
  the decision is a page to retire. It also puts a wrong-shaped site briefly in front of anyone who
  looks.
- **In its favour:** it is the only option that needs no decision today, and nothing is lost that
  cannot be rebuilt — the framework builds sites, that is the point of it.

### Option B — hold the build, repair, re-run, resume  ← the clean one
1. Hold copyonline's open work items so nothing new is decided on the blind classification.
2. Apply the held fix (`SQL_2026-09-03c_..._HOLD.sql`) so the brief becomes readable.
3. Supersede the six blind specs and let the pipeline rewrite them — **do not hand-write them**; the
   estate's rule is that the framework writes the content.
4. Re-run classification and **verify at the artefact**: the new `reasoning` must not say the brief was
   missing. That sentence is the pass/fail test, and the agent writes it itself.
5. Release the hold and let strategy, composition and the plan run against a brief they can see.
- **Cost:** one intervention now, plus a decision on the ten live pages. Everything after step 5 is
  the build the owner already approved.
- **Against it:** step 1 is a genuine intervention in a running build, which is the thing the owner
  said not to do. It needs his word.

### Option C — apply only the brief fix and leave the build running
Tempting and **I recommend against it.** It hands the planner a correct brief and a wrong
classification side by side, and the planner reads both. That is not a repair, it is a muddle, and a
half-blind plan is harder to reason about afterwards than a cleanly blind one. **A partial fix is
worse than either whole option.**

## 6. What I have and have not done

Written, guarded, verified, **not applied**: `SQL_2026-09-03c_..._HOLD.sql` (step 1 of option B).
Nothing has been changed on the running build. The owner's standing instruction is not to change a
build that is already running and he has not answered since; silence is not permission, so it waits.

What I got wrong and have corrected in place: I told him this was "one well-evidenced change". It was,
before the consumers ran. They have now run and persisted their blind reads, so it is a sequence, not
a change. The correction is at the foot of the SQL file and in `NOTES` entry (pp).
