package v1

import (
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/koomgr/apis/proto"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"net/http"
)

func ListToken(this *AdminV1Handler, usr tokenapi.UserDesc, response http.ResponseWriter, request *http.Request) {
	list, err := this.TokenBasket.GetAll()
	if err != nil {
		this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	data := proto.TokenListResponse{
		Tokens: list,
	}
	this.ServeJSON(response, data)
}

func DeleteToken(this *AdminV1Handler, usr tokenapi.UserDesc, response http.ResponseWriter, request *http.Request) {
	token := mux.Vars(request)["token"]
	ok, err := this.TokenBasket.Delete(token)
	if err != nil {
		this.HttpError(response, "Server error. Check server logs", http.StatusInternalServerError)
		return
	}
	if !ok {
		this.HttpError(response, "Not found", http.StatusNotFound)
		return
	}
	this.HttpClose(response, "", http.StatusOK)
}
