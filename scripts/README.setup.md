How It All Works Together

The scripts are designed to work in a specific, logical sequence that ensures a smooth deployment:

    Local Setup: A developer runs setup.sh once to prepare the cluster by creating the namespace and all necessary secrets.

    Deployment Orchestration: The developer runs make deploy (or make quickstart), which executes deploy-system.sh.

    Infrastructure Creation: deploy-system.sh applies the Kubernetes YAML files, creating the database and Kafka StatefulSets, services, and jobs.

    Container-Level Waiting: As the database-init job pod starts, its internal wait-for-services.sh script runs, pausing execution until the PostgreSQL and MySQL pods are fully ready to accept connections.

    Database Migration: Once the databases are ready, docker-run-migrations.sh runs within the job, applying all the table schemas.

    Data Seeding: The data-seeder job follows the same pattern, waiting for services before seed-data.sh inserts the initial persona templates.

    Service Startup: Finally, the main microservices (like auth-service and core-manager) start. Their initContainers also run wait scripts to ensure they don't start before their own dependencies (like databases or specific Kafka topics) are available.