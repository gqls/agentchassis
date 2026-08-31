-- ROLLBACK for 665: remove the fine-tuning service section from the terms, restoring the
-- lock exactly as 665 does (all four columns, original provenance).
BEGIN;
UPDATE page_components pc SET locked_at=NULL, locked_by=NULL, lock_type=NULL, lock_expires_at=NULL
  FROM pages p, sites s WHERE p.id=pc.page_id AND s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='terms';
UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{content}',
         to_jsonb(regexp_replace(pc.content_data->>'content',
           '<h2>The fine-tuning service</h2>.*?(?=<h2>Intellectual property</h2>)', '', 'g'))),
       updated_at = now()
  FROM pages p, sites s WHERE p.id=pc.page_id AND s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='terms';
UPDATE page_components pc
   SET locked_at=TIMESTAMPTZ '2026-07-21 09:21:45.96136+00', locked_by='182_legal_pages',
       lock_type='permanent', lock_expires_at=NULL
  FROM pages p, sites s WHERE p.id=pc.page_id AND s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='terms';
COMMIT;
