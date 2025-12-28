package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/cluster"
	cerrors "github.com/patrickvassell/cks-weight-room/internal/errors"
)

// ClusterResponse represents the API response for cluster operations
type ClusterResponse struct {
	Success         bool                       `json:"success"`
	Cluster         *cluster.Cluster           `json:"cluster,omitempty"`
	Message         string                     `json:"message,omitempty"`
	Error           string                     `json:"error,omitempty"`
	ActionableError *cerrors.ActionableError   `json:"actionableError,omitempty"`
}

// ProvisionRequest represents the request to provision a cluster
type ProvisionRequest struct {
	ExerciseSlug string `json:"exerciseSlug"`
}

// convertClusterError converts a cluster.ClusterError to an ActionableError
func convertClusterError(err error) *cerrors.ActionableError {
	var clusterErr *cluster.ClusterError
	if !errors.As(err, &clusterErr) {
		// Not a ClusterError, return generic error
		return cerrors.NewActionableError(
			cerrors.ErrOperationFailed,
			"Cluster operation failed",
			err.Error(),
			[]string{"Try the operation again", "Contact support if the issue persists"},
			true,
		).WithInternalError(err)
	}

	// Convert based on error code
	switch clusterErr.Code {
	case cluster.ErrCodeDockerNotRunning:
		return cerrors.NewDockerNotRunningError().WithInternalError(err)

	case cluster.ErrCodeProvisionFailed:
		return cerrors.NewClusterProvisionFailedError(clusterErr.Message).WithInternalError(err)

	default:
		return cerrors.NewActionableError(
			cerrors.ErrOperationFailed,
			"Cluster operation failed",
			clusterErr.Message,
			[]string{"Try the operation again", "Contact support if the issue persists"},
			true,
		).WithInternalError(err)
	}
}

// ProvisionCluster handles POST /api/cluster/provision
func ProvisionCluster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ClusterResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.ExerciseSlug == "" {
		response := ClusterResponse{
			Success: false,
			Error:   "exerciseSlug is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create context with timeout (2 minutes for cluster provisioning)
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	// Provision cluster (no progress channel for now - we'll add SSE later)
	clusterInfo, err := cluster.ProvisionCluster(ctx, req.ExerciseSlug, nil)

	if err != nil {
		actionableErr := convertClusterError(err)
		response := ClusterResponse{
			Success:         false,
			Cluster:         clusterInfo,
			Error:           err.Error(),
			ActionableError: actionableErr,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ClusterResponse{
		Success: true,
		Cluster: clusterInfo,
		Message: "Cluster provisioned successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetClusterStatus handles GET /api/cluster/status/{exerciseSlug}
func GetClusterStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exercise slug from path
	path := r.URL.Path
	slug := path[len("/api/cluster/status/"):]

	if slug == "" {
		response := ClusterResponse{
			Success: false,
			Error:   "exerciseSlug is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	clusterName := cluster.GetClusterName(slug)
	clusterInfo, err := cluster.GetClusterStatus(ctx, clusterName)

	if err != nil {
		response := ClusterResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ClusterResponse{
		Success: true,
		Cluster: clusterInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteCluster handles DELETE /api/cluster/{exerciseSlug}
func DeleteCluster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exercise slug from path
	path := r.URL.Path
	slug := path[len("/api/cluster/"):]

	if slug == "" {
		response := ClusterResponse{
			Success: false,
			Error:   "exerciseSlug is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	clusterName := cluster.GetClusterName(slug)
	err := cluster.DeleteCluster(ctx, clusterName)

	if err != nil {
		response := ClusterResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ClusterResponse{
		Success: true,
		Message: "Cluster deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
