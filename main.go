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
	"github.com/oracle/oci-go-sdk/v65/common/auth"
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

// func GetDeploySecret(configurationProvider common.ConfigurationProvider, secretID string) string {
func GetDeploySecret(secretID string) string {
	provider, err := auth.InstancePrincipalConfigurationProvider()
	helpers.FatalIfError(err)
	client, err := secrets.NewSecretsClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).
	req := secrets.GetSecretBundleRequest{
		//SecretId: common.String(secretID),
		SecretId: &secretID,
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

// so bad, refactor these GetDeploySecret funcs
func GetDeploySecret2(configurationProvider common.ConfigurationProvider, secretID string) string {
	client, err := secrets.NewSecretsClientWithConfigurationProvider(configurationProvider)
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).
	req := secrets.GetSecretBundleRequest{
		SecretId: common.String(secretID),
		//SecretId: &secretID,
		Stage: secrets.GetSecretBundleStageLatest,
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
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	apex_url := os.Args[2] + "deploy/registered_apps/"

	r, err := http.NewRequest("GET", apex_url, nil)
	if err != nil {
		fmt.Println("could not create request")
	}
	r.Header.Add("Authorization", "Bearer "+token)

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
	//apps := body["items"].(string)
	fmt.Println(body)
}

func GetToken(cred string) string {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	token_url := os.Args[2] + "oauth/token"
	//fmt.Println(token_url)

	r, err := http.NewRequest("POST", token_url, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("could not create token request")
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", strings.Split(cred, ":")[1])

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

func deployExport(token string) {
	f, err := os.Open(os.Args[3])
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	deployURL := os.Args[2] + "deploy/app/0/"
	fmt.Println(deployURL)
	req, err := http.NewRequest("POST", deployURL, f)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/sql")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
	defer resp.Body.Close()
}

func PrintArgsErrors() {
	fmt.Println("Please pass required args:")
	fmt.Println("command (deploy) url filename vault_secret_ocid -e|i (mode) optional_app_id")
	fmt.Println("deploy https://apex.oracle.com/pls/apex/endpoint/ ackie ocid1.vaultsecret.oc1.phx.realocidhereasljsaflkasdjfewkrjllhkjyfvj")
	os.Exit(1)
}

func spacer() {
	fmt.Println("|------------------------------------------------|")
	fmt.Println()
	fmt.Println("|------------------------------------------------|")
}

func ValidateArgs() bool {
	if len(os.Args) < 6 {
		fmt.Println("Too few arguments.")
		return false
	}
	if os.Args[1] != "deploy" && os.Args[1] != "export" {
		fmt.Println("First arg needs to be deploy or export")
		return false
	}
	if !ValidateUrl(os.Args[2]) {
		fmt.Println("Double check the URL")
		return false
	}
	if !ValidateOcid(os.Args[4], "secret") {
		fmt.Println("Double check the secret ocid")
		return false
	}
	if os.Args[5] != "-e" && os.Args[5] != "-i" {
		fmt.Println("Fifth argument, -e to use ENV variables, -i to use instance principals")
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
	var cred string
	if os.Args[5] == "-e" { // use below when not using instance principals
		configurationProvider := initializeLocalConfigurationProvider()
		cred = GetDeploySecret2(configurationProvider, os.Getenv("secret_ocid"))
	} else if os.Args[5] == "-i" {
		// below uses instance principals
		cred = GetDeploySecret(os.Args[4])
	}
	fmt.Println(cred)
	token := GetToken(cred)
	//spacer()
	//GetAllApps(token)
	//spacer()
	deployExport(token)
	//deployExport("putarealtokenhereok")
}
