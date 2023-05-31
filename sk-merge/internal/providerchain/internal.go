package providerchain

import (
	"github.com/go-logr/logr"
	"skas/sk-common/proto/v1/proto"
	"skas/sk-merge/internal/provider"
	"sort"
)

var _ ProviderChain = &providerChain{}

type providerChain struct {
	providers []provider.Provider
	logger    logr.Logger
}

var priorityByStatus = map[proto.Status]int{
	proto.Undefined:         0,
	proto.UserNotFound:      1,
	proto.PasswordMissing:   2,
	proto.PasswordUnchecked: 3,
	proto.PasswordChecked:   4,
	proto.PasswordFail:      4,
	proto.Disabled:          5,
}

func priority(status proto.Status) int {
	return priorityByStatus[status]
}

func (pc *providerChain) GetIdentity(login, password string, detailed bool) (*proto.IdentityResponse, error) {
	response := &proto.IdentityResponse{
		User:      proto.InitUser(login),
		Status:    proto.UserNotFound,
		Details:   make([]proto.UserDetail, 0, len(pc.providers)),
		Authority: "",
	}
	for _, prvd := range pc.providers {
		userDetail, err := prvd.GetUserDetail(login, password)
		if err != nil {
			// If provider is not critical, we do not land here. (A UserDetail with Status==Undefined is returned)
			// Error logging and formatting has been performed by caller
			return nil, err
		}
		if !userDetail.Provider.CredentialAuthority && priority(userDetail.Status) > priority(proto.PasswordMissing) {
			// A non-authority provider can't check a password or disable a user
			userDetail.Status = proto.PasswordMissing
		}
		if priority(userDetail.Status) > priority(response.Status) {
			response.Status = userDetail.Status
			if priority(userDetail.Status) > priority(proto.PasswordMissing) {
				// Uid must be provided by the authority provider who test the password
				response.Uid = userDetail.Translated.Uid
				response.Authority = prvd.GetName()
			}
		}
		if userDetail.Status != proto.Undefined {
			// A provider can carry Groups information even for an non existing-user
			if userDetail.Provider.GroupAuthority {
				response.User.Groups = append(response.User.Groups, userDetail.Translated.Groups...)
			}
		}
		if detailed {
			response.Details = append(response.Details, *userDetail)
		}
	}
	response.CommonNames = dedupAndSort(response.CommonNames)
	response.Emails = dedupAndSort(response.Emails)
	response.Groups = dedupAndSort(response.Groups)
	return response, nil
}

func dedupAndSort(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0, len(stringSlice))
	for _, entry := range stringSlice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	return list
}

func (pc *providerChain) lookupProvider(name string) provider.Provider {
	for idx, client := range pc.providers {
		if client.GetName() == name {
			return pc.providers[idx]
		}
	}
	return nil
}

func (pc *providerChain) ChangePassword(request proto.PasswordChangeRequest) (*proto.PasswordChangeResponse, error) {
	prvd := pc.lookupProvider(request.Provider)
	if prvd == nil {
		return &proto.PasswordChangeResponse{
			Login:  request.Login,
			Status: proto.UnknownProvider,
		}, nil
	}
	return prvd.ChangePassword(request)
}
