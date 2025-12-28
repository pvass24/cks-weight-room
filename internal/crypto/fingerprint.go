package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

// GetMachineID generates a unique machine identifier based on hostname and MAC address
// Returns a formatted machine ID like: ABCD-1234-EFGH-5678
func GetMachineID() (string, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Get MAC address of first non-loopback interface
	macAddr := ""
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Use first valid hardware address
		if len(iface.HardwareAddr) > 0 {
			macAddr = iface.HardwareAddr.String()
			break
		}
	}

	if macAddr == "" {
		return "", fmt.Errorf("no valid network interface found")
	}

	// Combine hostname and MAC address
	combined := hostname + "|" + macAddr

	// Hash the combined string for consistency
	hash := sha256.Sum256([]byte(combined))
	hashHex := hex.EncodeToString(hash[:16]) // Use first 16 bytes (32 hex chars)

	// Format as XXXX-XXXX-XXXX-XXXX
	formatted := fmt.Sprintf("%s-%s-%s-%s",
		strings.ToUpper(hashHex[0:4]),
		strings.ToUpper(hashHex[4:8]),
		strings.ToUpper(hashHex[8:12]),
		strings.ToUpper(hashHex[12:16]))

	return formatted, nil
}

// GetMachineIDForEncryption returns the raw machine identifier for use as encryption key material
func GetMachineIDForEncryption() (string, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Get MAC address
	macAddr := ""
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		if len(iface.HardwareAddr) > 0 {
			macAddr = iface.HardwareAddr.String()
			break
		}
	}

	if macAddr == "" {
		return "", fmt.Errorf("no valid network interface found")
	}

	// Return combined raw identifier (used for key derivation)
	return hostname + "|" + macAddr, nil
}
