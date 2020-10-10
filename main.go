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
	result = strings.ReplaceAll(result, "\r\n", "")
	return result, nil
}

func http2https(url string) string {
	return strings.ReplaceAll(url, "http", "https")
}

func main() {
	if len(os.Args) != 2 {
		return
	}
	str := `---
layout: bangumi
title: bangumi
comments: false
date: 2019-02-10 21:32:48
keywords:
description:
bangumis:` + "\n"
	username := os.Args[1]
	api := login.NoLogin().NewSession()
	result, err := api.UserCollection(username, true, false)
	if err != nil {
		return
	}
	for _, item := range result {
		subject := item["subject"].(map[string]interface{})
		var (
			img      = http2https(subject["images"].(map[string]interface{})["large"].(string))
			title    = subject["name_cn"].(string)
			progress = math.Floor((item["ep_status"].(float64) / subject["eps"].(float64)) * 100)
			jp       = subject["name"].(string)
			time     = fmt.Sprintf("%s %s", subject["air_date"].(string), convertWeekday(subject["air_weekday"].(float64)))
		)
		desc, err := getSummary(api, subject["id"].(float64))
		if err != nil {
			return
		}
		str += fmt.Sprintf(`  - img: %s
    title: %s
    status: %s
    progress: %.0f
    jp: %s
    time: %s
    desc: %s`, img, title, strconv.FormatFloat(progress, 'f', 0, 64)+"%", progress, jp, time, desc) + "\n"
	}
	str += "---\n"
	fmt.Println(str)
}