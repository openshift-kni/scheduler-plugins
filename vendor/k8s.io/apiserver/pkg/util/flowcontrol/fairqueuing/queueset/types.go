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
	"time"

	genericrequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/util/flowcontrol/debug"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	"k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/promise"
	fcrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
)

// request is a temporary container for "requests" with additional
<<<<<<< HEAD
// tracking fields required for the functionality FQScheduler
=======
// tracking fields required for QueueSet functionality.
>>>>>>> upstream/master
type request struct {
	ctx context.Context

	qs *queueSet

	flowDistinguisher string
	fsName            string

	// The relevant queue.  Is nil if this request did not go through
	// a queue.
	queue *queue

<<<<<<< HEAD
	// startTime is the real time when the request began executing
	startTime time.Time

	// width of the request
	width fcrequest.Width
=======
	// estimated amount of work of the request
	workEstimate completedWorkEstimate
>>>>>>> upstream/master

	// decision gets set to a `requestDecision` indicating what to do
	// with this request.  It gets set exactly once, when the request
	// is removed from its queue.  The value will be decisionReject,
<<<<<<< HEAD
	// decisionCancel, or decisionExecute; decisionTryAnother never
	// appears here.
	decision promise.LockingWriteOnce
=======
	// decisionCancel, or decisionExecute.
	//
	// decision.Set is called with the queueSet locked.
	// decision.Get is called without the queueSet locked.
	decision promise.WriteOnce
>>>>>>> upstream/master

	// arrivalTime is the real time when the request entered this system
	arrivalTime time.Time

	// descr1 and descr2 are not used in any logic but they appear in
	// log messages
	descr1, descr2 interface{}

<<<<<<< HEAD
	// Indicates whether client has called Request::Wait()
	waitStarted bool

	queueNoteFn fq.QueueNoteFn

	// Removes this request from its queue. If the request is not put into a
	// a queue it will be nil.
	removeFromQueueFn removeFromFIFOFunc
}

// queue is an array of requests with additional metadata required for
// the FQScheduler
type queue struct {
	// The requests are stored in a FIFO list.
	requests fifo

	// virtualStart is the virtual time (virtual seconds since process
	// startup) when the oldest request in the queue (if there is any)
	// started virtually executing
	virtualStart float64

	requestsExecuting int
	index             int
=======
	queueNoteFn fq.QueueNoteFn

	// The preceding fields are filled in at creation and not modified since;
	// the following fields may be modified later and must only be accessed while
	// holding the queueSet's lock.

	// Removes this request from its queue. If the request is not put into a
	// a queue it will be nil.
	removeFromQueueLocked removeFromFIFOFunc

	// arrivalR is R(arrivalTime).  R is, confusingly, also called "virtual time".
	// This field is meaningful only while the request is waiting in the virtual world.
	arrivalR fcrequest.SeatSeconds

	// startTime is the real time when the request began executing
	startTime time.Time

	// Indicates whether client has called Request::Wait()
	waitStarted bool
}

type completedWorkEstimate struct {
	fcrequest.WorkEstimate
	totalWork fcrequest.SeatSeconds // initial plus final work
	finalWork fcrequest.SeatSeconds // only final work
}

// queue is a sequence of requests that have arrived but not yet finished
// execution in both the real and virtual worlds.
type queue struct {
	// The requests not yet executing in the real world are stored in a FIFO list.
	requests fifo

	// nextDispatchR is the R progress meter reading at
	// which the next request will be dispatched in the virtual world.
	nextDispatchR fcrequest.SeatSeconds

	// requestsExecuting is the count in the real world.
	requestsExecuting int

	// index is the position of this queue among those in its queueSet.
	index int
>>>>>>> upstream/master

	// seatsInUse is the total number of "seats" currently occupied
	// by all the requests that are currently executing in this queue.
	seatsInUse int
}

<<<<<<< HEAD
// Enqueue enqueues a request into the queue and
// sets the removeFromQueueFn of the request appropriately.
func (q *queue) Enqueue(request *request) {
	request.removeFromQueueFn = q.requests.Enqueue(request)
}

// Dequeue dequeues a request from the queue
func (q *queue) Dequeue() (*request, bool) {
	request, ok := q.requests.Dequeue()
	return request, ok
}

func (q *queue) dump(includeDetails bool) debug.QueueDump {
=======
// queueSum tracks the sum of initial seats, max seats, and
// totalWork from all requests in a given queue
type queueSum struct {
	// InitialSeatsSum is the sum of InitialSeats
	// associated with all requests in a given queue.
	InitialSeatsSum int

	// MaxSeatsSum is the sum of MaxSeats
	// associated with all requests in a given queue.
	MaxSeatsSum int

	// TotalWorkSum is the sum of totalWork of the waiting requests
	TotalWorkSum fcrequest.SeatSeconds
}

func (req *request) totalWork() fcrequest.SeatSeconds {
	return req.workEstimate.totalWork
}

func (qs *queueSet) completeWorkEstimate(we *fcrequest.WorkEstimate) completedWorkEstimate {
	finalWork := qs.computeFinalWork(we)
	return completedWorkEstimate{
		WorkEstimate: *we,
		totalWork:    qs.computeInitialWork(we) + finalWork,
		finalWork:    finalWork,
	}
}

func (qs *queueSet) computeInitialWork(we *fcrequest.WorkEstimate) fcrequest.SeatSeconds {
	return fcrequest.SeatsTimesDuration(float64(we.InitialSeats), qs.estimatedServiceDuration)
}

func (qs *queueSet) computeFinalWork(we *fcrequest.WorkEstimate) fcrequest.SeatSeconds {
	return fcrequest.SeatsTimesDuration(float64(we.FinalSeats), we.AdditionalLatency)
}

func (q *queue) dumpLocked(includeDetails bool) debug.QueueDump {
>>>>>>> upstream/master
	digest := make([]debug.RequestDump, q.requests.Length())
	i := 0
	q.requests.Walk(func(r *request) bool {
		// dump requests.
		digest[i].MatchedFlowSchema = r.fsName
		digest[i].FlowDistinguisher = r.flowDistinguisher
		digest[i].ArriveTime = r.arrivalTime
		digest[i].StartTime = r.startTime
<<<<<<< HEAD
=======
		digest[i].WorkEstimate = r.workEstimate.WorkEstimate
>>>>>>> upstream/master
		if includeDetails {
			userInfo, _ := genericrequest.UserFrom(r.ctx)
			digest[i].UserName = userInfo.GetName()
			requestInfo, ok := genericrequest.RequestInfoFrom(r.ctx)
			if ok {
				digest[i].RequestInfo = *requestInfo
			}
		}
		i++
		return true
	})
<<<<<<< HEAD
	return debug.QueueDump{
		VirtualStart:      q.virtualStart,
		Requests:          digest,
		ExecutingRequests: q.requestsExecuting,
		SeatsInUse:        q.seatsInUse,
=======

	sum := q.requests.QueueSum()
	queueSum := debug.QueueSum{
		InitialSeatsSum: sum.InitialSeatsSum,
		MaxSeatsSum:     sum.MaxSeatsSum,
		TotalWorkSum:    sum.TotalWorkSum.String(),
	}

	return debug.QueueDump{
		NextDispatchR:     q.nextDispatchR.String(),
		Requests:          digest,
		ExecutingRequests: q.requestsExecuting,
		SeatsInUse:        q.seatsInUse,
		QueueSum:          queueSum,
>>>>>>> upstream/master
	}
}
