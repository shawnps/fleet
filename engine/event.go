package engine

import (
	log "github.com/coreos/fleet/third_party/github.com/golang/glog"

	"github.com/coreos/fleet/event"
	"github.com/coreos/fleet/job"
)

type EventHandler struct {
	engine *Engine
}

func NewEventHandler(engine *Engine) *EventHandler {
	return &EventHandler{engine}
}

func (self *EventHandler) HandleEventJobCreated(ev event.Event) {
	j := ev.Payload.(job.Job)

	log.V(1).Infof("EventJobCreated(%s): publishing JobOffer", j.Name)
	self.engine.OfferJob(j)
}

func (self *EventHandler) HandleEventJobBidSubmitted(ev event.Event) {
	jb := ev.Payload.(job.JobBid)

	log.V(1).Infof("EventJobBidSubmitted(%s): attempting to schedule Job to Machine(%s)", jb.JobName, jb.MachineName)
	err := self.engine.ResolveJobOffer(jb.JobName, jb.MachineName)
	if err == nil {
		log.V(1).Infof("EventJobBidSubmitted(%s): successfully scheduled Job to Machine(%s)", jb.JobName, jb.MachineName)
	} else {
		log.V(1).Infof("EventJobBidSubmitted(%s): failed to schedule Job to Machine(%s): %s", jb.JobName, jb.MachineName, err.Error())
	}
}

func (self *EventHandler) HandleEventMachineRemoved(ev event.Event) {
	machName := ev.Payload.(string)
	if !self.engine.ClaimMachine(machName) {
		log.V(2).Infof("EventMachineRemoved(%s): failed to lock Machine, ignoring event", machName)
		return
	}

	jobs := self.engine.GetJobsScheduledToMachine(machName)

	for _, j := range jobs {
		log.V(1).Infof("EventMachineRemoved(%s): unscheduling Job(%s)", machName, j.Name)
		self.engine.RemoveJobState(j.Name)
		self.engine.UnscheduleJob(j.Name)
	}

	for _, j := range jobs {
		log.V(1).Infof("EventMachineRemoved(%s): re-publishing JobOffer(%s)", machName, j.Name)
		self.engine.OfferJob(j)
	}
}
