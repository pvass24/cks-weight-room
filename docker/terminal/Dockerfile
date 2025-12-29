# Secure Terminal Container for CKS Weight Room
# This container provides an isolated environment for terminal sessions
# with limited access to the host system

FROM ubuntu:22.04

# Install only necessary tools
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    git \
    vim \
    nano \
    less \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# Install additional security tools commonly used in CKS
RUN curl -sfL https://get.docker.com | sh - || true

# Create non-root user for security
RUN useradd -m -s /bin/bash -u 1000 cksuser

# Set up working directory
WORKDIR /home/cksuser

# Create kubectl alias that uses the provided context
# The KUBECTL_CONTEXT env var is passed by the backend
# Must do this BEFORE switching to cksuser so we can write to .bashrc
RUN echo 'if [ -n "$KUBECTL_CONTEXT" ]; then' >> /home/cksuser/.bashrc && \
    echo '  alias kubectl="kubectl --context=$KUBECTL_CONTEXT"' >> /home/cksuser/.bashrc && \
    echo '  alias k="kubectl --context=$KUBECTL_CONTEXT"' >> /home/cksuser/.bashrc && \
    echo 'else' >> /home/cksuser/.bashrc && \
    echo '  alias k=kubectl' >> /home/cksuser/.bashrc && \
    echo 'fi' >> /home/cksuser/.bashrc && \
    chown -R cksuser:cksuser /home/cksuser

# Switch to non-root user
USER cksuser

# Set environment
ENV TERM=xterm-256color
ENV HOME=/home/cksuser

# Default command
CMD ["/bin/bash"]
