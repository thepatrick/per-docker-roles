package per_container_roles

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func AllIssuesHandlers(e *Endpoint) (http.HandlerFunc, http.HandlerFunc, http.HandlerFunc) {
	putTokenHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if DenyXForwardedFor(w, r) {
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		log.Println("Looking up container by IP:", remoteIP)

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "no known container with that IP")
			return
		}

		if container.RoleName == "" {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "container has no role")
			return
		}

		// TODO: Port should be customizable
		requestURL := fmt.Sprintf("http://localhost:%d%s", 9911, TOKEN_RESOURCE_PATH)

		req, err := http.NewRequest(http.MethodPut, requestURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "could not connect to IMDS")
			fmt.Printf("client: could not create request: %s\n", err)
			return
		}

		tokenTTLStr := r.Header.Get(EC2_METADATA_TOKEN_TTL_HEADER)
		if tokenTTLStr != "" {
			req.Header.Add(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTLStr)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "could not connect to IMDS")
			fmt.Printf("client: error making http request: %s\n", err)
			return
		}

		fmt.Printf("client: got response!\n")
		fmt.Printf("client: status code: %d\n", res.StatusCode)

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}

		if res.StatusCode == http.StatusOK {
			upstreamTTLHeader := res.Header.Get(EC2_METADATA_TOKEN_TTL_HEADER)
			w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, upstreamTTLHeader)
			w.Header().Set("X-Likely-Role", container.RoleName)
		}

		w.WriteHeader(res.StatusCode)
		io.WriteString(w, string(resBody))

	}

	// Handles requests to /latest/meta-data/iam/security-credentials/
	getRoleNameHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if DenyXForwardedFor(w, r) {
			return
		}

		tokenTTL, _, err := VerifyToken(w, r)
		if err != nil {
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		log.Println("Looking up container by IP:", remoteIP)

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "no known container with that IP")
			return
		}

		if container.RoleName == "" {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "container has no role")
			return
		}

		w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTL)
		io.WriteString(w, container.RoleName)
	}

	// Handles GET requests to /latest/meta-data/iam/security-credentials/<ROLE_NAME>
	getCredentialsHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		roleName := r.PathValue("roleName")

		if DenyXForwardedFor(w, r) {
			return
		}

		tokenTTL, upstreamRoleName, err := VerifyToken(w, r)
		if err != nil {
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		log.Println("Looking up container by IP:", remoteIP)

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "no known container with that IP")
			return
		}

		if container.RoleName == "" {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "container has no role")
			return
		}

		if roleName != container.RoleName {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "role does not match container role")
			log.Println("Role does not match container role, expected:", container.RoleName, "got:", roleName)
			return
		}

		w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTL)

		// container.Creds
		var nextRefreshTime = container.Creds.Expiration.Add(-RefreshTime)
		if time.Until(nextRefreshTime) < RefreshTime {
			// refresh the creds
			log.Println("Refreshing creds for container:", container.RoleARN, container.IPAddress)

			assumedRoleCredentials, err := GenerateCredentials(r.Header.Get(EC2_METADATA_TOKEN_HEADER), upstreamRoleName, container.RoleARN, container.RoleSessionName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "failed to generate credentials")
				log.Println("Failed to generate credentials:", err)
				return
			}

			container.UpdateCreds(RefreshableCred{
				AccessKeyId:     *assumedRoleCredentials.AccessKeyId,
				SecretAccessKey: *assumedRoleCredentials.SecretAccessKey,
				Token:           *assumedRoleCredentials.SessionToken,
				Expiration:      *assumedRoleCredentials.Expiration,
				LastUpdated:     time.Now(),
				Code:            REFRESHABLE_CRED_CODE,
				Type:            REFRESHABLE_CRED_TYPE,
			})
		}

		err = json.NewEncoder(w).Encode(container.Creds)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "failed to encode credentials")
			return
		}
	}

	return putTokenHandler, getRoleNameHandler, getCredentialsHandler
}
