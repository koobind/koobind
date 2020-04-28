package common

import "time"



const (
	V1ValidateTokenUrl = "/auth/v1/validateToken"
	V1GetToken         = "/auth/v1/getToken"
	V1Admin            = "/auth/v1/admin/"
)



type GetTokenResponse struct {
	Token string		`json:"token"`
	ClientTTL Duration	`json:"clientTTL"`
}


type TokenLifecycle struct {
	InactivityTimeout Duration	`json:"inactivityTimeout"`
	MaxTTL Duration				`json:"maxTTL"`
	ClientTTL Duration			`json:"clientTTL"`
}

type UserToken struct {
	Token string				`json:"token"`
	User  User					`json:"user"`
	Creation time.Time			`json:"creation"`
	LastHit time.Time			`json:"lasthit"`
	Lifecycle *TokenLifecycle	`json:"lifecycle"`
}

func (this *UserToken) StillValid(now time.Time) bool {
	return this.LastHit.Add(this.Lifecycle.InactivityTimeout.Duration).After(now) && this.Creation.Add(this.Lifecycle.MaxTTL.Duration).After(now)
}

func (this *UserToken) Touch(now time.Time) {
	this.LastHit = now
}

type UserDescribeResponse struct {
	UserStatuses []UserStatus	`json:"userStatuses"`
}


// This is here for the user describe exchange
type UserStatus struct {
	ProviderName string				`json:"provider"`	// Used for 'describe' command
	Found bool						`json:"found"`
	PasswordStatus PasswordStatus	`json:"passwordStatus"`
	Uid string						`json:"uid"`    // Issued from the authoritative server (The first one which checked the password).
	Groups []string					`json:"groups"`
	Email string					`json:"email"`
}

type PasswordStatus int

const (
	Unchecked PasswordStatus = iota		// Either password = "" (Caller don't want to check) or LdapClient.CredentialAuthority == False
	Checked
	Wrong
)

var passwordStatusByValue = map[PasswordStatus]string{
	Unchecked: "Unchecked",
	Checked: "Checked",
	Wrong: "Wrong",
}

func (ps PasswordStatus) String() string {
	return passwordStatusByValue[ps]
}



type TokenListResponse struct {
	Tokens []UserToken `json:"tokens"`
}


type User struct {
	Username	string 		`json:"username"`
	Uid			string		`json:"uid"`
	Groups		[]string	`json:"groups"`
}

type ValidateTokenRequest struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Spec struct {
		Token string `json:"token"`
	} `json:"spec"`
}

type ValidateTokenResponse struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Status struct {
		Authenticated bool `json:"authenticated"`
		User *User `json:"user,omitempty"`
	} `json:"status"`
}
