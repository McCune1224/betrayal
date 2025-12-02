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

// auditQuery represents different types of audit queries
type auditQuery struct {
	name        string
	description string
	run         func(pool *pgxpool.Pool, args []string) error
}

var availableQueries = map[string]auditQuery{
	"stats": {
		name:        "stats",
		description: "Show audit trail statistics",
		run:         queryStats,
	},
	"top-commands": {
		name:        "top-commands",
		description: "Show top executed commands",
		run:         queryTopCommands,
	},
	"top-users": {
		name:        "top-users",
		description: "Show top command executers",
		run:         queryTopUsers,
	},
	"failures": {
		name:        "failures",
		description: "Show failed commands",
		run:         queryFailures,
	},
	"user-activity": {
		name:        "user-activity",
		description: "Show activity for a specific user",
		run:         queryUserActivity,
	},
	"command-history": {
		name:        "command-history",
		description: "Show history for a specific command",
		run:         queryCommandHistory,
	},
	"audit-by-date": {
		name:        "audit-by-date",
		description: "Show audit records from a specific date",
		run:         queryAuditByDate,
	},
	"slow-commands": {
		name:        "slow-commands",
		description: "Show slowest executing commands",
		run:         querySlowCommands,
	},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Create database pool
	dsn := os.Getenv("DATABASE_POOLER_URL")
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "Error: DATABASE_POOLER_URL environment variable not set\n")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating database pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Check database connection
	if err := pool.Ping(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	queryType := os.Args[1]
	query, exists := availableQueries[queryType]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Unknown query type '%s'\n", queryType)
		printUsage()
		os.Exit(1)
	}

	if err := query.run(pool, os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Betrayal Audit Analysis Tool")
	fmt.Println("\nUsage: betrayal-audit <query-type> [options]\n")
	fmt.Println("Available queries:")

	for _, query := range availableQueries {
		fmt.Printf("  %-20s %s\n", query.name, query.description)
	}

	fmt.Println("\nExamples:")
	fmt.Println("  betrayal-audit stats")
	fmt.Println("  betrayal-audit top-commands --limit 10")
	fmt.Println("  betrayal-audit user-activity --user-id 123456789")
	fmt.Println("  betrayal-audit command-history --command /setup")
	fmt.Println("  betrayal-audit failures --hours 24")
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
	}

	var s stats

	// Total commands
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM command_audit").Scan(&s.Total)
	if err != nil {
		return err
	}

	// Commands today
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE DATE(timestamp) = CURRENT_DATE").Scan(&s.Today)
	if err != nil {
		return err
	}

	// Failed commands
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE status = 'error'").Scan(&s.Failures)
	if err != nil {
		return err
	}

	// Average execution time
	err = pool.QueryRow(ctx,
		"SELECT COALESCE(AVG(execution_time_ms), 0) FROM command_audit").Scan(&s.AvgTime)
	if err != nil {
		return err
	}

	fmt.Printf("Audit Trail Statistics\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Total Commands:      %d\n", s.Total)
	fmt.Printf("Commands Today:      %d\n", s.Today)
	fmt.Printf("Failed Commands:     %d\n", s.Failures)
	fmt.Printf("Average Exec Time:   %.2f ms\n", s.AvgTime)
	fmt.Printf("Failure Rate:        %.2f%%\n", float64(s.Failures)/float64(s.Total)*100)

	return nil
}

// queryTopCommands shows most frequently used commands
func queryTopCommands(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("top-commands", flag.ExitOnError)
	limit := fs.Int("limit", 10, "Number of commands to show")
	fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
		return err
	}
	defer rows.Close()

	fmt.Println("Top Commands")
	fmt.Println("============")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Command\tCount\tFailures\tAvg Time (ms)")
	fmt.Fprintln(w, "-------\t-----\t--------\t-------------")

	for rows.Next() {
		var cmd string
		var count, failures int64
		var avgTime float64

		if err := rows.Scan(&cmd, &count, &failures, &avgTime); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%.2f\n", cmd, count, failures, avgTime)
	}

	w.Flush()
	return rows.Err()
}

// queryTopUsers shows users who executed the most commands
func queryTopUsers(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("top-users", flag.ExitOnError)
	limit := fs.Int("limit", 10, "Number of users to show")
	fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT username, user_id, COUNT(*) as count
		FROM command_audit
		GROUP BY username, user_id
		ORDER BY count DESC
		LIMIT $1
	`, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Top Users")
	fmt.Println("=========")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Username\tUserID\tCount")
	fmt.Fprintln(w, "--------\t------\t-----")

	for rows.Next() {
		var username, userID string
		var count int64

		if err := rows.Scan(&username, &userID, &count); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\t%d\n", username, userID, count)
	}

	w.Flush()
	return rows.Err()
}

// queryFailures shows failed command executions
func queryFailures(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("failures", flag.ExitOnError)
	hours := fs.Int("hours", 24, "Look back N hours")
	limit := fs.Int("limit", 20, "Number of failures to show")
	fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, username, error_message
		FROM command_audit
		WHERE status = 'error' AND timestamp > NOW() - INTERVAL '1 hour' * $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *hours, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("Failed Commands (Last %d Hours)\n", *hours)
	fmt.Println("================================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Time\tCommand\tUser\tError")
	fmt.Fprintln(w, "----\t-------\t----\t-----")

	for rows.Next() {
		var timestamp time.Time
		var command, user string
		var errMsg *string

		if err := rows.Scan(&timestamp, &command, &user, &errMsg); err != nil {
			return err
		}

		err := ""
		if errMsg != nil {
			err = *errMsg
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			timestamp.Format("15:04:05"),
			command, user, err)
	}

	w.Flush()
	return rows.Err()
}

// queryUserActivity shows activity for a specific user
func queryUserActivity(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("user-activity", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID to query")
	limit := fs.Int("limit", 50, "Number of records to show")
	fs.Parse(args)

	if *userID == "" {
		return fmt.Errorf("--user-id flag is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, status, execution_time_ms
		FROM command_audit
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *userID, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("Activity for User: %s\n", *userID)
	fmt.Println("====================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Time\tCommand\tStatus\tExec Time (ms)")
	fmt.Fprintln(w, "----\t-------\t------\t--------------")

	for rows.Next() {
		var timestamp time.Time
		var command, status string
		var execTime int32

		if err := rows.Scan(&timestamp, &command, &status, &execTime); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
			timestamp.Format("15:04:05"),
			command, status, execTime)
	}

	w.Flush()
	return rows.Err()
}

// queryCommandHistory shows history for a specific command
func queryCommandHistory(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("command-history", flag.ExitOnError)
	command := fs.String("command", "", "Command name to query")
	limit := fs.Int("limit", 50, "Number of records to show")
	fs.Parse(args)

	if *command == "" {
		return fmt.Errorf("--command flag is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT timestamp, username, status, execution_time_ms
		FROM command_audit
		WHERE command_name = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, *command, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("History for Command: %s\n", *command)
	fmt.Println("======================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Time\tUser\tStatus\tExec Time (ms)")
	fmt.Fprintln(w, "----\t----\t------\t--------------")

	for rows.Next() {
		var timestamp time.Time
		var user, status string
		var execTime int32

		if err := rows.Scan(&timestamp, &user, &status, &execTime); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
			timestamp.Format("15:04:05"),
			user, status, execTime)
	}

	w.Flush()
	return rows.Err()
}

// queryAuditByDate shows audit records from a specific date
func queryAuditByDate(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("audit-by-date", flag.ExitOnError)
	date := fs.String("date", time.Now().Format("2006-01-02"), "Date to query (YYYY-MM-DD)")
	fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT timestamp, command_name, username, status
		FROM command_audit
		WHERE DATE(timestamp) = $1
		ORDER BY timestamp DESC
	`, *date)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("Audit Records for %s\n", *date)
	fmt.Println("=====================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Time\tCommand\tUser\tStatus")
	fmt.Fprintln(w, "----\t-------\t----\t------")

	for rows.Next() {
		var timestamp time.Time
		var command, user, status string

		if err := rows.Scan(&timestamp, &command, &user, &status); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			timestamp.Format("15:04:05"),
			command, user, status)
	}

	w.Flush()
	return rows.Err()
}

// querySlowCommands shows the slowest executing commands
func querySlowCommands(pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("slow-commands", flag.ExitOnError)
	limit := fs.Int("limit", 10, "Number of commands to show")
	fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT command_name, MAX(execution_time_ms) as max_time, 
		       AVG(execution_time_ms) as avg_time, COUNT(*) as count
		FROM command_audit
		GROUP BY command_name
		ORDER BY avg_time DESC
		LIMIT $1
	`, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Slowest Commands")
	fmt.Println("================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Command\tMax Time (ms)\tAvg Time (ms)\tCount")
	fmt.Fprintln(w, "-------\t-----------\t-----------\t-----")

	for rows.Next() {
		var command string
		var maxTime, avgTime int64
		var count int64

		if err := rows.Scan(&command, &maxTime, &avgTime, &count); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", command, maxTime, avgTime, count)
	}

	w.Flush()
	return rows.Err()
}
