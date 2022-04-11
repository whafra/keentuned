package modules

import (
	"fmt"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"math"
)

func (tuner *Tuner) analyseBestResult() string {
	if tuner.isSensitize {
		return ""
	}

	var currentRatioInfo string
	for _, name := range tuner.Benchmark.SortedItems {
		info, ok := tuner.bestInfo.Score[name]
		if !ok {
			log.Warnf("", "%vth best config [%v] info not exist", tuner.Iteration, name)
			continue
		}

		currentInfo, ok := tuner.feedbackScore[name]
		if !ok {
			log.Warnf("", "%vth bench config [%v] info not exist", tuner.Iteration, name)
			continue
		}

		oneRatioInfo := getBestRatio(info, currentInfo, tuner.Verbose, name)
		if oneRatioInfo == "" {
			continue
		}

		currentRatioInfo += fmt.Sprintf("\n\t%v", oneRatioInfo)
	}

	return currentRatioInfo
}

// analyseResult analyse benchmark score Result
func (tuner *Tuner) analyseResult() string {
	if tuner.isSensitize {
		return ""
	}

	if err := tuner.getBest(); err != nil {
		return ""
	}

	var currentRatioInfo string
	for _, name := range tuner.Benchmark.SortedItems {
		info, ok := tuner.bestInfo.Score[name]
		if !ok {
			log.Warnf("", "%vth config [%v] info not exist", tuner.Iteration, name)
			continue
		}

		oneRatioInfo := getRatio(info, tuner.Verbose, name)
		if oneRatioInfo == "" {
			continue
		}

		currentRatioInfo += fmt.Sprintf("\n\t%v", oneRatioInfo)
	}

	if currentRatioInfo != "" {
		return fmt.Sprintf("[Iteration %v]:%v", tuner.bestInfo.Round+1, currentRatioInfo)
	}

	return currentRatioInfo
}

func getRatio(info ItemDetail, verbose bool, name string) string {
	if len(info.Baseline) == 0 {
		return ""
	}

	var sum float32
	for _, base := range info.Baseline {
		sum += base
	}
	average := sum / float32(len(info.Baseline))
	score := utils.IncreaseRatio(info.Value, average)
	return sprintRatio(info, verbose, name, score, average)
}

func getBestRatio(info ItemDetail, currentScore []float32, verbose bool, name string) string {
	if len(info.Baseline) == 0 || len(currentScore) == 0 {
		return ""
	}

	var baseSum, currentSum float32
	for _, base := range info.Baseline {
		baseSum += base
	}

	for _, score := range currentScore {
		currentSum += score
	}

	baseAverage := baseSum / float32(len(info.Baseline))
	currentAverage := currentSum / float32(len(currentScore))
	score := utils.IncreaseRatio(currentAverage, baseAverage)
	return sprintRatio(info, verbose, name, score, baseAverage)
}

func sprintRatio(info ItemDetail, verbose bool, name string, score float32, average float32) string {
	if verbose {
		if (score < 0.0 && info.Negative) || (score > 0.0 && !info.Negative) {
			info := utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return fmt.Sprintf("[%v]\tImproved by %s;\t(baseline = %.3f)", name, info, average)
		} else {
			info := utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return fmt.Sprintf("[%v]\tDeclined by %s;\t(baseline = %.3f)", name, info, average)
		}
	}

	if !verbose && info.Weight > 0.0 {
		if (score < 0.0 && info.Negative) || (score > 0.0 && !info.Negative) {
			info := utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return fmt.Sprintf("[%v]\tImproved by %s", name, info)
		} else {
			info := utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return fmt.Sprintf("[%v]\tDeclined by %s", name, info)
		}
	}

	return ""
}

