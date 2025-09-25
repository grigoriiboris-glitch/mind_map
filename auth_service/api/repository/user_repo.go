package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/mymindmap/api/models"
)

type UserRepo struct {
    db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, entity *models.User) error {
    query := `INSERT INTO users (name, email, password, role, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6)
              RETURNING id`
    return r.db.QueryRow(ctx, query, entity.Name, entity.Email, entity.Password, entity.Role, entity.CreatedAt, entity.UpdatedAt).
        Scan(&entity.ID)
}

func (r *UserRepo) Find(ctx context.Context, id any) (*models.User, error) {
    query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE id = $1`
    row := r.db.QueryRow(ctx, query, id)

    entity := &models.User{}
    err := row.Scan(&entity.ID, &entity.Name, &entity.Email, &entity.Password, &entity.Role, &entity.CreatedAt, &entity.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return entity, nil
}

func (r *UserRepo) Update(ctx context.Context, entity *models.User) error {
    query := `UPDATE users SET name = $1, email = $2, password = $3, role = $4, created_at = $5, updated_at = $6 WHERE id = $7`
    _, err := r.db.Exec(ctx, query, entity.Name, entity.Email, entity.Password, entity.Role, entity.CreatedAt, entity.UpdatedAt, entity.ID)
    return err
}

func (r *UserRepo) Delete(ctx context.Context, id any) error {
    query := `DELETE FROM users WHERE id = $1`
    _, err := r.db.Exec(ctx, query, id)
    return err
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]models.User, error) {
    query := `SELECT id, name, email, password, role, created_at, updated_at FROM users ORDER BY id LIMIT $1 OFFSET $2`
    rows, err := r.db.Query(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []models.User
    for rows.Next() {
        entity := models.User{}
        err := rows.Scan(&entity.ID, &entity.Name, &entity.Email, &entity.Password, &entity.Role, &entity.CreatedAt, &entity.UpdatedAt)
        if err != nil {
            return nil, err
        }
        result = append(result, entity)
    }
    return result, nil
}
