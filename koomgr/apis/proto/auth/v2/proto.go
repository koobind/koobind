package v2

import "time"

// ------------------------------------------------------- Auth protocol

var LoginUrlPath = "/auth/v2/login"

type LoginRequest struct {
	Login         string `json:"login"`
	Password      string `json:"password"`
	GenerateToken bool   `json:"generateToken"`
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
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Token string `json:"token"`
	Valid bool   `json:"valid"`
}

// ---------------------------------------------------

var ChangePasswordUrlPath = "/auth/v1/changePassword"

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// Response by status code
