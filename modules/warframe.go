package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var platform string = "pc" // Example platform
var params map[string]string = map[string]string{
	"language": "zh",
}

// 订阅裂缝的结构体
type SubFissure struct {
	MissionType string
	IsHard      bool
	Tier        string
	Location    string
}

// 主结构体
type Warframe struct {
	SubsFissures []SubFissure
}

// Fissure represents the structure of the fissure data
type Fissure struct {
	ID          string `json:"id"`
	MissionType string `json:"missionType"`
	EnemyKey    string `json:"enemyKey"`
	Tier        string `json:"tier"`
	Node        string `json:"node"`
	ETA         string `json:"eta"`
	IsHard      bool   `json:"isHard"`
	IsStorm     bool   `json:"isStorm"`
	Active      bool   `json:"active"`
}

// GetFissuresWithRetry fetches fissures with retry logic
func GetFissuresWithRetry(f Warframe, maxRetries int, delay time.Duration) ([]byte, error) {
	url := fmt.Sprintf("https://api.warframestat.us/%s/fissures", platform)
	client := &http.Client{}

	// Construct the URL with query parameters
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	// Retry logic
	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				Log(fmt.Sprint("请求失败 (尝试 %d/%d): %v\n", attempt+1, maxRetries, err), ERROR)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("请求失败: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 && res.StatusCode < 600 {
			if attempt < maxRetries-1 {
				Log(fmt.Sprint("请求失败 (状态码 %d) (尝试 %d/%d)\n", res.StatusCode, attempt+1, maxRetries), ERROR)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("请求失败，状态码: %d", res.StatusCode)
		}

		// Read and return the response body
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}
		return body, nil
	}

	// Max retries reached
	return nil, fmt.Errorf("重试次数达到最大限制，退出。")
}

// 获取裂缝原始数据
func GetRawFissures(f Warframe) ([]Fissure, error) {
	// Get fissure data with retry
	data, err := GetFissuresWithRetry(f, 9999, 2*time.Second)
	if err != nil {
		return nil, err
	}

	// Convert data from byte slice to string and then to JSON
	var fissures []Fissure
	if err := json.Unmarshal(data, &fissures); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	return fissures, err
}

// GetFissures retrieves fissure data and processes it based on type
func GetFissures(f Warframe, type_ int) ([]map[string]string, error) {
	// Get fissure data with retry
	data, err := GetFissuresWithRetry(f, 9999, 2*time.Second)
	if err != nil {
		return nil, err
	}

	// Convert data from byte slice to string and then to JSON
	var fissures []Fissure
	if err := json.Unmarshal(data, &fissures); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	// Filter and categorize fissures based on conditions
	var dataList, dataListRegular, dataListStorm, dataListSteel []map[string]string

	for _, fissure := range fissures {
		if fissure.Active {
			// Group fissures by type
			fissureData := map[string]string{
				"id":       fissure.ID,
				"type":     fissure.MissionType,
				"level":    "",
				"title":    "",
				"subtitle": "",
				"location": fissure.Node,
				"time":     fissure.ETA,
			}
			if fissure.IsHard {
				dataListSteel = append(dataListSteel, fissureData)
			} else if fissure.IsStorm {
				dataListStorm = append(dataListStorm, fissureData)
			} else {
				dataListRegular = append(dataListRegular, fissureData)
			}
		}
	}

	// Add relevant fissures to dataList based on requested type
	if type_ == 0 || type_ == 1 {
		for _, fissure := range dataListSteel {
			fissure["level"] = "steel"
			fissure["title"] = fmt.Sprintf("钢铁之路 %s - %s", fissure["type"], fissure["type"])
			fissure["subtitle"] = fmt.Sprintf("%s 裂缝", fissure["level"])
			dataList = append(dataList, fissure)
		}
	}
	if type_ == 0 || type_ == 2 {
		for _, fissure := range dataListRegular {
			fissure["level"] = "regular"
			fissure["title"] = fmt.Sprintf("%s - %s", fissure["type"], fissure["type"])
			fissure["subtitle"] = fmt.Sprintf("%s 裂缝", fissure["level"])
			dataList = append(dataList, fissure)
		}
	}
	if type_ == 0 || type_ == 3 {
		for _, fissure := range dataListStorm {
			fissure["level"] = "storm"
			fissure["title"] = fmt.Sprintf("虚空风暴 %s - %s", fissure["type"], fissure["type"])
			fissure["subtitle"] = fmt.Sprintf("%s 裂缝", fissure["level"])
			dataList = append(dataList, fissure)
		}
	}

	// Return the final list
	return dataList, nil
}

// 发送邮件
func SendSubsFissures(f Warframe) error {
	data, err := GetRawFissures(f)
	if err != nil {
		return err
	}
	//检查已有裂缝
	for _, fissure := range data {
		if fissure.Active {
			for _, subsfissure := range f.SubsFissures {
				if subsfissure.IsHard == fissure.IsHard && subsfissure.MissionType == fissure.MissionType && subsfissure.Tier == fissure.Tier {

				}
			}
		}
	}
	return nil
}

// 检查已有裂缝
func CheckSubFissure(f Warframe) error {

	type existsID struct {
		ID []string `json:"id"`
	}

	// 读取 已存在裂缝ID 文件
	file, err := os.Open("exists.json")
	if err != nil {
		Log(fmt.Sprint("Error opening file:", err), ERROR)
		return err
	}
	defer file.Close()

	// 创建一个结构体实例来存储读取的 JSON 数据
	var ids existsID
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ids)
	if err != nil {
		Log(fmt.Sprint("Error decoding JSON:", err), ERROR)
		return err
	}

	// 输出当前已存在的裂缝
	for _, id := range ids.ID {
		Log_INFO("当前已存在ID: " + id)
	}

	data, err := GetRawFissures(f)
	if err != nil {
		return err
	}
	var newFissures []string = []string{"default"}
	//检查已有裂缝
	for _, fissure := range data {
		if fissure.Active {
			for _, subsfissure := range f.SubsFissures {
				if subsfissure.IsHard == fissure.IsHard && subsfissure.MissionType == fissure.MissionType && subsfissure.Tier == fissure.Tier {
					for _, s := range ids.ID {
						if s == fissure.ID {
							newFissures = append(newFissures, fissure.ID)
						}
					}
					newFissures = append(newFissures, fissure.ID)

				}
			}
		}
	}
	ids.ID = newFissures
	// 更新本地json文件
	file, err = os.Create("exists.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()
	return nil
}
