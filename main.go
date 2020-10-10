package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/amtoaer/bangumi/login"
	"github.com/amtoaer/bangumi/session"
)

func convertWeekday(day float64) string {
	switch day {
	case 1:
		return "MON."
	case 2:
		return "TUE."
	case 3:
		return "WED."
	case 4:
		return "THU."
	case 5:
		return "FRI."
	case 6:
		return "SAT."
	case 7:
		return "SUN."
	default:
		return ""
	}
}

// 因api的summary项返回为空，不得已使用正则匹配网页内容得到剧情简介
func getSummary(a *session.API, id float64) (string, error) {
	strID := strconv.FormatFloat(id, 'f', 0, 64)
	toMatch := regexp.MustCompile(`<div id="subject_summary" class="subject_summary" property="v:summary">([\s\S]+?)</div>`)
	resp, err := a.Client.Get(fmt.Sprintf("https://bangumi.tv/subject/%s", strID))
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	tmpResult := toMatch.FindStringSubmatch(string(content))
	if len(tmpResult) < 2 {
		return "", errors.New("no match")
	}
	result := strings.ReplaceAll(tmpResult[1], `<br />`, "")
	return result, nil
}

func main() {
	if len(os.Args) != 2 {
		return
	}
	username := os.Args[1]
	api := login.NoLogin().NewSession()
	result, err := api.UserCollection(username, true, false)
	if err != nil {
		return
	}
	for _, item := range result {
		subject := item["subject"].(map[string]interface{})
		desc, err := getSummary(api, subject["id"].(float64))
		if err != nil {
			return
		}
		fmt.Println("中文名", subject["name_cn"].(string))
		fmt.Println("日文名", subject["name"].(string))
		fmt.Println("简介")
		fmt.Println(desc)
		fmt.Println("放送时间", fmt.Sprintf("%s %s", subject["air_date"].(string), convertWeekday(subject["air_weekday"].(float64))))
		fmt.Println("图片url", subject["images"].(map[string]interface{})["large"].(string))
		fmt.Println("进度", math.Floor((item["ep_status"].(float64)/subject["eps"].(float64))*100))
		fmt.Println("-----------------------------------------------------------------------------")
	}
}
