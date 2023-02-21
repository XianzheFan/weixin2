package wxapi

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var (
	// APPID 公众号后台分配的APPID
	APPID = "wxecfcee9d7fd5785e"
	// APPSECRET 公众号后台分配的APPSECRET
	APPSECRET = "fa1dc9e9fe08de5cba8a645b66aa34d0"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// WxReceiveCommonMsg 接收普通消息
type WxReceiveCommonMsg struct {
	ToUserName   string //接收者 开发者 微信号
	FromUserName string //发送者 发送方帐号（一个OpenID）
	Content      string //文本内容
	CreateTime   int64  //创建时间
	MsgType      string //消息类型
	MsgId        int64  //消息id
	PicUrl       string //图片url
	MediaId      string //媒体id
	Event        string //事件类型，VIEW
	EventKey     string //事件KEY值，设置的跳转URL
	MenuId       string
	Format       string
	Recognition  string
	ThumbMediaId string //缩略图媒体ID
}

type WxCustomText struct {
	Content string `json:"content"`
}
type WxCustomTextMsg struct {
	ToUser string `json:"touser"`
	MsgType string `json:"msgtype"`
	Text WxCustomText `json:"text"`
}

func (msg *WxCustomTextMsg) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

// WxReceiveFunc (接收到消息之后，会将消息交于这个函数处理)
var WxReceiveFunc func(msg WxReceiveCommonMsg) error

// WxMakeSign 计算签名
func WxMakeSign(token, timestamp, nonce string) string {
	strs := []string{token, timestamp, nonce}
	sort.Strings(strs)
	sha := sha1.New()
	io.WriteString(sha, strings.Join(strs, ""))
	return fmt.Sprintf("%x", sha.Sum(nil))
}

// HandleWxLogin首次接入，成为开发者
func HandleWxLogin(c *gin.Context) {
	fmt.Printf("==>HandleWxLogin\n")
	echostr := c.DefaultQuery("echostr", "")
	if echostr != "" {
		fmt.Printf("==>echostr:%s\n", echostr)
		c.String(200, "%s", echostr)
		return
	}
}

// WxGetAccessToken 获取微信accesstoken
func WxGetAccessToken() string {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", APPID, APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}

	return token.AccessToken
}

// WxGetUserList 获取关注者列表
func WxGetUserList(accessToken string) []gjson.Result {
	url := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + accessToken + "&next_openid="
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取关注列表失败", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return nil
	}
	flist := gjson.Get(string(body), "data.openid").Array()
	return flist
}

// WxPostTemplate 发送模板消息
func WxPostTemplate(accessToken string, reqdata string, fxurl string, templateid string, openid string) {

	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + accessToken

	reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateid + "\", \"url\":\"" + fxurl + "\", \"data\": " + reqdata + "}"
	fmt.Printf("WxPostTemplate:%#v\n", reqbody)
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}

// 客服回复接口
func WxPostCustomTextMsg(accessToken string, touser string, content string) {

	url := "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken

	req:=&WxCustomTextMsg{ToUser:touser, MsgType:"text", Text:WxCustomText{Content:content}}
	jsonStr := req.ToJson()
	//fmt.Printf("WxPostCustomTextMsg:%#v\n", jsonStr)
	request ,_:= http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}


// 客服输入状态接口
func InputStatusRequest(accessToken string, touser string, command string) error {
    url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/typing?access_token=%s", accessToken)
    data := fmt.Sprintf(`{"touser":"%s", "command":"%s"}`, touser, command)

    resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer([]byte(data)))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Check response status code
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("InputStatusRequest request failed with status code %d", resp.StatusCode)
    }

    return nil
}


// ReceiveCommonMsg
func ReceiveCommonMsg(msgData []byte) (WxReceiveCommonMsg, error) {

	fmt.Printf("received weixin msgData:\n%s\n", msgData)
	msg := WxReceiveCommonMsg{}
	err := xml.Unmarshal(msgData, &msg)
	if WxReceiveFunc == nil {
		return msg, err
	}
	err = WxReceiveFunc(msg)
	return msg, err
}

// HandleWxPostRecv 处理微信公众号前端发起的消息事件
func HandleWxPostRecv(c *gin.Context) {
	fmt.Printf("==>HandleWxPostRecv Enter\n")
	data, err := c.GetRawData()
	if err != nil {
		log.Fatalln(err)
	}
	ReceiveCommonMsg(data)
}

// WxCreateMenu 创建菜单
func WxCreateMenu(accessToken, menustr string) (string, error) {

	url := "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + accessToken
	fmt.Printf("WxCreateMenu:%s\n", menustr)
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(menustr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}

// WxDelMenu 删除菜单
func WxDelMenu(accessToken string) (string, error) {
	url := "https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=" + accessToken
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("删除菜单失败", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}

// WxGetUserInfo 根据用户openid获取基本信息
func WxGetUserInfo(accessToken, openid string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN", accessToken, openid)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取信息失败", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil

}


func init() {
}
