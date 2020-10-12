// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers"
)

const (
	// ContentType represents JSON format content type.
	ContentType = "application/json"
)

var (
	errDecode = errors.New("failed to decode json")
)

type transformer struct {
	keys []string
}

// New returns transformer service implementation for SenML messages.
func New(keys []string) transformers.Transformer {
	return transformer{
		keys: keys,
	}
}

func (t transformer) Transform(mm messaging.Message) (interface{}, error) {
	var payload map[string]interface{}
	err := json.Unmarshal(mm.Payload, &payload)
	if err != nil {
		return []transformers.Message{}, errors.Wrap(errDecode, err)
	}

	var msgs []transformers.Message

	for k, v := range payload {
		// Apply key filter
		for _, tk := range t.keys {
			if tk == "*" || tk == k {
				kvmgs := createKeyValMsg(mm, k, v)
				msgs = append(msgs, kvmgs...)
				continue
			}
		}
	}

	return msgs, nil
}

func createKeyValMsg(mm messaging.Message, key string, val interface{}) []transformers.Message {
	msg := transformers.Message{
		Channel:   mm.Channel,
		Subtopic:  mm.Subtopic,
		Publisher: mm.Publisher,
		Protocol:  mm.Protocol,
		Name:      key,
		Time:      float64(mm.Created) / float64(1e9),
	}

	switch val.(type) {
	case string:
		s := val.(string)
		msg.StringValue = &s
	case float64:
		f := val.(float64)
		msg.Value = &f
	case bool:
		b := val.(bool)
		msg.BoolValue = &b
	case []byte:
		d := string(val.([]byte))
		msg.DataValue = &d
	case map[string]interface{}:
		var msgs []transformers.Message

		for k, v := range val.(map[string]interface{}) {
			nKey := key + "/" + k
			kvmgs := createKeyValMsg(mm, nKey, v)
			msgs = append(msgs, kvmgs...)
		}
		return msgs
	}

	return []transformers.Message{msg}
}
