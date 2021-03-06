// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// rawread attempts to read from the specified USB device.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/gousb"
	"log"
)

var (
	device   = flag.String("device", "03eb:2013", "Device to which to connect")
	config   = flag.Int("config", 1, "Endpoint to which to connect")
	iface    = flag.Int("interface", 0, "Endpoint to which to connect")
	setup    = flag.Int("setup", 0, "Endpoint to which to connect")
	endpoint = flag.Int("endpoint", 1, "Endpoint to which to connect")
	debug    = flag.Int("debug", 3, "Debug level for libusb")
)

func read_le_int16(data []byte) (ret int16) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := gousb.NewContext()
	defer ctx.Close()

	ctx.Debug(*debug)

	log.Printf("Scanning for device %q...", *device)

	//	// ListDevices is used to find the devices to open.
	//	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
	//		if fmt.Sprintf("%s:%s", desc.Vendor, desc.Product) != *device {
	//			return false
	//		}
	//
	//		// The usbid package can be used to print out human readable information.
	//		fmt.Printf("  Protocol: %d\n", desc)
	//
	//		// The configurations can be examined from the Descriptor, though they can only
	//		// be set once the device is opened.  All configuration references must be closed,
	//		// to free up the memory in libusb.
	//		for _, cfg := range desc.Configs {
	//			// This loop just uses more of the built-in and usbid pretty printing to list
	//			// the USB devices.
	//			fmt.Printf("  %s:\n", cfg)
	//			for _, alt := range cfg.Interfaces {
	//				fmt.Printf("    --------------\n")
	//				for _, iface := range alt.AltSettings {
	//					fmt.Printf("    %s\n", iface)
	//					for _, end := range iface.Endpoints {
	//						fmt.Printf("      %v\n", end)
	//					}
	//				}
	//			}
	//			fmt.Printf("    --------------\n")
	//		}
	//
	//		return true
	//	})
	//
	//	// All Devices returned from ListDevices must be closed.
	//	defer func() {
	//		for _, d := range devs {
	//			d.Close()
	//		}
	//	}()
	//
	//	// ListDevices can occaionally fail, so be sure to check its return value.
	//	if err != nil {
	//		log.Fatalf("list: %s", err)
	//	}
	//
	//	if len(devs) == 0 {
	//		log.Fatalf("no devices found")
	//	}
	//
	//	dev := devs[0]
	//	dev.Reset()
	//	//dev.ReadTimeout = (1 * time.Second)
	//	//dev.WriteTimeout = (1 * time.Second)
	//	//dev.ControlTimeout = (1 * time.Second)
	//
	//	//log.Printf("Using device descriptor: ")
	//	//spew.Dump(dev.Descriptor)
	//
	//	cfg, err := dev.ActiveConfig()
	//	if err != nil {
	//		log.Fatalf("Cannot get active config: %s", err)
	//	}
	//	log.Printf("Selecting config %d, endpoint 0x%x", cfg,
	//		uint8(1)|uint8(gousb.ENDPOINT_DIR_IN))
	//	ep_read, err := dev.OpenEndpoint(cfg, uint8(0),
	//		uint8(0), uint8(1)|uint8(gousb.ENDPOINT_DIR_IN))
	//	if err != nil {
	//		log.Fatalf("Failed to open read endpoint: %s", err)
	//	}
	//	log.Printf("Got read endpoint: ")
	//	spew.Dump(ep_read)
	//
	//	ep_write, err := dev.OpenEndpoint(uint8(*config), uint8(0),
	//		uint8(0), uint8(2)|uint8(gousb.ENDPOINT_DIR_OUT))
	//	if err != nil {
	//		log.Fatalf("Failed to open write endpoint: %s", err)
	//	}
	//	log.Printf("Got write endpoint: ")
	//	spew.Dump(ep_write)

	// Open any device with a given VID/PID using a convenience function.
	dev, err := ctx.OpenDeviceWithVIDPID(0x03eb, 0x2013)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	defer dev.Close()

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

	// Open an IN endpoint.
	ep_read, err := intf.InEndpoint(1)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(0): %v", intf, err)
	}

	// Open an OUT endpoint.
	ep_write, err := intf.OutEndpoint(2)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(0): %v", intf, err)
	}

	var buf []byte
	// Read invalid bytes from device
	num, err := ep_read.Read(buf)
	if err != nil {
		log.Fatal("Failed to read pending bytes into buffer")
	}
	log.Printf("Read %d bytes into temporary buffer", num)

	// request data step 1: send request command
	buf = []byte("\x40\x68\x2a\x54\x52\x0a\x40\x40\x40\x40\x40\x40\x40\x40\x40\x40")
	num, err = ep_write.Write(buf)
	if num != len(buf) {
		log.Fatalf("Failed to write request command: %s", err)
	}
	spew.Printf("Request data - wrote %d bytes: % x\n", num, buf)

	// request data step 2: read response
	num, err = ep_read.Read(buf)
	if err != nil {
		log.Fatal("Failed to read pending bytes into buffer")
	}
	spew.Printf("Response data - read %d bytes: % x\n", num, buf)
	spew.Dump(buf[2:4])
	voc := read_le_int16(buf[2:4])
	// check voc range - sensor docs says between 450 and 2000.
	// everything else is garbage.
	if (voc >= 450) && (voc <= 2000) {
		log.Printf("VOC concentration: %d ppm CO2-equivalent", voc)
	} else {
		log.Printf("ERROR: invalid VOC value %d received", voc)
	}

	// request data step 3: flush
	num, err = ep_read.Read(buf)
	if err != nil {
		log.Fatal("Failed to read pending bytes into buffer")
	}
	log.Printf("Read %d bytes into temporary buffer", num)

}
