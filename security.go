package arrest

import (
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

type SecurityScheme struct {
	SecurityScheme *highv3.SecurityScheme
}

func SecuritySchemeForType(typ string) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: typ,
		},
	}
}

func SecuritySchemeBearerAuth() *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type:   "http",
			Scheme: "bearer",
		},
	}
}

func SecuritySchemeBearerAuthWithFormat(format string) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: format,
		},
	}
}

func SecuritySchemeBasicAuth() *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		},
	}
}

func SecuritySchemeAPIAuthKey(in string, name string) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "apiKey",
			Name: name,
			In:   in,
		},
	}
}

func SecuritySchemeOAuth2Implicit(
	authorizationURL string,
	scopes map[string]string,
) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "oauth2",
			Flows: &highv3.OAuthFlows{
				Implicit: &highv3.OAuthFlow{
					AuthorizationUrl: authorizationURL,
					Scopes:           orderedmap.ToOrderedMap(scopes),
				},
			},
		},
	}
}

func SecuritySchemeOAuth2AuthorizationCode(
	authorizationURL string,
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "oauth2",
			Flows: &highv3.OAuthFlows{
				AuthorizationCode: &highv3.OAuthFlow{
					AuthorizationUrl: authorizationURL,
					TokenUrl:         tokenURL,
					Scopes:           orderedmap.ToOrderedMap(scopes),
				},
			},
		},
	}
}

func SecuritySchemeOAuth2Password(
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "oauth2",
			Flows: &highv3.OAuthFlows{
				Password: &highv3.OAuthFlow{
					TokenUrl: tokenURL,
					Scopes:   orderedmap.ToOrderedMap(scopes),
				},
			},
		},
	}
}

func SecuritySchemeOAuth2ClientCredentials(
	tokenURL string,
	scopes map[string]string,
) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "oauth2",
			Flows: &highv3.OAuthFlows{
				ClientCredentials: &highv3.OAuthFlow{
					TokenUrl: tokenURL,
					Scopes:   orderedmap.ToOrderedMap(scopes),
				},
			},
		},
	}
}

func SecuritySchemeCookieAuth(name string) *SecurityScheme {
	return &SecurityScheme{
		SecurityScheme: &highv3.SecurityScheme{
			Type: "apiKey",
			Name: name,
			In:   "cookie",
		},
	}
}

func (s *SecurityScheme) Description(description string) *SecurityScheme {
	s.SecurityScheme.Description = description
	return s
}

func (s *SecurityScheme) Name(name string) *SecurityScheme {
	s.SecurityScheme.Name = name
	return s
}

func (s *SecurityScheme) In(in string) *SecurityScheme {
	s.SecurityScheme.In = in
	return s
}

func (s *SecurityScheme) Scheme(scheme string) *SecurityScheme {
	s.SecurityScheme.Scheme = scheme
	return s
}

func (s *SecurityScheme) BearerFormat(format string) *SecurityScheme {
	s.SecurityScheme.BearerFormat = format
	return s
}

type regardingFlow struct {
	securityScheme *SecurityScheme
	flow           []*highv3.OAuthFlow
}

func (s *SecurityScheme) ImplicitFlow() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	if s.SecurityScheme.Flows.Implicit == nil {
		s.SecurityScheme.Flows.Implicit = &highv3.OAuthFlow{}
	}

	return &regardingFlow{
		securityScheme: s,
		flow:           []*highv3.OAuthFlow{s.SecurityScheme.Flows.Implicit},
	}
}

func (s *SecurityScheme) AuthorizationCodeFlow() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	if s.SecurityScheme.Flows.AuthorizationCode == nil {
		s.SecurityScheme.Flows.AuthorizationCode = &highv3.OAuthFlow{}
	}

	return &regardingFlow{
		securityScheme: s,
		flow:           []*highv3.OAuthFlow{s.SecurityScheme.Flows.AuthorizationCode},
	}
}

func (s *SecurityScheme) PasswordFlow() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	if s.SecurityScheme.Flows.Password == nil {
		s.SecurityScheme.Flows.Password = &highv3.OAuthFlow{}
	}

	return &regardingFlow{
		securityScheme: s,
		flow:           []*highv3.OAuthFlow{s.SecurityScheme.Flows.Password},
	}
}

func (s *SecurityScheme) ClientCredentialsFlow() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	if s.SecurityScheme.Flows.ClientCredentials == nil {
		s.SecurityScheme.Flows.ClientCredentials = &highv3.OAuthFlow{}
	}

	return &regardingFlow{
		securityScheme: s,
		flow:           []*highv3.OAuthFlow{s.SecurityScheme.Flows.ClientCredentials},
	}
}

func (s *SecurityScheme) AllFlows() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	if s.SecurityScheme.Flows.Implicit == nil {
		s.SecurityScheme.Flows.Implicit = &highv3.OAuthFlow{}
	}

	if s.SecurityScheme.Flows.AuthorizationCode == nil {
		s.SecurityScheme.Flows.AuthorizationCode = &highv3.OAuthFlow{}
	}

	if s.SecurityScheme.Flows.Password == nil {
		s.SecurityScheme.Flows.Password = &highv3.OAuthFlow{}
	}

	if s.SecurityScheme.Flows.ClientCredentials == nil {
		s.SecurityScheme.Flows.ClientCredentials = &highv3.OAuthFlow{}
	}

	return &regardingFlow{
		securityScheme: s,
		flow: []*highv3.OAuthFlow{
			s.SecurityScheme.Flows.Implicit,
			s.SecurityScheme.Flows.AuthorizationCode,
			s.SecurityScheme.Flows.Password,
			s.SecurityScheme.Flows.ClientCredentials,
		},
	}
}

func (s *SecurityScheme) AllDefinedFlows() *regardingFlow {
	if s.SecurityScheme.Flows == nil {
		s.SecurityScheme.Flows = &highv3.OAuthFlows{}
	}

	flows := make([]*highv3.OAuthFlow, 0, 4)
	if s.SecurityScheme.Flows.Implicit != nil {
		flows = append(flows, s.SecurityScheme.Flows.Implicit)
	}
	if s.SecurityScheme.Flows.AuthorizationCode != nil {
		flows = append(flows, s.SecurityScheme.Flows.AuthorizationCode)
	}
	if s.SecurityScheme.Flows.Password != nil {
		flows = append(flows, s.SecurityScheme.Flows.Password)
	}
	if s.SecurityScheme.Flows.ClientCredentials != nil {
		flows = append(flows, s.SecurityScheme.Flows.ClientCredentials)
	}

	return &regardingFlow{
		securityScheme: s,
		flow:           flows,
	}
}

func (f *regardingFlow) AddScope(name, description string) *regardingFlow {
	for _, flow := range f.flow {
		if flow.Scopes == nil {
			flow.Scopes = orderedmap.New[string, string]()
		}
		flow.Scopes.Set(name, description)
	}
	return f
}

func (f *regardingFlow) AuthorizationURL(url string) *regardingFlow {
	for _, flow := range f.flow {
		flow.AuthorizationUrl = url
	}
	return f
}

func (f *regardingFlow) TokenURL(url string) *regardingFlow {
	for _, flow := range f.flow {
		flow.TokenUrl = url
	}
	return f
}

func (f *regardingFlow) RefreshURL(url string) *regardingFlow {
	for _, flow := range f.flow {
		flow.RefreshUrl = url
	}
	return f
}

func (f *regardingFlow) AddExtension(name string, value *yaml.Node) *regardingFlow {
	for _, flow := range f.flow {
		if flow.Extensions == nil {
			flow.Extensions = orderedmap.New[string, *yaml.Node]()
		}
		flow.Extensions.Set(name, value)
	}
	return f
}
