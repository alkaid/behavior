package wrand

import (
	"errors"
	"math/rand"
	"sort"

	"github.com/samber/lo"
)

// Choice is a generic wrapper that can be used to add weights for any item.
type Choice struct {
	Key    any
	Item   any
	Weight uint
}

// NewChoice creates a new Choice with specified item and weight.
func NewChoice(key any, item any, weight uint) Choice {
	return Choice{Key: key, Item: item, Weight: weight}
}

// A Chooser caches many possible Choices in a structure designed to improve
// performance on repeated calls for weighted random selection.
type Chooser struct {
	data   []Choice
	totals []int
	max    int
}

func ShuffleWithWeights(weights []int) ([]int, error) {
	shuffled := make([]int, len(weights))
	items := make([]int, len(weights))
	for i := 0; i < len(weights); i++ {
		items[i] = i
	}
	pick := -1
	leng := len(weights)
	for i := 0; i < leng-1; i++ {
		if pick >= 0 {
			weights = lo.Drop(weights, 1)
			items = lo.Drop(items, 1)
		}
		chooser, err := NewChooserWithWeights(items, weights)
		if err != nil {
			return nil, err
		}
		_pick, _ := chooser.Pick()
		pick = _pick.(int)
		shuffled[i] = pick
	}
	shuffled[leng-1] = items[0]
	return shuffled, nil
}

func NewChooserWithWeights(items []int, weights []int) (*Chooser, error) {
	choices := make([]Choice, len(weights))
	for i := 0; i < len(weights); i++ {
		choices[i] = NewChoice(items[i], items[i], uint(weights[i]))
	}
	return NewChooser(choices...)
}

// NewChooser initializes a new Chooser for picking from the provided choices.
func NewChooser(choices ...Choice) (*Chooser, error) {
	sort.Slice(choices, func(i, j int) bool {
		return choices[i].Weight < choices[j].Weight
	})

	totals := make([]int, len(choices))
	runningTotal := 0
	for i, c := range choices {
		weight := int(c.Weight)
		if (maxInt - runningTotal) <= weight {
			return nil, errWeightOverflow
		}
		runningTotal += weight
		totals[i] = runningTotal
	}

	if runningTotal < 1 {
		return nil, errNoValidChoices
	}

	return &Chooser{data: choices, totals: totals, max: runningTotal}, nil
}

const (
	intSize = 32 << (^uint(0) >> 63) // cf. strconv.IntSize
	maxInt  = 1<<(intSize-1) - 1
)

// Possible errors returned by NewChooser, preventing the creation of a Chooser
// with unsafe runtime states.
var (
	// If the sum of provided Choice weights exceed the maximum integer value
	// for the current platform (e.g. math.MaxInt32 or math.MaxInt64), then
	// the internal running total will overflow, resulting in an imbalanced
	// distribution generating improper results.
	errWeightOverflow = errors.New("sum of Choice Weights exceeds max int")
	// If there are no Choices available to the Chooser with a weight >= 1,
	// there are no valid choices and Pick would produce a runtime panic.
	errNoValidChoices = errors.New("zero Choices with Weight >= 1")
)

// Pick returns a single weighted random Choice.Item from the Chooser.
//
// Utilizes global rand as the source of randomness.
func (c Chooser) Pick() (key any, item interface{}) {
	i := searchInts(c.totals, rand.Intn(c.max)+1)
	return c.data[i].Key, c.data[i].Item
}

// The standard library sort.SearchInts() just wraps the generic sort.Search()
// function, which takes a function closure to determine truthfulness. However,
// since this function is utilized within a for loop, it cannot currently be
// properly inlined by the compiler, resulting in non-trivial performance
// overhead.
//
// Thus, this is essentially manually inlined version.  In our use case here, it
// results in a up to ~33% overall throughput increase for Pick().
func searchInts(a []int, x int) int {
	// Possible further future optimization for searchInts via SIMD if we want
	// to write some Go assembly code: http://0x80.pl/articles/simd-search.html
	i, j := 0, len(a)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
