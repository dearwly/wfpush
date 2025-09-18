package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// ... platform, params, SubFissure, Warframe 保持不变 ...
var platform string = "pc"
var params map[string]string = map[string]string{
	"language": "en",
}

type SubFissure struct {
	MissionType string `yaml:"mission_type"`
	IsHard      bool   `yaml:"is_hard"`
	Tier        string `yaml:"tier"`
	Location    string `yaml:"location"`
}

type Warframe struct {
	SubsFissures []SubFissure
}

// Fissure 结构体修改：移除 ETA 字段
type Fissure struct {
	ID          string `json:"id"`
	Activation  string `json:"activation"`
	Expiry      string `json:"expiry"`
	Node        string `json:"node"`
	MissionType string `json:"missionType"`
	EnemyKey    string `json:"enemyKey"`
	Tier        string `json:"tier"`
	IsHard      bool   `json:"isHard"`
	IsStorm     bool   `json:"isStorm"`
}

// GetETA 方法修改：移除对 f.ETA 的依赖
func (f *Fissure) GetETA() string {
	if f.Expiry == "" {
		return "未知"
	}

	expiryTime, err := time.Parse(time.RFC3339Nano, f.Expiry)
	if err != nil {
		Log(fmt.Sprintf("解析时间戳失败 '%s': %v", f.Expiry, err), WARNING)
		return "解析失败"
	}

	remaining := time.Until(expiryTime)

	if remaining <= 0 {
		return "已结束"
	}

	// 格式化为 "Xh Ym Zs"
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d时 %d分 %d秒", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d分 %d秒", minutes, seconds)
}

// ... GetFissuresWithRetry, GetRawFissures, SendSubsFissures, CheckSubsFissure 保持不变 ...
func GetFissuresWithRetry(f Warframe, maxRetries int, delay time.Duration) ([]byte, error) {
	url := fmt.Sprintf("https://api.warframestat.us/%s/fissures", platform)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				Log(fmt.Sprintf("请求失败 (尝试 %d/%d): %v", attempt+1, maxRetries, err), ERROR)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("请求失败: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 && res.StatusCode < 600 {
			if attempt < maxRetries-1 {
				Log(fmt.Sprintf("请求失败 (状态码 %d) (尝试 %d/%d)", res.StatusCode, attempt+1, maxRetries), ERROR)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("请求失败，状态码: %d", res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}
		return body, nil
	}
	return nil, fmt.Errorf("重试次数达到最大限制，退出。")
}

func GetRawFissures(f Warframe) ([]Fissure, error) {
	data, err := GetFissuresWithRetry(f, 9999, 2*time.Second)
	if err != nil {
		return nil, err
	}

	var fissures []Fissure
	if err := json.Unmarshal(data, &fissures); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}
	return fissures, err
}

func SendSubsFissures(f Warframe, cfg *Config) error {
	fissures, isNew, err := CheckSubsFissure(f)
	if err != nil {
		return err
	}

	if cfg.Email.Enabled && isNew && len(fissures) > 0 {
		Log("发现新裂缝，准备发送邮件。", INFO)

		htmlBody, err := formatFissuresToHTML(fissures)
		if err != nil {
			Log(fmt.Sprintf("创建HTML邮件失败: %v", err), ERROR)
			return err
		}

		Send_email(htmlBody, cfg)

	} else if !cfg.Email.Enabled {
		Log("邮件发送功能已在配置中禁用。", INFO)
	}

	return nil
}

func CheckSubsFissure(f Warframe) ([]Fissure, bool, error) {
	type existsFissures struct {
		Fissures []Fissure `json:"fissure"`
	}

	file, err := os.Open("data.json")
	if err != nil {
		return []Fissure{}, false, err
	}
	defer file.Close()

	var existingData existsFissures
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&existingData)
	if err != nil {
		return []Fissure{}, false, err
	}

	currentFissures, err := GetRawFissures(f)
	if err != nil {
		return []Fissure{}, false, err
	}

	var allMatchingFissures []Fissure
	var newSubscribedFissures []Fissure
	isNew := false

	for _, fissure := range currentFissures {
		missionTypeCN := TranslateMissionType(fissure.MissionType)
		tierCN := TranslateTier(fissure.Tier)

		for _, sub := range f.SubsFissures {
			matchMission := (sub.MissionType == "" || sub.MissionType == missionTypeCN)
			matchTier := (sub.Tier == "" || sub.Tier == tierCN)
			matchHard := (sub.IsHard == fissure.IsHard)

			if matchMission && matchTier && matchHard {
				allMatchingFissures = append(allMatchingFissures, fissure)

				isAlreadyKnown := false
				for _, knownFissure := range existingData.Fissures {
					if knownFissure.ID == fissure.ID {
						isAlreadyKnown = true
						break
					}
				}

				if !isAlreadyKnown {
					isNew = true
					newSubscribedFissures = append(newSubscribedFissures, fissure)
				}
				break
			}
		}
	}

	existingData.Fissures = currentFissures

	Log_INFO("当前符合订阅的裂缝: ")
	if len(allMatchingFissures) == 0 {
		Log_INFO("当前暂无订阅裂缝。")
	} else {
		for _, fissure := range allMatchingFissures {
			texts, _ := formatPrint(fissure)
			Log_INFO(texts)
		}
	}

	file, err = os.Create("data.json")
	if err != nil {
		Log(fmt.Sprint("Error creating file:", err), ERROR)
		return []Fissure{}, false, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(existingData)
	if err != nil {
		Log(fmt.Sprint("Error encoding JSON:", err), ERROR)
		return []Fissure{}, false, err
	}

	return newSubscribedFissures, isNew, nil
}

// formatPrint 修改：使用 TranslateNode
func formatPrint(fissure Fissure) (string, error) {
	head := ""
	if fissure.IsHard {
		head = "钢铁之路 "
	}
	if fissure.IsStorm {
		head = "虚空风暴 "
	}

	return fmt.Sprintf("%s%s %s %s %s 剩余%s",
		head,
		TranslateMissionType(fissure.MissionType),
		TranslateFaction(fissure.EnemyKey),
		TranslateTier(fissure.Tier),
		TranslateNode(fissure.Node), // 修改
		fissure.GetETA(),
	), nil
}

// formatFissuresToHTML 修改：在 funcMap 中注册新的翻译函数
func formatFissuresToHTML(fissures []Fissure) (string, error) {
	funcMap := template.FuncMap{
		"TranslateMissionType": TranslateMissionType,
		"TranslateFaction":     TranslateFaction,
		"TranslateTier":        TranslateTier,
		"TranslateNode":        TranslateNode, // 新增
	}

	tmpl, err := template.New("email").Funcs(funcMap).Parse(HTMLTemplate)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, fissures); err != nil {
		return "", err
	}

	return body.String(), nil
}
