package auth

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashPass(t *testing.T) {
	tests := []struct {
		name       string
		pass       string
		hashedPass string
	}{
		{
			name:       "digits_small",
			pass:       "123",
			hashedPass: "343932676c3132624143744154314d7940bd001563085fc35165329ea1ff5c5ecbdbbeef",
		},
		{
			name:       "digits_large",
			pass:       "1233213213132123124465565681478732804197120947127480327560",
			hashedPass: "343932676c3132624143744154314d79732049e4e7cd01be7ca356d6af789551f44c9df2",
		},
		{
			name:       "alpha_small",
			pass:       "abc",
			hashedPass: "343932676c3132624143744154314d79a9993e364706816aba3e25717850c26c9cd0d89d",
		},
		{
			name:       "alpha_large",
			pass:       "abcdfjawfehuiwafhoeuwhfiowafluwuhfwiqwueiopuia",
			hashedPass: "343932676c3132624143744154314d7905d8287962a12ba087d15b65d258bd20632821de",
		},
		{
			name:       "random",
			pass:       "NnhqXCgRtX7FPSH!zTpKYkG#dAeoM!rJrpgEB6Hr",
			hashedPass: "343932676c3132624143744154314d79c800078494d3bde04c6b000f72f9e67c266baaaf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.hashedPass, HashPass(tt.pass))
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name  string
		login string
		token string
	}{
		{
			name:  "Test #1",
			login: "test1",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJNYXBDbGFpbXMiOm51bGwsImxvZ2luIjoidGVzdDEifQ.x2nikf-_mOoUJLn-LaqU7YdQDyZHiVHN-TKQjSmTxU0",
		},
		{
			name:  "Test #2",
			login: "test2",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJNYXBDbGFpbXMiOm51bGwsImxvZ2luIjoidGVzdDIifQ.36IDGm0dN5mtMvvIkQXaGuOnC_waB0DFPUL3aG_v2Fk",
		},
		{
			name:  "Test #3",
			login: "test3",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJNYXBDbGFpbXMiOm51bGwsImxvZ2luIjoidGVzdDMifQ.WmUPAvySi4hdV6ECKTUb3viF6Ug41hlCcTML6mk6zsY",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.login)
			require.NoError(t, err)
			require.Equal(t, tt.token, token)
		})
	}
}

func TestParseToken(t *testing.T) {
	type args struct {
		rawToken string
	}
	tests := []struct {
		name  string
		login string
	}{
		{
			name:  "digits",
			login: "2139087231",
		},
		{
			name:  "alpha",
			login: "ifoadfoapsfjaisofjaipsfapi",
		},
		{
			name:  "random",
			login: "af$QQScLQdcK6Xg#",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.login)
			require.NoError(t, err)
			login, err := ParseToken(token)
			require.NoError(t, err)
			require.Equal(t, tt.login, login)
		})
	}
}
