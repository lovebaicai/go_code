package main

import (
	"encoding/json"
	"log"
	"os"
	"rds_backup/util"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
)

type BackupInfo struct {
	Items            *BackupItems `json:"items"`
	TotalRecordCount int          `json:"totalRecordCount"`
}

type BackupItems struct {
	BackupItem *[2]Backup `json:"backup"`
}

type Backup struct {
	BackupStartTime           string
	BackupEndTime             string
	TotalBackupSize           string
	BackupStatus              string `json:"backupStatus"`
	BackupIntranetDownloadURL string `json:"backupIntranetDownloadURL"`
	BackupDownloadURL         string `json:"backupDownloadURL"`
}

func main() {
	log.Println("开始查询rds实例备份..")
	accountKey := os.Getenv("rds_key")
	accountSecret := os.Getenv("rds_secret")
	rdsId := os.Getenv("rds_id")

	client, err := rds.NewClientWithAccessKey("cn-shanghai", accountKey, accountSecret)
	if err != nil {
		panic(err)
	}

	request := rds.CreateDescribeBackupsRequest()
	request.StartTime = time.Now().AddDate(0, 0, -1).UTC().Format(time.RFC3339)[0:11] + "00:00Z"
	request.DBInstanceId = rdsId
	request.PageSize = "30"
	request.PageNumber = "1"
	request.BackupStatus = "Success"
	request.Scheme = "https"

	respJsonStr, err := client.DescribeBackups(request)
	if err != nil {
		log.Println("api获取备份文件错误:", err.Error())
		return
	}
	data := &BackupInfo{}
	if json.Unmarshal(respJsonStr.GetHttpContentBytes(), data) != nil {
		log.Println("json to struct error !")
		return
	}
	log.Println("开始下载rds实例备份..")
	if data.TotalRecordCount > 0 {
		downloadUrl := data.Items.BackupItem[0].BackupDownloadURL
		// downloadUrl := "https://down5.huorong.cn/sysdiag-all-5.0.67.2-2022.4.11.1.exe"
		backupDate := data.Items.BackupItem[0].BackupStartTime[0:10]
		log.Println("DB backup time: ", backupDate)
		log.Println("DB donwload url: ", downloadUrl)
		filePath := "/data/rds_backup/" + backupDate + ".tar.gz"
		log.Println("filepath: ", filePath)
		err := util.DownloadFile(filePath, downloadUrl)
		if err != nil {
			log.Println("下载失败:", err)
			return
		}
		log.Println("rds实例备份下载成功..", backupDate)
	}
}
