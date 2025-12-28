package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// OfflineActivationFile represents the structure of an offline activation file
type OfflineActivationFile struct {
	LicenseKey      string `json:"licenseKey"`
	MachineID       string `json:"machineId"`
	ActivationToken string `json:"activationToken"`
	IssuedAt        string `json:"issuedAt"`
	ExpiresAt       string `json:"expiresAt,omitempty"`
	Signature       string `json:"signature"`
}

func main() {
	licenseKey := flag.String("license", "CKSWT-ABCDE-12345-FGHIJ-67890", "License key")
	machineID := flag.String("machine", "", "Machine ID (required)")
	output := flag.String("output", "cks-weight-room-activation.json", "Output file path")
	flag.Parse()

	if *machineID == "" {
		fmt.Println("Error: Machine ID is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run tools/generate-activation-file.go -machine ABCD-1234-EFGH-5678")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Generate mock activation file
	now := time.Now()
	activationFile := OfflineActivationFile{
		LicenseKey:      *licenseKey,
		MachineID:       *machineID,
		ActivationToken: fmt.Sprintf("OFFLINE-TOKEN-%s-%d", *machineID, now.Unix()),
		IssuedAt:        now.Format("2006-01-02 15:04:05"),
		Signature:       "MOCK-SIGNATURE-" + fmt.Sprintf("%d", now.Unix()),
	}

	// Marshal to JSON with pretty printing
	jsonData, err := json.MarshalIndent(activationFile, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile(*output, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Activation file generated successfully!\n")
	fmt.Printf("  File: %s\n", *output)
	fmt.Printf("  License Key: %s\n", *licenseKey)
	fmt.Printf("  Machine ID: %s\n", *machineID)
	fmt.Printf("\nYou can now upload this file in the offline activation screen.\n")
}
