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

// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/users/jsmith2 -d '{}'
// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/users/jsmith3 -d '{ "email": "xx@xx" }'
// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/users/jsmith4 -d '{ "email": "xx@xx", "commonName": "John smith4", "passwordHash": "$2a$10$SxKQu8Ny54c/MuujiltVD.9J9P8kvSM01UK.sTh/bhAxYhoLGwjLi", "uid": 10009, "comment": "A test User", "disabled": false }'

// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/users/jsmith2 -d '{}'
// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/users/jsmith3 -d '{ "email": "xx@xx" }'
// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/users/jsmith4 -d '{ "email": "xx@xx", "commonName": "John smith4", "passwordHash": "$2a$10$SxKQu8Ny54c/MuujiltVD.9J9P8kvSM01UK.sTh/bhAxYhoLGwjLi", "uid": 10009, "comment": "A test User", "disabled": false }'

func AddApplyPatchUser(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	// Decode the payload
	var userSpec v1alpha1.UserSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&userSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
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
		// It exists.
		if request.Method == "POST" {
			handler.HttpError(response, fmt.Sprintf("User '%s' already exists!", userName), http.StatusConflict)
		} else {
			if request.Method == "PUT" {
				crdUser.Spec = userSpec
			} else if request.Method == "PATCH" {
				// We  overwrite only the provided fields
				if userSpec.Uid != nil {
					crdUser.Spec.Uid = userSpec.Uid
				}
				if userSpec.PasswordHash != "" {
					crdUser.Spec.PasswordHash = userSpec.PasswordHash
				}
				if userSpec.Comment != "" {
					crdUser.Spec.Comment = userSpec.Comment
				}
				if userSpec.Email != "" {
					crdUser.Spec.Email = userSpec.Email
				}
				if userSpec.CommonName != "" {
					crdUser.Spec.CommonName = userSpec.CommonName
				}
				if userSpec.Disabled != nil {
					crdUser.Spec.Disabled = userSpec.Disabled
				}
			}
			err = handler.KubeClient.Update(context.TODO(), crdUser)
			if err != nil {
				handler.HttpError(response, err.Error(), http.StatusInternalServerError)
				return
			}
			handler.HttpClose(response, "", http.StatusOK)
		}
	} else {
		if request.Method == "PATCH" {
			handler.HttpError(response, fmt.Sprintf("User '%s' does not exists!", userName), http.StatusNotFound)
		} else {
			// User does not exists. Must create it
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
	}
}

// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/users/jsmith2
// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/users/jsmith3
// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/users/jsmith4

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
		handler.HttpError(response, fmt.Sprintf("User '%s' does not exists!", userName), http.StatusNotFound)
		return
	}
	err = handler.KubeClient.Delete(context.TODO(), crdUser, client.GracePeriodSeconds(0))
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
}
