package friends

import (
	"github.com/abdiUNO/featherr/api/auth"
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

	response.Json(w, map[string]interface{}{
		"friendships": friendship,
	})
}
