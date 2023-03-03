package job

import (
	"x-ui/logger"
	"x-ui/web/service"
)

type CheckBusinessStatusJob struct {
	businessService service.BusinessService
}

func NewCheckBusinessStatusJob() *CheckBusinessStatusJob {
	return new(CheckBusinessStatusJob)
}

func (j *CheckBusinessStatusJob) Run() {
	err := j.businessService.EenewBusinessInfo()
	if err != nil {
		logger.Error("更新业务信息失败：", err)
	}
}
