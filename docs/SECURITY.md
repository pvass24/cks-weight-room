# Security Features

CKS Weight Room implements multiple layers of security to protect users' systems from potentially harmful commands executed during practice sessions.

## Overview

Practice terminal sessions can be run in two modes:

1. **Standard Mode** - Terminal runs directly on host (original behavior)
2. **Secure Mode** - Terminal runs in isolated Docker container with command filtering (recommended)

## Secure Mode Features

### 1. Containerized Terminals

Each terminal session runs in an isolated Docker container with:

- **Non-root user** - Sessions run as `cksuser` (UID 1000), not root
- **Read-only root filesystem** - Cannot modify system files
- **Limited tmpfs mounts** - Only `/tmp` and `/home/cksuser` are writable
- **Resource limits**:
  - Maximum memory: 512MB
  - Maximum CPU: 1 core
- **Dropped capabilities** - All Linux capabilities dropped except `NET_RAW` (for ping)
- **No privilege escalation** - `no-new-privileges` flag set
- **Auto-cleanup** - Containers removed after session ends

### 2. Command Filtering

All commands are validated before execution. The system blocks:

#### File System Destruction
- `rm -rf /` and variants
- `dd` commands (direct disk access)
- `mkfs`, `fdisk`, `parted` (filesystem/partition manipulation)

#### System Manipulation
- `reboot`, `shutdown`, `halt`, `poweroff`
- `init 0`, `init 6`

#### Privilege Escalation
- `sudo` commands
- `su` (switch user)

#### Network Attacks
- `nmap`, `masscan` (port scanning)
- `metasploit` framework

#### Process Manipulation
- `kill -9 1` (killing init)
- `killall -9` (force kill all)

#### Resource Exhaustion
- Fork bombs (`:(){:|:&};:`)
- Infinite loops

#### Escape Attempts
- `chroot` operations
- `docker run/exec` (container escape)
- `nsenter` (namespace escape)

#### Suspicious Patterns
- Base64 decode and execute
- Excessive command chaining (>5 pipes or semicolons)
- Commands over 1000 characters
- Multiple directory traversal attempts

### 3. Allowed Commands

The filter allows all commands needed for CKS practice:

- **Kubernetes**: `kubectl`, all kubectl subcommands
- **File operations**: `ls`, `cat`, `less`, `grep`, `find`, etc.
- **Editors**: `vim`, `nano`
- **Network tools**: `curl`, `wget`, `ping`, `netstat`
- **System info**: `ps`, `top`, `free`, `df`
- **CKS tools**: `falco`, `trivy`, `kube-bench`, `kubesec`, `opa`
- **Container tools**: `crictl`, `ctr`, `nerdctl`
- **Security tools**: `openssl`, `apparmor_parser`, `seccomp`

## Enabling Secure Mode

### Option 1: Environment Variable (Recommended)

Set the `SECURE_TERMINAL` environment variable:

```bash
export SECURE_TERMINAL=true
./cks-weight-room
```

### Option 2: Build and Use Secure Image

1. Build the terminal image:
```bash
./scripts/build-terminal-image.sh
```

2. The secure terminal will be used automatically when the image exists and `SECURE_TERMINAL=true`

## Security Trade-offs

### Secure Mode Advantages
✅ Complete isolation from host system
✅ Command filtering prevents dangerous operations
✅ Resource limits prevent DoS attacks
✅ Non-root execution
✅ Read-only filesystem
✅ Automatic cleanup

### Secure Mode Limitations
⚠️ Requires Docker
⚠️ Slightly slower startup (container creation)
⚠️ Some legitimate commands may be blocked (can be added to allowlist)
⚠️ Cannot access host filesystem (by design)

### Standard Mode
⚠️ Direct host access - use only on trusted systems
⚠️ No command filtering
⚠️ No resource limits
⚠️ User has full bash access

## Monitoring and Logging

All blocked commands are logged with:
- Exercise slug
- Command that was blocked
- Reason for blocking
- Timestamp

Example log entry:
```
Blocked command for disable-anonymous-access: rm -rf / (reason: Blocked: Recursive delete from root)
```

## Development

### Adding Allowed Commands

Edit `internal/security/commandfilter.go` and add to the `allowedCommands` slice:

```go
allowedCommands: []string{
    // ... existing commands ...
    "yournewcommand",
}
```

### Adding Blocked Patterns

Add to the `blockedCommands` slice:

```go
{regexp.MustCompile(`\byourpattern\b`), "Description of why it's blocked"},
```

### Modifying Container Settings

Edit `internal/api/terminal_secure.go` to adjust:
- Resource limits (`maxTerminalMemory`, `maxTerminalCPU`)
- Capabilities (`CapDrop`, `CapAdd`)
- Mount points
- Security options

## Best Practices

1. **Always use Secure Mode in production**
2. **Review logs regularly** for blocked commands
3. **Update allowed commands** as needed for new exercises
4. **Keep terminal image updated** with latest security patches
5. **Monitor resource usage** to adjust limits if needed

## Testing Security

To test the security features:

1. Try blocked commands:
```bash
# Should be blocked
rm -rf /
sudo su
dd if=/dev/zero of=/dev/sda
```

2. Try allowed commands:
```bash
# Should work
kubectl get pods
curl https://kubernetes.io
ping 8.8.8.8
```

3. Check logs for blocked attempts
4. Verify container isolation:
```bash
# From inside container - should fail
ls /Users  # Host filesystem not accessible
```

## FAQ

**Q: Can users bypass the command filter?**
A: The filter is not perfect, but combined with container isolation, the risk is minimal. The container itself provides the primary security boundary.

**Q: What if a legitimate command is blocked?**
A: Add it to the allowlist in `commandfilter.go` and rebuild. Consider if the block reason is valid first.

**Q: Does Secure Mode work offline?**
A: Yes, once the terminal image is built, it works offline. Only the initial build requires internet.

**Q: What about performance?**
A: Container startup adds ~1-2 seconds. Once running, performance is near-native.

**Q: Can users see other users' sessions?**
A: No, each session gets its own isolated container with unique namespace.
