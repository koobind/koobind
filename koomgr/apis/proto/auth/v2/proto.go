package v2

import (
	"time"
)

type AuthClient struct {
	Id     string `json:"id"`
	Secret string `json:"secret"`
}

// ------------------------------------------------------- Auth protocol

var LoginUrlPath = "/auth/v2/login"

type LoginRequest struct {
	Login         string     `json:"login"`
	Password      string     `json:"password"`
	GenerateToken bool       `json:"generateToken"`
	Client        AuthClient `json:"client"`
}

type LoginResponse struct {
	Username      string        `json:"username"`
	CommonNames   []string      `json:"commonNames"`
	Uid           string        `json:"uid"`
	Emails        []string      `json:"emails"`
	EmailVerified bool          `json:"emailVerified"` // Not used for now
	Groups        []string      `json:"groups"`
	Token         string        `json:"token"`
	ClientTTL     time.Duration `json:"clientTTL"`
}

// ---------------------------------------------------

var ValidateTokenUrlPath = "/auth/v2/validateToken"

type ValidateTokenRequest struct {
	Token  string     `json:"token"`
	Client AuthClient `json:"client"`
}

type ValidateTokenResponse struct {
	Token string `json:"token"`
	Valid bool   `json:"valid"`
}

// ---------------------------------------------------

var ChangePasswordUrlPath = "/auth/v1/changePassword"

type ChangePasswordRequest struct {
	OldPassword string     `json:"oldPassword"`
	NewPassword string     `json:"newPassword"`
	Client      AuthClient `json:"client"`
}

// Response by status code
