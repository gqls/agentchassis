#!/bin/bash
# Script to update agent definitions with current image information
# This can be run as part of deployment or as a Kubernetes Job

# Get the current image from the agent-chassis deployment
CURRENT_IMAGE=$(kubectl -n ai-persona-system get deployment agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}')

# Parse the image components
if [[ $CURRENT_IMAGE =~ ^(.+):(.+)$ ]]; then
    IMAGE_REPO="${BASH_REMATCH[1]}"
    IMAGE_TAG="${BASH_REMATCH[2]}"
else
    echo "Error: Could not parse image: $CURRENT_IMAGE"
    exit 1
fi

echo "Updating agent definitions with:"
echo "  Repository: $IMAGE_REPO"
echo "  Tag: $IMAGE_TAG"

# Update the database
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db << EOF
-- Update all agent definitions with the current image
UPDATE agent_definitions 
SET 
    image_repository = '$IMAGE_REPO',
    image_tag = '$IMAGE_TAG',
    updated_at = NOW()
WHERE 1=1;

-- Show the results
SELECT type, image_repository, image_tag 
FROM agent_definitions 
ORDER BY type
LIMIT 5;
EOF

echo "Agent definitions updated successfully"