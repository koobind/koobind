package v1

import (
	"context"
	"encoding/json"
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

func AddGroup(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
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
	if crdGroup != nil {
		handler.HttpError(response, "Group already exists!", http.StatusConflict)
		return
	}
	// Ok, now, we can create it
	var groupSpec v1alpha1.GroupSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&groupSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
		return
	}
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

// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{ "description": "Group1", "disabled": true }'
// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/groups/grp2 -d '{}'

func ApplyGroup(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
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
	if crdGroup == nil {
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
	} else {
		// It exists. We fully overwrite the group definition.
		crdGroup.Spec = groupSpec
		err = handler.KubeClient.Update(context.TODO(), crdGroup)
		if err != nil {
			handler.HttpError(response, err.Error(), http.StatusInternalServerError)
			return
		}
		handler.HttpClose(response, "", http.StatusOK)
	}
}

// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{ "disabled": false}'
// curl -k -i -u admin:admin -X PATCH https://koomgrdev:9444/auth/v1/admin/_/groups/grp2 -d '{ "description": "Group2", "disabled": true }'

func PatchGroup(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
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
		handler.HttpError(response, "Group does not exists!", http.StatusNotFound)
		return
	}
	// Parse the provided group definition
	var groupSpec v1alpha1.GroupSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&groupSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
		return
	}
	// We  overwrite only the provided fields
	if groupSpec.Description != "" {
		crdGroup.Spec.Description = groupSpec.Description
	}
	if groupSpec.Disabled != nil {
		crdGroup.Spec.Disabled = groupSpec.Disabled
	}
	err = handler.KubeClient.Update(context.TODO(), crdGroup)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
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
		handler.HttpError(response, "Group does not exists!", http.StatusNotFound)
		return
	}
	err = handler.KubeClient.Delete(context.TODO(), crdGroup, client.GracePeriodSeconds(0))
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
}
