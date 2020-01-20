package ca

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// CertificateAuthority for GCP Projects
type CertificateAuthority struct {
	signer          ssh.Signer
	signInternalIPs bool
	signExternalIPs bool
	parallelism     int
	daysValid       time.Duration
	projectIDs      []string
	folderIDs       []string
	rateLimiter     map[string]time.Time
}

// Create a new CA Signer
func Create(configFile string, keyfile string, signInternalIPs bool, signExternalIPs bool, parallelism int, daysValid int) *CertificateAuthority {
	ca := CertificateAuthority{
		signInternalIPs: signInternalIPs,
		signExternalIPs: signExternalIPs,
		signer:          loadCAKey(keyfile),
		parallelism:     parallelism,
		daysValid:       time.Duration(daysValid) * 24 * time.Hour,
	}
	// time.Duration
	// Load unique project IDs from the config file
	config := loadConfigFile(configFile)
	ca.projectIDs = config.Projects
	ca.folderIDs = config.Folders

	return &ca
}

// SignKeys for any GCE instances in a whitelisted project or folder
// that have published an SSH pubkey through their Guest Attributes
func (ca *CertificateAuthority) SignKeys() {
	sem := make(chan bool, ca.parallelism)
	for _, projectID := range ca.getUniqueProjectIDs() {
		sem <- true
		go func(id string) {
			ca.signKeysInProject(id)
			<-sem
		}(projectID)
	}

	// Wait for all goroutines to finish
	for i := 0; i < len(sem); i++ {
		sem <- true
	}
}

// Merge Project IDs with all projects in the initialized Folders
// into one unique list of Project IDs
func (ca *CertificateAuthority) getUniqueProjectIDs() []string {
	allProjectIDs := []string{}
	allProjectIDs = append(allProjectIDs, ca.projectIDs...)
	allProjectIDs = append(allProjectIDs, getActiveProjectIDs(ca.folderIDs)...)
	allProjectIDs = makeSliceUnique(allProjectIDs)
	return allProjectIDs
}

// The key to query each instance's guest attributes for an SSH pubkey
const guestAttributesQueryPath = "hostkeys/ssh-ed25519"

// The instance metadata key to inject the signed key back into
const metadataInjectedSigKey = "hostkeys-signed-ssh-ed25519"

func (ca *CertificateAuthority) signKeysInProject(projectID string) {
	debugln("Processing Project: ", projectID)

	// Query all instances in all zones
	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		log.Println("[signKeysInProject]", err)
		return
	}

	aggregatedList, err := computeService.Instances.AggregatedList(projectID).Do()
	if err != nil {
		log.Println("[signKeysInProject]", err)
		return
	}

	// Search for GuestAttributes that want to be signed
	for _, list := range aggregatedList.Items {
		for _, instance := range list.Instances {
			zone := strings.Split(instance.Zone, "/")[8]
			debugln("- Processing Instance: ", projectID, zone, instance.Name)

			// Get the SSH pubkey if it exists in guest attributes
			call := computeService.Instances.GetGuestAttributes(projectID, zone, instance.Name)
			call.QueryPath(guestAttributesQueryPath)
			attributes, err := call.Do()
			if err != nil {
				debugf("  - %v not found in GuestAttributes.  Skipping instance.", guestAttributesQueryPath)
				continue
			}
			sshKey := attributes.QueryValue.Items[0].Value
			debugln("  - Key found: ", sshKey)

			// Get instance IP addresses
			ips := []string{}
			for _, nic := range instance.NetworkInterfaces {
				if ca.signExternalIPs {
					for _, ac := range nic.AccessConfigs {
						if isIPValid(ac.NatIP) {
							ips = append(ips, ac.NatIP)
						}
					}
				}
				if ca.signExternalIPs {
					if isIPValid(nic.NetworkIP) {
						ips = append(ips, nic.NetworkIP)
					}
				}
			}

			// Validate & Sign the key
			signedKey := signPubkey(ca.signer, sshKey, ips, ca.daysValid)
			log.Printf("Signed Key for Project=%v, Instance=%v, IPs=%v\n", projectID, instance.Name, ips)

			// Inject into metadata
			debugln("  - Setting Metadata")
			addSignatureToMetadata(instance.Metadata, signedKey)
			setMetaCall := computeService.Instances.SetMetadata(
				projectID,
				zone,
				instance.Name,
				instance.Metadata)
			if _, err = setMetaCall.Do(); err != nil {
				log.Println("[signKeysInProject]", err)
				return
			}
			debugln("  - Metadata Set")
		}
	}

}

// Add or update the signed SSH key in the provided instance metadata
func addSignatureToMetadata(metadata *compute.Metadata, signedKey string) {
	// Update the signed key if it already exists
	for i, item := range metadata.Items {
		if item.Key == metadataInjectedSigKey {
			metadata.Items[i].Value = &signedKey
			return
		}
	}

	// Otherwise add a new item
	item := compute.MetadataItems{
		Key:   metadataInjectedSigKey,
		Value: &signedKey,
	}
	metadata.Items = append(metadata.Items, &item)
}

// isIPValid according to IPv4
func isIPValid(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip.To4() == nil {
		log.Printf("[WARN] %v is not a valid IPv4 Address\n", ipAddress)
		return false
	}
	return true
}

// makeSliceUnique with no duplicated elements.
func makeSliceUnique(slice []string) []string {
	unique := make([]string, 0, len(slice))
	tempMap := make(map[string]bool)

	for _, val := range slice {
		if _, ok := tempMap[val]; !ok {
			tempMap[val] = true
			unique = append(unique, val)
		}
	}
	return unique
}
