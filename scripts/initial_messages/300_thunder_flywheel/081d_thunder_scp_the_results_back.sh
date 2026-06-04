# (confirm the exact secret name + ip if the connection fails:)
#   SELECT instance_ip, ssh_user, ssh_key_secret_name FROM thunder_instances WHERE id='fabfd7fa-ac84-4476-86f3-f7ac57862214';

# 1) pull the box's SSH private key from its k8s secret
kubectl -n ai-persona-system get secret thunder-ssh-fabfd7fa-ac84-4476-86f3-f7ac57862214 \
  -o jsonpath='{.data.private_key}' | base64 -d > /tmp/iter0_key
chmod 600 /tmp/iter0_key

# 2) copy the adapter dir off the box
scp -P 30340 -i /tmp/iter0_key \
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  -r ubuntu@216.81.200.234:/workspace/adapter_out \
  ./iter0_adapter_out


  -- check first
  ssh -p 30340 -i /tmp/iter0_key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    ubuntu@216.81.200.234 \
    'grep -E "RUN_SH_FULL_OK|RUN_SH_DONE" /workspace/train.log; ls -la /workspace/adapter_out 2>/dev/null'
