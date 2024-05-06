package per_container_roles

import "time"

type RefreshableCred struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string    // SessionToken
	Code            string    // REFRESHABLE_CRED_CODE
	Type            string    // REFRESHABLE_CRED_TYPE
	Expiration      time.Time // time.Parse(time.RFC3339, credentialProcessOutput.Expiration)
	LastUpdated     time.Time // time.Now()
}
