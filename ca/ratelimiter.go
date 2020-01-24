package ca

import (
	"time"
)

var rateLimitDuration = 12 * time.Hour

// hasSignedMetadataInLastDay only returns false on
// first run 24 hours after the last signature
func (ca *CertificateAuthority) hasSignedMetadataInLastDay(instanceID uint64) bool {
	ca.rateLimiterMutex.RLock()
	defer ca.rateLimiterMutex.RUnlock()

	lastSignature, _ := ca.rateLimiter[instanceID]
	now := time.Now()
	if lastSignature == nil || now.Sub(*lastSignature) > rateLimitDuration {
		return false
	}
	return true
}

// hasSignedMetadataInLastDay only returns false on
// first run 24 hours after the last signature
func (ca *CertificateAuthority) recordSigningMetadata(instanceID uint64) {
	ca.rateLimiterMutex.Lock()
	defer ca.rateLimiterMutex.Unlock()

	now := time.Now()
	ca.rateLimiter[instanceID] = &now
}
