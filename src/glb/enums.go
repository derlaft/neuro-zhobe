package glb

import (
	"fmt"
)

type (
	ConnectionError     uint
	AuthenticationError uint
	PresenceType        uint
	Affiliation         uint
	Role                uint

	DisconnectError struct {
		ConnectionError     ConnectionError
		AuthenticationError AuthenticationError
	}
)

var (
	ConnectionErrors = []string{
		"NoError",
		"StreamError",
		"StreamVersionError",
		"StreamClosed",
		"ProxyAuthRequired",
		"ProxyAuthFailed",
		"ProxyNoSupportedAuth",
		"IoError",
		"ParseError",
		"ConnectionRefused",
		"DnsError",
		"OutOfMemory",
		"NoSupportedAuth",
		"TlsFailed",
		"TlsNotAvailable",
		"CompressionFailed",
		"AuthenticationFailed",
		"UserDisconnected",
		"NotConnected",
	}

	AuthenticationErrors = []string{
		"ErrorUndefined",
		"SaslAborted",
		"SaslIncorrectEncoding",
		"SaslInvalidAuthzid",
		"SaslInvalidMechanism",
		"SaslMalformedRequest",
		"SaslMechanismTooWeak",
		"SaslNotAuthorized",
		"SaslTemporaryAuthFailure",
		"NonSaslConflict",
		"NonSaslNotAcceptable",
		"NonSaslNotAuthorized",
	}
)

const (
	// Connection errors
	ConnErrNoError = ConnectionError(iota)
	ConnErrStreamError
	ConnErrStreamVersionError
	ConnErrStreamClosed
	ConnErrProxyAuthRequired
	ConnErrProxyAuthFailed
	ConnErrProxyNoSupportedAuth
	ConnErrIoError
	ConnErrParseError
	ConnErrConnectionRefused
	ConnErrDnsError
	ConnErrOutOfMemory
	ConnErrNoSupportedAuth
	ConnErrTlsFailed
	ConnErrTlsNotAvailable
	ConnErrCompressionFailed
	ConnErrAuthenticationFailed
	ConnErrUserDisconnected
	ConnErrNotConnected

	// Auth Errors
	AuthErrUndefined = AuthenticationError(iota)
	AuthErrSaslAborted
	AuthErrSaslIncorrectEncoding
	AuthErrSaslInvalidAuthzid
	AuthErrSaslInvalidMechanism
	AuthErrSaslMalformedRequest
	AuthErrSaslMechanismTooWeak
	AuthErrSaslNotAuthorized
	AuthErrSaslTemporaryAuthFailure
	AuthErrNonSaslConflict
	AuthErrNonSaslNotAcceptable
	AuthErrNonSaslNotAuthorized

	// Presence Types
	PresenceAvailable = PresenceType(iota)
	PresenceChat
	PresenceAway
	PresenceDND
	PresenceXA
	PresenceUnavailable
	PresenceProbe
	PresenceError
	PresenceInvalid

	// Affiliations
	AffiliationNone = Affiliation(iota)
	AffiliationOutcast
	AffiliationMember
	AffiliationOwner
	AffiliationAdmin
	AffiliationInvalid

	// Roles
	RoleNone = Role(iota)
	RoleVisitor
	RoleParticipant
	RoleModerator
	RoleInvalid
)

func (d DisconnectError) Error() string {
	return fmt.Sprintf(
		"Dissonnected with error (errCode=%v, authError=%v)",
		ConnectionErrors[d.ConnectionError],
		AuthenticationErrors[d.AuthenticationError],
	)
}
