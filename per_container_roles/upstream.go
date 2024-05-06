package per_container_roles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func VerifyToken(w http.ResponseWriter, r *http.Request) (string, string, error) {
	token := r.Header.Get(EC2_METADATA_TOKEN_HEADER)

	// TODO: Port should be customizable
	requestURL := fmt.Sprintf("http://localhost:%d%s", 9911, SECURITY_CREDENTIALS_RESOURCE_PATH)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "could not connect to IMDS"
		io.WriteString(w, msg)
		fmt.Printf("client: could not create request: %s\n", err)
		return "", "", errors.New(msg)
	}

	req.Header.Add(EC2_METADATA_TOKEN_HEADER, token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "could not connect to IMDS"
		io.WriteString(w, msg)
		fmt.Printf("client: error making http request: %s\n", err)
		return "", "", errors.New(msg)
	}

	log.Println("VerifyToken: got response!")
	log.Println("VerifyToken: status code:", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}

	if res.StatusCode != http.StatusOK {
		msg := "token verification failed"
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, msg)
		return "", "", errors.New(msg)
	}

	upstreamTTLHeader := res.Header.Get(EC2_METADATA_TOKEN_TTL_HEADER)
	upstreamRoleName := string(resBody)

	return upstreamTTLHeader, upstreamRoleName, nil
}

func GetUpstreamCreds(token string, upstreamRoleName string) (RefreshableCred, error) {
	requestURL := fmt.Sprintf("http://localhost:%d%s", 9911, SECURITY_CREDENTIALS_RESOURCE_PATH+upstreamRoleName)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return RefreshableCred{}, fmt.Errorf("could not connect to upstream: %s", err.Error())
	}
	req.Header.Add(EC2_METADATA_TOKEN_HEADER, token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		return RefreshableCred{}, fmt.Errorf("could not connect to upstream: %s", err.Error())
	}

	var upstreamCreds RefreshableCred
	err = json.NewDecoder(res.Body).Decode(&upstreamCreds)
	if err != nil {
		return RefreshableCred{}, fmt.Errorf("could not read response body from upstream: %s", err.Error())
	}

	log.Println("got credentials from upstream:", upstreamCreds)

	return upstreamCreds, nil
}
