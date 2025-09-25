package models

import (
	"time"
	"github.com/mymindmap/api/pkg/core/db/relations"
)

// теги для парсинга полей
type User struct {
	relations.BaseModel
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"password"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	MindMaps  []*MindMap `json:"mind_maps,omitempty"`
	Posts     []*Post       `json:"posts,omitempty"`
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