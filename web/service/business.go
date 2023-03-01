package service

import (
	"fmt"
	"os"
	"strings"
	"x-ui/database/model"
	"x-ui/logger"

	"github.com/google/uuid"
)

type Result struct {
	Result  bool   `json:"result"`
	Message string `json:"msg"`
	Data    struct {
		Id      string `json:"id"`
		Guid    string `json:"guid"`
		UserId  string `json:"user_id"`
		OrderId string `json:"order_id"`
		Name    string `json:"name"`
		BuyTime string `json:"buy_time"`
		EndTime string `json:"end_time"`
		Status  string `json:"status"`
		Area    string `json:"area"`
		Flow    string `json:"flow"`
	}
}

type BusinessService struct {
	settingService SettingService
	xrayService    XrayService
	inboundService InboundService
}

func (j *BusinessService) EenewBusinessInfo() error {
	inbo, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Error("获取入站信息失败：", err)
		return err
	}

	hostname, err := os.Hostname()

	guid := uuid.New()
	//默认参数
	settings := "{\n  \"clients\": [\n    {\n      \"id\": \"id-key\",\n      \"flow\": \"xtls-rprx-direct\"\n    }\n  ],\n  \"decryption\": \"none\",\n  \"fallbacks\": []\n}"
	streamSettings := "{\n  \"network\": \"ws\",\n  \"security\": \"none\",\n  \"wsSettings\": {\n    \"acceptProxyProtocol\": false,\n    \"path\": \"/jiulingyun\",\n    \"headers\": {}\n  }\n}"
	sniffing := "{\n  \"enabled\": true,\n  \"destOverride\": [\n    \"http\",\n    \"tls\"\n  ]\n}"
	settings = strings.Replace(settings, "id-key", guid.String(), 1)

	//检查入站是否添加，如果未添加则新增一个入站
	if len(inbo) < 1 {
		inbound := &model.Inbound{}
		inbound.UserId = 1
		inbound.Enable = true
		inbound.Down = 0
		//inbound.ExpiryTime = ''
		inbound.Id = 1
		inbound.Port = 80
		inbound.Protocol = model.VMess
		inbound.Remark = hostname
		inbound.Settings = settings
		inbound.Sniffing = sniffing
		inbound.StreamSettings = streamSettings
		inbound.Total = 0
		inbound.Up = 0
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)

		err := j.inboundService.AddInbound(inbound)
		if err != nil {
			logger.Error("添加入站失败：", err)
			return err
		}

		if err == nil {
			j.xrayService.SetToNeedRestart()
		}

		logger.Info("添加入站成功")

	}

	return nil
}
