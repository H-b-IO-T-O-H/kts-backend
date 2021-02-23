package usecase

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/timetable"
)

type TimetableUseCase struct {
	repos timetable.RepositoryTimetable
}

func (t TimetableUseCase) GetTimetableWeek(groupName string, weekNumber int) (*models.Week, common.Err) {
	return t.repos.GetTimetableWeek(groupName, weekNumber)
}

func NewTimetableUseCase(repos timetable.RepositoryTimetable) timetable.UseCase {
	return TimetableUseCase{
		repos: repos,
	}
}

func (t TimetableUseCase) CreateUpdateWeek(newWeek models.Week) common.Err {
	return t.repos.CreateUpdateWeek(newWeek)
}

func (t TimetableUseCase) DeleteTimetableWeek(weekInfo models.DeleteEntity) common.Err {
	return t.repos.DeleteTimetableWeek(weekInfo)
}
