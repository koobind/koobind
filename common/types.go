package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)


//const (
//	V1ValidateTokenUrl = "/auth/v1/validateToken"
//	V1GetToken         = "/auth/v1/getToken"
//)



type GetTokenResponse struct {
	Token string		`json:"token"`
	ClientTTL metav1.Duration	`json:"clientTTL"`
}


type TokenLifecycle struct {
	InactivityTimeout 	metav1.Duration	`json:"inactivityTimeout"`
	MaxTTL 				metav1.Duration	`json:"maxTTL"`
	ClientTTL 			metav1.Duration	`json:"clientTTL"`
}

type UserToken struct {
	Token string				`json:"token"`
	User  User					`json:"user"`
	Creation time.Time			`json:"creation"`
	LastHit time.Time			`json:"lasthit"`
	Lifecycle *TokenLifecycle	`json:"lifecycle"`
}

// NB: User and Authority are redundant as they can be computed from UserStatuses. But, we prefer code the logic of this in one place, on the server,
// instead recomputing each time we need such info.
type UserDescribeResponse struct {
	UserStatuses []UserStatus	`json:"userStatuses"`
	User User					`json:"user"`			// This field is computed
	Authority string			`json:"authority"`
}


// Used both internally and for the user describe exchange
type UserStatus struct {
	ProviderName string				`json:"provider"`		// Used for 'describe' command
	Authority bool                  `json:"authority"`		// Is this provider Authority for authentication (password) for this user (A password is defined)
	Found bool						`json:"found"`
	PasswordStatus PasswordStatus	`json:"passwordStatus"`	// For describe, always 'unchecked'
	Uid string						`json:"uid"`    		// Issued from the authoritative server (The first one which checked the password).
	Groups []string					`json:"groups"`
	Email string					`json:"email"`
	CommonName string				`json:"commonName"`
	Messages []string				`json:"messages"`		// To report error or explanation i.e broken link in crd provider, or disabled link
}

type PasswordStatus int

const (
	Unchecked PasswordStatus = iota		// Either caller don't want to check or LdapClient.CredentialAuthority == False
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

// Needed as member of  Token CRD
func (in *User) DeepCopyInto(out *User) {
	*out = *in
	out.Groups = make([]string, len(in.Groups))
	copy(out.Groups, in.Groups)
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

