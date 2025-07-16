# terraform/environments/production/uk/uk001/010-infrastructure/terraform.tfvars

instance_cluster_name           = "uk001-prod-cluster"
instance_rackspace_region       = "uk-lon-1"
instance_ondemand_node_flavor = "mh.vs1.medium-lon"
instance_spot_node_flavor     = "gp.vs1.large-lon" #mh.vs1.medium-lon has much more memory gp. is general
instance_slack_webhook_url    = "https://hooks.slack.com/services/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"
instance_ondemand_node_count = 0
instance_spot_min_nodes = 3
instance_spot_max_nodes = 6

# general purpose
# gp.vs1.medium-lon
# gp.vs1.large-lon
# gp.vs1.xlarge-lon
# gp.vs1.2xlarge-lon

# memory
# mh.vs1.medium-lon
# mh.vs1.large-lon
# mh.vs1.xlarge-lon
# mh.vs1.2xlarge-lon

# gp.bm2.medium-lon
# gp.bm2.large-lon

# ch.vs1.medium-lon
# ch.vs1.large-lon
# ch.vs1.xlarge-lon
# ch.vs1.2xlarge-lon





