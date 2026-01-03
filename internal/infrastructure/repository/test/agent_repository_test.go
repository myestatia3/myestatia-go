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

func setupAgentSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestAgentRepository_Create_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupAgentSQLMock(t)
	repo := repository.NewAgentRepository(db)
	agent := &entity.Agent{
		ID:    "a1",
		Email: "agent1@test.com",
		Name:  "Agent One",
	}

	mock.ExpectBegin()
	// GORM order for Agent: Name, Email, Phone, Password, Role, CompanyID, CreatedAt, UpdatedAt, DeletedAt, ID
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"agents\"")).
		WithArgs(
			agent.Name,       // $1
			agent.Email,      // $2
			sqlmock.AnyArg(), // $3 Phone
			sqlmock.AnyArg(), // $4 Password
			sqlmock.AnyArg(), // $5 Role (GORM default tag might use 'agent')
			sqlmock.AnyArg(), // $6 CompanyID
			sqlmock.AnyArg(), // $7 CreatedAt
			sqlmock.AnyArg(), // $8 UpdatedAt
			sqlmock.AnyArg(), // $9 DeletedAt
			agent.ID,         // $10
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(agent.ID))
	mock.ExpectCommit()

	// WHEN
	err := repo.Create(agent)

	// THEN
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentRepository_FindByEmail_SQLMock(t *testing.T) {
	// GIVEN
	db, mock := setupAgentSQLMock(t)
	repo := repository.NewAgentRepository(db)
	email := "search@test.com"

	rows := sqlmock.NewRows([]string{"id", "email", "deleted_at"}).
		AddRow("a2", email, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"agents\" WHERE email = $1 AND \"agents\".\"deleted_at\" IS NULL ORDER BY \"agents\".\"id\" LIMIT $2")).
		WithArgs(email, 1).
		WillReturnRows(rows)

	// WHEN
	result, err := repo.FindByEmail(context.Background(), email)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "a2", result.ID)
	assert.Equal(t, email, result.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
