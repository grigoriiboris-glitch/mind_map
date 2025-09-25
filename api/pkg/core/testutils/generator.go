package testutils

import (
	"reflect"
	"strconv"
	"time"
)

// FieldGenerator содержит настройки для генерации тестовых данных
type FieldGenerator struct {
	// CustomGenerators позволяет переопределить генерацию для конкретных полей
	CustomGenerators map[string]func(fieldName, dbTag string, id int) interface{}
}

// NewFieldGenerator создает новый генератор с настройками по умолчанию
func NewFieldGenerator() *FieldGenerator {
	return &FieldGenerator{
		CustomGenerators: make(map[string]func(fieldName, dbTag string, id int) interface{}),
	}
}

// GenerateTestData создает тестовые данные на основе структуры модели
func (g *FieldGenerator) GenerateTestData(modelType reflect.Type, id int) map[string]interface{} {
	testData := make(map[string]interface{})
	
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Проверяем кастомный генератор
		if generator, exists := g.CustomGenerators[dbTag]; exists {
			testData[dbTag] = generator(field.Name, dbTag, id)
			continue
		}

		fieldType := field.Type
		fieldName := field.Name

		testData[dbTag] = g.generateValue(fieldName, dbTag, fieldType, id)
	}

	return testData
}

// generateValue генерирует значение на основе типа поля
func (g *FieldGenerator) generateValue(fieldName, dbTag string, fieldType reflect.Type, id int) interface{} {
	switch fieldType.Kind() {
	case reflect.String:
		return g.generateStringValue(fieldName, dbTag, id)
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.generateIntValue(fieldName, dbTag, id)
		
	case reflect.Bool:
		return id%2 == 0
		
	case reflect.Float32, reflect.Float64:
		return float64(id) * 1.5
		
	default:
		if fieldType == reflect.TypeOf(time.Time{}) {
			return g.generateTimeValue(fieldName, dbTag, id)
		}
	}
	
	return nil
}

// generateStringValue генерирует строковые значения
func (g *FieldGenerator) generateStringValue(fieldName, dbTag string, id int) string {
	if fieldName == "ID" {
		return "" // ID должен быть int
	}
	
	switch {
	case contains([]string{"title", "name", "subject"}, dbTag):
		return "Test " + fieldName + " " + strconv.Itoa(id)
	case contains([]string{"content", "body", "description"}, dbTag):
		return "Test content for " + dbTag + " " + strconv.Itoa(id)
	case contains([]string{"email", "mail"}, dbTag):
		return "test" + strconv.Itoa(id) + "@example.com"
	case contains([]string{"phone", "telephone"}, dbTag):
		return "+123456789" + strconv.Itoa(id)
	case contains([]string{"url", "website", "link"}, dbTag):
		return "https://example.com/" + dbTag + "/" + strconv.Itoa(id)
	default:
		return "test_" + dbTag + "_" + strconv.Itoa(id)
	}
}

// generateIntValue генерирует целочисленные значения
func (g *FieldGenerator) generateIntValue(fieldName, dbTag string, id int) int {
	if fieldName == "ID" {
		return id
	}
	
	switch {
	case contains([]string{"user_id", "author_id", "creator_id"}, dbTag):
		return id * 10
	case contains([]string{"count", "quantity", "amount"}, dbTag):
		return id * 5
	case contains([]string{"status", "type", "category"}, dbTag):
		return id
	default:
		return id * 100
	}
}

// generateTimeValue генерирует временные значения
func (g *FieldGenerator) generateTimeValue(fieldName, dbTag string, id int) time.Time {
	switch {
	case contains([]string{"created_at", "created"}, dbTag):
		return time.Now().Add(-time.Duration(id) * time.Hour * 24)
	case contains([]string{"updated_at", "modified_at", "updated"}, dbTag):
		return time.Now().Add(-time.Duration(id) * time.Hour)
	case contains([]string{"deleted_at", "expires_at", "due_date"}, dbTag):
		return time.Now().Add(time.Duration(id) * time.Hour * 24)
	default:
		return time.Now().Add(-time.Duration(id) * time.Hour)
	}
}

// contains проверяет наличие строки в slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CreateTestEntity создает тестовую сущность с заполненными полями
func CreateTestEntity(model interface{}, id int) interface{} {
	return CreateTestEntityWithGenerator(model, id, NewFieldGenerator())
}

// CreateTestEntityWithGenerator создает тестовую сущность с использованием кастомного генератора
func CreateTestEntityWithGenerator(model interface{}, id int, generator *FieldGenerator) interface{} {
	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()
	
	testData := generator.GenerateTestData(modelType, id)
	
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		if value, exists := testData[dbTag]; exists {
			fieldValue := modelValue.Field(i)
			if fieldValue.CanSet() {
				rv := reflect.ValueOf(value)
				if rv.Type().ConvertibleTo(fieldValue.Type()) {
					fieldValue.Set(rv.Convert(fieldValue.Type()))
				}
			}
		}
	}

	return model
}