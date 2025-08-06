# test/README.md
# Agent Chassis Test Suite

This directory contains all tests for the Agent Chassis system.

## Test Categories

### Unit Tests (`./unit/`)
Fast, isolated tests for individual components.
```bash
make test-unit

Integration Tests (./integration/)
Tests that verify component interactions with real dependencies.
make test-integration

End-to-End Tests (./e2e/)
Full system tests that verify complete workflows.
make test-e2e

Performance Tests (./performance/)
Benchmarks and load tests.
make test-performance

Quick Start
Run All Tests
make test-all

Run Specific Agent Test
make test-agent AGENT=content-creator

Run in Kubernetes
make test-k8s

Test Data Setup
Before running tests, ensure test data is set up:
make test-setup


This structure provides:

Clear organization by test type
Easy-to-run test commands
Reusable test utilities
Proper fixtures and test data
K8s integration for cluster testing
Performance testing capabilities

The key improvements:

Consolidated duplicate test scripts
Organized by test type (unit/integration/e2e)
Centralized test configuration
Reusable test helpers
Clear documentation
Easy-to-use Makefile targets