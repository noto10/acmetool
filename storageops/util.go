package storageops

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/hlandau/acme/storage"
	"time"
)

type targetSorter []*storage.Target

func (ts targetSorter) Len() int {
	return len(ts)
}

func (ts targetSorter) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ts targetSorter) Less(i, j int) bool {
	return targetGt(ts[j], ts[i])
}

func targetGt(a *storage.Target, b *storage.Target) bool {
	if a == nil && b == nil {
		return false // equal
	} else if b == nil {
		return true // a > nil
	} else if a == nil {
		return false // nil < a
	}

	if a.Priority > b.Priority {
		return true
	} else if a.Priority < b.Priority {
		return false
	}

	return len(a.Satisfy.Names) > len(b.Satisfy.Names)
}

func renewTime(notBefore, notAfter time.Time) time.Time {
	validityPeriod := notAfter.Sub(notBefore)
	renewSpan := validityPeriod / 3
	if renewSpan > 30*24*time.Hour { // close enough to 30 days
		renewSpan = 30 * 24 * time.Hour
	}

	return notAfter.Add(-renewSpan)
}

func signatureAlgorithmFromKey(pk crypto.PrivateKey) (x509.SignatureAlgorithm, error) {
	switch pk.(type) {
	case *rsa.PrivateKey:
		return x509.SHA256WithRSA, nil
	case *ecdsa.PrivateKey:
		return x509.ECDSAWithSHA256, nil
	default:
		return x509.UnknownSignatureAlgorithm, fmt.Errorf("unknown key type %T", pk)
	}
}

func generateKey(trk *storage.TargetRequestKey) (pk crypto.PrivateKey, err error) {
	switch trk.Type {
	default:
		fallthrough // ...
	case "", "rsa":
		pk, err = rsa.GenerateKey(rand.Reader, clampRSAKeySize(trk.RSASize))
	case "ecdsa":
		pk, err = ecdsa.GenerateKey(getECDSACurve(trk.ECDSACurve), rand.Reader)
	}

	return
}

// Error associated with a specific target, for clarity of error messages.
type TargetSpecificError struct {
	Target *storage.Target
	Err    error
}

func (tse *TargetSpecificError) Error() string {
	return fmt.Sprintf("error satisfying %v: %v", tse.Target, tse.Err)
}
