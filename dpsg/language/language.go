package language

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/csvcfg"
	"golang-project/dpsg/rpc"
	"path"
	"strconv"
	"strings"
)

type LanguageCfg struct {
	TID string
	CH  string
	EN  string
}

var languageCfg map[string]*[]LanguageCfg

type LocationLanguageCfg struct {
	TID string
	STR string
}

var locationLanguageCfg map[string]*[]LocationLanguageCfg

func initLanguage() {
	designerDir := common.GetDesignerDir()
	fullPath := path.Join(designerDir, "textsutf8.csv")
	csvcfg.LoadCSVConfig(fullPath, &languageCfg)
}

//支持变参Format
func GetLanguage(id string, args ...interface{}) string {
	if len(languageCfg) == 0 {
		initLanguage()
	}

	id = strings.TrimSpace(strings.ToLower(id))

	if lan, ok := languageCfg[id]; ok {
		return fmt.Sprintf((*lan)[0].CH, args...)
	}

	return ""
}

func initLocationLanguage() {
	designerDir := common.GetDesignerDir()
	fullPath := path.Join(designerDir, "localtextutf8.csv")
	csvcfg.LoadCSVConfig(fullPath, &locationLanguageCfg)
}

//取地区语言
func GetLocationLanguage(id string, location rpc.GameLocation, args ...interface{}) string {
	if len(locationLanguageCfg) == 0 {
		initLocationLanguage()
	}

	id = strings.TrimSpace(strings.ToLower(id))

	tid := id + strconv.FormatInt(int64(location), 10)
	if lan, ok := locationLanguageCfg[tid]; ok {
		return fmt.Sprintf((*lan)[0].STR, args...)
	}

	//取不到就取默认值
	tid = id + "Default"
	if lan, ok := locationLanguageCfg[tid]; ok {
		return fmt.Sprintf((*lan)[0].STR, args...)
	}

	return ""
}
