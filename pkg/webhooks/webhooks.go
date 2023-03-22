// imported from https://github.com/sudosammy/knary/blob/master/libknary/webhooks.go
// This package is useful to send full output messages from interactsh-client to remote webhooks
// Configuration is made through environment variables what turns it simple to use with docker environments

package webhooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func SendMsg(msg string) {
	re := regexp.MustCompile(`\r?\n`)
	msg = re.ReplaceAllString(msg, "\\n")
	msg = strings.ReplaceAll(msg, "\"", "\\\"")

	if os.Getenv("SLACK_WEBHOOK") != "" {
		jsonMsg := []byte(`{"username":"knary","icon_emoji":":bird:","text":"` + msg + `"}`)
		_, err := http.Post(os.Getenv("SLACK_WEBHOOK"), "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("SLACK_WEBHOOK_ERROR: " + err.Error())
		}
	}

	if os.Getenv("PUSHOVER_TOKEN") != "" && os.Getenv("PUSHOVER_USER") != "" {
		jsonMsg := []byte(`{"token":"` + os.Getenv("PUSHOVER_TOKEN") + `","user":"` + os.Getenv("PUSHOVER_USER") + `","message":"` + msg + `"}`)
		_, err := http.Post("https://api.pushover.net/1/messages.json/", "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("PUSHOVER_WEBHOOK_ERROR: " + err.Error())
		}
	}

	if os.Getenv("TELEGRAM_CHATID") != "" && os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
		msg = strings.ReplaceAll(msg, "```From:", "\nFrom:")
		re = regexp.MustCompile("```\\n?")
		msg = re.ReplaceAllString(msg, "")

		jsonMsg := []byte(`{"chat_id": "` + os.Getenv("TELEGRAM_CHATID") + `", "text": "` + msg + `"}`)
		_, err := http.Post("https://api.telegram.org/bot"+os.Getenv("TELEGRAM_BOT_TOKEN")+"/sendMessage", "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("TELEGRAM_WEBHOOK_ERROR: " + err.Error())
		}
	}

	if os.Getenv("LARK_WEBHOOK") != "" {
		re = regexp.MustCompile("```\\n?")
		msg = re.ReplaceAllString(msg, "")

		jsonMsg := []byte("{\n")

		if larkSecret := os.Getenv("LARK_SECRET"); larkSecret != "" {
			// Generate signature
			timestamp := time.Now().Unix()
			sig, err := SignLark(os.Getenv("LARK_SECRET"), timestamp)
			if err != nil {
				fmt.Println("LARK_WEBHOOK_ERROR: " + err.Error())
			}

			// Add fields to payload
			sigFields := fmt.Sprintf(""+
				"    \"timestamp\": \"%d\",\n"+
				"    \"sign\": \"%s\",\n", timestamp, sig)

			jsonMsg = append(jsonMsg, sigFields...)
		}

		// Escape hell. Probably could have just backticked lol.
		postBody := fmt.Sprintf(""+
			"    \"msg_type\": \"post\",\n"+
			"    \"content\": {\n"+
			"        \"post\": {\n"+
			"            \"en_us\": {\n"+
			"                \"title\": \"Knary Triggered 🐦\",\n"+
			"                \"content\": [\n"+
			"                    [\n"+
			"                        {\n"+
			"                            \"tag\": \"text\",\n"+
			"                            \"text\": \"%s\"\n"+
			"                        }\n"+
			"                    ]\n"+
			"                ]\n"+
			"            }\n"+
			"        }\n"+
			"    }\n"+
			"}", msg)

		jsonMsg = append(jsonMsg, postBody...)

		_, err := http.Post(os.Getenv("LARK_WEBHOOK"), "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("LARK_WEBHOOK_ERROR: " + err.Error())
		}
	}

	if os.Getenv("DISCORD_WEBHOOK") != "" {
		jsonMsg := []byte(`{"username":"knary","text":"` + msg + `"}`)
		_, err := http.Post(os.Getenv("DISCORD_WEBHOOK")+"/slack", "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("DISCORD_WEBHOOK_ERROR: " + err.Error())
		}
	}

	if os.Getenv("TEAMS_WEBHOOK") != "" {
		// swap ``` with <pre> for MS teams :face-with-rolling-eyes:
		msg = strings.Replace(msg, "```", "</pre>", 2)
		msg = strings.Replace(msg, "</pre>", "<pre>", 1)

		jsonMsg := []byte(`{"text":"` + msg + `"}`)
		_, err := http.Post(os.Getenv("TEAMS_WEBHOOK"), "application/json", bytes.NewBuffer(jsonMsg))

		if err != nil {
			fmt.Println("TEAMS_WEBHOOK_ERROR: " + err.Error())
		}
	}

	// should be simple enough to add support for other webhooks here
}

// https://www.feishu.cn/hc/en-US/articles/360024984973-Bot-Use-bots-in-groups
func SignLark(secret string, timestamp int64) (string, error) {
	//timestamp + key as sha256, then base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
