UPDATE sites
SET content_data = content_data || '{"contact_phone": "+44 (0) 7934 524 911"}'::jsonb
WHERE domain = 'leopardessconsulting.co.uk';

