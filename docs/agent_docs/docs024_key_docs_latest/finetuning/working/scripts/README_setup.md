see RUNBOOK_phase_b_c_d_deploy.md latest version

path is: docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/00_vm_setup.sh
full path is: /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/

mkdir -p /tmp/bundle && cd /tmp/bundle
# copy in the THREE on-VM scripts (00_vm_setup.sh unchanged from the last bundle):
cp /path/to/00_vm_setup.sh .
cp /path/to/02_train_llama_3_3_70b.py .    # Phase A edited copy
cp /path/to/run.sh .                       # Phase C edited copy
chmod +x run.sh 00_vm_setup.sh
tar -czf bundle.tar.gz 00_vm_setup.sh 02_train_llama_3_3_70b.py run.sh
tar -tzf bundle.tar.gz    # MUST list the three files at the root, no leading dir

cp /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/00_vm_setup.sh .
cp /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/02_train_llama_3_3_70b.py .    # Phase A edited copy
cp /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/run.sh .                       # Phase C edited copy
chmod +x run.sh 00_vm_setup.sh
tar -czf bundle.tar.gz 00_vm_setup.sh 02_train_llama_3_3_70b.py run.sh
tar -tzf bundle.tar.gz    # MUST list the three files at the root, no leading dir

--

# b2 v4:
b2 file upload personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz
b2 ls --long "b2://personae-model-training/finetuning/scripts/"
# b2 v3 equivalents:
#   b2 upload-file personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz
#   b2 ls --long personae-model-training finetuning/scripts/