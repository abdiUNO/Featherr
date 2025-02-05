package chats

import (
	"fmt"
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/api/friends"
	"github.com/abdiUNO/featherr/utils"
	"github.com/abdiUNO/featherr/utils/response"
	"github.com/gorilla/mux"
	"net/http"
)

var CreateConversation = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	token := r.Context().Value("token").(*auth.Token)
	friendshipId := params["id"]
	friendship, _ := friends.FindFriendshipById(&friendshipId)

	fmt.Println(friendship)

	group := &Conversation{}

	group, err := group.Create(friendship.User, friendship.Friend)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	var userId = friendship.FriendId

	if friendship.FriendId == token.UserId {
		userId = friendship.UserID
	}

	go utils.SendToToic(&utils.MessageData{
		MsgType: utils.UPDATE_CONVERSATIONS,
		Topic:   userId,
		UserId:  token.UserId,
	})

	response.Json(w, map[string]interface{}{
		"group": group,
	})
}

var GetConversations = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)

	groups, err := AllConversations(user)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"groups": groups,
	})
}

var RemoveConversation = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupId := params["id"]
	group, err := FindConversationById(&groupId)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	if ok := DeleteConversation(group); ok != nil {
		response.HandleError(w, ok)
		return
	}

	token := r.Context().Value("token").(*auth.Token)
	go utils.SendToToic(&utils.MessageData{
		MsgType: utils.UPDATE_CONVERSATIONS,
		Topic:   group.ID,
		UserId:  token.UserId,
	})

	response.Json(w, map[string]interface{}{
		"groupId": group.ID,
	})
}
