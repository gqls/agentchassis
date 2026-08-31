-- ROLLBACK for 666: remove the two added paragraphs from the privacy policy.
BEGIN;
UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{content}',
         to_jsonb(regexp_replace(regexp_replace(pc.content_data->>'content',
           '<p><strong>Documents you send us for fine-tuning\.</strong>.*?</p>', '', 'g'),
           '<p><strong>Fine-tuning documents and models\.</strong>.*?</p>', '', 'g'))),
       updated_at = now()
  FROM pages p, sites s
 WHERE p.id=pc.page_id AND s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='privacy-policy';
COMMIT;
