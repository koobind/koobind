/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/

package ldap

import (
	"crypto/tls"
	"fmt"
	"github.com/go-logr/logr"
	tokenapi "github.com/koobind/koobind/koomgr/apis/tokens/v1alpha1"
	"github.com/koobind/koobind/koomgr/internal/providers"
	"gopkg.in/ldap.v2"
	"strconv"
	"strings"
)

var _ providers.Provider = &ldapProvider{}

type ldapProvider struct {
	*LdapProviderConfig
	hostPort         string
	tlsConfig        *tls.Config
	userSearchScope  int
	groupSearchScope int
	logger           logr.Logger
}

func (this *ldapProvider) IsCritical() bool {
	return *this.Critical
}

func scopeString(i int) string {
	switch i {
	case ldap.ScopeBaseObject:
		return "base"
	case ldap.ScopeSingleLevel:
		return "one"
	case ldap.ScopeWholeSubtree:
		return "sub"
	default:
		return ""
	}
}

func parseScope(s string) (int, bool) {
	// NOTE(ericchiang): ScopeBaseObject doesn't really make sense for us because we
	// never know the user's or group's DN.
	switch s {
	case "", "sub":
		return ldap.ScopeWholeSubtree, true
	case "one":
		return ldap.ScopeSingleLevel, true
	}
	return 0, false
}

// do initializes a connection to the LDAP directory and passes it to the
// provided function. It then performs appropriate teardown or reuse before
// returning.
func (this *ldapProvider) do(f func(c *ldap.Conn) error) error {
	var (
		conn *ldap.Conn
		err  error
	)
	switch {
	case this.InsecureNoSSL:
		this.logger.V(2).Info(fmt.Sprintf("Dial('tcp', %s)", this.hostPort))
		conn, err = ldap.Dial("tcp", this.hostPort)
	case this.StartTLS:
		this.logger.V(2).Info(fmt.Sprintf("Dial('tcp', %s)", this.hostPort))
		conn, err = ldap.Dial("tcp", this.hostPort)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		this.logger.V(2).Info(fmt.Sprintf("conn.StartTLS(tlsConfig)"))
		if err := conn.StartTLS(this.tlsConfig); err != nil {
			return fmt.Errorf("start TLS failed: %v", err)
		}
	default:
		this.logger.V(2).Info(fmt.Sprintf("DialTLS('tcp', %s, tlsConfig)", this.hostPort))
		conn, err = ldap.DialTLS("tcp", this.hostPort, this.tlsConfig)
	}
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer func() {
		this.logger.V(2).Info("Closing ldap connection")
		conn.Close()
	}()

	return f(conn)
}

func (this *ldapProvider) lookupUser(conn *ldap.Conn, login string) (ldap.Entry, bool, error) {
	filter := fmt.Sprintf("(%s=%s)", this.UserSearch.LoginAttr, ldap.EscapeFilter(login))
	if this.UserSearch.Filter != "" {
		filter = fmt.Sprintf("(&%s%s)", this.UserSearch.Filter, filter)
	}

	// Initial search.
	req := &ldap.SearchRequest{
		BaseDN: this.UserSearch.BaseDN,
		Filter: filter,
		Scope:  this.userSearchScope,
		// We only need to search for these specific requests.
		Attributes: []string{
			this.UserSearch.LoginAttr,
		},
	}
	if this.UserSearch.NumericalIdAttr != "" {
		req.Attributes = append(req.Attributes, this.UserSearch.NumericalIdAttr)
	}
	if this.UserSearch.EmailAttr != "" {
		req.Attributes = append(req.Attributes, this.UserSearch.EmailAttr)
	}
	if this.UserSearch.CnAttr != "" {
		req.Attributes = append(req.Attributes, this.UserSearch.CnAttr)
	}
	if *this.GroupAuthority {
		req.Attributes = append(req.Attributes, this.GroupSearch.LinkUserAttr)
	}

	searchDesc := fmt.Sprintf("baseDN:'%s' scope:'%s' filter:'%s'", req.BaseDN, scopeString(req.Scope), req.Filter)
	resp, err := conn.Search(req)
	if err != nil {
		return ldap.Entry{}, false, fmt.Errorf("Search [%s] failed: %v", searchDesc, err)
	}
	this.logger.V(2).Info(fmt.Sprintf("Performing search [%s] -> Found %d entries", searchDesc, len(resp.Entries)))

	switch n := len(resp.Entries); n {
	case 0:
		this.logger.V(2).Info(fmt.Sprintf("No results returned for filter: %q", filter))
		return ldap.Entry{}, false, nil
	case 1:
		user := *resp.Entries[0]
		this.logger.V(2).Info(fmt.Sprintf("username %q mapped to entry %s", login, user.DN))
		return user, true, nil
	default:
		return ldap.Entry{}, false, fmt.Errorf("Filter returned multiple (%d) results: %q", n, filter)
	}
}

func getAttrs(e ldap.Entry, name string) []string {
	if name == "DN" {
		return []string{e.DN}
	}
	for _, a := range e.Attributes {
		if a.Name == name {
			return a.Values
		}
	}
	return nil
}

func getAttr(e ldap.Entry, name string) string {
	if name == "" {
		return ""
	}
	if a := getAttrs(e, name); len(a) > 0 {
		return a[0]
	}
	return ""
}

func (this *ldapProvider) GetUserStatus(login string, password string, checkPassword bool) (tokenapi.UserEntry, error) {
	userEntry := tokenapi.UserEntry{
		ProviderName:   this.Name,
		Authority:      *this.CredentialAuthority,
		Found:          false,
		PasswordStatus: tokenapi.PasswordStatusUnchecked,
		Uid:            "",
		Groups:         []string{},
		Email:          "",
		Messages:       make([]string, 0, 0),
	}
	var ldapUser ldap.Entry
	err := this.do(func(conn *ldap.Conn) error {
		var err error
		// If bindDN and bindPW are empty this will default to an anonymous bind.
		bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", this.BindDN, "xxxxxxxx")
		if err = conn.Bind(this.BindDN, this.BindPW); err != nil {
			return fmt.Errorf("%s failed: %v", bindDesc, err)
		}
		this.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
		if ldapUser, userEntry.Found, err = this.lookupUser(conn, login); err != nil {
			return err
		}
		if userEntry.Found {
			if checkPassword && *this.CredentialAuthority {
				if userEntry.PasswordStatus, err = this.checkPassword(conn, ldapUser, password); err != nil {
					return err
				}
			} else {
				userEntry.PasswordStatus = tokenapi.PasswordStatusUnchecked
			}
			// No need to collect groups if auth failed
			if userEntry.PasswordStatus != tokenapi.PasswordStatusWrong && *this.GroupAuthority {
				// We need to bind again, as password check was performed by groupbinding on user
				bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", this.BindDN, "xxxxxxxx")
				if err := conn.Bind(this.BindDN, this.BindPW); err != nil {
					return fmt.Errorf("%s failed: %v", bindDesc, err)
				}
				this.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
				if userEntry.Groups, err = this.lookupGroups(conn, ldapUser); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err == nil && userEntry.Found {
		this.logger.V(2).Info(fmt.Sprint("Will fetch Attributes"))
		userEntry.Uid = getAttr(ldapUser, this.UserSearch.NumericalIdAttr)
		if userEntry.Uid != "" && this.UidOffet != 0 {
			if uid, err := strconv.Atoi(userEntry.Uid); err != nil {
				// Shoud be a Warning
				this.logger.Error(err, "Non numerical Uid value (%s) for user '%s'", userEntry.Uid, login)
			} else {
				uid = uid + this.UidOffet
				userEntry.Uid = strconv.Itoa(uid)
			}
		}
		userEntry.Email = getAttr(ldapUser, this.UserSearch.EmailAttr)
		userEntry.CommonName = getAttr(ldapUser, this.UserSearch.CnAttr)
	}
	return userEntry, err
}

func (this *ldapProvider) lookupGroups(conn *ldap.Conn, user ldap.Entry) ([]string, error) {
	ldapGroups := []*ldap.Entry{}
	groups := []string{}
	for _, attr := range getAttrs(user, this.GroupSearch.LinkUserAttr) {
		var req *ldap.SearchRequest
		filter := "(objectClass=top)" // The only way I found to have a pass through filter
		if this.GroupSearch.Filter != "" {
			filter = this.GroupSearch.Filter
		}
		if strings.ToUpper(this.GroupSearch.LinkGroupAttr) == "DN" {
			req = &ldap.SearchRequest{
				BaseDN:     attr,
				Filter:     filter,
				Scope:      ldap.ScopeBaseObject,
				Attributes: []string{this.GroupSearch.NameAttr},
			}
		} else {
			filter := fmt.Sprintf("(%s=%s)", this.GroupSearch.LinkGroupAttr, ldap.EscapeFilter(attr))
			if this.GroupSearch.Filter != "" {
				filter = fmt.Sprintf("(&%s%s)", this.GroupSearch.Filter, filter)
			}
			req = &ldap.SearchRequest{
				BaseDN:     this.GroupSearch.BaseDN,
				Filter:     filter,
				Scope:      this.groupSearchScope,
				Attributes: []string{this.GroupSearch.NameAttr},
			}

		}
		searchDesc := fmt.Sprintf("baseDN:'%s' scope:'%s' filter:'%s'", req.BaseDN, scopeString(req.Scope), req.Filter)
		resp, err := conn.Search(req)
		if err != nil {
			return []string{}, fmt.Errorf("Search [%s] failed: %v", searchDesc, err)
		}
		this.logger.V(2).Info(fmt.Sprintf("Performing search [%s] -> Found %d entries", searchDesc, len(resp.Entries)))
		ldapGroups = append(ldapGroups, resp.Entries...)
	}
	for _, ldapGroup := range ldapGroups {
		gname := ldapGroup.GetAttributeValue(this.GroupSearch.NameAttr)
		if gname != "" {
			gname = fmt.Sprintf(this.GroupPattern, gname)
			groups = append(groups, gname)
		}
	}
	return groups, nil
}

func (this *ldapProvider) checkPassword(conn *ldap.Conn, user ldap.Entry, password string) (tokenapi.PasswordStatus, error) {
	if password == "" {
		return tokenapi.PasswordStatusWrong, nil
	}
	// Try to authenticate as the distinguished name.
	bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", user.DN, "xxxxxxxx")
	if err := conn.Bind(user.DN, password); err != nil {
		// Detect a bad password through the LDAP error code.
		if ldapErr, ok := err.(*ldap.Error); ok {
			switch ldapErr.ResultCode {
			case ldap.LDAPResultInvalidCredentials:
				this.logger.V(2).Info(fmt.Sprintf("%s => invalid password", bindDesc))
				return tokenapi.PasswordStatusWrong, nil
			case ldap.LDAPResultConstraintViolation:
				// Should be a Warning
				this.logger.Error(nil, fmt.Sprintf("%s => constraint violation: %s", bindDesc, ldapErr.Error()))
				return tokenapi.PasswordStatusWrong, nil
			}
		} // will also catch all ldap.Error without a case statement above
		return tokenapi.PasswordStatusWrong, fmt.Errorf("%s => failed: %v", bindDesc, err)
	}
	this.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
	return tokenapi.PasswordStatusChecked, nil
}

func (this *ldapProvider) ChangePassword(user string, oldPassword string, newPassword string) error {
	return providers.ErrorChangePasswordNotSupported
}
