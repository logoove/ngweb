package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

//go:embed tpl/*
var tpls embed.FS
type ViewFunc func(http.ResponseWriter, *http.Request)
func BasicAuth(f ViewFunc, user, passwd []byte) ViewFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		basicAuthPrefix := "Basic "
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			payload, err := base64.StdEncoding.DecodeString(
				auth[len(basicAuthPrefix):],
			)
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], passwd) {
					f(w, r)
					return
				}
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="ngweb"`)
		w.WriteHeader(http.StatusUnauthorized)
	}
}
func main() {

	fmt.Println("linux端口进程管理\n http://localhost:8088"+"\n Ver:1.0.0 \n首次使用[ufw allow 8088]开启本应用端口默认账号密码(admin/123456)\n 作者:yoby \n date:"+time.Now().Format("2006-01-02 15:04:05"))
	http.HandleFunc("/", BasicAuth(home, []byte("admin"), []byte("123456")))
	http.HandleFunc("/getlist", getlist)
	http.HandleFunc("/add", add)
	http.HandleFunc("/del", del)
	http.HandleFunc("/delpid", delpid)
	http.HandleFunc("/so", so)//查询端口对应pid
	http.HandleFunc("/ufwa", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "*")
		c:=exec.Command("ufw", "enable")
		str,_:=c.Output()
		msg, _ := json.Marshal(map[string]interface{}{"code":200,"msg":string(str)})
		io.WriteString(w, string(msg))
	})
	http.HandleFunc("/ufwb", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "*")
		c:=exec.Command("ufw", "disable")
		str,_:=c.Output()
		msg, _ := json.Marshal(map[string]interface{}{"code":200,"msg":string(str)})
		io.WriteString(w, string(msg))

	})
	http.Handle("/tpl/",http.FileServer(http.FS(tpls)))
	http.ListenAndServe(":8088", nil)
}
func getlist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	c:=exec.Command("ufw","status")
	str,_:=c.Output()
	var lines []string
	buff := bufio.NewScanner(strings.NewReader(string(str)))
	for buff.Scan() {
		re:=regexp.MustCompile(`\s+`).Split(strings.TrimRight(buff.Text()," "),-1)
		lines = append(lines,re[0])
	}
	//fmt.Println(lines)
	var mp map[string]interface{}
	if len(lines) >1{
		sc:=lines[4:len(lines)-1]
		mp=map[string]interface{}{"code":200,"data":array_unique(sc),"msg":"读取端口成功"}
	}else{
		mp=map[string]interface{}{"code":400,"msg":"没找到"}
	}
	msg, _ := json.Marshal(mp)
	io.WriteString(w, string(msg))
	//io.WriteString(w,string(`{"code":200,"data":["80/tcp","53"],"msg":"读取端口成功"}`))
}
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.SetBasicAuth("admin","123456")
	t := template.Must(template.ParseFS(tpls, "tpl/*.html"))
	t.Execute(w, "你好")
}
func add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	k:=r.FormValue("k")
	c:=exec.Command("ufw", "allow",k)
	str,_:=c.Output()
	msg, _ := json.Marshal(map[string]interface{}{"code":200,"msg":string(str)})
	io.WriteString(w, string(msg))
}
func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	k:=r.FormValue("k")
	c:=exec.Command("ufw","delete", "allow",k)
	str,_:=c.Output()
	msg, _ := json.Marshal(map[string]interface{}{"code":200,"msg":string(str)})
	io.WriteString(w, string(msg))
}
func so(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	k:=r.FormValue("k")
	c:=exec.Command("lsof", "-i:"+k)
	str,_:=c.Output()
	ms:=make(map[string]string)
	buff := bufio.NewScanner(strings.NewReader(string(str)))
	for buff.Scan() {
		re:=regexp.MustCompile(`\s+`).Split(strings.TrimRight(buff.Text()," "),-1)
		ms[re[1]]=re[0]
	}
	delete(ms,"PID")
	mp:=map[string]interface{}{"code":200,"data":ms,"msg":"查询pid成功"}
	msg, _ := json.Marshal(mp)
	io.WriteString(w, string(msg))
}
func delpid(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	k:=r.FormValue("k")
	c:=exec.Command("kill",k)
	str,_:=c.Output()
	msg, _ := json.Marshal(map[string]interface{}{"code":200,"msg":string(str)})
	io.WriteString(w, string(msg))
}
func array_unique(l []string) []string {
	result := make([]string, 0, len(l))
	temp := map[string]struct{}{}
	for _, item := range l {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
