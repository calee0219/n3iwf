package handler

import (
	"time"

	"github.com/sirupsen/logrus"

	"free5gc/src/n3iwf/context"
	n3iwf_message "free5gc/src/n3iwf/handler/message"
	"free5gc/src/n3iwf/ike"
	"free5gc/src/n3iwf/logger"
	"free5gc/src/n3iwf/ngap"
	"free5gc/src/n3iwf/ngap/handler"
	ngap_message "free5gc/src/n3iwf/ngap/message"
	"free5gc/src/n3iwf/relay"
)

var handlerLog *logrus.Entry

func init() {
	// init pool
	handlerLog = logger.HandlerLog
}

func Handle() {
	for {
		select {
		case msg, ok := <-n3iwf_message.N3iwfChannel:
			if ok {
				switch msg.Event {
				case n3iwf_message.EventSCTPConnectMessage:
					handler.HandleEventSCTPConnect(msg.SCTPAddr)
				case n3iwf_message.EventNGAPMessage:
					ngap.Dispatch(msg.SCTPAddr, msg.Value.([]byte))
				case n3iwf_message.EventTimerSendRanConfigUpdateMessage:
					handlerLog.Infof("Re-send Ran Configuration Update Message when waiting time expired")
					self := context.N3IWFSelf()
					self.AMFReInitAvailableList[msg.SCTPAddr] = true
					ngap_message.SendRANConfigurationUpdate(self.AMFPool[msg.SCTPAddr])
				case n3iwf_message.EventN1UDPMessage:
					ike.Dispatch(msg.UDPSendInfo, msg.Value.([]byte))
				case n3iwf_message.EventN1TunnelUPMessage:
					self := context.N3IWFSelf()
					ue, ok := self.AllocatedUEIPAddress[msg.UEInnerIP]
					if !ok {
						handlerLog.Error("UE context not found")
						continue
					}
					relay.ForwardUPTrafficFromN1(ue, msg.Value.([]byte))
				case n3iwf_message.EventGTPMessage:
					self := context.N3IWFSelf()
					ue, ok := self.AllocatedUETEID[msg.TEID]
					if !ok {
						handlerLog.Error("UE context not found")
						continue
					}
					relay.ForwardUPTrafficFromN3(ue, msg.Value.([]byte))
				case n3iwf_message.EventN1TunnelCPMessage:
					self := context.N3IWFSelf()
					ue, ok := self.AllocatedUEIPAddress[msg.UEInnerIP]
					if !ok {
						handlerLog.Error("UE context not found")
						continue
					}
					relay.ForwardCPTrafficFromN1(ue, msg.Value.([]byte))
				}
			}
		case <-time.After(1 * time.Second):
		}
	}
}
