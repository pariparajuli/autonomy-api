package consts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/consts"
)

func TestTwCountyKey(t *testing.T) {
	mapping := map[string]string{
		"台北市": "taipei_city",
		"台中市": "taichung_city",
		"台南市": "tainan_city",
		"高雄市": "kaohsiung_city",
		"基隆市": "keelung_city",
		"新竹市": "hsinchu_city",
		"嘉義市": "chiayi_city",
		"新北市": "new_taipei_city",
		"桃園市": "taoyuan_city",
		"新竹縣": "hsinchu_county",
		"宜蘭縣": "yilan_county",
		"苗栗縣": "miaoli_county",
		"彰化縣": "changhua_county",
		"南投縣": "nantou_county",
		"雲林縣": "yunlin_county",
		"嘉義縣": "chiayi_county",
		"屏東縣": "pingtung_county",
		"澎湖縣": "penghu_county",
		"花蓮縣": "hualien_county",
		"台東縣": "taitung_county",
		"金門縣": "kinmen_county",
		"連江縣": "lianjiang_county",
	}

	for key, value := range mapping {
		actual, _ := consts.TwCountyKey(key)
		assert.Equal(t, value, actual, "wrong key")
	}
}
