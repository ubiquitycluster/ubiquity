// Ubiquity PXE Installer
//
// A Go-based PXE boot server that:
// 1. Listens for PXE boot requests using the pixiecore library
// 2. Serves kernel/initrd to machines on the network
// 3. Provides a phone-home HTTP API that machines call after installation
// 4. Triggers Ansible playbooks for further provisioning
//
// Usage:
//   ubiquity-installer --kernel ./vmlinuz --initrd ./initrd.img
//
// Environment variables:
//   UBQUITY_INSTALLER_PORT  — phone-home API port (default: 8080)
//   UBQUITY_INSTALLER_ADDR  — bind address (default: 0.0.0.0)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"go.universe.tf/netboot/pixiecore"
)

// hostMap maps MAC addresses to hostnames.
// Each host that PXE-boots is identified by MAC and then
// an Ansible playbook is triggered for that specific host.
var hostMap = map[string]string{
	// Example:
	// "bc:24:11:d0:28:34": "node1",
	// "bc:24:11:0d:2f:20": "node2",
}

// PhoneHomePayload is sent by machines after they finish installing.
type PhoneHomePayload struct {
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
	Host string `json:"host,omitempty"`
}

var (
	installed = make(map[string]bool)
	inFlight  = make(map[string]bool)
	mu        sync.Mutex

	kernelPath = flag.String("kernel", "", "Path to kernel bzImage")
	initrdPath = flag.String("initrd", "", "Path to initrd")
	address    = flag.String("address", "0.0.0.0", "Address to listen on")
	apiPort    = flag.Int("api-port", 8080, "Phone-home API port")
)

func phoneHomeHandler(w http.ResponseWriter, r *http.Request) {
	var p PhoneHomePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Phone home from %s (%s)", p.MAC, p.IP)

	mu.Lock()
	installed[p.MAC] = true
	host, hasAnsible := hostMap[p.MAC]
	delete(inFlight, p.MAC)
	mu.Unlock()

	if hasAnsible {
		go runAnsible(host, p.IP)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"installed": installed,
		"in_flight": inFlight,
		"hosts":     hostMap,
	})
}

func runAnsible(host, ip string) {
	log.Printf("Triggering Ansible for host %s (%s)", host, ip)
	cmd := exec.Command("ansible-playbook",
		"-i", fmt.Sprintf("%s,", ip),
		"metal/boot.yml",
		"--limit", host,
		"--extra-vars", fmt.Sprintf("ansible_host=%s", ip),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Ansible failed for %s: %v", host, err)
	}
}

func main() {
	flag.Parse()

	// Override from env
	if v := os.Getenv("UBQUITY_INSTALLER_ADDR"); v != "" {
		*address = v
	}
	if v := os.Getenv("UBQUITY_INSTALLER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", apiPort)
	}

	if *kernelPath == "" || *initrdPath == "" {
		log.Print("WARNING: --kernel and --initrd are not set.")
		log.Print("The PXE server will start but no boot images will be served.")
		log.Print("Pass --kernel /path/to/vmlinuz --initrd /path/to/initrd.img")
	}

	// Build a static booter from the provided kernel/initrd paths
	spec := &pixiecore.Spec{
		Kernel:  pixiecore.ID(*kernelPath),
		Initrd:  []pixiecore.ID{pixiecore.ID(*initrdPath)},
		Cmdline: "console=ttyS0,115200n8 console=tty0 net.ifnames=0",
	}

	booter, err := pixiecore.StaticBooter(spec)
	if err != nil {
		log.Fatalf("Failed to create booter: %v", err)
	}

	// Start PXE server
	pxe := &pixiecore.Server{
		Booter:     booter,
		Address:    *address,
		DHCPNoBind: true, // DHCP proxy mode — coexists with existing DHCP
		Log: func(subsystem, msg string) {
			log.Printf("[%s] %s", subsystem, msg)
		},
	}

	go func() {
		log.Printf("PXE server listening on %s (DHCP proxy mode)", *address)
		if err := pxe.Serve(); err != nil {
			log.Fatalf("PXE server error: %v", err)
		}
	}()

	// Phone-home API
	mux := http.NewServeMux()
	mux.HandleFunc("/phone-home", phoneHomeHandler)
	mux.HandleFunc("/status", statusHandler)

	apiAddr := fmt.Sprintf("%s:%d", *address, *apiPort)
	log.Printf("Phone-home API listening on %s", apiAddr)

	go func() {
		if err := http.ListenAndServe(apiAddr, mux); err != nil {
			log.Fatalf("API server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	pxe.Shutdown()
}