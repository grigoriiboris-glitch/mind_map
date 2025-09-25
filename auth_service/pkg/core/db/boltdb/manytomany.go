// manytomany/repository.go
package manytomany

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.etcd.io/bbolt"
)

// RelationConfig конфигурация для конкретного типа отношений
type RelationConfig struct {
	Name          string // Уникальное имя отношения (например: "user_mindmaps")
	LeftEntity    string // Название левой сущности (например: "user")
	RightEntity   string // Название правой сущности (например: "mindmap")
	LeftIDType    IDType // Тип ID левой сущности
	RightIDType   IDType // Тип ID правой сущности
}

// IDType тип идентификатора
type IDType int

const (
	IntID IDType = iota
	StringID
)

// ManyToManyRepository универсальный репозиторий для Many-to-Many отношений
type ManyToManyRepository struct {
	db         *bbolt.DB
	configs    map[string]*RelationConfig
	bucketNames map[string][]string
}

// NewRepository создает новый репозиторий
func NewRepository(db *bbolt.DB) *ManyToManyRepository {
	return &ManyToManyRepository{
		db:         db,
		configs:    make(map[string]*RelationConfig),
		bucketNames: make(map[string][]string),
	}
}

// RegisterRelation регистрирует новый тип отношений
func (r *ManyToManyRepository) RegisterRelation(config *RelationConfig) error {
	if config.Name == "" {
		return fmt.Errorf("relation name cannot be empty")
	}
	if _, exists := r.configs[config.Name]; exists {
		return fmt.Errorf("relation '%s' already registered", config.Name)
	}

	r.configs[config.Name] = config
	
	// Генерируем имена бакетов
	leftBucket := fmt.Sprintf("%s_%s", config.LeftEntity, config.RightEntity)
	rightBucket := fmt.Sprintf("%s_%s", config.RightEntity, config.LeftEntity)
	r.bucketNames[config.Name] = []string{leftBucket, rightBucket}

	return r.initializeBuckets(config.Name)
}

// initializeBuckets инициализирует бакеты для отношения
func (r *ManyToManyRepository) initializeBuckets(relationName string) error {
	buckets := r.bucketNames[relationName]
	if len(buckets) != 2 {
		return fmt.Errorf("invalid bucket configuration for relation '%s'", relationName)
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return fmt.Errorf("failed to create bucket '%s': %w", bucketName, err)
			}
		}
		return nil
	})
}

// AddRelation добавляет отношение между двумя сущностями
func (r *ManyToManyRepository) AddRelation(relationName string, leftID, rightID interface{}) error {
	config, buckets, err := r.getConfigAndBuckets(relationName)
	if err != nil {
		return err
	}

	// Валидация ID
	leftIDBytes, err := r.encodeID(leftID, config.LeftIDType)
	if err != nil {
		return fmt.Errorf("invalid left ID: %w", err)
	}
	rightIDBytes, err := r.encodeID(rightID, config.RightIDType)
	if err != nil {
		return fmt.Errorf("invalid right ID: %w", err)
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		leftBucket := tx.Bucket([]byte(buckets[0]))
		rightBucket := tx.Bucket([]byte(buckets[1]))

		// Левая сущность → правые сущности
		if err := r.addToBucket(leftBucket, leftIDBytes, rightID, config.RightIDType); err != nil {
			return fmt.Errorf("failed to add to left bucket: %w", err)
		}

		// Правая сущность → левые сущности
		if err := r.addToBucket(rightBucket, rightIDBytes, leftID, config.LeftIDType); err != nil {
			return fmt.Errorf("failed to add to right bucket: %w", err)
		}

		return nil
	})
}

// RemoveRelation удаляет отношение между двумя сущностями
func (r *ManyToManyRepository) RemoveRelation(relationName string, leftID, rightID interface{}) error {
	config, buckets, err := r.getConfigAndBuckets(relationName)
	if err != nil {
		return err
	}

	leftIDBytes, err := r.encodeID(leftID, config.LeftIDType)
	if err != nil {
		return fmt.Errorf("invalid left ID: %w", err)
	}
	rightIDBytes, err := r.encodeID(rightID, config.RightIDType)
	if err != nil {
		return fmt.Errorf("invalid right ID: %w", err)
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		leftBucket := tx.Bucket([]byte(buckets[0]))
		rightBucket := tx.Bucket([]byte(buckets[1]))

		// Левая сущность → правые сущности
		if err := r.removeFromBucket(leftBucket, leftIDBytes, rightID, config.RightIDType); err != nil {
			return fmt.Errorf("failed to remove from left bucket: %w", err)
		}

		// Правая сущность → левые сущности
		if err := r.removeFromBucket(rightBucket, rightIDBytes, leftID, config.LeftIDType); err != nil {
			return fmt.Errorf("failed to remove from right bucket: %w", err)
		}

		return nil
	})
}

// GetLeftEntities возвращает левые сущности для правой сущности
func (r *ManyToManyRepository) GetLeftEntities(relationName string, rightID interface{}) ([]interface{}, error) {
	config, buckets, err := r.getConfigAndBuckets(relationName)
	if err != nil {
		return nil, err
	}

	rightIDBytes, err := r.encodeID(rightID, config.RightIDType)
	if err != nil {
		return nil, fmt.Errorf("invalid right ID: %w", err)
	}

	var result []interface{}
	err = r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(buckets[1])) // mindmap_users bucket
		data := bucket.Get(rightIDBytes)
		if data == nil {
			return nil
		}

		return r.decodeIDList(data, config.LeftIDType, &result)
	})

	return result, err
}

// GetRightEntities возвращает правые сущности для левой сущности
func (r *ManyToManyRepository) GetRightEntities(relationName string, leftID interface{}) ([]interface{}, error) {
	config, buckets, err := r.getConfigAndBuckets(relationName)
	if err != nil {
		return nil, err
	}

	leftIDBytes, err := r.encodeID(leftID, config.LeftIDType)
	if err != nil {
		return nil, fmt.Errorf("invalid left ID: %w", err)
	}

	var result []interface{}
	err = r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(buckets[0])) // user_mindmaps bucket
		data := bucket.Get(leftIDBytes)
		if data == nil {
			return nil
		}

		return r.decodeIDList(data, config.RightIDType, &result)
	})

	return result, err
}

// HasRelation проверяет существование отношения
func (r *ManyToManyRepository) HasRelation(relationName string, leftID, rightID interface{}) (bool, error) {
	rightEntities, err := r.GetRightEntities(relationName, leftID)
	if err != nil {
		return false, err
	}

	for _, entity := range rightEntities {
		if r.idsEqual(entity, rightID) {
			return true, nil
		}
	}

	return false, nil
}

// Вспомогательные методы

func (r *ManyToManyRepository) getConfigAndBuckets(relationName string) (*RelationConfig, []string, error) {
	config, exists := r.configs[relationName]
	if !exists {
		return nil, nil, fmt.Errorf("relation '%s' not registered", relationName)
	}

	buckets, exists := r.bucketNames[relationName]
	if !exists || len(buckets) != 2 {
		return nil, nil, fmt.Errorf("bucket configuration not found for relation '%s'", relationName)
	}

	return config, buckets, nil
}

func (r *ManyToManyRepository) encodeID(id interface{}, idType IDType) ([]byte, error) {
	switch idType {
	case IntID:
		intID, ok := id.(int)
		if !ok {
			return nil, fmt.Errorf("expected int ID, got %T", id)
		}
		return []byte(strconv.Itoa(intID)), nil
	case StringID:
		strID, ok := id.(string)
		if !ok {
			return nil, fmt.Errorf("expected string ID, got %T", id)
		}
		return []byte(strID), nil
	default:
		return nil, fmt.Errorf("unsupported ID type: %v", idType)
	}
}

func (r *ManyToManyRepository) addToBucket(bucket *bbolt.Bucket, key []byte, valueID interface{}, valueType IDType) error {
	var existingIDs []interface{}
	
	if data := bucket.Get(key); data != nil {
		if err := r.decodeIDList(data, valueType, &existingIDs); err != nil {
			return err
		}
	}

	// Проверяем, существует ли уже отношение
	for _, existingID := range existingIDs {
		if r.idsEqual(existingID, valueID) {
			return nil // Отношение уже существует
		}
	}

	existingIDs = append(existingIDs, valueID)
	data, err := r.encodeIDList(existingIDs, valueType)
	if err != nil {
		return err
	}

	return bucket.Put(key, data)
}

func (r *ManyToManyRepository) removeFromBucket(bucket *bbolt.Bucket, key []byte, valueID interface{}, valueType IDType) error {
	var existingIDs []interface{}
	
	data := bucket.Get(key)
	if data == nil {
		return nil // Ничего нет для удаления
	}

	if err := r.decodeIDList(data, valueType, &existingIDs); err != nil {
		return err
	}

	// Фильтруем массив
	var newIDs []interface{}
	for _, existingID := range existingIDs {
		if !r.idsEqual(existingID, valueID) {
			newIDs = append(newIDs, existingID)
		}
	}

	if len(newIDs) == 0 {
		return bucket.Delete(key)
	}

	newData, err := r.encodeIDList(newIDs, valueType)
	if err != nil {
		return err
	}

	return bucket.Put(key, newData)
}

func (r *ManyToManyRepository) encodeIDList(ids []interface{}, idType IDType) ([]byte, error) {
	switch idType {
	case IntID:
		var intIDs []int
		for _, id := range ids {
			intIDs = append(intIDs, id.(int))
		}
		return json.Marshal(intIDs)
	case StringID:
		var strIDs []string
		for _, id := range ids {
			strIDs = append(strIDs, id.(string))
		}
		return json.Marshal(strIDs)
	default:
		return nil, fmt.Errorf("unsupported ID type: %v", idType)
	}
}

func (r *ManyToManyRepository) decodeIDList(data []byte, idType IDType, result *[]interface{}) error {
	switch idType {
	case IntID:
		var intIDs []int
		if err := json.Unmarshal(data, &intIDs); err != nil {
			return err
		}
		for _, id := range intIDs {
			*result = append(*result, id)
		}
	case StringID:
		var strIDs []string
		if err := json.Unmarshal(data, &strIDs); err != nil {
			return err
		}
		for _, id := range strIDs {
			*result = append(*result, id)
		}
	default:
		return fmt.Errorf("unsupported ID type: %v", idType)
	}
	return nil
}

func (r *ManyToManyRepository) idsEqual(id1, id2 interface{}) bool {
	switch id1.(type) {
	case int:
		return id1.(int) == id2.(int)
	case string:
		return id1.(string) == id2.(string)
	default:
		return false
	}
}