package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/nmcclain/ldap"
	"net"
	"regexp"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-padl/internal/config"
	"strconv"
	"strings"
)

type LdapHandler interface {
	ldap.Binder
	ldap.Searcher
	ldap.Closer

	ldap.Comparer
	ldap.Abandoner
	ldap.Extender
	ldap.Unbinder

	ldap.Adder
	ldap.Modifier
	ldap.Deleter
	ldap.ModifyDNr
}

type ldapHandler struct {
	log logr.Logger
}

func New(logger logr.Logger) LdapHandler {
	return &ldapHandler{
		log: logger,
	}
}

var _ LdapHandler = &ldapHandler{}

func (h ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	h.log.V(1).Info("Bind()", "bindDN", bindDN, "remote", conn.RemoteAddr().String())

	if bindDN == config.Conf.RoBindDn && bindSimplePw == config.Conf.RoBindPassword {
		h.log.V(1).Info("Bind() Success", "bindDN", bindDN, "remote", conn.RemoteAddr().String())
		return ldap.LDAPResultSuccess, nil
	} else {
		h.log.V(1).Info("Bind() FAILED", "bindDN", bindDN, "remote", conn.RemoteAddr().String())
		return ldap.LDAPResultInvalidCredentials, nil
	}
}

func (h ldapHandler) Search(boundDN string, req ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	if boundDN != config.Conf.RoBindDn {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInvalidCredentials}, nil
	}

	h.log.V(1).Info("Search()", "boundDN", boundDN, "remote", conn.RemoteAddr().String())
	h.log.V(2).Info(dumpSearchRequest("", req))

	if req.DerefAliases != ldap.NeverDerefAliases { // [-a {never|always|search|find}
		// Server DerefAliases not supported: RFC4511 4.5.1.3
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("req.DerefAliases != ldap.NeverDerefAliases. Not supported")
	}
	if req.TimeLimit > 0 {
		// Server TimeLimit not implemented
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: req.TimeLimit > 0. Not supported")
	}

	entries := []*ldap.Entry{}
	if req.BaseDN == config.Conf.UsersBaseDn {
		// It is a search for user information
		uid := extractUidFromFilter(req.Filter, config.UidFromUserFilterRegexes)
		if uid == "" {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'uid' from user filter '%s'", req.Filter)
		}
		user := getUserFromUid(uid)
		if user != nil {
			attrs := []*ldap.EntryAttribute{}
			attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"top", "inetOrgPerson"}})

			attrs = append(attrs, &ldap.EntryAttribute{Name: "uid", Values: []string{user.Login}})
			if len(user.CommonNames) > 0 {
				attrs = append(attrs, &ldap.EntryAttribute{Name: "cn", Values: user.CommonNames})
				if user.CommonNames[0] != "" {
					// Get the surname as last work of the common name
					sp := strings.Split(user.CommonNames[0], " ")
					attrs = append(attrs, &ldap.EntryAttribute{Name: "sn", Values: []string{sp[len(sp)-1]}})
				}
			}
			if user.Uid != 0 {
				attrs = append(attrs, &ldap.EntryAttribute{Name: "uidNumber", Values: []string{strconv.Itoa(user.Uid)}})
			}
			if len(user.Emails) > 0 {
				attrs = append(attrs, &ldap.EntryAttribute{Name: "mail", Values: user.Emails})
			}

			dn := fmt.Sprintf("uid=%s,%s", uid, config.Conf.UsersBaseDn)
			h.log.V(1).Info("Search user result", "dn", dn)
			entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
		}
	} else if req.BaseDN == config.Conf.GroupsBaseDn {
		// It is a search for group list
		uid := extractUidFromFilter(req.Filter, config.UidFromGroupFilterRegexes)
		if uid == "" {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'uid' from group filter '%s'", req.Filter)
		}
		user := getUserFromUid(uid)
		if user != nil {
			for _, grp := range user.Groups {
				attrs := []*ldap.EntryAttribute{}
				attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"top", "groupOfUniqueNames"}})
				attrs = append(attrs, &ldap.EntryAttribute{Name: "cn", Values: []string{grp}})
				attrs = append(attrs, &ldap.EntryAttribute{Name: "memberUid", Values: []string{uid}})
				dn := fmt.Sprintf("cn=%s,%s", grp, config.Conf.GroupsBaseDn)
				h.log.V(1).Info("Search group result", "dn", dn)
				entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
			}
		}
	} else {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Invalid baseDN=%s on search request", req.BaseDN)
	}

	return ldap.ServerSearchResult{
		Entries:    entries,
		Referrals:  make([]string, 0),
		Controls:   make([]ldap.Control, 0),
		ResultCode: ldap.LDAPResultSuccess,
	}, nil
}

func (h ldapHandler) Close(boundDN string, conn net.Conn) error {
	h.log.V(1).Info("Close()", "boundDN", boundDN, "remote", conn.RemoteAddr().String())
	return nil
}

func (h ldapHandler) Compare(boundDN string, req ldap.CompareRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Compare() not implemented")
}

func (h ldapHandler) Abandon(boundDN string, conn net.Conn) error {
	return fmt.Errorf("function Abandon() not implemented")
}

func (h ldapHandler) Extended(boundDN string, req ldap.ExtendedRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Extended() not implemented")
}

func (h ldapHandler) Unbind(boundDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Unbind() not implemented")
}

func (h ldapHandler) Add(boundDN string, req ldap.AddRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Add() not implemented")
}

func (h ldapHandler) Modify(boundDN string, req ldap.ModifyRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Modify() not implemented")
}

func (h ldapHandler) Delete(boundDN, deleteDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function Delete() not implemented")
}

func (h ldapHandler) ModifyDN(boundDN string, req ldap.ModifyDNRequest, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultOperationsError, fmt.Errorf("function ModifyDN() not implemented")
}

// -------------------------------------------------------------------------------------

func dumpSearchRequest(prefix string, req ldap.SearchRequest) string {
	var b strings.Builder
	b.WriteString(prefix + "\n{\n")
	b.WriteString(prefix + "\tBaseDN: '" + req.BaseDN + "'\n")
	b.WriteString(prefix + "\tFilter: '" + req.Filter + "'\n")
	b.WriteString(prefix + "\tAttributes : [" + strings.Join(req.Attributes, ", ") + "]\n")
	if len(req.Controls) > 0 {
		for idx, c := range req.Controls {
			b.WriteString(fmt.Sprintf("%s\tControl[%d] : %v\n", prefix, idx, c))
		}
	} else {
		b.WriteString(prefix + "\tControls : []\n")
	}
	b.WriteString(fmt.Sprintf("%s\tScope: %s\n", prefix, ldap.ScopeMap[req.Scope]))

	b.WriteString(fmt.Sprintf("%s\tScope: %d\n", prefix, req.DerefAliases))
	b.WriteString(fmt.Sprintf("%s\tSizeLimit: %d\n", prefix, req.SizeLimit))
	b.WriteString(fmt.Sprintf("%s\tTimeLimit: %d\n", prefix, req.TimeLimit))
	b.WriteString(fmt.Sprintf("%s\tDerefAliases: %d\n", prefix, req.DerefAliases))
	b.WriteString(fmt.Sprintf("%s\tTypesOnly: %t\n", prefix, req.TypesOnly))

	b.WriteString(prefix + "}")
	return b.String()
}

//
//var r2 = regexp.MustCompile(`^\((\w+)=(\w+)\)$`)
//var r1 = regexp.MustCompile(`^\(\&\(objectClass=inetOrgPerson\)\((\w+)=(\w+)\)\)$`)
//
//func parseFilter(filter string) (attr string, value string, err error) {
//	matches := r1.FindStringSubmatch(filter)
//	if len(matches) == 3 {
//		return matches[1], matches[2], nil
//	}
//	matches = r2.FindStringSubmatch(filter)
//	if len(matches) == 3 {
//		return matches[1], matches[2], nil
//	}
//	return "", "", fmt.Errorf("invalid filter: '%s'", filter)
//}

// Fake function
func getUserFromUid(uid string) *proto.User {
	if uid == "user1" {
		u := &proto.User{
			Login:       uid,
			Uid:         2333,
			CommonNames: []string{"User ONE"},
			Emails:      []string{"xxx@yy.com"},
			Groups:      []string{"grp1", "grp2"},
		}
		return u
	}
	return nil
}

func extractUidFromFilter(filter string, regexps []*regexp.Regexp) string {
	for _, regex := range regexps {
		matches := regex.FindStringSubmatch(filter)
		fmt.Printf("Trying regeg %s with '%s'  len(matches)=%d\n", regex.String(), filter, len(matches))
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}
