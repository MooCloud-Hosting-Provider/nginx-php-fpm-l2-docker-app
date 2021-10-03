package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"./nginxParser"
	"./phpParser"
)

var mooCLIConfigParserMap map[string]map[string]interface{} = map[string]map[string]interface{}{
	"nginx": {
		"fastcgi_cache_profile":  nginxParser.FastcgiCacheProfile,
		"fastcgi_cache_location": nginxParser.FastcgiCacheLocation,
	},
	"php": {
		"upload_max_filesize": phpParser.UploadMaxFilesize,
		"post_max_size":       phpParser.PostMaxSize,
		"memory_limit":        phpParser.MemoryLimit,
		"allow_url_fopen":     phpParser.AllowUrlFopen,
	},
}

type MooCLILicense struct {
	Token   string
	License string
}

// Hardcoded
var MOOCLOUD_LICENSE_ACTIVATION_SERVER_URL string = "https://apps-store.moocloud.ch/version-test/api/1.1/wf/activate"
var MOOCLOUD_MOOCLI_CONFIG_SERVER_URL string = "https://alpha.worker.tools.moocloud.ch/webhook/02b4adb1-1cfd-47b3-aedd-3db0ac47b3aa"
var MOOCLOUD_MOOCLI_CONFIG_UPDATE_LISTENING_PORT = 8080
var MOOCLOUD_MOOCLI_CONFIG_UPDATE_LISTENING_ENDPOINT = "config-update"

var ENV_CLOUDRON_APP_DOMAIN string
var ENV_CLOUDRON_API_ORIGIN string
var ENV_FILEPATH_MOOCLI_LICENSE string

var mooCLILicense MooCLILicense

var phpReturned chan bool
var nginxReturned chan bool

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func licenseCheck(isMatch bool) {
	if !isMatch {
		fmt.Println("The License doesn't seem to be valid. Please contact MooCloud.")
		os.Exit(0)
	}
}

func setupEnvironmentVariables() {

	// TODO: There's some testing code in here that
	// needs to be removed after everything works.
	// TODO: Think of a better testing functionality.

	var err error
	var ok bool

	ENV_FILEPATH_MOOCLI_LICENSE, ok = os.LookupEnv("FILEPATH_MOOCLI_LICENSE")

	if !ok {
		ENV_FILEPATH_MOOCLI_LICENSE, err = os.Getwd()
		check(err)
		ENV_FILEPATH_MOOCLI_LICENSE = ENV_FILEPATH_MOOCLI_LICENSE + "/license"
	}

	ENV_CLOUDRON_APP_DOMAIN, ok = os.LookupEnv("CLOUDRON_APP_DOMAIN")

	if !ok {
		ENV_CLOUDRON_APP_DOMAIN = "my.tools.moocloud.ch"
	}

	ENV_CLOUDRON_API_ORIGIN, ok = os.LookupEnv("CLOUDRON_API_ORIGIN")

	if !ok {
		ENV_CLOUDRON_API_ORIGIN = "https://swagger.tools.moocloud.ch"
	}

	ENV_CLOUDRON_API_ORIGIN = strings.TrimPrefix(ENV_CLOUDRON_API_ORIGIN, "https://")
}

func initMooCLILicense() {

	mooCLILicenseFilePointer, err := ioutil.ReadFile(ENV_FILEPATH_MOOCLI_LICENSE)

	check(err)

	mooCLILicenseFileStr := string(mooCLILicenseFilePointer)

	isMatch, err := regexp.Match("Token: .*\\n", []byte(mooCLILicenseFileStr))
	check(err)
	licenseCheck(isMatch)

	isMatch, err = regexp.Match("License: .*\\n", []byte(mooCLILicenseFileStr))
	check(err)
	licenseCheck(isMatch)

	re1 := regexp.MustCompile("\\n(.*)")
	re2 := regexp.MustCompile("(.*)\\n(.*)License: ")

	token := strings.Replace(mooCLILicenseFileStr, "Token: ", "", 1)
	token = re1.ReplaceAllString(token, "")

	license := re2.ReplaceAllString(mooCLILicenseFileStr, "")
	license = re1.ReplaceAllString(license, "")

	mooCLILicense.Token = token
	mooCLILicense.License = license
}

func requestMooCLIConfig(mooCLILicense MooCLILicense) []byte {
	requestBody, err := json.Marshal(map[string]string{
		"Token":   mooCLILicense.Token,
		"License": mooCLILicense.License,
	})

	check(err)

	request, err := http.NewRequest("GET", MOOCLOUD_MOOCLI_CONFIG_SERVER_URL, bytes.NewBuffer(requestBody))

	check(err)

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)

	check(err)

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	return body
}

func parseMooCLIConfig(body []byte) {
	var data map[string]map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	for app, options := range data {
		for option, value := range options {
			appMap, doesExist := mooCLIConfigParserMap[app]
			if doesExist {
				optionMap, doesExist := appMap[option]
				if doesExist {
					valueStr := fmt.Sprintf("%v", value)
					err := optionMap.(func(string) error)(valueStr)
					if err != nil {
						log.Print(err)
					}
				}

			}

		}

	}
}

func nginxStartGo(returned chan bool) {
	fmt.Println("Started running nginx")
	nginxStart := exec.Command("/usr/sbin/nginx", "-c", os.Getenv("DIRPATH_NGINX")+"/nginx.conf")
	nginxStart.Run()
	fmt.Println("Stopped running nginx")
	returned <- true
}

func nginxStopGo(stopped chan bool) {
	nginxStop := exec.Command("/usr/sbin/nginx", "-c", os.Getenv("DIRPATH_NGINX")+"/nginx.conf", "-s", "stop")
	nginxStop.Run()
	stopped <- true
}

func phpStartGo(returned chan bool) {
	fmt.Println("Started running php")
	phpStart := exec.Command("/usr/sbin/php-fpm"+os.Getenv("PHP_VERSION"), "--nodaemonize", "--fpm-config", os.Getenv("DIRPATH_PHP")+"/php-fpm.conf")
	phpStart.Run()
	fmt.Println("Stopped running php")
	returned <- true
}

func phpStopGo(stopped chan bool) {
	phpStop := exec.Command("service", "php"+os.Getenv("PHP_VERSION")+"-fpm", "stop")
	phpStop.Run()
	stopped <- true
}

func restartApps(n int) {
AppStoppLoop:
	for i := 0; true; i++ {

		if i >= n {
			log.Printf("Unsuccessfully tried %v times to stop the apps, please check for bugs.\n", n)
			os.Exit(1)
		}

		phpStopped := make(chan bool)
		nginxStopped := make(chan bool)

		go phpStopGo(phpStopped)
		go nginxStopGo(nginxStopped)

		<-phpStopped
		<-nginxStopped

		select {
		case <-phpReturned:
			select {
			case <-nginxReturned:
				break AppStoppLoop
			default:
			}
		default:
		}

		time.Sleep(2 * time.Second)
	}

	phpReturned = make(chan bool)
	nginxReturned = make(chan bool)

	go phpStartGo(phpReturned)
	go nginxStartGo(nginxReturned)
}

func responseToUpdateRequest(response http.ResponseWriter, request *http.Request) {
	var err error
	var body []byte
	switch request.Method {
	case http.MethodGet:
		body = requestMooCLIConfig(mooCLILicense)
	case http.MethodPost:
		body, err = ioutil.ReadAll(request.Body)
		if err != nil {
			check(err)
		}
	default:
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
	}
	parseMooCLIConfig(body)
	restartApps(10)
	response.WriteHeader(http.StatusOK)
}

func listenForUpdates() error {
	http.HandleFunc("/config-update", responseToUpdateRequest)
	return http.ListenAndServe("localhost:8080", nil)
}

func routineLoop() {
	listenError := listenForUpdates()
	check(listenError)
	fmt.Println("Shouldn't have come so far, but oh well.")
	os.Exit(1)
}

func main() {

	setupEnvironmentVariables()

	if _, err := os.Stat(ENV_FILEPATH_MOOCLI_LICENSE); os.IsNotExist(err) {

		fmt.Println("\nNo MooCLI License file found. Request from server." + "\n")

		requestStr, err := json.Marshal(map[string]string{
			"Domain": string(ENV_CLOUDRON_APP_DOMAIN),
			"server": string(ENV_CLOUDRON_API_ORIGIN),
		})

		check(err)

		req, err := http.NewRequest("POST", MOOCLOUD_LICENSE_ACTIVATION_SERVER_URL, bytes.NewBuffer(requestStr))

		check(err)

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		check(err)

		defer resp.Body.Close()

		if resp.Status != "200 OK" {
			fmt.Println("Could not generate a valid license file. Please note that this application is designed to be run within the MooCloud ecosystem.")
			os.Exit(0)
		}

		/*
			fmt.Println("request: ", req)
			fmt.Println()
			fmt.Println("response Status: " + resp.Status + "\n")
			fmt.Println("response Headers: ", resp.Header)
			fmt.Println()
		*/

		body, _ := ioutil.ReadAll(resp.Body)

		//fmt.Println("response Body: " + string(body) + "\n")

		var data map[string]map[string]string

		json.Unmarshal([]byte(body), &data)

		map_token := data["response"]["Token"]
		map_license := data["response"]["License"]

		/*
			fmt.Println("response Body Status11: ", dat)
			fmt.Println("response License", map_license)
			fmt.Println("response Token", map_token)
			fmt.Println()
		*/

		var licenseString string
		licenseString += "Token: " + map_token + "\n"
		licenseString += "License: " + map_license + "\n"

		mooCLILicenseFilePointer, err := os.Create(ENV_FILEPATH_MOOCLI_LICENSE)
		check(err)

		mooCLILicenseFilePointer.WriteString(licenseString)
		mooCLILicenseFilePointer.Close()

		fmt.Print("\n\nMooCLI License successfully obtained.\n\n")
	} else {
		fmt.Print("\n\nMooCLI License found.\n\n")
	}

	initMooCLILicense()

	var body []byte = requestMooCLIConfig(mooCLILicense)

	parseMooCLIConfig(body)

	phpReturned = make(chan bool)
	nginxReturned = make(chan bool)

	go phpStartGo(phpReturned)
	go nginxStartGo(nginxReturned)

	routineLoop()
}
