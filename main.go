package main

import (
	"flag"
	"log"
	"time"

	"gitlab.com/unit410/gcp-ssh-ca/ca"
)

var (
	caKeyFile       = flag.String("ca-keyfile", "./ca_key", "The PEM Formatted SSH CA Key")
	configFile      = flag.String("config-file", "./config.yaml", "Yaml formatted config file")
	debug           = flag.Bool("debug", false, "Enable debug mode")
	oneshot         = flag.Bool("oneshot", false, "Run once then quit")
	signInternalIPs = flag.Bool("sign-internal-ips", true, "Sign the internal IPs of instances")
	signExternalIPs = flag.Bool("sign-external-ips", true, "Sign the external IPs of instances")
	parallelism     = flag.Int("parallelism", 3, "Max number of simultaneous GCP Signers to run.")
	simulate        = flag.Bool("simulate", false, "Don't inject any signatures, just simulate a run")
	validity        = flag.Int("validity", 365, "Number of days that this cert signature is valid.")
)

func main() {
	log.Println("GCS SSH Certificate Authority starting up...")

	flag.Parse()
	ca.DebugEnabled = *debug

	// Run loop
	certAuthority := ca.Create(*configFile, *caKeyFile, *signInternalIPs, *signExternalIPs, *parallelism, *simulate, *validity)
	for {
		certAuthority.SignKeys()
		log.Println("Ran successfully")
		if *oneshot {
			break
		}
		time.Sleep(10 * time.Second)
	}
}
