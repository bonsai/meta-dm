package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Message struct {
	ID string `json:"id"`
	ConversationID string `json:"conversation_id"`
	From json.RawMessage `json:"from,omitempty"`
	To json.RawMessage `json:"to,omitempty"`
	Text string `json:"text,omitempty"`
	Attachments json.RawMessage `json:"attachments,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Source string `json:"source"`
	EventType string `json:"event_type"`
}
type page struct {
	Data []map[string]any `json:"data"`
	Paging struct { Next string `json:"next"` } `json:"paging"`
}

func main() {
	if len(os.Args)<2 { usage(); os.Exit(2) }
	var err error
	switch os.Args[1] {
	case "conversations": err=conversations(os.Args[2:])
	case "history": err=history(os.Args[2:])
	case "workflow": err=workflow(os.Args[2:])
	default: usage(); os.Exit(2)
	}
	if err!=nil { fmt.Fprintln(os.Stderr,"meta-dm:",err); os.Exit(1) }
}
func usage(){fmt.Println("meta-dm conversations");fmt.Println("meta-dm history <conversation_id> [--limit N] [--before ISO8601] [--all]");fmt.Println("meta-dm workflow run <conversation_id> [--before ISO8601] [--all]");fmt.Println("meta-dm workflow runs");fmt.Println("meta-dm workflow download <run_id> [--dir DIR]")}
func gh(args ...string) error {
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func workflow(args []string) error {
	if len(args) == 0 { return errors.New("workflow subcommand is required: run, runs, download") }
	switch args[0] {
	case "run":
		if len(args) < 2 { return errors.New("conversation_id is required") }
		id := args[1]
		fs := flag.NewFlagSet("workflow run", flag.ContinueOnError)
		before := fs.String("before", "", "optional RFC3339 cutoff")
		all := fs.Bool("all", true, "follow pagination toward older messages")
		if err := fs.Parse(args[2:]); err != nil { return err }
		cmdArgs := []string{"workflow", "run", ".github/workflows/dm-history.yml", "-f", "conversation_id="+id, "-f", "all="+strconv.FormatBool(*all)}
		if *before != "" { cmdArgs = append(cmdArgs, "-f", "before="+*before) }
		fmt.Printf("dispatching dm-history.yml for conversation=%s\n", id)
		return gh(cmdArgs...)
	case "runs":
		return gh("run", "list", "--workflow", "dm-history.yml", "--limit", "10")
	case "download":
		if len(args) < 2 { return errors.New("run_id is required") }
		fs := flag.NewFlagSet("workflow download", flag.ContinueOnError)
		dir := fs.String("dir", "private-archive", "download directory")
		if err := fs.Parse(args[2:]); err != nil { return err }
		return gh("run", "download", args[1], "-D", *dir)
	default:
		return errors.New("unknown workflow subcommand: " + args[0])
	}
}

func token()(string,error){v:=strings.TrimSpace(os.Getenv("META_ACCESS_TOKEN"));if v==""{return "",errors.New("META_ACCESS_TOKEN is required")};return v,nil}
func base()string{v:=strings.TrimRight(os.Getenv("META_GRAPH_URL"),"/");if v==""{v="https://graph.facebook.com"};return v}
func version()string{v:=strings.Trim(os.Getenv("META_GRAPH_VERSION"),"/ ");if v==""{v="v23.0"};return v}
func get(raw string)([]byte,error){
	t,err:=token();if err!=nil{return nil,err};u,err:=url.Parse(raw);if err!=nil{return nil,err}
	q:=u.Query();q.Set("access_token",t);u.RawQuery=q.Encode()
	req,err:=http.NewRequest(http.MethodGet,u.String(),nil);if err!=nil{return nil,err}
	req.Header.Set("Accept","application/json");resp,err:=http.DefaultClient.Do(req);if err!=nil{return nil,err};defer resp.Body.Close()
	b,_:=io.ReadAll(resp.Body);if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("Meta API %s: %s",resp.Status,strings.TrimSpace(string(b)))};return b,nil
}
func conversations(args []string)error{
	fs:=flag.NewFlagSet("conversations",flag.ContinueOnError);limit:=fs.Int("limit",100,"page size");if err:=fs.Parse(args);err!=nil{return err}
	ig:=strings.TrimSpace(os.Getenv("META_IG_USER_ID"));if ig==""{return errors.New("META_IG_USER_ID is required")}
	u:=fmt.Sprintf("%s/%s/%s/conversations",base(),version(),url.PathEscape(ig));q:=url.Values{};q.Set("platform","instagram");q.Set("limit",fmt.Sprint(*limit));u+="?"+q.Encode()
	b,err:=get(u);if err!=nil{return err};var p page;if err=json.Unmarshal(b,&p);err!=nil{return err};return json.NewEncoder(os.Stdout).Encode(p.Data)
}
func history(args []string)error{
	if len(args)==0{return errors.New("conversation_id is required")};id:=args[0]
	fs:=flag.NewFlagSet("history",flag.ContinueOnError);limit:=fs.Int("limit",0,"maximum messages");before:=fs.String("before","","stop at RFC3339");all:=fs.Bool("all",false,"follow all older pages");pageSize:=fs.Int("page-size",100,"Meta API page size");out:=fs.String("out","data/messages","output directory");if err:=fs.Parse(args[1:]);err!=nil{return err}
	if *before!=""{if _,err:=time.Parse(time.RFC3339,*before);err!=nil{return err}}
	path:=filepath.Join(*out,id+".jsonl");seen,err:=readSeen(path);if err!=nil{return err}
	next:=fmt.Sprintf("%s/%s/%s/messages",base(),version(),url.PathEscape(id));q:=url.Values{};q.Set("fields","id,created_time,from,to,message,attachments");q.Set("limit",fmt.Sprint(*pageSize));next+="?"+q.Encode()
	var added []Message
	for next!=""{
		b,err:=get(next);if err!=nil{return err};var p page;if err=json.Unmarshal(b,&p);err!=nil{return err};stop:=false
		for _,r:=range p.Data{
			m:=normalize(id,r);if m.ID==""{continue}
			if *before!=""&&m.CreatedAt!=""{if t,e:=time.Parse(time.RFC3339,m.CreatedAt);e==nil{cut,_:=time.Parse(time.RFC3339,*before);if !t.After(cut){stop=true;continue}}}
			if _,ok:=seen[m.ID];ok{continue};seen[m.ID]=struct{}{};added=append(added,m)
			if *limit>0&&len(added)>=*limit{stop=true;break}
		}
		if stop||!*all{break};next=p.Paging.Next
	}
	sort.Slice(added,func(i,j int)bool{return added[i].CreatedAt<added[j].CreatedAt})
	if err:=os.MkdirAll(*out,0700);err!=nil{return err};f,err:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600);if err!=nil{return err};defer f.Close();enc:=json.NewEncoder(f);for _,m:=range added{if err:=enc.Encode(m);err!=nil{return err}}
	if err:=os.MkdirAll("data/sync",0700);err!=nil{return err};state:=map[string]any{"conversation_id":id,"retrieved_at":time.Now().UTC().Format(time.RFC3339),"added":len(added),"total_known":len(seen),"mode_all":*all,"before":*before};sb,_:=json.MarshalIndent(state,"","  ");if err:=os.WriteFile(filepath.Join("data/sync",id+".json"),append(sb,'\n'),0600);err!=nil{return err}
	fmt.Printf("conversation=%s added=%d known=%d file=%s\n",id,len(added),len(seen),path);return nil
}
func readSeen(path string)(map[string]struct{},error){seen:=map[string]struct{}{};f,err:=os.Open(path);if errors.Is(err,os.ErrNotExist){return seen,nil};if err!=nil{return nil,err};defer f.Close();s:=bufio.NewScanner(f);for s.Scan(){var m Message;if json.Unmarshal(s.Bytes(),&m)==nil&&m.ID!=""{seen[m.ID]=struct{}{}}};return seen,s.Err()}
func normalize(id string,r map[string]any)Message{m:=Message{ConversationID:id,Source:"instagram",EventType:"message_received"};m.ID,_=r["id"].(string);m.CreatedAt,_=r["created_time"].(string);m.Text,_=r["message"].(string);if v,ok:=r["from"];ok{m.From,_=json.Marshal(v)};if v,ok:=r["to"];ok{m.To,_=json.Marshal(v)};if v,ok:=r["attachments"];ok{m.Attachments,_=json.Marshal(v)};return m}
