package clientproviderchain

import (
	"skas/sk-common/proto/v1/proto"
	"sort"
)

func Merge(login string, scanItems []ScanItem) (merged *proto.UserStatusResponse, credentialAuthorityProvider string) {
	merged = &proto.UserStatusResponse{
		User: proto.User{
			Login:       login,
			Uid:         0,
			CommonNames: make([]string, 0, 2),
			Emails:      make([]string, 0, 2),
			Groups:      make([]string, 0, 10),
		},
		UserStatus: proto.NotFound,
	}

	for _, item := range scanItems {
		newUserStatus := item.UserStatusResponse.UserStatus
		if isUserFound(newUserStatus) {
			if merged.UserStatus == proto.NotFound {
				// User found. Must be at least 'PasswordUnchecked'
				merged.UserStatus = proto.PasswordUnchecked
			}
			if merged.UserStatus == proto.PasswordUnchecked && (*item.Provider).IsCredentialAuthority() {
				merged.UserStatus = newUserStatus // May be PasswordChecked, PasswordFailed ro PasswordUnchecked
				if newUserStatus != proto.PasswordUnchecked {
					// Uid must be provided by the provider who validate the password.
					merged.Uid = item.Translated.Uid
					credentialAuthorityProvider = (*item.Provider).GetName()
				}
			}
			merged.CommonNames = append(merged.CommonNames, item.UserStatusResponse.CommonNames...)
			merged.Emails = append(merged.Emails, item.UserStatusResponse.Emails...)
			if (*item.Provider).IsGroupAuthority() {
				merged.Groups = append(merged.Groups, item.Translated.Groups...)
			}
		}
	}
	merged.CommonNames = dedupAndSort(merged.CommonNames)
	merged.Emails = dedupAndSort(merged.Emails)
	merged.Groups = dedupAndSort(merged.Groups)
	return merged, credentialAuthorityProvider
}

func dedupAndSort(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	return list
}
