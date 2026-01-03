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

func setupSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestCompanyRepository_Create_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupSQLMock(t)
	repo := repository.NewCompanyRepository(db)
	company := &entity.Company{
		ID:     "1",
		Name:   "Test SQLMock Company",
		City:   "Test City",
		Email1: "test@company.com",
	}

	mock.ExpectBegin()
	// GORM with Postgres: Name, Address, PostalCode, City, Province, Country, OfficeLocation, ContactPerson, Email1, Email2, Phone1, Phone2, Website, PageLink, WebDeveloper, CreatedAt, UpdatedAt, DeletedAt, ID
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"companies\"")).
		WithArgs(
			company.Name,     // $1
			sqlmock.AnyArg(), // $2 Address
			sqlmock.AnyArg(), // $3 PostalCode
			company.City,     // $4
			sqlmock.AnyArg(), // $5 Province
			sqlmock.AnyArg(), // $6 Country
			sqlmock.AnyArg(), // $7 OfficeLocation
			sqlmock.AnyArg(), // $8 ContactPerson
			company.Email1,   // $9
			sqlmock.AnyArg(), // $10 Email2
			sqlmock.AnyArg(), // $11 Phone1
			sqlmock.AnyArg(), // $12 Phone2
			sqlmock.AnyArg(), // $13 Website
			sqlmock.AnyArg(), // $14 PageLink
			sqlmock.AnyArg(), // $15 WebDeveloper
			sqlmock.AnyArg(), // $16 CreatedAt
			sqlmock.AnyArg(), // $17 UpdatedAt
			sqlmock.AnyArg(), // $18 DeletedAt
			company.ID,       // $19 (ID at the end)
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(company.ID))
	mock.ExpectCommit()

	// WHEN
	err := repo.Create(company)

	// THEN
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_FindByName_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupSQLMock(t)
	repo := repository.NewCompanyRepository(db)
	companyName := "Search Company"

	rows := sqlmock.NewRows([]string{"id", "name", "deleted_at"}).
		AddRow("1", companyName, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"companies\" WHERE name = $1 AND \"companies\".\"deleted_at\" IS NULL ORDER BY \"companies\".\"id\" LIMIT $2")).
		WithArgs(companyName, 1).
		WillReturnRows(rows)

	// WHEN
	result, err := repo.FindByName(context.Background(), companyName)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1", result.ID)
	assert.Equal(t, companyName, result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
