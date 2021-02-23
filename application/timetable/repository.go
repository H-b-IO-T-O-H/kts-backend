package timetable

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
)

type RepositoryTimetable interface {
	CreateUpdateWeek(newWeek models.Week) common.Err
	GetTimetableWeek(groupName string, weekNumber int) (*models.Week, common.Err)
	DeleteTimetableWeek(weekInfo models.DeleteEntity) common.Err
}