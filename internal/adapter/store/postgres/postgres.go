package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/dedpnd/unifier/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

type DataBase struct {
	pool *pgxpool.Pool
}

type ErrUserUniq struct {
	Login string
	Err   error
}

func (e *ErrUserUniq) Error() string {
	return fmt.Sprintf("unique violation for login: %s", e.Login)
}

func (e *ErrUserUniq) Unwrap() error {
	return e.Err
}

// var ErrUserUniq = errors.New("such a user exists")

func NewDB(ctx context.Context, dsn string, lg *zap.Logger) (DataBase, error) {
	pool, err := connection(ctx, dsn)
	if err != nil {
		return DataBase{}, fmt.Errorf("failed connection to postgre: %w", err)
	}

	lg.Info("Connection to postgre: success")

	if err := runMigrations(dsn); err != nil {
		return DataBase{}, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	return DataBase{
		pool: pool,
	}, nil
}

func (db DataBase) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}

	return nil
}

//nolint:dupl // This legal code
func (db DataBase) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	row := db.pool.QueryRow(ctx,
		`SELECT id, login, hash FROM users WHERE login=$1`,
		login,
	)

	u := models.User{}
	err := row.Scan(&u.ID, &u.Login, &u.Hash)
	if err != nil {
		var pgErr *pgconn.PgError
		// Если данные не найдены возвращаем пустую структуру
		if !(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.NoDataFound) {
			return u, nil
		}

		return u, fmt.Errorf("failed scan row: %w", err)
	}

	return u, nil
}

func (db DataBase) CreateUser(ctx context.Context, user models.User) (int, error) {
	row := db.pool.QueryRow(ctx,
		`INSERT INTO users (Login, Hash) VALUES($1, $2) RETURNING id`,
		user.Login,
		user.Hash,
	)

	var id int
	err := row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, &ErrUserUniq{Login: user.Login, Err: err}
		}

		return 0, fmt.Errorf("failed scan row: %w", err)
	}

	return id, nil
}

func (db DataBase) GetAllRules(ctx context.Context) ([]models.Rule, error) {
	rows, err := db.pool.Query(ctx, `SELECT ID, Rule, Owner FROM Rules`)
	if err != nil {
		return nil, fmt.Errorf("failed rules query records: %w", err)
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var rule models.Rule
		if err = rows.Scan(&rule.ID, &rule.Rule, &rule.Owner); err != nil {
			return nil, fmt.Errorf("failed scan rules records: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

//nolint:dupl // This legal code
func (db DataBase) GetRuleByID(ctx context.Context, id int) (models.Rule, error) {
	row := db.pool.QueryRow(ctx,
		`SELECT ID, Rule, Owner FROM Rules WHERE id=$1`,
		id,
	)

	r := models.Rule{}
	err := row.Scan(&r.ID, &r.Rule, &r.Owner)
	if err != nil {
		var pgErr *pgconn.PgError
		// Если данные не найдены возвращаем пустую структуру
		if !(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.NoDataFound) {
			return r, nil
		}

		return r, fmt.Errorf("failed scan row: %w", err)
	}

	return r, nil
}

func (db DataBase) CreateRule(ctx context.Context, rule models.Config, owner int) (int, error) {
	row := db.pool.QueryRow(context.Background(),
		`INSERT INTO rules (Rule, Owner) VALUES($1, $2) RETURNING id`,
		rule,
		owner,
	)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed scan rule record row: %w", err)
	}

	return id, nil
}

func (db DataBase) DeleteRule(ctx context.Context, id int) error {
	_, err := db.pool.Exec(context.Background(),
		`DELETE FROM rules WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed delete record in rules: %w", err)
	}

	return nil
}

// -----------------------.

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func connection(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed connection to postgre: %w", err)
	}

	return pool, nil
}
