package friends

import (
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/database/orm"
	"github.com/abdiUNO/featherr/utils"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"time"
)

type Friendship struct {
	ID        string     `database:"primary_key;type:varchar(255);" json:"id"`
	User      *auth.User `json:"-"`
	UserID    string     `json:"-"`
	FriendId  string     `json:"friendId"`
	Friend    *auth.User `json:"friend";gorm:"association_foreignkey:id;foreignkey:friend_id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
}

func (model *Friendship) BeforeCreate(scope *gorm.Scope) error {
	u1 := uuid.Must(uuid.NewV4(), nil)
	scope.SetColumn("ID", u1.String())
	return nil
}

func GetDB() *gorm.DB {
	return orm.DBCon
}

func FindFriendshipById(id *string) (*Friendship, *utils.Error) {
	friendship := &Friendship{}
	err := GetDB().Table("friendships").Preload("User").Preload("Friend").Where("id = ?", id).First(friendship).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &Friendship{}, utils.NewError(utils.ENOTFOUND, "Conversation not found", nil)
		} else {
			return &Friendship{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}
	}

	return friendship, nil
}

func (friendship *Friendship) Create(user *auth.User, friend *auth.User) (*Friendship, *utils.Error) {
	var userId = user.ID
	var friendId = friend.ID

	err := GetDB().Where("friend_id = ? AND  user_id = ?", userId, friendId).Or("user_id = ? AND  friend_id = ?", userId, friendId).Find(&friendship).Error

	if err != gorm.ErrRecordNotFound {
		return &Friendship{}, utils.NewError(utils.ECONFLICT, "friend already added", err)
	}

	friendship.User = user
	friendship.UserID = userId
	friendship.Friend = friend
	friendship.FriendId = friendId

	err = GetDB().Save(&friendship).Error

	if err != nil {
		return &Friendship{}, utils.NewError(utils.EINVALID, "could not add friend", err)
	}

	return friendship, nil
}

func FindFriends(user *auth.User) ([]*auth.User, *utils.Error) {
	var userId = user.ID
	var friends []*auth.User

	//err := GetDB().Table("friendships").Select("*").Joins("JOIN users ON users.id = friend_id").Where("user_id = ?", user.ID).Scan(&friends).Error
	err := GetDB().Raw(`SELECT * from friendships 
							JOIN users ON users.id = user_id 
							WHERE friend_id = ? 
							UNION 
							SELECT * from friendships 
							JOIN users ON users.id = friend_id
							WHERE user_id = ?`, userId, userId).Scan(&friends).Error

	if err != nil {
		return []*auth.User{}, utils.NewError(utils.EINTERNAL, "internal database error", nil)
	}

	return friends, nil
}

//func (friendship *Friendship) AddFriend(user *users.UserModel, friendId string) (*Friendship, *utils.Error) {
//	err := GetDB().Where("friend_id = ? AND  user_id = ?", user.ID, friendId).Or("user_id = ? AND  friend_id = ?", user.ID, friendId).Find(&friendship).Error
//
//	if err != gorm.ErrRecordNotFound {
//		return &Friendship{}, utils.NewError(utils.EINVALID, "friend already added", nil)
//	} else if err != nil {
//		return &Friendship{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
//	}
//
//	friend := users.GetUser(friendId)
//
//	if err := GetDB().Model(&user).Association("Friends").Append(friend).Error; err != nil {
//		return &Friendship{}, utils.NewError(utils.EINTERNAL, "could not add friend", err)
//	}
//
//	return friendship, nil
//}
