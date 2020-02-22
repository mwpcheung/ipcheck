package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/yl2chen/cidranger"
)

var (
	gChinaMainlandRange cidranger.Ranger
)

func loadLookupTable(name string) cidranger.Ranger {
	ranger := cidranger.NewPCTrieRanger()
	fi, err := os.Open(name)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		_, network, _ := net.ParseCIDR(string(a))
		entry := cidranger.NewBasicRangerEntry(*network)
		ranger.Insert(entry)
	}
	return ranger
}

func isChinaMainlandIP(IP string) bool {
	ipp := net.ParseIP(IP)
	if gChinaMainlandRange != nil {
		contains, err := gChinaMainlandRange.Contains(ipp)
		if err != nil {
			fmt.Printf("to query ip is  %s failed %v", IP, err)
			return false
		}
		return contains
	}
	return false
}
func requestGet(uri string) []byte {
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Printf("sending Rio server request failed %s", err.Error())
		return nil
	}
	defer resp.Body.Close()
	nowResp, _ := ioutil.ReadAll(resp.Body)
	return nowResp
}
func downloadIPTable() []byte {
	uri := "https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt"
	body := requestGet(uri)
	return body
}

func handleQueryChinaMainland(ctx *context.Context) {
	ip := ctx.Input.Query("ip")
	if ip != "" {
		if strings.Contains(ip, ":") {
			ip = strings.Split(ip, ":")[1]
		}
	}
	isChina := isChinaMainlandIP(ip)
	if isChina {
		ctx.WriteString("yes")
	} else {
		ctx.WriteString("no")
	}

}

//Exists file exist
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
func initLookupTable() {
	name := "iptable.txt"
	if !Exists(name) {
		ips := downloadIPTable()
		if len(ips) > 0 {
			ioutil.WriteFile(name, ips, 0644)
		}
	}
	gChinaMainlandRange = loadLookupTable(name)
}
func main() {
	initLookupTable()
	beego.Any("/api", handleQueryChinaMainland)
	beego.Run()
}
