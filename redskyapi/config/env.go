/*
Copyright 2020 GramLabs, Inc.

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

package config

import "os"

// NOTE: The environment loader is NOT backward compatible with the old environment variables

func envLoader(cfg *ClientConfig) error {
	srv, az, _, _, err := contextConfig(&cfg.data, cfg.data.CurrentContext)
	if err != nil {
		return err
	}

	// Build a server using environment variables, re-compute defaults and merge it in to overwrite
	envSrv := &Server{
		Identifier: os.Getenv("REDSKY_ADDRESS"),
		RedSky: RedSkyServer{
			ExperimentsEndpoint: "",
			AccountsEndpoint:    "",
		},
		Authorization: AuthorizationServer{
			AuthorizationEndpoint: "",
			TokenEndpoint:         os.Getenv("REDSKY_OAUTH2_TOKEN_URL"),
			RegistrationEndpoint:  "",
		},
	}
	if err := defaultServer(envSrv); err != nil {
		return err
	}
	mergeServer(srv, envSrv)

	// Only take a complete credential
	cc := &ClientCredential{
		ClientID:     os.Getenv("REDSKY_OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("REDSKY_OAUTH2_CLIENT_SECRET"),
	}
	if cc.ClientID != "" && cc.ClientSecret != "" {
		mergeAuthorization(az, &Authorization{Credential: Credential{ClientCredential: cc}})
	}

	return nil
}

// LegacyEnvMapping produces a map of environment variables generated from a configuration
func LegacyEnvMapping(cc *ClientConfig, includeController bool) (map[string][]byte, error) {
	srv, az, _, ctrl, err := contextConfig(&cc.data, cc.data.CurrentContext)
	if err != nil {
		return nil, err
	}

	env := make(map[string][]byte)

	// Record the server information
	env["REDSKY_ADDRESS"] = []byte(srv.Identifier)
	env["REDSKY_OAUTH2_TOKEN_URL"] = []byte(srv.Authorization.TokenEndpoint)

	// Record the authorization information
	if az.Credential.ClientCredential != nil {
		env["REDSKY_OAUTH2_CLIENT_ID"] = []byte(az.Credential.ClientID)
		env["REDSKY_OAUTH2_CLIENT_SECRET"] = []byte(az.Credential.ClientSecret)
	}

	// Optionally record environment variables from the controller configuration
	if includeController {
		for i := range ctrl.Env {
			env[ctrl.Env[i].Name] = []byte(ctrl.Env[i].Value)
		}
	}

	// Strip out blanks
	for k, v := range env {
		if len(v) == 0 {
			delete(env, k)
		}
	}
	return env, nil
}
