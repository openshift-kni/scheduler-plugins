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

package queueset

import (
	"context"
<<<<<<< HEAD
=======
	"errors"
>>>>>>> upstream/master
	"fmt"
	"math"
	"sync"
	"time"

<<<<<<< HEAD
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/util/flowcontrol/counter"
	"k8s.io/apiserver/pkg/util/flowcontrol/debug"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	"k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/promise/lockingpromise"
=======
	"k8s.io/apiserver/pkg/util/flowcontrol/debug"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	"k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/eventclock"
	"k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/promise"
>>>>>>> upstream/master
	"k8s.io/apiserver/pkg/util/flowcontrol/metrics"
	fqrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
	"k8s.io/apiserver/pkg/util/shufflesharding"
	"k8s.io/klog/v2"
<<<<<<< HEAD
=======

	// The following hack is needed to work around a tooling deficiency.
	// Packages imported only for test code are not included in vendor.
	// See https://kubernetes.slack.com/archives/C0EG7JC6T/p1626985671458800?thread_ts=1626983387.450800&cid=C0EG7JC6T
	_ "k8s.io/utils/clock/testing"
>>>>>>> upstream/master
)

const nsTimeFmt = "2006-01-02 15:04:05.000000000"

// queueSetFactory implements the QueueSetFactory interface
// queueSetFactory makes QueueSet objects.
type queueSetFactory struct {
<<<<<<< HEAD
	counter counter.GoRoutineCounter
	clock   clock.PassiveClock
}

// `*queueSetCompleter` implements QueueSetCompleter.  Exactly one of
// the fields `factory` and `theSet` is non-nil.
type queueSetCompleter struct {
	factory *queueSetFactory
	obsPair metrics.TimedObserverPair
	theSet  *queueSet
	qCfg    fq.QueuingConfig
	dealer  *shufflesharding.Dealer
=======
	clock                 eventclock.Interface
	promiseFactoryFactory promiseFactoryFactory
}

// promiseFactory returns a WriteOnce
// - whose Set method is invoked with the queueSet locked, and
// - whose Get method is invoked with the queueSet not locked.
// The parameters are the same as for `promise.NewWriteOnce`.
type promiseFactory func(initial interface{}, doneCh <-chan struct{}, doneVal interface{}) promise.WriteOnce

// promiseFactoryFactory returns the promiseFactory to use for the given queueSet
type promiseFactoryFactory func(*queueSet) promiseFactory

// `*queueSetCompleter` implements QueueSetCompleter.  Exactly one of
// the fields `factory` and `theSet` is non-nil.
type queueSetCompleter struct {
	factory      *queueSetFactory
	reqsObsPair  metrics.RatioedChangeObserverPair
	execSeatsObs metrics.RatioedChangeObserver
	theSet       *queueSet
	qCfg         fq.QueuingConfig
	dealer       *shufflesharding.Dealer
>>>>>>> upstream/master
}

// queueSet implements the Fair Queuing for Server Requests technique
// described in this package's doc, and a pointer to one implements
<<<<<<< HEAD
// the QueueSet interface.  The clock, GoRoutineCounter, and estimated
// service time should not be changed; the fields listed after the
=======
// the QueueSet interface.  The fields listed before the lock
// should not be changed; the fields listed after the
>>>>>>> upstream/master
// lock must be accessed only while holding the lock.  The methods of
// this type follow the naming convention that the suffix "Locked"
// means the caller must hold the lock; for a method whose name does
// not end in "Locked" either acquires the lock or does not care about
// locking.
type queueSet struct {
<<<<<<< HEAD
	clock                clock.PassiveClock
	counter              counter.GoRoutineCounter
	estimatedServiceTime float64
	obsPair              metrics.TimedObserverPair
=======
	clock                    eventclock.Interface
	estimatedServiceDuration time.Duration

	reqsObsPair metrics.RatioedChangeObserverPair // .RequestsExecuting covers regular phase only

	execSeatsObs metrics.RatioedChangeObserver // for all phases of execution

	promiseFactory promiseFactory
>>>>>>> upstream/master

	lock sync.Mutex

	// qCfg holds the current queuing configuration.  Its
	// DesiredNumQueues may be less than the current number of queues.
	// If its DesiredNumQueues is zero then its other queuing
	// parameters retain the settings they had when DesiredNumQueues
	// was last non-zero (if ever).
	qCfg fq.QueuingConfig

	// the current dispatching configuration.
	dCfg fq.DispatchingConfig

<<<<<<< HEAD
	// If `config.DesiredNumQueues` is non-zero then dealer is not nil
	// and is good for `config`.
=======
	// If `qCfg.DesiredNumQueues` is non-zero then dealer is not nil
	// and is good for `qCfg`.
>>>>>>> upstream/master
	dealer *shufflesharding.Dealer

	// queues may be longer than the desired number, while the excess
	// queues are still draining.
	queues []*queue

<<<<<<< HEAD
	// virtualTime is the number of virtual seconds since process startup
	virtualTime float64
=======
	// currentR is the amount of seat-seconds allocated per queue since process startup.
	// This is our generalization of the progress meter named R in the original fair queuing work.
	currentR fqrequest.SeatSeconds
>>>>>>> upstream/master

	// lastRealTime is what `clock.Now()` yielded when `virtualTime` was last updated
	lastRealTime time.Time

	// robinIndex is the index of the last queue dispatched
	robinIndex int

	// totRequestsWaiting is the sum, over all the queues, of the
	// number of requests waiting in that queue
	totRequestsWaiting int

	// totRequestsExecuting is the total number of requests of this
	// queueSet that are currently executing.  That is the same as the
	// sum, over all the queues, of the number of requests executing
	// from that queue.
	totRequestsExecuting int

	// totSeatsInUse is the number of total "seats" in use by all the
	// request(s) that are currently executing in this queueset.
	totSeatsInUse int
<<<<<<< HEAD
}

// NewQueueSetFactory creates a new QueueSetFactory object
func NewQueueSetFactory(c clock.PassiveClock, counter counter.GoRoutineCounter) fq.QueueSetFactory {
	return &queueSetFactory{
		counter: counter,
		clock:   c,
	}
}

func (qsf *queueSetFactory) BeginConstruction(qCfg fq.QueuingConfig, obsPair metrics.TimedObserverPair) (fq.QueueSetCompleter, error) {
=======

	// enqueues is the number of requests that have ever been enqueued
	enqueues int
}

// NewQueueSetFactory creates a new QueueSetFactory object
func NewQueueSetFactory(c eventclock.Interface) fq.QueueSetFactory {
	return newTestableQueueSetFactory(c, ordinaryPromiseFactoryFactory)
}

// newTestableQueueSetFactory creates a new QueueSetFactory object with the given promiseFactoryFactory
func newTestableQueueSetFactory(c eventclock.Interface, promiseFactoryFactory promiseFactoryFactory) fq.QueueSetFactory {
	return &queueSetFactory{
		clock:                 c,
		promiseFactoryFactory: promiseFactoryFactory,
	}
}

func (qsf *queueSetFactory) BeginConstruction(qCfg fq.QueuingConfig, reqsObsPair metrics.RatioedChangeObserverPair, execSeatsObs metrics.RatioedChangeObserver) (fq.QueueSetCompleter, error) {
>>>>>>> upstream/master
	dealer, err := checkConfig(qCfg)
	if err != nil {
		return nil, err
	}
	return &queueSetCompleter{
<<<<<<< HEAD
		factory: qsf,
		obsPair: obsPair,
		qCfg:    qCfg,
		dealer:  dealer}, nil
=======
		factory:      qsf,
		reqsObsPair:  reqsObsPair,
		execSeatsObs: execSeatsObs,
		qCfg:         qCfg,
		dealer:       dealer}, nil
>>>>>>> upstream/master
}

// checkConfig returns a non-nil Dealer if the config is valid and
// calls for one, and returns a non-nil error if the given config is
// invalid.
func checkConfig(qCfg fq.QueuingConfig) (*shufflesharding.Dealer, error) {
	if qCfg.DesiredNumQueues == 0 {
		return nil, nil
	}
	dealer, err := shufflesharding.NewDealer(qCfg.DesiredNumQueues, qCfg.HandSize)
	if err != nil {
		err = fmt.Errorf("the QueueSetConfig implies an invalid shuffle sharding config (DesiredNumQueues is deckSize): %w", err)
	}
	return dealer, err
}

func (qsc *queueSetCompleter) Complete(dCfg fq.DispatchingConfig) fq.QueueSet {
	qs := qsc.theSet
	if qs == nil {
		qs = &queueSet{
<<<<<<< HEAD
			clock:                qsc.factory.clock,
			counter:              qsc.factory.counter,
			estimatedServiceTime: 60,
			obsPair:              qsc.obsPair,
			qCfg:                 qsc.qCfg,
			virtualTime:          0,
			lastRealTime:         qsc.factory.clock.Now(),
		}
	}
	qs.setConfiguration(qsc.qCfg, qsc.dealer, dCfg)
=======
			clock:                    qsc.factory.clock,
			estimatedServiceDuration: 3 * time.Millisecond,
			reqsObsPair:              qsc.reqsObsPair,
			execSeatsObs:             qsc.execSeatsObs,
			qCfg:                     qsc.qCfg,
			currentR:                 0,
			lastRealTime:             qsc.factory.clock.Now(),
		}
		qs.promiseFactory = qsc.factory.promiseFactoryFactory(qs)
	}
	qs.setConfiguration(context.Background(), qsc.qCfg, qsc.dealer, dCfg)
>>>>>>> upstream/master
	return qs
}

// createQueues is a helper method for initializing an array of n queues
func createQueues(n, baseIndex int) []*queue {
	fqqueues := make([]*queue, n)
	for i := 0; i < n; i++ {
		fqqueues[i] = &queue{index: baseIndex + i, requests: newRequestFIFO()}
	}
	return fqqueues
}

func (qs *queueSet) BeginConfigChange(qCfg fq.QueuingConfig) (fq.QueueSetCompleter, error) {
	dealer, err := checkConfig(qCfg)
	if err != nil {
		return nil, err
	}
	return &queueSetCompleter{
		theSet: qs,
		qCfg:   qCfg,
		dealer: dealer}, nil
}

<<<<<<< HEAD
// SetConfiguration is used to set the configuration for a queueSet.
// Update handling for when fields are updated is handled here as well -
// eg: if DesiredNum is increased, SetConfiguration reconciles by
// adding more queues.
func (qs *queueSet) setConfiguration(qCfg fq.QueuingConfig, dealer *shufflesharding.Dealer, dCfg fq.DispatchingConfig) {
	qs.lockAndSyncTime()
=======
// setConfiguration is used to set the configuration for a queueSet.
// Update handling for when fields are updated is handled here as well -
// eg: if DesiredNum is increased, setConfiguration reconciles by
// adding more queues.
func (qs *queueSet) setConfiguration(ctx context.Context, qCfg fq.QueuingConfig, dealer *shufflesharding.Dealer, dCfg fq.DispatchingConfig) {
	qs.lockAndSyncTime(ctx)
>>>>>>> upstream/master
	defer qs.lock.Unlock()

	if qCfg.DesiredNumQueues > 0 {
		// Adding queues is the only thing that requires immediate action
<<<<<<< HEAD
		// Removing queues is handled by omitting indexes >DesiredNum from
		// chooseQueueIndexLocked
=======
		// Removing queues is handled by attrition, removing a queue when
		// it goes empty and there are too many.
>>>>>>> upstream/master
		numQueues := len(qs.queues)
		if qCfg.DesiredNumQueues > numQueues {
			qs.queues = append(qs.queues,
				createQueues(qCfg.DesiredNumQueues-numQueues, len(qs.queues))...)
		}
	} else {
		qCfg.QueueLengthLimit = qs.qCfg.QueueLengthLimit
		qCfg.HandSize = qs.qCfg.HandSize
		qCfg.RequestWaitLimit = qs.qCfg.RequestWaitLimit
	}

	qs.qCfg = qCfg
	qs.dCfg = dCfg
	qs.dealer = dealer
	qll := qCfg.QueueLengthLimit
	if qll < 1 {
		qll = 1
	}
<<<<<<< HEAD
	qs.obsPair.RequestsWaiting.SetX1(float64(qll))
	qs.obsPair.RequestsExecuting.SetX1(float64(dCfg.ConcurrencyLimit))
=======
	qs.reqsObsPair.RequestsWaiting.SetDenominator(float64(qll))
	qs.reqsObsPair.RequestsExecuting.SetDenominator(float64(dCfg.ConcurrencyLimit))
	qs.execSeatsObs.SetDenominator(float64(dCfg.ConcurrencyLimit))
>>>>>>> upstream/master

	qs.dispatchAsMuchAsPossibleLocked()
}

// A decision about a request
type requestDecision int

// Values passed through a request's decision
const (
<<<<<<< HEAD
	decisionExecute requestDecision = iota
	decisionReject
=======
	// Serve this one
	decisionExecute requestDecision = iota

	// Reject this one due to APF queuing considerations
	decisionReject

	// This one's context timed out / was canceled
>>>>>>> upstream/master
	decisionCancel
)

// StartRequest begins the process of handling a request.  We take the
// approach of updating the metrics about total requests queued and
// executing at each point where there is a change in that quantity,
// because the metrics --- and only the metrics --- track that
// quantity per FlowSchema.
<<<<<<< HEAD
func (qs *queueSet) StartRequest(ctx context.Context, width *fqrequest.Width, hashValue uint64, flowDistinguisher, fsName string, descr1, descr2 interface{}, queueNoteFn fq.QueueNoteFn) (fq.Request, bool) {
	qs.lockAndSyncTime()
=======
// The queueSet's promiseFactory is invoked once if the returned Request is non-nil,
// not invoked if the Request is nil.
func (qs *queueSet) StartRequest(ctx context.Context, workEstimate *fqrequest.WorkEstimate, hashValue uint64, flowDistinguisher, fsName string, descr1, descr2 interface{}, queueNoteFn fq.QueueNoteFn) (fq.Request, bool) {
	qs.lockAndSyncTime(ctx)
>>>>>>> upstream/master
	defer qs.lock.Unlock()
	var req *request

	// ========================================================================
	// Step 0:
	// Apply only concurrency limit, if zero queues desired
	if qs.qCfg.DesiredNumQueues < 1 {
<<<<<<< HEAD
		if !qs.canAccommodateSeatsLocked(int(width.Seats)) {
			klog.V(5).Infof("QS(%s): rejecting request %q %#+v %#+v because %d seats are asked for, %d seats are in use (%d are executing) and the limit is %d",
				qs.qCfg.Name, fsName, descr1, descr2, width, qs.totSeatsInUse, qs.totRequestsExecuting, qs.dCfg.ConcurrencyLimit)
			metrics.AddReject(ctx, qs.qCfg.Name, fsName, "concurrency-limit")
			return nil, qs.isIdleLocked()
		}
		req = qs.dispatchSansQueueLocked(ctx, width, flowDistinguisher, fsName, descr1, descr2)
=======
		if !qs.canAccommodateSeatsLocked(workEstimate.MaxSeats()) {
			klog.V(5).Infof("QS(%s): rejecting request %q %#+v %#+v because %d seats are asked for, %d seats are in use (%d are executing) and the limit is %d",
				qs.qCfg.Name, fsName, descr1, descr2, workEstimate, qs.totSeatsInUse, qs.totRequestsExecuting, qs.dCfg.ConcurrencyLimit)
			metrics.AddReject(ctx, qs.qCfg.Name, fsName, "concurrency-limit")
			return nil, qs.isIdleLocked()
		}
		req = qs.dispatchSansQueueLocked(ctx, workEstimate, flowDistinguisher, fsName, descr1, descr2)
>>>>>>> upstream/master
		return req, false
	}

	// ========================================================================
	// Step 1:
	// 1) Start with shuffle sharding, to pick a queue.
	// 2) Reject old requests that have been waiting too long
	// 3) Reject current request if there is not enough concurrency shares and
	// we are at max queue length
	// 4) If not rejected, create a request and enqueue
<<<<<<< HEAD
	req = qs.timeoutOldRequestsAndRejectOrEnqueueLocked(ctx, width, hashValue, flowDistinguisher, fsName, descr1, descr2, queueNoteFn)
=======
	req = qs.timeoutOldRequestsAndRejectOrEnqueueLocked(ctx, workEstimate, hashValue, flowDistinguisher, fsName, descr1, descr2, queueNoteFn)
>>>>>>> upstream/master
	// req == nil means that the request was rejected - no remaining
	// concurrency shares and at max queue length already
	if req == nil {
		klog.V(5).Infof("QS(%s): rejecting request %q %#+v %#+v due to queue full", qs.qCfg.Name, fsName, descr1, descr2)
		metrics.AddReject(ctx, qs.qCfg.Name, fsName, "queue-full")
		return nil, qs.isIdleLocked()
	}

	// ========================================================================
	// Step 2:
	// The next step is to invoke the method that dequeues as much
	// as possible.
	// This method runs a loop, as long as there are non-empty
	// queues and the number currently executing is less than the
	// assured concurrency value.  The body of the loop uses the
	// fair queuing technique to pick a queue and dispatch a
	// request from that queue.
	qs.dispatchAsMuchAsPossibleLocked()

<<<<<<< HEAD
	// ========================================================================
	// Step 3:

	// Set up a relay from the context's Done channel to the world
	// of well-counted goroutines. We Are Told that every
	// request's context's Done channel gets closed by the time
	// the request is done being processed.
	doneCh := ctx.Done()

	// Retrieve the queueset configuration name while we have the lock
	// and use it in the goroutine below.
	configName := qs.qCfg.Name

	if doneCh != nil {
		qs.preCreateOrUnblockGoroutine()
		go func() {
			defer runtime.HandleCrash()
			qs.goroutineDoneOrBlocked()
			_ = <-doneCh
			// Whatever goroutine unblocked the preceding receive MUST
			// have already either (a) incremented qs.counter or (b)
			// known that said counter is not actually counting or (c)
			// known that the count does not need to be accurate.
			// BTW, the count only needs to be accurate in a test that
			// uses FakeEventClock::Run().
			klog.V(6).Infof("QS(%s): Context of request %q %#+v %#+v is Done", configName, fsName, descr1, descr2)
			qs.cancelWait(req)
			qs.goroutineDoneOrBlocked()
		}()
	}
	return req, false
}

// Seats returns the number of seats this request requires.
func (req *request) Seats() int {
	return int(req.width.Seats)
=======
	return req, false
}

// ordinaryPromiseFactoryFactory is the promiseFactoryFactory that
// a queueSetFactory would ordinarily use.
// Test code might use something different.
func ordinaryPromiseFactoryFactory(qs *queueSet) promiseFactory {
	return promise.NewWriteOnce
}

// MaxSeats returns the maximum number of seats this request requires, it is
// the maxumum of the two - WorkEstimate.InitialSeats, WorkEstimate.FinalSeats.
func (req *request) MaxSeats() int {
	return req.workEstimate.MaxSeats()
}

func (req *request) InitialSeats() int {
	return int(req.workEstimate.InitialSeats)
>>>>>>> upstream/master
}

func (req *request) NoteQueued(inQueue bool) {
	if req.queueNoteFn != nil {
		req.queueNoteFn(inQueue)
	}
}

func (req *request) Finish(execFn func()) bool {
	exec, idle := req.wait()
	if !exec {
		return idle
	}
	func() {
		defer func() {
			idle = req.qs.finishRequestAndDispatchAsMuchAsPossible(req)
		}()

		execFn()
	}()

	return idle
}

func (req *request) wait() (bool, bool) {
	qs := req.qs
<<<<<<< HEAD
	qs.lock.Lock()
=======

	// ========================================================================
	// Step 3:
	// The final step is to wait on a decision from
	// somewhere and then act on it.
	decisionAny := req.decision.Get()
	qs.lockAndSyncTime(req.ctx)
>>>>>>> upstream/master
	defer qs.lock.Unlock()
	if req.waitStarted {
		// This can not happen, because the client is forbidden to
		// call Wait twice on the same request
<<<<<<< HEAD
		panic(fmt.Sprintf("Multiple calls to the Wait method, QueueSet=%s, startTime=%s, descr1=%#+v, descr2=%#+v", req.qs.qCfg.Name, req.startTime, req.descr1, req.descr2))
	}
	req.waitStarted = true

	// ========================================================================
	// Step 4:
	// The final step is to wait on a decision from
	// somewhere and then act on it.
	decisionAny := req.decision.GetLocked()
	qs.syncTimeLocked()
	decision, isDecision := decisionAny.(requestDecision)
	if !isDecision {
		panic(fmt.Sprintf("QS(%s): Impossible decision %#+v (of type %T) for request %#+v %#+v", qs.qCfg.Name, decisionAny, decisionAny, req.descr1, req.descr2))
	}
	switch decision {
=======
		klog.Errorf("Duplicate call to the Wait method!  Immediately returning execute=false.  QueueSet=%s, startTime=%s, descr1=%#+v, descr2=%#+v", req.qs.qCfg.Name, req.startTime, req.descr1, req.descr2)
		return false, qs.isIdleLocked()
	}
	req.waitStarted = true
	switch decisionAny {
>>>>>>> upstream/master
	case decisionReject:
		klog.V(5).Infof("QS(%s): request %#+v %#+v timed out after being enqueued\n", qs.qCfg.Name, req.descr1, req.descr2)
		metrics.AddReject(req.ctx, qs.qCfg.Name, req.fsName, "time-out")
		return false, qs.isIdleLocked()
	case decisionCancel:
<<<<<<< HEAD
		// TODO(aaron-prindle) add metrics for this case
		klog.V(5).Infof("QS(%s): Ejecting request %#+v %#+v from its queue", qs.qCfg.Name, req.descr1, req.descr2)
		return false, qs.isIdleLocked()
=======
>>>>>>> upstream/master
	case decisionExecute:
		klog.V(5).Infof("QS(%s): Dispatching request %#+v %#+v from its queue", qs.qCfg.Name, req.descr1, req.descr2)
		return true, false
	default:
		// This can not happen, all possible values are handled above
<<<<<<< HEAD
		panic(decision)
	}
=======
		klog.Errorf("QS(%s): Impossible decision (type %T, value %#+v) for request %#+v %#+v!  Treating as cancel", qs.qCfg.Name, decisionAny, decisionAny, req.descr1, req.descr2)
	}
	// TODO(aaron-prindle) add metrics for this case
	klog.V(5).Infof("QS(%s): Ejecting request %#+v %#+v from its queue", qs.qCfg.Name, req.descr1, req.descr2)
	// remove the request from the queue as it has timed out
	if req.removeFromQueueLocked() != nil {
		qs.totRequestsWaiting--
		metrics.AddReject(req.ctx, qs.qCfg.Name, req.fsName, "cancelled")
		metrics.AddRequestsInQueues(req.ctx, qs.qCfg.Name, req.fsName, -1)
		req.NoteQueued(false)
		qs.reqsObsPair.RequestsWaiting.Add(-1)
	}
	return false, qs.isIdleLocked()
>>>>>>> upstream/master
}

func (qs *queueSet) IsIdle() bool {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	return qs.isIdleLocked()
}

func (qs *queueSet) isIdleLocked() bool {
	return qs.totRequestsWaiting == 0 && qs.totRequestsExecuting == 0
}

// lockAndSyncTime acquires the lock and updates the virtual time.
<<<<<<< HEAD
// Doing them together avoids the mistake of modify some queue state
// before calling syncTimeLocked.
func (qs *queueSet) lockAndSyncTime() {
	qs.lock.Lock()
	qs.syncTimeLocked()
=======
// Doing them together avoids the mistake of modifying some queue state
// before calling syncTimeLocked.
func (qs *queueSet) lockAndSyncTime(ctx context.Context) {
	qs.lock.Lock()
	qs.syncTimeLocked(ctx)
>>>>>>> upstream/master
}

// syncTimeLocked updates the virtual time based on the assumption
// that the current state of the queues has been in effect since
// `qs.lastRealTime`.  Thus, it should be invoked after acquiring the
// lock and before modifying the state of any queue.
<<<<<<< HEAD
func (qs *queueSet) syncTimeLocked() {
	realNow := qs.clock.Now()
	timeSinceLast := realNow.Sub(qs.lastRealTime).Seconds()
	qs.lastRealTime = realNow
	qs.virtualTime += timeSinceLast * qs.getVirtualTimeRatioLocked()
	metrics.SetCurrentR(qs.qCfg.Name, qs.virtualTime)
=======
func (qs *queueSet) syncTimeLocked(ctx context.Context) {
	realNow := qs.clock.Now()
	timeSinceLast := realNow.Sub(qs.lastRealTime)
	qs.lastRealTime = realNow
	prevR := qs.currentR
	incrR := fqrequest.SeatsTimesDuration(qs.getVirtualTimeRatioLocked(), timeSinceLast)
	qs.currentR = prevR + incrR
	switch {
	case prevR > qs.currentR:
		klog.ErrorS(errors.New("queueset::currentR overflow"), "Overflow", "QS", qs.qCfg.Name, "when", realNow.Format(nsTimeFmt), "prevR", prevR, "incrR", incrR, "currentR", qs.currentR)
	case qs.currentR >= highR:
		qs.advanceEpoch(ctx, realNow, incrR)
	}
	metrics.SetCurrentR(qs.qCfg.Name, qs.currentR.ToFloat())
}

// rDecrement is the amount by which the progress meter R is wound backwards
// when needed to avoid overflow.
const rDecrement = fqrequest.MaxSeatSeconds / 2

// highR is the threshold that triggers advance of the epoch.
// That is, decrementing the global progress meter R by rDecrement.
const highR = rDecrement + rDecrement/2

// advanceEpoch subtracts rDecrement from the global progress meter R
// and all the readings that have been taked from that meter.
// The now and incrR parameters are only used to add info to the log messages.
func (qs *queueSet) advanceEpoch(ctx context.Context, now time.Time, incrR fqrequest.SeatSeconds) {
	oldR := qs.currentR
	qs.currentR -= rDecrement
	klog.InfoS("Advancing epoch", "QS", qs.qCfg.Name, "when", now.Format(nsTimeFmt), "oldR", oldR, "newR", qs.currentR, "incrR", incrR)
	success := true
	for qIdx, queue := range qs.queues {
		if queue.requests.Length() == 0 && queue.requestsExecuting == 0 {
			// Do not just decrement, the value could be quite outdated.
			// It is safe to reset to zero in this case, because the next request
			// will overwrite the zero with `qs.currentR`.
			queue.nextDispatchR = 0
			continue
		}
		oldNextDispatchR := queue.nextDispatchR
		queue.nextDispatchR -= rDecrement
		if queue.nextDispatchR > oldNextDispatchR {
			klog.ErrorS(errors.New("queue::nextDispatchR underflow"), "Underflow", "QS", qs.qCfg.Name, "queue", qIdx, "oldNextDispatchR", oldNextDispatchR, "newNextDispatchR", queue.nextDispatchR, "incrR", incrR)
			success = false
		}
		queue.requests.Walk(func(req *request) bool {
			oldArrivalR := req.arrivalR
			req.arrivalR -= rDecrement
			if req.arrivalR > oldArrivalR {
				klog.ErrorS(errors.New("request::arrivalR underflow"), "Underflow", "QS", qs.qCfg.Name, "queue", qIdx, "request", *req, "oldArrivalR", oldArrivalR, "incrR", incrR)
				success = false
			}
			return true
		})
	}
	metrics.AddEpochAdvance(ctx, qs.qCfg.Name, success)
>>>>>>> upstream/master
}

// getVirtualTimeRatio calculates the rate at which virtual time has
// been advancing, according to the logic in `doc.go`.
func (qs *queueSet) getVirtualTimeRatioLocked() float64 {
	activeQueues := 0
	seatsRequested := 0
	for _, queue := range qs.queues {
<<<<<<< HEAD
		seatsRequested += (queue.seatsInUse + queue.requests.SeatsSum())
=======
		// here we want the sum of the maximum width of the requests in this queue since our
		// goal is to find the maximum rate at which the queue could work.
		seatsRequested += (queue.seatsInUse + queue.requests.QueueSum().MaxSeatsSum)
>>>>>>> upstream/master
		if queue.requests.Length() > 0 || queue.requestsExecuting > 0 {
			activeQueues++
		}
	}
	if activeQueues == 0 {
		return 0
	}
	return math.Min(float64(seatsRequested), float64(qs.dCfg.ConcurrencyLimit)) / float64(activeQueues)
}

// timeoutOldRequestsAndRejectOrEnqueueLocked encapsulates the logic required
// to validate and enqueue a request for the queueSet/QueueSet:
// 1) Start with shuffle sharding, to pick a queue.
// 2) Reject old requests that have been waiting too long
// 3) Reject current request if there is not enough concurrency shares and
// we are at max queue length
// 4) If not rejected, create a request and enqueue
// returns the enqueud request on a successful enqueue
// returns nil in the case that there is no available concurrency or
// the queuelengthlimit has been reached
<<<<<<< HEAD
func (qs *queueSet) timeoutOldRequestsAndRejectOrEnqueueLocked(ctx context.Context, width *fqrequest.Width, hashValue uint64, flowDistinguisher, fsName string, descr1, descr2 interface{}, queueNoteFn fq.QueueNoteFn) *request {
	// Start with the shuffle sharding, to pick a queue.
	queueIdx := qs.chooseQueueIndexLocked(hashValue, descr1, descr2)
=======
func (qs *queueSet) timeoutOldRequestsAndRejectOrEnqueueLocked(ctx context.Context, workEstimate *fqrequest.WorkEstimate, hashValue uint64, flowDistinguisher, fsName string, descr1, descr2 interface{}, queueNoteFn fq.QueueNoteFn) *request {
	// Start with the shuffle sharding, to pick a queue.
	queueIdx := qs.shuffleShardLocked(hashValue, descr1, descr2)
>>>>>>> upstream/master
	queue := qs.queues[queueIdx]
	// The next step is the logic to reject requests that have been waiting too long
	qs.removeTimedOutRequestsFromQueueLocked(queue, fsName)
	// NOTE: currently timeout is only checked for each new request.  This means that there can be
	// requests that are in the queue longer than the timeout if there are no new requests
	// We prefer the simplicity over the promptness, at least for now.

<<<<<<< HEAD
=======
	defer qs.boundNextDispatchLocked(queue)

>>>>>>> upstream/master
	// Create a request and enqueue
	req := &request{
		qs:                qs,
		fsName:            fsName,
		flowDistinguisher: flowDistinguisher,
		ctx:               ctx,
<<<<<<< HEAD
		decision:          lockingpromise.NewWriteOnce(&qs.lock, qs.counter),
		arrivalTime:       qs.clock.Now(),
=======
		decision:          qs.promiseFactory(nil, ctx.Done(), decisionCancel),
		arrivalTime:       qs.clock.Now(),
		arrivalR:          qs.currentR,
>>>>>>> upstream/master
		queue:             queue,
		descr1:            descr1,
		descr2:            descr2,
		queueNoteFn:       queueNoteFn,
<<<<<<< HEAD
		width:             *width,
=======
		workEstimate:      qs.completeWorkEstimate(workEstimate),
>>>>>>> upstream/master
	}
	if ok := qs.rejectOrEnqueueLocked(req); !ok {
		return nil
	}
	metrics.ObserveQueueLength(ctx, qs.qCfg.Name, fsName, queue.requests.Length())
	return req
}

<<<<<<< HEAD
// chooseQueueIndexLocked uses shuffle sharding to select a queue index
// using the given hashValue and the shuffle sharding parameters of the queueSet.
func (qs *queueSet) chooseQueueIndexLocked(hashValue uint64, descr1, descr2 interface{}) int {
	bestQueueIdx := -1
	bestQueueSeatsSum := int(math.MaxInt32)
	// the dealer uses the current desired number of queues, which is no larger than the number in `qs.queues`.
	qs.dealer.Deal(hashValue, func(queueIdx int) {
		// TODO: Consider taking into account `additional latency` of requests
		// in addition to their widths.
		// Ideally, this should be based on projected completion time in the
		// virtual world of the youngest request in the queue.
		thisSeatsSum := qs.queues[queueIdx].requests.SeatsSum()
		klog.V(7).Infof("QS(%s): For request %#+v %#+v considering queue %d of seatsSum %d", qs.qCfg.Name, descr1, descr2, queueIdx, thisSeatsSum)
		if thisSeatsSum < bestQueueSeatsSum {
			bestQueueIdx, bestQueueSeatsSum = queueIdx, thisSeatsSum
		}
	})
	klog.V(6).Infof("QS(%s) at r=%s v=%.9fs: For request %#+v %#+v chose queue %d, had %d waiting & %d executing", qs.qCfg.Name, qs.clock.Now().Format(nsTimeFmt), qs.virtualTime, descr1, descr2, bestQueueIdx, bestQueueSeatsSum, qs.queues[bestQueueIdx].requestsExecuting)
=======
// shuffleShardLocked uses shuffle sharding to select a queue index
// using the given hashValue and the shuffle sharding parameters of the queueSet.
func (qs *queueSet) shuffleShardLocked(hashValue uint64, descr1, descr2 interface{}) int {
	var backHand [8]int
	// Deal into a data structure, so that the order of visit below is not necessarily the order of the deal.
	// This removes bias in the case of flows with overlapping hands.
	hand := qs.dealer.DealIntoHand(hashValue, backHand[:])
	handSize := len(hand)
	offset := qs.enqueues % handSize
	qs.enqueues++
	bestQueueIdx := -1
	minQueueSeatSeconds := fqrequest.MaxSeatSeconds
	for i := 0; i < handSize; i++ {
		queueIdx := hand[(offset+i)%handSize]
		queue := qs.queues[queueIdx]
		queueSum := queue.requests.QueueSum()

		// this is the total amount of work in seat-seconds for requests
		// waiting in this queue, we will select the queue with the minimum.
		thisQueueSeatSeconds := queueSum.TotalWorkSum
		klog.V(7).Infof("QS(%s): For request %#+v %#+v considering queue %d with sum: %#v and %d seats in use, nextDispatchR=%v", qs.qCfg.Name, descr1, descr2, queueIdx, queueSum, queue.seatsInUse, queue.nextDispatchR)
		if thisQueueSeatSeconds < minQueueSeatSeconds {
			minQueueSeatSeconds = thisQueueSeatSeconds
			bestQueueIdx = queueIdx
		}
	}
	if klog.V(6).Enabled() {
		chosenQueue := qs.queues[bestQueueIdx]
		klog.V(6).Infof("QS(%s) at t=%s R=%v: For request %#+v %#+v chose queue %d, with sum: %#v & %d seats in use & nextDispatchR=%v", qs.qCfg.Name, qs.clock.Now().Format(nsTimeFmt), qs.currentR, descr1, descr2, bestQueueIdx, chosenQueue.requests.QueueSum(), chosenQueue.seatsInUse, chosenQueue.nextDispatchR)
	}
>>>>>>> upstream/master
	return bestQueueIdx
}

// removeTimedOutRequestsFromQueueLocked rejects old requests that have been enqueued
// past the requestWaitLimit
func (qs *queueSet) removeTimedOutRequestsFromQueueLocked(queue *queue, fsName string) {
	timeoutCount := 0
	now := qs.clock.Now()
	reqs := queue.requests
	// reqs are sorted oldest -> newest
	// can short circuit loop (break) if oldest requests are not timing out
	// as newer requests also will not have timed out

<<<<<<< HEAD
	// now - requestWaitLimit = waitLimit
	waitLimit := now.Add(-qs.qCfg.RequestWaitLimit)
	reqs.Walk(func(req *request) bool {
		if waitLimit.After(req.arrivalTime) {
			req.decision.SetLocked(decisionReject)
			timeoutCount++
			metrics.AddRequestsInQueues(req.ctx, qs.qCfg.Name, req.fsName, -1)
			req.NoteQueued(false)

			// we need to check if the next request has timed out.
			return true
		}

=======
	// now - requestWaitLimit = arrivalLimit
	arrivalLimit := now.Add(-qs.qCfg.RequestWaitLimit)
	reqs.Walk(func(req *request) bool {
		if arrivalLimit.After(req.arrivalTime) {
			if req.decision.Set(decisionReject) && req.removeFromQueueLocked() != nil {
				timeoutCount++
				req.NoteQueued(false)
				metrics.AddRequestsInQueues(req.ctx, qs.qCfg.Name, req.fsName, -1)
			}
			// we need to check if the next request has timed out.
			return true
		}
>>>>>>> upstream/master
		// since reqs are sorted oldest -> newest, we are done here.
		return false
	})

	// remove timed out requests from queue
	if timeoutCount > 0 {
<<<<<<< HEAD
		// The number of requests we have timed out is timeoutCount,
		// so, let's dequeue the exact number of requests for this queue.
		for i := 0; i < timeoutCount; i++ {
			queue.requests.Dequeue()
		}
		// decrement the # of requestsEnqueued
		qs.totRequestsWaiting -= timeoutCount
		qs.obsPair.RequestsWaiting.Add(float64(-timeoutCount))
=======
		qs.totRequestsWaiting -= timeoutCount
		qs.reqsObsPair.RequestsWaiting.Add(float64(-timeoutCount))
>>>>>>> upstream/master
	}
}

// rejectOrEnqueueLocked rejects or enqueues the newly arrived
// request, which has been assigned to a queue.  If up against the
// queue length limit and the concurrency limit then returns false.
// Otherwise enqueues and returns true.
func (qs *queueSet) rejectOrEnqueueLocked(request *request) bool {
	queue := request.queue
	curQueueLength := queue.requests.Length()
	// rejects the newly arrived request if resource criteria not met
	if qs.totSeatsInUse >= qs.dCfg.ConcurrencyLimit &&
		curQueueLength >= qs.qCfg.QueueLengthLimit {
		return false
	}

	qs.enqueueLocked(request)
	return true
}

// enqueues a request into its queue.
func (qs *queueSet) enqueueLocked(request *request) {
	queue := request.queue
	now := qs.clock.Now()
	if queue.requests.Length() == 0 && queue.requestsExecuting == 0 {
<<<<<<< HEAD
		// the queue’s virtual start time is set to the virtual time.
		queue.virtualStart = qs.virtualTime
		if klog.V(6).Enabled() {
			klog.Infof("QS(%s) at r=%s v=%.9fs: initialized queue %d virtual start time due to request %#+v %#+v", qs.qCfg.Name, now.Format(nsTimeFmt), queue.virtualStart, queue.index, request.descr1, request.descr2)
		}
	}
	queue.Enqueue(request)
	qs.totRequestsWaiting++
	metrics.AddRequestsInQueues(request.ctx, qs.qCfg.Name, request.fsName, 1)
	request.NoteQueued(true)
	qs.obsPair.RequestsWaiting.Add(1)
}

// dispatchAsMuchAsPossibleLocked runs a loop, as long as there
// are non-empty queues and the number currently executing is less than the
// assured concurrency value.  The body of the loop uses the fair queuing
// technique to pick a queue, dequeue the request at the head of that
// queue, increment the count of the number executing, and send true
// to the request's channel.
func (qs *queueSet) dispatchAsMuchAsPossibleLocked() {
	for qs.totRequestsWaiting != 0 && qs.totSeatsInUse < qs.dCfg.ConcurrencyLimit {
		ok := qs.dispatchLocked()
		if !ok {
			break
		}
	}
}

func (qs *queueSet) dispatchSansQueueLocked(ctx context.Context, width *fqrequest.Width, flowDistinguisher, fsName string, descr1, descr2 interface{}) *request {
=======
		// the queue’s start R is set to the virtual time.
		queue.nextDispatchR = qs.currentR
		if klog.V(6).Enabled() {
			klog.Infof("QS(%s) at t=%s R=%v: initialized queue %d start R due to request %#+v %#+v", qs.qCfg.Name, now.Format(nsTimeFmt), queue.nextDispatchR, queue.index, request.descr1, request.descr2)
		}
	}
	request.removeFromQueueLocked = queue.requests.Enqueue(request)
	qs.totRequestsWaiting++
	metrics.AddRequestsInQueues(request.ctx, qs.qCfg.Name, request.fsName, 1)
	request.NoteQueued(true)
	qs.reqsObsPair.RequestsWaiting.Add(1)
}

// dispatchAsMuchAsPossibleLocked does as many dispatches as possible now.
func (qs *queueSet) dispatchAsMuchAsPossibleLocked() {
	for qs.totRequestsWaiting != 0 && qs.totSeatsInUse < qs.dCfg.ConcurrencyLimit && qs.dispatchLocked() {
	}
}

func (qs *queueSet) dispatchSansQueueLocked(ctx context.Context, workEstimate *fqrequest.WorkEstimate, flowDistinguisher, fsName string, descr1, descr2 interface{}) *request {
>>>>>>> upstream/master
	// does not call metrics.SetDispatchMetrics because there is no queuing and thus no interesting virtual world
	now := qs.clock.Now()
	req := &request{
		qs:                qs,
		fsName:            fsName,
		flowDistinguisher: flowDistinguisher,
		ctx:               ctx,
		startTime:         now,
<<<<<<< HEAD
		decision:          lockingpromise.NewWriteOnce(&qs.lock, qs.counter),
		arrivalTime:       now,
		descr1:            descr1,
		descr2:            descr2,
		width:             *width,
	}
	req.decision.SetLocked(decisionExecute)
	qs.totRequestsExecuting++
	qs.totSeatsInUse += req.Seats()
	metrics.AddRequestsExecuting(ctx, qs.qCfg.Name, fsName, 1)
	metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, fsName, req.Seats())
	qs.obsPair.RequestsExecuting.Add(1)
	if klog.V(5).Enabled() {
		klog.Infof("QS(%s) at r=%s v=%.9fs: immediate dispatch of request %q %#+v %#+v, qs will have %d executing", qs.qCfg.Name, now.Format(nsTimeFmt), qs.virtualTime, fsName, descr1, descr2, qs.totRequestsExecuting)
=======
		decision:          qs.promiseFactory(decisionExecute, ctx.Done(), decisionCancel),
		arrivalTime:       now,
		arrivalR:          qs.currentR,
		descr1:            descr1,
		descr2:            descr2,
		workEstimate:      qs.completeWorkEstimate(workEstimate),
	}
	qs.totRequestsExecuting++
	qs.totSeatsInUse += req.MaxSeats()
	metrics.AddRequestsExecuting(ctx, qs.qCfg.Name, fsName, 1)
	metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, fsName, req.MaxSeats())
	qs.reqsObsPair.RequestsExecuting.Add(1)
	qs.execSeatsObs.Add(float64(req.MaxSeats()))
	if klog.V(5).Enabled() {
		klog.Infof("QS(%s) at t=%s R=%v: immediate dispatch of request %q %#+v %#+v, qs will have %d executing", qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, fsName, descr1, descr2, qs.totRequestsExecuting)
>>>>>>> upstream/master
	}
	return req
}

// dispatchLocked uses the Fair Queuing for Server Requests method to
// select a queue and dispatch the oldest request in that queue.  The
<<<<<<< HEAD
// return value indicates whether a request was dispatched; this will
// be false when there are no requests waiting in any queue.
func (qs *queueSet) dispatchLocked() bool {
	queue := qs.selectQueueLocked()
	if queue == nil {
		return false
	}
	request, ok := queue.Dequeue()
	if !ok { // This should never happen.  But if it does...
		return false
	}
=======
// return value indicates whether a request was dequeued; this will
// be false when either all queues are empty or the request at the head
// of the next queue cannot be dispatched.
func (qs *queueSet) dispatchLocked() bool {
	queue, request := qs.findDispatchQueueLocked()
	if queue == nil {
		return false
	}
	if request == nil { // This should never happen.  But if it does...
		return false
	}
	qs.totRequestsWaiting--
	metrics.AddRequestsInQueues(request.ctx, qs.qCfg.Name, request.fsName, -1)
	request.NoteQueued(false)
	qs.reqsObsPair.RequestsWaiting.Add(-1)
	defer qs.boundNextDispatchLocked(queue)
	if !request.decision.Set(decisionExecute) {
		return true
	}
>>>>>>> upstream/master
	request.startTime = qs.clock.Now()
	// At this moment the request leaves its queue and starts
	// executing.  We do not recognize any interim state between
	// "queued" and "executing".  While that means "executing"
	// includes a little overhead from this package, this is not a
	// problem because other overhead is also included.
<<<<<<< HEAD
	qs.totRequestsWaiting--
	qs.totRequestsExecuting++
	qs.totSeatsInUse += request.Seats()
	queue.requestsExecuting++
	queue.seatsInUse += request.Seats()
	metrics.AddRequestsInQueues(request.ctx, qs.qCfg.Name, request.fsName, -1)
	request.NoteQueued(false)
	metrics.AddRequestsExecuting(request.ctx, qs.qCfg.Name, request.fsName, 1)
	metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, request.fsName, request.Seats())
	qs.obsPair.RequestsWaiting.Add(-1)
	qs.obsPair.RequestsExecuting.Add(1)
	if klog.V(6).Enabled() {
		klog.Infof("QS(%s) at r=%s v=%.9fs: dispatching request %#+v %#+v from queue %d with virtual start time %.9fs, queue will have %d waiting & %d executing",
			qs.qCfg.Name, request.startTime.Format(nsTimeFmt), qs.virtualTime, request.descr1, request.descr2,
			queue.index, queue.virtualStart, queue.requests.Length(), queue.requestsExecuting)
	}
	// When a request is dequeued for service -> qs.virtualStart += G
	queue.virtualStart += qs.estimatedServiceTime * float64(request.Seats())
	request.decision.SetLocked(decisionExecute)
	return ok
}

// cancelWait ensures the request is not waiting.  This is only
// applicable to a request that has been assigned to a queue.
func (qs *queueSet) cancelWait(req *request) {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	if req.decision.IsSetLocked() {
		// The request has already been removed from the queue
		// and so we consider its wait to be over.
		return
	}
	req.decision.SetLocked(decisionCancel)

	// remove the request from the queue as it has timed out
	req.removeFromQueueFn()
	qs.totRequestsWaiting--
	metrics.AddRequestsInQueues(req.ctx, qs.qCfg.Name, req.fsName, -1)
	req.NoteQueued(false)
	qs.obsPair.RequestsWaiting.Add(-1)
=======
	qs.totRequestsExecuting++
	qs.totSeatsInUse += request.MaxSeats()
	queue.requestsExecuting++
	queue.seatsInUse += request.MaxSeats()
	metrics.AddRequestsExecuting(request.ctx, qs.qCfg.Name, request.fsName, 1)
	metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, request.fsName, request.MaxSeats())
	qs.reqsObsPair.RequestsExecuting.Add(1)
	qs.execSeatsObs.Add(float64(request.MaxSeats()))
	if klog.V(6).Enabled() {
		klog.Infof("QS(%s) at t=%s R=%v: dispatching request %#+v %#+v work %v from queue %d with start R %v, queue will have %d waiting & %d requests occupying %d seats, set will have %d seats occupied",
			qs.qCfg.Name, request.startTime.Format(nsTimeFmt), qs.currentR, request.descr1, request.descr2,
			request.workEstimate, queue.index, queue.nextDispatchR, queue.requests.Length(), queue.requestsExecuting, queue.seatsInUse, qs.totSeatsInUse)
	}
	// When a request is dequeued for service -> qs.virtualStart += G * width
	queue.nextDispatchR += request.totalWork()
	return true
>>>>>>> upstream/master
}

// canAccommodateSeatsLocked returns true if this queueSet has enough
// seats available to accommodate a request with the given number of seats,
// otherwise it returns false.
func (qs *queueSet) canAccommodateSeatsLocked(seats int) bool {
	switch {
	case seats > qs.dCfg.ConcurrencyLimit:
		// we have picked the queue with the minimum virtual finish time, but
		// the number of seats this request asks for exceeds the concurrency limit.
		// TODO: this is a quick fix for now, once we have borrowing in place we will not need it
		if qs.totRequestsExecuting == 0 {
			// TODO: apply additional lateny associated with this request, as described in the KEP
			return true
		}
		// wait for all "currently" executing requests in this queueSet
		// to finish before we can execute this request.
<<<<<<< HEAD
		if klog.V(4).Enabled() {
			klog.Infof("QS(%s): seats (%d) asked for exceeds concurrency limit, waiting for currently executing requests to complete, %d seats are in use (%d are executing) and the limit is %d",
				qs.qCfg.Name, seats, qs.totSeatsInUse, qs.totRequestsExecuting, qs.dCfg.ConcurrencyLimit)
		}
=======
>>>>>>> upstream/master
		return false
	case qs.totSeatsInUse+seats > qs.dCfg.ConcurrencyLimit:
		return false
	}

	return true
}

<<<<<<< HEAD
// selectQueueLocked examines the queues in round robin order and
// returns the first one of those for which the virtual finish time of
// the oldest waiting request is minimal.
func (qs *queueSet) selectQueueLocked() *queue {
	minVirtualFinish := math.Inf(1)
	sMin := math.Inf(1)
	dsMin := math.Inf(1)
	sMax := math.Inf(-1)
	dsMax := math.Inf(-1)
=======
// findDispatchQueueLocked examines the queues in round robin order and
// returns the first one of those for which the virtual finish time of
// the oldest waiting request is minimal, and also returns that request.
// Returns nils if the head of the selected queue can not be dispatched now.
func (qs *queueSet) findDispatchQueueLocked() (*queue, *request) {
	minVirtualFinish := fqrequest.MaxSeatSeconds
	sMin := fqrequest.MaxSeatSeconds
	dsMin := fqrequest.MaxSeatSeconds
	sMax := fqrequest.MinSeatSeconds
	dsMax := fqrequest.MinSeatSeconds
>>>>>>> upstream/master
	var minQueue *queue
	var minIndex int
	nq := len(qs.queues)
	for range qs.queues {
		qs.robinIndex = (qs.robinIndex + 1) % nq
		queue := qs.queues[qs.robinIndex]
<<<<<<< HEAD
		if queue.requests.Length() != 0 {
			sMin = math.Min(sMin, queue.virtualStart)
			sMax = math.Max(sMax, queue.virtualStart)
			estimatedWorkInProgress := qs.estimatedServiceTime * float64(queue.seatsInUse)
			dsMin = math.Min(dsMin, queue.virtualStart-estimatedWorkInProgress)
			dsMax = math.Max(dsMax, queue.virtualStart-estimatedWorkInProgress)
			// the virtual finish time of the oldest request is:
			//   virtual start time + G
			// we are not taking the width of the request into account when
			// we calculate the virtual finish time of the request because
			// it can starve requests with smaller wdith in other queues.
			//
			// so let's draw an example of the starving scenario:
			//  - G=60 (estimated service time in seconds)
			//  - concurrency limit=2
			//  - we have two queues, q1 and q2
			//  - q1 has an infinite supply of requests with width W=1
			//  - q2 has one request waiting in the queue with width W=2
			//  - virtual start time for both q1 and q2 are at t0
			//  - requests complete really fast, S=1ms on q1
			// in this scenario we will execute roughly 60,000 requests
			// from q1 before we pick the request from q2.
			currentVirtualFinish := queue.virtualStart + qs.estimatedServiceTime

=======
		oldestWaiting, _ := queue.requests.Peek()
		if oldestWaiting != nil {
			sMin = ssMin(sMin, queue.nextDispatchR)
			sMax = ssMax(sMax, queue.nextDispatchR)
			estimatedWorkInProgress := fqrequest.SeatsTimesDuration(float64(queue.seatsInUse), qs.estimatedServiceDuration)
			dsMin = ssMin(dsMin, queue.nextDispatchR-estimatedWorkInProgress)
			dsMax = ssMax(dsMax, queue.nextDispatchR-estimatedWorkInProgress)
			currentVirtualFinish := queue.nextDispatchR + oldestWaiting.totalWork()
			klog.V(11).InfoS("Considering queue to dispatch", "queueSet", qs.qCfg.Name, "queue", qs.robinIndex, "finishR", currentVirtualFinish)
>>>>>>> upstream/master
			if currentVirtualFinish < minVirtualFinish {
				minVirtualFinish = currentVirtualFinish
				minQueue = queue
				minIndex = qs.robinIndex
			}
		}
	}

<<<<<<< HEAD
	// TODO: add a method to fifo that lets us peek at the oldest request
	var oldestReqFromMinQueue *request
	minQueue.requests.Walk(func(r *request) bool {
		oldestReqFromMinQueue = r
		return false
	})
	if oldestReqFromMinQueue == nil || !qs.canAccommodateSeatsLocked(oldestReqFromMinQueue.Seats()) {
		// since we have not picked the queue with the minimum virtual finish
		// time, we are not going to advance the round robin index here.
		return nil
=======
	oldestReqFromMinQueue, _ := minQueue.requests.Peek()
	if oldestReqFromMinQueue == nil {
		// This cannot happen
		klog.ErrorS(errors.New("selected queue is empty"), "Impossible", "queueSet", qs.qCfg.Name)
		return nil, nil
	}
	if !qs.canAccommodateSeatsLocked(oldestReqFromMinQueue.MaxSeats()) {
		// since we have not picked the queue with the minimum virtual finish
		// time, we are not going to advance the round robin index here.
		if klog.V(4).Enabled() {
			klog.Infof("QS(%s): request %v %v seats %d cannot be dispatched from queue %d, waiting for currently executing requests to complete, %d requests are occupying %d seats and the limit is %d",
				qs.qCfg.Name, oldestReqFromMinQueue.descr1, oldestReqFromMinQueue.descr2, oldestReqFromMinQueue.MaxSeats(), minQueue.index, qs.totRequestsExecuting, qs.totSeatsInUse, qs.dCfg.ConcurrencyLimit)
		}
		return nil, nil
	}
	oldestReqFromMinQueue.removeFromQueueLocked()

	// If the requested final seats exceed capacity of that queue,
	// we reduce them to current capacity and adjust additional latency
	// to preserve the total amount of work.
	if oldestReqFromMinQueue.workEstimate.FinalSeats > uint(qs.dCfg.ConcurrencyLimit) {
		finalSeats := uint(qs.dCfg.ConcurrencyLimit)
		additionalLatency := oldestReqFromMinQueue.workEstimate.finalWork.DurationPerSeat(float64(finalSeats))
		oldestReqFromMinQueue.workEstimate.FinalSeats = finalSeats
		oldestReqFromMinQueue.workEstimate.AdditionalLatency = additionalLatency
>>>>>>> upstream/master
	}

	// we set the round robin indexing to start at the chose queue
	// for the next round.  This way the non-selected queues
	// win in the case that the virtual finish times are the same
	qs.robinIndex = minIndex
<<<<<<< HEAD
	// according to the original FQ formula:
	//
	//   Si = MAX(R(t), Fi-1)
	//
	// the virtual start (excluding the estimated cost) of the chose
	// queue should always be greater or equal to the global virtual
	// time.
	//
	// hence we're refreshing the per-queue virtual time for the chosen
	// queue here. if the last virtual start time (excluded estimated cost)
	// falls behind the global virtual time, we update the latest virtual
	// start by: <latest global virtual time> + <previously estimated cost>
	previouslyEstimatedServiceTime := float64(minQueue.seatsInUse) * qs.estimatedServiceTime
	if qs.virtualTime > minQueue.virtualStart-previouslyEstimatedServiceTime {
		// per-queue virtual time should not fall behind the global
		minQueue.virtualStart = qs.virtualTime + previouslyEstimatedServiceTime
	}
	metrics.SetDispatchMetrics(qs.qCfg.Name, qs.virtualTime, minQueue.virtualStart, sMin, sMax, dsMin, dsMax)
	return minQueue
=======

	if minQueue.nextDispatchR < oldestReqFromMinQueue.arrivalR {
		klog.ErrorS(errors.New("dispatch before arrival"), "Inconceivable!", "QS", qs.qCfg.Name, "queue", minQueue.index, "dispatchR", minQueue.nextDispatchR, "request", oldestReqFromMinQueue)
	}
	metrics.SetDispatchMetrics(qs.qCfg.Name, qs.currentR.ToFloat(), minQueue.nextDispatchR.ToFloat(), sMin.ToFloat(), sMax.ToFloat(), dsMin.ToFloat(), dsMax.ToFloat())
	return minQueue, oldestReqFromMinQueue
}

func ssMin(a, b fqrequest.SeatSeconds) fqrequest.SeatSeconds {
	if a > b {
		return b
	}
	return a
}

func ssMax(a, b fqrequest.SeatSeconds) fqrequest.SeatSeconds {
	if a < b {
		return b
	}
	return a
>>>>>>> upstream/master
}

// finishRequestAndDispatchAsMuchAsPossible is a convenience method
// which calls finishRequest for a given request and then dispatches
// as many requests as possible.  This is all of what needs to be done
// once a request finishes execution or is canceled.  This returns a bool
// indicating whether the QueueSet is now idle.
func (qs *queueSet) finishRequestAndDispatchAsMuchAsPossible(req *request) bool {
<<<<<<< HEAD
	qs.lockAndSyncTime()
=======
	qs.lockAndSyncTime(req.ctx)
>>>>>>> upstream/master
	defer qs.lock.Unlock()

	qs.finishRequestLocked(req)
	qs.dispatchAsMuchAsPossibleLocked()
	return qs.isIdleLocked()
}

// finishRequestLocked is a callback that should be used when a
// previously dispatched request has completed it's service.  This
// callback updates important state in the queueSet
func (qs *queueSet) finishRequestLocked(r *request) {
	now := qs.clock.Now()
	qs.totRequestsExecuting--
<<<<<<< HEAD
	qs.totSeatsInUse -= r.Seats()
	metrics.AddRequestsExecuting(r.ctx, qs.qCfg.Name, r.fsName, -1)
	metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, r.fsName, -r.Seats())
	qs.obsPair.RequestsExecuting.Add(-1)

	if r.queue == nil {
		if klog.V(6).Enabled() {
			klog.Infof("QS(%s) at r=%s v=%.9fs: request %#+v %#+v finished, qs will have %d executing", qs.qCfg.Name, now.Format(nsTimeFmt), qs.virtualTime, r.descr1, r.descr2, qs.totRequestsExecuting)
		}
		return
	}

	S := now.Sub(r.startTime).Seconds()

	// When a request finishes being served, and the actual service time was S,
	// the queue’s virtual start time is decremented by (G - S)*width.
	r.queue.virtualStart -= (qs.estimatedServiceTime - S) * float64(r.Seats())

	// request has finished, remove from requests executing
	r.queue.requestsExecuting--
	r.queue.seatsInUse -= r.Seats()

	if klog.V(6).Enabled() {
		klog.Infof("QS(%s) at r=%s v=%.9fs: request %#+v %#+v finished, adjusted queue %d virtual start time to %.9fs due to service time %.9fs, queue will have %d waiting & %d executing",
			qs.qCfg.Name, now.Format(nsTimeFmt), qs.virtualTime, r.descr1, r.descr2, r.queue.index,
			r.queue.virtualStart, S, r.queue.requests.Length(), r.queue.requestsExecuting)
=======
	metrics.AddRequestsExecuting(r.ctx, qs.qCfg.Name, r.fsName, -1)
	qs.reqsObsPair.RequestsExecuting.Add(-1)

	actualServiceDuration := now.Sub(r.startTime)

	// TODO: for now we keep the logic localized so it is easier to see
	//  how the counters are tracked for queueset and queue, in future we
	//  can refactor to move this function.
	releaseSeatsLocked := func() {
		defer qs.removeQueueIfEmptyLocked(r)

		qs.totSeatsInUse -= r.MaxSeats()
		metrics.AddRequestConcurrencyInUse(qs.qCfg.Name, r.fsName, -r.MaxSeats())
		qs.execSeatsObs.Add(-float64(r.MaxSeats()))
		if r.queue != nil {
			r.queue.seatsInUse -= r.MaxSeats()
		}
	}

	defer func() {
		if r.workEstimate.AdditionalLatency <= 0 {
			// release the seats allocated to this request immediately
			releaseSeatsLocked()
			if !klog.V(6).Enabled() {
			} else if r.queue != nil {
				klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished all use of %d seats, adjusted queue %d start R to %v due to service time %.9fs, queue will have %d requests with %#v waiting & %d requests occupying %d seats",
					qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.MaxSeats(), r.queue.index,
					r.queue.nextDispatchR, actualServiceDuration.Seconds(), r.queue.requests.Length(), r.queue.requests.QueueSum(), r.queue.requestsExecuting, r.queue.seatsInUse)
			} else {
				klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished all use of %d seats, qs will have %d requests occupying %d seats", qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.InitialSeats, qs.totRequestsExecuting, qs.totSeatsInUse)
			}
			return
		}

		additionalLatency := r.workEstimate.AdditionalLatency
		if !klog.V(6).Enabled() {
		} else if r.queue != nil {
			klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished main use of %d seats but lingering on %d seats for %v seconds, adjusted queue %d start R to %v due to service time %.9fs, queue will have %d requests with %#v waiting & %d requests occupying %d seats",
				qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.InitialSeats, r.workEstimate.FinalSeats, additionalLatency.Seconds(), r.queue.index,
				r.queue.nextDispatchR, actualServiceDuration.Seconds(), r.queue.requests.Length(), r.queue.requests.QueueSum(), r.queue.requestsExecuting, r.queue.seatsInUse)
		} else {
			klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished main use of %d seats but lingering on %d seats for %v seconds, qs will have %d requests occupying %d seats", qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.InitialSeats, r.workEstimate.FinalSeats, additionalLatency.Seconds(), qs.totRequestsExecuting, qs.totSeatsInUse)
		}
		// EventAfterDuration will execute the event func in a new goroutine,
		// so the seats allocated to this request will be released after
		// AdditionalLatency elapses, this ensures that the additional
		// latency has no impact on the user experience.
		qs.clock.EventAfterDuration(func(_ time.Time) {
			qs.lockAndSyncTime(r.ctx)
			defer qs.lock.Unlock()
			now := qs.clock.Now()
			releaseSeatsLocked()
			if !klog.V(6).Enabled() {
			} else if r.queue != nil {
				klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished lingering on %d seats, queue %d will have %d requests with %#v waiting & %d requests occupying %d seats",
					qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.FinalSeats, r.queue.index,
					r.queue.requests.Length(), r.queue.requests.QueueSum(), r.queue.requestsExecuting, r.queue.seatsInUse)
			} else {
				klog.Infof("QS(%s) at t=%s R=%v: request %#+v %#+v finished lingering on %d seats, qs will have %d requests occupying %d seats", qs.qCfg.Name, now.Format(nsTimeFmt), qs.currentR, r.descr1, r.descr2, r.workEstimate.FinalSeats, qs.totRequestsExecuting, qs.totSeatsInUse)
			}
			qs.dispatchAsMuchAsPossibleLocked()
		}, additionalLatency)
	}()

	if r.queue != nil {
		// request has finished, remove from requests executing
		r.queue.requestsExecuting--

		// When a request finishes being served, and the actual service time was S,
		// the queue’s start R is decremented by (G - S)*width.
		r.queue.nextDispatchR -= fqrequest.SeatsTimesDuration(float64(r.InitialSeats()), qs.estimatedServiceDuration-actualServiceDuration)
		qs.boundNextDispatchLocked(r.queue)
	}
}

// boundNextDispatchLocked applies the anti-windup hack.
// We need a hack because all non-empty queues are allocated the same
// number of seats.  A queue that can not use all those seats and does
// not go empty accumulates a progresively earlier `virtualStart` compared
// to queues that are using more than they are allocated.
// The following hack addresses the first side of that inequity,
// by insisting that dispatch in the virtual world not precede arrival.
func (qs *queueSet) boundNextDispatchLocked(queue *queue) {
	oldestReqFromMinQueue, _ := queue.requests.Peek()
	if oldestReqFromMinQueue == nil {
		return
	}
	var virtualStartBound = oldestReqFromMinQueue.arrivalR
	if queue.nextDispatchR < virtualStartBound {
		if klog.V(4).Enabled() {
			klog.InfoS("AntiWindup tweaked queue", "QS", qs.qCfg.Name, "queue", queue.index, "time", qs.clock.Now().Format(nsTimeFmt), "requestDescr1", oldestReqFromMinQueue.descr1, "requestDescr2", oldestReqFromMinQueue.descr2, "newVirtualStart", virtualStartBound, "deltaVirtualStart", (virtualStartBound - queue.nextDispatchR))
		}
		queue.nextDispatchR = virtualStartBound
	}
}

func (qs *queueSet) removeQueueIfEmptyLocked(r *request) {
	if r.queue == nil {
		return
>>>>>>> upstream/master
	}

	// If there are more queues than desired and this one has no
	// requests then remove it
	if len(qs.queues) > qs.qCfg.DesiredNumQueues &&
		r.queue.requests.Length() == 0 &&
		r.queue.requestsExecuting == 0 {
		qs.queues = removeQueueAndUpdateIndexes(qs.queues, r.queue.index)

		// decrement here to maintain the invariant that (qs.robinIndex+1) % numQueues
		// is the index of the next queue after the one last dispatched from
		if qs.robinIndex >= r.queue.index {
			qs.robinIndex--
		}
	}
}

// removeQueueAndUpdateIndexes uses reslicing to remove an index from a slice
// and then updates the 'index' field of the queues to be correct
func removeQueueAndUpdateIndexes(queues []*queue, index int) []*queue {
	keptQueues := append(queues[:index], queues[index+1:]...)
	for i := index; i < len(keptQueues); i++ {
		keptQueues[i].index--
	}
	return keptQueues
}

<<<<<<< HEAD
// preCreateOrUnblockGoroutine needs to be called before creating a
// goroutine associated with this queueSet or unblocking a blocked
// one, to properly update the accounting used in testing.
func (qs *queueSet) preCreateOrUnblockGoroutine() {
	qs.counter.Add(1)
}

// goroutineDoneOrBlocked needs to be called at the end of every
// goroutine associated with this queueSet or when such a goroutine is
// about to wait on some other goroutine to do something; this is to
// properly update the accounting used in testing.
func (qs *queueSet) goroutineDoneOrBlocked() {
	qs.counter.Add(-1)
}

func (qs *queueSet) UpdateObservations() {
	qs.obsPair.RequestsWaiting.Add(0)
	qs.obsPair.RequestsExecuting.Add(0)
=======
func (qs *queueSet) UpdateObservations() {
	qs.reqsObsPair.RequestsWaiting.Add(0)
	qs.reqsObsPair.RequestsExecuting.Add(0)
	qs.execSeatsObs.Add(0)
>>>>>>> upstream/master
}

func (qs *queueSet) Dump(includeRequestDetails bool) debug.QueueSetDump {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	d := debug.QueueSetDump{
		Queues:     make([]debug.QueueDump, len(qs.queues)),
		Waiting:    qs.totRequestsWaiting,
		Executing:  qs.totRequestsExecuting,
		SeatsInUse: qs.totSeatsInUse,
	}
	for i, q := range qs.queues {
<<<<<<< HEAD
		d.Queues[i] = q.dump(includeRequestDetails)
=======
		d.Queues[i] = q.dumpLocked(includeRequestDetails)
>>>>>>> upstream/master
	}
	return d
}
