-- ROLLBACK for 649: remove the image branch from hero-tool and case-studies-hero,
-- returning them to components that silently orphan any imagery filed for them.
BEGIN;
UPDATE content_components
   SET html_template = regexp_replace(html_template,
         '<section class="hero hero-case-studies\{\{if or \.hero_url \.background_image\}\}[^>]*>',
         '<section class="hero hero-case-studies" data-component="hero-case-studies">'),
       updated_at = now()
 WHERE name = 'case-studies-hero';
UPDATE content_components
   SET html_template = regexp_replace(html_template,
         '<section class="hero-tool-section\{\{if or \.hero_url \.background_image\}\}[^>]*>',
         '<section class="hero-tool-section" data-component="hero-tool">'),
       updated_at = now()
 WHERE name = 'hero-tool';
COMMIT;
