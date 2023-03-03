package job

import (
	"time"
	"x-ui/logger"
	"x-ui/web/service"
)

type DetectTrafficResetJob struct {
	xrayService    service.XrayService
	inboundService service.InboundService
	settingService service.SettingService
}

func TrafficResetJob() *DetectTrafficResetJob {
	return new(DetectTrafficResetJob)
}

func (j *DetectTrafficResetJob) Run() {
	logger.Info("今天是", time.Now().Day(), "号，正在检查是否重置流量")
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("获取配置失败:", err)
	}
	trafficResetDay := allSetting.TrafficResetDay
	if trafficResetDay == time.Now().Day() {
		count, err := j.inboundService.ResetTraffic()
		if err != nil {
			logger.Warning("Check reset flow error:", err)
		} else if count > 0 {
			logger.Info("reset %v users' traffic", count)
			j.xrayService.SetToNeedRestart()
		}
	} else {
		logger.Info("检查完毕，重置日是", trafficResetDay, "号，今天无须重置流量")
	}
}
