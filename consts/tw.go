package consts

import (
	"fmt"
	"strings"
)

var TwCountyEnglish map[string]string

func init() {
	TwCountyEnglish = make(map[string]string)

	TwCountyEnglish["台北市"] = "Taipei City"
	TwCountyEnglish["台中市"] = "Taichung City"
	TwCountyEnglish["台南市"] = "Tainan City"
	TwCountyEnglish["高雄市"] = "Kaohsiung City"
	TwCountyEnglish["基隆市"] = "Keelung City"
	TwCountyEnglish["新竹市"] = "Hsinchu City"
	TwCountyEnglish["嘉義市"] = "Chiayi City"
	TwCountyEnglish["新北市"] = "New Taipei City"
	TwCountyEnglish["桃園市"] = "Taoyuan City"
	TwCountyEnglish["新竹縣"] = "Hsinchu County"
	TwCountyEnglish["宜蘭縣"] = "Yilan County"
	TwCountyEnglish["苗栗縣"] = "Miaoli County"
	TwCountyEnglish["彰化縣"] = "Changhua County"
	TwCountyEnglish["南投縣"] = "Nantou County"
	TwCountyEnglish["雲林縣"] = "Yunlin County"
	TwCountyEnglish["嘉義縣"] = "Chiayi County"
	TwCountyEnglish["屏東縣"] = "Pingtung County"
	TwCountyEnglish["澎湖縣"] = "Penghu County"
	TwCountyEnglish["花蓮縣"] = "Hualien County"
	TwCountyEnglish["台東縣"] = "Taitung County"
	TwCountyEnglish["金門縣"] = "Kinmen County"
	TwCountyEnglish["連江縣"] = "Lianjiang County"
}

// TwCountyKey - convert chinese tw county name into key
func TwCountyKey(county string) (string, error) {
	if zh, ok := TwCountyEnglish[county]; !ok {
		return "", fmt.Errorf("%s not exist", county)
	} else {
		return strings.Replace(strings.ToLower(zh), " ", "_", -1), nil
	}
}
