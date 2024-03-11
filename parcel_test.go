package main

import (
	"database/sql"
	"log/slog"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type SQLiteSuite struct {
	suite.Suite
	db         *sql.DB
	driverName string
	dbName     string
	store      ParcelStore
	logger     *slog.Logger
}

func (s *SQLiteSuite) SetupSuite() {
	s.driverName = "sqlite"
	s.dbName = "tracker.db"
	db, err := sql.Open(s.driverName, s.dbName)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return
	}
	s.db = db
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s.store = NewParcelStore(db, s.logger)
}

func (s *SQLiteSuite) TearDownSuite() {
	s.db.Close()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(SQLiteSuite))
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func (s *SQLiteSuite) TestAddGetDelete() {
	parcel := getTestParcel()
	s.T().Run("TestAddGetDelete", func(t *testing.T) {
		// add
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := s.store.Add(parcel)
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// get
		// получите только что добавленную посылку, убедитесь в отсутствии ошибки
		// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
		p, err := s.store.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, p.Number)
		assert.Equal(t, parcel.CreatedAt, p.CreatedAt)
		assert.Equal(t, parcel.Status, p.Status)
		assert.Equal(t, parcel.Address, p.Address)
		assert.Equal(t, parcel.Client, p.Client)

		// delete
		// удалите добавленную посылку, убедитесь в отсутствии ошибки
		// проверьте, что посылку больше нельзя получить из БД
		err = s.store.Delete(id)
		require.NoError(t, err)

		_, err = s.store.Get(id)
		require.Error(t, err)
		assert.Error(t, err, sql.ErrNoRows)
	})
}

// TestSetAddress проверяет обновление адреса
func (s *SQLiteSuite) TestSetAddress() {
	parcel := getTestParcel()
	s.T().Run("TestSetAddress", func(t *testing.T) {
		// add
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := s.store.Add(parcel)
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// set address
		// обновите адрес, убедитесь в отсутствии ошибки
		newAddress := "new test address"
		err = s.store.SetAddress(id, newAddress)
		require.NoError(t, err)

		// check
		// получите добавленную посылку и убедитесь, что адрес обновился
		p, err := s.store.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, p.Number)
		assert.Equal(t, parcel.CreatedAt, p.CreatedAt)
		assert.Equal(t, parcel.Status, p.Status)
		assert.Equal(t, newAddress, p.Address)
		assert.Equal(t, parcel.Client, p.Client)
	})
}

// TestSetStatus проверяет обновление статуса
func (s *SQLiteSuite) TestSetStatus() {
	parcel := getTestParcel()
	s.T().Run("TestSetAddress", func(t *testing.T) {
		// add
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := s.store.Add(parcel)
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// set status
		// обновите статус, убедитесь в отсутствии ошибки
		newStatus := ParcelStatusSent
		err = s.store.SetStatus(id, newStatus)
		require.NoError(t, err)

		// check
		// получите добавленную посылку и убедитесь, что статус обновился
		p, err := s.store.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, p.Number)
		assert.Equal(t, parcel.CreatedAt, p.CreatedAt)
		assert.Equal(t, newStatus, p.Status)
		assert.Equal(t, parcel.Address, p.Address)
		assert.Equal(t, parcel.Client, p.Client)
	})
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func (s *SQLiteSuite) TestGetByClient() {
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
	s.T().Run("TestSetAddress", func(t *testing.T) {
		// add
		for i := 0; i < len(parcels); i++ {
			// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
			id, err := s.store.Add(parcels[i])
			require.NoError(t, err)

			// обновляем идентификатор добавленной у посылки
			parcels[i].Number = id

			// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
			parcelMap[id] = parcels[i]
		}

		// get by client
		// получите список посылок по идентификатору клиента, сохранённого в переменной client
		// убедитесь в отсутствии ошибки
		// убедитесь, что количество полученных посылок совпадает с количеством добавленных
		storedParcels, err := s.store.GetByClient(client)
		require.NoError(t, err)
		require.Equal(t, len(storedParcels), len(parcels))

		// check
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		for _, parcel := range storedParcels {
			p, ok := parcelMap[parcel.Number]
			require.True(t, ok)
			assert.Equal(t, parcel.CreatedAt, p.CreatedAt)
			assert.Equal(t, parcel.Status, p.Status)
			assert.Equal(t, parcel.Address, p.Address)
			assert.Equal(t, parcel.Client, p.Client)
		}
	})
}
