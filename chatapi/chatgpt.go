package chatapi
import (
	"strings"
	"fmt"
	"time"
	"context"
	"github.com/otiai10/openaigo"
)

var (
	// CHATKEY 
	CHATKEY = "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"
	RESPOND = "免费、高效的错别字校对工具\n\n公告：爱校对官方暂未发布任何移动端应用！！\n\n官方校对入口 ：www.ijiaodui.com ，Wps插件、Word插件和Chrome（谷歌浏览器）插件可在PC端官网 “下载中心” 内页获取。\n\n欢迎提出您的宝贵意见，我们将第一时间回复，祝您生活愉快 ~"
)

func AskChatAI(question string) string{
	if strings.HasPrefix(question, "#") {
		fmt.Printf("\033[31mAsk chatGPT,\033[0m question:%s\n", strings.TrimPrefix(question, "#"))
		begin := time.Now()
		client := openaigo.NewClient(CHATKEY)
		request := openaigo.CompletionRequestBody {
			Model:  "text-davinci-003",
			Prompt: []string{question},
			MaxTokens: 2000,
		}
		ctx := context.Background()
		response, err := client.Completion(ctx, request)
		elapsed := time.Since(begin)
		//fmt.Println(response, err)
		if err != nil {
			fmt.Printf("\033[31mError:\033[0m %v\nTime: ", err, elapsed)
			return ""
		} else {
			fmt.Printf("Answer:%+v\n\033[32mTime:\033[0m %v\n\n", response, elapsed)
			return strings.TrimLeft(response.Choices[0].Text, "\n")
		}
	} else {
		return RESPOND
	}
}