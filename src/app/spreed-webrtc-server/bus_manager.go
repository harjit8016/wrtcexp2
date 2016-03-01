/*
 * Spreed WebRTC.
 * Copyright (C) 2013-2015 struktur AG
 *
 * This file is part of Spreed WebRTC.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats"
)

const (
	BusManagerOffer      = "offer"
	BusManagerAnswer     = "answer"
	BusManagerBye        = "bye"
	BusManagerConnect    = "connect"
	BusManagerDisconnect = "disconnect"
	BusManagerSession    = "session"
)

// A BusManager provides the API to interact with a bus.
type BusManager interface {
	Trigger(name, from, payload string, data interface{}) error
}

// A BusTrigger is a container to serialize trigger events
// for the bus backend.
type BusTrigger struct {
	Name    string
	From    string
	Payload string      `json:",omitempty"`
	Data    interface{} `json:",omitempty"`
}

// BusSubjectTrigger returns the bus subject for trigger payloads.
func BusSubjectTrigger(prefix, suffix string) string {
	return fmt.Sprintf("%s.%s", prefix, suffix)
}

type busManager struct {
	BusManager
}

// NewBusManager creates and initializes a new BusMager with the
// provided flags for NATS support. It is intended to connect the
// backend bus with a easy to use API to send and receive bus data.
func NewBusManager(useNats bool, subjectPrefix string) BusManager {
	var b BusManager
	if useNats {
		var err error
		b, err = newNatsBus(subjectPrefix)
		if err == nil {
			log.Println("Nats bus connected")
		} else {
			log.Println("Error connecting nats bus", err)
			b = &noopBus{}
		}
	} else {
		b = &noopBus{}
	}
	return &busManager{b}
}

type noopBus struct {
}

func (bus *noopBus) Trigger(name, from, payload string, data interface{}) error {
	return nil
}

type natsBus struct {
	prefix string
	ec     *nats.EncodedConn
}

func newNatsBus(prefix string) (*natsBus, error) {
	ec, err := EstablishNatsConnection(nil)
	if err != nil {
		return nil, err
	}
	if prefix == "" {
		prefix = "channelling.trigger"
	}
	return &natsBus{prefix, ec}, nil
}

func (bus *natsBus) Trigger(name, from, payload string, data interface{}) (err error) {
	if bus.ec != nil {
		trigger := &BusTrigger{
			Name:    name,
			From:    from,
			Payload: payload,
			Data:    data,
		}
		err = bus.ec.Publish(BusSubjectTrigger(bus.prefix, name), trigger)
		if err != nil {
			log.Println("Failed to trigger NATS event", err)
		}
	}
	return err
}