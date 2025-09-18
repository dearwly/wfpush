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

// 将API语言参数改为 "en"
var platform string = "pc"
var params map[string]string = map[string]string{
	"language": "en",
}

// ... SubFissure, Warframe, Fissure 结构体保持不变 ...
type SubFissure struct {
	MissionType string `yaml:"mission_type"`
	IsHard      bool   `yaml:"is_hard"`
	Tier        string `yaml:"tier"`
	Location    string `yaml:"location"`
}

type Warframe struct {
	SubsFissures []SubFissure
}

type Fissure struct {
	ID          string `json:"id"`
	MissionType string `json:"missionType"`
	EnemyKey    string `json:"enemyKey"`
	Tier        string `json:"tier"`
	Node        string `json:"node"`
	ETA         string `json:"eta"`
	IsHard      bool   `json:"isHard"`
	IsStorm     bool   `json:"isStorm"`
}

// ... GetFissuresWithRetry, GetRawFissures, GetFissures 函数保持不变 ...
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

func GetFissures(f Warframe, type_ int) ([]map[string]string, error) {
	// ... 此函数在此项目中未被直接调用，可以保持不变或按需修改 ...
	return nil, nil
}

// 修改 SendSubsFissures，生成并发送HTML邮件
func SendSubsFissures(f Warframe, cfg *Config) error {
	fissures, isNew, err := CheckSubsFissure(f)
	if err != nil {
		return err
	}

	if cfg.Email.Enabled && isNew && len(fissures) > 0 {
		Log("发现新裂缝，准备发送邮件。", INFO)

		// 生成HTML邮件正文
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

// 修改 CheckSubsFissure，使用翻译进行匹配
func CheckSubsFissure(f Warframe) ([]Fissure, bool, error) {
	type existsFissures struct {
		Fissures []Fissure `json:"fissure"`
	}

	file, err := os.Open("data.json") // 修改文件名
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
		// 将API返回的英文数据翻译成中文
		missionTypeCN := TranslateMissionType(fissure.MissionType)
		tierCN := TranslateTier(fissure.Tier)

		for _, sub := range f.SubsFissures {
			// 使用翻译后的中文数据与订阅条件进行匹配
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

	existingData.Fissures = currentFissures // 保存的是原始的、未经筛选的当前所有裂缝

	Log_INFO("当前符合订阅的裂缝: ")
	if len(allMatchingFissures) == 0 {
		Log_INFO("当前暂无订阅裂缝。")
	} else {
		for _, fissure := range allMatchingFissures {
			texts, _ := formatPrint(fissure)
			Log_INFO(texts)
		}
	}

	file, err = os.Create("data.json") // 修改文件名
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

// 修改 formatPrint，使用翻译函数
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
		fissure.Node,
		fissure.ETA,
	), nil
}

// 新增函数：将新裂缝格式化为HTML
func formatFissuresToHTML(fissures []Fissure) (string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
<style>
  body { font-family: Arial, sans-serif; margin: 20px; color: #333; }
  .container { border: 1px solid #ddd; padding: 20px; border-radius: 8px; max-width: 700px; margin: auto; }
  h1 { color: #0056b3; }
  .fissure-table { width: 100%; border-collapse: collapse; margin-top: 20px; }
  .fissure-table th, .fissure-table td { border: 1px solid #ddd; padding: 12px; text-align: left; }
  .fissure-table th { background-color: #f2f2f2; }
  .fissure-type { font-weight: bold; }
  .steel-path { color: #d9534f; font-weight: bold; }
  .void-storm { color: #5bc0de; font-weight: bold; }
</style>
</head>
<body>
  <div class="container">
    <h1>Warframe 裂缝订阅通知</h1>
    <p>您好！您订阅的以下裂缝任务已出现：</p>
    <table class="fissure-table">
      <thead>
        <tr>
          <th>任务</th>
          <th>地点</th>
          <th>纪元</th>
          <th>阵营</th>
          <th>剩余时间</th>
        </tr>
      </thead>
      <tbody>
        {{range .}}
        <tr>
          <td>
            <span class="fissure-type">
              {{if .IsHard}}<span class="steel-path">钢铁之路</span> {{end}}
              {{if .IsStorm}}<span class="void-storm">虚空风暴</span> {{end}}
              {{.MissionType | TranslateMissionType}}
            </span>
          </td>
          <td>{{.Node}}</td>
          <td>{{.Tier | TranslateTier}}</td>
          <td>{{.EnemyKey | TranslateFaction}}</td>
          <td>{{.ETA}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
    <p style="font-size: 12px; color: #888; margin-top: 20px;">
      这是由 Warframe 小助手自动发送的邮件。
    </p>
  </div>
</body>
</html>`

	// 注册可以在模板中使用的函数
	funcMap := template.FuncMap{
		"TranslateMissionType": TranslateMissionType,
		"TranslateFaction":     TranslateFaction,
		"TranslateTier":        TranslateTier,
	}

	tmpl, err := template.New("email").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, fissures); err != nil {
		return "", err
	}

	return body.String(), nil
}
