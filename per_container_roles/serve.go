package per_container_roles

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"

	"github.com/docker/docker/client"
)

const DefaultLocalHostAddress = "127.0.0.1"
const DefaultPort = 9912
const DefaultDockerNetwork = "bridge"

var RefreshTime = time.Minute * time.Duration(5)

const TOKEN_RESOURCE_PATH = "/latest/api/token"
const SECURITY_CREDENTIALS_RESOURCE_PATH = "/latest/meta-data/iam/security-credentials/"

const EC2_METADATA_TOKEN_HEADER = "x-aws-ec2-metadata-token"
const EC2_METADATA_TOKEN_TTL_HEADER = "x-aws-ec2-metadata-token-ttl-seconds"
const DEFAULT_TOKEN_TTL_SECONDS = "21600"

const X_FORWARDED_FOR_HEADER = "X-Forwarded-For"

const REFRESHABLE_CRED_TYPE = "AWS-HMAC"
const REFRESHABLE_CRED_CODE = "Success"

func GenerateCredentials(token string, upstreamRoleName string, roleArn string, roleSessionName string) (*stsTypes.Credentials, error) {
	upstreamCreds, err := GetUpstreamCreds(token, upstreamRoleName)

	if err != nil {
		return nil, fmt.Errorf("could not get upstream creds: %s", err.Error())
	}

	client := sts.New(sts.Options{
		Region:      "ap-southeast-2", // TODO: make this configurable
		Credentials: credentials.NewStaticCredentialsProvider(upstreamCreds.AccessKeyId, upstreamCreds.SecretAccessKey, upstreamCreds.Token),
	})

	assumedRole, err := client.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		DurationSeconds: aws.Int32(3600), // 1 hour
		RoleSessionName: aws.String(roleSessionName),
		Tags:            []stsTypes.Tag{},
	})

	if err != nil {
		return nil, fmt.Errorf("could not assume role: %s", err.Error())
	}

	log.Println("Assumed Role:", assumedRole.AssumedRoleUser)

	return assumedRole.Credentials, nil
}

func GetRemoteIPFromRequest(r *http.Request) (string, error) {
	remoteIP := strings.Split(r.RemoteAddr, ":")[0]

	return remoteIP, nil
}

func GetRemoteIP(r *http.Request) (string, error) {
	// remoteIP := r.Header.Get("X-Real-IP")
	// if remoteIP == "" {
	// 	return "", errors.New("unable to process requests without X-Real-IP header")
	// }

	// return remoteIP, nil
	return GetRemoteIPFromRequest(r)
}

func Serve(port int, listenAddress string, dockerNetwork string) {
	// todo: specify the network name
	endpoint := &Endpoint{PortNum: port, NetworkID: dockerNetwork, ByContainer: make(map[string]*ContainerWithCreds)}
	endpoint.Server = &http.Server{}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	log.Println("Connected to Docker daemon")

	ready, errors := endpoint.ConfigureFromDocker(cli, ctx)

	select {
	case err := <-errors:
		log.Println("Error configuring from Docker:", err)
		panic(err)
	case <-ready:
		log.Println("Basic configuration available, go!")
	}

	putTokenHandler, getRoleNameHandler, getCredentialsHandler := AllIssuesHandlers(endpoint)

	http.HandleFunc(TOKEN_RESOURCE_PATH, putTokenHandler)
	http.HandleFunc(SECURITY_CREDENTIALS_RESOURCE_PATH, getRoleNameHandler)
	http.HandleFunc(SECURITY_CREDENTIALS_RESOURCE_PATH+"{roleName}", getCredentialsHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddress, endpoint.PortNum))
	if err != nil {
		log.Println("failed to create listener")
		os.Exit(1)
	}
	endpoint.PortNum = listener.Addr().(*net.TCPAddr).Port
	log.Println("Local server started on port:", endpoint.PortNum)
	log.Println("Make it available to the sdk by running:")
	log.Printf("export AWS_EC2_METADATA_SERVICE_ENDPOINT=http://%s:%d/", listenAddress, endpoint.PortNum)
	if err := endpoint.Server.Serve(listener); err != nil {
		log.Println("Httpserver: ListenAndServe() error")
		os.Exit(1)
	}

}
