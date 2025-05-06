package internals

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type EventParseError struct {
	err error
}

func (epe *EventParseError) Error() string {
	return epe.err.Error()
}

type EventOrderError struct {
	from uint
	to   uint
}

func (eoe *EventOrderError) Error() string {
	return "Can not proccess event from status (" + strconv.FormatUint(uint64(eoe.from), 10) + ") to (" + strconv.FormatUint(uint64(eoe.to), 10) + ")"
}

func ProccessRace(race *Race, pathLog string, outputPath string) error {
	// Входной файл логов
	file, err := os.Open(pathLog)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Выходной файл логов
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// Главная часть СХД
	results := make(map[uint]*Result)

	// Главный цикл обработки
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		event, err := parseEvent(line)
		if err != nil {
			writer.WriteString("ERROR:" + err.Error() + "\n")
			continue
		}

		handleEvent(event, &results, writer, race)
	}

	sortedResults := make([]*Result, 0)
	for _, item := range results {
		sortedResults = append(sortedResults, item)
	}
	sort.Slice(sortedResults, func(i, j int) bool {
		if sortedResults[i].MainResult == "NotStarted" {
			return false
		} else if sortedResults[i].MainResult == "NotStarted" {
			return true
		} else if sortedResults[i].MainResult == "NotFinished" {
			return false
		} else if sortedResults[j].MainResult == "NotFinished" {
			return true
		}
		answer, _ := compareWithThreshold(sortedResults[i].MainResult, sortedResults[j].MainResult)
		return answer
	})
	makeReport(sortedResults, writer, race)

	return nil
}

func makeReport(model []*Result, writer *bufio.Writer, race *Race) {
	for _, result := range model {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("[%s] %d [", result.MainResult, result.Id))

		for i := 0; i < int(race.Laps); i++ {
			if len(result.Laps) > i {
				builder.WriteString(fmt.Sprintf("{%s, %f}", getFromDurationToString(result.Laps[i].Duration), result.Laps[i].Speed))
			} else {
				builder.WriteString("{,}")
			}
			if i != int(race.Laps)-1 {
				builder.WriteString(", ")
			} else {
				builder.WriteString("] ")
			}
		}

		var penaltyTime time.Duration
		for _, pairLap := range result.PenaltyLap {
			penaltyTime += pairLap.Duration
		}
		var avSpeed float64 = float64(race.PenaltyLen) * float64(len(result.PenaltyLap)) / penaltyTime.Seconds()
		builder.WriteString(fmt.Sprintf("{%s, %f} ", getFromDurationToString(penaltyTime), avSpeed))

		builder.WriteString(fmt.Sprintf("%d/%d", result.Hits, result.Shoots))

		writer.WriteString(builder.String() + "\n")

	}
}

func handleEvent(event *Event, model *map[uint]*Result, writer *bufio.Writer, race *Race) error {
	modelResult, ok := (*model)[event.Participant]
	if !ok {
		modelResult = getDefaultResult()
		modelResult.Id = event.Participant
	}
	switch event.Id {
	case 1:
		modelResult.Registered = true
	case 2:
		modelResult.ScheduledStartTime = event.ExtraParams
	case 4:
		modelResult.StartTime = event.Time
		modelResult.MainStartTime = event.Time
		diffString, _ := subTime(modelResult.ScheduledStartTime, event.Time)
		if canRun, err := compareWithThreshold(getFromDurationToString(diffString), race.StartDelta); err != nil || !canRun {
			modelResult.Status = 0
			modelResult.MainResult = "NotStarted"
		}
	case 5:
		modelResult.Shoots += 5
	case 6:
		modelResult.Hits += 1
	case 8:
		modelResult.PenaltyStartTime = event.Time
	case 9:
		modelResult.PenaltyLap = append(modelResult.PenaltyLap, *calculateTimeForPenaltyLoop(event, modelResult.PenaltyStartTime, race))
	case 10:
		modelResult.Laps = append(modelResult.Laps, *calculateTimeForMainLoop(event, modelResult.MainStartTime, race))
		modelResult.MainStartTime = event.Time
		if len(modelResult.Laps) == int(race.Laps) {
			modelResult.FinishTime = event.Time
			diff, _ := subTime(modelResult.StartTime, modelResult.FinishTime)
			modelResult.MainResult = getFromDurationToString(diff)
		}
	case 11:
		modelResult.MainResult = "NotFinished"
	}
	modelResult.Status = event.Id

	writer.WriteString(fmt.Sprintf("[%s] %s\n", event.Time, getOutLogLine(event)))
	(*model)[event.Participant] = modelResult

	return nil
}

func parseEvent(line string) (*Event, error) {
	data := strings.Fields(line)
	if len(data) < 3 {
		return nil, &EventParseError{}
	}

	var event Event
	event.Time = strings.Trim(data[0], string(data[0][0])+string(data[0][len(data[0])-1]))

	eventId, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		return nil, &EventParseError{err: err}
	}
	event.Id = uint(eventId)

	participant, err := strconv.ParseUint(data[2], 10, 64)
	if err != nil {
		return nil, &EventParseError{err: err}
	}
	event.Participant = uint(participant)

	if len(data) == 4 && (event.Id == 2 || event.Id == 5 || event.Id == 6 || event.Id == 11) {
		event.ExtraParams = data[3]
	}

	return &event, nil

}
