package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/abdiUNO/featherr/utils"

	"github.com/abdiUNO/featherr/config"
	"github.com/abdiUNO/featherr/database/orm"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

/*
JWT claims struct
*/

type Token struct {
	UserId string
	jwt.StandardClaims
}

func GetDB() *gorm.DB {
	return orm.DBCon
}

var cfg = config.GetConfig()

//a struct to rep user
type User struct {
	orm.GormModel
	FullName      string  `json:"fullname"`
	Username      string  `json:"username"`
	Email         string  `json:"email"`
	Password      string  `json:",omitempty"`
	Friends       []*User `gorm:"many2many:friendships;association_jointable_foreignkey:friend_id" json:"-"`
	Image         string  `json:"image"`
	FcmToken      string  `json:"fcmToken"`
	JwtToken      string  `sql:"-" json:"jwtToken"`
	EmailVerified bool    `sql:"not null;DEFAULT:false" json:"emailVerified"`
}

func (user *User) TableName() string {
	return "users"
}

func (user *User) Validate() *utils.Error {
	if !strings.Contains(user.Email, "@") {
		return utils.NewError(utils.EINVALID, "email address is required", nil)

	}

	if len(user.Password) < 6 {
		return utils.NewError(utils.EINVALID, "password is required", nil)
	}

	temp := &User{}

	err := GetDB().Table("users").Where("email = ?", user.Email).Or("username = ?", user.Username).First(temp).Error

	//fmt.Println(temp == nil)

	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.NewError(utils.EINVALID, "connection error. Please retry", err)
	}

	if temp.Username != "" && temp.Username == user.Username {
		return utils.NewError(utils.EINVALID, "username is already in use by another user", nil)
	}

	if temp.Email != "" && temp.Email == user.Email {
		return utils.NewError(utils.EINVALID, "email address already in use by another user", nil)
	}

	return nil
}

func (user *User) Create() (*User, *utils.Error) {
	hashedPassword := hashAndSalt([]byte(user.Password))

	user.Password = string(hashedPassword)

	err := GetDB().Create(user).Error
	if err != nil {
		return &User{}, utils.NewError(utils.ECONFLICT, "could not create user", nil)
	}

	//Create new JWT token for the newly registered account
	tk := &Token{UserId: user.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
	user.JwtToken = tokenString //Store the token in the response

	user.Password = "" //delete password

	return user, nil
}

func ValidateUserInfo(id string, user *User) *utils.Error {
	if !strings.Contains(user.Email, "@") {
		return utils.NewError(utils.EINVALID, "email address is required", nil)

	}

	temp := &User{}
	err := GetDB().Table("users").Where("id = ?", id).First(temp).Error

	if temp.Username != "" && temp.Username != user.Username {
		tempUsername := &User{}
		tempErr := GetDB().Table("users").Where("username = ?", user.Username).First(tempUsername).Error
		if tempErr != gorm.ErrRecordNotFound && tempUsername.Username == user.Username {
			return utils.NewError(utils.EINVALID, "username is already in use by another user", nil)
		}
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.NewError(utils.EINVALID, "connection error. Please retry", err)
	}

	if user.Email != "" && temp.Email != user.Email {
		tempUserEmail := &User{}
		tempErr := GetDB().Table("users").Where("email = ?", user.Email).First(tempUserEmail).Error
		fmt.Println(tempUserEmail.Email)
		if tempErr != gorm.ErrRecordNotFound && tempUserEmail.Email == user.Email {
			return utils.NewError(utils.EINVALID, "email address already in use by another user", nil)
		}
	}

	return nil
}

func Update(id string, user *User) (*User, *utils.Error) {

	if validateErr := ValidateUserInfo(id, user); validateErr != nil {
		return &User{}, validateErr
	}

	updateUser := GetUser(id)

	//Create JWT token
	tk := &Token{UserId: updateUser.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
	updateUser.JwtToken = tokenString //Store the token in the response

	err := GetDB().Model(&updateUser).Updates(&user).Error
	if err != nil {
		return &User{}, utils.NewError(utils.ECONFLICT, "could not create user", nil)
	}

	return updateUser, nil
}

func (user *User) UpdatePassword(oldPassword string, newPassword string) *utils.Error {

	newPasswordHashed := hashAndSalt([]byte(newPassword))

	fmt.Println(newPassword)
	fmt.Println(user.Password)

	if comparePasswords(user.Password, []byte(oldPassword)) == false { //Password does not match!
		return utils.NewError(utils.EINVALID, "Incorrect current password", nil)
	}

	if len(newPassword) < 6 {
		return utils.NewError(utils.EINVALID, "Password is required", nil)
	}

	user.Password = string(newPasswordHashed)

	//Create JWT token
	tk := &Token{UserId: user.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
	user.JwtToken = tokenString //Store the token in the response

	err := GetDB().Save(&user).Updates(&user).Error
	if err != nil {
		return utils.NewError(utils.ECONFLICT, "could not update password", nil)
	}

	return nil
}

func Login(email string, password string, fcmToken string) (*User, *utils.Error) {

	user := &User{}
	err := GetDB().Table("users").Where("email = ?", email).First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &User{}, utils.NewError(utils.ENOTFOUND, "Your email or password is incorrect.", nil)
		}

		return &User{}, utils.NewError(utils.EINTERNAL, "internal server error", nil)
	}

	if comparePasswords(user.Password, []byte(password)) == false { //Password does not match!
		return &User{}, utils.NewError(utils.EINVALID, "Your email or password is incorrect.", nil)
	}

	GetDB().Model(&user).Update("fcm_token", fcmToken)

	//Worked! Logged In
	user.Password = ""

	//Create JWT token
	tk := &Token{UserId: user.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
	user.JwtToken = tokenString //Store the token in the response

	return user, nil
}

func QueryUsers(query string) (*[]User, *utils.Error) {
	users := &[]User{}

	err := GetDB().Table("users").Where("full_name LIKE ?", query+"%").Find(&users).Error

	if err != nil {
		return &[]User{}, utils.NewError(utils.EINVALID, "invalid login credentials. Please try again", err)
	}

	return users, nil
}

func GetUser(u string) *User {

	user := &User{}
	GetDB().Table("users").Where("id = ?", u).First(user)
	if user.Email == "" { //User not found!
		return nil
	}

	return user
}

func FindUserById(u string) (*User, error) {

	user := &User{}
	err := GetDB().Table("users").Where("id = ?", u).First(user).Error

	if user.Email == "" { //User not found!
		return nil, err
	}

	return user, err
}

func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
