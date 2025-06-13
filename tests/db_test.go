package tests

import (
	"testing"

	"github.com/suryanshu-09/pomodoro/todos"
)

func TestDB(t *testing.T) {
	t.Run("create and read", func(t *testing.T) {
		db, err := todos.NewDB()
		defer db.Flush()
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
}
