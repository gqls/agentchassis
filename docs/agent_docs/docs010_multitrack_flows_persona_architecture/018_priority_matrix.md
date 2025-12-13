# Implementation Priority Matrix

## Visual Phase Dependencies

```
Phase 0: Fix Multipage (REQUIRED)
    ↓
Phase 1: Simple Flows ──┐
    ↓                   │
Phase 2: Personas       │
    ↓                   │
    ├───────────────────┘
    ↓
Phase 3: Components ────┐
    ↓                   │
Phase 4: Interrogation  │
    ↓                   │
    └───────────────────┘
    ↓
Phase 5: Cognitive Architecture
```

## What Can Run in Parallel

### After Phase 0 Complete:

**Track A (Content Quality):**
- Phase 1: Flows
- Phase 2: Personas

**Track B (Site Structure):**
- Phase 3: Components
- Phase 4: Interrogation

**These tracks are independent!**
You can work on both simultaneously if you have the resources.

## Effort vs Impact Matrix

```
High Impact
    │
    │  Phase 0        Phase 2
    │  (Fix it)       (Personas)
    │     ┌────┐         ┌────┐
    │     │ ** │         │ ** │
    │     └────┘         └────┘
    │                              Phase 4
    │                              (Learn)
    │                              ┌────┐
    │                              │ ** │
    │   Phase 1                    └────┘
    │   (Flows)
    │   ┌────┐         Phase 3
    │   │ ** │         (Components)
    │   └────┘         ┌────┐
    │                  │ ** │
    │                  └────┘
    │                              Phase 5
    │                              (Cognitive)
    │                              ┌────┐
    │                              │ ** │
    │                              └────┘
    └────────────────────────────────────── High Effort
```

## Priority Scoring

| Phase | Impact | Effort | ROI | Priority |
|-------|--------|--------|-----|----------|
| Phase 0: Fix Multipage | 10 | 2 | **5.0** | **1** |
| Phase 1: Simple Flows | 6 | 1 | **6.0** | **2** |
| Phase 2: Personas | 8 | 1 | **8.0** | **3** |
| Phase 3: Components | 7 | 4 | 1.75 | 5 |
| Phase 4: Interrogation | 9 | 4 | 2.25 | 4 |
| Phase 5: Cognitive | 10 | 6 | 1.67 | 6 |

**ROI = Impact / Effort**

## Decision Tree

```
START
  │
  ├─> Need working multipage?
  │   └─> YES → Phase 0 (2 weeks)
  │       │
  │       ├─> Content all sounds same?
  │       │   └─> YES → Phase 1+2 (3 weeks)
  │       │       │
  │       │       ├─> Sites look amateur?
  │       │       │   └─> YES → Phase 3 (4 weeks)
  │       │       │       │
  │       │       │       └─> Need to improve quality?
  │       │       │           └─> YES → Phase 4 (4 weeks)
  │       │       │               │
  │       │       │               └─> Want AI personas?
  │       │       │                   └─> YES → Phase 5 (6 weeks)
  │       │       │
  │       │       └─> Sites look good enough?
  │       │           └─> STOP (6 weeks total)
  │       │
  │       └─> Content sounds good?
  │           └─> STOP (2 weeks total)
  │
  └─> Already works?
      └─> Skip to Phase 1
```

## Recommended Sequences

### Sequence 1: Minimum Viable (6 weeks)
```
Week 1-2:  Phase 0 (Fix multipage)
Week 3-4:  Phase 1 (Add flows)
Week 5-6:  Phase 2 (Add personas)
STOP - Assess quality
```

**Delivers:** Working multipage sites with persona-based content
**Risk:** Low
**Cost:** 6 weeks
**Quality:** Good

### Sequence 2: Professional Quality (14 weeks)
```
Week 1-2:   Phase 0 (Fix multipage)
Week 3-4:   Phase 1 (Add flows)
Week 5-6:   Phase 2 (Add personas)
Week 7-10:  Phase 3 (Components)
Week 11-14: Phase 4 (Interrogation)
STOP - Assess quality
```

**Delivers:** Professional-looking sites that learn from examples
**Risk:** Medium
**Cost:** 14 weeks
**Quality:** Very Good

### Sequence 3: Full Vision (20 weeks)
```
Week 1-2:   Phase 0
Week 3-4:   Phase 1
Week 5-6:   Phase 2
Week 7-10:  Phase 3
Week 11-14: Phase 4
Week 15-20: Phase 5 (Cognitive)
```

**Delivers:** AI personas with memory and learning
**Risk:** High
**Cost:** 20 weeks
**Quality:** Cutting Edge

### Sequence 4: Parallel Tracks (10 weeks)
```
Track A (Content):          Track B (Structure):
Week 1-2:  Phase 0          Week 1-2:  Phase 0
Week 3-4:  Phase 1          Week 3-6:  Phase 3
Week 5-6:  Phase 2          Week 7-10: Phase 4
Week 7-10: (Wait for B)
```

**Delivers:** Both content quality AND structure improvements
**Risk:** High (coordination complexity)
**Cost:** 10 weeks (but 2x team)
**Quality:** Very Good

## Quick Wins

### Week 1 Quick Win
Just fix the loop timing in multipage-builder:
```go
// Add delays between spawns
time.Sleep(15 * time.Second)
```
**Impact:** Multipage might just work
**Effort:** 1 day
**Try this first!**

### Week 2 Quick Win
Copy landing-page-builder workflow structure:
```json
{
    "spawn_strategist": ...,
    "call_strategist": ...,
    "loop_pages": {
        "call_researcher": ...,
        "call_writer": ...,
        "store_page": ...
    },
    "wrap_pages": ...,
    "deploy": ...
}
```
**Impact:** Reliable multipage generation
**Effort:** 3-5 days

## Risk Assessment

### Phase 0: LOW RISK ✅
- Clear problem (race conditions)
- Clear solution (sequential generation)
- Similar pattern works (landing-page-builder)
- Can verify easily

### Phase 1: LOW RISK ✅
- Simple database table
- Minimal code changes
- Easy to test
- Easy to rollback

### Phase 2: LOW RISK ✅
- Just adding data to prompts
- No complex logic
- Personas are just profiles
- Measurable impact

### Phase 3: MEDIUM RISK ⚠️
- New action to implement
- Component system is complex
- Rendering logic can be tricky
- Need good testing

### Phase 4: MEDIUM RISK ⚠️
- External dependencies (web scraping)
- VLM analysis costs
- Pattern extraction is fuzzy
- Quality hard to measure

### Phase 5: HIGH RISK 🔴
- 8 new actions
- Complex state management
- Memory persistence tricky
- Hard to debug
- Benefits uncertain until tested

## What to Measure

### Phase 0 Success Metrics
- [ ] Site generation completes without timeout
- [ ] All pages generated
- [ ] Navigation works
- [ ] Deployed to git successfully

### Phase 1 Success Metrics
- [ ] Voice formality measurably different across stages
- [ ] Home page reads more casual than contact
- [ ] Flow stored in database
- [ ] Reproducible results

### Phase 2 Success Metrics
- [ ] Distinct writing styles per persona
- [ ] Stage 1 vs Stage 3 text analysis shows difference
- [ ] Users can identify which persona wrote what
- [ ] Quality score improves

### Phase 3 Success Metrics
- [ ] Sites built from components (no hardcoded HTML)
- [ ] Components reused across sites
- [ ] CSS/JS dependencies managed correctly
- [ ] Professional appearance

### Phase 4 Success Metrics
- [ ] Can scrape and analyze sites
- [ ] Patterns extracted and stored
- [ ] Sites informed by patterns measurably better
- [ ] Quality improves over time

### Phase 5 Success Metrics
- [ ] Persona remembers previous tasks
- [ ] Content quality improves with experience
- [ ] Personality remains consistent
- [ ] Memory queries fast enough

## Budget Estimates

### Development Hours

| Phase | Backend | Frontend | Testing | Total Hours |
|-------|---------|----------|---------|-------------|
| Phase 0 | 40 | 0 | 20 | 60 |
| Phase 1 | 20 | 5 | 15 | 40 |
| Phase 2 | 20 | 5 | 15 | 40 |
| Phase 3 | 80 | 40 | 40 | 160 |
| Phase 4 | 80 | 20 | 60 | 160 |
| Phase 5 | 160 | 20 | 60 | 240 |

### At $100/hour

| Sequence | Cost | Timeline |
|----------|------|----------|
| Sequence 1 (Phase 0-2) | $14,000 | 6 weeks |
| Sequence 2 (Phase 0-4) | $46,000 | 14 weeks |
| Sequence 3 (Phase 0-5) | $70,000 | 20 weeks |

### External Costs

| Item | Phase | Cost |
|------|-------|------|
| LLM API calls | All | $500-2000/month |
| Vector DB (Pinecone) | Phase 5 | $70-500/month |
| Web scraping service | Phase 4 | $200/month |
| Screenshot API | Phase 4 | $50/month |
| Total Monthly | Full | $820-2750/month |

## Stopping Points

### Stop After Phase 0 If:
- Multipage generation works reliably
- Content quality acceptable
- No budget for improvements

### Stop After Phase 2 If:
- Personas provide enough variety
- Sites look good enough
- Phase 3-5 ROI unclear

### Stop After Phase 4 If:
- Component library working well
- Quality plateaus
- Phase 5 seems like overkill

### Continue to Phase 5 If:
- Want cutting-edge AI personas
- Have budget and time
- Quality improvements worth investment
- Building showcase system

## Summary

**Minimum Viable:** Phase 0-2 (6 weeks, $14k)
**Recommended:** Phase 0-4 (14 weeks, $46k)
**Full Vision:** Phase 0-5 (20 weeks, $70k)

**Start with:** Phase 0 + Week 1 quick win
**Assess after:** Phase 2
**Decide then:** Continue or stop

The roadmap gives you options at every step.