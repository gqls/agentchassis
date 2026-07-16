FROM ubuntu:24.04

ARG RUNNER_VERSION=2.333.1
ARG RUNNER_ARCH=x64

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    git \
    jq \
    openssh-client \
    python3 \
    python3-pip \
    rsync \
    && rm -rf /var/lib/apt/lists/*

# Install B2 CLI
RUN pip3 install --break-system-packages b2

# Create runner user (GitHub runner refuses to run as root)
RUN useradd -m -d /home/runner runner

# Install GitHub Actions runner
WORKDIR /home/runner

RUN curl -fsSL -o runner.tar.gz \
    "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz" \
    && tar xzf runner.tar.gz \
    && rm runner.tar.gz \
    && ./bin/installdependencies.sh

COPY build/docker/backend/github-actions-runner-entrypoint.sh /home/runner/entrypoint.sh
RUN chmod +x /home/runner/entrypoint.sh

RUN chown -R runner:runner /home/runner

USER runner

ENTRYPOINT ["/home/runner/entrypoint.sh"]
