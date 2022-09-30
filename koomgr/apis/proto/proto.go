package proto

import (
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"time"
)

// -------------------------------------------------------- Deprecated protocol

// This UserToken is analogous to tokenapi.Token, but for usage out of k8s/controllerRuntime (admin api, memory token storage, ...)
type UserToken struct {
	Token   string             `json:"token"`
	Spec    tokenapi.TokenSpec `json:"spec"`
	LastHit time.Time          `json:"lasthit"`
}

type TokenListResponse struct {
	Tokens []UserToken `json:"tokens"`
}

// ------------------------------------- DEX protocol

var DexLoginUrlPath = "/dex/v1/login"

type DexLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DexLoginResponse struct {
	Name       string   `json:"name"`
	CommonName string   `json:"commonName"`
	Uid        string   `json:"uid"`
	Email      string   `json:"email"`
	Groups     []string `json:"groups"`
	Token      string   `json:"token"`
}
