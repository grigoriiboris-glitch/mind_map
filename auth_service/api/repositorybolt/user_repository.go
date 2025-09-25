package repositorybolt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	//"log"
	"strconv"
	"time"

	"github.com/mymindmap/api/models"
	"go.etcd.io/bbolt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	db *bbolt.DB
}

func NewUserRepository(db *bbolt.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Initialize() error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		
		_, err = tx.CreateBucketIfNotExists([]byte("user_emails"))
		return err
	})
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		emailsBucket := tx.Bucket([]byte("user_emails"))
		
		if usersBucket == nil || emailsBucket == nil {
			return errors.New("buckets not initialized")
		}

		// Проверяем, существует ли пользователь с таким email
		existingID := emailsBucket.Get([]byte(user.Email))
		if existingID != nil {
			return errors.New("user with this email already exists")
		}

		// Генерируем ID
		id, err := usersBucket.NextSequence()
		if err != nil {
			return fmt.Errorf("generate ID: %w", err)
		}
		user.ID = int(id)

		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		// Сериализуем пользователя
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user: %w", err)
		}

		// Сохраняем пользователя
		idBytes := []byte(strconv.Itoa(user.ID))
		err = usersBucket.Put(idBytes, userJSON)
		if err != nil {
			return fmt.Errorf("save user: %w", err)
		}

		// Сохраняем mapping email -> id
		return emailsBucket.Put([]byte(user.Email), idBytes)
	})
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    
    err := r.db.View(func(tx *bbolt.Tx) error {
        emailsBucket := tx.Bucket([]byte("user_emails"))
        usersBucket := tx.Bucket([]byte("users"))
        
        if emailsBucket == nil || usersBucket == nil {
            return errors.New("buckets not initialized")
        }
        
        // Получаем ID по email
        idBytes := emailsBucket.Get([]byte(email))
        if idBytes == nil {
            return ErrUserNotFound // Возвращаем ошибку внутри транзакции
        }

        // Получаем пользователя по ID
        userJSON := usersBucket.Get(idBytes)
        if userJSON == nil {
            return ErrUserNotFound // Возвращаем ошибку внутри транзакции
        }

        return json.Unmarshal(userJSON, &user)
    })

    if err != nil {
        // ВАЖНО: преобразуем ErrUserNotFound в nil, nil
        if errors.Is(err, ErrUserNotFound) {
            return nil, nil // Пользователь не найден - это не ошибка
        }
        return nil, err // Другие ошибки возвращаем как есть
    }

    return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user *models.User
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		if usersBucket == nil {
			return errors.New("users bucket not initialized")
		}

		idBytes := []byte(strconv.Itoa(id))
		userJSON := usersBucket.Get(idBytes)
		if userJSON == nil {
			return ErrUserNotFound
		}

		user = &models.User{}
		return json.Unmarshal(userJSON, user)
	})

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetSecureUserByID(ctx context.Context, id int) (*models.User, error) {
	var user *models.User
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		if usersBucket == nil {
			return errors.New("users bucket not initialized")
		}

		idBytes := []byte(strconv.Itoa(id))
		userJSON := usersBucket.Get(idBytes)
		if userJSON == nil {
			return ErrUserNotFound
		}

		user = &models.User{}
		return json.Unmarshal(userJSON, user)
	})

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}
	user.Password = ""

	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		emailsBucket := tx.Bucket([]byte("user_emails"))
		
		if usersBucket == nil || emailsBucket == nil {
			return errors.New("buckets not initialized")
		}

		// Получаем текущего пользователя для проверки email
		idBytes := []byte(strconv.Itoa(user.ID))
		existingUserJSON := usersBucket.Get(idBytes)
		if existingUserJSON == nil {
			return ErrUserNotFound
		}

		var existingUser models.User
		if err := json.Unmarshal(existingUserJSON, &existingUser); err != nil {
			return fmt.Errorf("unmarshal existing user: %w", err)
		}

		// Если email изменился, проверяем новый email
		if existingUser.Email != user.Email {
			// Проверяем, не занят ли новый email
			if existingID := emailsBucket.Get([]byte(user.Email)); existingID != nil {
				return errors.New("email already taken")
			}

			// Удаляем старый email mapping
			if err := emailsBucket.Delete([]byte(existingUser.Email)); err != nil {
				return fmt.Errorf("delete old email mapping: %w", err)
			}

			// Добавляем новый email mapping
			if err := emailsBucket.Put([]byte(user.Email), idBytes); err != nil {
				return fmt.Errorf("add new email mapping: %w", err)
			}
		}

		// Обновляем данные
		user.UpdatedAt = time.Now()
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user: %w", err)
		}

		return usersBucket.Put(idBytes, userJSON)
	})
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		emailsBucket := tx.Bucket([]byte("user_emails"))
		
		if usersBucket == nil || emailsBucket == nil {
			return errors.New("buckets not initialized")
		}

		idBytes := []byte(strconv.Itoa(id))
		
		// Получаем пользователя для удаления email mapping
		userJSON := usersBucket.Get(idBytes)
		if userJSON == nil {
			return ErrUserNotFound
		}

		var user models.User
		if err := json.Unmarshal(userJSON, &user); err != nil {
			return fmt.Errorf("unmarshal user: %w", err)
		}

		// Удаляем email mapping
		if err := emailsBucket.Delete([]byte(user.Email)); err != nil {
			return fmt.Errorf("delete email mapping: %w", err)
		}

		// Удаляем пользователя
		return usersBucket.Delete(idBytes)
	})
}

// Дополнительные методы для работы со всеми пользователями
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	
	err := r.db.View(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("users"))
		if usersBucket == nil {
			return errors.New("users bucket not initialized")
		}

		return usersBucket.ForEach(func(k, v []byte) error {
			var user models.User
			if err := json.Unmarshal(v, &user); err != nil {
				return fmt.Errorf("unmarshal user %s: %w", string(k), err)
			}
			users = append(users, &user)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}