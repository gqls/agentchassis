# Landing Page vs Brochure Site - Quick Reference

## Classification Signals

### Landing Page
```
✓ Single conversion goal
✓ Specific target action (signup, purchase, download)
✓ Product/service launch focus
✓ Lead generation emphasis
✓ One primary CTA
✓ Short, focused messaging
```

### Brochure Site
```
✓ Multiple information sections
✓ Company overview/about us
✓ Team/leadership pages
✓ Service/product portfolio
✓ Case studies/portfolio
✓ Blog/resources section
✓ Multiple navigation pages
```

---

## Recommended Builders

| Site Type | Builder Agent | Purpose |
|-----------|---------------|---------|
| Landing | `landing-page-builder` | Single-page conversion sites |
| Brochure | `content-site-builder` | Multi-page corporate sites |
| Large Multi-page | `multipage-website-builder` | 20+ page sites with batching |
| Generic | `website-builder` | General-purpose fallback |

---

## Questionnaire Differences

### Landing Page Questions
```json
{
  "brand_name": "Product/service name",
  "tagline": "Value prop in one line",
  "tone": "professional|friendly|bold",
  "primary_benefit": "Main value for visitors",
  "unique_selling_points": "3-5 bullet points",
  "target_audience": "Ideal customer",
  "primary_cta": "Sign Up / Buy Now / etc",
  "primary_cta_url": "/action-url",
  "secondary_cta": "Learn More",
  "has_testimonials": true/false,
  "client_count": "Social proof numbers",
  "notable_clients": "Logo wall clients"
}
```

### Brochure Site Questions
```json
{
  "company_name": "Full company name",
  "tagline": "Corporate positioning",
  "about_us": "Company history/mission",
  "services": [
    {
      "name": "Service name",
      "description": "Service details"
    }
  ],
  "target_audience": "B2B/B2C audience",
  "tone": "professional|friendly|corporate",
  "key_differentiators": "What sets company apart",
  "leadership_team": [
    {
      "name": "Executive name",
      "title": "C-level title",
      "bio": "Background/credentials"
    }
  ],
  "case_studies": [
    {
      "client": "Client name",
      "challenge": "Problem solved",
      "result": "Quantified outcomes"
    }
  ],
  "contact_email": "info@...",
  "contact_phone": "+1...",
  "headquarters": "City, State",
  "has_blog": true/false,
  "has_careers": true/false,
  "primary_cta": "Contact Us / Schedule Demo",
  "primary_cta_url": "/contact"
}
```

---

## Example Objectives

### Landing Page Objective
```
"Showcase our revolutionary AI agent orchestration platform 
that enables autonomous multi-agent workflows. Convert 
technical decision-makers and CTOs looking to build 
scalable AI systems."

Signals: Product focus, conversion goal, specific audience
Classification: "landing" (confidence: 0.95)
```

### Brochure Site Objective
```
"Create a comprehensive corporate website for ACME Consulting,
a management consulting firm. Showcase company overview, 
leadership team bios, service offerings, case studies, 
insights blog, careers section, and contact information."

Signals: Multiple sections, company focus, information architecture
Classification: "brochure" (confidence: 0.92)
```

---

## Workflow Comparison

| Step | Landing Page | Brochure Site |
|------|--------------|---------------|
| 1. Classify | → "landing" | → "brochure" |
| 2. Recommend | → landing-page-builder | → content-site-builder |
| 3. Questionnaire | 10 fields (conversion-focused) | 15+ fields (comprehensive) |
| 4. Build | Single page HTML/CSS | Multi-page structure |
| 5. Output | index.html + styles | index.html, about.html, services.html, etc. |

---

## HITL Response Templates

### Landing Page Confirmation
```bash
kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P -b kafka:9092 -t system.agent.generic.responses \
-H correlation_id=$CORRELATION_ID \
-H in_response_to_request_id=$HITL_REQUEST_ID \
... <<'EOF'
{
  "site_type": "landing",
  "recommended_builder": "landing-page-builder"
}
EOF
```

### Brochure Site Confirmation
```bash
kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P -b kafka:9092 -t system.agent.generic.responses \
-H correlation_id=$CORRELATION_ID \
-H in_response_to_request_id=$HITL_REQUEST_ID \
... <<'EOF'
{
  "site_type": "brochure",
  "recommended_builder": "content-site-builder"
}
EOF
```

---

## Testing Scripts

### Landing Page Test
```bash
./start_landing_test.sh          # Start with landing objective
./send_landing_confirm.sh ...    # Confirm "landing" type
./send_landing_brief.sh ...      # Fill conversion questionnaire
```

### Brochure Site Test
```bash
./start_brochure_test.sh         # Start with brochure objective
./send_brochure_confirm.sh ...   # Confirm "brochure" type
./send_brochure_brief.sh ...     # Fill corporate questionnaire
```

---

## Common Issues

### Classifier Picked Wrong Type

**Scenario:** Brochure objective → classified as "landing"

**Solutions:**
1. ✅ Override via HITL: Send `"site_type": "brochure"` in confirmation
2. ✅ Improve objective wording: Add "multiple pages", "company sections"
3. ✅ Check classifier logs: See what signals it detected

### Wrong Builder Spawned

**Scenario:** Confirmed "brochure" but spawned landing-page-builder

**Solutions:**
1. ✅ Check HITL response: Ensure `"recommended_builder": "content-site-builder"`
2. ✅ Verify response matched: Look for "Found awaited request"
3. ✅ Check CollectedData: Should have `confirmed_type.recommended_builder`

### Questionnaire Doesn't Match

**Scenario:** Brochure builder but got landing questionnaire

**Solutions:**
1. ✅ Check fetch_questionnaire step: Should query correct builder
2. ✅ Verify builder type: `fetch_agent_questionnaire` uses `confirmed_type.recommended_builder`
3. ✅ Check response data: Should have brochure-specific fields

---

## Success Checklist

### Landing Page Path ✅
- [ ] Classifier suggests "landing"
- [ ] HITL confirms landing-page-builder
- [ ] Questionnaire has conversion-focused fields
- [ ] landing-page-builder spawned
- [ ] Single-page HTML generated

### Brochure Site Path ⏳ (To Test)
- [ ] Classifier suggests "brochure"
- [ ] HITL confirms content-site-builder
- [ ] Questionnaire has corporate/multi-page fields
- [ ] content-site-builder spawned
- [ ] Multi-page HTML structure generated

---

## Key Learning: HITL Enables Flexibility

The HITL pauses allow humans to:
1. **Override classifier:** Change "landing" → "brochure" if AI misclassified
2. **Select different builder:** Choose alternate builder even if classifier was right
3. **Customize questionnaire:** Fill in or modify generated data
4. **Quality control:** Review before expensive generation step

This makes the system robust to both AI errors and changing requirements.