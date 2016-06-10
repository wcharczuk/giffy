package util

import (
	"math"
	"sort"
	"time"
)

//----------------------------------------------------------------------------------------------------
// NOTE: This was (mostly) purloined from https://github.com/montanaflynn/stats/blob/master/stats.go
// WCNOTE: I've removed errors from the returns as that felt really stupid
//----------------------------------------------------------------------------------------------------

// PowOfInt returns the base to the power.
func PowOfInt(base, power uint) int {
	if base == 2 {
		return 1 << power
	}
	return float64ToInt(math.Pow(float64(base), float64(power)))
}

// Min finds the lowest value in a slice.
func Min(input []float64) float64 {
	if len(input) == 0 {
		return 0
	}

	min := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] < min {
			min = input[i]
		}
	}
	return min
}

// MinOfInt finds the lowest value in a slice.
func MinOfInt(input []int) int {
	if len(input) == 0 {
		return 0
	}

	min := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] < min {
			min = input[i]
		}
	}
	return min
}

// MinOfDuration finds the lowest value in a slice.
func MinOfDuration(input []time.Duration) time.Duration {
	if len(input) == 0 {
		return time.Duration(0)
	}

	min := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] < min {
			min = input[i]
		}
	}
	return min
}

// Max finds the highest value in a slice.
func Max(input []float64) float64 {

	if len(input) == 0 {
		return 0
	}

	max := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] > max {
			max = input[i]
		}
	}

	return max
}

// MaxOfInt finds the highest value in a slice.
func MaxOfInt(input []int) int {
	if len(input) == 0 {
		return 0
	}

	max := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] > max {
			max = input[i]
		}
	}

	return max
}

// MaxOfDuration finds the highest value in a slice.'
func MaxOfDuration(input []time.Duration) time.Duration {
	if len(input) == 0 {
		return time.Duration(0)
	}

	max := input[0]

	for i := 1; i < len(input); i++ {
		if input[i] > max {
			max = input[i]
		}
	}

	return max
}

// Sum adds all the numbers of a slice together
func Sum(input []float64) float64 {

	if len(input) == 0 {
		return 0
	}

	sum := float64(0)

	// Add em up
	for _, n := range input {
		sum += n
	}

	return sum
}

// SumOfInt adds all the numbers of a slice together
func SumOfInt(values []int) int {
	total := 0
	for x := 0; x < len(values); x++ {
		total += values[x]
	}

	return total
}

// SumOfDuration adds all the values of a slice together
func SumOfDuration(values []time.Duration) time.Duration {
	total := time.Duration(0)
	for x := 0; x < len(values); x++ {
		total += values[x]
	}

	return total
}

// Mean gets the average of a slice of numbers
func Mean(input []float64) float64 {
	if len(input) == 0 {
		return 0
	}

	sum := Sum(input)

	return sum / float64(len(input))
}

// MeanOfInt gets the average of a slice of numbers
func MeanOfInt(input []int) float64 {
	if len(input) == 0 {
		return 0
	}

	sum := SumOfInt(input)
	return float64(sum) / float64(len(input))
}

// MeanOfDuration gets the average of a slice of numbers
func MeanOfDuration(input []time.Duration) time.Duration {
	if len(input) == 0 {
		return 0
	}

	sum := SumOfDuration(input)
	mean := uint64(sum) / uint64(len(input))
	return time.Duration(mean)
}

// Median gets the median number in a slice of numbers
func Median(input []float64) float64 {
	median := float64(0)
	l := len(input)
	if l == 0 {
		return 0
	}
	c := copyslice(input)
	sort.Float64s(c)

	if l%2 == 0 {
		median = Mean(c[l/2-1 : l/2+1])
	} else {
		median = float64(c[l/2])
	}

	return median
}

// Mode gets the mode of a slice of numbers
// `Mode` generally is the most frequently occurring values within the input set.
func Mode(input []float64) []float64 {

	l := len(input)
	if l == 1 {
		return input
	} else if l == 0 {
		return []float64{}
	}

	m := make(map[float64]int)
	for _, v := range input {
		m[v]++
	}

	mode := []float64{}

	var current int
	for k, v := range m {
		switch {
		case v < current:
		case v > current:
			current = v
			mode = append(mode[:0], k)
		default:
			mode = append(mode, k)
		}
	}

	lm := len(mode)
	if l == lm {
		return []float64{}
	}

	return mode
}

// Variance finds the variance for both population and sample data
func Variance(input []float64, sample int) float64 {

	if len(input) == 0 {
		return 0
	}

	variance := float64(0)
	m := Mean(input)

	for _, n := range input {
		variance += (float64(n) - m) * (float64(n) - m)
	}

	// When getting the mean of the squared differences
	// "sample" will allow us to know if it's a sample
	// or population and wether to subtract by one or not
	return variance / float64((len(input) - (1 * sample)))
}

// VarP finds the amount of variance within a population
func VarP(input []float64) float64 {
	return Variance(input, 0)
}

// VarS finds the amount of variance within a sample
func VarS(input []float64) float64 {
	return Variance(input, 1)
}

// StdDevP finds the amount of variation from the population
func StdDevP(input []float64) float64 {

	if len(input) == 0 {
		return 0
	}

	// stdev is generally the square root of the variance
	return math.Pow(VarP(input), 0.5)
}

// StdDevS finds the amount of variation from a sample
func StdDevS(input []float64) float64 {

	if len(input) == 0 {
		return 0
	}

	// stdev is generally the square root of the variance
	return math.Pow(VarS(input), 0.5)
}

// Round a float to a specific decimal place or precision
func Round(input float64, places int) float64 {
	if math.IsNaN(input) {
		return 0.0
	}

	sign := 1.0
	if input < 0 {
		sign = -1
		input *= -1
	}

	rounded := float64(0)
	precision := math.Pow(10, float64(places))
	digit := input * precision
	_, decimal := math.Modf(digit)

	if decimal >= 0.5 {
		rounded = math.Ceil(digit)
	} else {
		rounded = math.Floor(digit)
	}

	return rounded / precision * sign
}

// Percentile finds the relative standing in a slice of floats
func Percentile(input []float64, percent float64) float64 {
	if len(input) == 0 {
		return 0
	}

	c := copyslice(input)
	sort.Float64s(c)
	index := (percent / 100.0) * float64(len(c))

	percentile := float64(0)
	if index == float64(int64(index)) {
		i := float64ToInt(index)
		percentile = Mean([]float64{c[i-1], c[i]})
	} else {
		i := float64ToInt(index)
		percentile = c[i-1]
	}

	return percentile
}

// float64ToInt rounds a float64 to an int
func float64ToInt(input float64) (output int) {
	r := Round(input, 0)
	return int(r)
}

// copyslice copies a slice of float64s
func copyslice(input []float64) []float64 {
	s := make([]float64, len(input))
	copy(s, input)
	return s
}
