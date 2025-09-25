package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"your-app/relations"
	
	_ "github.com/lib/pq"
)

package models

import (
	"time"
	"github.com/mymindmap/api/pkg/core/relations"
)

// теги для парсинга полей
type User struct {
	relations.BaseModel
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"password"` // Не отправляем пароль в JSON
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	MindMaps  []*MindMap `json:"mind_maps,omitempty"`
	Posts  []*Post       `json:"posts,omitempty"`
}

func (u *User) GetID() interface{} {
    return u.ID
}

func (u *User) GetTableName() string { return "users" }

func (u *User) GetRelations() map[string]relations.Relation {
	return map[string]relations.Relation{
		"mindMaps": relations.NewHasMany(
			func() relations.Model { return &MindMap{} },
			relations.RelationConfig{
				ForeignKey: "user_id",
				LocalKey:   "id",
			},
		),
		"posts": relations.NewHasMany(
			func() relations.Model { return &Post{} },
			relations.RelationConfig{
				ForeignKey: "user_id",
				LocalKey:   "id",
			},
		),
	}
}

func main() {
	var err error
	db, err = sql.Open("postgres", "user=postgres dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	ctx := context.Background()
	
	// Создание отношений
	user := &User{}
	post := &Post{}
	
	builder := relations.NewQueryBuilder(user)
	
	// Регистрация отношений
	builder.RegisterRelation("posts", relations.NewHasMany(
		func() relations.Model { return &Post{} },
		relations.RelationConfig{},
	))
	
	// Загрузка пользователя с постами
	if err := builder.With("posts").Get(ctx, 1); err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("User: %s\n", user.Name)
	fmt.Printf("Posts count: %d\n", len(user.Posts))
	
	// Отношение "принадлежит"
	postBuilder := relations.NewQueryBuilder(post)
	postBuilder.RegisterRelation("user", relations.NewBelongsTo(
		&User{},
		relations.RelationConfig{},
	))
	
	if err := postBuilder.With("user").Get(ctx, 1); err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Post: %s, Author: %s\n", post.Title, post.User.Name)
}