package test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupPropertySQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestPropertyRepository_Create_Simplified(t *testing.T) {
	// GIVEN
	db, mock := setupPropertySQLMock(t)
	repo := repository.NewPropertyRepository(db)
	prop := &entity.Property{
		ID:        "P1",
		Reference: "REF1",
	}

	mock.ExpectBegin()
	// GORM with Postgres: uses RETURNING "id" and many fields.
	// We use WillReturnRows and AnyArg for simplicity.
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"properties\"")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(prop.ID))
	mock.ExpectCommit()

	// WHEN
	err := repo.Create(prop)

	// THEN
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPropertyRepository_FindByReference_Simplified(t *testing.T) {
	// GIVEN
	db, mock := setupPropertySQLMock(t)
	repo := repository.NewPropertyRepository(db)
	ref := "REF1"

	rows := sqlmock.NewRows([]string{"id", "reference", "deleted_at"}).
		AddRow("P1", ref, nil)

	// GORM adds ORDER BY id and LIMIT 1 for First()
	mock.ExpectQuery("SELECT .* FROM \"properties\" WHERE reference = \\$1").
		WithArgs(ref, 1). // 1 for the LIMIT
		WillReturnRows(rows)

	// WHEN
	result, err := repo.FindByReference(context.Background(), ref)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "P1", result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
