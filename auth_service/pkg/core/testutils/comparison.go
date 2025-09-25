package testutils

import (
	"fmt"
	"reflect"
)

// CompareEntities сравнивает две сущности, исключая указанные поля
func CompareEntities(a, b interface{}, excludeFields []string) error {
	aValue := reflect.ValueOf(a).Elem()
	bValue := reflect.ValueOf(b).Elem()
	aType := aValue.Type()

	for i := 0; i < aType.NumField(); i++ {
		field := aType.Field(i)
		fieldName := field.Name

		// Пропускаем исключенные поля
		if containsString(excludeFields, fieldName) {
			continue
		}

		aField := aValue.Field(i)
		bField := bValue.Field(i)

		if !reflect.DeepEqual(aField.Interface(), bField.Interface()) {
			return fmt.Errorf("field %s: got %v, want %v", fieldName, aField.Interface(), bField.Interface())
		}
	}

	return nil
}

// containsString проверяет наличие строки в slice
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}