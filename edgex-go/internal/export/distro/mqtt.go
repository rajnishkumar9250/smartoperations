//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"encoding/json"
//	"regexp"
)

type mqttSender struct {
	client MQTT.Client
	topic  string
}

// newMqttSender - create new mqtt sender
func newMqttSender(addr models.Addressable, cert string, key string) sender {
	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if protocol == "tcps" || protocol == "ssl" || protocol == "tls" {
		cert, err := tls.LoadX509KeyPair(cert, key)

		if err != nil {
			LoggingClient.Error("Failed loading x509 data")
			return nil
		}

		tlsConfig := &tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		}

		opts.SetTLSConfig(tlsConfig)

	}

	sender := &mqttSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}

	return sender
}

func (sender *mqttSender) Send(data []byte, event *models.Event) bool {
	fmt.Println("json string")
	LoggingClient.Info(fmt.Sprintf("Receive Data to Sent"))
	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mqtt server")
		if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
			LoggingClient.Error(fmt.Sprintf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error()))
			return false
		}
	}
	//var result map[string]interface{}
	//json.Unmarshal(data, &result)
	//LoggingClient.Debug(fmt.Sprintf("Sent data: %X", result))
	//fmt.Printf("readings %X", string(result['readings']))



	var jsondata map[string]interface{}
	json.Unmarshal(data, &jsondata)
	var strs1  []interface{}
	strs, _ := json.Marshal(jsondata["readings"])
	json.Unmarshal(strs, &strs1)
	var result = make(map[string]interface{})
	result["name"] = jsondata["device"]
	//var js map[string]interface{}
	//var js  interface{}
	var js map[string]interface{}
	for _, v := range strs1 {
		key := v.(map[string]interface{})["name"]
	        value := v.(map[string]interface{})["value"]
		fmt.Println(value)
		//err := json.Unmarshal([]byte(value.(string)), &js)
		//fmt.Println(err.Error())
		/*if innt, err := strconv.Atoi(value.(string)); err == nil {
			result[key.(string)] = innt
		} else if flt, err := strconv.ParseFloat(value.(string), 10); err == nil {
			result[key.(string)] = flt
		} else if   b, err := strconv.ParseBool(value.(string)); err == nil {
			result[key.(string)] = b
		} else*/ if  json.Unmarshal([]byte(value.(string)), &js) == nil {
			fmt.Println("json string")
			//result[key.(string)] = js 
			js["name"] =  jsondata["device"]
			result = js
                } else {
			fmt.Println("string")
			result[key.(string)] = value.(string)
                }
	}

	empData, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	jsonStr := string(empData)
	token := sender.client.Publish(sender.topic, 0, false, jsonStr)
	LoggingClient.Info(fmt.Sprintf("Data Sent"))

	//token := sender.client.Publish(sender.topic, 0, false, data)
	// FIXME: could be removed? set of tokens?
	token.Wait()
	if token.Error() != nil {
		LoggingClient.Error(token.Error().Error())
		return false
	} else {
		LoggingClient.Debug(fmt.Sprintf("Sent data: %X", data))
		return true
	}
}
