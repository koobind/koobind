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

func getGroup(handler *AdminV1Handler, namespace string, groupName string) (crdGroup *v1alpha1.Group, err error) {
	crdGroup = &v1alpha1.Group{}
	err = handler.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      groupName,
	}, crdGroup)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if err != nil {
		return nil, nil
	}
	return crdGroup, nil
}

// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{}'
// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/groups/grp2 -d '{ "description": "Group2", "disabled": true }'

// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{ "description": "Group1", "disabled": true }'
// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/groups/grp2 -d '{}'

// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{ "disabled": false}'
// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/groups/grp2 -d '{ "description": "Group2", "disabled": true }'

func AddApplyPatchGroup(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	// Decode the payload
	var groupSpec v1alpha1.GroupSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&groupSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
		return
	}
	// Check if group exists
	groupName := mux.Vars(request)["group"]
	crdGroup, err := getGroup(handler, namespace, groupName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdGroup != nil {
		// It exists.
		if request.Method == "POST" {
			handler.HttpError(response, fmt.Sprintf("Group '%s' already exists!", groupName), http.StatusConflict)
		} else {
			if request.Method == "PUT" {
				crdGroup.Spec = groupSpec
			} else if request.Method == "PATCH" {
				// We  overwrite only the provided fields
				if groupSpec.Description != "" {
					crdGroup.Spec.Description = groupSpec.Description
				}
				if groupSpec.Disabled != nil {
					crdGroup.Spec.Disabled = groupSpec.Disabled
				}
			}
			err = handler.KubeClient.Update(context.TODO(), crdGroup)
			if err != nil {
				handler.HttpError(response, err.Error(), http.StatusInternalServerError)
				return
			}
			handler.HttpClose(response, "", http.StatusOK)
		}
	} else {
		if request.Method == "PATCH" {
			handler.HttpError(response, fmt.Sprintf("Group '%s' does not exists!", groupName), http.StatusNotFound)
		} else {
			// Group does not exists. Must create it
			crdGroup = &v1alpha1.Group{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupName,
					Namespace: namespace,
				},
				Spec:   groupSpec,
				Status: v1alpha1.GroupStatus{},
			}
			err = handler.KubeClient.Create(context.TODO(), crdGroup)
			if err != nil {
				handler.HttpError(response, err.Error(), http.StatusInternalServerError)
				return
			}
			handler.HttpClose(response, "", http.StatusCreated)
		}
	}
}

// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/groups/grp1
// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/groups/grp2

func DeleteGroup(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if group exists
	groupName := mux.Vars(request)["group"]
	crdGroup, err := getGroup(handler, namespace, groupName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdGroup == nil {
		handler.HttpError(response, fmt.Sprintf("Group '%s' does not exists!", groupName), http.StatusNotFound)
		return
	}
	err = handler.KubeClient.Delete(context.TODO(), crdGroup, client.GracePeriodSeconds(0))
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
}
