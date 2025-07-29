
https://claude.ai/chat/878cdc53-1a28-4598-beb9-0e0926b63bb8

~/projects/agent-chassis/tests/agents$ kubectl apply -f content-creator-test.yaml 

# Copy the test file into the pod
kubectl cp tests/agents/test_content_creator.go ai-persona-system/content-creator-test:/test.go

# Exec into the pod
kubectl exec -it -n ai-persona-system content-creator-test -- /bin/sh

# Inside the pod
cd /
go mod init test
go get github.com/segmentio/kafka-go
go get github.com/google/uuid
go get go.uber.org/zap

# Run the test
go run test.go