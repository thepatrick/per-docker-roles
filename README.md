## Whacky idea

- [x] if x-forwarded-for is set, 403
- [x] PUT /latest/api/token/ -> aws_signing_helper
- [x] GET /latest/meta-data/iam/security-credentials/ ->
    1. [x] -> aws_signing_helper (as a cheap hack to validate the token is valid)
    2. [x] -> based on incoming IP, lookup docker container, find label (if none, error out), return that role name
- [x] GET /latest/meta-data/iam/security-credentials/<ROLE_NAME> ->
    1. [x] -> aws_signing_helper/latest/meta-data/iam/security-credentials/ (both as a cheap hack to validate the token is valid, and because we need to know it)
    2. [x] -> based on incoming IP, lookup docker container, find label (if none, error out)
    3. [x] -> do we have a current cached credential for this role? if yes, return cached value now
    3. [x] -> aws_signing_helper/latest/meta-data/iam/security-credentials/<1:ROLE_NAME>
    4. [x] -> aws sdk: assume role <ROLE_NAME>, cache, return cred

the refreshable cred looks like:

```go
const REFRESHABLE_CRED_TYPE = "AWS-HMAC"
const REFRESHABLE_CRED_CODE = "Success"

type RefreshableCred struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string      // SessionToken
	Code            string      // REFRESHABLE_CRED_CODE
	Type            string      // REFRESHABLE_CRED_TYPE
	Expiration      time.Time   // time.Parse(time.RFC3339, credentialProcessOutput.Expiration)
	LastUpdated     time.Time   // time.Now()
}
```

Maybe

```shell
echo $(docker ps -a -q) | xargs docker inspect --format '{{ .NetworkSettings.IPAddress }}  {{.Id}}' | grep MY_IP

# this would work, if we required things to come through a named network, and also that the network name could be consisent...
# otherwise we might need to get this in JSON format
docker inspect --format '{{.Id}} {{.NetworkSettings.Networks.traefik_default.IPAddress}} {{.Config.Labels.maintainer}}' 8697aa4bc346
```

## ⚠️⚠️ TODOs ⚠️⚠️

These are currently littered through the codebase. 

- [ ] script building arm64 for linux builds
- [ ] github actions build and add as a release
- [ ] add ansible stuff to nyx
- [ ] ansible: pull a release of the aws cred helper, run with systemd
- [ ] ansible: pull a release of per-docker-roles, run with systemd