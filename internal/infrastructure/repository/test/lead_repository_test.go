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

func setupLeadSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestLeadRepository_Create_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupLeadSQLMock(t)
	repo := repository.NewLeadRepository(db)
	lead := &entity.Lead{
		ID:    "L1",
		Email: "lead1@test.com",
		Name:  "Lead One",
		Phone: "123456",
	}

	mock.ExpectBegin()
	// GORM order for Lead (approximate based on struct): Name, Email, Phone, Status, PropertyID, CompanyID, AssignedAgentID, LastInteraction, Language, Source, Budget, Zone, PropertyType, Channel, SuggestedPropertiesCount, CreatedAt, UpdatedAt, DeletedAt, ID
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"leads\"")).
		WithArgs(
			lead.Name,        // $1
			lead.Email,       // $2
			lead.Phone,       // $3
			sqlmock.AnyArg(), // $4 Status
			sqlmock.AnyArg(), // $5 PropertyID
			sqlmock.AnyArg(), // $6 CompanyID
			sqlmock.AnyArg(), // $7 AssignedAgentID
			sqlmock.AnyArg(), // $8 LastInteraction
			sqlmock.AnyArg(), // $9 Language
			sqlmock.AnyArg(), // $10 Source
			sqlmock.AnyArg(), // $11 Budget
			sqlmock.AnyArg(), // $12 Zone
			sqlmock.AnyArg(), // $13 PropertyType
			sqlmock.AnyArg(), // $14 Channel
			sqlmock.AnyArg(), // $15 SuggestedPropertiesCount
			sqlmock.AnyArg(), // $16 CreatedAt
			sqlmock.AnyArg(), // $17 UpdatedAt
			sqlmock.AnyArg(), // $18 DeletedAt
			lead.ID,          // $19
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(lead.ID))
	mock.ExpectCommit()

	// WHEN
	err := repo.Create(lead)

	// THEN
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLeadRepository_FindByEmail_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupLeadSQLMock(t)
	repo := repository.NewLeadRepository(db)
	email := "search@lead.com"

	rows := sqlmock.NewRows([]string{"id", "email", "deleted_at"}).
		AddRow("L2", email, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"leads\" WHERE email = $1 AND \"leads\".\"deleted_at\" IS NULL ORDER BY \"leads\".\"id\" LIMIT $2")).
		WithArgs(email, 1).
		WillReturnRows(rows)

	// WHEN
	result, err := repo.FindByEmail(context.Background(), email)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "L2", result.ID)
	assert.Equal(t, email, result.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
