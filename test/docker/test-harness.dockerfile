# test/docker/test-harness.dockerfile
# Go-focused test harness for Kubernetes environments

# Stage 1: Go builder with test tools
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    g++ \
    musl-dev \
    bash \
    curl

WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install Go testing tools
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/onsi/gomega/...@latest && \
    go install gotest.tools/gotestsum@latest && \
    go install github.com/vektra/mockery/v2@latest && \
    go install github.com/golang/mock/mockgen@latest && \
    go install github.com/rakyll/gotest@latest && \
    go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install github.com/matm/gocov-html/cmd/gocov-html@latest && \
    go install github.com/jstemmer/go-junit-report/v2@latest

# Copy source code
COPY . .

# Build test binaries
RUN go test -c -o /test-runner ./test/cmd/runner

# Stage 2: Kafka tools
FROM confluentinc/cp-kafka:7.5.0 AS kafka-tools

# Stage 3: Final test harness
FROM golang:1.21-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    git \
    make \
    postgresql-client \
    mysql-client \
    redis \
    curl \
    wget \
    jq \
    vim \
    tmux \
    htop \
    ca-certificates \
    netcat-openbsd

# Install k6 for load testing
RUN wget -q https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz && \
    tar -xzf k6-v0.47.0-linux-amd64.tar.gz && \
    mv k6-v0.47.0-linux-amd64/k6 /usr/local/bin/ && \
    rm -rf k6-v0.47.0-linux-amd64*

# Install kafkacat/kcat
RUN wget -q https://github.com/edenhill/kcat/releases/download/1.7.1/kcat-1.7.1-linux-amd64 -O /usr/local/bin/kcat && \
    chmod +x /usr/local/bin/kcat && \
    ln -s /usr/local/bin/kcat /usr/local/bin/kafkacat

# Copy Kafka tools
COPY --from=kafka-tools /usr/bin/kafka-* /usr/local/bin/
COPY --from=kafka-tools /usr/share/java/kafka /usr/share/java/kafka

# Copy Go test tools from builder
COPY --from=builder /go/bin/* /usr/local/bin/
COPY --from=builder /test-runner /usr/local/bin/test-runner

# Create directories
RUN mkdir -p /workspace /results /fixtures /scripts /tools

# Copy test files
COPY test/scripts /scripts/
COPY test/fixtures /fixtures/
COPY test/tools /tools/
COPY test /workspace/test/

# Create main test harness script
RUN cat > /usr/local/bin/test-harness << 'EOF'
#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}=== Agent Chassis Test Harness ===${NC}"
echo -e "${BLUE}Running in Kubernetes${NC}"
echo "Namespace: ${NAMESPACE:-default}"
echo "Test Suite: ${TEST_SUITE:-all}"
echo ""

# Function to wait for service
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local timeout=${4:-30}

    echo -e "${YELLOW}Waiting for ${service} at ${host}:${port}...${NC}"
    for i in $(seq 1 $timeout); do
        if nc -z "$host" "$port" 2>/dev/null; then
            echo -e "${GREEN}✓ ${service} is ready${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
    done
    echo ""
    echo -e "${RED}✗ ${service} timeout${NC}"
    return 1
}

# Function to check Kubernetes services
check_k8s_services() {
    echo -e "${YELLOW}Checking Kubernetes services...${NC}"

    # Check PostgreSQL
    if wait_for_service "${DB_HOST}" "${DB_PORT:-5432}" "PostgreSQL" 10; then
        echo -e "${GREEN}✓ Database connection verified${NC}"
    else
        echo -e "${RED}✗ Database not accessible${NC}"
    fi

    # Check Kafka
    if wait_for_service "${KAFKA_BROKERS%%:*}" "${KAFKA_BROKERS##*:}" "Kafka" 10; then
        echo -e "${GREEN}✓ Kafka connection verified${NC}"
    else
        echo -e "${RED}✗ Kafka not accessible${NC}"
    fi
}

# Function to setup test database
setup_test_db() {
    echo -e "${YELLOW}Setting up test database...${NC}"

    if [ -f "/workspace/test/migrations/001_test_schema.sql" ]; then
        PGPASSWORD="${DB_PASSWORD}" psql \
            -h "${DB_HOST}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER}" \
            -d "${DB_NAME}" \
            -f /workspace/test/migrations/001_test_schema.sql
    fi

    if [ -f "/workspace/test/migrations/002_test_data.sql" ]; then
        PGPASSWORD="${DB_PASSWORD}" psql \
            -h "${DB_HOST}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER}" \
            -d "${DB_NAME}" \
            -f /workspace/test/migrations/002_test_data.sql
    fi

    echo -e "${GREEN}✓ Test database ready${NC}"
}

# Function to run test suite
run_test_suite() {
    local suite=$1
    echo -e "${YELLOW}Running ${suite} tests...${NC}"

    cd /workspace

    case $suite in
        unit)
            gotestsum --format=testname \
                --junitfile=/results/unit-tests.xml \
                -- -v -short -coverprofile=/results/coverage-unit.out \
                ./test/unit/...
            ;;
        integration)
            check_k8s_services
            setup_test_db

            gotestsum --format=testname \
                --junitfile=/results/integration-tests.xml \
                -- -v -coverprofile=/results/coverage-integration.out \
                ./test/integration/...
            ;;
        e2e)
            check_k8s_services
            setup_test_db

            # Run E2E setup script if exists
            if [ -f "/scripts/setup_e2e.sh" ]; then
                /scripts/setup_e2e.sh
            fi

            gotestsum --format=testname \
                --junitfile=/results/e2e-tests.xml \
                -- -v -timeout=30m -coverprofile=/results/coverage-e2e.out \
                ./test/e2e/...

            # Cleanup
            if [ -f "/scripts/cleanup_e2e.sh" ]; then
                /scripts/cleanup_e2e.sh
            fi
            ;;
        performance)
            check_k8s_services

            # Run Go benchmarks
            go test -bench=. -benchmem -benchtime=10s \
                -cpu=1,2,4 \
                ./test/performance/... | tee /results/bench_results.txt

            # Generate benchmark report
            if [ -f "/results/bench_results.txt" ]; then
                echo -e "\n${GREEN}=== Benchmark Summary ===${NC}"
                grep -E "^Benchmark" /results/bench_results.txt | column -t
            fi

            # Run k6 load tests if available
            if [ -f "/scripts/k6/load_test.js" ]; then
                echo -e "\n${YELLOW}Running k6 load tests...${NC}"
                k6 run --out json=/results/k6_results.json /scripts/k6/load_test.js
            fi
            ;;
        specific)
            # Run specific test
            if [ -n "$TEST_PATTERN" ]; then
                gotestsum --format=testname \
                    -- -v -run "$TEST_PATTERN" ./test/...
            else
                echo -e "${RED}No TEST_PATTERN specified${NC}"
                exit 1
            fi
            ;;
        all)
            run_test_suite unit
            run_test_suite integration
            run_test_suite e2e

            # Merge coverage reports
            echo -e "\n${YELLOW}Merging coverage reports...${NC}"
            gocov convert /results/coverage-unit.out > /results/coverage-unit.json
            gocov convert /results/coverage-integration.out > /results/coverage-integration.json
            gocov convert /results/coverage-e2e.out > /results/coverage-e2e.json

            # Generate HTML report
            gocov-html < /results/coverage-unit.json > /results/coverage.html
            ;;
        *)
            echo -e "${RED}Unknown test suite: ${suite}${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}✓ ${suite} tests completed${NC}"
}

# Main execution
case ${1:-test} in
    test)
        # Run specified test suite
        run_test_suite "${TEST_SUITE:-all}"
        ;;
    shell)
        # Interactive shell for debugging
        echo -e "${GREEN}Entering test harness shell...${NC}"
        echo -e "${YELLOW}Available commands:${NC}"
        echo "  - gotestsum: Enhanced test runner"
        echo "  - gocov: Coverage analysis"
        echo "  - k6: Load testing"
        echo "  - kcat: Kafka testing"
        echo "  - psql: PostgreSQL client"
        echo ""
        exec /bin/bash
        ;;
    validate)
        # Validate test environment
        echo -e "${YELLOW}Validating test environment...${NC}"

        # Check tools
        for tool in go gotestsum gocov k6 kcat psql; do
            if command -v $tool &> /dev/null; then
                echo -e "${GREEN}✓ ${tool} is available${NC}"
            else
                echo -e "${RED}✗ ${tool} is missing${NC}"
            fi
        done

        # Check Kubernetes services
        check_k8s_services
        ;;
    kafka-test)
        # Test Kafka connectivity
        /usr/local/bin/run-kafka-test "${2:-test.topic}"
        ;;
    db-test)
        # Test database connectivity
        /usr/local/bin/run-db-test
        ;;
    coverage)
        # Generate coverage report
        echo -e "${YELLOW}Generating coverage report...${NC}"

        if [ -f "/results/coverage-unit.out" ]; then
            go tool cover -html=/results/coverage-unit.out -o /results/coverage.html
            echo -e "${GREEN}✓ Coverage report generated: /results/coverage.html${NC}"
        else
            echo -e "${RED}No coverage data found. Run tests first.${NC}"
        fi
        ;;
    clean)
        # Clean test results
        echo -e "${YELLOW}Cleaning test results...${NC}"
        rm -rf /results/*
        echo -e "${GREEN}✓ Test results cleaned${NC}"
        ;;
    *)
        # Pass through to command
        exec "$@"
        ;;
esac

# Generate test summary if results exist
if compgen -G "/results/*.xml" > /dev/null; then
    echo -e "\n${GREEN}=== Test Summary ===${NC}"

    # Parse JUnit XML results
    for xml in /results/*.xml; do
        if [ -f "$xml" ]; then
            suite=$(basename "$xml" .xml)
            tests=$(grep -c '<testcase' "$xml" || echo "0")
            failures=$(grep -c '<failure' "$xml" || echo "0")
            errors=$(grep -c '<error' "$xml" || echo "0")

            if [ "$failures" -eq 0 ] && [ "$errors" -eq 0 ]; then
                echo -e "${GREEN}✓ ${suite}: ${tests} tests passed${NC}"
            else
                echo -e "${RED}✗ ${suite}: ${failures} failures, ${errors} errors out of ${tests} tests${NC}"
            fi
        fi
    done
fi

EOF
RUN chmod +x /usr/local/bin/test-harness

# Create Kafka test helper
COPY --from=builder /usr/local/bin/run-kafka-test /usr/local/bin/run-kafka-test

# Create database test helper
COPY --from=builder /usr/local/bin/run-db-test /usr/local/bin/run-db-test

# Set up workspace
WORKDIR /workspace

# Environment variables
ENV GO_ENV=test \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    TEST_RESULTS_DIR=/results \
    TEST_FIXTURES_DIR=/fixtures \
    TEST_TIMEOUT=30m \
    GOTESTSUM_FORMAT=testname

# Default entrypoint
ENTRYPOINT ["test-harness"]
CMD ["test"]