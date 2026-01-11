public | affiliate_products      | table             | clients_user
 public | affiliate_programs      | table             | clients_user
 public | agent_capabilities      | table             | clients_user
 public | agent_default_configs   | table             | clients_user
 public | agent_definitions       | table             | clients_user
 public | agent_dependencies      | table             | clients_user
 public | agent_group_definitions | table             | clients_user
 public | agent_group_members     | table             | clients_user
 public | agent_groups            | table             | clients_user
 public | agent_metrics           | table             | clients_user
 public | agent_metrics_config    | table             | clients_user
 public | approval_requests       | table             | clients_user
 public | assets                  | table             | clients_user
 public | awaited_requests        | table             | clients_user
 public | clients                 | table             | clients_user
 public | content_components      | table             | clients_user
 public | content_items           | table             | clients_user
 public | css_snippets            | table             | clients_user
 public | css_themes              | table             | clients_user
 public | entity_state_log        | table             | clients_user
 public | entity_state_log_id_seq | sequence          | clients_user
 public | event_statistics        | materialized view | clients_user
 public | flow_pages              | table             | clients_user
 public | improvement_proposals   | table             | clients_user
 public | input_requests          | table             | clients_user
 public | js_snippets             | table             | clients_user
 public | link_registry           | table             | clients_user
 public | navigation_structures   | table             | clients_user
 public | networks                | table             | clients_user
 public | orchestration_requests  | table             | clients_user
 public | orchestration_states    | table             | clients_user
 public | page_components         | table             | clients_user
 public | pages                   | table             | clients_user
 public | pending_input_requests  | view              | clients_user
 public | pending_requests        | table             | clients_user
 public | processed_messages      | table             | clients_user
 public | product_assets          | table             | clients_user
 public | products                | table             | clients_user
 public | recent_agent_events     | view              | clients_user
 public | redirects               | table             | clients_user
 public | relationships           | table             | clients_user
 public | research_results        | table             | clients_user
 public | site_flows              | table             | clients_user
 public | sites                   | table             | clients_user
 public | style_collections       | table             | clients_user
 public | system_events           | table             | clients_user
 public | theme_tags              | table             | clients_user
 public | v_active_workflows      | view              | clients_user
 public | v_agents_by_category    | view              | clients_user
 public | v_all_workflows         | view              | clients_user
 public | v_content_usage         | view              | clients_user
 public | v_navigation_pages      | view              | clients_user
 public | v_page_build_status     | view              | clients_user
 public | v_research_summary      | view              | clients_user
 public | v_site_assets           | view              | clients_user
 public | v_site_links            | view              | clients_user
 public | workflow_contract_chain | view              | clients_user
 public | workflow_templates      | table             | clients_user

e.g. these tables
content_components
content_items
css_snippets
css_themes
js_snippets
page_components
pages
 product_assets
products
style_collections
theme_tags
sites
site_flows
relationships

-----

clients_db=# \d content_components
                                Table "public.content_components"
      Column      |           Type           | Collation | Nullable |          Default
------------------+--------------------------+-----------+----------+----------------------------
 id               | uuid                     |           | not null | gen_random_uuid()
 name             | text                     |           | not null |
 description      | text                     |           |          |
 html_template    | text                     |           | not null |
 input_schema     | jsonb                    |           |          |
 function         | text                     |           | not null | 'generic-text-block'::text
 created_at       | timestamp with time zone |           |          | now()
 updated_at       | timestamp with time zone |           |          | now()
 display_name     | text                     |           |          | ''::text
 category         | character varying(255)   |           |          | ''::character varying
 semantic_tags    | jsonb                    |           |          |
 sort_order       | integer                  |           |          |
 render_mode      | text                     |           |          | 'template'::text
 agent_type       | text                     |           |          |
 agent_workflow   | text                     |           |          |
 data_sources     | text[]                   |           |          |
 child_components | jsonb                    |           |          |
 component_level  | text                     |           |          | 'section'::text
 is_active        | boolean                  |           |          | true
Indexes:
    "content_components_pkey" PRIMARY KEY, btree (id)
    "content_components_name_key" UNIQUE CONSTRAINT, btree (name)
    "idx_components_agent_type" btree (agent_type) WHERE agent_type IS NOT NULL
    "idx_components_level" btree (component_level)
    "idx_components_render_mode" btree (render_mode)
    "idx_content_components_category" btree (category)
    "idx_content_components_function" btree (function)
Referenced by:
    TABLE "page_components" CONSTRAINT "page_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_component_id_fkey" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_fk" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_component_id_fkey" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_fk" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_component_id_fkey" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_fk" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)

--

clients_db=# \d content_items
                                   Table "public.content_items"
       Column       |           Type           | Collation | Nullable |          Default
--------------------+--------------------------+-----------+----------+----------------------------
 id                 | uuid                     |           | not null | gen_random_uuid()
 site_id            | uuid                     |           |          |
 content_type       | character varying(100)   |           | not null |
 content_key        | character varying(255)   |           |          |
 content_data       | jsonb                    |           | not null | '{}'::jsonb
 plain_text         | text                     |           |          |
 is_library         | boolean                  |           |          | false
 library_tags       | text[]                   |           |          |
 industry_vertical  | character varying(100)   |           |          |
 origin_type        | text                     |           | not null | 'generated'::text
 origin_agent       | text                     |           |          |
 origin_research_id | uuid                     |           |          |
 origin_content_id  | uuid                     |           |          |
 status             | character varying(50)    |           |          | 'draft'::character varying
 version            | integer                  |           |          | 1
 created_at         | timestamp with time zone |           |          | now()
 updated_at         | timestamp with time zone |           |          | now()
 approved_at        | timestamp with time zone |           |          |
 approved_by        | character varying(100)   |           |          |
Indexes:
    "content_items_pkey" PRIMARY KEY, btree (id)
    "idx_content_items_key" btree (site_id, content_key)
    "idx_content_items_library" btree (is_library, industry_vertical) WHERE is_library = true
    "idx_content_items_search" gin (to_tsvector('english'::regconfig, plain_text))
    "idx_content_items_site" btree (site_id)
    "idx_content_items_status" btree (status)
    "idx_content_items_type" btree (content_type)
    "idx_content_items_unique_key" UNIQUE, btree (site_id, content_key) WHERE site_id IS NOT NULL AND content_key IS NOT NULL
Foreign-key constraints:
    "content_items_origin_content_id_fkey" FOREIGN KEY (origin_content_id) REFERENCES content_items(id)
    "content_items_origin_research_id_fkey" FOREIGN KEY (origin_research_id) REFERENCES research_results(id) ON DELETE SET NULL
    "content_items_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:
    TABLE "content_items" CONSTRAINT "content_items_origin_content_id_fkey" FOREIGN KEY (origin_content_id) REFERENCES content_items(id)
    TABLE "page_components" CONSTRAINT "page_components_content_item_id_fkey" FOREIGN KEY (content_item_id) REFERENCES content_items(id) ON DELETE SET NULL


--
\d content_items
                                   Table "public.content_items"
       Column       |           Type           | Collation | Nullable |          Default
--------------------+--------------------------+-----------+----------+----------------------------
 id                 | uuid                     |           | not null | gen_random_uuid()
 site_id            | uuid                     |           |          |
 content_type       | character varying(100)   |           | not null |
 content_key        | character varying(255)   |           |          |
 content_data       | jsonb                    |           | not null | '{}'::jsonb
 plain_text         | text                     |           |          |
 is_library         | boolean                  |           |          | false
 library_tags       | text[]                   |           |          |
 industry_vertical  | character varying(100)   |           |          |
 origin_type        | text                     |           | not null | 'generated'::text
 origin_agent       | text                     |           |          |
 origin_research_id | uuid                     |           |          |
 origin_content_id  | uuid                     |           |          |
 status             | character varying(50)    |           |          | 'draft'::character varying
 version            | integer                  |           |          | 1
 created_at         | timestamp with time zone |           |          | now()
 updated_at         | timestamp with time zone |           |          | now()
 approved_at        | timestamp with time zone |           |          |
 approved_by        | character varying(100)   |           |          |
Indexes:
    "content_items_pkey" PRIMARY KEY, btree (id)
    "idx_content_items_key" btree (site_id, content_key)
    "idx_content_items_library" btree (is_library, industry_vertical) WHERE is_library = true
    "idx_content_items_search" gin (to_tsvector('english'::regconfig, plain_text))
    "idx_content_items_site" btree (site_id)
    "idx_content_items_status" btree (status)
    "idx_content_items_type" btree (content_type)
    "idx_content_items_unique_key" UNIQUE, btree (site_id, content_key) WHERE site_id IS NOT NULL AND content_key IS NOT NULL
Foreign-key constraints:
    "content_items_origin_content_id_fkey" FOREIGN KEY (origin_content_id) REFERENCES content_items(id)
    "content_items_origin_research_id_fkey" FOREIGN KEY (origin_research_id) REFERENCES research_results(id) ON DELETE SET NULL
    "content_items_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:
    TABLE "content_items" CONSTRAINT "content_items_origin_content_id_fkey" FOREIGN KEY (origin_content_id) REFERENCES content_items(id)
    TABLE "page_components" CONSTRAINT "page_components_content_item_id_fkey" FOREIGN KEY (content_item_id) REFERENCES content_items(id) ON DELETE SET NULL

clients_db=# \d css_snippets
                              Table "public.css_snippets"
    Column     |            Type             | Collation | Nullable |      Default
---------------+-----------------------------+-----------+----------+-------------------
 id            | uuid                        |           | not null | gen_random_uuid()
 name          | character varying(100)      |           | not null |
 description   | text                        |           |          |
 css_content   | text                        |           | not null |
 semantic_tags | jsonb                       |           |          | '[]'::jsonb
 applies_to    | jsonb                       |           |          | '[]'::jsonb
 created_at    | timestamp without time zone |           |          | now()
Indexes:
    "css_snippets_pkey" PRIMARY KEY, btree (id)
    "css_snippets_name_key" UNIQUE CONSTRAINT, btree (name)


--

\d css_themes
                              Table "public.css_themes"
    Column     |           Type           | Collation | Nullable |      Default
---------------+--------------------------+-----------+----------+-------------------
 id            | uuid                     |           | not null | gen_random_uuid()
 name          | text                     |           | not null |
 display_name  | text                     |           | not null |
 description   | text                     |           |          |
 category      | text                     |           |          |
 css_content   | text                     |           | not null |
 version       | integer                  |           |          | 1
 is_active     | boolean                  |           |          | true
 created_at    | timestamp with time zone |           |          | now()
 updated_at    | timestamp with time zone |           |          | now()
 semantic_tags | text[]                   |           |          |
 color_palette | jsonb                    |           |          |
 typography    | jsonb                    |           |          |
Indexes:
    "css_themes_pkey" PRIMARY KEY, btree (id)
    "css_themes_name_key" UNIQUE CONSTRAINT, btree (name)
    "idx_css_themes_category" btree (category)
    "idx_css_themes_name" btree (name)
Referenced by:
    TABLE "style_collections" CONSTRAINT "style_collections_css_theme_id_fkey" FOREIGN KEY (css_theme_id) REFERENCES css_themes(id)


--

clients_db=# \d js_snippets
                               Table "public.js_snippets"
    Column     |            Type             | Collation | Nullable |      Default
---------------+-----------------------------+-----------+----------+-------------------
 id            | uuid                        |           | not null | gen_random_uuid()
 name          | character varying(100)      |           | not null |
 description   | text                        |           |          |
 js_content    | text                        |           | not null |
 semantic_tags | jsonb                       |           |          | '[]'::jsonb
 applies_to    | jsonb                       |           |          | '[]'::jsonb
 dependencies  | jsonb                       |           |          | '[]'::jsonb
 created_at    | timestamp without time zone |           |          | now()
Indexes:
    "js_snippets_pkey" PRIMARY KEY, btree (id)
    "js_snippets_name_key" UNIQUE CONSTRAINT, btree (name)


--
 Table "public.page_components"
       Column       |           Type           | Collation | Nullable |      Default
--------------------+--------------------------+-----------+----------+-------------------
 id                 | uuid                     |           | not null | gen_random_uuid()
 page_id            | uuid                     |           |          |
 component_id       | uuid                     |           |          |
 position           | integer                  |           | not null |
 slot_name          | character varying(100)   |           |          |
 parent_instance_id | uuid                     |           |          |
 rendered_html      | text                     |           |          |
 content_data       | jsonb                    |           |          |
 content_hash       | character varying(64)    |           |          |
 data_path          | character varying(500)   |           |          |
 data_uuid          | uuid                     |           |          | gen_random_uuid()
 created_at         | timestamp with time zone |           |          | now()
 updated_at         | timestamp with time zone |           |          | now()
 build_status       | text                     |           |          | 'pending'::text
 reviewed_at        | timestamp with time zone |           |          |
 reviewed_by        | text                     |           |          |
 deploy_commit      | text                     |           |          |
 research_id        | uuid                     |           |          |
 sources_displayed  | boolean                  |           |          | false
 content_item_id    | uuid                     |           |          |
Indexes:
    "page_components_pkey" PRIMARY KEY, btree (id)
    "idx_page_components_content" btree (content_item_id)
    "idx_page_components_page" btree (page_id)
    "idx_page_components_parent" btree (parent_instance_id)
    "idx_page_components_status" btree (build_status)
    "idx_page_components_template" btree (component_id)
Foreign-key constraints:
    "page_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    "page_components_content_item_id_fkey" FOREIGN KEY (content_item_id) REFERENCES content_items(id) ON DELETE SET NULL
    "page_components_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
    "page_components_parent_instance_id_fkey" FOREIGN KEY (parent_instance_id) REFERENCES page_components(id)
    "page_components_research_id_fkey" FOREIGN KEY (research_id) REFERENCES research_results(id) ON DELETE SET NULL
Referenced by:

--

 Table "public.pages"
      Column      |           Type           | Collation | Nullable |           Default
------------------+--------------------------+-----------+----------+-----------------------------
 id               | uuid                     |           | not null | gen_random_uuid()
 site_id          | uuid                     |           |          |
 name             | character varying(255)   |           | not null |
 url              | character varying(500)   |           | not null |
 title            | character varying(500)   |           |          |
 page_type        | character varying(50)    |           |          |
 status           | character varying(50)    |           |          | 'active'::character varying
 content_hash     | character varying(64)    |           |          |
 meta_description | text                     |           |          |
 topics           | text[]                   |           |          |
 nav_label        | character varying(255)   |           |          |
 nav_order        | integer                  |           |          | 100
 in_header        | boolean                  |           |          | true
 in_footer        | boolean                  |           |          | true
 last_built_at    | timestamp with time zone |           |          |
 expires_at       | timestamp with time zone |           |          |
 created_at       | timestamp with time zone |           |          | now()
 updated_at       | timestamp with time zone |           |          | now()
 build_status     | text                     |           |          | 'pending'::text
 deployed_at      | timestamp with time zone |           |          |
 sections         | jsonb                    |           |          | '[]'::jsonb
 version          | integer                  |           |          | 1
Indexes:
    "pages_pkey" PRIMARY KEY, btree (id)
    "idx_pages_build_status" btree (site_id, build_status)
    "idx_pages_nav" btree (site_id, in_header, nav_order) WHERE status::text = 'active'::text
    "idx_pages_needs_build" btree (site_id) WHERE build_status = ANY (ARRAY['planned'::text, 'needs_rebuild'::text])
    "idx_pages_site" btree (site_id)
    "idx_pages_status" btree (status)
    "idx_pages_type" btree (page_type)
    "pages_site_id_name_key" UNIQUE CONSTRAINT, btree (site_id, name)
Foreign-key constraints:
    "pages_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:


--

\d product_assets
                          Table "public.product_assets"
   Column   |           Type           | Collation | Nullable |      Default
------------+--------------------------+-----------+----------+-------------------
 id         | uuid                     |           | not null | gen_random_uuid()
 product_id | uuid                     |           | not null |
 asset_id   | uuid                     |           | not null |
 position   | integer                  |           |          | 0
 is_primary | boolean                  |           |          | false
 asset_role | text                     |           |          |
 created_at | timestamp with time zone |           |          | now()
Indexes:
    "product_assets_pkey" PRIMARY KEY, btree (id)
    "idx_product_assets_asset" btree (asset_id)
    "idx_product_assets_product" btree (product_id)
    "product_assets_product_id_asset_id_key" UNIQUE CONSTRAINT, btree (product_id, asset_id)
Foreign-key constraints:
    "product_assets_asset_id_fkey" FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
    "product_assets_product_id_fkey" FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE


--

                                 Table "public.products"
      Column       |           Type           | Collation | Nullable |      Default
-------------------+--------------------------+-----------+----------+-------------------
 id                | uuid                     |           | not null | gen_random_uuid()
 site_id           | uuid                     |           |          |
 name              | text                     |           | not null |
 slug              | text                     |           | not null |
 sku               | text                     |           |          |
 short_description | text                     |           |          |
 description       | text                     |           |          |
 features          | jsonb                    |           |          |
 specifications    | jsonb                    |           |          |
 price             | numeric(10,2)            |           |          |
 compare_at_price  | numeric(10,2)            |           |          |
 currency          | text                     |           |          | 'GBP'::text
 price_display     | text                     |           |          |
 category          | text                     |           |          |
 subcategory       | text                     |           |          |
 tags              | text[]                   |           |          |
 meta_title        | text                     |           |          |
 meta_description  | text                     |           |          |
 content_data      | jsonb                    |           |          | '{}'::jsonb
 status            | text                     |           |          | 'draft'::text
 published_at      | timestamp with time zone |           |          |
 created_at        | timestamp with time zone |           |          | now()
 updated_at        | timestamp with time zone |           |          | now()
Indexes:
    "products_pkey" PRIMARY KEY, btree (id)
    "idx_products_category" btree (category)
    "idx_products_site" btree (site_id)
    "idx_products_status" btree (status)
    "idx_products_tags" gin (tags)
    "products_site_id_slug_key" UNIQUE CONSTRAINT, btree (site_id, slug)
Foreign-key constraints:
    "products_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:
    TABLE "product_assets" CONSTRAINT "product_assets_product_id_fkey" FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE


--

                                                                                                 Table "public.style_collections"
          Column          |           Type           | Collation | Nullable |                                                                                    Default
--------------------------+--------------------------+-----------+----------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 id                       | uuid                     |           | not null | gen_random_uuid()
 name                     | text                     |           | not null |
 display_name             | text                     |           | not null |
 description              | text                     |           |          |
 header_component_id      | uuid                     |           |          |
 header_home_component_id | uuid                     |           |          |
 footer_component_id      | uuid                     |           |          |
 css_theme_id             | uuid                     |           |          |
 color_palette            | jsonb                    |           |          | '{"text": "#333333", "accent": "#16a085", "primary": "#1a1a2e", "secondary": "#2d2d44", "background": "#ffffff", "text_light": "#666666", "background_alt": "#f8f9fa"}'::jsonb
 typography               | jsonb                    |           |          | '{"base_size": "16px", "font_family": "-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif", "line_height": "1.6", "heading_font": "inherit"}'::jsonb
 category                 | text                     |           |          |
 industry_tags            | text[]                   |           |          |
 is_active                | boolean                  |           |          | true
 created_at               | timestamp with time zone |           |          | now()
 updated_at               | timestamp with time zone |           |          | now()
Indexes:
    "style_collections_pkey" PRIMARY KEY, btree (id)
    "idx_style_collections_category" btree (category)
    "idx_style_collections_industry" gin (industry_tags)
    "style_collections_name_key" UNIQUE CONSTRAINT, btree (name)
Foreign-key constraints:
    "style_collections_css_theme_id_fkey" FOREIGN KEY (css_theme_id) REFERENCES css_themes(id)
    "style_collections_footer_component_id_fkey" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    "style_collections_footer_fk" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    "style_collections_header_component_id_fkey" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    "style_collections_header_fk" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    "style_collections_header_home_component_id_fkey" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
    "style_collections_header_home_fk" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
Referenced by:
    TABLE "sites" CONSTRAINT "sites_style_collection_id_fkey" FOREIGN KEY (style_collection_id) REFERENCES style_collections(id)


clients_db=# \d theme_tags
                               Table "public.theme_tags"
    Column    |            Type             | Collation | Nullable |      Default
--------------+-----------------------------+-----------+----------+-------------------
 id           | uuid                        |           | not null | gen_random_uuid()
 name         | character varying(50)       |           | not null |
 category     | character varying(30)       |           | not null |
 description  | text                        |           |          |
 related_tags | jsonb                       |           |          | '[]'::jsonb
 created_at   | timestamp without time zone |           |          | now()
Indexes:
    "theme_tags_pkey" PRIMARY KEY, btree (id)
    "theme_tags_name_key" UNIQUE CONSTRAINT, btree (name)


--

Table "public.sites"
       Column        |           Type           | Collation | Nullable |           Default
---------------------+--------------------------+-----------+----------+-----------------------------
 id                  | uuid                     |           | not null | gen_random_uuid()
 network_id          | uuid                     |           |          |
 domain              | character varying(255)   |           | not null |
 name                | character varying(255)   |           |          |
 brand_dna           | jsonb                    |           |          | '{}'::jsonb
 github_repo         | character varying(500)   |           |          |
 github_branch       | character varying(100)   |           |          | 'main'::character varying
 settings            | jsonb                    |           |          | '{}'::jsonb
 status              | character varying(50)    |           |          | 'active'::character varying
 last_built_at       | timestamp with time zone |           |          |
 last_deployed_at    | timestamp with time zone |           |          |
 created_at          | timestamp with time zone |           |          | now()
 updated_at          | timestamp with time zone |           |          | now()
 style_collection_id | uuid                     |           |          |
 style_overrides     | jsonb                    |           |          | '{}'::jsonb
 content_data        | jsonb                    |           |          | '{}'::jsonb
 brand_assets        | jsonb                    |           |          | '{}'::jsonb
 default_components  | jsonb                    |           |          | '{}'::jsonb
 deploy_config       | jsonb                    |           |          | '{}'::jsonb
 build_status        | text                     |           |          | 'pending'::text
Indexes:
    "sites_pkey" PRIMARY KEY, btree (id)
    "idx_sites_build_status" btree (build_status)
    "idx_sites_content" gin (content_data)
    "idx_sites_domain" btree (domain)
    "idx_sites_network" btree (network_id)
    "idx_sites_status" btree (status)
    "idx_sites_style_collection" btree (style_collection_id)
    "sites_domain_key" UNIQUE CONSTRAINT, btree (domain)
Foreign-key constraints:
    "sites_network_id_fkey" FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE CASCADE
    "sites_style_collection_id_fkey" FOREIGN KEY (style_collection_id) REFERENCES style_collections(id)
Referenced by:
    TABLE "affiliate_products" CONSTRAINT "affiliate_products_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCA

    --

\d site_flows
                               Table "public.site_flows"
      Column      |           Type           | Collation | Nullable |      Default
------------------+--------------------------+-----------+----------+-------------------
 id               | uuid                     |           | not null | gen_random_uuid()
 site_id          | uuid                     |           |          |
 name             | character varying(255)   |           | not null |
 slug             | character varying(100)   |           | not null |
 audience_segment | character varying(255)   |           |          |
 narrative_arc    | jsonb                    |           |          |
 entry_points     | text[]                   |           |          |
 success_metric   | text                     |           |          |
 voice_parameters | jsonb                    |           |          | '{}'::jsonb
 is_default       | boolean                  |           |          | false
 created_at       | timestamp with time zone |           |          | now()
 updated_at       | timestamp with time zone |           |          | now()
Indexes:
    "site_flows_pkey" PRIMARY KEY, btree (id)
    "idx_flows_site" btree (site_id)
    "site_flows_site_id_slug_key" UNIQUE CONSTRAINT, btree (site_id, slug)
Foreign-key constraints:
    "site_flows_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:
    TABLE "flow_pages" CONSTRAINT "flow_pages_flow_id_fkey" FOREIGN KEY (flow_id) REFERENCES site_flows(id) ON DELETE CASCADE


-

clients_db=# \d relationships
                                      Table "public.relationships"
       Column       |            Type             | Collation | Nullable |           Default
--------------------+-----------------------------+-----------+----------+------------------------------
 id                 | uuid                        |           | not null | gen_random_uuid()
 source_entity_id   | character varying(255)      |           | not null |
 source_entity_type | character varying(100)      |           | not null |
 target_entity_id   | character varying(255)      |           | not null |
 target_entity_type | character varying(100)      |           | not null |
 relationship_type  | character varying(100)      |           | not null |
 direction          | character varying(20)       |           |          | 'one_way'::character varying
 properties         | jsonb                       |           |          | '{}'::jsonb
 status             | character varying(50)       |           |          | 'active'::character varying
 created_at         | timestamp without time zone |           |          | now()
 updated_at         | timestamp without time zone |           |          | now()
 ended_at           | timestamp without time zone |           |          |
Indexes:
    "relationships_pkey" PRIMARY KEY, btree (id)
    "idx_relationships_active" btree (source_entity_id, target_entity_id) WHERE status::text = 'active'::text
    "idx_relationships_pages" btree (source_entity_id, target_entity_id) WHERE source_entity_type::text = 'page'::text AND target_entity_type::text = 'page'::text
    "idx_relationships_source" btree (source_entity_id, source_entity_type)
    "idx_relationships_target" btree (target_entity_id, target_entity_type)
    "idx_relationships_type" btree (relationship_type)
    "unique_relationship" UNIQUE CONSTRAINT, btree (source_entity_id, source_entity_type, target_entity_id, target_entity_type, relationship_type)



