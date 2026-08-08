// Package datasync syncs the game catalog (roles/items/abilities/perks) from
// Google Sheets CSV exports into the database.
//
// It is the shared engine behind two surfaces:
//   - the web panel's /sync page (fetch → preview diff → validate → apply), and
//   - the archived CLI (cmd/data-entry), which is now a thin wrapper.
//
// Semantics: UPSERT by exact name. Sheet edits propagate to existing rows
// (description/alignment/rarity/charges/cost), new rows are created, and
// nothing is ever deleted — syncs are additive, matching how the tool is used
// ("run once before a game starts").
package datasync

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// infiniteCharges is the DB encoding for "∞" (unlimited) charges.
const infiniteCharges = int32(999999)

// Action is the outcome of comparing one catalog item against the database.
type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionSkip   Action = "skip"
)

// AbilityDoc is one ability as parsed from a role's CSV chunk.
type AbilityDoc struct {
	Name           string
	Description    string
	DefaultCharges int32
	Rarity         models.Rarity
	AnyAbility     bool
	Categories     []string
}

// PerkDoc is one passive (perk) as parsed from a role's CSV chunk.
type PerkDoc struct {
	Name        string
	Description string
}

// RoleDoc is one role (plus its abilities and passives) parsed from a sheet.
type RoleDoc struct {
	Name        string
	Description string
	Alignment   models.Alignment
	Abilities   []AbilityDoc
	Perks       []PerkDoc
}

// ItemDoc is one item parsed from the items sheet.
type ItemDoc struct {
	Name        string
	Description string
	Rarity      models.Rarity
	Cost        int32
	Categories  []string
}

// AbilityPlan is a diff result for one ability.
type AbilityPlan struct {
	Doc     AbilityDoc
	Action  Action
	Changes []string
	OldDesc string
}

// PerkPlan is a diff result for one perk.
type PerkPlan struct {
	Doc     PerkDoc
	Action  Action
	Changes []string
	OldDesc string
}

// RolePlan is a diff result for one role, including its abilities and perks.
type RolePlan struct {
	Doc       RoleDoc
	Action    Action
	Changes   []string
	OldDesc   string
	Abilities []AbilityPlan
	Perks     []PerkPlan
}

// RoleSourcePlan is the full diff for one roles CSV source.
type RoleSourcePlan struct {
	Alignment models.Alignment
	Roles     []RolePlan
	Counts    map[Action]int
	Warnings  []string
}

// ItemPlan is a diff result for one item.
type ItemPlan struct {
	Doc     ItemDoc
	Action  Action
	Changes []string
	OldDesc string
}

// ItemSourcePlan is the full diff for the items CSV source.
type ItemSourcePlan struct {
	Items    []ItemPlan
	Counts   map[Action]int
	Warnings []string
}

// Source mirrors the sync_source table row for the web UI.
type Source = models.SyncSource

// CanonicalSources are the four spreadsheet feeds seeded at startup. The
// alignment column drives role parsing; items use "".
var CanonicalSources = []struct {
	Name      string
	Kind      string
	Alignment string
}{
	{"good_roles", "roles", string(models.AlignmentGOOD)},
	{"evil_roles", "roles", string(models.AlignmentEVIL)},
	{"neutral_roles", "roles", string(models.AlignmentNEUTRAL)},
	{"items", "items", ""},
}

// envKeys maps each canonical source to the env var holding its CSV URL.
var envKeys = map[string]string{
	"good_roles":    "GOOD_ROLES_CSV",
	"evil_roles":    "EVIL_ROLES_CSV",
	"neutral_roles": "NEUTRAL_ROLES_CSV",
	"items":         "ITEM_CSV",
}

// Service ties the sync engine to a database pool. envURLs maps source name →
// CSV URL from the environment (used only to seed/backfill empty URLs).
type Service struct {
	pool            *pgxpool.Pool
	envURLs         map[string]string
	client          *http.Client
	allowUnsafeURLs bool
}

// New creates a Service. envURLs is the map built from GOOD_ROLES_CSV /
// EVIL_ROLES_CSV / NEUTRAL_ROLES_CSV / ITEM_CSV (empty values are fine).
func New(pool *pgxpool.Pool, envURLs map[string]string) *Service {
	return &Service{
		pool:    pool,
		envURLs: envURLs,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SetAllowUnsafeURLs exists for localhost fixture tests only. Production
// callers must leave it false so arbitrary admin-entered URLs cannot become
// an SSRF primitive.
func (s *Service) SetAllowUnsafeURLs(allow bool) { s.allowUnsafeURLs = allow }

// ChargesDisplay renders the DB charge encoding for the UI ("∞" for
// infiniteCharges, otherwise the number).
func ChargesDisplay(charges int32) string {
	if charges == infiniteCharges {
		return "∞"
	}
	return itoa(int64(charges))
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// Fetch downloads the CSV for a source. The caller closes the returned reader.
func (s *Service) Fetch(ctx context.Context, src Source) (io.ReadCloser, error) {
	parsed, err := url.Parse(src.Url)
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return nil, &URLPolicyError{Reason: "invalid URL"}
	}
	if !s.allowUnsafeURLs && !allowedSyncURL(parsed) {
		return nil, &URLPolicyError{Reason: "URL must be an HTTPS Google Sheets export"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.Url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &HTTPError{Status: resp.StatusCode}
	}
	return resp.Body, nil
}

// allowedSyncURL permits only the Google export hosts used by configured
// spreadsheet sources. Unsafe fixture URLs are enabled explicitly in tests.
func allowedSyncURL(u *url.URL) bool {
	if u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "docs.google.com" || host == "docs.googleusercontent.com" || host == "sheets.googleapis.com"
}

type URLPolicyError struct{ Reason string }

func (e *URLPolicyError) Error() string { return "sync URL rejected: " + e.Reason }

// HTTPError reports a non-200 CSV fetch.
type HTTPError struct{ Status int }

func (e *HTTPError) Error() string {
	return "csv fetch failed: HTTP " + itoa(int64(e.Status))
}
