package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mfmayer/goham"
	"github.com/mfmayer/gosml"
)

// SensorValue represents the information of a sensor value
type sensorValue struct {
	ObisCode         gosml.OctetString
	ValueName        string
	DeviceClass      string
	UnitOfMeasure    string
	CorrectionFactor float64
}

// sensorValueFlag is a custom flag.Value implementation to parse sensor values
type sensorValueFlag struct {
	sensorValues *[]sensorValue
}

func (sf *sensorValueFlag) String() string {
	return fmt.Sprintf("%v", sf.sensorValues)
}

func (sf *sensorValueFlag) Set(value string) error {
	fields := strings.Split(value, ",")
	correctionFactor := 1.0
	if len(fields) != 4 {
		if len(fields) != 5 {
			return fmt.Errorf("invalid sensor value format: %s", value)
		}
		var err error
		if correctionFactor, err = strconv.ParseFloat(fields[4], 64); err != nil {
			return fmt.Errorf("invalid correction factor format: %s", value)
		}
	}

	obisCodeString := fields[0]
	valueName := fields[1]
	deviceClass := fields[2]
	unitOfMeasure := fields[3]

	var obisCode []byte
	for _, val := range strings.Split(obisCodeString, ".") {
		num, err := strconv.Atoi(val)
		if err != nil || num < 0 || num > 255 {
			return fmt.Errorf("invalid Obis Code format: %s", obisCodeString)
		}
		obisCode = append(obisCode, byte(num))
	}

	sensor := sensorValue{obisCode, valueName, deviceClass, unitOfMeasure, correctionFactor}
	*sf.sensorValues = append(*sf.sensorValues, sensor)
	return nil
}

func connect(url string) (client mqtt.Client, err error) {
	opts := mqtt.NewClientOptions().AddBroker(url)
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err = token.Error()
	}
	return
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\n %s [options] \n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Reads file and publishes specified read values to given MQTT broker.\n\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	// Define flags
	filePath := flag.String("file", "", "Path to the file with sml encoded sensor values")
	mqttBrokerURL := flag.String("broker", "", "URL to the MQTT broker")
	sensorValues := []sensorValue{}
	flag.Var(&sensorValueFlag{&sensorValues}, "value", "Value's OBIS code, ValueName, DeviceClass, UnitOfMeasure and optional CorrectionFactor. Format: ObisCode,ValueName,DeviceClass,UnitOfMeasure (e.g. \"1.0.1.8.0,Energy,energy,kWh[,0.001]\"). Multiple sensor values can be specified.")

	// Parse command line arguemnts into flags
	flag.Usage = printUsage
	flag.Parse()

	// Validate flags
	if *filePath == "" || *mqttBrokerURL == "" || len(sensorValues) == 0 {
		fmt.Fprintln(os.Stderr, "Please provide the required flags: file, broker, and value(s).")
		flag.Usage()
		os.Exit(1)
	}

	// Connect to the broker
	client, err := connect(*mqttBrokerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while connecting to MQTT broker: %s\n", err)
		os.Exit(1)
	}

	// Create a publisher function that matches the MQTTPublisher interface
	mqttPublisher := goham.MQTTPublisherFunc(func(topic string, qos byte, retained bool, payload interface{}) {
		token := client.Publish(topic, qos, retained, payload)
		token.WaitTimeout(time.Millisecond * 50)
	})

	// check if argument is a valid file path
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n\n", *filePath)
		printUsage()
		os.Exit(1)
	}
	// try to open the file
	f, err := os.OpenFile(*filePath, os.O_RDONLY|256, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		os.Exit(1)
	}

	// Create MQTT Sensor
	mqttSensor := goham.NewMQTTSensor(mqttPublisher, path.Base(*filePath))
	_ = mqttSensor

	readOptions := []gosml.ReadOption{}
	for _, v := range sensorValues {
		func(value sensorValue) {
			mqttValue := mqttSensor.AddValue(
				value.ValueName,
				goham.WithDeviceClass(value.DeviceClass),
				goham.WithUnitOfMeasurement(value.UnitOfMeasure),
			)
			readOptions = append(readOptions,
				gosml.WithObisCallback(value.ObisCode, func(message *gosml.ListEntry) {
					floatValue := message.Float() * value.CorrectionFactor
					fmt.Printf("%s %f\n", message.ObjectName(), floatValue)
					mqttValue.Update(floatValue)
				}),
			)
		}(v)
	}

	// create a buffered reader for the file
	r := bufio.NewReader(f)
	// read the file using gosml module with the option
	gosml.Read(r, readOptions...)
}
