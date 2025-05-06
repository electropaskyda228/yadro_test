package internals

import (
	"fmt"
	"time"
)

type Race struct {
	Laps        uint   `json:"laps"`
	LapLen      uint   `json:"lapLen"`
	PenaltyLen  uint   `json:"penaltyLen"`
	FiringLines uint   `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

type Event struct {
	Id          uint
	Participant uint
	ExtraParams string
	Time        string
}

type Result struct {
	Registered         bool
	Status             uint
	StartTime          string
	ScheduledStartTime string
	Laps               []PairLap
	PenaltyLap         []PairLap
	Hits               uint
	Shoots             uint
	MainResult         string
	PenaltyStartTime   string
	MainStartTime      string
	FinishTime         string
	Id                 uint
}

type PairLap struct {
	Duration time.Duration
	Speed    float64
}

func getDefaultResult() *Result {
	return &Result{
		Registered:         false,
		Status:             0,
		ScheduledStartTime: "",
		StartTime:          "",
		Laps:               make([]PairLap, 0),
		PenaltyLap:         make([]PairLap, 0),
		Hits:               0,
		Shoots:             0,
		MainResult:         "",
		PenaltyStartTime:   "",
		MainStartTime:      "",
		FinishTime:         "",
		Id:                 0,
	}
}

func getOutLogLine(event *Event) string {
	switch event.Id {
	case 1:
		return fmt.Sprintf("The competitor(%d) registered", event.Participant)
	case 2:
		return fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s", event.Participant, event.ExtraParams)
	case 3:
		return fmt.Sprintf("The competitor(%d) is on the start line", event.Participant)
	case 4:
		return fmt.Sprintf("The competitor(%d) has started", event.Participant)
	case 5:
		return fmt.Sprintf("The competitor(%d) is on the firing range(%s)", event.Participant, event.ExtraParams)
	case 6:
		return fmt.Sprintf("The target(%s) has been hit by competitor(%d)", event.ExtraParams, event.Participant)
	case 7:
		return fmt.Sprintf("The competitor(%d) left the firing range", event.Participant)
	case 8:
		return fmt.Sprintf("The competitor(%d) entered the penalty laps", event.Participant)
	case 9:
		return fmt.Sprintf("The competitor(%d) left the penalty laps", event.Participant)
	case 10:
		return fmt.Sprintf("The competitor(%d) ended the main lap", event.Participant)
	case 11:
		return fmt.Sprintf("The competitor(%d) can`t continue: %s", event.Participant, event.ExtraParams)
	}
	return ""
}

func parseTime(timeString string) (time.Time, error) {
	return time.Parse("15:04:05.000", timeString)
}

func subTime(start, end string) (time.Duration, error) {
	t1, err := parseTime(start)
	if err != nil {
		return 0, err
	}

	t2, err := parseTime(end)
	if err != nil {
		return 0, err
	}

	return t2.Sub(t1), nil
}

func getFromDurationToString(diff time.Duration) string {
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60
	millis := int(diff.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

func compareWithThreshold(diffStr, thresholdStr string) (bool, error) {
	diff, err := parseTime(diffStr)
	if err != nil {
		return false, err
	}

	threshold, err := parseTime(thresholdStr)
	if err != nil {
		return false, err
	}

	diffDur := time.Duration(diff.Hour())*time.Hour +
		time.Duration(diff.Minute())*time.Minute +
		time.Duration(diff.Second())*time.Second +
		time.Duration(diff.Nanosecond())

	thresholdDur := time.Duration(threshold.Hour())*time.Hour +
		time.Duration(threshold.Minute())*time.Minute +
		time.Duration(threshold.Second())*time.Second +
		time.Duration(threshold.Nanosecond())

	if diffDur <= thresholdDur {
		return true, nil
	}
	return false, nil
}

func calculateTimeForPenaltyLoop(event *Event, penaltyStartTime string, race *Race) *PairLap {
	diff, _ := subTime(penaltyStartTime, event.Time)
	var speed float64 = float64(race.PenaltyLen) / diff.Seconds()
	return &PairLap{Duration: diff, Speed: speed}
}

func calculateTimeForMainLoop(event *Event, mainStartTime string, race *Race) *PairLap {
	diff, _ := subTime(mainStartTime, event.Time)
	var speed float64 = float64(race.LapLen) / diff.Seconds()
	return &PairLap{Duration: diff, Speed: speed}
}
