/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"sync"
	"time"

<<<<<<< HEAD
	"k8s.io/apimachinery/pkg/util/clock"
	compbasemetrics "k8s.io/component-base/metrics"
	"k8s.io/klog/v2"
=======
	compbasemetrics "k8s.io/component-base/metrics"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
>>>>>>> upstream/master
)

const (
	labelNameMark       = "mark"
	labelValueLo        = "low"
	labelValueHi        = "high"
	labelNamePhase      = "phase"
	labelValueWaiting   = "waiting"
	labelValueExecuting = "executing"
)

<<<<<<< HEAD
// SampleAndWaterMarkPairGenerator makes pairs of TimedObservers that
=======
// SampleAndWaterMarkPairGenerator makes pairs of RatioedChangeObservers that
>>>>>>> upstream/master
// track samples and watermarks.
type SampleAndWaterMarkPairGenerator struct {
	urGenerator SampleAndWaterMarkObserverGenerator
}

<<<<<<< HEAD
var _ TimedObserverPairGenerator = SampleAndWaterMarkPairGenerator{}
=======
var _ RatioedChangeObserverPairGenerator = SampleAndWaterMarkPairGenerator{}
>>>>>>> upstream/master

// NewSampleAndWaterMarkHistogramsPairGenerator makes a new pair generator
func NewSampleAndWaterMarkHistogramsPairGenerator(clock clock.PassiveClock, samplePeriod time.Duration, sampleOpts, waterMarkOpts *compbasemetrics.HistogramOpts, labelNames []string) SampleAndWaterMarkPairGenerator {
	return SampleAndWaterMarkPairGenerator{
		urGenerator: NewSampleAndWaterMarkHistogramsGenerator(clock, samplePeriod, sampleOpts, waterMarkOpts, append([]string{labelNamePhase}, labelNames...)),
	}
}

// Generate makes a new pair
<<<<<<< HEAD
func (spg SampleAndWaterMarkPairGenerator) Generate(waiting1, executing1 float64, labelValues []string) TimedObserverPair {
	return TimedObserverPair{
		RequestsWaiting:   spg.urGenerator.Generate(0, waiting1, append([]string{labelValueWaiting}, labelValues...)),
		RequestsExecuting: spg.urGenerator.Generate(0, executing1, append([]string{labelValueExecuting}, labelValues...)),
=======
func (spg SampleAndWaterMarkPairGenerator) Generate(initialWaitingDenominator, initialExecutingDenominator float64, labelValues []string) RatioedChangeObserverPair {
	return RatioedChangeObserverPair{
		RequestsWaiting:   spg.urGenerator.Generate(0, initialWaitingDenominator, append([]string{labelValueWaiting}, labelValues...)),
		RequestsExecuting: spg.urGenerator.Generate(0, initialExecutingDenominator, append([]string{labelValueExecuting}, labelValues...)),
>>>>>>> upstream/master
	}
}

func (spg SampleAndWaterMarkPairGenerator) metrics() Registerables {
	return spg.urGenerator.metrics()
}

<<<<<<< HEAD
// SampleAndWaterMarkObserverGenerator creates TimedObservers that
=======
// SampleAndWaterMarkObserverGenerator creates RatioedChangeObservers that
>>>>>>> upstream/master
// populate histograms of samples and low- and high-water-marks.  The
// generator has a samplePeriod, and the histograms get an observation
// every samplePeriod.  The sampling windows are quantized based on
// the monotonic rather than wall-clock times.  The `t0` field is
// there so to provide a baseline for monotonic clock differences.
type SampleAndWaterMarkObserverGenerator struct {
	*sampleAndWaterMarkObserverGenerator
}

type sampleAndWaterMarkObserverGenerator struct {
	clock        clock.PassiveClock
	t0           time.Time
	samplePeriod time.Duration
	samples      *compbasemetrics.HistogramVec
	waterMarks   *compbasemetrics.HistogramVec
}

<<<<<<< HEAD
var _ TimedObserverGenerator = (*sampleAndWaterMarkObserverGenerator)(nil)
=======
var _ RatioedChangeObserverGenerator = SampleAndWaterMarkObserverGenerator{}
>>>>>>> upstream/master

// NewSampleAndWaterMarkHistogramsGenerator makes a new one
func NewSampleAndWaterMarkHistogramsGenerator(clock clock.PassiveClock, samplePeriod time.Duration, sampleOpts, waterMarkOpts *compbasemetrics.HistogramOpts, labelNames []string) SampleAndWaterMarkObserverGenerator {
	return SampleAndWaterMarkObserverGenerator{
		&sampleAndWaterMarkObserverGenerator{
			clock:        clock,
			t0:           clock.Now(),
			samplePeriod: samplePeriod,
			samples:      compbasemetrics.NewHistogramVec(sampleOpts, labelNames),
			waterMarks:   compbasemetrics.NewHistogramVec(waterMarkOpts, append([]string{labelNameMark}, labelNames...)),
		}}
}

func (swg *sampleAndWaterMarkObserverGenerator) quantize(when time.Time) int64 {
	return int64(when.Sub(swg.t0) / swg.samplePeriod)
}

<<<<<<< HEAD
// Generate makes a new TimedObserver
func (swg *sampleAndWaterMarkObserverGenerator) Generate(x, x1 float64, labelValues []string) TimedObserver {
	relX := x / x1
=======
// Generate makes a new RatioedChangeObserver
func (swg *sampleAndWaterMarkObserverGenerator) Generate(initialNumerator, initialDenominator float64, labelValues []string) RatioedChangeObserver {
	ratio := initialNumerator / initialDenominator
>>>>>>> upstream/master
	when := swg.clock.Now()
	return &sampleAndWaterMarkHistograms{
		sampleAndWaterMarkObserverGenerator: swg,
		labelValues:                         labelValues,
		loLabelValues:                       append([]string{labelValueLo}, labelValues...),
		hiLabelValues:                       append([]string{labelValueHi}, labelValues...),
<<<<<<< HEAD
		x1:                                  x1,
		sampleAndWaterMarkAccumulator: sampleAndWaterMarkAccumulator{
			lastSet:    when,
			lastSetInt: swg.quantize(when),
			x:          x,
			relX:       relX,
			loRelX:     relX,
			hiRelX:     relX,
=======
		denominator:                         initialDenominator,
		sampleAndWaterMarkAccumulator: sampleAndWaterMarkAccumulator{
			lastSet:    when,
			lastSetInt: swg.quantize(when),
			numerator:  initialNumerator,
			ratio:      ratio,
			loRatio:    ratio,
			hiRatio:    ratio,
>>>>>>> upstream/master
		}}
}

func (swg *sampleAndWaterMarkObserverGenerator) metrics() Registerables {
	return Registerables{swg.samples, swg.waterMarks}
}

type sampleAndWaterMarkHistograms struct {
	*sampleAndWaterMarkObserverGenerator
	labelValues                  []string
	loLabelValues, hiLabelValues []string

	sync.Mutex
<<<<<<< HEAD
	x1 float64
=======
	denominator float64
>>>>>>> upstream/master
	sampleAndWaterMarkAccumulator
}

type sampleAndWaterMarkAccumulator struct {
<<<<<<< HEAD
	lastSet        time.Time
	lastSetInt     int64 // lastSet / samplePeriod
	x              float64
	relX           float64 // x / x1
	loRelX, hiRelX float64
}

var _ TimedObserver = (*sampleAndWaterMarkHistograms)(nil)

func (saw *sampleAndWaterMarkHistograms) Add(deltaX float64) {
	saw.innerSet(func() {
		saw.x += deltaX
	})
}

func (saw *sampleAndWaterMarkHistograms) Set(x float64) {
	saw.innerSet(func() {
		saw.x = x
	})
}

func (saw *sampleAndWaterMarkHistograms) SetX1(x1 float64) {
	saw.innerSet(func() {
		saw.x1 = x1
	})
}

func (saw *sampleAndWaterMarkHistograms) innerSet(updateXOrX1 func()) {
=======
	lastSet          time.Time
	lastSetInt       int64 // lastSet / samplePeriod
	numerator        float64
	ratio            float64 // numerator/denominator
	loRatio, hiRatio float64
}

var _ RatioedChangeObserver = (*sampleAndWaterMarkHistograms)(nil)

func (saw *sampleAndWaterMarkHistograms) Add(deltaNumerator float64) {
	saw.innerSet(func() {
		saw.numerator += deltaNumerator
	})
}

func (saw *sampleAndWaterMarkHistograms) Observe(numerator float64) {
	saw.innerSet(func() {
		saw.numerator = numerator
	})
}

func (saw *sampleAndWaterMarkHistograms) SetDenominator(denominator float64) {
	saw.innerSet(func() {
		saw.denominator = denominator
	})
}

func (saw *sampleAndWaterMarkHistograms) innerSet(updateNumeratorOrDenominator func()) {
>>>>>>> upstream/master
	when, whenInt, acc, wellOrdered := func() (time.Time, int64, sampleAndWaterMarkAccumulator, bool) {
		saw.Lock()
		defer saw.Unlock()
		// Moved these variables here to tiptoe around https://github.com/golang/go/issues/43570 for #97685
		when := saw.clock.Now()
		whenInt := saw.quantize(when)
		acc := saw.sampleAndWaterMarkAccumulator
		wellOrdered := !when.Before(acc.lastSet)
<<<<<<< HEAD
		updateXOrX1()
		saw.relX = saw.x / saw.x1
		if wellOrdered {
			if acc.lastSetInt < whenInt {
				saw.loRelX, saw.hiRelX = acc.relX, acc.relX
=======
		updateNumeratorOrDenominator()
		saw.ratio = saw.numerator / saw.denominator
		if wellOrdered {
			if acc.lastSetInt < whenInt {
				saw.loRatio, saw.hiRatio = acc.ratio, acc.ratio
>>>>>>> upstream/master
				saw.lastSetInt = whenInt
			}
			saw.lastSet = when
		}
		// `wellOrdered` should always be true because we are using
		// monotonic clock readings and they never go backwards.  Yet
		// very small backwards steps (under 1 microsecond) have been
		// observed
		// (https://github.com/kubernetes/kubernetes/issues/96459).
		// In the backwards case, treat the current reading as if it
		// had occurred at time `saw.lastSet` and log an error.  It
		// would be wrong to update `saw.lastSet` in this case because
		// that plants a time bomb for future updates to
		// `saw.lastSetInt`.
<<<<<<< HEAD
		if saw.relX < saw.loRelX {
			saw.loRelX = saw.relX
		} else if saw.relX > saw.hiRelX {
			saw.hiRelX = saw.relX
=======
		if saw.ratio < saw.loRatio {
			saw.loRatio = saw.ratio
		} else if saw.ratio > saw.hiRatio {
			saw.hiRatio = saw.ratio
>>>>>>> upstream/master
		}
		return when, whenInt, acc, wellOrdered
	}()
	if !wellOrdered {
		lastSetS := acc.lastSet.String()
		whenS := when.String()
		klog.Errorf("Time went backwards from %s to %s for labelValues=%#+v", lastSetS, whenS, saw.labelValues)
	}
	for acc.lastSetInt < whenInt {
<<<<<<< HEAD
		saw.samples.WithLabelValues(saw.labelValues...).Observe(acc.relX)
		saw.waterMarks.WithLabelValues(saw.loLabelValues...).Observe(acc.loRelX)
		saw.waterMarks.WithLabelValues(saw.hiLabelValues...).Observe(acc.hiRelX)
		acc.lastSetInt++
		acc.loRelX, acc.hiRelX = acc.relX, acc.relX
=======
		saw.samples.WithLabelValues(saw.labelValues...).Observe(acc.ratio)
		saw.waterMarks.WithLabelValues(saw.loLabelValues...).Observe(acc.loRatio)
		saw.waterMarks.WithLabelValues(saw.hiLabelValues...).Observe(acc.hiRatio)
		acc.lastSetInt++
		acc.loRatio, acc.hiRatio = acc.ratio, acc.ratio
>>>>>>> upstream/master
	}
}
