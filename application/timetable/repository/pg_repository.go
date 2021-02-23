package repository

import (
	"fmt"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/timetable"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type pgStorage struct {
	db *gorm.DB
}

func NewPgRepository(db *gorm.DB) timetable.RepositoryTimetable {
	return &pgStorage{db: db}
}

func (p pgStorage) CreateUpdateWeek(newWeek models.Week) common.Err {
	var err error
	timetableOld := models.Timetable{GroupName: newWeek.GroupName}
	group := models.Group{GroupName: newWeek.GroupName}

	err = p.db.Raw("select group_id, timetable_id from public.groups where group_name = ?", newWeek.GroupName).
		Row().Scan(&timetableOld.GroupId, &timetableOld.TimetableId)
	if err != nil {
		return common.RespErr{Status: http.StatusNotFound, Message: err.Error()}
	}

	if timetableOld.TimetableId == uuid.Nil {
		if err := p.db.Create(&timetableOld).Error; err != nil {
			return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
		}
		group.TimetableId = timetableOld.TimetableId
		group.GroupId = timetableOld.GroupId
		if err1 := p.db.Model(&group).Where("group_id = ?", group.GroupId).
			Update("timetable_id", group.TimetableId.String()).Error; err1 != nil {
			return common.RespErr{Status: http.StatusInternalServerError, Message: err1.Error()}
		}
	}

	if newWeek.WeekId == uuid.Nil {
		err = p.db.Raw(fmt.Sprintf("select week%d_id from public.timetables where timetable_id = '%s'",
			newWeek.WeekNumber, timetableOld.TimetableId)).Row().Scan(&newWeek.WeekId)
	}

	if err = p.insertDataInTimetable(timetableOld, newWeek); err != nil {
		return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
	}

	return nil
}

func (p pgStorage) insertDataInTimetable(newTimetable models.Timetable, newWeek models.Week) error {
	var err error
	var weekId = newWeek.WeekId

	tx := p.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}


	for j := 0; j < len(newWeek.Days); j++ {
		day := newWeek.Days[j]
		for k := 0; k < len(day.Lessons); k++ {
			lesson := day.Lessons[k]

			_ = tx.Raw(fmt.Sprintf("select day%d_id from public.weeks where week_id='%s'", day.DayOrder, weekId)).Row().Scan(&day.DayId)
			if day.DayId == uuid.Nil {
				err = tx.Create(&lesson).Error
				if err != nil {
					tx.Rollback()
					return err
				}

				q := fmt.Sprintf("insert into public.days(lesson%d_id) values ('%s') returning day_id", lesson.LessonOrder, lesson.LessonId)
				err = tx.Raw(q).Row().Scan(&day.DayId)

				if err != nil {
					tx.Rollback()
					return err
				}

				err = tx.Exec(fmt.Sprintf("update public.weeks set day%d_id='%s' where week_id='%s'", day.DayOrder, day.DayId, weekId)).Error
				if err != nil {
					tx.Rollback()
					return err
				}

			} else {
				id := uuid.Nil
				_ = tx.Raw(fmt.Sprintf("select lesson%d_id from public.days where day_id='%s'", lesson.LessonOrder, day.DayId)).Row().Scan(&id)
				fmt.Println(id)
				fmt.Println(lesson.LessonId)

				if id == uuid.Nil {

					id = lesson.LessonId
					lesson.LessonId = uuid.Nil

					err = tx.Create(&lesson).Error
					if err != nil {
						tx.Rollback()
						return err
					}
					err = tx.Exec(fmt.Sprintf("update public.days set lesson%d_id='%s' where day_id='%s'", lesson.LessonOrder, lesson.LessonId, day.DayId)).Error
					if err != nil {
						tx.Rollback()
						return err
					}
					if id != uuid.Nil {
						fmt.Println(id)
						fmt.Println(lesson.LessonId)
						lessonOld := models.Lesson{LessonId: id}
						_ = tx.Take(&lessonOld).Error
						if lessonOld.Title == lesson.Title && lessonOld.LessonType == lesson.LessonType && lessonOld.Auditorium == lesson.Auditorium {
							err = tx.Exec("delete from public.lessons where lesson_id = ?", id).Error
							if err != nil {
								tx.Rollback()
								return err
							}
						}
					}
				} else {
					lesson.LessonId = id
				}
			}

			//TODO:check update real need?
			q := fmt.Sprintf("update public.lessons set title='%s', auditorium='%s', lesson_type='%s', comment='%s' where lesson_id = '%s'",
				lesson.Title, lesson.Auditorium, lesson.LessonType, lesson.Comment, lesson.LessonId)

			err = tx.Exec(q).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		if newWeek.WeekId == uuid.Nil {
			q := fmt.Sprintf("insert into public.weeks(day%d_id) values ('%s') returning week_id", day.DayOrder, day.DayId)
			err = tx.Raw(q).Row().Scan(&weekId)
		}

		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if newWeek.WeekId == uuid.Nil {
		newWeek.WeekId = weekId
		q := fmt.Sprintf("update public.timetables set week%d_id='%s' where timetable_id='%s'", newWeek.WeekNumber, weekId, newTimetable.TimetableId)
		err = tx.Exec(q).Error
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (p pgStorage) GetTimetableWeek(groupName string, weekNumber int) (*models.Week, common.Err) {
	groupId := uuid.Nil

	err := p.db.Raw("select group_id from public.groups where group_name = ?", groupName).Row().Scan(&groupId)
	if err != nil {
		msg := err.Error()
		if common.NoRows(msg) {
			return nil, common.RespErr{Status: http.StatusNotFound, Message: common.NotFound}
		}
		return nil, common.RespErr{Status: http.StatusInternalServerError, Message: msg}
	}

	week := models.Week{GroupName: groupName, WeekNumber: weekNumber}
	dayIds := make([]uuid.UUID, 7)
	err = p.db.Raw(fmt.Sprintf("select w.day1_id, w.day2_id, w.day3_id, w.day4_id, w.day5_id, w.day6_id, "+
		"w.day7_id, w.week_type, w.week_id from  public.weeks w join timetables t on w.week_id = t.week%d_id where t.group_id = '%s'",
		weekNumber, groupId)).Row().Scan(&dayIds[0], &dayIds[1], &dayIds[2], &dayIds[3], &dayIds[4], &dayIds[5],
		&dayIds[6], &week.WeekType, &week.WeekId)
	if err != nil {
		if common.NoRows(err.Error()) {
			return nil, common.RespErr{Status: http.StatusNotFound, Message: common.NotFound}
		}
		return nil, common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
	}
	var oldTimetable = new(models.Timetable)
	oldTimetable.Weeks = []models.Week{}
	for i := 0; i < 7; i++ {
		if dayIds[i] != uuid.Nil {
			lessonsIds := make([]uuid.UUID, 8)
			err = p.db.Raw("select d.lesson1_id, d.lesson2_id, d.lesson3_id, d.lesson4_id, d.lesson5_id, d.lesson6_id, d.lesson7_id, d.lesson8_id from public.days d where d.day_id = ?",
				dayIds[i]).Row().Scan(&lessonsIds[0], &lessonsIds[1], &lessonsIds[2], &lessonsIds[3], &lessonsIds[4], &lessonsIds[5],
				&lessonsIds[6], &lessonsIds[7])
			if err != nil {
				return nil, common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
			}
			day := models.Day{DayId: dayIds[i], DayOrder: i}
			for j := 0; j < 8; j++ {
				if lessonsIds[j] != uuid.Nil {
					lesson := models.Lesson{LessonId: lessonsIds[j], LessonOrder: j}
					err = p.db.Raw("select l.title, l.auditorium, l.lesson_type, l.comment from public.lessons l where l.lesson_id = ?", lesson.LessonId).Row().
						Scan(&lesson.Title, &lesson.Auditorium, &lesson.LessonType, &lesson.Comment)
					if err != nil {
						return nil, common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
					}
					day.Lessons = append(day.Lessons, lesson)
				}
			}
			week.Days = append(week.Days, day)
		}
	}

	return &week, nil
}

func (p pgStorage) DeleteTimetableWeek(weekInfo models.DeleteEntity) common.Err {
	var err error
	var errDelete common.Err

	tx := p.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
	}

	if weekInfo.WeekId != uuid.Nil {
		errDelete = deleteWeek(tx, weekInfo.WeekId)
	}
	if weekInfo.DaysIds != nil && len(weekInfo.DaysIds) > 0 {
		errDelete = deleteDays(tx, weekInfo.DaysIds)
	}
	if weekInfo.LessonsIds != nil && len(weekInfo.LessonsIds) > 0 {
		errDelete = deleteLessons(tx, weekInfo.LessonsIds)
	}
	if errDelete != nil {
		return errDelete
	}
	err = tx.Commit().Error
	if err != nil {
		return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
	}
	return nil
}

func deleteWeek(tx *gorm.DB, weekId uuid.UUID) common.Err {
	var err error
	dayIds := make([]uuid.UUID, 7)
	errInternal := common.RespErr{Status: http.StatusInternalServerError}

	err = tx.Raw(fmt.Sprintf("select w.day1_id, w.day2_id, w.day3_id, w.day4_id, w.day5_id, w.day6_id, "+
		"w.day7_id from  public.weeks w where w.week_id ='%s'", weekId)).Row().Scan(&dayIds[0],
		&dayIds[1], &dayIds[2], &dayIds[3], &dayIds[4], &dayIds[5], &dayIds[6])
	if err != nil {
		msg := err.Error()
		if common.NoRows(err.Error()) {
			return common.RespErr{Status: http.StatusNotFound, Message: common.NotFound}
		}
		errInternal.Message = msg
		return errInternal
	}

	if err := deleteDays(tx, dayIds); err != nil {
		return err
	}

	err = tx.Exec("delete from public.weeks where week_id = ?", weekId).Error
	if err != nil {
		tx.Rollback()
		errInternal.Message = err.Error()
		return errInternal
	}

	return nil
}

func deleteDays(tx *gorm.DB, daysIds []uuid.UUID) common.Err {
	var err error
	errInternal := common.RespErr{Status: http.StatusInternalServerError}

	for i := 0; i < 7; i++ {
		if daysIds[i] != uuid.Nil {
			lessonsIds := make([]uuid.UUID, 8)
			err = tx.Raw("select d.lesson1_id, d.lesson2_id, d.lesson3_id, d.lesson4_id, d.lesson5_id, d.lesson6_id,"+
				" d.lesson7_id, d.lesson8_id from public.days d where d.day_id = ?",
				daysIds[i]).Row().Scan(&lessonsIds[0], &lessonsIds[1], &lessonsIds[2], &lessonsIds[3], &lessonsIds[4], &lessonsIds[5],
				&lessonsIds[6], &lessonsIds[7])
			if err != nil {
				errInternal.Message = err.Error()
				return errInternal
			}

			if err := deleteLessons(tx, lessonsIds); err != nil {
				return err
			}

			err = tx.Exec("delete from public.days where day_id = ?", daysIds[i]).Error
			if err != nil {
				tx.Rollback()
				errInternal.Message = err.Error()
				return errInternal
			}
		}
	}
	return nil
}

func deleteLessons(tx *gorm.DB, lessonsIds []uuid.UUID) common.Err {
	var err error
	errInternal := common.RespErr{Status: http.StatusInternalServerError}

	for j := 0; j < len(lessonsIds); j++ {
		if lessonsIds[j] != uuid.Nil {
			fmt.Println(lessonsIds[j])
			err = tx.Exec("delete from public.lessons where lesson_id = ?", lessonsIds[j]).Error
			if err != nil {
				tx.Rollback()
				errInternal.Message = err.Error()
				return errInternal
			}
		}
	}
	return nil
}
