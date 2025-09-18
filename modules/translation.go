package modules

import "strings"

// missionTranslations 存储任务类型的中英文对照
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

// factionTranslations 存储派系的中英文对照
var factionTranslations = map[string]string{
	"Grineer":   "Grineer",
	"Corpus":    "Corpus",
	"Infested":  "Infested",
	"Corrupted": "Corrupted",
	"Narmer":    "Narmer",
}

// tierTranslations 存储裂缝纪元的中英文对照
var tierTranslations = map[string]string{
	"Lith":    "古纪",
	"Meso":    "前纪",
	"Neo":     "中纪",
	"Axi":     "后纪",
	"Requiem": "安魂",
	"Omnia":   "Omnia",
}

// TranslateMissionType 将英文任务类型翻译为简体中文
func TranslateMissionType(english string) string {
	if translated, ok := missionTranslations[english]; ok {
		return translated
	}
	return english // 如果找不到翻译，返回原文
}

// TranslateFaction 将英文派系翻译为简体中文
func TranslateFaction(english string) string {
	if translated, ok := factionTranslations[english]; ok {
		return translated
	}
	// API有时会返回如 "Corpus Faction" 的字符串，我们只取第一部分
	factionKey := strings.Split(english, " ")[0]
	if translated, ok := factionTranslations[factionKey]; ok {
		return translated
	}
	return english
}

// TranslateTier 将英文纪元翻译为简体中文
func TranslateTier(english string) string {
	if translated, ok := tierTranslations[english]; ok {
		return translated
	}
	return english
}
