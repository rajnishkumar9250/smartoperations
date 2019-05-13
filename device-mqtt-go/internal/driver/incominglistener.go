// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	sdk "github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"reflect"
)

func startIncomingListening() error {
	var scheme = driver.Config.Incoming.Protocol
	var brokerUrl = driver.Config.Incoming.Host
	var brokerPort = driver.Config.Incoming.Port
	var username = driver.Config.Incoming.Username
	var password = driver.Config.Incoming.Password
	var mqttClientId = driver.Config.Incoming.MqttClientId
	var qos = byte(driver.Config.Incoming.Qos)
	var keepAlive = driver.Config.Incoming.KeepAlive
	var topic = driver.Config.Incoming.Topic

	uri := &url.URL{
		Scheme: strings.ToLower(scheme),
		Host:   fmt.Sprintf("%s:%d", brokerUrl, brokerPort),
		User:   url.UserPassword(username, password),
	}

	client, err := createClient(mqttClientId, uri, keepAlive)
	if err != nil {
		return err
	}

	defer func() {
		if client.IsConnected() {
			client.Disconnect(5000)
		}
	}()

	token := client.Subscribe(topic, qos, onIncomingDataReceived)
	if token.Wait() && token.Error() != nil {
		driver.Logger.Info(fmt.Sprintf("[Incoming listener] Stop incoming data listening. Cause:%v", token.Error()))
		return token.Error()
	}

	driver.Logger.Info("[Incoming listener] Start incoming data listening. Port is  %X and Topic is %X", brokerPort, topic)
	select {}
}

func onIncomingDataReceived(client mqtt.Client, message mqtt.Message) {
	var data map[string]interface{}
	json.Unmarshal(message.Payload(), &data)
	fmt.Printf("IncomingDataReceived is %X", data)

	fmt.Printf("checkDataWithKey")
	//if !checkDataWithKey(data, "name") || !checkDataWithKey(data, "cmd") {
	if !checkDataWithKey(data, "name") {
		return
	}

	/*
	var data_org map[string]interface{}
	data_org = data
	var cmd map[string]interface{}
	for key, value := range data {
                //err := json.Unmarshal([]byte(value.(string)), &js)
                fmt.Println(value)
                if innt, err := strconv.Atoi(value.(string)); err == nil {
                        cmd[key.(string)] = innt
                } else if flt, err := strconv.ParseFloat(value.(string), 10); err == nil {
                        cmd[key.(string)] = flt
                } else if   b, err := strconv.ParseBool(value.(string)); err == nil {
                        cmd[key.(string)] = b
                } else if  json.Unmarshal([]byte(value.(string)), &js) == nil {
                        fmt.Println("json string")
                        cmd[key.(string)] = js 
                } else {
                        fmt.Println("string")
                        cmd[key.(string)] = value.(string)
                }
        }


	*/


	deviceName := data["name"].(string)
	fmt.Println("deviceName: %X", deviceName)
	/*
	cmd := data["cmd"].(string)
	reading, ok := data[cmd]
	fmt.Println(reflect.TypeOf(reading))
	fmt.Printf("reading: %X", reading)
	if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No reading data found : topic=%v msg=%v", message.Topic(), string(message.Payload())))
		return
	}
	*/
	service := sdk.RunningService()
        


	/*
	deviceObject, ok := service.DeviceResource(deviceName, cmd, "get")
	fmt.Println(reflect.TypeOf(deviceObject))
	fmt.Printf("deviceObject: %X", deviceObject)
	if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No DeviceObject found : topic=%v msg=%v", message.Topic(), string(message.Payload())))
		return
	}

	ro, ok := service.ResourceOperation(deviceName, cmd, "get")
	fmt.Println(reflect.TypeOf(ro))
	fmt.Printf("ro: %X", ro)
	if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No ResourceOperation found : topic=%v msg=%v", message.Topic(), string(message.Payload())))
		return
	}

	result, err := newResult(deviceObject, ro, reading)

	if err != nil {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored.   topic=%v msg=%v error=%v", message.Topic(), string(message.Payload()), err))
		return
	}
	fmt.Println(reflect.TypeOf(result))
	*/	
	var res  []*sdkModel.CommandValue
	delete(data, "name");
	fmt.Printf("data: %X", data)
	for key, val := range data {
	  fmt.Println("iterating")
	  fmt.Println("key :%X", key)
	  fmt.Println(strings.ToLower(reflect.TypeOf(val).String()))
	  fmt.Println(reflect.TypeOf(val).Kind().String())
	  if reflect.TypeOf(val).Kind().String() == "map" {
	    //val = val.(string)
	    val = fmt.Sprintf("%v", val)
	  } /*else if reflect.TypeOf(val).Kind().String() == "slice"  {
	    val = strings.Join(val, ", ")
	  }*/
	  fmt.Println(val)
	  deviceObject, ok := service.DeviceResource(deviceName, key, "get")
	  fmt.Println(reflect.TypeOf(deviceObject))
	  fmt.Printf("deviceObject: %X", deviceObject)
	  if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No DeviceObject found : topic=%v msg=%v", message.Topic(), string(message.Payload())))
		return
	  }

	  ro, ok := service.ResourceOperation(deviceName, key, "get")
	  fmt.Println(reflect.TypeOf(ro))
	  fmt.Printf("ro: %X", ro)
	  if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No ResourceOperation found : topic=%v msg=%v", message.Topic(), string(message.Payload())))
		return
	  }

	  result, err := newResult(deviceObject, ro, val)

	  if err != nil {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored.   topic=%v msg=%v error=%v", message.Topic(), string(message.Payload()), err))
		return
	  }
	  fmt.Println(reflect.TypeOf(result))

	  res = append(res, result)
	}





	asyncValues := &sdkModel.AsyncValues{
		DeviceName:    deviceName,
		//CommandValues: []*sdkModel.CommandValue{result},
		CommandValues: res,
	}
	
	driver.Logger.Info(fmt.Sprintf("[Incoming listener] Incoming reading received: topic=%v msg=%v", message.Topic(), string(message.Payload())))

	driver.AsyncCh <- asyncValues
	//driver.DataCh <- data

}

func checkDataWithKey(data map[string]interface{}, key string) bool {
	val, ok := data[key]
	if !ok {
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. No %v found : msg=%v", key, data))
		return false
	}

	switch val.(type) {
	case string:
		return true
	default:
		driver.Logger.Warn(fmt.Sprintf("[Incoming listener] Incoming reading ignored. %v should be string : msg=%v", key, data))
		return false
	}
}
