package tool

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var ConfEnvPath string //配置文件夹
var ConfEnv string     //配置文件名

//解析配置文件目录

//配置文件必须放到一个文件夹中
//如：config = conf/dev/base.json  ConfEnvPath=conf/dev   ConfEnv=dev
//如：config = conf/base.json      ConfEnvPath=conf       ConfEnv=conf

func ParseConPath(config string) error {
	path := strings.Split(config, "/")
	length := len(path)
	prefix := strings.Join(path[:length-1], "/")

	ConfEnvPath = prefix
	ConfEnv = path[length-2]
	return nil
}

//获取配置环境名称
func GetConfEnv() string {
	return ConfEnv
}

//获取配置文件完整路径
func GetConfPath(fileName string) string {
	return ConfEnvPath + "/" + fileName + ".toml"
}

func GetConfFilePath(fileName string) string {
	return ConfEnvPath + "/" + fileName
}

//解析配置文件
func ParseLocalConfig(fileName string, conf interface{}) error {
	path := GetConfFilePath(fileName)
	err := ParseConfig(path, conf)
	if err != nil {
		return err
	}
	return nil
}

//解析配置
func ParseConfig(path string, conf interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Open config %v failed,errMessge:%v", path, err)
	}
	//读取文件
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Read config failed,errMessage:%v", err)
	}
	v := viper.New()
	//设置配置类型
	v.SetConfigType("toml")
	v.ReadConfig(bytes.NewBuffer(data))
	if err := v.Unmarshal(conf); err != nil {
		return fmt.Errorf("Parsing config failed ,config:%v,err:%v", string(data), err)
	}
	return nil

}
