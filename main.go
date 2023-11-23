//go:build windows
// +build windows

package main

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"syscall"
	"time"
)

var (
	username   string
	password   string
	interval   time.Duration
	autoLogin  bool
	iface      string
	configFile string
	client     http.Client
)

//go:embed index.html
var indexHTML []byte

type Config struct {
	Username  string
	Password  string
	Interval  time.Duration
	AutoLogin bool
	Iface     string
}

func main() {
	loadConfig()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	configFile = homeDir + "\\Documents\\autologin\\config.dat"
	fmt.Println("配置文件路径：", configFile)
	// 如果没有相关文件夹，则创建
	_, err = os.Stat(homeDir + "\\Documents\\autologin")
	if os.IsNotExist(err) {
		err = os.Mkdir(homeDir+"\\Documents\\autologin", os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	client = http.Client{
		Timeout: 3 * time.Second,
	}

	runtime.LockOSThread()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)

	go func() {
		for {
			if autoLogin && !checkLogin() {
				login()
			}
			time.Sleep(interval)
		}
	}()

	fmt.Println("\033[31m已启动自动登录程序，请在浏览器打开http://localhost:1580 进行配置，\nPowered by Tianli 2023 For SZCU\033[0m")
	// 如果没有config.dat文件不在本地，则自动打开浏览器
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		cmd := exec.Command("cmd", "/c", "start", "http://localhost:1580")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = http.ListenAndServe(":1580", nil)
	if err != nil {
		// 打开浏览器
		cmd := exec.Command("cmd", "/c", "start", "http://localhost:1580")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	saveConfig()
}

// 遍历网卡列表，返回IP为20开头的网卡名称
func getIface() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println(err)
			return ""
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				if ipNet.IP.String()[0:2] == "20" {
					return iface.Name
				}
			}
		}
	}

	return ""
}

func getIPAndMAC() (string, string) {
	resp, err := http.Get("http://172.16.8.22/")
	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}

	// 使用正则表达式提取IP地址
	var ipMatches []string
	var ip string
	ipRegex := regexp.MustCompile(`v46ip='([\d.]+)'`)
	ipMatches = ipRegex.FindStringSubmatch(string(body))
	if len(ipMatches) < 2 {
		ipRegex := regexp.MustCompile(`v4ip='([\d.]+)'`)
		ipMatches = ipRegex.FindStringSubmatch(string(body))
		if len(ipMatches) < 2 {
			fmt.Println("无法获取IP地址")
			return "", ""
		}
	}
	ip = ipMatches[1]
	fmt.Println("IP地址：", ip)

	mac := "000000000000"

	return ip, mac
}

func login() {
	ip, mac := getIPAndMAC()
	//请注意针对您的校园网修改登录请求
	macEncoded := url.QueryEscape(mac)
	loginURL := fmt.Sprintf("http://172.16.8.22:801/eportal/?c=Portal&a=login&callback=dr1004&login_method=1&user_account=%%2C0%%2C%s%%40telecom&user_password=%s&wlan_user_ip=%s&wlan_user_ipv6=&wlan_user_mac=%s&wlan_ac_ip=&wlan_ac_name=&jsVersion=3.3.3&v=9431", username, password, ip, macEncoded)

	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "登录成功")
}

func checkLogin() bool {
	req, err := http.NewRequest("GET", "https://www.baidu.com/", nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginData struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Interval  string `json:"interval"`
		AutoLogin bool   `json:"autoLogin"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, "错误响应", http.StatusBadRequest)
		return
	}

	username = loginData.Username
	password = loginData.Password
	intervalStr := loginData.Interval
	autoLogin = loginData.AutoLogin

	if iface == "" {
		iface = getIface()
	}

	fmt.Fprint(w, "登录信息已配置")

	if intervalStr != "" {
		duration, err := time.ParseDuration(intervalStr + "s")
		if err != nil {
			fmt.Println(err)
			return
		}
		interval = duration
	}

	saveConfig() // 保存配置项到文件

	if autoLogin {
		login()
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	ip, mac := getIPAndMAC()

	var status string
	if checkLogin() {
		status = "已经登录"
	} else {
		status = "还没有登录"
	}

	data := struct {
		Status    string
		Interval  string
		AutoLogin bool
		IP        string
		MAC       string
	}{
		status,
		interval.String(),
		autoLogin,
		ip,
		mac,
	}

	tmpl, err := template.New("index").Parse(string(indexHTML))
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, "服务器出错")
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, "服务器出错")
		return
	}
}

func saveConfig() {
	var config Config
	config.Username = username
	config.Password = password
	config.Interval = interval
	config.AutoLogin = autoLogin

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	configFile, err := os.Create(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer configFile.Close()

	_, err = configFile.Write(buf.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func loadConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	configFile = homeDir + "\\Documents\\autologin\\config.dat"
	fmt.Println("配置文件路径：", configFile)

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		// 如果配置文件不存在，则使用默认配置
		username = "default_username"
		password = "default_password"
		interval = 10 * time.Second
		autoLogin = false
		iface = "Ethernet"
		return
	}

	configFile, err := os.Open(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer configFile.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	if buf.Len() == 0 {
		// 如果配置文件为空，则使用默认配置
		username = "default_username"
		password = "default_password"
		interval = 10 * time.Second
		autoLogin = false
		return
	}

	var config Config
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&config)
	if err != nil {
		fmt.Println(err)
		return
	}

	username = config.Username
	password = config.Password
	interval = config.Interval
	autoLogin = config.AutoLogin
}
