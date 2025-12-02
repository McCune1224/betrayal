package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

// ANSI color codes for terminal output
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

// auditQuery represents different types of audit queries
type auditQuery struct {
	name        string
	description string
	run         func(pool *pgxpool.Pool, args []string) error
}

var availableQueries = map[string]auditQuery{
	"stats": {
		name:        "stats",
		description: "Show audit trail statistics and health overview",
		run:         queryStats,
	},
	"top-commands": {
		name:        "top-commands",
		description: "Show most frequently executed commands",
		run:         queryTopCommands,
	},
	"top-users": {
		name:        "top-users",
		description: "Show most active users by command count",
		run:         queryTopUsers,
	},
	"failures": {
		name:        "failures",
		description: "Show failed command executions with error details",
		run:         queryFailures,
	},
	"user-activity": {
		name:        "user-activity",
		description: "Show detailed activity timeline for a specific user",
		run:         queryUserActivity,
	},
	"command-history": {
		name:        "command-history",
		description: "Show execution history for a specific command",
		run:         queryCommandHistory,
	},
	"audit-by-date": {
		name:        "audit-by-date",
		description: "Show all audit records from a specific date",
		run:         queryAuditByDate,
	},
	"slow-commands": {
		name:        "slow-commands",
		description: "Show slowest commands by average execution time",
		run:         querySlowCommands,
	},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check for help flag
	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		printUsage()
		os.Exit(0)
	}

	// Create database pool
	dsn := os.Getenv("DATABASE_POOLER_URL")
	if dsn == "" {
		printError("DATABASE_POOLER_URL environment variable not set")
		fmt.Fprintf(os.Stderr, "\nSet it with: export DATABASE_POOLER_URL='postgresql://user:pass@host:port/db'\n")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		printError("Failed to create database connection pool")
		fmt.Fprintf(os.Stderr, "Details: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Check database connection
	if err := pool.Ping(context.Background()); err != nil {
		printError("Failed to connect to database")
		fmt.Fprintf(os.Stderr, "Details: %v\n", err)
		os.Exit(1)
	}

	queryType := os.Args[1]
	query, exists := availableQueries[queryType]
	if !exists {
		printError(fmt.Sprintf("Unknown query type: %s", queryType))
		fmt.Fprintf(os.Stderr, "\nRun '%s help' to see available commands\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf("%s▶ Executing audit query:%s %s\n", colorCyan, colorReset, colorBold+queryType+colorReset)
	fmt.Println(string(make([]byte, 60))) // visual separator

	if err := query.run(pool, os.Args[2:]); err != nil {
		printError(fmt.Sprintf("Query failed: %v", err))
		os.Exit(1)
	}

	fmt.Println(string(make([]byte, 60))) // visual separator
	fmt.Printf("%s✓ Query completed successfully%s\n", colorGreen, colorReset)
}

func printUsage() {
	fmt.Printf("%s╔════════════════════════════════════════════════════════╗%s\n", colorBlue, colorReset)
	fmt.Printf("%s║          Betrayal Audit Analysis Tool                 ║%s\n", colorBlue, colorReset)
	fmt.Printf("%s║        Query and analyze command audit trails         ║%s\n", colorBlue, colorReset)
	fmt.Printf("%s╚════════════════════════════════════════════════════════╝%s\n\n", colorBlue, colorReset)

	fmt.Printf("%sUsage:%s\n", colorBold, colorReset)
	fmt.Printf("  %s <query-type> [options]\n\n", os.Args[0])

	fmt.Printf("%sAvailable Queries:%s\n", colorBold, colorReset)
	for _, query := range availableQueries {
		fmt.Printf("  %s%-18s%s  %s\n", colorGreen, query.name, colorReset, query.description)
	}

	fmt.Printf("\n%sExamples:%s\n", colorBold, colorReset)
	fmt.Printf("  %s stats%s\n", colorCyan, colorReset)
	fmt.Printf("  %s top-commands --limit 15%s\n", colorCyan, colorReset)
	fmt.Printf("  %s failures --hours 24%s\n", colorCyan, colorReset)
	fmt.Printf("  %s user-activity --user-id 123456789%s\n", colorCyan, colorReset)
	fmt.Printf("  %s command-history --command /setup%s\n", colorCyan, colorReset)
	fmt.Printf("  %s slow-commands --limit 10%s\n\n", colorCyan, colorReset)

	fmt.Printf("%sFlags:%s\n", colorBold, colorReset)
	fmt.Printf("  --limit N        Number of results to show (default varies by query)\n")
	fmt.Printf("  --user-id ID     Discord user ID for user-activity query\n")
	fmt.Printf("  --command CMD    Command name for command-history query\n")
	fmt.Printf("  --date DATE      Date in YYYY-MM-DD format (default: today)\n")
	fmt.Printf("  --hours N        Look back N hours for failures query\n\n")

	fmt.Printf("%sEnvironment:%s\n", colorBold, colorReset)
	fmt.Printf("  DATABASE_POOLER_URL  PostgreSQL connection string (required)\n\n")
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s✗ Error:%s %s\n", colorRed, colorReset, msg)
}

func printSuccess(msg string) {
	fmt.Printf("%s✓ %s%s\n", colorGreen, msg, colorReset)
}

func printInfo(msg string) {
	fmt.Printf("%sℹ %s%s\n", colorCyan, msg, colorReset)
}

func printSection(title string) {
	fmt.Printf("\n%s%s%s\n", colorBold+colorMagenta, title, colorReset)
	fmt.Println(string(make([]byte, len(title))))
}

// queryStats shows overall audit statistics
func queryStats(pool *pgxpool.Pool, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type stats struct {
		Total    int64
		Today    int64
		Failures int64
		AvgTime  float64
		MaxTime  int32
	}

	var s stats

	printInfo("Querying database for audit statistics...")

	// Total commands
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM command_audit").Scan(&s.Total)
	if err != nil {
		return fmt.Errorf("failed to get total commands: %w", err)
	}

	// Commands today
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE DATE(timestamp) = CURRENT_DATE").Scan(&s.Today)
	if err != nil {
		return fmt.Errorf("failed to get today's commands: %w", err)
	}

	// Failed commands
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE status = 'error'").Scan(&s.Failures)
	if err != nil {
		return fmt.Errorf("failed to get failed commands: %w", err)
	}

	// Average execution time
	err = pool.QueryRow(ctx,
		"SELECT COALESCE(AVG(execution_time_ms), 0), COALESCE(MAX(execution_time_ms), 0) FROM command_audit").Scan(&s.AvgTime, &s.MaxTime)
	if err != nil {
		return fmt.Errorf("failed to get execution times: %w", err)
	}

	printSection("📊 Audit Trail Statistics")

	failureRate := 0.0
	if s.Total > 0 {
		failureRate = (float64(s.Failures) / float64(s.Total)) * 100
	}

	// Color code the health status
	healthStatus := fmt.Sprintf("%s✓ HEALTHY%s", colorGreen, colorReset)
	if failureRate > 5 {
		healthStatus = fmt.Sprintf("%s⚠ WARNING%s", colorYellow, colorReset)
	}
	if failureRate > 10 {
		healthStatus = fmt.Sprintf("%s✗ CRITICAL%s", colorRed, colorReset)
	}

	fmt.Printf("  System Health:     %s\n", healthStatus)
	fmt.Printf("  Total Commands:    %s%d%s\n", colorBold, s.Total, colorReset)
	fmt.Printf("  Commands Today:    %s%d%s\n", colorBold, s.Today, colorReset)
	fmt.Printf("  Failed Commands:   %s%d%s\n", colorBold, s.Failures, colorReset)

	failureRateColor := colorGreen
	if failureRate > 5 {
		failureRateColor = colorYellow
	}
	fmt.Printf("  Failure Rate:      %s%.2f%%%s\n", failureRateColor, failureRate, colorReset)
	fmt.Printf("  Avg Exec Time:     %s%.2f ms%s\n", colorBold, s.AvgTime, colorReset)
	fmt.Printf("  Max Exec Time:     %s%d ms%s\n", colorBold, s.MaxTime, colorReset)

	printSuccess("Statistics retrieved successfully")
	return nil
}

// queryTopCommands shows most frequently used commands
func queryTopCommands(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("top-commands", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "Number of commands to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching top %d commands by execution count...", *limit))

	rows, err := pool.Query(ctx, `
		SELECT command_name, COUNT(*) as count, 
		       SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as failures,
		       ROUND(AVG(execution_time_ms)::numeric, 2) as avg_time
		FROM command_audit
		GROUP BY command_name
		ORDER BY count DESC
		LIMIT $1
	`, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection("🏆 Top Commands")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "  %sCommand%s\t%sCount%s\t%sFailures%s\t%sAvg Time (ms)%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s────────────────%s\t%s─────%s\t%s────────%s\t%s──────────────%s\n", colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset)

	rowCount := 0
	for rows.Next() {
		var cmd string
		var count, failures int64
		var avgTime float64

		if err := rows.Scan(&cmd, &count, &failures, &avgTime); err != nil {
			return err
		}

		rowCount++
		fmt.Fprintf(w, "  %d. %s\t%d\t%d\t%.2f\n", rowCount, cmd, count, failures, avgTime)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d commands", rowCount))
	return rows.Err()
}

// queryTopUsers shows users who executed the most commands
func queryTopUsers(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("top-users", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "Number of users to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching top %d users by command count...", *limit))

	rows, err := pool.Query(ctx, `
		SELECT username, user_id, COUNT(*) as count
		FROM command_audit
		GROUP BY username, user_id
		ORDER BY count DESC
		LIMIT $1
	`, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection("👥 Top Users")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "  %sUsername%s\t%sUserID%s\t%sCommand Count%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s────────────────%s\t%s──────────────────%s\t%s──────────────%s\n", colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset)

	rowCount := 0
	for rows.Next() {
		var username, userID string
		var count int64

		if err := rows.Scan(&username, &userID, &count); err != nil {
			return err
		}

		rowCount++
		fmt.Fprintf(w, "  %d. %s\t%s\t%d\n", rowCount, username, userID, count)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d users", rowCount))
	return rows.Err()
}

// queryFailures shows failed command executions
func queryFailures(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("failures", flag.ContinueOnError)
	hours := fs.Int("hours", 24, "Look back N hours")
	limit := fs.Int("limit", 20, "Number of failures to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching failed commands from last %d hours (limit %d)...", *hours, *limit))

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, username, error_message
		FROM command_audit
		WHERE status = 'error' AND timestamp > NOW() - INTERVAL '1 hour' * $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *hours, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection(fmt.Sprintf("❌ Failed Commands (Last %d Hours)", *hours))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  %sTime%s\t%sCommand%s\t%sUser%s\t%sError%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s──────────────%s\t%s──────────%s\t%s────────────%s\t%s──────────────────────────────%s\n", colorRed, colorReset, colorRed, colorReset, colorRed, colorReset, colorRed, colorReset)

	rowCount := 0
	for rows.Next() {
		var timestamp time.Time
		var command, user string
		var errMsg *string

		if err := rows.Scan(&timestamp, &command, &user, &errMsg); err != nil {
			return err
		}

		rowCount++
		errText := ""
		if errMsg != nil {
			errText = *errMsg
			if len(errText) > 30 {
				errText = errText[:30] + "..."
			}
		}

		fmt.Fprintf(w, "  %s%s%s\t%s%s%s\t%s%s%s\t%s\n",
			colorYellow, timestamp.Format("15:04:05"), colorReset,
			colorBold, command, colorReset,
			colorBold, user, colorReset,
			errText)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d failed commands", rowCount))
	return rows.Err()
}

// queryUserActivity shows activity for a specific user
func queryUserActivity(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("user-activity", flag.ContinueOnError)
	userID := fs.String("user-id", "", "User ID to query")
	limit := fs.Int("limit", 50, "Number of records to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *userID == "" {
		return fmt.Errorf("--user-id flag is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching activity for user %s (last %d records)...", *userID, *limit))

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, status, execution_time_ms
		FROM command_audit
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *userID, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection(fmt.Sprintf("📋 Activity for User: %s", *userID))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  %sTime%s\t%sCommand%s\t%sStatus%s\t%sExec Time (ms)%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s──────────────%s\t%s──────────%s\t%s────────%s\t%s──────────────%s\n", colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset)

	rowCount := 0
	for rows.Next() {
		var timestamp time.Time
		var command, status string
		var execTime int32

		if err := rows.Scan(&timestamp, &command, &status, &execTime); err != nil {
			return err
		}

		rowCount++
		statusColor := colorGreen
		if status == "error" {
			statusColor = colorRed
		}

		fmt.Fprintf(w, "  %s%s%s\t%s%s%s\t%s%s%s\t%d\n",
			colorCyan, timestamp.Format("15:04:05"), colorReset,
			colorBold, command, colorReset,
			statusColor, status, colorReset,
			execTime)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d activity records", rowCount))
	return rows.Err()
}

// queryCommandHistory shows history for a specific command
func queryCommandHistory(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("command-history", flag.ContinueOnError)
	command := fs.String("command", "", "Command name to query")
	limit := fs.Int("limit", 50, "Number of records to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *command == "" {
		return fmt.Errorf("--command flag is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching history for command %s (last %d executions)...", *command, *limit))

	rows, err := pool.Query(ctx, `
		SELECT timestamp, username, status, execution_time_ms
		FROM command_audit
		WHERE command_name = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *command, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection(fmt.Sprintf("⏱ History for Command: %s", *command))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  %sTime%s\t%sUser%s\t%sStatus%s\t%sExec Time (ms)%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s──────────────%s\t%s────────────────%s\t%s────────%s\t%s──────────────%s\n", colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset)

	rowCount := 0
	for rows.Next() {
		var timestamp time.Time
		var user, status string
		var execTime int32

		if err := rows.Scan(&timestamp, &user, &status, &execTime); err != nil {
			return err
		}

		rowCount++
		statusColor := colorGreen
		if status == "error" {
			statusColor = colorRed
		}

		fmt.Fprintf(w, "  %s%s%s\t%s%s%s\t%s%s%s\t%d\n",
			colorCyan, timestamp.Format("15:04:05"), colorReset,
			colorBold, user, colorReset,
			statusColor, status, colorReset,
			execTime)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d execution records", rowCount))
	return rows.Err()
}

// queryAuditByDate shows audit records from a specific date
func queryAuditByDate(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("audit-by-date", flag.ContinueOnError)
	date := fs.String("date", time.Now().Format("2006-01-02"), "Date to query (YYYY-MM-DD)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching audit records for %s...", *date))

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, username, status
		FROM command_audit
		WHERE DATE(timestamp) = $1
		ORDER BY timestamp DESC
	`, *date)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection(fmt.Sprintf("📅 Audit Records for %s", *date))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  %sTime%s\t%sCommand%s\t%sUser%s\t%sStatus%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s──────────────%s\t%s──────────%s\t%s────────────────%s\t%s────────%s\n", colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset, colorCyan, colorReset)

	rowCount := 0
	for rows.Next() {
		var timestamp time.Time
		var command, user, status string

		if err := rows.Scan(&timestamp, &command, &user, &status); err != nil {
			return err
		}

		rowCount++
		statusColor := colorGreen
		if status == "error" {
			statusColor = colorRed
		}

		fmt.Fprintf(w, "  %s%s%s\t%s%s%s\t%s%s%s\t%s%s%s\n",
			colorCyan, timestamp.Format("15:04:05"), colorReset,
			colorBold, command, colorReset,
			colorBold, user, colorReset,
			statusColor, status, colorReset)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d records for %s", rowCount, *date))
	return rows.Err()
}

// querySlowCommands shows the slowest executing commands
func querySlowCommands(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("slow-commands", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "Number of commands to show")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	printInfo(fmt.Sprintf("Fetching %d slowest commands by average execution time...", *limit))

	rows, err := pool.Query(ctx, `
		SELECT command_name, MAX(execution_time_ms) as max_time, 
		       AVG(execution_time_ms) as avg_time, COUNT(*) as count
		FROM command_audit
		GROUP BY command_name
		ORDER BY avg_time DESC
		LIMIT $1
	`, *limit)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	printSection("🐢 Slowest Commands")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "  %sCommand%s\t%sMax (ms)%s\t%sAvg (ms)%s\t%sCount%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Fprintf(w, "  %s────────────────%s\t%s─────────%s\t%s─────────%s\t%s─────%s\n", colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)

	rowCount := 0
	for rows.Next() {
		var command string
		var maxTime, avgTime int64
		var count int64

		if err := rows.Scan(&command, &maxTime, &avgTime, &count); err != nil {
			return err
		}

		rowCount++
		// Color code slow commands
		timeColor := colorGreen
		if avgTime > 500 {
			timeColor = colorYellow
		}
		if avgTime > 1000 {
			timeColor = colorRed
		}

		fmt.Fprintf(w, "  %d. %s\t%s%d%s\t%s%d%s\t%d\n", rowCount, command, timeColor, maxTime, colorReset, timeColor, avgTime, colorReset, count)
	}

	w.Flush()
	printSuccess(fmt.Sprintf("Retrieved %d commands", rowCount))
	return rows.Err()
}
