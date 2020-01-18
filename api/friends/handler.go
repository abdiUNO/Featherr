package friends

import (
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/utils"
	"github.com/abdiUNO/featherr/utils/response"
	"github.com/gorilla/mux"
	"net/http"
)

var GetFriends = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)

	friends, err := FindFriends(user)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"friends": friends,
	})
}

var BlockUser = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	params := mux.Vars(r)
	friendshipId := params["id"]

	friendShip, err := FindFriendshipById(&friendshipId)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	blocked := &auth.Blocked{}
	if friendShip.User.ID == token.UserId {
		blocked, err = blocked.BlockUser(friendShip.User, friendShip.Friend)
	} else {
		blocked, err = blocked.BlockUser(friendShip.Friend, friendShip.User)
	}

	if err != nil {
		response.HandleError(w, err)
		return
	}

	err = DeleteFriendShip(&friendshipId)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"blocked": blocked,
	})
}

var AddFriend = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)
	params := mux.Vars(r)
	friendId := params["id"]
	friend := auth.GetUser(friendId)

	friendship := &Friendship{}

	friendship, err := friendship.Create(user, friend)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	go utils.SendToToic(&utils.MessageData{
		MsgType: utils.UPDATE_FRIENDS,
		Topic:   friendId,
		UserId:  token.UserId,
	})

	response.Json(w, map[string]interface{}{
		"friendships": friendship,
	})
}
