package cliques

import (
	"github.com/abdiUNO/featherr/api/auth"
	"github.com/abdiUNO/featherr/utils"
	"github.com/abdiUNO/featherr/utils/response"
	"github.com/gorilla/mux"
	"net/http"
)

var CreateGroup = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)

	group := &Group{}

	group, err := group.Create(user)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"group": group,
	})
}

var GetGroups = func(w http.ResponseWriter, r *http.Request) {

	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)

	groups, err := AllGroups(user)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Json(w, map[string]interface{}{
		"groups": groups,
	})
}

var JoinGroup = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)

	group, err := FindGroup(user)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	go utils.SendToToic(&utils.MessageData{
		MsgType: utils.UPDATE_CLIQUES,
		Topic:   group.ID,
		UserId:  token.UserId,
	})

	response.Json(w, map[string]interface{}{
		"group": group,
	})
}

var LeaveGroup = func(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token").(*auth.Token)
	user := auth.GetUser(token.UserId)
	params := mux.Vars(r)
	groupId := params["id"]
	group, err := FindGroupById(&groupId)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	if ok := DeleteGroupAssoc(group, user); ok != nil {
		response.HandleError(w, ok)
		return
	}

	go utils.SendToToic(&utils.MessageData{
		MsgType: utils.UPDATE_CLIQUES,
		Topic:   group.ID,
		UserId:  token.UserId,
	})

	response.Json(w, map[string]interface{}{
		"groupId": group.ID,
	})
}
