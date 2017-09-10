// @Author: Neils <dengqi935@gmail.cm>
// @Describtion:
//     Login into ShanghaiTech's network
//     Update the dns of domain with the
//     api of clouddxns
// @Dependencies: Standard Library
// @Version: 0.1
// @Updated: Aug 2, 2017

package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/buger/jsonparser"
	"github.com/robfig/cron"
)

const api_base_url = "https://www.cloudxns.net/api2"

// const base_url = "https://controller.shanghaitech.edu.cn:8445/PortalServer/"
const base_url = "https://10.15.44.172:8445/PortalServer/"

//
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type UpdateData struct {
	Domain string `json:"domain"`
	Ip     string `json:"ip"`
}

type Config struct {
	Username  string
	Password  string
	ApiKey    string
	SecretKey string
	Domains   []string
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func Login(username, password string) (status bool, ip string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	loginData := url.Values{}
	loginData.Add("userName", username)
	loginData.Add("password", password)
	loginData.Add("hasValidateCode", "false")
	loginData.Add("authLan", "zh_CN")
	loginUrl := "/Webauth/webAuthAction!login.action"
	req, err := http.NewRequest("POST", base_url+loginUrl,
		strings.NewReader(loginData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	Check(err)
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	Check(err)
	status, err2 := jsonparser.GetBoolean(body, "success")
	if err2 == nil {
		if !status {
			ip = ""
			fmt.Println("Emmmmm, something wrong!")
			status = false
			return
		}
		ip2, err2 := jsonparser.GetString(body, "data", "ip")
		if err2 == nil {
			status = true
			ip = ip2
			return
		}
		fmt.Println(err)
		status = false
		ip = ""
		return
	}
	status = false
	ip = ""
	return
}

func UpdateDDNS(ip, domain, secret_key, api_key string) {
	client := http.Client{}
	request_url := api_base_url + "/ddns"
	data := url.Values{}
	data.Add("domian", domain)
	data.Add("ip", ip)
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	req_data := UpdateData{
		Domain: domain,
		Ip:     ip,
	}
	err := encoder.Encode(req_data)
	Check(err)
	req, err := http.NewRequest("POST", request_url, buf)
	Check(err)
	req.Header.Set("Content-Type", "application/json")
	// request_data := "{\"domain\":\"" + domain + "\",\"ip\":\"" + ip + "\",\"line_id\":\"1\"}"
	// request_data := buf.String()
	req.Header.Set("API-KEY", api_key)
	request_time := string(time.Now().Format(time.RFC1123Z))
	request_time = strings.Replace(request_time, "UTC", "GMT", 1)
	req.Header.Set("API-REQUEST-DATE", request_time)
	hmac_raw := api_key + request_url + buf.String() + request_time + secret_key
	hula := md5.Sum([]byte(hmac_raw))
	hmac := hex.EncodeToString(hula[:])
	req.Header.Set("API-HMAC", hmac)
	resp, err := client.Do(req)
	Check(err)
	// fmt.Println(hmac_raw)
	// fmt.Println(hmac)
	// fmt.Println(request_data)
	// fmt.Println(request_time)
	defer resp.Body.Close()
	r := Response{}
	json.NewDecoder(resp.Body).Decode(&r)
	if r.Code != 1 {
		fmt.Println(r.Message)
	}
	fmt.Println(domain + " : " + r.Message)
}

func ddns() {
	if len(os.Args) != 2 {
		fmt.Println("You need to specife a config file!")
		os.Exit(1)
	}

	file, err := ioutil.ReadFile(os.Args[1])
	Check(err)
	config := Config{}
	json.Unmarshal(file, &config)
	_, ip := Login(config.Username, config.Password)
	fmt.Println(time.Now())
	fmt.Println("Start")
	fmt.Println(ip)
	for _, domain := range config.Domains {
		//fmt.Println(domain)
		UpdateDDNS(ip, domain, config.SecretKey, config.ApiKey)
	}
	fmt.Println("Finished")
}

func handleSignal() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Kill, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	fmt.Println("Close singal recived!")
}

func main() {
	cronTab := cron.New()
	cronTab.AddFunc("@every 30m", ddns)
	cronTab.Start()
	fmt.Println("server started")
	ddns()

	handleSignal()
	fmt.Println("gracefully shutdown")
}
