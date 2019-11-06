package server

import (
	"encoding/json"
	"github.com/abdiUNO/featherr/api/chats"
	"github.com/abdiUNO/featherr/api/cliques"
	"github.com/abdiUNO/featherr/api/friends"
	"net/http"

	"github.com/abdiUNO/featherr/api/auth"
)

func (s *Server) SetupRoutes() {
	s.router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	s.router.HandleFunc("/api/users/new", auth.CreateUser).Methods("POST")
	s.router.HandleFunc("/api/users/login", auth.Authenticate).Methods("POST")
	s.router.HandleFunc("/api/users", auth.FindUsers).Queries("query", "{query}").Methods("GET")
	s.router.HandleFunc("/api/users/{id}", auth.UpdateUser).Methods("PATCH")
	s.router.HandleFunc("/api/users/{id}/change-password", auth.ChangePassword).Methods("PATCH")

	s.router.HandleFunc("/api/users/{id}/otp-code", auth.GenerateOTP).Methods("GET")
	s.router.HandleFunc("/api/users/{id}/otp-code", auth.ValidateOTP).Queries("code", "{code}").Methods("POST")

	s.router.HandleFunc("/api/friends", friends.GetFriends).Methods("GET")
	s.router.HandleFunc("/api/users/{id}/add", friends.AddFriend).Methods("PUT")

	s.router.HandleFunc("/api/friends/{id}/conversations", chats.CreateConversation).Methods("POST")
	s.router.HandleFunc("/api/conversations", chats.GetConversations).Methods("GET")
	s.router.HandleFunc("/api/conversations/{id}", chats.RemoveConversation).Methods("DELETE")

	s.router.HandleFunc("/api/chat/", cliques.GetGroups).Methods("GET")
	s.router.HandleFunc("/api/chat/new", cliques.CreateGroup).Methods("POST")
	s.router.HandleFunc("/api/chat/find", cliques.JoinGroup).Methods("POST")
	s.router.HandleFunc("/api/chat/{id}/leave", cliques.LeaveGroup).Methods("PUT")
	s.router.HandleFunc("/api/upload", auth.UploadProfileImage).Methods("POST")
}
