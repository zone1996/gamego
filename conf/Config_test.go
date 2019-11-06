package conf

import "testing"
import "encoding/json"
import "fmt"

func TestConfig(t *testing.T) {
	// marshal
	dbconfig := make(map[string]interface{})
	dbconfig["host"] = "localhost"
	dbconfig["port"] = "3306"
	dbconfig["password"] = "123456"

	redisconfig := make(map[string]interface{})
	redisconfig["host"] = "localhost"
	redisconfig["port"] = "6379"
	redisconfig["password"] = "123456"

	config := &Config{
		ServerName:  "第1区",
		Port:        "12345",
		AdminPort:   "12346",
		DbConfig:    dbconfig,
		RedisConfig: redisconfig,
	}

	bytes, err := json.Marshal(config)
	if err == nil {
		fmt.Println(string(bytes))
	} else {
		t.Errorf("error:%v", err)
	}

	// unmarshal
	unmarshaledConfig := &Config{}
	data := `
	{
		"ServerName":"第1区",
		"Port":"12345",
		"AdminPort":"12346",
		"DbConfig":
		{
			"host":"localhost",
			"password":"123456",
			"port":"3306"
		},
		"RedisConfig":
		{
			"host":"localhost",
			"password":"123456",
			"port":"6379"
		}
	}`

	err = json.Unmarshal([]byte(data), unmarshaledConfig)
	if err != nil {
		t.Errorf("error:%v", err)
	}

	if unmarshaledConfig.Port != "12345" || unmarshaledConfig.RedisConfig["port"] != "6379" {
		t.Errorf("unmarshal failed")
	} else {
		// fmt.Println(unmarshaledConfig)
	}
}

func TestInit(t *testing.T) {
	filepath := "config.conf"
	err := Init(filepath)
	if err == nil {
		fmt.Println(GetConfig())
	} else {
		t.Errorf("error:%v", err)
	}
}
