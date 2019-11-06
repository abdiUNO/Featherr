package response

import (
	"encoding/json"
	"github.com/abdiUNO/featherr/utils"
	"net/http"
	"strconv"
)

func HandleError(w http.ResponseWriter, err *utils.Error) {
	status, _ := strconv.Atoi(err.Code)
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"data":  nil,
		"error": err,
	}

	w.Header().Add("Content-Type", "application/json")

	if _err := json.NewEncoder(w).Encode(resp); _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	}
}

func Json(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusOK)

	resp := data
	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}
