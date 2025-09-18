package modules

import (
	"regexp"
	"strings"
)

// ... missionTranslations 和 factionTranslations 保持不变 ...
var missionTranslations = map[string]string{
	"Extermination":    "歼灭",
	"Capture":          "捕获",
	"Mobile Defense":   "移动防御",
	"Defense":          "防御",
	"Rescue":           "救援",
	"Sabotage":         "破坏",
	"Spy":              "间谍",
	"Survival":         "生存",
	"Interception":     "拦截",
	"Excavation":       "挖掘",
	"Disruption":       "中断",
	"Hijack":           "劫持",
	"Assassination":    "刺杀",
	"Infested Salvage": "Infested资源回收",
}

var factionTranslations = map[string]string{
	"Grineer":   "Grineer",
	"Corpus":    "Corpus",
	"Infested":  "Infested",
	"Corrupted": "Corrupted",
	"Narmer":    "Narmer",
}

var tierTranslations = map[string]string{
	"Lith":    "古纪",
	"Meso":    "前纪",
	"Neo":     "中纪",
	"Axi":     "后纪",
	"Requiem": "安魂",
	"Omnia":   "Omnia",
}

// 新增：星球和关键地点的中英文对照
var planetTranslations = map[string]string{
	"Mercury":          "水星",
	"Venus":            "金星",
	"Earth":            "地球",
	"Mars":             "火星",
	"Phobos":           "火卫一",
	"Ceres":            "谷神星",
	"Jupiter":          "木星",
	"Europa":           "木卫二",
	"Saturn":           "土星",
	"Uranus":           "天王星",
	"Neptune":          "海王星",
	"Pluto":            "冥王星",
	"Eris":             "阋神星",
	"Sedna":            "赛德娜",
	"Lua":              "月球",
	"Deimos":           "火卫二",
	"Kuva Fortress":    "赤毒要塞",
	"Zariman Ten Zero": "扎里曼",
	"Void":             "虚空",
}

// 使用 MustCompile，如果正则表达式无效，程序会在启动时 panic，便于立即发现问题
var nodeRegex = regexp.MustCompile(`\((.*?)\)`)

// TranslateMissionType, TranslateFaction, TranslateTier 函数保持不变
func TranslateMissionType(english string) string {
	if translated, ok := missionTranslations[english]; ok {
		return translated
	}
	return english
}

func TranslateFaction(english string) string {
	if translated, ok := factionTranslations[english]; ok {
		return translated
	}
	factionKey := strings.Split(english, " ")[0]
	if translated, ok := factionTranslations[factionKey]; ok {
		return translated
	}
	return english
}

func TranslateTier(english string) string {
	if translated, ok := tierTranslations[english]; ok {
		return translated
	}
	return english
}

// 新增：翻译地点中的星球名称
func TranslateNode(englishNode string) string {
	// 查找括号内的内容，例如 "Ultor (Mars)" -> "Mars"
	matches := nodeRegex.FindStringSubmatch(englishNode)

	// 如果没有找到括号或括号内没有内容，直接返回原字符串
	if len(matches) < 2 {
		return englishNode
	}

	planetInEnglish := matches[1]
	// 在字典中查找翻译
	if planetInChinese, ok := planetTranslations[planetInEnglish]; ok {
		// 替换原字符串中的英文星球名为中文
		return strings.Replace(englishNode, planetInEnglish, planetInChinese, 1)
	}

	// 如果字典里没有，返回原样
	return englishNode
}
