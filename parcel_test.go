package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Импортируйте драйвер modernc.org/sqlite
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// createTable создает таблицу parcel, если она не существует
func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS parcel (number INTEGER PRIMARY KEY AUTOINCREMENT,client INTEGER NOT NULL,status INTEGER NOT NULL,address TEXT NOT NULL,created_at TEXT NOT NULL);`
	_, err := db.Exec(query)
	return err
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "test_tracker.db")
	require.NoError(t, err, "failed to open test db")
	defer db.Close()

	err = createTable(db) // Создаем таблицу
	require.NoError(t, err, "failed to create table")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	require.Greater(t, id, 0, "invalid parcel id")

	parcel.Number = id // <--- ADD THIS LINE: Update the parcel's Number with the generated ID

	// get
	storedParcel, err := store.Get(id)
	require.NoError(t, err, "failed to get parcel")
	assert.Equal(t, parcel.Number, storedParcel.Number, "number mismatch") //ADDED ASSERT

	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	require.Equal(t, parcel.Client, storedParcel.Client, "client mismatch")
	require.Equal(t, parcel.Status, storedParcel.Status, "status mismatch")
	require.Equal(t, parcel.Address, storedParcel.Address, "address mismatch")
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt, "created_at mismatch")
	// delete
	err = store.Delete(id)
	require.NoError(t, err, "failed to delete parcel")
	_, err = store.Get(id)
	require.Error(t, err, "parcel should not exist")
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "test_tracker.db") // настройте подключение к БД
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	err = createTable(db) // Создаем таблицу
	require.NoError(t, err, "failed to create table")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	require.NoError(t, err, "failed to add parcel")
	require.Greater(t, id, 0, "invalid parcel id")
	// set address
	newAddress := "new test address"
	// обновите адрес, убедитесь в отсутствии ошибки
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "failed to set address")

	// check
	storedParcel, err := store.Get(id)
	// получите добавленную посылку и убедитесь, что адрес обновился
	require.NoError(t, err, "failed to get parcel")
	assert.Equal(t, newAddress, storedParcel.Address, "address should be updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "test_tracker.db") // настройте подключение к БД
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	err = createTable(db) // Создаем таблицу
	require.NoError(t, err, "failed to create table")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	require.NoError(t, err, "failed to add parcel")
	require.Greater(t, id, 0, "invalid parcel id")

	// set status
	newStatus := ParcelStatusSent // обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "failed to set status")

	// check
	storedParcel, err := store.Get(id) // получите добавленную посылку и убедитесь, что статус обновился
	require.NoError(t, err, "failed to get parcel")
	assert.Equal(t, newStatus, storedParcel.Status, "status should be updated")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "test_tracker.db") // настройте подключение к БД
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	err = createTable(db) // Создаем таблицу
	require.NoError(t, err, "failed to create table")

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "failed to add parcel")
		require.Greater(t, id, 0, "invalid parcel id")
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err, "failed to get parcels by client")
	assert.Len(t, storedParcels, len(parcels), "incorrect number of parcels")
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		originalParcel, ok := parcelMap[parcel.Number]
		assert.True(t, ok, "parcel not found in map")

		assert.Equal(t, originalParcel, parcel, "parcel mismatch")
		assert.Equal(t, originalParcel.Status, parcel.Status, "status mismatch")
		assert.Equal(t, originalParcel.Address, parcel.Address, "address mismatch")
		assert.Equal(t, originalParcel.CreatedAt, parcel.CreatedAt, "created_at mismatch")
	}
}
