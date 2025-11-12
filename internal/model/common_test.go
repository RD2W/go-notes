package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeFieldsInitialization(t *testing.T) {
	tf := &TimeFields{}

	// Проверяем, что временные метки изначально равны нулю
	assert.Zero(t, tf.GetCreatedAt())
	assert.Zero(t, tf.GetUpdatedAt())

	// Инициализируем временные метки
	tf.initializeTimestamps()

	// Проверяем, что временные метки теперь не равны нулю
	assert.NotZero(t, tf.GetCreatedAt())
	assert.NotZero(t, tf.GetUpdatedAt())

	// Проверяем, что время создания и обновления примерно равны
	assert.WithinDuration(t, tf.GetCreatedAt(), tf.GetUpdatedAt(), time.Second)
}

func TestTimeFieldsUpdateTimestamp(t *testing.T) {
	tf := &TimeFields{}
	tf.initializeTimestamps()

	originalUpdatedAt := tf.GetUpdatedAt()

	// Ждем немного времени
	time.Sleep(10 * time.Millisecond)

	// Обновляем временную метку
	tf.updateTimestamp()

	// Проверяем, что время обновления изменилось
	assert.True(t, tf.GetUpdatedAt().After(originalUpdatedAt))

	// Проверяем, что время создания осталось прежним
	assert.WithinDuration(t, tf.GetCreatedAt(), originalUpdatedAt, time.Second)
}

func TestTimeFieldsGetters(t *testing.T) {
	tf := &TimeFields{}
	tf.initializeTimestamps()

	createdAt := tf.GetCreatedAt()
	updatedAt := tf.GetUpdatedAt()

	assert.NotZero(t, createdAt)
	assert.NotZero(t, updatedAt)
	assert.WithinDuration(t, createdAt, updatedAt, time.Second)
}

func TestTimeFieldsMultipleUpdates(t *testing.T) {
	tf := &TimeFields{}
	tf.initializeTimestamps()

	// Сохраняем начальные времена
	initialCreatedAt := tf.GetCreatedAt()
	initialUpdatedAt := tf.GetUpdatedAt()

	// Обновляем несколько раз
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Millisecond)
		tf.updateTimestamp()

		// Проверяем, что время создания не изменилось
		assert.Equal(t, initialCreatedAt, tf.GetCreatedAt())

		// Проверяем, что время обновления изменилось
		assert.True(t, tf.GetUpdatedAt().After(initialUpdatedAt))

		initialUpdatedAt = tf.GetUpdatedAt()
	}
}

func TestTimeFieldsMarshalJSON(t *testing.T) {
	tf := &TimeFields{}
	tf.initializeTimestamps()

	jsonData, err := json.Marshal(tf)
	assert.NoError(t, err)

	// Проверяем, что JSON содержит ожидаемые поля
	var timeMap map[string]interface{}
	err = json.Unmarshal(jsonData, &timeMap)
	assert.NoError(t, err)

	assert.Contains(t, timeMap, "created_at")
	assert.Contains(t, timeMap, "updated_at")

	// Проверяем, что значения соответствуют ожидаемым
	createdAtStr, ok := timeMap["created_at"].(string)
	assert.True(t, ok)

	updatedAtStr, ok := timeMap["updated_at"].(string)
	assert.True(t, ok)

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	assert.NoError(t, err)

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	assert.NoError(t, err)

	assert.WithinDuration(t, tf.GetCreatedAt(), createdAt, time.Second)
	assert.WithinDuration(t, tf.GetUpdatedAt(), updatedAt, time.Second)
}

func TestTimeFieldsUnmarshalJSON(t *testing.T) {
	// Создаем JSON с временными метками
	createdAt := time.Now().Add(-1 * time.Hour)
	updatedAt := time.Now()

	jsonStr := `{"created_at":"` + createdAt.Format(time.RFC3339) + `","updated_at":"` + updatedAt.Format(time.RFC3339) + `"}`

	tf := &TimeFields{}
	err := tf.UnmarshalJSON([]byte(jsonStr))
	assert.NoError(t, err)

	assert.WithinDuration(t, createdAt, tf.GetCreatedAt(), time.Second)
	assert.WithinDuration(t, updatedAt, tf.GetUpdatedAt(), time.Second)
}

func TestTimeFieldsMarshalUnmarshalRoundTrip(t *testing.T) {
	// Создаем TimeFields инициализируем
	tf1 := &TimeFields{}
	tf1.initializeTimestamps()

	// Сохраняем времена до сериализации
	createdAtBefore := tf1.GetCreatedAt()
	updatedAtBefore := tf1.GetUpdatedAt()

	// Маршалим в JSON
	jsonData, err := json.Marshal(tf1)
	assert.NoError(t, err)

	// Создаем новый объект и десериализуем
	tf2 := &TimeFields{}
	err = tf2.UnmarshalJSON(jsonData)
	assert.NoError(t, err)

	// Проверяем, что значения сохранились
	assert.WithinDuration(t, createdAtBefore, tf2.GetCreatedAt(), time.Second)
	assert.WithinDuration(t, updatedAtBefore, tf2.GetUpdatedAt(), time.Second)

	// Проверяем, что можно снова обновить временную метку
	time.Sleep(10 * time.Millisecond)
	tf2.updateTimestamp()

	// Время обновления должно быть больше, чем до десериализации
	assert.True(t, tf2.GetUpdatedAt().After(updatedAtBefore))

	// Время создания должно остаться тем же
	assert.WithinDuration(t, createdAtBefore, tf2.GetCreatedAt(), time.Second)
}

func TestTimeFieldsUnmarshalInvalidJSON(t *testing.T) {
	// Проверяем десериализацию с неправильным JSON
	invalidJSON := `{"created_at":"invalid_time","updated_at":"invalid_time"}`

	tf := &TimeFields{}
	err := tf.UnmarshalJSON([]byte(invalidJSON))
	assert.Error(t, err)

	// Проверяем десериализацию с неполным JSON
	incompleteJSON := `{"created_at":"2023-01-01T00:00:00Z"}`

	tf2 := &TimeFields{}
	err = tf2.UnmarshalJSON([]byte(incompleteJSON))
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC), tf2.GetCreatedAt())
	// updatedAt будет равен нулю, так как не был задан в JSON
	assert.Zero(t, tf2.GetUpdatedAt())
}
