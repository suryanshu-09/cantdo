package todos

import (
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Status int

const (
	PENDING Status = iota
	COMPLETE
)

type ToDo struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey;autoincrement"`
	Title       string `gorm:"unique"`
	Description string
	Status      Status
}

type DB struct {
	db_ *gorm.DB
}

func NewDB() (*DB, error) {
	db, err := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(ToDo{})

	return &DB{db_: db}, nil
}

func (db *DB) CreateTodo(todo ListTodo) {
	td := ToDo{Title: todo.Title, Description: todo.Description, Status: PENDING}
	if result := db.db_.Create(&td); result.Error != nil {
		errFile, _ := os.OpenFile("error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer errFile.Close()
		errFile.WriteString("For Input: " + todo.Title + "\n")
		errFile.WriteString(result.Error.Error() + "\n")
	}
}

func (db *DB) UpdateStatus(todo ListTodo) {
	var td ToDo
	db.db_.Where("title = ?", todo.Title).First(&td)
	td.Status = (td.Status + 1) % 2
	if td.Status == COMPLETE {
		go db.DeleteQueue(todo, &DefaultSleeper{})
	}
	db.db_.Save(&td)
}

type Sleeper interface {
	Sleep()
}
type DefaultSleeper struct{}

func (d *DefaultSleeper) Sleep() {
	time.Sleep(1 * time.Hour)
}

func (db *DB) DeleteQueue(todo ListTodo, sleeper Sleeper) {
	sleeper.Sleep()
	db.db_.Where("title = ?", todo.Title).Delete(&ToDo{})
}

func (db *DB) UpdateTodo(todos []ListTodo) {
	db.Flush()
	for _, todo := range todos {
		db.CreateTodo(todo)
	}
}

func (db *DB) ReadTodo() []ListTodo {
	ltd := make([]ListTodo, 0)
	todos := make([]ToDo, 0)
	db.db_.Find(&todos)

	for _, td := range todos {
		ltd = append(ltd, ListTodo{Title: td.Title, Description: td.Description, Status: td.Status})
	}

	return ltd
}

func (db *DB) Flush() {
	db.db_.Migrator().DropTable(&ToDo{})
	db.db_.AutoMigrate(ToDo{})
}
