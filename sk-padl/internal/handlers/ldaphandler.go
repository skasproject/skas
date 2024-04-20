package handlers

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/nmcclain/ldap"
	"net"
	"regexp"
	"skas/sk-common/pkg/skclient"
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
	log      logr.Logger
	provider skclient.SkClient
}

func New(logger logr.Logger, provider skclient.SkClient) LdapHandler {
	return &ldapHandler{
		log:      logger,
		provider: provider,
	}
}

var _ LdapHandler = &ldapHandler{}

func (h ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	h.log.V(1).Info("Bind()", "bindDN", bindDN, "remote", conn.RemoteAddr().String())

	if bindDN == config.Conf.RoBindDn {
		// It is THE admin (ro) bind password
		if bindSimplePw != config.Conf.RoBindPassword {
			h.log.V(1).Info("Bind() FAILED", "bindDN", bindDN, "remote", conn.RemoteAddr().String())
			return ldap.LDAPResultInvalidCredentials, nil
		}
		h.log.V(1).Info("Bind() Success", "bindDN", bindDN, "remote", conn.RemoteAddr().String())
		return ldap.LDAPResultSuccess, nil
	} else {
		// It is a bind for a standard user password
		uid := h.getUserIdFromDn(bindDN)
		if uid == "" {
			return ldap.LDAPResultOperationsError, fmt.Errorf("unable to extract uid from DN '%s'", bindDN)
		}
		_, status, err := h.getSkasUserFromUid(uid, bindSimplePw)
		if err != nil {
			return ldap.LDAPResultOperationsError, err
		}
		if status != proto.PasswordChecked {
			h.log.V(1).Info("Bind() FAILED", "bindDN", bindDN, "remote", conn.RemoteAddr().String(), "status", status)
			return ldap.LDAPResultInvalidCredentials, nil
		}
		h.log.V(1).Info("Bind() Success", "bindDN", bindDN, "remote", conn.RemoteAddr().String())
		return ldap.LDAPResultSuccess, nil
	}
}

func (h ldapHandler) Search(boundDN string, req ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	if boundDN != config.Conf.RoBindDn {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInvalidCredentials}, nil
	}

	h.log.V(1).Info("Search()", "boundDN", boundDN, "remote", conn.RemoteAddr().String(), "baseDN", req.BaseDN, "filter", req.Filter)
	h.log.V(2).Info(dumpSearchRequest("", req))

	// We can't have aliases. So whatever value in req.DerefAliases is OK.
	//if req.DerefAliases != ldap.NeverDerefAliases { // [-a {never|always|search|find}
	//	// Server DerefAliases not supported: RFC4511 4.5.1.3
	//	return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("req.DerefAliases != ldap.NeverDerefAliases. Not supported")
	//}
	// We don't handle req.TimeLimit. But, let's do as we do it
	//if req.TimeLimit > 0 {
	//	// Server TimeLimit not implemented
	//	return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: req.TimeLimit > 0. Not supported")
	//}

	entries := []*ldap.Entry{}
	if req.BaseDN == config.Conf.UsersBaseDn {
		// It is a search for user information from baseDN, using filter
		uid := h.extractUidFromFilter(req.Filter, config.UidFromUserFilterRegexes)
		if uid == "" {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'uid' from user filter '%s'", req.Filter)
		}
		entry, err := h.getUserEntryFromUid(uid)
		if err != nil {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, err
		}
		if entry != nil {
			entries = append(entries, entry)
		}
	} else if req.BaseDN == config.Conf.GroupsBaseDn {
		// It is a search for group list, for a given uid
		uid := h.extractUidFromFilter(req.Filter, config.UidFromGroupFilterRegexes)
		if uid == "" {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'uid' from group filter '%s'", req.Filter)
		}
		user, _, err := h.getSkasUserFromUid(uid, "")
		if err != nil {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, err
		}
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
	} else if isFilterEmpty(req.Filter) {
		if strings.HasSuffix(req.BaseDN, config.Conf.UsersBaseDn) {
			// The baseDN is the searched DN. MinIO use this form on 'mc idp ldap policy attach --user=....'
			uid := h.getUserIdFromDn(req.BaseDN)
			if uid == "" {
				return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'uid' from baseDN '%s' (filter='%s')", req.BaseDN, req.Filter)
			}
			entry, err := h.getUserEntryFromUid(uid)
			if err != nil {
				return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, err
			}
			if entry != nil {
				entries = append(entries, entry)
			}
		} else if strings.HasSuffix(req.BaseDN, config.Conf.GroupsBaseDn) {
			// The baseDN is the searched DN. MinIO use this form on 'mc idp ldap policy attach --groups=....'
			cn := h.getCnFromDn(req.BaseDN)
			if cn == "" {
				return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Unable to extract 'cn' from baseDN '%s' (filter='%s')", req.BaseDN, req.Filter)
			}
			// In such case, we don't try to check if group really exits. All groups may potentially exits
			attrs := []*ldap.EntryAttribute{}
			attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"top", "groupOfUniqueNames"}})
			attrs = append(attrs, &ldap.EntryAttribute{Name: "cn", Values: []string{cn}})
			dn := fmt.Sprintf("cn=%s,%s", cn, config.Conf.GroupsBaseDn)
			h.log.V(1).Info("Search group result", "dn", dn)
			entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
		} else {
			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Filter is empty and baseDn ('%s') does not match users or groups one", req.BaseDN)
		}
	} else {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("ERROR: Invalid baseDN=%s on search request", req.BaseDN)
	}

	result := ldap.ServerSearchResult{
		Entries:    entries,
		Referrals:  make([]string, 0),
		Controls:   make([]ldap.Control, 0),
		ResultCode: ldap.LDAPResultSuccess,
	}
	logSearchResult(result, h.log)
	return result, nil
}

func (h ldapHandler) getUserEntryFromUid(uid string) (*ldap.Entry, error) {
	user, _, err := h.getSkasUserFromUid(uid, "")
	if err != nil {
		return nil, err
	}
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
		return &ldap.Entry{DN: dn, Attributes: attrs}, nil
	} else {
		return nil, nil
	}
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

func (h ldapHandler) extractUidFromFilter(filter string, regexps []*regexp.Regexp) string {
	for _, regex := range regexps {
		matches := regex.FindStringSubmatch(filter)
		if len(matches) == 2 {
			h.log.V(2).Info("extractUidFromFilter() SUCCESS", "regex", regex.String(), "filter", filter)
			return matches[1]
		}
		h.log.V(2).Info("extractUidFromFilter() attempt failure", "regex", regex.String(), "filter", filter)
	}
	return ""
}

func (h ldapHandler) getUserIdFromDn(dn string) string {
	for _, regex := range config.UidFromDnRegexes {
		matches := regex.FindStringSubmatch(dn)
		if len(matches) == 2 {
			h.log.V(2).Info("getUserIdFromDn() SUCCESS", "regex", regex.String(), "dn", dn, "uid", matches[1])
			return matches[1]
		}
		h.log.V(2).Info("getUserIdFromDn() attempt failure", "regex", regex.String(), "dn", dn)
	}
	return ""
}

func (h ldapHandler) getCnFromDn(dn string) string {
	for _, regex := range config.CnFromDnRegexes {
		matches := regex.FindStringSubmatch(dn)
		if len(matches) == 2 {
			h.log.V(2).Info("getCnFromDn() SUCCESS", "regex", regex.String(), "dn", dn, "cn", matches[1])
			return matches[1]
		}
		h.log.V(2).Info("getCnFromDn() attempt failure", "regex", regex.String(), "dn", dn)
	}
	return ""
}

func (h ldapHandler) getSkasUserFromUid(uid string, password string) (*proto.User, proto.Status, error) {
	request := &proto.IdentityRequest{
		ClientAuth: h.provider.GetClientAuth(),
		Login:      uid,
		Password:   password,
		Detailed:   false,
	}
	response := &proto.IdentityResponse{}
	err := h.provider.Do(proto.IdentityMeta, request, response, nil)
	if err != nil {
		return nil, proto.Undefined, err
	}
	if response.Status == proto.Undefined || response.Status == proto.UserNotFound || response.Status == proto.Disabled {
		return nil, response.Status, nil
	}
	return &response.User, response.Status, nil
}

func isFilterEmpty(filter string) bool {
	for _, f := range config.Conf.EmptyFilters {
		if f == filter {
			return true
		}
	}
	return false
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

	b.WriteString(fmt.Sprintf("%s\tScope: %d\n", prefix, req.Scope))
	b.WriteString(fmt.Sprintf("%s\tSizeLimit: %d\n", prefix, req.SizeLimit))
	b.WriteString(fmt.Sprintf("%s\tTimeLimit: %d\n", prefix, req.TimeLimit))
	b.WriteString(fmt.Sprintf("%s\tDerefAliases: %d\n", prefix, req.DerefAliases))
	b.WriteString(fmt.Sprintf("%s\tTypesOnly: %t\n", prefix, req.TypesOnly))

	b.WriteString(prefix + "}")
	return b.String()
}

func logSearchResult(result ldap.ServerSearchResult, logger logr.Logger) {
	entries := make([]string, 0, 10)
	for _, entry := range result.Entries {
		entries = append(entries, entry.DN)
	}
	logger.V(0).Info("ServerSearchResult", "entries", entries)
}
