package v1

import (
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/common"
	"net/http"
)

func listToken(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	list, err := handler.TokenBasket.GetAll()
	if err != nil {
		http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	data := common.TokenListResponse{
		Tokens: list,
	}
	handler.ServeJSON(response, data)
}

func deleteToken(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	token := mux.Vars(request)["token"]
	ok, err := handler.TokenBasket.Delete(token)
	if err != nil {
		http.Error(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(response, "Not found", http.StatusNotFound)
	}
	handler.HttpClose(response, "", http.StatusOK)
}
