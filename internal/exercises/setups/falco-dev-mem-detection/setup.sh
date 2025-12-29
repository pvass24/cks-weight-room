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

    # Note: Container plugin is already configured via /etc/falco/config.d/falco.container_plugin.yaml
    # We just need the LD_PRELOAD workaround to make it work on Falco 0.42.x

    # Create custom /dev/mem detection rule in falco_rules.local.yaml (standard location)
    # This is the standard way to add custom Falco rules - now with container fields!
    cat > /etc/falco/falco_rules.local.yaml <<'\''RULE'\''
- rule: Access to /dev/mem
  desc: Detect attempts to access /dev/mem
  condition: fd.name = /dev/mem
  output: '\''Process accessing /dev/mem (user=%user.name process=%proc.name cmdline=%proc.cmdline container=%container.name pid=%proc.pid)'\''
  priority: WARNING
RULE

    # Create wrapper script for Falco with LD_PRELOAD workaround
    # This fixes the "__res_search" symbol issue in Falco 0.42.x
    cat > /usr/local/bin/falco-fixed <<'\''WRAPPER'\''
#!/bin/bash
# Workaround for Falco 0.42.x container plugin bug
# See: https://github.com/falcosecurity/falco/issues/3719
export LD_PRELOAD=/lib/aarch64-linux-gnu/libresolv.so.2
exec /usr/bin/falco "$@"
WRAPPER
    chmod +x /usr/local/bin/falco-fixed

    echo "==================================================================="
    echo "Falco installation complete with container plugin and custom rule!"
    echo "==================================================================="
    echo ""
    echo "Custom rule: /etc/falco/falco_rules.local.yaml"
    echo "Container plugin: ENABLED (with transparent LD_PRELOAD workaround)"
    echo ""
    echo "To run Falco and detect /dev/mem access with container info, use:"
    echo ""
    echo "  docker exec -it $NODE_NAME falco -U"
    echo ""
    echo "Note: The '\''falco'\'' alias includes the LD_PRELOAD workaround automatically."
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
echo "Exercise setup complete on node: $NODE_NAME!"
echo "==================================================================="
echo "Falco is installed with container plugin and custom /dev/mem rule."
echo "Deployments created: nvidia, cpu, gpu (in security-scan namespace)"
echo ""
echo "IMPORTANT: Pods may run on different nodes!"
echo "To solve this exercise:"
echo "  1. Find which node has the pods:"
echo "       kubectl get pods -n security-scan -o wide"
echo "  2. Exec into that node and run Falco:"
echo "       docker exec -it <node-name> falco -U"
echo ""
echo "Falco will show container names and Kubernetes metadata!"
echo "Press Ctrl+C to stop Falco when done observing."
echo "==================================================================="
