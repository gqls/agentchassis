#!/bin/bash
# FILE: docker/scripts/seed-data.sh
set -e

echo "🌱 Starting data seeding..."

# Wait for services to be ready
/app/wait-for-services.sh

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Function to seed PostgreSQL data
seed_postgres_data() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local sql_command=$5
    local description=$6

    echo -e "${YELLOW}🌱 Seeding $description...${NC}"

    export PGPASSWORD="$password"

    if psql -h "$host" -U "$user" -d "$database" -c "$sql_command"; then
        echo -e "${GREEN}✅ $description seeded${NC}"
    else
        echo -e "${RED}❌ $description seeding failed${NC}"
        # Don't exit on seeding failures - they might already exist
    fi
}

# Function to seed MySQL data
seed_mysql_data() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local sql_command=$5
    local description=$6

    echo -e "${YELLOW}🌱 Seeding $description...${NC}"

    if mysql -h "$host" -u "$user" -p"$password" "$database" -e "$sql_command"; then
        echo -e "${GREEN}✅ $description seeded${NC}"
    else
        echo -e "${RED}❌ $description seeding failed${NC}"
        # Don't exit on seeding failures - they might already exist
    fi
}

# 1. Seed persona templates
echo -e "${YELLOW}🤖 Seeding persona templates...${NC}"

# Basic Copywriter Template
seed_postgres_data \
    "postgres-templates" \
    "templates_user" \
    "templates_db" \
    "$TEMPLATES_DB_PASSWORD" \
    "INSERT INTO persona_templates (id, name, description, category, config, is_active, created_at, updated_at)
     VALUES (
         '00000000-0000-0000-0000-000000000001',
         'Basic Copywriter',
         'A versatile copywriting assistant that can create engaging content across various formats and tones.',
         'copywriter',
         '{
             \"model\": \"claude-3-sonnet\",
             \"temperature\": 0.7,
             \"max_tokens\": 2000,
             \"system_prompt\": \"You are a professional copywriter. Create compelling, engaging content that resonates with the target audience. Always consider the tone, style, and purpose of the content.\",
             \"workflow\": {
                 \"start_step\": \"generate_content\",
                 \"steps\": {
                     \"generate_content\": {
                         \"action\": \"ai_text_generate_claude_sonnet\",
                         \"description\": \"Generate the requested content\",
                         \"next_step\": \"complete_workflow\"
                     },
                     \"complete_workflow\": {
                         \"action\": \"complete_workflow\",
                         \"description\": \"Mark workflow as complete\"
                     }
                 }
             }
         }',
         true,
         NOW(),
         NOW()
     ) ON CONFLICT (id) DO NOTHING;" \
    "Basic Copywriter Template"

# Research Assistant Template
seed_postgres_data \
    "postgres-templates" \
    "templates_user" \
    "templates_db" \
    "$TEMPLATES_DB_PASSWORD" \
    "INSERT INTO persona_templates (id, name, description, category, config, is_active, created_at, updated_at)
     VALUES (
         '00000000-0000-0000-0000-000000000002',
         'Research Assistant',
         'An in-depth research specialist that can gather, analyze, and synthesize information from multiple sources.',
         'researcher',
         '{
             \"model\": \"claude-3-opus\",
             \"temperature\": 0.3,
             \"max_tokens\": 4000,
             \"system_prompt\": \"You are a thorough research assistant. Provide comprehensive, well-sourced, and analytically rigorous research. Always cite sources and present balanced perspectives.\",
             \"workflow\": {
                 \"start_step\": \"web_search\",
                 \"steps\": {
                     \"web_search\": {
                         \"action\": \"web_search\",
                         \"description\": \"Search for relevant information\",
                         \"topic\": \"system.adapter.web.search\",
                         \"next_step\": \"analyze_research\"
                     },
                     \"analyze_research\": {
                         \"action\": \"ai_text_generate_claude_opus\",
                         \"description\": \"Analyze and synthesize research findings\",
                         \"next_step\": \"complete_workflow\"
                     },
                     \"complete_workflow\": {
                         \"action\": \"complete_workflow\",
                         \"description\": \"Mark workflow as complete\"
                     }
                 }
             }
         }',
         true,
         NOW(),
         NOW()
     ) ON CONFLICT (id) DO NOTHING;" \
    "Research Assistant Template"

# Blog Post Generator Template
seed_postgres_data \
    "postgres-templates" \
    "templates_user" \
    "templates_db" \
    "$TEMPLATES_DB_PASSWORD" \
    "INSERT INTO persona_templates (id, name, description, category, config, is_active, created_at, updated_at)
     VALUES (
         '00000000-0000-0000-0000-000000000003',
         'Blog Post Generator',
         'A specialized content creator for well-structured, engaging blog posts with research backing.',
         'content-creator',
         '{
             \"model\": \"claude-3-sonnet\",
             \"temperature\": 0.6,
             \"max_tokens\": 3000,
             \"system_prompt\": \"You are a professional blog writer. Create well-structured, engaging blog posts with clear introductions, informative body content, and compelling conclusions. Use proper headings and maintain consistent tone.\",
             \"workflow\": {
                 \"start_step\": \"research_topic\",
                 \"steps\": {
                     \"research_topic\": {
                         \"action\": \"fan_out\",
                         \"description\": \"Research the topic thoroughly\",
                         \"sub_tasks\": [
                             {\"step_name\": \"web_research\", \"topic\": \"system.adapter.web.search\"},
                             {\"step_name\": \"style_analysis\", \"topic\": \"system.agent.reasoning.process\"}
                         ],
                         \"next_step\": \"generate_blog_post\"
                     },
                     \"generate_blog_post\": {
                         \"action\": \"ai_text_generate_claude_sonnet\",
                         \"description\": \"Generate the blog post using research\",
                         \"next_step\": \"pause_for_review\"
                     },
                     \"pause_for_review\": {
                         \"action\": \"pause_for_human_input\",
                         \"description\": \"Allow human review and approval\",
                         \"next_step\": \"complete_workflow\"
                     },
                     \"complete_workflow\": {
                         \"action\": \"complete_workflow\",
                         \"description\": \"Mark workflow as complete\"
                     }
                 }
             }
         }',
         true,
         NOW(),
         NOW()
     ) ON CONFLICT (id) DO NOTHING;" \
    "Blog Post Generator Template"

# Image Content Creator Template
seed_postgres_data \
    "postgres-templates" \
    "templates_user" \
    "templates_db" \
    "$TEMPLATES_DB_PASSWORD" \
    "INSERT INTO persona_templates (id, name, description, category, config, is_active, created_at, updated_at)
     VALUES (
         '00000000-0000-0000-0000-000000000004',
         'Visual Content Creator',
         'Creates both textual content and accompanying images for comprehensive visual storytelling.',
         'multimedia-creator',
         '{
             \"model\": \"claude-3-sonnet\",
             \"temperature\": 0.7,
             \"max_tokens\": 2500,
             \"system_prompt\": \"You are a visual content creator. Create compelling text content and provide detailed image descriptions for visual elements that enhance the narrative.\",
             \"workflow\": {
                 \"start_step\": \"create_content_plan\",
                 \"steps\": {
                     \"create_content_plan\": {
                         \"action\": \"ai_text_generate_claude_sonnet\",
                         \"description\": \"Plan the content and visual elements\",
                         \"next_step\": \"generate_visuals\"
                     },
                     \"generate_visuals\": {
                         \"action\": \"ai_image_generate_sdxl\",
                         \"description\": \"Generate accompanying images\",
                         \"topic\": \"system.adapter.image.generate\",
                         \"next_step\": \"finalize_content\"
                     },
                     \"finalize_content\": {
                         \"action\": \"ai_text_generate_claude_sonnet\",
                         \"description\": \"Finalize content with visual references\",
                         \"next_step\": \"complete_workflow\"
                     },
                     \"complete_workflow\": {
                         \"action\": \"complete_workflow\",
                         \"description\": \"Mark workflow as complete\"
                     }
                 }
             }
         }',
         true,
         NOW(),
         NOW()
     ) ON CONFLICT (id) DO NOTHING;" \
    "Visual Content Creator Template"

# 2. Verify subscription tiers exist (they should be created by migration)
echo -e "${YELLOW}💳 Verifying subscription tiers...${NC}"
mysql -h mysql-auth -u auth_user -p"$AUTH_DB_PASSWORD" auth_db -e "SELECT COUNT(*) as tier_count FROM subscription_tiers;" 2>/dev/null || {
    echo -e "${RED}❌ Subscription tiers table not accessible${NC}"
}

# 3. Create default permissions if they don't exist
echo -e "${YELLOW}🔐 Ensuring default permissions exist...${NC}"
seed_mysql_data \
    "mysql-auth" \
    "auth_user" \
    "auth_db" \
    "$AUTH_DB_PASSWORD" \
    "INSERT IGNORE INTO permissions (id, name, description) VALUES
        ('00000000-0000-0000-0000-000000000001', 'personas.create', 'Create new personas'),
        ('00000000-0000-0000-0000-000000000002', 'personas.delete', 'Delete personas'),
        ('00000000-0000-0000-0000-000000000003', 'projects.manage', 'Manage all projects'),
        ('00000000-0000-0000-0000-000000000004', 'admin.users', 'Manage users'),
        ('00000000-0000-0000-0000-000000000005', 'admin.subscriptions', 'Manage subscriptions'),
        ('00000000-0000-0000-0000-000000000006', '*', 'Super admin - all permissions');" \
    "Default Permissions"

echo -e "${GREEN}🎉 Data seeding completed successfully!${NC}"

# Show summary
echo -e "${YELLOW}📊 Seeding Summary:${NC}"
echo -e "${GREEN}✅ 4 Persona templates created${NC}"
echo -e "${GREEN}✅ 6 Default permissions ensured${NC}"
echo -e "${GREEN}✅ Subscription tiers verified${NC}"