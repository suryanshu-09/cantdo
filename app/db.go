package app

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Status int

const (
	PENDING Status = iota
	INPROGRESS
	COMPLETE
)

type Todo struct {
	Title_       string
	Description_ string
	Status_      Status
}

type ToDoSchema struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey;autoincrement"`
	Title       string `gorm:"unique"`
	Description string
	Status      Status
}

type DB struct {
	DB_ *gorm.DB
}

func NewDB() (*DB, error) {
	homeDir, _ := os.UserHomeDir()
	dbPath := homeDir + "/.local/share/cantdo/cantdo.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(ToDoSchema{})
	return &DB{DB_: db}, nil
}

func (db *DB) Close() {
	squeel, err := db.DB_.DB()
	if err != nil {
		writeError(err, "could not close db")
	}
	squeel.Close()
}

func writeError(err error, str string) {
	homeDir, _ := os.UserHomeDir()
	errPath := homeDir + "/.local/share/cantdo/error.txt"
	errFile, _ := os.OpenFile(errPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer errFile.Close()
	errFile.WriteString("For Input: " + str + "\n")
	errFile.WriteString(err.Error() + "\n")
}

func (db *DB) CreateTodo(todo Todo) {
	td := ToDoSchema{Title: todo.Title_, Description: todo.Description_, Status: PENDING}
	if result := db.DB_.Create(&td); result.Error != nil {
		writeError(result.Error, todo.Title_)
	}
}

func (db *DB) UpdateStatus(todo Todo) {
	var td ToDoSchema
	result := db.DB_.Where("title = ?", todo.Title_).First(&td)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			writeError(result.Error, todo.Title_)
			return
		}
	}
	td.Status = (td.Status + 1) % 3
	db.DB_.Save(&td)
}

func (db *DB) Delete(todo Todo) {
	db.DB_.Where("title = ?", todo.Title_).Delete(&ToDoSchema{})
}

func (db *DB) UpdateDescription(todo Todo) {
	var td ToDoSchema
	found := db.DB_.Where("title = ?", todo.Title_).First(&td)
	if found.Error != nil {
		if found.Error == gorm.ErrRecordNotFound {
			db.CreateTodo(todo)
		} else {
			writeError(found.Error, todo.Title_)
		}
	} else {
		td.Description = todo.Description_
		db.DB_.Save(&td)
	}
}

func (db *DB) GetTodo(todo Todo) Todo {
	found := ToDoSchema{}
	f := db.DB_.Where("title = ?", todo.Title_).First(&found)
	if f.Error != nil {
		if f.Error == gorm.ErrRecordNotFound {
			return Todo{}
		}
		writeError(f.Error, todo.Title_)
		return Todo{}
	}
	return Todo{Title_: found.Title, Description_: found.Description, Status_: found.Status}
}

func (db *DB) ReadTodo() []Todo {
	ltd := make([]Todo, 0)
	todos := make([]ToDoSchema, 0)
	db.DB_.Find(&todos)

	for _, td := range todos {
		ltd = append(ltd, Todo{Title_: td.Title, Description_: td.Description, Status_: td.Status})
	}

	return ltd
}

func (db *DB) Flush() {
	db.Close()
	db.DB_.Migrator().DropTable(&ToDoSchema{})
	db.DB_.AutoMigrate(ToDoSchema{})
}
