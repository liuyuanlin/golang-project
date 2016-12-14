package connector

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//手机qq登陆
//登陆
type stMobileQQLogin struct {
	Appid   int    `json:"appid"`
	Openid  string `json:"openid"`
	Openkey string `json:"openkey"`
	Userip  string `json:"userip"`
}

//登陆返回
type stMobileQQLoginRet struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

var sLoginErrMsg string = "login failed!"
var sPayErrMsg string = "pay failed!"
var sQueryErrMsg string = "query failed!"
var sBalanceErrMsg string = "query balance failed!"
var sFriendsErrMsg string = "query friends failed"

func MobileQQAuth(login *rpc.Login, IP string) (success bool, errmsg string) {
	//配置表
	tencentAppId, tencentAppKey := common.GetQQAppInfo()
	sUrlBase := common.GetQQLoginUrl()

	//当前时间
	st := stMobileQQLogin{
		Appid:   tencentAppId,
		Openid:  login.GetOpenid(),
		Openkey: login.GetOpenkey(),
		Userip:  IP,
	}
	body, err := json.Marshal(st)
	if err != nil {
		logger.Error("MobileQQAuth Marshal stMobileQQLogin error: %v", err)
		return false, sLoginErrMsg
	}
	logger.Info("MobileQQAuth client info: ", string(body))
	buf := bytes.NewBuffer(body)

	timecur := strconv.FormatInt(time.Now().Unix(), 10)
	h := md5.New()
	io.WriteString(h, tencentAppKey)
	io.WriteString(h, timecur)
	sig := fmt.Sprintf("%x", h.Sum(nil))
	fullurl := fmt.Sprintf("%s/auth/verify_login/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
		sUrlBase, timecur, tencentAppId, sig, login.GetOpenid())
	logger.Info("MobileQQAuth tx url: ", fullurl)
	res, err := http.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("MobileQQAuth http.Post error: %v", err)
		return false, sLoginErrMsg
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("MobileQQAuth body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("MobileQQAuth ioutil.ReadAll error: %v", err)
		return false, sLoginErrMsg
	}

	rst := stMobileQQLoginRet{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("MobileQQAuth ioutil.ReadAll error: %v", err)
		return false, sLoginErrMsg
	}

	if rst.Ret != 0 {
		return false, rst.Msg
	}

	return true, ""
}

//QQ支付
//支付返回
type stMobileQQPayRet struct {
	Ret     int    `json:"ret"`
	Msg     string `json:"msg"`
	BillNo  string `json:"billno"`
	Balance int    `json:"balance"`
}

//cookies
type Jar struct {
	cookies []*http.Cookie
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

func MobileQQPay(p *player, number int) (success bool, errmsg string, billno string, balance int) {
	if p == nil || p.mobileqqinfo == nil {
		return false, sPayErrMsg, "", 0
	}

	//配置表
	tencentAppId, tencentAppKey := common.GetQQAppInfo()
	sUrlBase := common.GetQQLoginUrl()

	openid := p.mobileqqinfo.Openid
	//非qq渠道就不用走这个流程了
	if len(openid) == 0 {
		return
	}

	openkey := p.mobileqqinfo.Openkey
	pay_token := p.mobileqqinfo.Pay_token
	//pfkey := p.mobileqqinfo.Pfkey
	pfkey := "pfkey"
	pf := p.mobileqqinfo.Pf
	timecur := strconv.FormatInt(time.Now().Unix(), 10)
	//ip := p.conn.GetRemoteIp()
	amt := strconv.FormatInt(int64(number), 10)

	v := make(url.Values)
	v.Add("openid", openid)
	v.Add("openkey", openkey)
	v.Add("pay_token", pay_token)
	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
	v.Add("ts", timecur)
	v.Add("pf", pf)
	v.Add("format", "json")
	//v.Add("userip", ip)
	v.Add("zoneid", "1")
	v.Add("amt", amt)
	v.Add("pfkey", pfkey)

	sigurl := "GET&" + url.QueryEscape("/mpay/pay_m") + "&" + url.QueryEscape(v.Encode())
	sigappkey := tencentAppKey + "&"
	h := sha1.New()
	io.WriteString(h, sigurl)
	mac := hmac.New(sha1.New, []byte(sigappkey))
	mac.Write([]byte(sigurl))
	dec := fmt.Sprintf("%s", mac.Sum(nil))
	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

	v.Add("sig", sig)

	fullurl := sUrlBase + "/mpay/pay_m?" + v.Encode()
	logger.Info("MobileQQPay tx url: ", fullurl)

	request, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		logger.Error("MobileQQPay http.NewRequest error: %v", err)
		return false, sPayErrMsg, "", 0
	}
	request.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: url.QueryEscape("openid"),
	})
	request.AddCookie(&http.Cookie{
		Name:  "session_type",
		Value: url.QueryEscape("kp_actoken"),
	})
	request.AddCookie(&http.Cookie{
		Name:  "org_loc",
		Value: url.QueryEscape("/mpay/pay_m"),
	})

	jar := &Jar{cookies: make([]*http.Cookie, 0)}
	client := &http.Client{Jar: jar}

	resp, err := client.Do(request)
	if err != nil {
		logger.Error("MobileQQPay client.Do error: %v", err)
		return false, sPayErrMsg, "", 0
	}

	b, err := ioutil.ReadAll(resp.Body)
	logger.Info("MobileQQPay body info:%s", string(b))
	resp.Body.Close()
	if err != nil {
		logger.Error("MobileQQPay ioutil.ReadAll error: %v", err)
		return false, sPayErrMsg, "", 0
	}

	rst := stMobileQQPayRet{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("MobileQQPay ioutil.ReadAll error: %v", err)
		return false, sPayErrMsg, "", 0
	}

	/*0：成功；
	1004：余额不足。
	1018：登陆校验失败。
	其它：失败*/
	if rst.Ret != 0 {
		return false, rst.Msg, "", 0
	}

	return true, "", rst.BillNo, rst.Balance
}

//查询名字等基础信息
type stMobileQQQuery struct {
	Appid       int    `json:"appid"`
	AccessToken string `json:"accessToken"`
	Openid      string `json:"openid"`
}

type stMobileQQQueryResult struct {
	Ret        int    `json:"ret"`
	Msg        string `json:"msg"`
	NickName   string `json:"nickName"`
	Gender     string `json:"gender"`
	Picture40  string `json:"picture40"`
	Picture100 string `json:"picture100"`
}

func MobileQQQuery(p *player) (success bool, errmsg string, nickname string, gender string, picture40 string, picture100 string) {
	//返回值初始化
	success = false
	errmsg = sQueryErrMsg
	nickname = ""
	gender = ""
	picture40 = ""
	picture100 = ""

	if p == nil || p.mobileqqinfo == nil {
		return
	}

	//配置表
	tencentAppId, tencentAppKey := common.GetQQAppInfo()
	sUrlBase := common.GetQQLoginUrl()

	openid := p.mobileqqinfo.Openid
	openkey := p.mobileqqinfo.Openkey

	//非qq渠道就不用走这个流程了
	if len(openid) == 0 {
		return
	}

	//当前时间
	st := stMobileQQQuery{
		Appid:       tencentAppId,
		AccessToken: openkey,
		Openid:      openid,
	}
	body, err := json.Marshal(st)
	if err != nil {
		logger.Error("MobileQQQuery Marshal stMobileQQLogin error: %v", err)
		return
	}
	logger.Info("MobileQQQuery client info: ", string(body))
	buf := bytes.NewBuffer(body)

	timecur := strconv.FormatInt(time.Now().Unix(), 10)
	h := md5.New()
	io.WriteString(h, tencentAppKey)
	io.WriteString(h, timecur)
	sig := fmt.Sprintf("%x", h.Sum(nil))
	fullurl := fmt.Sprintf("%s/relation/qqprofile/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
		sUrlBase, timecur, tencentAppId, sig, openid)
	logger.Info("MobileQQQuery tx url: ", fullurl)
	res, err := http.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("MobileQQQuery http.Post error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("MobileQQQuery body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("MobileQQQuery ioutil.ReadAll error: %v", err)
		return
	}

	rst := stMobileQQQueryResult{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("MobileQQQuery ioutil.ReadAll error: %v", err)
		return
	}

	if rst.Ret != 0 {
		errmsg = rst.Msg
		return
	}

	success = true
	errmsg = ""
	nickname = rst.NickName
	gender = rst.Gender
	picture40 = rst.Picture40
	picture100 = rst.Picture100

	return
}

//查询余额
type stMobileQQBalanceRet struct {
	Ret int `json:"ret"`
	//总游戏币个数，包括赠送
	Balance int `json:"balance"`
	//赠送游戏币个数
	GenBalance int `json:"gen_balance"`
	FirstSave  int `json:"first_save"`
}

func MobileQQBalance(p *player) (success bool, errmsg string, number int) {
	success = false
	errmsg = sBalanceErrMsg
	number = 0

	if p == nil || p.mobileqqinfo == nil {
		return
	}

	//配置表
	tencentAppId, tencentAppKey := common.GetQQAppInfo()
	sUrlBase := common.GetQQLoginUrl()

	openid := p.mobileqqinfo.Openid
	//非qq渠道就不用走这个流程了
	if len(openid) == 0 {
		return
	}

	openkey := p.mobileqqinfo.Openkey
	pay_token := p.mobileqqinfo.Pay_token
	//pfkey := p.mobileqqinfo.Pfkey
	pfkey := "pfkey"
	pf := p.mobileqqinfo.Pf
	timecur := strconv.FormatInt(time.Now().Unix(), 10)
	//ip := p.conn.GetRemoteIp()

	v := make(url.Values)
	v.Add("openid", openid)
	v.Add("openkey", openkey)
	v.Add("pay_token", pay_token)
	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
	v.Add("ts", timecur)
	v.Add("pf", pf)
	v.Add("format", "json")
	//v.Add("userip", ip)
	v.Add("zoneid", "1")
	v.Add("pfkey", pfkey)

	sigurl := "GET&" + url.QueryEscape("/mpay/get_balance_m") + "&" + url.QueryEscape(v.Encode())
	sigappkey := tencentAppKey + "&"
	h := sha1.New()
	io.WriteString(h, sigurl)
	mac := hmac.New(sha1.New, []byte(sigappkey))
	mac.Write([]byte(sigurl))
	dec := fmt.Sprintf("%s", mac.Sum(nil))
	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

	v.Add("sig", sig)

	fullurl := sUrlBase + "/mpay/get_balance_m?" + v.Encode()
	logger.Info("MobileQQBalance tx url: ", fullurl)

	request, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		logger.Error("MobileQQBalance http.NewRequest error: %v", err)
		return
	}
	request.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: url.QueryEscape("openid"),
	})
	request.AddCookie(&http.Cookie{
		Name:  "session_type",
		Value: url.QueryEscape("kp_actoken"),
	})
	request.AddCookie(&http.Cookie{
		Name:  "org_loc",
		Value: url.QueryEscape("/mpay/get_balance_m"),
	})

	jar := &Jar{cookies: make([]*http.Cookie, 0)}
	client := &http.Client{Jar: jar}

	resp, err := client.Do(request)
	if err != nil {
		logger.Error("MobileQQBalance client.Do error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	logger.Info("MobileQQBalance body info:%s", string(b))
	resp.Body.Close()
	if err != nil {
		logger.Error("MobileQQBalance ioutil.ReadAll error: %v", err)
		return
	}

	rst := stMobileQQBalanceRet{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("MobileQQBalance ioutil.ReadAll error: %v", err)
		return
	}

	/*0：成功；
	1001：参数错误
	1018：登陆校验失败。*/
	if rst.Ret != 0 {
		return
	}

	success = true
	errmsg = ""
	number = rst.Balance

	return
}

//好友列表
type stMobileQQFriendBase struct {
	OpenId   string `json:"openid"`
	NickName string `json:"nickName"`
	Gender   string `json:"gender"`
	Picture  string `json:"figureurl_qq"`
}

type stMobileQQFriendsResult struct {
	Ret  int                     `json:"ret"`
	Msg  string                  `json:"msg"`
	List []*stMobileQQFriendBase `json:"lists"`
}

func MobileQQFriends(p *player) (success bool, errmsg string, list []*stMobileQQFriendBase) {
	//返回值初始化
	success = false
	errmsg = sFriendsErrMsg

	if p == nil || p.mobileqqinfo == nil {
		return
	}

	//配置表
	tencentAppId, tencentAppKey := common.GetQQAppInfo()
	sUrlBase := common.GetQQLoginUrl()

	openid := p.mobileqqinfo.Openid
	openkey := p.mobileqqinfo.Openkey

	//非qq渠道就不用走这个流程了
	if len(openid) == 0 {
		return
	}

	//当前时间
	st := stMobileQQQuery{
		Appid:       tencentAppId,
		AccessToken: openkey,
		Openid:      openid,
	}
	body, err := json.Marshal(st)
	if err != nil {
		logger.Error("MobileQQFriends Marshal stMobileQQLogin error: %v", err)
		return
	}
	logger.Info("MobileQQFriends client info: ", string(body))
	buf := bytes.NewBuffer(body)

	timecur := strconv.FormatInt(time.Now().Unix(), 10)
	h := md5.New()
	io.WriteString(h, tencentAppKey)
	io.WriteString(h, timecur)
	sig := fmt.Sprintf("%x", h.Sum(nil))
	fullurl := fmt.Sprintf("%s/relation/qqfriends_detail/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
		sUrlBase, timecur, tencentAppId, sig, openid)
	logger.Info("MobileQQFriends tx url: ", fullurl)
	res, err := http.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("MobileQQFriends http.Post error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("MobileQQFriends body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("MobileQQFriends ioutil.ReadAll error: %v", err)
		return
	}

	rst := stMobileQQFriendsResult{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("MobileQQFriends ioutil.ReadAll error: %v", err)
		return
	}

	if rst.Ret != 0 {
		errmsg = rst.Msg
		return
	}

	success = true
	errmsg = ""
	list = rst.List

	return
}
