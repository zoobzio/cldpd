# NOTE: For production use, pin node:22-slim to a specific digest
# and specify a version for @anthropic-ai/claude-code.
FROM node:22-slim

# Install git and SSH client
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# Add GitHub to known hosts so SSH operations don't prompt for host verification
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh \
    && ssh-keyscan github.com >> /root/.ssh/known_hosts \
    && chmod 600 /root/.ssh/known_hosts

# Install Claude Code
RUN npm install -g @anthropic-ai/claude-code

WORKDIR /workspace
