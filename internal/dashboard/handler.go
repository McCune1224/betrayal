package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditResponse represents a single audit record for API responses
type AuditResponse struct {
	ID               int64                  `json:"id"`
	Timestamp        time.Time              `json:"timestamp"`
	CommandName      string                 `json:"command_name"`
	UserID           string                 `json:"user_id"`
	Username         string                 `json:"username"`
	UserRoles        []string               `json:"user_roles"`
	IsAdmin          bool                   `json:"is_admin"`
	CommandArguments map[string]interface{} `json:"command_arguments"`
	Status           string                 `json:"status"`
	ErrorMessage     *string                `json:"error_message"`
	ExecutionTimeMs  int32                  `json:"execution_time_ms"`
}

// AuditListResponse is the response for listing audit records
type AuditListResponse struct {
	Total   int64           `json:"total"`
	Records []AuditResponse `json:"records"`
	Page    int32           `json:"page"`
	Limit   int32           `json:"limit"`
}

// StatsResponse shows audit trail statistics
type StatsResponse struct {
	TotalCommands      int64          `json:"total_commands"`
	CommandsToday      int64          `json:"commands_today"`
	FailedCommands     int64          `json:"failed_commands"`
	AverageExecutionMs float64        `json:"average_execution_ms"`
	TopCommands        []CommandStats `json:"top_commands"`
	TopUsers           []UserStats    `json:"top_users"`
}

// CommandStats shows command execution statistics
type CommandStats struct {
	CommandName string `json:"command_name"`
	Count       int64  `json:"count"`
	Failures    int64  `json:"failures"`
}

// UserStats shows user activity statistics
type UserStats struct {
	Username     string `json:"username"`
	UserID       string `json:"user_id"`
	CommandCount int64  `json:"command_count"`
}

// Handler holds dependencies for the dashboard API
type Handler struct {
	db *pgxpool.Pool
}

// NewHandler creates a new dashboard handler
func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers all dashboard routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/audit/list", h.listAudits)
	mux.HandleFunc("/api/audit/stats", h.getStats)
	mux.HandleFunc("/api/audit/command/", h.getCommandDetails)
	mux.HandleFunc("/api/audit/user/", h.getUserDetails)
	mux.HandleFunc("/health", h.healthCheck)
}

// listAudits returns a paginated list of audit records
func (h *Handler) listAudits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	commandStr := r.URL.Query().Get("command")

	page := int32(1)
	if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
		page = int32(p)
	}

	limit := int32(50)
	if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 && l <= 1000 {
		limit = int32(l)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Build query
	query := `SELECT id, timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms 
		FROM command_audit`

	args := []interface{}{}
	argCount := 1

	if commandStr != "" {
		query += fmt.Sprintf(" WHERE command_name = $%d", argCount)
		args = append(args, commandStr)
		argCount++
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM command_audit`
	if commandStr != "" {
		countQuery += fmt.Sprintf(" WHERE command_name = $%d", 1)
	}

	var total int64
	if err := h.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get count: %v", err), http.StatusInternalServerError)
		return
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	records := []AuditResponse{}
	for rows.Next() {
		var record AuditResponse
		var userRolesJSON []byte
		var argsJSON []byte

		if err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.CommandName,
			&record.UserID,
			&record.Username,
			&userRolesJSON,
			&record.IsAdmin,
			&argsJSON,
			&record.Status,
			&record.ErrorMessage,
			&record.ExecutionTimeMs,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse JSON fields
		_ = json.Unmarshal(userRolesJSON, &record.UserRoles)
		_ = json.Unmarshal(argsJSON, &record.CommandArguments)

		records = append(records, record)
	}

	response := AuditListResponse{
		Total:   total,
		Records: records,
		Page:    page,
		Limit:   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getStats returns audit trail statistics
func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats := StatsResponse{}

	// Total commands
	err := h.db.QueryRow(ctx, "SELECT COUNT(*) FROM command_audit").Scan(&stats.TotalCommands)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get total commands: %v", err), http.StatusInternalServerError)
		return
	}

	// Commands today
	err = h.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE DATE(timestamp) = CURRENT_DATE").Scan(&stats.CommandsToday)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get today's commands: %v", err), http.StatusInternalServerError)
		return
	}

	// Failed commands
	err = h.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE status = 'ERROR'").Scan(&stats.FailedCommands)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get failed commands: %v", err), http.StatusInternalServerError)
		return
	}

	// Average execution time
	err = h.db.QueryRow(ctx,
		"SELECT COALESCE(AVG(execution_time_ms), 0) FROM command_audit").Scan(&stats.AverageExecutionMs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get average execution time: %v", err), http.StatusInternalServerError)
		return
	}

	// Top 5 commands
	rows, err := h.db.Query(ctx,
		`SELECT command_name, COUNT(*), SUM(CASE WHEN status = 'ERROR' THEN 1 ELSE 0 END)
		 FROM command_audit
		 GROUP BY command_name
		 ORDER BY COUNT(*) DESC
		 LIMIT 5`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get top commands: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stats.TopCommands = []CommandStats{}
	for rows.Next() {
		var cmd CommandStats
		if err := rows.Scan(&cmd.CommandName, &cmd.Count, &cmd.Failures); err != nil {
			http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
			return
		}
		stats.TopCommands = append(stats.TopCommands, cmd)
	}

	// Top 5 users
	rows, err = h.db.Query(ctx,
		`SELECT username, user_id, COUNT(*)
		 FROM command_audit
		 GROUP BY username, user_id
		 ORDER BY COUNT(*) DESC
		 LIMIT 5`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get top users: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stats.TopUsers = []UserStats{}
	for rows.Next() {
		var user UserStats
		if err := rows.Scan(&user.Username, &user.UserID, &user.CommandCount); err != nil {
			http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
			return
		}
		stats.TopUsers = append(stats.TopUsers, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getCommandDetails returns audit history for a specific command
func (h *Handler) getCommandDetails(w http.ResponseWriter, r *http.Request) {
	commandName := r.URL.Path[len("/api/audit/command/"):]
	if commandName == "" {
		http.Error(w, "Command name required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `SELECT id, timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms
		FROM command_audit
		WHERE command_name = $1
		ORDER BY timestamp DESC
		LIMIT 100`

	rows, err := h.db.Query(ctx, query, commandName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	records := []AuditResponse{}
	for rows.Next() {
		var record AuditResponse
		var userRolesJSON []byte
		var argsJSON []byte

		if err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.CommandName,
			&record.UserID,
			&record.Username,
			&userRolesJSON,
			&record.IsAdmin,
			&argsJSON,
			&record.Status,
			&record.ErrorMessage,
			&record.ExecutionTimeMs,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
			return
		}

		_ = json.Unmarshal(userRolesJSON, &record.UserRoles)
		_ = json.Unmarshal(argsJSON, &record.CommandArguments)

		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// getUserDetails returns audit history for a specific user
func (h *Handler) getUserDetails(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/api/audit/user/"):]
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `SELECT id, timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms
		FROM command_audit
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT 100`

	rows, err := h.db.Query(ctx, query, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	records := []AuditResponse{}
	for rows.Next() {
		var record AuditResponse
		var userRolesJSON []byte
		var argsJSON []byte

		if err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.CommandName,
			&record.UserID,
			&record.Username,
			&userRolesJSON,
			&record.IsAdmin,
			&argsJSON,
			&record.Status,
			&record.ErrorMessage,
			&record.ExecutionTimeMs,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
			return
		}

		_ = json.Unmarshal(userRolesJSON, &record.UserRoles)
		_ = json.Unmarshal(argsJSON, &record.CommandArguments)

		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// healthCheck returns server health status
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
