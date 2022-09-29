package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/secrets"
)

// example cli signature string
// rockette deploy -a 19731 -u https://apex.oracle.com/pls/apex/erudition/ -s ocid1.vaultsecret.oc1.phx.realocidhereasljsaflkasdjfewkrjllhkjyfvj -f ackie
// simplified
// rocketo deploy https://apex.oracle.com/pls/apex/erudition/ ocid1.vaultsecret.oc1.phx.realocidhereasljsaflkasdjfewkrjllhkjyfvj <optional: file> (if deploying) <optional: app id>

// Provider used for testing locally - normally use resource principal
// Environment Variables must be properly set for this to work
// Most come from api key config, secret ocid is obtained from secrets

type Config struct {
	//GrantType   string `yaml:"grant_type"`
	//ContentType string `yaml:"content_type"`
	//Authorization string `yaml:"Authorization"`
	Token string `json:"access_token"`
}

func initializeLocalConfigurationProvider() common.ConfigurationProvider {
	privateKey := os.Getenv("pem")
	tenancyOCID := os.Getenv("tenancyOCID")
	userOCID := os.Getenv("userOCID")
	fingerprint := os.Getenv("fingerprint")
	region := os.Getenv("region")
	configurationProvider := common.NewRawConfigurationProvider(tenancyOCID, userOCID, region, fingerprint, string(privateKey), nil)
	return configurationProvider
}

func GetDeploySecret(configurationProvider common.ConfigurationProvider, secretID string) string {
	client, err := secrets.NewSecretsClientWithConfigurationProvider(configurationProvider)
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).
	req := secrets.GetSecretBundleRequest{
		SecretId: common.String(secretID),
		Stage:    secrets.GetSecretBundleStageLatest,
	}

	// Send the request using the service client
	resp, err := client.GetSecretBundle(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	var content string
	base64Details, ok := resp.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	if ok {
		content = *base64Details.Content
	}
	// Decode Base64
	rawDecodedText, err := b64.StdEncoding.DecodeString(content)
	if err != nil {
		panic(err)
	}

	// Convert to string
	cred := string(rawDecodedText)

	return cred
}

func GetAllApps(token string) {
	//see list of registered_apps
	// curl -X GET -H "Authorization:Bearer sgsRXX1xg6v-uCrGNEWqvw" https://apex.oracle.com/pls/apex/dco/deploy/registered_apps/
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	apex_url := os.Args[2] + "deploy/registered_apps/"

	r, err := http.NewRequest("GET", apex_url, nil)
	if err != nil {
		fmt.Println("could not create request")
	}
	//r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", "Bearer "+token)
	//fmt.Println(r.Header)

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	// Get token from map
	body := map[string]interface{}{}
	json.NewDecoder(res.Body).Decode(&body)
	//token := body["access_token"].(string)
	fmt.Println(body)
}

func GetToken(cred string) string {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	token_url := os.Args[2] + "oauth/token"
	fmt.Println(token_url)

	r, err := http.NewRequest("POST", token_url, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("could not create token request")
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", strings.Split(cred, ":")[1])
	//fmt.Println(r.Header)

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	// Get token from map
	body := map[string]interface{}{}
	json.NewDecoder(res.Body).Decode(&body)
	token := body["access_token"].(string)

	return token
}

func PrintArgsErrors() {
	fmt.Println("Please pass required args:")
	fmt.Println("command (deploy) url filename vault_secret_ocid optional_app_id")
	fmt.Println("deploy https://apex.oracle.com/pls/apex/endpoint/ ackie ocid1.vaultsecret.oc1.phx.realocidhereasljsaflkasdjfewkrjllhkjyfvj")
	os.Exit(1)
}

func spacer() {
	fmt.Println("|------------------------------------------------|")
	fmt.Println()
	fmt.Println("|------------------------------------------------|")
}

func ValidateArgs() bool {
	if len(os.Args) < 4 || len(os.Args[3]) < 75 {
		return false
	}
	if os.Args[1] != "deploy" {
		return false
	}
	if !ValidateUrl(os.Args[2]) {
		return false
	}
	if !ValidateOcid(os.Args[3], "secret") {
		return false
	}
	return true
}

func ValidateOcid(ocid string, resource string) bool {
	if len(ocid) < 75 || !strings.Contains(ocid, "ocid") || !strings.Contains(ocid, ".") || !strings.Contains(ocid, resource) {
		return false
	}
	return true
}

func ValidateUrl(url string) bool {
	if len(url) < 12 || !strings.Contains(url, "https://") || !strings.Contains(url, ".") {
		return false
	}
	return true
}

func main() {
	if !ValidateArgs() {
		PrintArgsErrors()
		os.Exit(1)
	}
	fmt.Println("Connecting to OCI and grabbing vault secret...")
	configurationProvider := initializeLocalConfigurationProvider()
	spacer()
	cred := GetDeploySecret(configurationProvider, os.Getenv("secret_ocid"))
	token := GetToken(cred)
	GetAllApps(token)
}
