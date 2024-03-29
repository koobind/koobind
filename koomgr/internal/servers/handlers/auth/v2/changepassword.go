package v2

import (
	"encoding/json"
	"fmt"
	proto "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
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
		var requestPayload proto.ChangePasswordRequest
		err := json.NewDecoder(request.Body).Decode(&requestPayload)
		if err != nil {
			this.HttpError(response, err.Error(), http.StatusBadRequest)
			return
		}
		if clientId := this.LookupClient(requestPayload.Client); clientId == "" {
			this.Logger.Info("Invalid clientId or clientSecret for password change.", "clientId", requestPayload.Client.Id, "user", usr.Name)
			this.HttpError(response, "Invalid clientId/clientSecret", http.StatusForbidden)
			return
		}
		found, userDescription := this.Providers.DescribeUser(usr.Name)
		if !found {
			this.Logger.Error(nil, "User authenticated but not found by Describe.", "clientId", requestPayload.Client.Id, "user", usr.Name)
			this.HttpError(response, "User authenticated but not found by Describe.", http.StatusInternalServerError)
			return
		}
		prvd, err := this.Providers.GetProvider(userDescription.Authority)
		if err != nil {
			this.Logger.Error(err, "User authenticated but its authority is not found.", "clientId", requestPayload.Client.Id, "user", usr.Name, "authority", userDescription.Authority)
			this.HttpError(response, "User authenticated its Authority is not found.", http.StatusInternalServerError)
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
