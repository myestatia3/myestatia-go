package test

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMessageSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestMessageRepository_Create_Simplified(t *testing.T) {
	// GIVEN
	db, mock := setupMessageSQLMock(t)
	repo := repository.NewMessageRepository(db)
	msg := &entity.Message{ID: "M1", Content: "Test content"}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"messages\"")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(msg.ID))
	mock.ExpectCommit()

	// WHEN
	err := repo.Create(msg)

	// THEN
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_FindByLeadID_Simplified(t *testing.T) {
	// GIVEN
	db, mock := setupMessageSQLMock(t)
	repo := repository.NewMessageRepository(db)
	leadID := "L1"

	rows := sqlmock.NewRows([]string{"id", "content", "lead_id"}).
		AddRow("M1", "Hello", leadID)

	// GORM adds ORDER BY timestamp asc
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"messages\" WHERE lead_id = $1")).
		WithArgs(leadID).
		WillReturnRows(rows)

	// WHEN
	result, err := repo.FindByLeadID(leadID)

	// THEN
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Hello", result[0].Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}
