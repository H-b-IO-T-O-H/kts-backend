package models

import (
	"github.com/google/uuid"
)

type Timetable struct {
	TimetableId uuid.UUID `gorm:"column:timetable_id;default:uuid_generate_v4()" json:"timetable_id"`
	GroupId     uuid.UUID `gorm:"column:group_id" json:"group_id"`
	GroupName   string    `gorm:"-" json:"group_name"`
	Weeks       []Week    `gorm:"-" json:"weeks"`
}

func (t Timetable) TableName() string {
	return "public.timetables"
}

type Lesson struct {
	LessonId    uuid.UUID `gorm:"column:lesson_id;default:uuid_generate_v4()" json:"lesson_id"`
	LessonOrder int       `gorm:"-" json:"lesson_order"`
	Title       string    `gorm:"column:title" json:"title"`
	Auditorium  string    `gorm:"column:auditorium" json:"auditorium"`
	LessonType  string    `gorm:"column:lesson_type" json:"lesson_type"`
	Comment     string    `gorm:"column:comment" json:"comment"`
}

func (l Lesson) TableName() string {
	return "public.lessons"
}

type Day struct {
	DayId    uuid.UUID `gorm:"column:day_id;default:uuid_generate_v4()" json:"day_id"`
	DayOrder int       `gorm:"-" json:"day_order"`
	Lessons  []Lesson  `gorm:"-" json:"lessons"`
}

func (d Day) TableName() string {
	return "public.days"
}

type Week struct {
	WeekId     uuid.UUID `gorm:"column:week_id; default:uuid_generate_v4()" json:"week_id"`
	WeekType   string    `gorm:"column:week_type" json:"week_type"`
	GroupName  string    `gorm:"-" json:"group"`
	WeekNumber int       `gorm:"-" json:"week_number"`
	Days       []Day     `gorm:"-" json:"days"`
}

func (w Week) TableName() string {
	return "public.weeks"
}

type Group struct {
	GroupId     uuid.UUID `gorm:"column:group_id; default:uuid_generate_v4()" json:"group_id"`
	GroupName   string    `gorm:"column:group_name" json:"group_name"`
	TimetableId uuid.UUID `gorm:"column:timetable_id" json:"timetable_id"`
}

func (g Group) TableName() string {
	return "public.groups"
}

type DeleteEntity struct {
	WeekId     uuid.UUID   `json:"week_id"`
	DaysIds    []uuid.UUID `json:"days_ids"`
	LessonsIds []uuid.UUID `json:"lessons_ids"`
}
