-- Add the Part 26A mechanism diagram to the Thames Water case page.
--
-- EVERY step below paraphrases a fact that is already in this site's evidence
-- register, quote-verified, with a source. No statutory content is written from
-- memory: that is the exact class of claim the grounded-explainer audit caught
-- on this site once already (three unsourced legal generalisations, pre-publication).
-- Fact ids are carried in content_data.grounded_in so the trail survives.
--
-- The component is locked `permanent`. The page's rebuild_policy is 'generic',
-- so a later rebuild would otherwise hand cited legal text to a writer and
-- replace it with plausible prose. A per-component lock protects this section
-- without freezing the whole page, which is what setting the page to 'owned'
-- would have done.

BEGIN;

WITH pg AS (
  SELECT id FROM pages
  WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name = 'thames-water'
), comp AS (
  SELECT id FROM content_components WHERE name = 'mechanism-flow'
)
INSERT INTO page_components (
  page_id, component_id, position, slot_name, content_data,
  build_status, locked_at, locked_by, lock_type
)
SELECT pg.id, comp.id, 2, 'mechanism-flow',
$json${
  "eyebrow": "The mechanism",
  "section_title": "How a restructuring plan binds a class that voted against it",
  "intro": "Part 26A of the Companies Act 2006 lets a court sanction a plan over the objection of an entire class of creditors. The sequence below is the route from financial difficulty to a binding order, and the two conditions at steps four and five are where a contested plan is usually won or lost.",
  "steps": [
    {
      "title": "The company must be in financial difficulty",
      "body": "Part 26A is available to a company that has encountered, or is likely to encounter, financial difficulties affecting its ability to carry on business as a going concern. That gateway is tested before any question of classes or votes arises.",
      "note": "The statute calls this Condition A. A different Condition A appears later, at section 901G, governing cross-class cram down. The label is reused for two unrelated tests, and mixing them up is a common source of confusion in commentary."
    },
    {
      "title": "Everyone whose rights are affected is entitled to vote",
      "body": "Every creditor or member whose rights are affected by the compromise or arrangement must be permitted to participate in the meeting. Class composition therefore determines who has a say, and it is settled before the vote rather than contested after it."
    },
    {
      "title": "Each class votes, and the threshold is 75% by value",
      "body": "Section 901F(1) requires 75% in value of creditors or members present and voting at the meeting to agree the compromise or arrangement before the court may sanction it. The threshold is applied class by class, and value is counted among those who actually vote.",
      "branches": [
        {
          "label": "Every class clears 75%",
          "body": "The plan proceeds to sanction on the ordinary footing. The cram down provisions are never reached."
        },
        {
          "label": "A class falls short",
          "body": "Section 901G engages and that class becomes the dissenting class. The plan can still be sanctioned, and the next two conditions decide whether it is."
        }
      ]
    },
    {
      "title": "Condition A: no member of the dissenting class is worse off than under the relevant alternative",
      "body": "The court must be satisfied that no member of the dissenting class would be worse off than they would be under the relevant alternative.",
      "note": "The relevant alternative is whatever the court considers would be most likely to occur if the compromise or arrangement were not sanctioned. The test is anchored to a counterfactual the court selects, which is why so much of the contested valuation evidence in a Part 26A case is aimed at establishing what that counterfactual is."
    },
    {
      "title": "Condition B: a class with genuine economic interest has approved it",
      "body": "The plan must have been agreed by 75% in value of at least one class of creditors or members who would receive a payment, or who have a genuine economic interest in the relevant alternative. A class that recovers nothing in the counterfactual cannot supply that approval."
    },
    {
      "title": "If both conditions are met, the dissent does not prevent sanction",
      "body": "Where Conditions A and B are satisfied, a dissenting class's rejection does not prevent the court sanctioning the plan. This is the cross-class cram down, and it is the power that separates Part 26A from a scheme of arrangement under Part 26.",
      "note": "Meeting the conditions opens the court's discretion rather than settling it. In Re Nasmyth Group Ltd the High Court declined to sanction a plan opposed by HMRC on the basis that it was unfair. In Adler the Court of Appeal allowed the appeal and set aside the sanction order, where Snowden LJ held that the overall value of claims voting in favour across all classes is not a relevant factor in the discretion to cram down a dissenting class."
    },
    {
      "title": "A sanctioned plan binds everyone it covers",
      "body": "A sanctioned compromise or arrangement is binding on all the creditors, or the class of creditors, or the members or class of members, and on the company itself. The court may depart from the pari passu principle that would apply in the relevant alternative, but only where there is good justification.",
      "note": "That last point is the practical reason a plan is worth fighting over. Priority in the counterfactual is the starting position, not a guarantee."
    }
  ],
  "footnote": "Every statement above is drawn from this site's evidence register, where each is recorded against the source it was read from and the date it was read. That shows you where our reading came from; it does not prove the reading is right. This describes the mechanism in general and is not a description of any particular plan, and it is not legal advice.",
  "grounded_in": [
    "CIT-ab259475d2781337", "CIT-de7e9d8f9827c5ac", "CIT-a7f91f88754d560",
    "CIT-709026d97df11645", "CIT-917cae2baf2a9069", "CIT-f729a1b60a30481",
    "CIT-81d532ab22dcb359", "CIT-3cd41ecf235e9df9", "CIT-f31377f0424f08b7",
    "CIT-7cd6c88635c7ce16", "CIT-3a4297f974e5b5b8", "CIT-39cc4b6148d4db47",
    "CIT-c4ecd449365be098"
  ]
}$json$::jsonb,
'pending', now(), 'oufe-workstream', 'permanent'
FROM pg, comp;

-- Register the slot on the page. sections is the list the renderer walks; a
-- component absent from it renders nothing and still reports COMPLETED
-- (bugs_closed/095).
UPDATE pages
SET sections = (
      SELECT jsonb_agg(DISTINCT x)
      FROM jsonb_array_elements(COALESCE(sections, '[]'::jsonb) || '["mechanism-flow"]'::jsonb) x
    ),
    build_status = 'needs_rebuild'
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name = 'thames-water';

COMMIT;

SELECT p.name, p.sections, p.build_status,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id = p.id) AS components
FROM pages p
WHERE p.site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name = 'thames-water';
