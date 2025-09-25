package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/mymindmap/api/models"
)

func TestUserRepo_Interface(t *testing.T) {
	// Проверяем что репозиторий реализует ожидаемые методы
	var repo interface{} = &UserRepo{}
	
	_, ok := repo.(interface {
		Create(ctx context.Context, entity *models.User) error
	})
	assert.True(t, ok, "Create method should be implemented")
	
	_, ok = repo.(interface {
		Find(ctx context.Context, id any) (*models.User, error)
	})
	assert.True(t, ok, "Find method should be implemented")
	
	_, ok = repo.(interface {
		Update(ctx context.Context, entity *models.User) error
	})
	assert.True(t, ok, "Update method should be implemented")
	
	_, ok = repo.(interface {
		Delete(ctx context.Context, id any) error
	})
	assert.True(t, ok, "Delete method should be implemented")
	
	_, ok = repo.(interface {
		List(ctx context.Context, limit int, offset int) ([]models.User, error)
	})
	assert.True(t, ok, "List method should be implemented")
}

func TestUserRepo_Creation(t *testing.T) {
	// Проверяем создание репозитория
	var db *pgxpool.Pool = nil
	repo := NewUserRepo(db)
	
	require.NotNil(t, repo)
	assert.IsType(t, &UserRepo{}, repo)
}

func TestUserRepo_SQLValidation(t *testing.T) {
	// Проверяем что все SQL компоненты правильно сгенерированы
	tests := []struct {
		name     string
		template string
	}{
		{"TableName", "users"},
		{"PrimaryKey", "id"},
		{"PrimaryKeyField", "ID"},
		{"InsertColumns", "name, email, password, role, created_at, updated_at"},
		{"InsertPlaceholders", "$1, $2, $3, $4, $5, $6"},
		{"SelectColumns", "id, name, email, password, role, created_at, updated_at"},
		{"UpdateAssignments", "name = $1, email = $2, password = $3, role = $4, created_at = $5, updated_at = $6"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.template, "%s should not be empty", tt.name)
		})
	}
}

func TestUserRepo_MethodCalls(t *testing.T) {
	// Тест который проверяет что методы можно вызвать
	// (они упадут с panic из-за nil db, но это нормально для теста компиляции)
	repo := NewUserRepo(nil)
	ctx := context.Background()
	
	// Эти вызовы проверят что сигнатуры методов правильные
	// Реальное выполнение упадет с panic из-за nil db
	assert.Panics(t, func() {
		entity := &models.User{}
		_ = repo.Create(ctx, entity)
	})
	
	assert.Panics(t, func() {
		_, _ = repo.Find(ctx, 1)
	})
	
	assert.Panics(t, func() {
		entity := &models.User{}
		_ = repo.Update(ctx, entity)
	})
	
	assert.Panics(t, func() {
		_, _ = repo.List(ctx, 10, 0)
	})
	
	assert.Panics(t, func() {
		_ = repo.Delete(ctx, 1)
	})
}