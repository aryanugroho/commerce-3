package cron

import (
	"github.com/gorhill/cronexpr"
	"github.com/ottemo/foundation/env"
	"time"
)

// Package global constants
const (
	ConstErrorModule = "env/cron"
	ConstErrorLevel  = env.ConstErrorLevelService
)

// DefaultCronScheduler is a default implementer of InterfaceIniConfig
type DefaultCronScheduler struct {
	tasks     map[string]env.FuncCronTask
	schedules []StructCronSchedule

	appStarted bool
}

// StructCronSchedule structure to hold schedule information (for internal usage)
type StructCronSchedule struct {
	CronExpr string
	TaskName string
	Params   map[string]interface{}
	Repeat   bool
	Time	 time.Time

	task env.FuncCronTask
	expr *cronexpr.Expression
}
