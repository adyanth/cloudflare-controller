package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkingv1alpha1 "github.com/adyanth/cloudflare-operator/api/v1alpha1"
	"github.com/go-logr/logr"
)

const (
	// Either tunnel or clustertunnel is mandatory
	// Tunnel CR Name
	tunnelAnnotation = "cfargotunnel.com/tunnel"
	// ClusterTunnel CR Name
	clusterTunnelAnnotation = "cfargotunnel.com/cluster-tunnel"
	// FQDN to create a DNS entry for and route traffic from internet on, defaults to Service name + cloudflare domain
	fqdnAnnotation = "cfargotunnel.com/fqdn"
	// Target can be used to override the target to send traffic to. Ex: Can be used to point to an ingress rather than the service directly
	targetAnnotation = "cfargotunnel.com/target"
	// Setting this annotation skips TLS verification for this ingress. Content does not matter. Delete the annotation if not desired. https://github.com/cloudflare/cloudflared/issues/585
	noTlsVerifyAnnotation = "cfargotunnel.com/noTlsVerify"
	// Name of the key containing the origin CA certificate in the Secret mentioned under the Tunnel.spec.originCaPool
	caPoolAnnotation = "cfargotunnel.com/caPool"

	// Protocol to use between cloudflared and the Service.
	tunnelProtoAnnotation = "cfargotunnel.com/proto"
	tunnelProtoHTTP       = "http"
	tunnelProtoHTTPS      = "https"
	tunnelProtoTCP        = "tcp"
	tunnelProtoUDP        = "udp"
	tunnelProtoSSH        = "ssh"
	tunnelProtoRDP        = "rdp"

	// Checksum of the config, used to restart pods in the deployment
	tunnelConfigChecksum = "cfargotunnel.com/checksum"

	// Tunnel properties labels
	tunnelLabel          = "cfargotunnel.com/tunnel"
	clusterTunnelLabel   = "cfargotunnel.com/cluster-tunnel"
	isClusterTunnelLabel = "cfargotunnel.com/is-cluster-tunnel"
	tunnelIdLabel        = "cfargotunnel.com/id"
	tunnelNameLabel      = "cfargotunnel.com/name"
	tunnelAppLabel       = "cfargotunnel.com/app"
	tunnelDomainLabel    = "cfargotunnel.com/domain"
	tunnelFinalizer      = "cfargotunnel.com/finalizer"
	// Service labels
	configHostnameLabel     = "cfargotunnel.com/hostname"
	configServiceLabel      = "cfargotunnel.com/service"
	configServiceLabelSplit = "."
	configmapKey            = "config.yaml"
)

var tunnelValidProtoMap map[string]bool = map[string]bool{
	tunnelProtoHTTP:  true,
	tunnelProtoHTTPS: true,
	tunnelProtoTCP:   true,
	tunnelProtoUDP:   true,
	tunnelProtoSSH:   true,
	tunnelProtoRDP:   true,
}

func getAPIDetails(ctx context.Context, c client.Client, log logr.Logger, tunnelSpec networkingv1alpha1.TunnelSpec, tunnelStatus networkingv1alpha1.TunnelStatus, namespace string) (*CloudflareAPI, *corev1.Secret, error) {

	// Get secret containing API token
	cfSecret := &corev1.Secret{}
	if err := c.Get(ctx, apitypes.NamespacedName{Name: tunnelSpec.Cloudflare.Secret, Namespace: namespace}, cfSecret); err != nil {
		log.Error(err, "secret not found", "secret", tunnelSpec.Cloudflare.Secret)
		return &CloudflareAPI{}, &corev1.Secret{}, err
	}

	// Read secret for API Token
	cfAPITokenB64, ok := cfSecret.Data[tunnelSpec.Cloudflare.CLOUDFLARE_API_TOKEN]
	if !ok {
		log.Info("key not found in secret", "secret", tunnelSpec.Cloudflare.Secret, "key", tunnelSpec.Cloudflare.CLOUDFLARE_API_TOKEN)
	}

	// Read secret for API Key
	cfAPIKeyB64, ok := cfSecret.Data[tunnelSpec.Cloudflare.CLOUDFLARE_API_KEY]
	if !ok {
		log.Info("key not found in secret", "secret", tunnelSpec.Cloudflare.Secret, "key", tunnelSpec.Cloudflare.CLOUDFLARE_API_KEY)
	}

	cfAPI := &CloudflareAPI{
		Log:             log,
		AccountName:     tunnelSpec.Cloudflare.AccountName,
		AccountId:       tunnelSpec.Cloudflare.AccountId,
		Domain:          tunnelSpec.Cloudflare.Domain,
		APIToken:        string(cfAPITokenB64),
		APIKey:          string(cfAPIKeyB64),
		APIEmail:        tunnelSpec.Cloudflare.Email,
		ValidAccountId:  tunnelStatus.AccountId,
		ValidTunnelId:   tunnelStatus.TunnelId,
		ValidTunnelName: tunnelStatus.TunnelName,
		ValidZoneId:     tunnelStatus.ZoneId,
	}
	return cfAPI, cfSecret, nil
}
