package proto

import (
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type GetTokenResponse struct {
	Token     string          `json:"token"`
	ClientTTL metav1.Duration `json:"clientTTL"`
}

type UserDescribeResponse struct {
	User tokenapi.UserDesc `json:"user"`
}

// This UserToken is analogous to tokenapi.Token, but for usage out of k8s/controllerRuntime (admin api, memory token storage, ...)
type UserToken struct {
	Token   string             `json:"token"`
	Spec    tokenapi.TokenSpec `json:"spec"`
	LastHit time.Time          `json:"lasthit"`
}

type TokenListResponse struct {
	Tokens []UserToken `json:"tokens"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// The two following messages are dictated by API server webhook protocol
type ValidateTokenRequest struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Token string `json:"token"`
	} `json:"spec"`
}

type ValidateTokenUser struct {
	Username string   `json:"username"`
	Uid      string   `json:"uid"`
	Groups   []string `json:"groups"`
}

type ValidateTokenResponse struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Authenticated bool               `json:"authenticated"`
		User          *ValidateTokenUser `json:"user,omitempty"`
	} `json:"status"`
}
