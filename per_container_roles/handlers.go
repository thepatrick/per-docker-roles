package per_container_roles

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func AllIssuesHandlers(e *Endpoint) (http.HandlerFunc, http.HandlerFunc, http.HandlerFunc) {
	putTokenHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			log.Println("PUT Token - Error: not a PUT request.")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if DenyXForwardedFor(w, r) {
			log.Println("PUT Token - Error: X-Forwareded-For is not allowed")
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			log.Println("PUT Token - Error:", err.Error(), "Container IP:", remoteIP)
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			log.Println("PUT Token - Error: No container with this IP.", "Container IP:", remoteIP)
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "no known container with that IP")
			return
		}

		if container.RoleName == "" {
			log.Println("PUT Token - Error: Container has no role", "Container IP:", remoteIP)
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "container has no role")
			return
		}

		// TODO: Port should be customizable
		requestURL := fmt.Sprintf("http://localhost:%d%s", 9911, TOKEN_RESOURCE_PATH)

		req, err := http.NewRequest(http.MethodPut, requestURL, nil)
		if err != nil {
			log.Println("PUT Token - IMDS - Error:", err.Error(), "Container IP:", remoteIP)

			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "could not connect to IMDS")
			return
		}

		tokenTTLStr := r.Header.Get(EC2_METADATA_TOKEN_TTL_HEADER)
		if tokenTTLStr != "" {
			req.Header.Add(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTLStr)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("PUT Token - IMDS - Error:", err.Error(), "Container IP:", remoteIP)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "could not connect to IMDS")
			return
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("PUT Token - IMDS - Error:", err.Error(), "Container IP:", remoteIP)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "could not connect to IMDS")
			return
		}

		if res.StatusCode == http.StatusOK {
			upstreamTTLHeader := res.Header.Get(EC2_METADATA_TOKEN_TTL_HEADER)
			w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, upstreamTTLHeader)
			w.Header().Set("X-Likely-Role", container.RoleName)
			log.Println("PUT Token - OK", "Container IP:", remoteIP)
		} else {
			log.Println("PUT Token - IDMS - Error:", res.StatusCode, "Container IP:", remoteIP)

		}

		w.WriteHeader(res.StatusCode)
		io.WriteString(w, string(resBody))
	}

	// Handles requests to /latest/meta-data/iam/security-credentials/
	getRoleNameHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			log.Println("GET RoleList - Error: Not a GET request.")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if DenyXForwardedFor(w, r) {
			log.Println("GET RoleList - Error: X-Forwarded-For header is not allowed.")
			return
		}

		tokenTTL, _, err := VerifyToken(w, r)
		if err != nil {
			log.Println("GET RoleList - Error: Token is not valid.", err.Error())
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			log.Println("GET RoleList - Could not get IP.", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			log.Println("GET RoleList - Error: No container known with this IP. Contaienr IP:", remoteIP)
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "no known container with that IP")
			return
		}

		if container.RoleName == "" {
			log.Println("GET RoleList - Error: Container has no role label. Contaienr IP:", remoteIP)
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "container has no role")
			return
		}

		log.Println("GET RoleList - OK. Contaienr IP:", remoteIP)

		w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTL)
		io.WriteString(w, container.RoleName)
	}

	// Handles GET requests to /latest/meta-data/iam/security-credentials/<ROLE_NAME>
	getCredentialsHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			log.Println("GET RoleCredentials - Error: Not a GET request.")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		roleName := r.PathValue("roleName")

		if DenyXForwardedFor(w, r) {
			log.Println("GET RoleCredentials - Error: X-Forwarded-For header is not allowed.")
			return
		}

		tokenTTL, upstreamRoleName, err := VerifyToken(w, r)
		if err != nil {
			log.Println("GET RoleCredentials - Error: Token is not valid.", err.Error())
			return
		}

		remoteIP, err := GetRemoteIP(r)
		if err != nil {
			log.Println("GET RoleCredentials - Error: Could not get IP.", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		container, ok := e.CredsByIP(remoteIP)

		if !ok {
			log.Println("GET RoleCredentials - Error: No container known with this IP. Contaienr IP:", remoteIP)
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
			log.Println("GET RoleCredentials - Error: Container has no role label. Contaienr IP:", remoteIP)
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "role does not match container role")
			return
		}

		w.Header().Set(EC2_METADATA_TOKEN_TTL_HEADER, tokenTTL)

		// container.Creds
		var nextRefreshTime = container.Creds.Expiration.Add(-RefreshTime)
		if time.Until(nextRefreshTime) < RefreshTime {
			// refresh the creds

			log.Println("GET RoleCredentials - Refreshing... Contaienr IP:", remoteIP, "Role ARN:", container.RoleARN)

			assumedRoleCredentials, err := GenerateCredentials(r.Header.Get(EC2_METADATA_TOKEN_HEADER), upstreamRoleName, container.RoleARN, container.RoleSessionName)
			if err != nil {
				log.Println("GET RoleCredentials - Refreshing... Error:", err.Error(), "Container IP:", remoteIP, "Role ARN:", container.RoleARN)
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "failed to generate credentials")
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
			log.Println("GET RoleCredentials - Error: Writing JSON", err.Error(), ". Contaienr IP:", remoteIP)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "failed to encode credentials")
			return
		}

		log.Println("GET RoleCredentials - OK. Contaienr IP:", remoteIP)
	}

	return putTokenHandler, getRoleNameHandler, getCredentialsHandler
}
