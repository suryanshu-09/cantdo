package tests

import (
	"testing"
	"time"

	"github.com/suryanshu-09/cantdo/todos"
)

func TestDB(t *testing.T) {
	t.Run("create and read", func(t *testing.T) {
		db, err := todos.NewDB()
		defer db.Flush()
		defer db.Close()
		if err != nil {
			t.Errorf("db creation failed:%s", err.Error())
		}

		todos := []todos.ListTodo{{Title: "Gae Man tries Golang", Description: "Let's see if he's any good.", Status: todos.PENDING}, {Title: "Trying harder he is", Description: "Won't make it he.", Status: todos.PENDING}}

		db.CreateTodo(todos[0])
		db.CreateTodo(todos[1])

		read := db.ReadTodo()

		for i, todo := range read {
			title := todos[i].Title == todo.Title
			desc := todos[i].Description == todo.Description
			status := todos[i].Status == todo.Status
			if !title || !desc || !status {
				t.Errorf("got:%v\nwant:%v", todo, todos[i])
			}
		}
	})

	t.Run("update status and delete", func(t *testing.T) {
		db, err := todos.NewDB()
		defer db.Flush()
		defer db.Close()
		if err != nil {
			t.Errorf("db creation failed:%s", err.Error())
		}

		todo := []todos.ListTodo{{Title: "Gae Man tries Golang", Description: "Let's see if he's any good.", Status: todos.PENDING}, {Title: "Trying harder he is", Description: "Won't make it he.", Status: todos.PENDING}}

		db.CreateTodo(todo[0])
		db.CreateTodo(todo[1])

		db.UpdateStatus(todo[0])
		read := db.ReadTodo()
		if read[0].Status != todos.COMPLETE {
			t.Errorf("did not update status %d", read[0].Status)
		}
		spySleeper := &SpySleeper{}
		db.DeleteQueue(todo[1], spySleeper)
		if spySleeper.Calls != 1 {
			t.Errorf("sleeping didn't workout")
		}
		time.Sleep(1 * time.Second)
		read = db.ReadTodo()
		if len(read) != 0 {
			t.Errorf("did not delete, %d", len(read))
		}
	})
}

type Sleeper interface {
	Sleep()
}
type SpySleeper struct {
	Calls int
}

func (s *SpySleeper) Sleep() {
	s.Calls++
}
