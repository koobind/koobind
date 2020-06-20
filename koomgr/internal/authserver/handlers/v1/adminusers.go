package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/koobind/koobind/common"
	"github.com/koobind/koobind/koomgr/apis/directory/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
  Some REST function are documented by example, assuming:
  - admin:admin is a valid user:password, with user belonging the the 'kooadmin' group
  - The server is koomgrdev:9444
*/

// curl -k  -u admin:admin -X GET https://koomgrdev:9444/auth/v1/admin/users/jsmith | jq
// curl -k -i -u admin:admin -X GET https://koomgrdev:9444/auth/v1/admin/users/jsmith

func DescribeUser(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	user := mux.Vars(request)["user"]
	found, userDescribeResponse := handler.Providers.DescribeUser(user)
	if !found {
		handler.HttpError(response, fmt.Sprintf("User %s not found", user), http.StatusNotFound)
		return
	}
	handler.ServeJSON(response, userDescribeResponse)
}

func getUser(handler *AdminV1Handler, namespace string, userName string) (crdUser *v1alpha1.User, err error) {
	crdUser = &v1alpha1.User{}
	err = handler.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      userName,
	}, crdUser)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if err != nil {
		return nil, nil
	}
	return crdUser, nil
}

// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/users/jsmith2 -d '{}'
// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/users/jsmith3 -d '{ "email": "xx@xx" }'
// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/users/jsmith4 -d '{ "email": "xx@xx", "commonName": "John smith4", "passwordHash": "$2a$10$SxKQu8Ny54c/MuujiltVD.9J9P8kvSM01UK.sTh/bhAxYhoLGwjLi", "uid": 10009, "comment": "A test User", "disabled": false }'

func AddUser(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if user exists
	userName := mux.Vars(request)["user"]
	crdUser, err := getUser(handler, namespace, userName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdUser != nil {
		handler.HttpError(response, "User already exists!", http.StatusConflict)
		return
	}
	// Ok, now, we can create it
	var userSpec v1alpha1.UserSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&userSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
		return
	}
	crdUser = &v1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userName,
			Namespace: namespace,
		},
		Spec:   userSpec,
		Status: v1alpha1.UserStatus{},
	}
	err = handler.KubeClient.Create(context.TODO(), crdUser)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusCreated)
}

func DeleteUser(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if user exists
	userName := mux.Vars(request)["user"]
	crdUser, err := getUser(handler, namespace, userName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdUser == nil {
		handler.HttpError(response, "User does not exists!", http.StatusNotFound)
		return
	}
	err = handler.KubeClient.Delete(context.TODO(), crdUser, client.GracePeriodSeconds(0))
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
}
