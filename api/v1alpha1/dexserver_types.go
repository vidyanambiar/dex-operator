/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StaticPasswordSpec allows us to define login credentials. Do not expect us to use this.
type StaticPasswordSpec struct {
	Email string `json:"email,omitempty"`
}

// StorageSpec defines how/if we persist the configuration to a database on store in K8s.
type StorageSpec struct {
	Type string `json:"type,omitempty"`
}

// WebSpec defines override for cert to dex server.
type WebSpec struct {
	Http    string `json:"http,omitempty"`
	Https   string `json:"https,omitempty"`
	TlsCert string `json:"tlsCert,omitempty"`
	TlsKey  string `json:"tlsKey,omitempty"`
}

// GrpcSpec defines override options on how we run grpc server. Addr should not need to change. The certs are required.
type GrpcSpec struct {
	Addr        string `json:"addr,omitempty"`
	TlsCert     string `json:"tlsCert,omitempty"`
	TlsKey      string `json:"tlsKey,omitempty"`
	TlsClientCA string `json:"tlsClientCA,omitempty"`
}

// ExpirySpec defines how we expire
type ExpirySpec struct {
	DeviceRequests string `json:"deviceRequests,omitempty"`
}

// LoggerSpec defines loggingoptions. Optional
type LoggerSpec struct {
	Level  string `json:"level,omitempty"`
	Format string `json:"format,omitempty"`
}

// Oauth2Spec defines dex behavior flags
type Oauth2Spec struct {
	ResponseTypes         []string `json:"responseTypes,omitempty"`
	SkipApprovalScreen    bool     `json:"skipApprovalScreen,omitempty"`
	AlwaysShowLoginScreen bool     `json:"alwaysShowLoginScreen,omitempty"`
	PasswordConnector     string   `json:"passwordConnector,omitempty"`
}

// LDAP UserMatcher holds information about user and group matching
type UserMatcher struct {
	UserAttr  string `json:"userAttr"`
	GroupAttr string `json:"groupAttr"`
}

// LDAP User entry search configuration
type UserSearchSpec struct {
	// BaseDN to start the search from. For example "cn=users,dc=example,dc=com"
	BaseDN string `json:"baseDN,omitempty"`

	// Optional filter to apply when searching the directory. For example "(objectClass=person)"
	Filter string `json:"filter,omitempty"`

	// Attribute to match against the inputted username. This will be translated and combined
	// with the other filter as "(<attr>=<username>)".
	Username string `json:"username,omitempty"`

	// Can either be:
	// * "sub" - search the whole sub tree
	// * "one" - only search one level
	Scope string `json:"scope,omitempty"`

	// A mapping of attributes on the user entry to claims.
	IDAttr                    string `json:"idAttr,omitempty"`                // Defaults to "uid"
	EmailAttr                 string `json:"emailAttr,omitempty"`             // Defaults to "mail"
	NameAttr                  string `json:"nameAttr,omitempty"`              // No default.
	PreferredUsernameAttrAttr string `json:"preferredUsernameAttr,omitempty"` // No default.

	// If this is set, the email claim of the id token will be constructed from the idAttr and
	// value of emailSuffix. This should not include the @ character.
	EmailSuffix string `json:"emailSuffix,omitempty"` // No default.
}

// LDAP Group search configuration
type GroupSearchSpec struct {
	// BaseDN to start the search from. For example "cn=groups,dc=example,dc=com"
	BaseDN string `json:"baseDN,omitempty"`

	// Optional filter to apply when searching the directory. For example "(objectClass=posixGroup)"
	Filter string `json:"filter,omitempty"`

	Scope string `json:"scope,omitempty"` // Defaults to "sub"

	// DEPRECATED config options. Those are left for backward compatibility.
	// See "UserMatchers" below for the current group to user matching implementation
	// TODO: should be eventually removed from the code
	UserAttr  string `json:"userAttr,omitempty"`
	GroupAttr string `json:"groupAttr,omitempty"`

	// Array of the field pairs used to match a user to a group.
	// See the "UserMatcher" struct for the exact field names
	//
	// Each pair adds an additional requirement to the filter that an attribute in the group
	// match the user's attribute value. For example that the "members" attribute of
	// a group matches the "uid" of the user. The exact filter being added is:
	//
	//   (userMatchers[n].<groupAttr>=userMatchers[n].<userAttr value>)
	//
	UserMatchers []UserMatcher `json:"userMatchers,omitempty"`

	// The attribute of the group that represents its name.
	NameAttr string `json:"nameAttr,omitempty"`
}

// GitHubConfigSpec describes the configuration specific to the GitHub connector
type GitHubConfigSpec struct {
	ClientID        string                 `json:"clientID,omitempty"`
	ClientSecretRef corev1.ObjectReference `json:"clientSecretRef,omitempty"`
	// TODO: confirm if we set this, or allow this to be passed in?
	RedirectURI string `json:"redirectURI,omitempty"`
	Org         string `json:"org,omitempty"`
}

// LDAPConfigSpec describes the configuration specific to the LDAP connector
type LDAPConfigSpec struct {
	// The following fields are associated with connector type 'ldap'
	// The host and optional port of the LDAP server. If port isn't supplied, it will be guessed based on the TLS configuration. 389 or 636.
	Host string `json:"host"`
	// insecureNoSSL is required if the LDAP host is not using TLS (port 389). Because this option inherently leaks passwords to anyone on the same network as dex, this option may be removed in a future version of dex
	InsecureNoSSL bool `json:"insecureNoSSL,omitempty"`
	// If a custom certificate isn't provided, insecureSkipVerify can be used to turn on TLS certificate checks. However, it is insecure and shouldn't be used outside of explorative phases
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// The startTLS option indicates that when connecting to the server, connect using the ldap:// protocol then issue a StartTLS command. If unspecified, connections will use the ldaps:// protocol
	StartTLS bool `json:"startTLS,omitempty"`
	// Path to a trusted root certificate file. Default: use the host's root CA
	RootCA string `json:"rootCA,omitempty"`
	// A raw certificate file can also be provided inline as a base64 encoded PEM file.
	RootCAData []byte `json:"rootCAData,omitempty"`
	// The DN for an application service account. The connector uses the bindDN and bindPW as credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
	BindDN string `json:"bindDN"`
	// The password for an application service account. The connector uses the bindDN and bindPW as credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
	// Please note that if the bind password contains a `$`, it has to be saved in an environment variable which should be given as the value to `bindPW`.
	BindPWRef corev1.SecretReference `json:"bindPWRef"`
	// The attribute to display in the provided password prompt. If unset, will display "Username"
	UsernamePrompt string `json:"usernamePrompt,omitempty"`

	UserSearch UserSearchSpec `json:"userSearch,omitempty"`

	GroupSearch GroupSearchSpec `json:"groupSearch,omitempty"`
}

// ConfigSpec describes the configuration properties associated with a connector
type ConfigSpec struct {
	GitHub GitHubConfigSpec `json:"github,omitempty"`

	LDAP LDAPConfigSpec `json:"ldap,omitempty"`
}

// ConnectorSpec defines the OIDC connector config details
type ConnectorSpec struct {
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Enum=github;ldap
	Type   ConnectorType `json:"type,omitempty"`
	Id     string        `json:"id,omitempty"`
	Config ConfigSpec    `json:",inline"`
}

type ConnectorType string

const (
	// ConnectorTypeGitHub enables Dex to use the GitHub OAuth2 flow to identify the end user through their GitHub account
	ConnectorTypeGitHub ConnectorType = "github"

	// ConnectorTypeLDAP enables Dex to allow email/password based authentication, backed by an LDAP directory
	ConnectorTypeLDAP ConnectorType = "ldap"
)

// DexServerSpec defines the desired state of DexServer
type DexServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Foo is an example field of DexServer. Edit dexserver_types.go to remove/update
	Foo string `json:"foo,omitempty"`
	// TODO: Issuer references the dex instance web URI. Should this be returned as status?
	Issuer           string               `json:"issuer,omitempty"`
	EnablePasswordDB bool                 `json:"enablepassworddb,omitempty"`
	StaticPasswords  []StaticPasswordSpec `json:"staticpasswords,omitempty"`
	Storage          StorageSpec          `json:"storage,omitempty"`
	Web              WebSpec              `json:"web,omitempty"`
	Grpc             GrpcSpec             `json:"grpc,omitempty"`
	Expiry           ExpirySpec           `json:"expiry,omitempty"`
	Logger           LoggerSpec           `json:"logger,omitempty"`
	Oauth2           Oauth2Spec           `json:"oauth2,omitempty"`
	Connectors       []ConnectorSpec      `json:"connectors,omitempty"`
}

// DexServerStatus defines the observed state of DexServer
type DexServerStatus struct {
	// +optional
	State string `json:"state,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DexServer is the Schema for the dexservers API
type DexServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DexServerSpec   `json:"spec,omitempty"`
	Status DexServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DexServerList contains a list of DexServer
type DexServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DexServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DexServer{}, &DexServerList{})
}
