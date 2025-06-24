package app

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
	INPROGRESS
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
	db_ *gorm.DB
}

func NewDB() (*DB, error) {
	db, err := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(ToDoSchema{})

	return &DB{db_: db}, nil
}

func (db *DB) Close() {
	squeel, err := db.db_.DB()
	if err != nil {
		writeError(err, "could not close db")
	}
	squeel.Close()
}

func writeError(err error, str string) {
	errFile, _ := os.OpenFile("error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer errFile.Close()
	errFile.WriteString("For Input: " + str + "\n")
	errFile.WriteString(err.Error() + "\n")
}

func (db *DB) CreateTodo(todo Todo) {
	td := ToDoSchema{Title: todo.Title_, Description: todo.Description_, Status: PENDING}
	if result := db.db_.Create(&td); result.Error != nil {
		writeError(result.Error, todo.Title_)
	}
}

func (db *DB) UpdateStatus(todo Todo) {
	var td ToDoSchema
	db.db_.Where("title = ?", todo.Title_).First(&td)
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

func (db *DB) Delete(todo Todo) {
	db.db_.Where("title = ?", todo.Title_).Delete(&ToDoSchema{})
}

func (db *DB) DeleteQueue(todo Todo, sleeper Sleeper) {
	sleeper.Sleep()
	db.db_.Where("title = ?", todo.Title_).Delete(&ToDoSchema{})
}

func (db *DB) UpdateTodo(todos []Todo) {
	db.Flush()
	for _, todo := range todos {
		db.CreateTodo(todo)
	}
}

func (db *DB) ReadTodo() []Todo {
	ltd := make([]Todo, 0)
	todos := make([]ToDoSchema, 0)
	db.db_.Find(&todos)

	for _, td := range todos {
		ltd = append(ltd, Todo{Title_: td.Title, Description_: td.Description, Status_: td.Status})
	}

	return ltd
}

func (db *DB) Flush() {
	db.db_.Migrator().DropTable(&ToDoSchema{})
	db.db_.AutoMigrate(ToDoSchema{})
}
