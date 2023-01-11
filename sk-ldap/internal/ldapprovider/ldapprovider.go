package ldapprovider

import (
	"crypto/tls"
	"fmt"
	"github.com/go-logr/logr"
	"gopkg.in/ldap.v2"
	"skas/sk-common/pkg/httpserver/handlers"
	"skas/sk-common/proto"
	"strconv"
	"strings"
)

var _ handlers.StatusProvider = &ldapProvider{}

type ldapProvider struct {
	*Config
	hostPort         string
	tlsConfig        *tls.Config
	userSearchScope  int
	groupSearchScope int
	logger           logr.Logger
}

func (l *ldapProvider) GetUserStatus(request proto.UserStatusRequest) (*proto.UserStatusResponse, error) {
	// Set some default values
	response := proto.UserStatusResponse{
		UserStatus: proto.NotFound,
	}
	var ldapUser *ldap.Entry
	err := l.do(func(conn *ldap.Conn) error {
		var err error
		// If bindDN and bindPW are empty this will default to an anonymous bind.
		bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", l.BindDN, "xxxxxxxx")
		if err = conn.Bind(l.BindDN, l.BindPW); err != nil {
			return fmt.Errorf("%s failed: %v", bindDesc, err)
		}
		l.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
		if ldapUser, err = l.lookupUser(conn, request.Login); err != nil {
			return err
		}
		if ldapUser != nil {
			response.User = &proto.User{
				Login: request.Login,
			}
			if request.Password != "" {
				if response.UserStatus, err = l.checkPassword(conn, *ldapUser, request.Password); err != nil {
					return err
				}
			} else {
				response.UserStatus = proto.PasswordUnchecked
			}
			// We need to bind again, as password check was performed on user
			bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", l.BindDN, "xxxxxxxx")
			if err := conn.Bind(l.BindDN, l.BindPW); err != nil {
				return fmt.Errorf("%s failed: %v", bindDesc, err)
			}
			l.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
			if response.User.Groups, err = l.lookupGroups(conn, *ldapUser); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if ldapUser != nil {
		l.logger.V(2).Info(fmt.Sprint("Will fetch Attributes"))
		uid := getAttr(*ldapUser, l.UserSearch.NumericalIdAttr)

		if response.User.Uid, err = strconv.ParseInt(uid, 10, 64); err != nil {
			l.logger.Error(err, "Non numerical Uid value (%s) for user '%s'", uid, request.Login)
		}
		response.User.Emails = getAttrs(*ldapUser, l.UserSearch.EmailAttr)
		response.User.CommonNames = getAttrs(*ldapUser, l.UserSearch.CnAttr)
	}
	return &response, nil
}

// do() initializes a connection to the LDAP directory and passes it to the
// provided function. It then performs appropriate teardown or reuse before
// returning.
func (l *ldapProvider) do(f func(c *ldap.Conn) error) error {
	var (
		conn *ldap.Conn
		err  error
	)
	switch {
	case l.InsecureNoSSL:
		l.logger.V(2).Info(fmt.Sprintf("Dial('tcp', %s)", l.hostPort))
		conn, err = ldap.Dial("tcp", l.hostPort)
	case l.StartTLS:
		l.logger.V(2).Info(fmt.Sprintf("Dial('tcp', %s)", l.hostPort))
		conn, err = ldap.Dial("tcp", l.hostPort)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		l.logger.V(2).Info(fmt.Sprintf("conn.StartTLS(tlsConfig)"))
		if err := conn.StartTLS(l.tlsConfig); err != nil {
			return fmt.Errorf("start TLS failed: %v", err)
		}
	default:
		l.logger.V(2).Info(fmt.Sprintf("DialTLS('tcp', %s, tlsConfig)", l.hostPort))
		conn, err = ldap.DialTLS("tcp", l.hostPort, l.tlsConfig)
	}
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer func() {
		l.logger.V(2).Info("Closing ldap connection")
		conn.Close()
	}()

	return f(conn)
}

func (l *ldapProvider) lookupUser(conn *ldap.Conn, login string) (*ldap.Entry, error) {
	filter := fmt.Sprintf("(%s=%s)", l.UserSearch.LoginAttr, ldap.EscapeFilter(login))
	if l.UserSearch.Filter != "" {
		filter = fmt.Sprintf("(&%s%s)", l.UserSearch.Filter, filter)
	}
	// Initial search.
	req := &ldap.SearchRequest{
		BaseDN: l.UserSearch.BaseDN,
		Filter: filter,
		Scope:  l.userSearchScope,
		// We only need to search for these specific requests.
		Attributes: []string{
			l.UserSearch.LoginAttr,
		},
	}
	if l.UserSearch.NumericalIdAttr != "" {
		req.Attributes = append(req.Attributes, l.UserSearch.NumericalIdAttr)
	}
	if l.UserSearch.EmailAttr != "" {
		req.Attributes = append(req.Attributes, l.UserSearch.EmailAttr)
	}
	if l.UserSearch.CnAttr != "" {
		req.Attributes = append(req.Attributes, l.UserSearch.CnAttr)
	}
	if l.GroupSearch.LinkUserAttr != "" {
		req.Attributes = append(req.Attributes, l.GroupSearch.LinkUserAttr)
	}

	searchDesc := fmt.Sprintf("baseDN:'%s' scope:'%s' filter:'%s'", req.BaseDN, scopeString(req.Scope), req.Filter)
	resp, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("search [%s] failed: %v", searchDesc, err)
	}
	l.logger.V(2).Info(fmt.Sprintf("Performing search [%s] -> Found %d entries", searchDesc, len(resp.Entries)))

	switch n := len(resp.Entries); n {
	case 0:
		l.logger.V(2).Info(fmt.Sprintf("No results returned for filter: %q", filter))
		return nil, nil
	case 1:
		l.logger.V(2).Info(fmt.Sprintf("username %q mapped to entry %s", login, resp.Entries[0].DN))
		return resp.Entries[0], nil
	default:
		return nil, fmt.Errorf("filter returned multiple (%d) results: %q", n, filter)
	}
}

func (l *ldapProvider) checkPassword(conn *ldap.Conn, user ldap.Entry, password string) (proto.UserStatus, error) {
	if password == "" {
		return proto.PasswordFail, nil
	}
	// Try to authenticate as the distinguished name.
	bindDesc := fmt.Sprintf("conn.Bind(%s, %s)", user.DN, "xxxxxxxx")
	if err := conn.Bind(user.DN, password); err != nil {
		// Detect a bad password through the LDAP error code.
		if ldapErr, ok := err.(*ldap.Error); ok {
			switch ldapErr.ResultCode {
			case ldap.LDAPResultInvalidCredentials:
				l.logger.V(2).Info(fmt.Sprintf("%s => invalid password", bindDesc))
				return proto.PasswordFail, nil
			case ldap.LDAPResultConstraintViolation:
				// Should be a Warning
				l.logger.Error(nil, fmt.Sprintf("%s => constraint violation: %s", bindDesc, ldapErr.Error()))
				return proto.PasswordFail, nil
			}
		} // will also catch all ldap.Error without a case statement above
		return proto.PasswordFail, fmt.Errorf("%s => failed: %v", bindDesc, err)
	}
	l.logger.V(2).Info(fmt.Sprintf("%s => success", bindDesc))
	return proto.PasswordChecked, nil
}

func (l *ldapProvider) lookupGroups(conn *ldap.Conn, user ldap.Entry) ([]string, error) {
	ldapGroups := make([]*ldap.Entry, 0, 2)
	groups := make([]string, 0, 2)
	for _, attr := range getAttrs(user, l.GroupSearch.LinkUserAttr) {
		var req *ldap.SearchRequest
		filter := "(objectClass=top)" // The only way I found to have a pass through filter
		if l.GroupSearch.Filter != "" {
			filter = l.GroupSearch.Filter
		}
		if strings.ToUpper(l.GroupSearch.LinkGroupAttr) == "DN" {
			req = &ldap.SearchRequest{
				BaseDN:     attr,
				Filter:     filter,
				Scope:      ldap.ScopeBaseObject,
				Attributes: []string{l.GroupSearch.NameAttr},
			}
		} else {
			filter := fmt.Sprintf("(%s=%s)", l.GroupSearch.LinkGroupAttr, ldap.EscapeFilter(attr))
			if l.GroupSearch.Filter != "" {
				filter = fmt.Sprintf("(&%s%s)", l.GroupSearch.Filter, filter)
			}
			req = &ldap.SearchRequest{
				BaseDN:     l.GroupSearch.BaseDN,
				Filter:     filter,
				Scope:      l.groupSearchScope,
				Attributes: []string{l.GroupSearch.NameAttr},
			}

		}
		searchDesc := fmt.Sprintf("baseDN:'%s' scope:'%s' filter:'%s'", req.BaseDN, scopeString(req.Scope), req.Filter)
		resp, err := conn.Search(req)
		if err != nil {
			return []string{}, fmt.Errorf("search [%s] failed: %v", searchDesc, err)
		}
		l.logger.V(2).Info(fmt.Sprintf("Performing search [%s] -> Found %d entries", searchDesc, len(resp.Entries)))
		ldapGroups = append(ldapGroups, resp.Entries...)
	}
	for _, ldapGroup := range ldapGroups {
		gname := ldapGroup.GetAttributeValue(l.GroupSearch.NameAttr)
		if gname != "" {
			groups = append(groups, gname)
		}
	}
	return groups, nil
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
	return []string{}
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
