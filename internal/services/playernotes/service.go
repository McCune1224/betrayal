package playernotes

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

var ErrUnauthorized = errors.New("player notes authorization required")
var ErrNotFound = errors.New("player note not found")

// Authorization identifies the already-authenticated transport principal.
// The service accepts only explicit principals so callers cannot accidentally
// rely on route or command location as authorization.
type Authorization struct {
	Scope AuthorizationScope
}

type AuthorizationScope string

const (
	DiscordAdminScope AuthorizationScope = "discord-admin"
	WebAdminScope     AuthorizationScope = "web-admin"
)

func DiscordAdminAuthorization() Authorization { return Authorization{Scope: DiscordAdminScope} }
func WebAdminAuthorization() Authorization     { return Authorization{Scope: WebAdminScope} }

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func authorize(auth Authorization) error {
	if auth.Scope != DiscordAdminScope && auth.Scope != WebAdminScope {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) List(ctx context.Context, auth Authorization, playerID int64) ([]models.PlayerNote, error) {
	if err := authorize(auth); err != nil {
		return nil, err
	}
	return models.New(s.pool).ListPlayerNote(ctx, playerID)
}

// Add appends a note after the highest existing position. Positions are not
// reused after deletion so references shown to admins remain stable.
func (s *Service) Add(ctx context.Context, auth Authorization, playerID int64, info string) (*models.PlayerNote, error) {
	if err := authorize(auth); err != nil {
		return nil, err
	}
	if info == "" {
		return nil, errors.New("note cannot be empty")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "select pg_advisory_xact_lock($1)", playerID); err != nil {
		return nil, fmt.Errorf("lock player notes: %w", err)
	}
	q := models.New(tx)
	notes, err := q.ListPlayerNote(ctx, playerID)
	if err != nil {
		return nil, err
	}
	var next int32
	for _, note := range notes {
		if note.Position > next {
			next = note.Position
		}
	}
	note, err := q.CreatePlayerNote(ctx, models.CreatePlayerNoteParams{
		PlayerID: playerID, Position: next + 1, Info: info,
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &note, nil
}

// Update changes an existing note. It deliberately does not create a missing
// position because Discord's /inv notes update command is update-only.
func (s *Service) Update(ctx context.Context, auth Authorization, playerID int64, position int, info string) (*models.PlayerNote, error) {
	if err := authorize(auth); err != nil {
		return nil, err
	}
	if position < 1 {
		return nil, errors.New("position must be positive")
	}
	if info == "" {
		return nil, errors.New("note cannot be empty")
	}
	note, err := models.New(s.pool).UpdatePlayerNoteByPosition(ctx, models.UpdatePlayerNoteByPositionParams{
		PlayerID: playerID, Position: int32(position), Info: info,
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// Save updates a requested position or creates it when absent. The pair is a
// single transaction so concurrent web saves cannot observe a half-completed
// upsert.
func (s *Service) Save(ctx context.Context, auth Authorization, playerID int64, position int, info string) (*models.PlayerNote, error) {
	if err := authorize(auth); err != nil {
		return nil, err
	}
	if position < 1 {
		return nil, errors.New("position must be positive")
	}
	if info == "" {
		return nil, errors.New("note cannot be empty")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "select pg_advisory_xact_lock($1)", playerID); err != nil {
		return nil, fmt.Errorf("lock player notes: %w", err)
	}
	q := models.New(tx)
	note, err := q.UpdatePlayerNoteByPosition(ctx, models.UpdatePlayerNoteByPositionParams{
		PlayerID: playerID, Position: int32(position), Info: info,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		note, err = q.CreatePlayerNote(ctx, models.CreatePlayerNoteParams{
			PlayerID: playerID, Position: int32(position), Info: info,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("save player note: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &note, nil
}

func (s *Service) Delete(ctx context.Context, auth Authorization, playerID int64, position int32) error {
	if err := authorize(auth); err != nil {
		return err
	}
	if position < 1 {
		return errors.New("position must be positive")
	}
	note, err := models.New(s.pool).GetPlayerNoteByPosition(ctx, models.GetPlayerNoteByPositionParams{
		PlayerID: playerID, Position: position,
	})
	if err != nil {
		return err
	}
	return models.New(s.pool).DeletePlayerNote(ctx, models.DeletePlayerNoteParams{
		PlayerID: playerID, NoteID: note.NoteID,
	})
}

func (s *Service) DeleteByID(ctx context.Context, auth Authorization, playerID int64, noteID int32) error {
	if err := authorize(auth); err != nil {
		return err
	}
	if _, err := models.New(s.pool).GetPlayerNote(ctx, models.GetPlayerNoteParams{
		PlayerID: playerID, NoteID: noteID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return models.New(s.pool).DeletePlayerNote(ctx, models.DeletePlayerNoteParams{
		PlayerID: playerID, NoteID: noteID,
	})
}
