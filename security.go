package arrest

import (
	"fmt"

	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// SecurityScheme provides DSL methods for creating OpenAPI security schemes.
// The libopenapi scheme underneath is available through OpenAPISecurityScheme.
type SecurityScheme struct {
	state any // *v3.SecurityScheme; see state.go
}

//go:noinline
func SecuritySchemeForType(typ string) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: typ,
	})
}

//go:noinline
func SecuritySchemeBearerAuth() *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type:   "http",
		Scheme: "bearer",
	})
}

//go:noinline
func SecuritySchemeBearerAuthWithFormat(format string) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: format,
	})
}

//go:noinline
func SecuritySchemeBasicAuth() *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type:   "http",
		Scheme: "basic",
	})
}

//go:noinline
func SecuritySchemeAPIAuthKey(in string, name string) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "apiKey",
		Name: name,
		In:   in,
	})
}

//go:noinline
func SecuritySchemeOAuth2Implicit(
	authorizationURL string,
	scopes map[string]string,
) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "oauth2",
		Flows: &highv3.OAuthFlows{
			Implicit: &highv3.OAuthFlow{
				AuthorizationUrl: authorizationURL,
				Scopes:           orderedmap.ToOrderedMap(scopes),
			},
		},
	})
}

//go:noinline
func SecuritySchemeOAuth2AuthorizationCode(
	authorizationURL string,
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "oauth2",
		Flows: &highv3.OAuthFlows{
			AuthorizationCode: &highv3.OAuthFlow{
				AuthorizationUrl: authorizationURL,
				TokenUrl:         tokenURL,
				Scopes:           orderedmap.ToOrderedMap(scopes),
			},
		},
	})
}

//go:noinline
func SecuritySchemeOAuth2Password(
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "oauth2",
		Flows: &highv3.OAuthFlows{
			Password: &highv3.OAuthFlow{
				TokenUrl: tokenURL,
				Scopes:   orderedmap.ToOrderedMap(scopes),
			},
		},
	})
}

//go:noinline
func SecuritySchemeOAuth2ClientCredentials(
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "oauth2",
		Flows: &highv3.OAuthFlows{
			ClientCredentials: &highv3.OAuthFlow{
				TokenUrl: tokenURL,
				Scopes:   orderedmap.ToOrderedMap(scopes),
			},
		},
	})
}

//go:noinline
func SecuritySchemeCookieAuth(name string) *SecurityScheme {
	return newSecurityScheme(&highv3.SecurityScheme{
		Type: "apiKey",
		Name: name,
		In:   "cookie",
	})
}

//go:noinline
func (s *SecurityScheme) Description(description string) *SecurityScheme {
	schemeOf(s).Description = description
	return s
}

//go:noinline
func (s *SecurityScheme) Name(name string) *SecurityScheme {
	schemeOf(s).Name = name
	return s
}

//go:noinline
func (s *SecurityScheme) In(in string) *SecurityScheme {
	schemeOf(s).In = in
	return s
}

//go:noinline
func (s *SecurityScheme) Scheme(scheme string) *SecurityScheme {
	schemeOf(s).Scheme = scheme
	return s
}

//go:noinline
func (s *SecurityScheme) BearerFormat(format string) *SecurityScheme {
	schemeOf(s).BearerFormat = format
	return s
}

// RegardingFlow configures one or more OAuth flows of a security scheme. The
// flows underneath are available through OpenAPIFlows.
type RegardingFlow struct {
	securityScheme *SecurityScheme
	state          any // []*v3.OAuthFlow; see state.go
}

//go:noinline
func (s *SecurityScheme) ImplicitFlow() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	if ss.Flows.Implicit == nil {
		ss.Flows.Implicit = &highv3.OAuthFlow{}
	}

	return &RegardingFlow{
		securityScheme: s,
		state:          []*highv3.OAuthFlow{ss.Flows.Implicit},
	}
}

//go:noinline
func (s *SecurityScheme) AuthorizationCodeFlow() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	if ss.Flows.AuthorizationCode == nil {
		ss.Flows.AuthorizationCode = &highv3.OAuthFlow{}
	}

	return &RegardingFlow{
		securityScheme: s,
		state:          []*highv3.OAuthFlow{ss.Flows.AuthorizationCode},
	}
}

//go:noinline
func (s *SecurityScheme) PasswordFlow() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	if ss.Flows.Password == nil {
		ss.Flows.Password = &highv3.OAuthFlow{}
	}

	return &RegardingFlow{
		securityScheme: s,
		state:          []*highv3.OAuthFlow{ss.Flows.Password},
	}
}

//go:noinline
func (s *SecurityScheme) ClientCredentialsFlow() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	if ss.Flows.ClientCredentials == nil {
		ss.Flows.ClientCredentials = &highv3.OAuthFlow{}
	}

	return &RegardingFlow{
		securityScheme: s,
		state:          []*highv3.OAuthFlow{ss.Flows.ClientCredentials},
	}
}

//go:noinline
func (s *SecurityScheme) AllFlows() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	if ss.Flows.Implicit == nil {
		ss.Flows.Implicit = &highv3.OAuthFlow{}
	}

	if ss.Flows.AuthorizationCode == nil {
		ss.Flows.AuthorizationCode = &highv3.OAuthFlow{}
	}

	if ss.Flows.Password == nil {
		ss.Flows.Password = &highv3.OAuthFlow{}
	}

	if ss.Flows.ClientCredentials == nil {
		ss.Flows.ClientCredentials = &highv3.OAuthFlow{}
	}

	return &RegardingFlow{
		securityScheme: s,
		state: []*highv3.OAuthFlow{
			ss.Flows.Implicit,
			ss.Flows.AuthorizationCode,
			ss.Flows.Password,
			ss.Flows.ClientCredentials,
		},
	}
}

//go:noinline
func (s *SecurityScheme) AllDefinedFlows() *RegardingFlow {
	ss := schemeOf(s)
	if ss.Flows == nil {
		ss.Flows = &highv3.OAuthFlows{}
	}

	flows := make([]*highv3.OAuthFlow, 0, 4)
	if ss.Flows.Implicit != nil {
		flows = append(flows, ss.Flows.Implicit)
	}
	if ss.Flows.AuthorizationCode != nil {
		flows = append(flows, ss.Flows.AuthorizationCode)
	}
	if ss.Flows.Password != nil {
		flows = append(flows, ss.Flows.Password)
	}
	if ss.Flows.ClientCredentials != nil {
		flows = append(flows, ss.Flows.ClientCredentials)
	}

	return &RegardingFlow{
		securityScheme: s,
		state:          flows,
	}
}

//go:noinline
func (f *RegardingFlow) AddScope(name, description string) *RegardingFlow {
	for _, flow := range flowsOf(f) {
		if flow.Scopes == nil {
			flow.Scopes = orderedmap.New[string, string]()
		}
		flow.Scopes.Set(name, description)
	}
	return f
}

//go:noinline
func (f *RegardingFlow) AuthorizationURL(url string) *RegardingFlow {
	for _, flow := range flowsOf(f) {
		flow.AuthorizationUrl = url
	}
	return f
}

//go:noinline
func (f *RegardingFlow) TokenURL(url string) *RegardingFlow {
	for _, flow := range flowsOf(f) {
		flow.TokenUrl = url
	}
	return f
}

//go:noinline
func (f *RegardingFlow) RefreshURL(url string) *RegardingFlow {
	for _, flow := range flowsOf(f) {
		flow.RefreshUrl = url
	}
	return f
}

// AddExtension sets the extension name on every flow being configured. A
// *yaml.Node is used as given; any other value is encoded as YAML first.
//
//go:noinline
func (f *RegardingFlow) AddExtension(name string, value any) *RegardingFlow {
	node, ok := value.(*yaml.Node)
	if !ok {
		node = &yaml.Node{}
		if err := node.Encode(value); err != nil {
			node = &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprint(value)}
		}
	}
	for _, flow := range flowsOf(f) {
		if flow.Extensions == nil {
			flow.Extensions = orderedmap.New[string, *yaml.Node]()
		}
		flow.Extensions.Set(name, node)
	}
	return f
}
