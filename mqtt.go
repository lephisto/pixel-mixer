package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"encoding/hex"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func startMqtt(config *Config, mixerOutputEnabler chan bool, mixerInputSelector chan int, rgbInputColor chan *Color) {
	connectionMadeHandler := func(client mqtt.Client) {
		if token := client.Subscribe(config.Mqtt.Topics.Power, 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error 2")
			fmt.Println(token.Error())
			os.Exit(1)
		}

		if token := client.Subscribe(config.Mqtt.Topics.Input, 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error 3")
			fmt.Println(token.Error())
			os.Exit(1)
		}

		if token := client.Subscribe(config.Mqtt.Topics.Color, 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println("Error 4")
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		message := string(msg.Payload())
		fmt.Printf("%s: %s\n", msg.Topic(), message)

		switch msg.Topic() {
		case config.Mqtt.Topics.Power:
			if message == "ON" {
				mixerOutputEnabler <- true
			} else {
				mixerOutputEnabler <- false
			}
		case config.Mqtt.Topics.Input:
			for i := 0; i < len(config.Inputs); i++ {
				if config.Inputs[i].MqttMessage == message {
					mixerInputSelector <- i + 1 // add 1 because input 0 is always solid RGB
				}
			}
		case config.Mqtt.Topics.Color:
			c, err := parseColor(message)
			if err != nil {
				fmt.Printf("Parse Error: %s\n", *err)
				return
			}
			mixerInputSelector <- rgbInput
			rgbInputColor <- c
		}
	}

	connectionLostHandler := func(client mqtt.Client, err error) {
		fmt.Println("Connection was lost, here is the error")
		fmt.Println(err)
	}

	mqtt.ERROR = log.New(os.Stdout, "ERROR ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "CRITICAL ", 0)

	opts := mqtt.NewClientOptions().
		SetKeepAlive(2 * time.Second).
		AddBroker(config.Mqtt.Server).
		SetOnConnectHandler(connectionMadeHandler).
		SetDefaultPublishHandler(messageHandler).
		SetConnectionLostHandler(connectionLostHandler).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(time.Minute)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Error 1")
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func parseColor(message string) (*Color, *string) {

        r,err := hex.DecodeString(message[0:2])
        g,err := hex.DecodeString(message[2:4])
        b,err := hex.DecodeString(message[4:6])

        if err != nil {
                errorMessage := "Invalid Hex"
                return nil, &errorMessage
        }

	c := Color{uint8(r[0]), uint8(g[0]), uint8(b[0])}

	return &c, nil
}
