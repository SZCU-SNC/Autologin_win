//go:build windows
// +build windows

package main

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"text/template"
	"time"

	"golang.org/x/sys/windows/registry"
)

var (
	username   string
	password   string
	interval   time.Duration
	autoLogin  bool
	iface      string
	configFile string = "config.dat"
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

func getIPAndMAC(iface string) (string, string) {
	ifaceObj, err := net.InterfaceByName(iface)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}

	addrs, err := ifaceObj.Addrs()
	if err != nil {
		fmt.Println(err)
		return "", ""
	}

	var ip string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ip = ipNet.IP.String()
			break
		}
	}

	mac := ifaceObj.HardwareAddr.String()

	return ip, mac
}

func login() {
	ip, mac := getIPAndMAC(iface)
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

	username = r.FormValue("username")
	password = r.FormValue("password")
	intervalStr := r.FormValue("interval")
	autoLogin = r.FormValue("auto_login") == "1"
	iface = r.FormValue("iface")

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
	ip, mac := getIPAndMAC(iface)

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
		Iface     string
		IP        string
		MAC       string
	}{
		status,
		interval.String(),
		autoLogin,
		iface,
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
	config.Iface = iface

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
	_, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		fmt.Println(err)
	}

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
		iface = "Ethernet"
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
	iface = config.Iface
}

func listInterfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, iface := range ifaces {
		fmt.Println(iface.Name)
	}
}

func main() {
	loadConfig()

	client = http.Client{
		Timeout: 3 * time.Second,
	}

	listInterfaces()
	runtime.LockOSThread()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	// 列出网卡列表，list函数
	http.HandleFunc("/list", listInterfacesHandler)

	go func() {
		for {
			if autoLogin && !checkLogin() {
				login()
			}
			time.Sleep(interval)
		}
	}()

	fmt.Println("\033[31m网卡名称请参考以上列表进行配置测试\n已启动自动登录程序，请在浏览器打开http://localhost:1580 进行配置，后续使用及配置也请在此页面进行\nPowered by Tianli 2023 For SZCU\033[0m")
	// 如果没有config.dat文件不在本地，则自动打开浏览器
	_, err := os.Stat(configFile)
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

// 列出网卡列表，以json格式返回
func listInterfacesHandler(w http.ResponseWriter, _ *http.Request) {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return
	}

	var buf bytes.Buffer
	buf.WriteString("[")
	for i, iface := range ifaces {
		buf.WriteString(fmt.Sprintf("\"%s\"", iface.Name))
		if i != len(ifaces)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}
