package v1

import (
	"encoding/json"
	"fmt"
	"github.com/koobind/koobind/koomgr/apis/proto"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"github.com/koobind/koobind/koomgr/internal/servers/handlers"
	"net/http"
)

type ChangePasswordHandler struct {
	handlers.AuthHandler
}

// export HASH=$(kubectl koo hash --password user1); echo $HASH
// curl -k -i -u admin:admin -X PUT https://koomgrdev:9444/auth/v1/admin/_/users/user1 -d "{ \"passwordHash\": \"$HASH\" }"

// curl -k -i -u user1:user1 -X POST https://koomgrdev:9444/auth/v1/changePassword -d '{ "oldPassword": "user1", "newPassword": "user1b" }'

func (this *ChangePasswordHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.ServeAuthenticatedHTTP(response, request, func(usr tokenapi.UserDesc) {
		found, userDescription := this.Providers.DescribeUser(usr.Name)
		if !found {
			this.Logger.Error(nil, "User authenticated but not found by Describe.", "user", usr.Name)
			this.HttpError(response, "User authenticated but not found by Describe.", http.StatusInternalServerError)
			return
		}
		prvd, err := this.Providers.GetProvider(userDescription.Authority)
		if err != nil {
			this.Logger.Error(err, "User authenticated but its authority is not found.", "user", usr.Name, "authority", userDescription.Authority)
			this.HttpError(response, "User authenticated its Authority is not found.", http.StatusInternalServerError)
			return
		}
		var requestPayload proto.ChangePasswordRequest
		err = json.NewDecoder(request.Body).Decode(&requestPayload)
		if err != nil {
			this.HttpError(response, err.Error(), http.StatusBadRequest)
			return
		}
		err = prvd.ChangePassword(usr.Name, requestPayload.OldPassword, requestPayload.NewPassword)
		if err != nil {
			if err == providers.ErrorInvalidOldPassword {
				this.HttpError(response, err.Error(), http.StatusBadRequest)
			} else if err == providers.ErrorChangePasswordNotSupported {
				this.HttpError(response, fmt.Sprintf("The password authority for '%s' is the '%s' provider and this provider (type: '%s') does not allow password change from koobind", usr.Name, prvd.GetName(), prvd.GetType()), http.StatusBadRequest)
			} else {
				this.HttpError(response, err.Error(), http.StatusInternalServerError)
			}
		} else {
			this.HttpClose(response, "Password changed successfully.", http.StatusOK)
		}
	})
}
