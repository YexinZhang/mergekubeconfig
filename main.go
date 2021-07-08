package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type KubectlCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
}

type KubectlContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type KubectlUser struct {
	ClientCertificateDate string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type KubectlClusterWithName struct {
	Name    string         `yaml:"name"`
	Cluster KubectlCluster `yaml:"cluster"`
}

type KubectlContextWithName struct {
	Name    string         `yaml:"name"`
	Context KubectlContext `yaml:"context"`
}

type KubectlUserWithName struct {
	Name string      `yaml:"name"`
	User KubectlUser `yaml:"user"`
}

type yamlFile struct {
	ApiVersion     string                   `yaml:"apiVersion" json:"apiVersion"`
	Kind           string                   `yaml:"kind" json:"kind"`
	Clusters       []KubectlClusterWithName `yaml:"clusters" json:"clusters"`
	Context        []KubectlContextWithName `yaml:"contexts" json:"contexts"`
	Preferences    struct{}   				`yaml:"preferences" json:"preferences"`
	CurrentContext string                   `yaml:"current-context" json:"current-context"`
	Users          []KubectlUserWithName    `yaml:"users" json:"users"`
}

var (
	yamlFilelists []*yamlFile
	x             *string // 存放kubeconfig文件的路径
	_CurrentCtx   *string
	output 		  *string
)

func init() {
	x = flag.String("d","./","指定kubeconfig存放的目录")
	_CurrentCtx = flag.String("ctx","","指定kubeconfig的当前上下文")
	output = flag.String("o", "mergedkubeconfig","输出的文件名")
}

func main() {
	flag.Parse()
	readfromdir()
	bigYaml := Filetogether(yamlFilelists)
	_mergedyamlFile, err := yaml.Marshal(bigYaml); if err != nil {
		panic(err)
	}

	err = writeTofile(_mergedyamlFile)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func writeTofile(yamlfile []byte) error {
	err := ioutil.WriteFile(*output, yamlfile, fs.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

// 将文件解析到yamlFile结构体， 并且存在一个切片中
func readfromdir() {

	// 删除当前文件夹下已经存在的mergedkubeconfig
	_, err := os.Stat(*x + *output)
	if err == nil {
		_ = os.Remove(*x + *output)
	}
	// 读取当前文件夹下的所有文件
	y := filepath.Dir(*x)
	fmt.Println(y)
	files, err := ioutil.ReadDir(*x)
	if err != nil {
		fmt.Println(err)
		return
	}

	var filelists []string
	// 如果程序放在通kubeconfig一个文件夹， 那么要排除这个文件
	execpath, err  := exec.LookPath(os.Args[0]); if err != nil { return }
	for _, f := range files {
		if f.Name() == filepath.Base(execpath) {
			continue
		}
		filelists = append(filelists, f.Name())
	}
	
	// 存入slice
	for _, f := range filelists {
		filep := *x + f
		yamlfile, err := os.Open(filep)
		if err != nil {
			fmt.Println(err)
			continue
		}
		yamlstruct := &yamlFile{}
		err = yaml.NewDecoder(yamlfile).Decode(yamlstruct)
		if err != nil {
			fmt.Println("Yaml decoder error: ", err)
			continue
		}
		yamlFilelists = append(yamlFilelists, yamlstruct)
		yamlfile.Close()
	}
}


//将解析到的yamlstruct聚合到一个yamlFile struct中

func Filetogether(list []*yamlFile) *yamlFile {
	bigYaml := &yamlFile{
		ApiVersion: "v1",
		Kind: "Config",
		Preferences: struct{}{},
		Context: make([]KubectlContextWithName,0),
		Clusters: make([]KubectlClusterWithName,0),
		Users: make([]KubectlUserWithName,0),
	}

	for _, v := range list {
		for _, c := range v.Clusters {
			bigYaml.Clusters = append(bigYaml.Clusters, c)
		}

		for _, u := range v.Users {
			bigYaml.Users = append(bigYaml.Users,u)
		}

		for _, ctx := range v.Context {
			bigYaml.Context = append(bigYaml.Context, ctx)
		}
	}
	// 配置current-context
	if *_CurrentCtx == "" {
		bigYaml.CurrentContext = list[0].Context[0].Name
	} else {
		bigYaml.CurrentContext = *_CurrentCtx
	}
	return bigYaml
}
