CREATE OR REPLACE FUNCTION public.auto_lock_on_deploy()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
    -- Only act if changing to 'deployed' status
    IF NEW.build_status = 'deployed' AND OLD.build_status != 'deployed' THEN
        -- Check if site is configured for first_deploy locking
        IF EXISTS (
            SELECT 1 FROM pages p
            JOIN sites s ON p.site_id = s.id
            WHERE p.id = NEW.page_id
            AND s.strict_mode_trigger = 'first_deploy'
        ) THEN
            -- Lock to strict mode if not already locked
            IF NEW.schema_mode IS NULL OR NEW.schema_mode = 'flexible' THEN
                NEW.schema_mode := 'strict';
                NEW.locked_at := now();
                NEW.locked_by := 'first_deploy';
                -- Note: schema_snapshot and content_snapshot should be set before deploy
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$function$

