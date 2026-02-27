package app

import (
	"context"
	"crypto/tls"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	configclientset "github.com/openshift/client-go/config/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getTLSConfigFromAPIServer fetches the TLS profile from the API Server configuration.
// This is the default source for most components.
func getTLSConfigFromAPIServer(configClient configclientset.Interface) (*tls.Config, error) {
	apiserver, err := configClient.ConfigV1().APIServers().Get(context.Background(), "cluster", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get APIServer config: %w", err)
	}

	profile := apiserver.Spec.TLSSecurityProfile
	if profile == nil {
		profile = &configv1.TLSSecurityProfile{
			Type: configv1.TLSProfileIntermediateType,
		}
	}

	return buildTLSConfigFromProfile(profile)
}

func buildTLSConfigFromProfile(profile *configv1.TLSSecurityProfile) (*tls.Config, error) {
	profileSpec, err := getTLSProfileSpec(profile)
	if err != nil {
		return nil, err
	}

	minVersion, err := parseTLSVersion(string(profileSpec.MinTLSVersion))
	if err != nil {
		return nil, fmt.Errorf("invalid MinTLSVersion: %w", err)
	}

	config := &tls.Config{
		MinVersion: minVersion,
	}

	if minVersion == tls.VersionTLS13 {
		config.MaxVersion = tls.VersionTLS13
	} else {
		cipherSuites := parseCipherSuites(profileSpec.Ciphers)
		if len(cipherSuites) == 0 {
			return nil, fmt.Errorf("no valid cipher suites found")
		}
		config.CipherSuites = cipherSuites
	}

	return config, nil
}

func getTLSProfileSpec(profile *configv1.TLSSecurityProfile) (*configv1.TLSProfileSpec, error) {
	switch profile.Type {
	case configv1.TLSProfileOldType,
		configv1.TLSProfileIntermediateType,
		configv1.TLSProfileModernType:
		return configv1.TLSProfiles[profile.Type], nil
	case configv1.TLSProfileCustomType:
		if profile.Custom == nil {
			return nil, fmt.Errorf("custom TLS profile specified but Custom field is nil")
		}
		return &profile.Custom.TLSProfileSpec, nil
	default:
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
	}
}

func parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "VersionTLS10", "TLSv1.0":
		return tls.VersionTLS10, nil
	case "VersionTLS11", "TLSv1.1":
		return tls.VersionTLS11, nil
	case "VersionTLS12", "TLSv1.2":
		return tls.VersionTLS12, nil
	case "VersionTLS13", "TLSv1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unknown TLS version: %s", version)
	}
}

// parseCipherSuites converts OpenSSL-style cipher names (as used in OpenShift TLS profiles)
// to Go's crypto/tls package constants. This mapping is based on the cipher suites defined
// in github.com/openshift/api/config/v1 TLSProfiles.
func parseCipherSuites(names []string) []uint16 {
	cipherMap := map[string]uint16{
		"ECDHE-RSA-AES128-GCM-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-ECDSA-AES128-GCM-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-RSA-AES256-GCM-SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-ECDSA-AES256-GCM-SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-RSA-CHACHA20-POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		"ECDHE-ECDSA-CHACHA20-POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		// "DHE-RSA-AES128-GCM-SHA256":     tls.TLS_DHE_RSA_WITH_AES_128_GCM_SHA256,
		// "DHE-RSA-AES256-GCM-SHA384":     tls.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-RSA-AES128-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-ECDSA-AES128-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-RSA-AES128-SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"ECDHE-ECDSA-AES128-SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		// "ECDHE-RSA-AES256-SHA384":       tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
		// "ECDHE-ECDSA-AES256-SHA384":     tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
		"ECDHE-RSA-AES256-SHA":   tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"ECDHE-ECDSA-AES256-SHA": tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"AES128-GCM-SHA256":      tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"AES256-GCM-SHA384":      tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"AES128-SHA256":          tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		//"AES256-SHA256":                 tls.TLS_RSA_WITH_AES_256_CBC_SHA256,
		"AES128-SHA":   tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"AES256-SHA":   tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"DES-CBC3-SHA": tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}

	suites := make([]uint16, 0, len(names))
	for _, name := range names {
		if suite, ok := cipherMap[name]; ok {
			suites = append(suites, suite)
		}
	}
	return suites
}
