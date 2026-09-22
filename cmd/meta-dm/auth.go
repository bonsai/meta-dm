package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type savedToken struct {
	AccessToken string `json:"access_token"`
	UserID string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	ExpiresIn int64 `json:"expires_in,omitempty"`
	CreatedAt string `json:"created_at"`
}

func login(args []string) error {
	fs:=flag.NewFlagSet("login",flag.ContinueOnError)
	redirect:=fs.String("redirect",envOr("META_REDIRECT_URI","http://127.0.0.1:8787/callback"),"OAuth redirect URI")
	scope:=fs.String("scope",envOr("META_SCOPE","instagram_business_basic,instagram_business_manage_messages"),"Instagram Login scopes")
	if err:=fs.Parse(args);err!=nil{return err}
	appID:=strings.TrimSpace(os.Getenv("META_APP_ID"));appSecret:=strings.TrimSpace(os.Getenv("META_APP_SECRET"))
	if appID==""||appSecret==""{return errors.New("META_APP_ID and META_APP_SECRET are required")}
	u,err:=url.Parse(*redirect);if err!=nil{return err}
	if u.Hostname()!="127.0.0.1"&&u.Hostname()!="localhost"{return errors.New("redirect must use 127.0.0.1 or localhost")}
	state:=randomState()
	authBase:=strings.TrimRight(envOr("META_AUTH_URL","https://www.instagram.com"),"/")
	authURL:=authBase+"/oauth/authorize?"+url.Values{"client_id":{appID},"redirect_uri":{*redirect},"response_type":{"code"},"scope":{*scope},"state":{state}}.Encode()
	fmt.Printf("Open this URL in the account owner's browser:\n\n    %s\n\n",authURL);openBrowser(authURL)
	code,err:=waitCallback(u,state);if err!=nil{return err}
	t:=savedToken{}
	form:=url.Values{"client_id":{appID},"client_secret":{appSecret},"grant_type":{"authorization_code"},"redirect_uri":{*redirect},"code":{code}}
	resp,err:=http.PostForm("https://api.instagram.com/oauth/access_token",form);if err!=nil{return err};defer resp.Body.Close()
	body,_:=io.ReadAll(resp.Body);if resp.StatusCode<200||resp.StatusCode>=300{return fmt.Errorf("Instagram token exchange %s: %s",resp.Status,strings.TrimSpace(string(body)))}
	if err:=json.Unmarshal(body,&t);err!=nil{return fmt.Errorf("decode token response: %w",err)};if t.AccessToken==""{return errors.New("token response has no access_token")}
	t.CreatedAt=time.Now().UTC().Format(time.RFC3339)
	if t.UserID==""{if id,name,e:=instagramMe(t.AccessToken);e==nil{t.UserID=id;t.Username=name}}
	if err:=os.MkdirAll(".meta-dm",0700);err!=nil{return err}
	b,_:=json.MarshalIndent(t,"","  ");if err:=os.WriteFile(filepath.Join(".meta-dm","token.json"),append(b,'\n'),0600);err!=nil{return err}
	fmt.Printf("login ok: user_id=%s username=%s\n",t.UserID,t.Username);fmt.Println("saved: .meta-dm/token.json (0600)")
	return nil
}

func waitCallback(redirect *url.URL,state string)(string,error){
	path:=redirect.Path;if path==""{path="/"}
	ch:=make(chan *url.URL,1)
	srv:=&http.Server{Addr:redirect.Host,Handler:http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.URL.Path!=path{http.NotFound(w,r);return}
		select{case ch<-r.URL:default:}
		w.Header().Set("Content-Type","text/plain; charset=utf-8");_,_=io.WriteString(w,"meta-dm login complete. You can close this tab.\n")
	})}
	go func(){_=srv.ListenAndServe()}();defer srv.Close()
	fmt.Println("Waiting for authorization callback on",redirect.Host,"...")
	select{case u:=<-ch:
		if e:=u.Query().Get("error");e!=""{return "",fmt.Errorf("Instagram authorization error: %s: %s",e,u.Query().Get("error_description"))}
		if u.Query().Get("state")!=state{return "",errors.New("OAuth state mismatch")}
		code:=u.Query().Get("code");if code==""{return "",errors.New("authorization code is missing")};return code,nil
	case <-time.After(5*time.Minute):return "",errors.New("OAuth callback timeout")}
}

func instagramMe(token string)(string,string,error){
	u:="https://graph.instagram.com/me?fields=id,username&access_token="+url.QueryEscape(token)
	resp,err:=http.Get(u);if err!=nil{return "","",err};defer resp.Body.Close();b,_:=io.ReadAll(resp.Body)
	if resp.StatusCode<200||resp.StatusCode>=300{return "","",fmt.Errorf("Instagram /me %s: %s",resp.Status,strings.TrimSpace(string(b)))}
	var v struct{ID string `json:"id"`;Username string `json:"username"`};if err:=json.Unmarshal(b,&v);err!=nil{return "","",err};return v.ID,v.Username,nil
}

func randomState()string{b:=make([]byte,16);if _,err:=rand.Read(b);err!=nil{return fmt.Sprint(time.Now().UnixNano())};return hex.EncodeToString(b)}
func envOr(k,d string)string{if v:=strings.TrimSpace(os.Getenv(k));v!=""{return v};return d}

func openBrowser(u string){
	var cmd *exec.Cmd
	switch runtime.GOOS{case "windows":cmd=exec.Command("cmd","/c","start","",u)
	case "darwin":cmd=exec.Command("open",u)
	default:cmd=exec.Command("xdg-open",u)}
	_=cmd.Start()
}
