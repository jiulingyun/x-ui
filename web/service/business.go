package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"x-ui/database/model"
	"x-ui/logger"

	"github.com/google/uuid"
)

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Id            int    `json:"id"`
		Ip            string `json:"ip"`
		ClientName    string `json:"client_name"`
		AreaName      string `json:"area_name"`
		PanelClientId int    `json:"panel_client_id"`
		PanelAreaId   int    `json:"panel_area_id"`
		OpenedDate    string `json:"opened_date"`
	}
}

type PostResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time string `json:"time"`
}

type BusinessService struct {
	settingService SettingService
	xrayService    XrayService
	inboundService InboundService
}

/**
获取入站信息
*/
func (j *BusinessService) GetBusinessInfo() (*Result, error) {
	data := &Result{}
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("获取全部配置信息失败", err)
	}
	if allSetting.ApiUrl != "" && allSetting.ApiKey != "" && allSetting.BusinessId > 0 {
		url := allSetting.ApiUrl + "/panel/getNode?token=" + allSetting.ApiKey + "&id=" + strconv.Itoa(allSetting.BusinessId)
		resp, err := http.Get(url)
		if err != nil {
			logger.Warning("http请求错误", err)
			return nil, err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Warning("http的body解析错误", err)
			return nil, err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Warning("json解析失败:", err)
			return nil, err
		}

	}
	return data, err
}

/**
推送节点链接
*/
func (j *BusinessService) PullNodeLink(inbound *model.Inbound) error {
	data := &PostResult{}
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("获取全部配置信息失败", err)
	}

	jsonstr, err := json.Marshal(inbound)

	qinqibody := "id=" + strconv.Itoa(inbound.Id) + "&node=" + string(jsonstr)

	if allSetting.ApiUrl != "" && allSetting.ApiKey != "" && allSetting.BusinessId > 0 {
		url := allSetting.ApiUrl + "/panel/setLink?token=" + allSetting.ApiKey
		resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(qinqibody))
		if err != nil {
			logger.Warning("链接推送-http请求错误", err)
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Warning("链接推送-http的body解析错误", err)
			return err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Warning("链接推送-json解析失败:", err)
			return err
		}

		if data.Code == 1 {
			logger.Info("推送节点连接成功！")
		} else {
			logger.Error("节点推送失败，服务器返回错误：" + data.Msg)
		}
	}

	return err

}

//推送在线状态
func (j *BusinessService) NodeStatus(inbound *model.Inbound) error {
	data := &PostResult{}
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("获取全部配置信息失败", err)
	}

	qinqibody := "id=" + strconv.Itoa(inbound.Id) + "&load=0,0,0&total=" + strconv.FormatInt(inbound.Total, 10) + "&down=" + strconv.FormatInt(inbound.Down, 10) + "&up=" + strconv.FormatInt(inbound.Up, 10) + "&enable=" + strconv.FormatBool(inbound.Enable) + "&xray=" + strconv.FormatBool(j.xrayService.IsXrayRunning())

	if allSetting.ApiUrl != "" && allSetting.ApiKey != "" && allSetting.BusinessId > 0 {
		url := allSetting.ApiUrl + "/panel/nodeStatus?token=" + allSetting.ApiKey
		resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(qinqibody))
		if err != nil {
			logger.Warning("节点状态-http请求错误", err)
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Warning("节点状态-http的body解析错误", err)
			return err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Warning("节点状态-json解析失败:", err)
			return err
		}

		if data.Code == 1 {
			logger.Info("节点状态推送成功！")
		} else {
			logger.Error("节点状态推送失败，服务器返回错误：" + data.Msg)
		}
	}

	return err
}

func (j *BusinessService) EenewBusinessInfo() error {
	data, err := j.GetBusinessInfo()
	if err != nil {
		logger.Error("获取业务信息错误：", err)
		return err
	}

	//将到期时间转换成时间戳
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")

	if data.Code == 1 {
		inbo, err := j.inboundService.GetAllInbounds()
		if err != nil {
			logger.Error("获取入站信息失败：", err)
			return err
		}

		//自动设置流量重置日
		startTime, _ := time.ParseInLocation(timeLayout, data.Data.OpenedDate, loc)
		j.settingService.SetTrafficResetDay(startTime.Day())

		//strArr := strings.Split(data.Data.Flow, "G")
		guid := uuid.New()
		//默认参数
		settings := "{\n  \"clients\": [\n    {\n      \"id\": \"id-key\",\n      \"flow\": \"xtls-rprx-direct\"\n    }\n  ],\n  \"decryption\": \"none\",\n  \"fallbacks\": []\n}"
		streamSettings := "{\n  \"network\": \"ws\",\n  \"security\": \"none\",\n  \"wsSettings\": {\n    \"acceptProxyProtocol\": false,\n    \"path\": \"/jiulingyun\",\n    \"headers\": {}\n  }\n}"
		sniffing := "{\n  \"enabled\": true,\n  \"destOverride\": [\n    \"http\",\n    \"tls\"\n  ]\n}"
		settings = strings.Replace(settings, "id-key", guid.String(), 1)
		//total, _ := strconv.Atoi(strArr[0])

		//检查入站是否添加，如果未添加则新增一个入站
		if len(inbo) < 1 {
			inbound := &model.Inbound{}
			inbound.UserId = 1
			inbound.Enable = true
			inbound.Down = 0
			inbound.ExpiryTime = 0
			inbound.Id = data.Data.Id
			inbound.Port = 80
			inbound.Protocol = model.VMess
			inbound.Remark = data.Data.AreaName + "-" + strconv.Itoa(data.Data.Id)
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
			j.PullNodeLink(inbound)
		} else {
			//这里开始推送在线状态
			inbound, err := j.inboundService.GetInbound(data.Data.Id)
			if err != nil {
				inbound, err := j.inboundService.GetInbound(1)
				if err != nil {
					logger.Error("获取入站失败：", err)
					return err
				}
			}
			j.NodeStatus(inbound)
		}

	}

	return nil
}
