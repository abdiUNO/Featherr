package database

import (
	"fmt"
	"github.com/abdiUNO/featherr/api/chats"
	"github.com/abdiUNO/featherr/api/cliques"
	"github.com/abdiUNO/featherr/api/friends"

	"github.com/abdiUNO/featherr/database/orm"

	"github.com/abdiUNO/featherr/api/auth"

	"github.com/abdiUNO/featherr/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func InitDatabase() {
	cfg := config.GetConfig()

	username := cfg.DBUser
	password := cfg.DBPass
	dbName := cfg.DBName
	dbHost := cfg.DBHost
	dbPort := cfg.DBPort
	dbType := cfg.DBType

	dbUri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", username, password, dbHost, dbPort, dbName)
	fmt.Println(dbUri)

	conn, err := gorm.Open(dbType, dbUri)
	if err != nil {
		fmt.Print(err)
	}

	orm.DBCon = conn

	orm.DBCon.Set("database:table_options", "ENGINE=InnoDB")
	orm.DBCon.Set("database:table_options", "collation_connection=utf8_general_ci")

	orm.DBCon.Debug().AutoMigrate(&auth.User{}, &friends.Friendship{}, &cliques.Group{}, &cliques.GroupUser{}, &chats.Conversation{}, &chats.ConversationUser{})
	orm.DBCon.LogMode(false)

}
