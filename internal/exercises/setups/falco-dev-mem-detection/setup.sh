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

    # Create custom /dev/mem detection rule in falco_rules.local.yaml (standard location)
    # This is the standard way to add custom Falco rules
    # Note: Cannot use container.name - container plugin has ARM64 compatibility issue
    #       (undefined symbol: __res_search in libcontainer.so)
    cat > /etc/falco/falco_rules.local.yaml <<'\''RULE'\''
- rule: Access to /dev/mem
  desc: Detect attempts to access /dev/mem
  condition: fd.name = /dev/mem
  output: '\''Process accessing /dev/mem (user=%user.name process=%proc.name cmdline=%proc.cmdline pid=%proc.pid)'\''
  priority: WARNING
RULE

    echo "==================================================================="
    echo "Falco installation complete with custom /dev/mem detection rule"
    echo "==================================================================="
    echo ""
    echo "Custom rule created in /etc/falco/falco_rules.local.yaml"
    echo "This is the standard location for local rules and will be loaded automatically."
    echo ""
    echo "To run Falco and detect /dev/mem access, use:"
    echo ""
    echo "  docker exec -it $NODE_NAME falco -U"
    echo ""
    echo "The '\''cpu'\'' pod is configured to access /dev/mem every 10 seconds."
    echo "==================================================================="
'

# Deployments are already applied by SetupExercise function
echo "Deployments should already be created..."

echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=nvidia -n security-scan --timeout=60s || true
kubectl wait --for=condition=ready pod -l app=cpu -n security-scan --timeout=60s || true
kubectl wait --for=condition=ready pod -l app=gpu -n security-scan --timeout=60s || true

echo ""
echo "==================================================================="
echo "Exercise setup complete!"
echo "==================================================================="
echo "Falco is installed with custom /dev/mem detection rule."
echo "Deployments created: nvidia, cpu, gpu (in security-scan namespace)"
echo ""
echo "To detect which pod is accessing /dev/mem, run:"
echo "  docker exec -it $NODE_NAME falco -U"
echo ""
echo "Press Ctrl+C to stop Falco when done observing."
echo "==================================================================="
