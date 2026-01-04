package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	dbPool *pgxpool.Pool
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{dbPool: pool}
}

// Dashboard handles GET /
func (h *DashboardHandler) Dashboard(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	cyclePhase, cycleNumber := h.getCycleInfo(ctx, q)
	playersAlive, playersDead, totalPlayers := h.getPlayerCounts(ctx, q)

	activitySummary, summaryErr := q.GetCommandActivitySummary(ctx)
	if summaryErr != nil {
		log.Printf("dashboard: failed to load command activity summary: %v", summaryErr)
	}

	topCommandRows, topCommandsErr := q.ListTopCommandsLastHour(ctx, 5)
	if topCommandsErr != nil {
		log.Printf("dashboard: failed to load top commands: %v", topCommandsErr)
	}

	recentErrorRows, recentErrorsErr := q.ListRecentCommandErrors(ctx, 5)
	if recentErrorsErr != nil {
		log.Printf("dashboard: failed to load recent command errors: %v", recentErrorsErr)
	}

	data := pages.DashboardData{
		CyclePhase:     cyclePhase,
		CycleNumber:    cycleNumber,
		PlayersAlive:   playersAlive,
		PlayersDead:    playersDead,
		TotalPlayers:   totalPlayers,
		CommandSummary: buildCommandSummary(activitySummary, summaryErr),
		TopCommands:    buildTopCommands(topCommandRows, topCommandsErr),
		RecentErrors:   buildRecentErrors(recentErrorRows, recentErrorsErr),
	}

	return render(c, http.StatusOK, pages.Dashboard(data))
}

func (h *DashboardHandler) getCycleInfo(ctx context.Context, q *models.Queries) (string, int) {
	cycle, err := q.GetCycle(ctx)
	if err != nil {
		return "Day", 1
	}

	phase := "Day"
	if cycle.IsElimination {
		phase = "Elimination"
	}

	return phase, int(cycle.Day)
}

func (h *DashboardHandler) getPlayerCounts(ctx context.Context, q *models.Queries) (alive, dead, total int) {
	players, err := q.ListPlayer(ctx)
	if err != nil {
		return 0, 0, 0
	}

	for _, p := range players {
		if p.Alive {
			alive++
		} else {
			dead++
		}
	}

	return alive, dead, len(players)
}

func buildCommandSummary(summary models.GetCommandActivitySummaryRow, err error) *pages.CommandActivitySummaryData {
	if err != nil {
		return nil
	}

	avg := convertAverageDuration(summary.AvgExecutionTimeMsLast24h)

	return &pages.CommandActivitySummaryData{
		CommandsLastHour:     int(summary.CommandsLastHour),
		CommandsLast24h:      int(summary.CommandsLast24h),
		SuccessLast24h:       int(summary.SuccessCountLast24h),
		FailuresLast24h:      int(summary.FailureCountLast24h),
		AdminCommandsLast24h: int(summary.AdminCommandsLast24h),
		AvgExecutionTimeMs:   avg,
	}
}

func buildTopCommands(rows []models.ListTopCommandsLastHourRow, err error) []pages.TopCommandData {
	if err != nil || len(rows) == 0 {
		return nil
	}

	commands := make([]pages.TopCommandData, 0, len(rows))
	for _, row := range rows {
		lastUsed := convertTimestamp(row.LastUsedAt)

		commands = append(commands, pages.TopCommandData{
			CommandName:  row.CommandName,
			UsageCount:   int(row.UsageCount),
			FailureCount: int(row.FailureCount),
			LastUsedAt:   lastUsed,
		})
	}

	return commands
}

func buildRecentErrors(rows []models.ListRecentCommandErrorsRow, err error) []pages.CommandErrorData {
	if err != nil || len(rows) == 0 {
		return nil
	}

	errors := make([]pages.CommandErrorData, 0, len(rows))
	for _, row := range rows {
		occurredAt := time.Time{}
		if row.Timestamp.Valid {
			occurredAt = row.Timestamp.Time
		}

		errors = append(errors, pages.CommandErrorData{
			CorrelationID: formatUUID(row.CorrelationID),
			CommandName:   row.CommandName,
			Arguments:     FormatCommandArguments(row.CommandArguments),
			Username:      row.Username,
			Status:        safeText(row.Status),
			ErrorMessage:  safeText(row.ErrorMessage),
			OccurredAt:    occurredAt,
		})
	}

	return errors
}

func convertTimestamp(value interface{}) time.Time {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v
	case pgtype.Timestamp:
		return v.Time
	default:
		return time.Time{}
	}
}

func convertAverageDuration(value interface{}) int {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int64:
		return int(v)
	case int32:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case pgtype.Numeric:
		res, err := util.IntFromNumeric(v)
		if err != nil {
			log.Printf("dashboard: failed to parse average execution time: %v", err)
			return 0
		}
		return res
	default:
		return 0
	}
}

func formatUUID(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}

	return uuid.UUID(id.Bytes).String()
}

func safeText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

// formatPlayerID formats a Discord user ID for display
func formatPlayerID(id int64) string {
	return fmt.Sprintf("%d", id)
}
