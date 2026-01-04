package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const userSelectFields = `
    id,
    email,
    full_name,
    region,
    primary_currency,
    role,
    status,
    permissions,
    created_at,
    updated_at,
    last_login_at,
    password_hash
`

// ListUsersOptions allows callers to filter and paginate results.
type ListUsersOptions struct {
	Role      Role
	Status    string
	Search    string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// Repository defines authentication persistence behavior.
type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, opts ListUsersOptions) ([]*User, int, error)
	DeleteUser(ctx context.Context, id string) error
}

// InMemoryRepo is a lightweight repository used during initial development.
type InMemoryRepo struct {
	mu           sync.RWMutex
	usersByID    map[string]*User
	usersByEmail map[string]*User
}

// NewInMemoryRepo constructs an in-memory user repository.
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		usersByID:    make(map[string]*User),
		usersByEmail: make(map[string]*User),
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *InMemoryRepo) CreateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	email := normalizeEmail(user.Email)
	if _, ok := r.usersByEmail[email]; ok {
		return appErrors.UserAlreadyExists
	}
	user.Email = email
	r.usersByEmail[email] = user
	r.usersByID[user.ID] = user
	return nil
}

func (r *InMemoryRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	email = normalizeEmail(email)

	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, appErrors.UserNotFound
	}
	if user.Status == "deleted" {
		return nil, appErrors.UserNotFound
	}
	return user, nil
}

func (r *InMemoryRepo) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
	if !ok {
		return nil, appErrors.UserNotFound
	}
	if user.Status == "deleted" {
		return nil, appErrors.UserNotFound
	}
	return user, nil
}

func (r *InMemoryRepo) UpdateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	email := normalizeEmail(user.Email)
	if existing, ok := r.usersByEmail[email]; ok && existing.ID != user.ID {
		return appErrors.UserAlreadyExists
	}
	r.usersByID[user.ID] = user
	r.usersByEmail[email] = user
	return nil
}

func (r *InMemoryRepo) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.usersByID[id]
	if !ok {
		return appErrors.UserNotFound
	}
	user.Status = "deleted"
	r.usersByID[id] = user
	r.usersByEmail[normalizeEmail(user.Email)] = user
	return nil
}

func (r *InMemoryRepo) ListUsers(ctx context.Context, opts ListUsersOptions) ([]*User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*User, 0, len(r.usersByID))
	search := strings.ToLower(strings.TrimSpace(opts.Search))
	for _, user := range r.usersByID {
		if user == nil {
			continue
		}
		if user.Status == "deleted" {
			continue
		}
		if opts.Role != "" && user.Role != opts.Role {
			continue
		}
		if opts.Status != "" && !strings.EqualFold(user.Status, opts.Status) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(user.FullName), search) && !strings.Contains(strings.ToLower(user.Email), search) {
			continue
		}
		results = append(results, user)
	}

	sortBy := cleanSortField(opts.SortBy)
	sortOrder := strings.ToLower(opts.SortOrder)
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	sort.Slice(results, func(i, j int) bool {
		var left, right time.Time
		switch sortBy {
		case "last_login_at":
			left = parseTime(results[i].LastLoginAt)
			right = parseTime(results[j].LastLoginAt)
		default:
			left = parseTime(results[i].CreatedAt)
			right = parseTime(results[j].CreatedAt)
		}
		if sortOrder == "asc" {
			return left.Before(right)
		}
		return right.Before(left)
	})

	total := len(results)
	page, limit := sanitizePagination(opts.Page, opts.Limit)
	start := (page - 1) * limit
	if start > total {
		return []*User{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}

	snapshot := make([]*User, 0, end-start)
	for _, user := range results[start:end] {
		snapshot = append(snapshot, sanitizeUser(user))
	}
	return snapshot, total, nil
}

func cleanSortField(value string) string {
	switch strings.ToLower(value) {
	case "last_login_at":
		return "last_login_at"
	default:
		return "created_at"
	}
}

func sanitizePagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return page, limit
}

func parseTime(value string) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t
}

func sanitizeUser(u *User) *User {
	if u == nil {
		return nil
	}
	copy := *u
	copy.PasswordHash = ""
	return &copy
}

// PostgresRepository persists users in PostgreSQL.
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository returns a repository backed by SQL.
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	perms, err := json.Marshal(user.Permissions)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO users (
			id, email, full_name, password_hash, region, primary_currency,
			role, status, permissions, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`, user.ID, normalizeEmail(user.Email), user.FullName, user.PasswordHash, user.Region, user.PrimaryCurrency, user.Role, user.Status, perms, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation (duplicate email)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return appErrors.UserAlreadyExists
		}
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE email = $1 AND deleted_at IS NULL", userSelectFields)
	return r.fetchUser(ctx, query, normalizeEmail(email))
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1 AND deleted_at IS NULL", userSelectFields)
	return r.fetchUser(ctx, query, id)
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) error {
	perms := []byte("[]")
	if len(user.Permissions) > 0 {
		data, err := json.Marshal(user.Permissions)
		if err != nil {
			return err
		}
		perms = data
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			full_name = $1,
			region = $2,
			primary_currency = $3,
			role = $4,
			status = $5,
			permissions = $6,
			last_login_at = $7,
			updated_at = $8,
			password_hash = $9
		WHERE id = $10 AND deleted_at IS NULL
	`, user.FullName, user.Region, user.PrimaryCurrency, user.Role, user.Status, perms, user.LastLoginAt, user.UpdatedAt, user.PasswordHash, user.ID)
	if err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			status = 'deleted',
			deleted_at = now(),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (r *PostgresRepository) ListUsers(ctx context.Context, opts ListUsersOptions) ([]*User, int, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	sortField := cleanSortField(opts.SortBy)
	sortOrder := strings.ToUpper(opts.SortOrder)
	if sortOrder != "ASC" {
		sortOrder = "DESC"
	}

	whereClauses := []string{"deleted_at IS NULL"}
	args := make([]interface{}, 0, 6)

	if opts.Role != "" {
		whereClauses = append(whereClauses, "role = $"+fmt.Sprint(len(args)+1))
		args = append(args, opts.Role)
	}
	if opts.Status != "" {
		whereClauses = append(whereClauses, "LOWER(status) = LOWER($"+fmt.Sprint(len(args)+1)+")")
		args = append(args, opts.Status)
	}
	if opts.Search != "" {
		search := "%" + strings.ToLower(strings.TrimSpace(opts.Search)) + "%"
		whereClauses = append(whereClauses, "(LOWER(full_name) LIKE $"+fmt.Sprint(len(args)+1)+" OR LOWER(email) LIKE $"+fmt.Sprint(len(args)+1)+")")
		args = append(args, search)
	}

	whereClause := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, appErrors.DatabaseError
	}

	offset := (opts.Page - 1) * opts.Limit
	limitStmt := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, opts.Limit, offset)
	query := fmt.Sprintf(`
		SELECT %s FROM users
		WHERE %s
		ORDER BY %s %s
		%s
	`, userSelectFields, whereClause, sortField, sortOrder, limitStmt)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, appErrors.DatabaseError
	}
	defer rows.Close()

	result := make([]*User, 0)
	for rows.Next() {
		var row userRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, appErrors.DatabaseError
		}
		user, err := mapRowToUser(row)
		if err != nil {
			return nil, 0, appErrors.DatabaseError
		}
		result = append(result, sanitizeUser(user))
	}
	return result, total, nil
}

func (r *PostgresRepository) fetchUser(ctx context.Context, query string, args ...interface{}) (*User, error) {
	var row userRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.UserNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToUser(row)
}

type userRow struct {
	ID              string         `db:"id"`
	Email           string         `db:"email"`
	FullName        string         `db:"full_name"`
	Region          string         `db:"region"`
	PrimaryCurrency string         `db:"primary_currency"`
	Role            Role           `db:"role"`
	Status          string         `db:"status"`
	Permissions     sql.NullString `db:"permissions"`
	CreatedAt       string         `db:"created_at"`
	UpdatedAt       string         `db:"updated_at"`
	LastLoginAt     sql.NullString `db:"last_login_at"`
	PasswordHash    string         `db:"password_hash"`
}

func mapRowToUser(row userRow) (*User, error) {
	perms := []string{}
	if row.Permissions.Valid {
		if err := json.Unmarshal([]byte(row.Permissions.String), &perms); err != nil {
			return nil, err
		}
	}
	user := &User{
		ID:              row.ID,
		Email:           row.Email,
		FullName:        row.FullName,
		Region:          row.Region,
		PrimaryCurrency: row.PrimaryCurrency,
		Role:            row.Role,
		Status:          row.Status,
		Permissions:     perms,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.LastLoginAt.Valid {
		user.LastLoginAt = row.LastLoginAt.String
	}
	user.PasswordHash = row.PasswordHash
	return user, nil
}
