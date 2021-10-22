package icinga2

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/taosdata/blm3/config"
	"github.com/taosdata/driver-go/v2/af"
)

func TestGatherServicesStatus(t *testing.T) {
	s := `{
    "results": [
        {
            "attrs": {
                "check_command": "c_c",
                "display_name": "d_n",
                "host_name": "h_n",
                "name": "n",
                "state": 0
            },
            "joins": {},
            "meta": {},
            "name": "name",
            "type": "Service"
        }
    ]
}
`

	checks := Result{}
	require.NoError(t, json.Unmarshal([]byte(s), &checks))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	}))

	defer ts.Close()
	api := ts.URL
	u, _ := url.Parse(ts.URL)
	config.Init()
	viper.Set("icinga2.enable", true)
	viper.Set("icinga2.host", api)
	viper.Set("icinga2.httpUsername", "root")
	viper.Set("icinga2.httpPassword", "9fba18a59df3d9f1")
	icinga2 := new(Icinga2)
	err := icinga2.Init(nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = icinga2.GatherStatus("services", icinga2.request["services"])
	if err != nil {
		t.Error(err)
		return
	}
	conn, err := af.Open("", "", "", "icinga2", 0)
	if err != nil {
		t.Error(err)
		return
	}
	rows, err := conn.Query("select * from services order by ts desc limit 1")
	if err != nil {
		t.Error(err)
		return
	}
	defer rows.Close()
	k := rows.Columns()
	v := make([]driver.Value, len(k))
	mapV := make(map[string]interface{}, len(k))
	err = rows.Next(v)
	if err != nil {
		t.Error(err)
		return
	}
	for index, name := range k {
		mapV[name] = v[index]
	}
	expectTags := map[string]string{
		"display_name":  "d_n",
		"check_command": "c_c",
		"_state":        "ok",
		"source":        "h_n",
		"server":        u.Hostname(),
		"port":          u.Port(),
		"scheme":        "http",
	}
	for tn, tv := range expectTags {
		vv, exist := mapV[tn]
		if !exist {
			t.Errorf("tag not exist, %s", tn)
			return
		}
		if vv.(string) != tv {
			t.Errorf("tag expect %s ,got %s", tv, vv.(string))
			return
		}
	}
	state, exist := mapV["state_code"]
	if !exist {
		t.Errorf("column state_code not exist")
		return
	}
	if state.(int64) != 0 {
		t.Errorf("column state_code expect 0 got %d", state.(int64))
		return
	}
	name, exist := mapV["name"]
	if !exist {
		t.Errorf("column name not exist")
		return
	}
	if name.(string) != "n" {
		t.Errorf("column name expect n got %s", name.(string))
		return
	}
}
