package main

import (
	"fmt"
	"net/http"
	"weixin2/wxapi"
	_"weixin2/convert"
	"weixin2/chatapi"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
)

func handleTest(c *gin.Context) {
	fmt.Printf("==>handleTest Enter\n")
	//say, url := apis.GetSay()
	//fmt.Printf("GetSay:say=%#v,url=%s\n", say, url)
	//apis.SendEverydaySay()
	//apis.SendWeather()
	accessToken := wxapi.WxGetAccessToken()
	fmt.Printf("WxGetAccessToken:%s\n", accessToken)
	ulist := wxapi.WxGetUserList(accessToken)
	fmt.Printf("WxGetUserList:%#v\n", ulist)
	//WxDelMenu(accessToken)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "成功",
	})
}

var wxReceiveFunc = func(msg wxapi.WxReceiveCommonMsg) error {
    fmt.Println("weixin msg received")
    fmt.Printf("%#v\n", msg)
    touser := msg.FromUserName
    content := msg.Content
    accessToken := wxapi.WxGetAccessToken()

    go func() {
        // 用户刚关注
        if msg.MsgType == "event" && msg.Event == "subscribe" {
            wxapi.WxPostCustomTextMsg(accessToken, touser, "终于等到你\n\n免费、高效的错别字校对工具\n\n公告：爱校对官方暂未发布任何移动端应用！！\n\n官方校对入口 ：www.ijiaodui.com ，Wps插件、Word插件和Chrome（谷歌浏览器）插件可在PC端官网 “下载中心” 内页获取。\n\n欢迎提出您的宝贵意见，我们将第一时间回复，祝您生活愉快~")
        } else {
            // Call the InputStatusRequest function before sending the message
            wxapi.InputStatusRequest(accessToken, touser, "Typing")

            // Set up a timer to limit response time to 14 seconds
            timer := time.NewTimer(14 * time.Second)
            defer timer.Stop()

            // Set up a channel to signal when a message is received
            ch := make(chan string)

            go func() {
                resp := chatapi.AskChatAI(content)
                if resp != "" {
                    ch <- resp
                } else {
                    ch <- "chatGPT服务异常"
                }
            }()

			select {
			case res := <-ch:
				wxapi.WxPostCustomTextMsg(accessToken, touser, res)
			case <-timer.C:
				wxapi.WxPostCustomTextMsg(accessToken, touser, "请耐心等待")
				go func() {
					resp := chatapi.AskChatAI(content)
					if resp != "" {
						wxapi.WxPostCustomTextMsg(accessToken, touser, resp)
					} else {
						wxapi.WxPostCustomTextMsg(accessToken, touser, "chatGPT服务异常")
					}
					wxapi.InputStatusRequest(accessToken, touser, "CancelTyping")
				}()
			}
		
            wxapi.InputStatusRequest(accessToken, touser, "CancelTyping")
        }
    }()

    if msg.Event == "VIEW" {
        fmt.Printf("WxGetAccessToken:%s\n", accessToken)
        ulist := wxapi.WxGetUserList(accessToken)
        fmt.Printf("WxGetUserList:%#v\n", ulist)

        resp, _ := wxapi.WxGetUserInfo(accessToken, msg.FromUserName)
        //WxCreateMenu(accessToken)
        fmt.Printf("%s\n", resp)
    }

    return nil
}


func main() {
	fmt.Println("开启定时触发任务")

	gocron.Start()

	wxapi.WxReceiveFunc = wxReceiveFunc

	router := gin.Default()
	router.GET("/", wxapi.HandleWxLogin)
	router.GET("/test", handleTest)
	router.POST("/", wxapi.HandleWxPostRecv)

	// 创建菜单
	router.GET("/add", func(c *gin.Context) {
		fmt.Printf("==>add Enter\n")

		accessToken := wxapi.WxGetAccessToken()
		fmt.Printf("WxGetAccessToken:%s\n", accessToken)
		ulist := wxapi.WxGetUserList(accessToken)
		fmt.Printf("WxGetUserList:%#v\n", ulist)

		menu := ``
		
		resp, _ := wxapi.WxCreateMenu(accessToken, menu)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  resp,
		})
	})

	//　删除菜单
	router.GET("/del", func(c *gin.Context) {
		fmt.Printf("==>del Enter\n")

		accessToken := wxapi.WxGetAccessToken()
		fmt.Printf("WxGetAccessToken:%s\n", accessToken)
		ulist := wxapi.WxGetUserList(accessToken)
		fmt.Printf("WxGetUserList:%#v\n", ulist)

		resp, _ := wxapi.WxDelMenu(accessToken)
		//WxCreateMenu(accessToken)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  resp,
		})
	})
	//运行的端口
	router.Run(":80")
	// router.Run(":8000")
}
