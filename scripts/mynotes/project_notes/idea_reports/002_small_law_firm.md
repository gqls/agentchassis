IDEA REPORT — a small legal-AI services firm
------------------------------------------------------------

        WHO IT'S FOR
        Regional law firms with their own documents but no ML staff
        
        WHY THEY'D PAY
        They pay because the bottleneck is turning their own messy, sensitive documents into something usable, which they can't safely paste into a public chatbot.
        
        ADVANCING IDEAS (best first)
        
        1) Drawing/Form Reader to Structured Data  [consider]
           Idea:       reads a customer's drawings and returns validated, structured fields
           Built on:   the customer's own engineering drawings, using tuned vision reading with schema-checked output.
           Checks out: Checked that tools like this exist and have matured in the last year; the edge is tuning on the customer's own drawing conventions, which generic tools miss.
           Scores:     Defensibility 4/5, Willingness 4/5, Buildability 3/5, Reuse 3/5, Durability 3/5 (total 18/25).
           Risk:       4/5 (low — refunds make customers whole)
           First test: Ask one engineering prospect for 20 anonymised drawings, run a plain pass against a 5-example tuned pass, and show them the accuracy gap.
        
        2) Firm-Voice Drafting Model  [consider]
           Note: short-lived — base-model progress may erode this.
           Note: needs liability work before building (risk 2/5).
           Idea:       drafts in the firm's actual style and clause preferences, not a generic template
           Built on:   the firm's past letters and contracts, using a private model tuned on that house style.
           Checks out: A paying market for legal drafting tools is real (several priced per user per month). A true tune on a firm's own past work is the defensible part.
           Scores:     Defensibility 4/5, Willingness 5/5, Buildability 3/5, Reuse 3/5, Durability 2/5 (total 17/25).
           Risk:       2/5 (high — needs review, insurance, tight T&Cs before building)
           First test: Validate demand first; do not build until PII insurance is in force and T&Cs are reviewed by a UK solicitor. Then ask three firms whether they'd pay a monthly per-seat fee for a model trained on their own drafts.
        
        ------------------------------------------------------------
        DIDN'T MAKE THE CUT (not hard enough to copy, or too little willingness to pay)
           - Stale-Doc Drift Detector (Defensibility 2/5, Willingness 4/5)
        
        ------------------------------------------------------------
        SET ASIDE ON RISK (regulated-profession territory or similar)
        These may be real opportunities, but we can't build them safely without the
        right qualifications and cover, so they're here for awareness, not as advice.
           - Symptom Triage Assistant (Risk 1/5; Defensibility 4/5, Willingness 5/5, Buildability 3/5)
--- PASS: TestRenderReadable (0.00s)