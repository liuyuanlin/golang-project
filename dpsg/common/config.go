package common

import (
	"golang-project/dpsg/csvcfg"
	"golang-project/dpsg/jscfg"
	"golang-project/dpsg/logger"
	"os"
	"path"
	"strings"
)

//gateserver配置
type GateServerCfg struct {
	GsIpForClient string
	GsIpForServer string
	DebugHost     string
	GcTime        uint8
}

//lockserver配置
type LockServerCfg struct {
	LockHost  string
	DebugHost string
	GcTime    uint8
}

//center配置
type CenterConfig struct {
	CenterHost       string
	CenterForGm      string
	DebugHost        string
	GcTime           uint8
	MainCacheProfile CacheConfig `json:"maincache"`
	ClanCacheProfile CacheConfig `json:"clancache"`
	//add for update rankresult
	UpdateTime string
}

//chatserver配置
type ChatServerCfg struct {
	ListenForClient  string
	ListenForServer  string
	ListenForGm      string
	DebugHost        string
	GcTime           uint8
	MainCacheProfile CacheConfig `json:"maincache"`
}

//gmserver配置
type GmServerCfg struct {
	GmServerIp string
	DbServerIp string
}

//logserver配置
type LogServerCfg struct {
	LogHost   string
	Host      string
	Port      uint16
	User      string
	Pass      string
	Dbname    string
	Charset   string
	DebugHost string
	GcTime    uint8
}

//cns配置
type CnsConfig struct {
	CnsHost          string
	CnsHostForClient string
	CnsForCenter     string
	FsHost           []string
	DebugHost        string
	GcTime           uint8
}

//designer路径
type DesignerDir struct {
	Designer           string
	OpenGm             uint8
	TW_GooglePayPubKey string
	OpenGemBuy         uint8
	OpenTrophyAdd      uint8
	TW_3PaySecretKey   string
	IosPayTest         uint8
	TXAppId            int
	TXAppKey           string
	TXLoginUrl         string
}

//GooglePay订单信息
type GooglePayOrderInfo struct {
	OrderId          string
	PackageName      string
	ProductId        string
	PurchaseTime     uint64
	PurchaseState    int
	DeveloperPayload string
	PurchaseToken    string
}

var tttawardCfg map[string]*[]TttAwardCfg

type TttAwardCfg struct {
	ID            string
	RankRangeType string
	RankRangeMin  uint32
	RankRangeMax  uint32
	AwardType1    string
	Award1        uint32
	AwardType2    string
	Award2        uint32
}

//add for get daily money
type GlobalInfo struct {
	TID         string //分数
	Mark        string //爵位
	AwardType1  string //宝石
	AwardCount1 uint32 //宝石数量
	AwardType2  string //武魂
	AwardCount2 uint32 //武魂数量
}

var globalinfoCfg map[string]*[]GlobalInfo

//add for challenge
var challengeCfg map[string]*[]ChallengeCfg

type ChallengeCfg struct {
	Value int32
}

//add for send present
type PresentInfo struct {
	Tite   string //标志
	Pic    string //资源类型图标
	Number string //数量
	Type   string //资源名称
}

var sendInfoCfg map[string]*[]PresentInfo

//add for send present

func LoadPresentCfg() {
	globalInfo := path.Join(GetDesignerDir(), "liwu.csv")
	csvcfg.LoadCSVConfig(globalInfo, &sendInfoCfg)
}

func GetPresentCfg(key string) *PresentInfo {
	cfg, exist := sendInfoCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//读取配置的接口
func LoadConfigFiles() {
	tttawards := path.Join(GetDesignerDir(), "ttt_award.csv")
	csvcfg.LoadCSVConfig(tttawards, &tttawardCfg)
}
func GetTttAwardCfg(key string) *TttAwardCfg {
	cfg, exist := tttawardCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//add for challenge
func LoadChallengeConfigFiles() {
	challengeInfo := path.Join(GetDesignerDir(), "globals.csv")
	csvcfg.LoadCSVConfig(challengeInfo, &challengeCfg)
}

func GetChallengeInfoCfg(key string) int32 {
	cfg, exist := challengeCfg[strings.ToLower(key)]
	if !exist {
		return 0
	}

	return (*cfg)[0].Value
}

//add for get daily money
func LoadDailyMoney() {
	globalInfo := path.Join(GetDesignerDir(), "juewei.csv")
	csvcfg.LoadCSVConfig(globalInfo, &globalinfoCfg)
}

func GetTitleInfoCfg(key string) *GlobalInfo {
	cfg, exist := globalinfoCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

func GetCfgSize() int {
	globalInfo := path.Join(GetDesignerDir(), "juewei.csv")
	size := csvcfg.GetCfgSize(globalInfo)

	return size
}

//center
func ReadCenterConfig(cfg *CenterConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "center.json"), cfg); err != nil {
		logger.Fatal("read center config failed, %v", err)
		return err
	}

	return nil
}

//gate
func ReadGateConfig(cfg *GateServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "gscfg.json"), cfg); err != nil {
		logger.Fatal("read gate config failed, %v", err)
		return err
	}

	return nil
}

//chat
func ReadChatConfig(cfg *ChatServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "chatservercfg.json"), cfg); err != nil {
		logger.Fatal("read chat config failed, %v", err)
		return err
	}

	return nil
}

//gm
func ReadGmConfig(cfg *GmServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "gmservercfg.json"), cfg); err != nil {
		logger.Fatal("read chat config failed, %v", err)
		return err
	}

	return nil
}

//log
func ReadLogConfig(cfg *LogServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "lgscfg.json"), cfg); err != nil {
		logger.Fatal("read chat config failed, %v", err)
		return err
	}

	return nil
}

//加锁服务器
func ReadLockServerConfig(cfg *LockServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "lscfg.json"), cfg); err != nil {
		logger.Fatal("read lock config failed, %v", err)
		return err
	}

	return nil
}

//cns服务器配置
func ReadCnsServerConfig(file string, cfg *CnsConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, file), cfg); err != nil {
		logger.Fatal("read center config failed, %v", err)
		return err
	}

	return nil
}

//designer配置
var pDesignerCfg *DesignerDir

func GetDesignerCfg() *DesignerDir {
	if pDesignerCfg != nil {
		return pDesignerCfg
	}

	cfgpath, _ := os.Getwd()
	if err := jscfg.ReadJson(path.Join(cfgpath, "designer.json"), &pDesignerCfg); err != nil {
		logger.Fatal("read designer config failed, %v", err)
		return nil
	}

	return pDesignerCfg
}

//designer
func GetDesignerDir() string {
	return GetDesignerCfg().Designer
}

//是否打开gm指令
func IsOpenGm() bool {
	return GetDesignerCfg().OpenGm == 1
}

//获得台湾版本的google支付publickkey
func GetTWGooglePubkey() string {
	return GetDesignerCfg().TW_GooglePayPubKey
}

//获得台湾版本的第三方支付secretkey
func GetTW3PaySecretKey() string {
	return GetDesignerCfg().TW_3PaySecretKey
}

//ios测试支付
func IsIosPayTest() bool {
	return GetDesignerCfg().IosPayTest == 1
}

func ReadGooglePayOrder(order string) GooglePayOrderInfo {
	var odi GooglePayOrderInfo
	jscfg.ReadJsonByDataStr(order, &odi)
	return odi
}

//是否打开GemBuy接口
func IsOpenGemBuy() bool {
	return GetDesignerCfg().OpenGemBuy == 1
}

//是否打开TrophyAdd接口
func IsOpenTrophyAdd() bool {
	return GetDesignerCfg().OpenTrophyAdd == 1
}

//腾讯对接使用
func GetQQAppInfo() (int, string) {
	info := GetDesignerCfg()

	return info.TXAppId, info.TXAppKey
}

//qq登陆地址
func GetQQLoginUrl() string {
	return GetDesignerCfg().TXLoginUrl
}
