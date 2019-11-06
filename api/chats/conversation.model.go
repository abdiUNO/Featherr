package chats

import (
	"fmt"
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/database/orm"
	"github.com/abdiUNO/featherr/utils"
	"github.com/jinzhu/gorm"
	"time"
)

//type Room struct {
//	ID        string `database:"primary_key;type:varchar(255);" json:"id"`
//	IsPrivate bool   `json:"isPrivate"`
//
//	CreatedAt time.Time `json:"-"`
//	UpdatedAt time.Time `json:"-"`
//}

type Conversation struct {
	orm.GormModel
	Members []*auth.User `gorm:"many2many:conversation_users;save_association:false" json:"members"`
}

type ConversationUser struct {
	ConversationId string        `json:"conversationId"`
	Conversation   *Conversation `json:"conversation"`
	UserId         string        `json:"userId"`
	User           *auth.User    `json:"user"`
	DeletedAt      *time.Time    `sql:"index"`
}

func GetDB() *gorm.DB {
	return orm.DBCon
}

func (convo *Conversation) Validate(user *auth.User, friend *auth.User) *utils.Error {
	var groupIds []string
	var idStr string
	groups := &[]Conversation{}

	if err := GetDB().Table("conversation_users").Where("user_id = ? AND deleted_at IS NULL", friend.ID).Pluck("conversation_id", &groupIds).Error; err != nil {
		return utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	fmt.Println("one")

	for i, id := range groupIds {
		if i == 0 {
			idStr += "'" + id + "'"
		} else {
			idStr += ",'" + id + "'"
		}
	}

	if len(groupIds) == 0 {
		return nil
	}

	if err := GetDB().Table("conversation_users").Where("conversation_id IN ("+idStr+") AND user_id = ?", user.ID).Find(&groups).Error; err != nil {
		fmt.Println(err)
		if err != gorm.ErrRecordNotFound {
			return utils.NewError(utils.EINTERNAL, "internal database error", err)
		}
	}

	fmt.Println("two")

	if len(*groups) > 0 {
		return utils.NewError(utils.ECONFLICT, "conversation already added", nil)
	}

	return nil
}

func (convo *Conversation) Create(user *auth.User, friend *auth.User) (*Conversation, *utils.Error) {
	convo.Members = append(convo.Members, user)
	convo.Members = append(convo.Members, friend)

	if ok := convo.Validate(user, friend); ok != nil {
		return &Conversation{}, utils.NewError(utils.ECONFLICT, "conversation already added", ok)
	}

	err := GetDB().Create(&convo).Error
	if err != nil {
		return &Conversation{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return convo, nil
}

func AllConversations(user *auth.User) (*[]Conversation, *utils.Error) {
	var groupIds []string
	var idStr string
	groups := &[]Conversation{}

	if err := GetDB().Table("conversation_users").Where("user_id = ? AND deleted_at IS NULL", user.ID).Pluck("conversation_id", &groupIds).Error; err != nil {
		return groups, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	if len(groupIds) <= 0 {
		return groups, nil
	}

	for i, id := range groupIds {
		if i == 0 {
			idStr += "'" + id + "'"
		} else {
			idStr += ",'" + id + "'"
		}
	}

	if err := GetDB().Table("conversations").Preload("Members", "id != ? ", user.ID).Where("id IN (" + idStr + ")").Find(&groups).Error; err != nil {
		return groups, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return groups, nil
}

func FindConversationById(groupId *string) (*Conversation, *utils.Error) {
	group := &Conversation{}
	err := GetDB().Table("conversations").Preload("Members").Where("id = ?", groupId).First(group).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &Conversation{}, utils.NewError(utils.ENOTFOUND, "Conversation not found", nil)
		} else {
			return &Conversation{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}
	}

	return group, nil
}

func DeleteConversation(group *Conversation) *utils.Error {
	groupId := group.ID

	if err := GetDB().Exec(`
		UPDATE conversation_users
		SET deleted_at = NOW()
		WHERE conversation_users.conversation_id = ?
	`, groupId).Error; err != nil {
		return utils.NewError(utils.EINTERNAL, err.Error(), err)
	}

	if ok := GetDB().Unscoped().Table("conversations").Where("id = ?", groupId).Delete(&group).Error; ok != nil {
		return utils.NewError(utils.EINTERNAL, "internal database error", ok)
	}

	return nil
}
