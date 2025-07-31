
# Apply and test
cd ~/projects/agent-chassis/tests/agents/mysql/
kubectl apply -f mysql-test-pod.yaml
kubectl exec -it mysql-connection-test -n ai-persona-system -- mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASSWORD -e "SELECT 1"