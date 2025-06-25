package tests

import (
	"testing"

	"github.com/suryanshu-09/cantdo/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestDB() (*app.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(app.ToDoSchema{})
	return &app.DB{DB_: db}, nil
}

func TestDB(t *testing.T) {
	db, err := NewTestDB()
	if err != nil {
		t.Error(err.Error())
	}
	t.Run("add to db", func(t *testing.T) {
		todo := app.Todo{
			Title_:       "This is a completely new todo",
			Description_: "This is its description",
		}
		db.CreateTodo(todo)
		created := db.GetTodo(todo)
		if created.Title_ != todo.Title_ {
			t.Errorf("got:%v\nwant:%v", created, todo)
		}
	})
	t.Run("UpdateStatus", func(t *testing.T) {
		todo := app.Todo{
			Title_:       "This is a completely new todo",
			Description_: "This is its description",
		}
		db.UpdateStatus(todo)
		updated := db.GetTodo(todo)
		if updated.Status_ != app.INPROGRESS {
			t.Errorf("got:%v\nwant:%v", updated, todo)
		}
	})

	t.Run("UpdateDescription", func(t *testing.T) {
		todo := app.Todo{
			Title_:       "This is a completely new todo",
			Description_: "This is its completely new description",
		}
		db.UpdateDescription(todo)
		updated := db.GetTodo(todo)
		if updated.Description_ != todo.Description_ {
			t.Errorf("got:%v\nwant:%v", updated, todo)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		todo := app.Todo{
			Title_:       "This is a completely new todo",
			Description_: "This is its description",
		}
		db.Delete(todo)
		updated := db.GetTodo(todo)
		if updated.Title_ != "" {
			t.Errorf("got:%v\nwant:%v", updated, todo)
		}
	})
	db.Flush()
	db.Close()
}
