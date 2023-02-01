
## All

- Secure communication with client certificate and/or clientId/secretID, and/or netpols
- Ensure copyright message in all relevant location
- Inject environment variable in config files
- Reference and normalize os.exit() code
- Switch to logrus, to have more level (i.e WARNING)
- Display client.id in relevant message
- Rename sk-xxxx to skas-xxxxx ? or ska-xxxx
- Rename userDescribe to userExplain
- Rename userStatus to userIdentity

## sk-static

- Make user file dynamic (cf dexgate or certwatcher)

## sk-ldap

- Manage certificate with the same logic as sk-merge

## sk-merge

- Manage rootCaPath in helm chart (Global and  by provider)
- On startup, perform a scan to check underlying providers (A flag to disable)
- rename to sk-bind ?
- Add an optional providerList, to modify order of provider. Needed when appending a new provider to list as 'extraProvider' in helm chart

## Doc:

Refer to topolvm docs for the following:
- Manually create a secret



