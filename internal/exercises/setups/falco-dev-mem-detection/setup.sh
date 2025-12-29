#!/bin/bash
set -e

NODE_NAME=$1
if [ -z "$NODE_NAME" ]; then
    echo "Usage: $0 <node-name>"
    exit 1
fi

echo "Setting up Falco exercise in node: $NODE_NAME"

# Install Falco in the KIND node
echo "Installing Falco..."
docker exec "$NODE_NAME" bash -c '
    # Install dependencies
    apt-get update -qq
    apt-get install -y -qq curl gnupg2 lsb-release

    # Add Falco repository
    curl -fsSL https://falco.org/repo/falcosecurity-packages.asc | gpg --dearmor -o /usr/share/keyrings/falco-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/falco-archive-keyring.gpg] https://download.falco.org/packages/deb stable main" | tee /etc/apt/sources.list.d/falcosecurity.list

    # Install Falco
    apt-get update -qq
    apt-get install -y -qq falco

    # Disable container plugin (ARM64 compatibility issue)
    rm -f /etc/falco/config.d/falco.container_plugin.yaml || true

    # Create simple falco config for /dev/mem detection
    cat > /etc/falco/rules.d/dev_mem.yaml <<EOF
- rule: Access to /dev/mem
  desc: Detect attempts to access /dev/mem
  condition: fd.name = /dev/mem
  output: "Process accessing /dev/mem (user=%user.name command=%proc.cmdline container=%container.name)"
  priority: WARNING
EOF

    # Enable and start Falco service
    systemctl enable falco || true
    systemctl start falco || true

    echo "Falco installation complete with custom /dev/mem rule"
'

# Deployments are already applied by SetupExercise function
echo "Deployments should already be created..."

echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=nvidia -n security-scan --timeout=60s || true
kubectl wait --for=condition=ready pod -l app=cpu -n security-scan --timeout=60s || true
kubectl wait --for=condition=ready pod -l app=gpu -n security-scan --timeout=60s || true

echo "Setup complete! Falco is running and deployments are created."
echo "Test with: docker exec $NODE_NAME falco -U"
