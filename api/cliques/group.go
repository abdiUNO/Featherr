package cliques

import (
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/database/orm"
	"github.com/abdiUNO/featherr/utils"
	"github.com/jinzhu/gorm"
	"time"
)

func GetDB() *gorm.DB {
	return orm.DBCon
}

//type Room struct {
//	ID        string `database:"primary_key;type:varchar(255);" json:"id"`
//	IsPrivate bool   `json:"isPrivate"`
//
//	CreatedAt time.Time `json:"-"`
//	UpdatedAt time.Time `json:"-"`
//}

type Group struct {
	orm.GormModel
	Admin   *auth.User   `json:"-" json:"admin"`
	Members []*auth.User `gorm:"many2many:group_users;save_association:false" json:"members"`
	Count   int          `json:"count"`
}

type GroupUser struct {
	GroupId   string     `json:"groupId"`
	Group     *Group     `json:"group"`
	UserId    string     `json:"userId"`
	User      *auth.User `json:"user"`
	Saved     bool       `sql:"not null;DEFAULT:false" json:"saved"`
	DeletedAt *time.Time `sql:"index"`
}

func AllGroups(user *auth.User) (*[]Group, *utils.Error) {
	var groupIds []string
	var idStr string
	groups := &[]Group{}

	if err := GetDB().Table("group_users").Where("user_id = ? AND deleted_at IS NULL", user.ID).Pluck("group_id", &groupIds).Error; err != nil {
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

	if err := GetDB().Table("groups").Preload("Members", "group_users.deleted_at IS NULL").Where("id IN (" + idStr + ")").Find(&groups).Error; err != nil {
		return groups, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return groups, nil
}

func (group *Group) Create(user *auth.User) (*Group, *utils.Error) {
	group.Admin = user
	group.Members = append(group.Members, user)
	group.Count += 1
	err := GetDB().Create(&group).Error

	if err != nil {
		return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return group, nil
}

func (group *Group) AddUser(user *auth.User) (*Group, *utils.Error) {
	group.Members = append(group.Members, user)
	err := GetDB().Save(&group).Error

	if err != nil {
		return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return group, nil
}

func DeleteGroupAssoc(group *Group, user *auth.User) *utils.Error {
	groupId := group.ID
	userId := user.ID

	err := GetDB().Exec(`
	UPDATE group_users
	SET deleted_at = NOW()
	WHERE group_users.group_id = ?
	AND group_users.user_id = ?
	`, groupId, userId).Error

	if err != nil {
		return utils.NewError(utils.EINTERNAL, err.Error(), err)
	}

	GetDB().Model(&group).Update("count", group.Count-1)

	return nil
}

func FindGroupById(groupId *string) (*Group, *utils.Error) {
	group := &Group{}
	err := GetDB().Table("groups").Preload("Members").Where("id = ?", groupId).First(group).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &Group{}, utils.NewError(utils.ENOTFOUND, "Group not found", nil)
		} else {
			return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}
	}

	return group, nil
}

func FindGroup(user *auth.User) (*Group, *utils.Error) {
	var groupIds []string
	var idStr string
	var query = ""

	group := &Group{}

	if err := GetDB().Table("group_users").Where("user_id = ?", user.ID).Pluck("group_id", &groupIds).Error; err != nil {
		return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	for i, id := range groupIds {
		if i == 0 {
			idStr += "'" + id + "'"
		} else {
			idStr += ",'" + id + "'"
		}
	}

	if len(groupIds) <= 0 {
		query = "count < 3"
	} else {
		query = "count < 3 AND id NOT IN (" + idStr + ")"
	}

	if err := GetDB().Table("groups").Preload("Members").Where(query).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return group.Create(user)
		} else {
			return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}
	}

	group.Members = append(group.Members, user)
	group.Count += 1

	if ok := GetDB().Save(&group).Error; ok != nil {
		return &Group{}, utils.NewError(utils.EINTERNAL, "internal database error", ok)
	}

	return group, nil
}
