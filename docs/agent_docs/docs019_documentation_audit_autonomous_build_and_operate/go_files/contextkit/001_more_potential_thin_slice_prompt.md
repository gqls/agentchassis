please start or update a runbook, a running notes and a plan. Please present the sql as migrations and always backup before running them e.g. using the snapshot_agent function if it's agent_definitions.

     | func

public | snapshot_agent                            | uuid                                                                                                                                                                                        | p_agent_type text                                                                                                                            | func

public | snapshot_agent                            | uuid                                                                                                                                                                                        | p_agent_type text, p_reason text DEFAULT NULL::text

public | take_site_snapshot                        | uuid                                                                                                                                                                                        | p_site_id uuid, p_trigger text, p_git_sha text DEFAULT NULL::text, p_label text DEFAULT NULL::text, p_created_by text DEFAULT 'system'::text | func


