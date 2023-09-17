package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"html/template"
	"log"
	"net/http"
)

func handleSummonWebpage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./src/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Buttons []string
	}{
		Buttons: selectAbleDevices(),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleUpdatePackets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan map[string][][]byte)

	clientMu.Lock()
	clients = append(clients, clientChan)
	clientMu.Unlock()

	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	go func() {
		<-closeNotifier
		removeClient(clientChan)
	}()

	for {
		select {
		case counterValue := <-clientChan:
			fmt.Fprintf(w, "data: %d\n\n", counterValue)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			removeClient(clientChan)
			return
		}
	}
}

func handleSelectDevice(device string) {
	//func handleSelectDevice(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(handle)
	//if handle != nil {
	//	handle.Close()
	//}
	//body, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	http.Error(w, "Failed to read request body", http.StatusInternalServerError)
	//	return
	//}
	//defer r.Body.Close()
	//
	//device := string("")
	//for _, value := range body {
	//	asciiChar := fmt.Sprintf("%c", value)
	//
	//	device += asciiChar
	//}
	//
	//if body == nil {
	//	device = "\\Device\\NPF_Loopback"
	//}
	//fmt.Println(test)

	fmt.Println(device)

	ipv4AddrOfInterface = getInterfaceFromDeviceName(device).Addresses[0].IP.String()
	fmt.Println(ipv4AddrOfInterface)
	handle, err := pcap.OpenLive(device, 65536, true, pcap.BlockForever)

	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	showpackets(packetSource)

	//defer handle.Close()
}

func getInterfaceFromDeviceName(device string) pcap.Interface {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, devicelist := range devices {
		if devicelist.Name == device {
			return devicelist
		}
	}
	return pcap.Interface{}
}
