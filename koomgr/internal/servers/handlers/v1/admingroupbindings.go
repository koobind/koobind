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

func getBinding(handler *AdminV1Handler, namespace string, bindingName string) (crdGroupBinding *v1alpha1.GroupBinding, err error) {
	crdGroupBinding = &v1alpha1.GroupBinding{}
	err = handler.KubeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      bindingName,
	}, crdGroupBinding)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if err != nil {
		return nil, nil
	}
	return crdGroupBinding, nil
}

// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/groups/grp1 -d '{}'
// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/users/user1 -d '{ "commonName": "User 1", "passwordHash": "$2a$10$zRW1QM2ZLLGwl3S9ebys3.gUOgsHyaOCdihJ590q.B58IUQsxgE9y"}'
// curl -k -i -u admin:admin -X POST https://koomgrdev:9444/auth/v1/admin/_/groupbindings/user1/grp1 -d '{}'

func AddApplyPatchGroupBinding(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	userName := mux.Vars(request)["user"]
	groupName := mux.Vars(request)["group"]
	crdUser, err := getUser(handler, namespace, userName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdUser == nil {
		handler.HttpError(response, fmt.Sprintf("User '%s' does not exists", userName), http.StatusNotFound)
		return
	}
	// Check if group exists
	crdGroup, err := getGroup(handler, namespace, groupName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdGroup == nil {
		handler.HttpError(response, fmt.Sprintf("Group '%s' does not exists", groupName), http.StatusNotFound)
		return
	}
	bindingName := fmt.Sprintf("%s-%s", userName, groupName)
	// We decode the body, but only 'disabled' flag is of interest
	var groupBindingSpec v1alpha1.GroupBindingSpec
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&groupBindingSpec)
	if err != nil {
		handler.HttpError(response, "Error while parsing body data:"+err.Error(), http.StatusBadRequest)
		return
	}
	groupBindingSpec.User = userName
	groupBindingSpec.Group = groupName
	// Check if binding already exists
	crdGroupBinding, err := getBinding(handler, namespace, bindingName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdGroupBinding != nil {
		// GroupBinding exists.
		if request.Method == "POST" {
			handler.HttpError(response, fmt.Sprintf("GroupBinding '%s' already exists!", bindingName), http.StatusConflict)
		} else {
			if request.Method == "PUT" {
				crdGroupBinding.Spec = groupBindingSpec
			} else if request.Method == "PATCH" {
				// If patch, do not alter previous value if not in payload
				if groupBindingSpec.Disabled != nil {
					crdGroupBinding.Spec.Disabled = groupBindingSpec.Disabled
				}
			}
			err = handler.KubeClient.Update(context.TODO(), crdGroupBinding)
			if err != nil {
				handler.HttpError(response, err.Error(), http.StatusInternalServerError)
				return
			}
			handler.HttpClose(response, "", http.StatusOK)
		}
	} else {
		// GroupBinding does not exists. Create it
		if request.Method == "PATCH" {
			handler.HttpError(response, fmt.Sprintf("GroupBinding '%s' does not exists!", bindingName), http.StatusNotFound)
		}
		crdGroupBinding = &v1alpha1.GroupBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: namespace,
			},
			Spec:   groupBindingSpec,
			Status: v1alpha1.GroupBindingStatus{},
		}
		err = handler.KubeClient.Create(context.TODO(), crdGroupBinding)
		if err != nil {
			handler.HttpError(response, err.Error(), http.StatusInternalServerError)
			return
		}
		handler.HttpClose(response, "", http.StatusCreated)
	}
}

// curl -k -i -u admin:admin -X DELETE https://koomgrdev:9444/auth/v1/admin/_/groupbindings/user1/grp1

func DeleteGroupBinding(handler *AdminV1Handler, usr common.User, response http.ResponseWriter, request *http.Request) {
	provider := mux.Vars(request)["provider"]
	namespace, err := handler.Providers.GetNamespace(provider)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusBadRequest)
		return
	}
	userName := mux.Vars(request)["user"]
	groupName := mux.Vars(request)["group"]
	bindingName := fmt.Sprintf("%s-%s", userName, groupName)
	// Check if binding already exists
	crdGroupBinding, err := getBinding(handler, namespace, bindingName)
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if crdGroupBinding == nil {
		handler.HttpError(response, fmt.Sprintf("GroupBinding '%s' does not exists!", bindingName), http.StatusConflict)
		return
	}
	err = handler.KubeClient.Delete(context.TODO(), crdGroupBinding, client.GracePeriodSeconds(0))
	if err != nil {
		handler.HttpError(response, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.HttpClose(response, "", http.StatusOK)
}
